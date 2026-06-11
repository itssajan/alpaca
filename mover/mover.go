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

// Start launches the idle-detection and scheduler goroutines.
func (m *Mover) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	m.mu.Lock()
	m.cancel = cancel
	m.mu.Unlock()

	go m.pollUserActivity(ctx)
	go m.scheduler(ctx)
}

// Stop shuts down all goroutines.
func (m *Mover) Stop() {
	m.mu.Lock()
	if m.cancel != nil {
		m.cancel()
	}
	m.mu.Unlock()
}

// SetEnabled toggles the mover on/off without stopping goroutines.
func (m *Mover) SetEnabled(v bool) {
	m.enabled.Store(v)
}

// Enabled returns current enabled state.
func (m *Mover) Enabled() bool {
	return m.enabled.Load()
}

// ReloadConfig sends a new config to the running scheduler.
func (m *Mover) ReloadConfig(cfg config.Config) {
	select {
	case m.configCh <- cfg:
	default:
		// drain stale update and replace
		select {
		case <-m.configCh:
		default:
		}
		m.configCh <- cfg
	}
}

// MoveNow triggers an immediate movement regardless of schedule.
func (m *Mover) MoveNow() {
	go func() {
		cfg := config.Get()
		execPattern(cfg)
	}()
}

// pollUserActivity tracks the last time the real user moved the mouse.
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
			if cx != px || cy != py {
				m.lastUserMove.Store(time.Now().UnixNano())
				px, py = cx, cy
			}
		}
	}
}

// scheduler fires movement patterns at randomized intervals.
func (m *Mover) scheduler(ctx context.Context) {
	cfg := config.Get()

	for {
		delay := randomInterval(cfg.Interval.MinSeconds, cfg.Interval.MaxSeconds)
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

		if !m.enabled.Load() {
			continue
		}

		// skip if user was active within quiet period
		quiet := time.Duration(cfg.Idle.QuietPeriodSeconds) * time.Second
		last := time.Unix(0, m.lastUserMove.Load())
		if time.Since(last) < quiet {
			continue
		}

		execPattern(cfg)
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
