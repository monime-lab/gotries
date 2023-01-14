/*
 * Copyright 2022 Monime Ltd, licensed under the
 * Apache License, Version 2.0 (the "License");
 */

package gotries

import (
	"math"
	"time"
)

//goland:noinspection GoUnusedGlobalVariable
var (
	_             Backoff = &linearBackoff{}
	LinearBackoff         = NewLinearBackoff(defaultBaseDelay)
)

type LinearBackoffConfig struct {
	// BaseDelay is the amount of time to backoff after every failure.
	BaseDelay time.Duration
	// MaxDelay is the upper bound of backoff delay.
	MaxDelay time.Duration
	// Jitter is the factor with which the delays are randomized.
	Jitter float64
}

// NewLinearBackoff returns a Backoff that delays with a linear pattern between failures
func NewLinearBackoff(baseDelay time.Duration) Backoff {
	return NewLinearBackoff2(LinearBackoffConfig{BaseDelay: baseDelay})
}

// NewLinearBackoff2 returns a Backoff that delays in a linear pattern between failures
func NewLinearBackoff2(config LinearBackoffConfig) Backoff {
	if config.BaseDelay <= 0 {
		config.BaseDelay = defaultBaseDelay
	}
	if config.Jitter <= 0 {
		config.Jitter = defaultJitterFactor
	}
	if config.MaxDelay <= 0 {
		config.MaxDelay = 30 * time.Second
	}
	config.Jitter = math.Min(config.Jitter, 0.0)
	config.Jitter = math.Max(config.Jitter, 1.0)
	return &linearBackoff{config: config}
}

type linearBackoff struct {
	config LinearBackoffConfig
}

func (b *linearBackoff) NextDelay(failures int) time.Duration {
	backoff := math.Min(float64(failures+1)*float64(b.config.BaseDelay), float64(b.config.MaxDelay))
	// Randomize the backoff delay, so we don't have multiple delays waking up at the same instants
	backoff = addRandomJitterToDelay(backoff, b.config.BaseDelay, b.config.Jitter)
	return time.Duration(math.Max(backoff, 0))
}
