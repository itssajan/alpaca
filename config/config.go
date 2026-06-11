package config

import (
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/BurntSushi/toml"
)

type Interval struct {
	MinSeconds int `toml:"min_seconds"`
	MaxSeconds int `toml:"max_seconds"`
}

type Idle struct {
	QuietPeriodSeconds int `toml:"quiet_period_seconds"`
}

type Movement struct {
	JitterWeight int `toml:"jitter_weight"`
	DriftWeight  int `toml:"drift_weight"`
	WanderWeight int `toml:"wander_weight"`
}

type Config struct {
	Interval Interval `toml:"interval"`
	Idle     Idle     `toml:"idle"`
	Movement Movement `toml:"movement"`
}

var defaults = Config{
	Interval: Interval{MinSeconds: 30, MaxSeconds: 300},
	Idle:     Idle{QuietPeriodSeconds: 3},
	Movement: Movement{JitterWeight: 30, DriftWeight: 30, WanderWeight: 40},
}

var (
	mu       sync.RWMutex
	current  Config
	filePath string
)

func configDir() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("APPDATA"), "alpaca")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "alpaca")
}

func Load() (Config, error) {
	filePath = filepath.Join(configDir(), "config.toml")

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if err2 := save(defaults); err2 != nil {
			return defaults, err2
		}
		mu.Lock()
		current = defaults
		mu.Unlock()
		return defaults, nil
	}

	var cfg Config
	if _, err := toml.DecodeFile(filePath, &cfg); err != nil {
		return defaults, err
	}
	applyDefaults(&cfg)
	mu.Lock()
	current = cfg
	mu.Unlock()
	return cfg, nil
}

func Save(cfg Config) error {
	applyDefaults(&cfg)
	if err := save(cfg); err != nil {
		return err
	}
	mu.Lock()
	current = cfg
	mu.Unlock()
	return nil
}

func Get() Config {
	mu.RLock()
	defer mu.RUnlock()
	return current
}

func save(cfg Config) error {
	dir := configDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	f, err := os.Create(filepath.Join(dir, "config.toml"))
	if err != nil {
		return err
	}
	defer f.Close()
	return toml.NewEncoder(f).Encode(cfg)
}

func applyDefaults(cfg *Config) {
	if cfg.Interval.MinSeconds <= 0 {
		cfg.Interval.MinSeconds = defaults.Interval.MinSeconds
	}
	if cfg.Interval.MaxSeconds <= 0 {
		cfg.Interval.MaxSeconds = defaults.Interval.MaxSeconds
	}
	if cfg.Interval.MaxSeconds < cfg.Interval.MinSeconds {
		cfg.Interval.MaxSeconds = cfg.Interval.MinSeconds + 30
	}
	if cfg.Idle.QuietPeriodSeconds <= 0 {
		cfg.Idle.QuietPeriodSeconds = defaults.Idle.QuietPeriodSeconds
	}
	total := cfg.Movement.JitterWeight + cfg.Movement.DriftWeight + cfg.Movement.WanderWeight
	if total <= 0 {
		cfg.Movement = defaults.Movement
	}
}
