package lifecycle

import (
	"context"
	"time"
)

type HeartbeatLoopConfig struct {
	Interval      time.Duration
	RetryBackoff  time.Duration
	Heartbeat     func(context.Context) error
	Recover       func(context.Context) error
	ShouldRecover func(error) bool
}

func RunHeartbeatLoop(ctx context.Context, cfg HeartbeatLoopConfig) {
	if cfg.Interval <= 0 || cfg.Heartbeat == nil {
		return
	}

	ticker := time.NewTicker(cfg.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			err := cfg.Heartbeat(ctx)
			if err == nil {
				continue
			}

			if cfg.Recover != nil && cfg.ShouldRecover != nil && cfg.ShouldRecover(err) {
				if recoverErr := cfg.Recover(ctx); recoverErr == nil {
					continue
				}
			}

			if cfg.RetryBackoff <= 0 {
				continue
			}
			select {
			case <-ctx.Done():
				return
			case <-time.After(cfg.RetryBackoff):
			}
		}
	}
}
