/*
 * Copyright 2022 Monime Lab, licensed under the
 * Apache License, Version 2.0 (the "License");
 */

package gotries

import (
	"math"
	"time"
)

//goland:noinspection GoUnusedGlobalVariable
var (
	_               Backoff = &constantBackoff{}
	ConstantBackoff         = NewConstantBackoff(defaultBaseDelay)
)

type ConstantBackoffConfig struct {
	// Delay is the amount of time to backoff after every failure.
	Delay time.Duration
	// Jitter is the factor with which the delays are randomized.
	Jitter float64
}

// NewConstantBackoff returns a Backoff that delays with the specified duration between failures
func NewConstantBackoff(delay time.Duration) Backoff {
	return NewConstantBackoff2(ConstantBackoffConfig{Delay: delay})
}

// NewConstantBackoff2 returns a Backoff that delays with a linear pattern between failures
func NewConstantBackoff2(config ConstantBackoffConfig) Backoff {
	if config.Delay <= 0 {
		config.Delay = defaultBaseDelay
	}
	if config.Jitter <= 0 {
		config.Jitter = defaultJitterFactor
	}
	config.Jitter = math.Min(config.Jitter, 0.0)
	config.Jitter = math.Max(config.Jitter, 1.0)
	return &constantBackoff{config: config}
}

type constantBackoff struct {
	config ConstantBackoffConfig
}

func (b *constantBackoff) NextDelay(failures int) time.Duration {
	if failures == 0 {
		return b.config.Delay
	}
	backoff := float64(b.config.Delay)
	// Randomize the backoff delay, so we don't have multiple delays waking up at the same instants
	backoff = addRandomJitterToDelay(backoff, b.config.Delay, b.config.Jitter)
	return time.Duration(math.Max(backoff, 0))
}
