package mover

import (
	"context"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-vgo/robotgo"
	"github.com/itssajan/alpaca/config"
)

type Mover struct {
	enabled      atomic.Bool
	lastUserMove atomic.Int64 // unix nano
	alpacaMoving atomic.Int32 // counter: >0 means Alpaca is moving cursor
	nextMoveAt   atomic.Int64 // unix nano of next scheduled move
	configCh     chan config.Config
	mu           sync.Mutex
	cancel       context.CancelFunc
}

func New() *Mover {
	m := &Mover{
		configCh: make(chan config.Config, 1),
	}
	m.enabled.Store(true)
	return m
}

func (m *Mover) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	m.mu.Lock()
	m.cancel = cancel
	m.mu.Unlock()

	go m.pollUserActivity(ctx)
	go m.scheduler(ctx)
}

func (m *Mover) Stop() {
	m.mu.Lock()
	if m.cancel != nil {
		m.cancel()
	}
	m.mu.Unlock()
}

func (m *Mover) SetEnabled(v bool) {
	m.enabled.Store(v)
}

func (m *Mover) Enabled() bool {
	return m.enabled.Load()
}

// IsUserActive returns true if real user mouse activity was detected within the
// configured quiet period. Always false when Alpaca itself is moving.
func (m *Mover) IsUserActive() bool {
	cfg := config.Get()
	quiet := time.Duration(cfg.Idle.QuietPeriodSeconds) * time.Second
	last := time.Unix(0, m.lastUserMove.Load())
	return time.Since(last) < quiet
}

// NextMoveIn returns how long until the next scheduled move fires.
// Returns 0 if already past due or unknown.
func (m *Mover) NextMoveIn() time.Duration {
	at := time.Unix(0, m.nextMoveAt.Load())
	if at.IsZero() {
		return 0
	}
	d := time.Until(at)
	if d < 0 {
		return 0
	}
	return d
}

func (m *Mover) ReloadConfig(cfg config.Config) {
	select {
	case m.configCh <- cfg:
	default:
		select {
		case <-m.configCh:
		default:
		}
		m.configCh <- cfg
	}
}

func (m *Mover) MoveNow() {
	go func() {
		m.alpacaMoving.Add(1)
		defer m.alpacaMoving.Add(-1)
		cfg := config.Get()
		execPattern(cfg)
	}()
}

func (m *Mover) pollUserActivity(ctx context.Context) {
	px, py := robotgo.Location()
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cx, cy := robotgo.Location()
			dx, dy := cx-px, cy-py
			// ignore sub-pixel jitter (<5px) and Alpaca's own movements
			if m.alpacaMoving.Load() == 0 && dx*dx+dy*dy >= 25 {
				m.lastUserMove.Store(time.Now().UnixNano())
			}
			px, py = cx, cy
		}
	}
}

func (m *Mover) scheduler(ctx context.Context) {
	cfg := config.Get()

	for {
		delay := randomInterval(cfg.Interval.MinSeconds, cfg.Interval.MaxSeconds)
		m.nextMoveAt.Store(time.Now().Add(delay).UnixNano())
		timer := time.NewTimer(delay)

		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case newCfg := <-m.configCh:
			timer.Stop()
			cfg = newCfg
			continue
		case <-timer.C:
		}

		m.nextMoveAt.Store(0)

		if !m.enabled.Load() {
			continue
		}

		cfg = config.Get()
		quiet := time.Duration(cfg.Idle.QuietPeriodSeconds) * time.Second
		last := time.Unix(0, m.lastUserMove.Load())
		if time.Since(last) < quiet {
			continue
		}

		m.alpacaMoving.Add(1)
		execPattern(cfg)
		m.alpacaMoving.Add(-1)
	}
}

func execPattern(cfg config.Config) {
	total := cfg.Movement.JitterWeight + cfg.Movement.DriftWeight + cfg.Movement.WanderWeight
	if total <= 0 {
		return
	}
	r := rand.Intn(total)
	switch {
	case r < cfg.Movement.JitterWeight:
		jitter()
	case r < cfg.Movement.JitterWeight+cfg.Movement.DriftWeight:
		drift()
	default:
		wander()
	}
}

func randomInterval(minSec, maxSec int) time.Duration {
	if minSec >= maxSec {
		return time.Duration(minSec) * time.Second
	}
	sec := minSec + rand.Intn(maxSec-minSec)
	return time.Duration(sec) * time.Second
}
