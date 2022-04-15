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
	_                  Backoff = &exponentialBackoff{}
	ExponentialBackoff         = NewExponentialBackoff(ExponentialBackoffConfig{
		Multiplier: 2.0,
		Jitter:     defaultJitterFactor,
		BaseDelay:  defaultBaseDelay,
		MaxDelay:   30 * time.Second,
	})
)

type ExponentialBackoffConfig struct {
	// BaseDelay is the amount of time to backoff after the first failure.
	BaseDelay time.Duration
	// Multiplier is the factor with which to multiply backoff after a
	// failed retry. Should ideally be greater than 1.
	Multiplier float64
	// Jitter is the factor with which the delays are randomized.
	Jitter float64
	// MaxDelay is the upper bound of backoff delay.
	MaxDelay time.Duration
}

// NewExponentialBackoff returns a Backoff that delays with an exponential pattern between failures
func NewExponentialBackoff(config ExponentialBackoffConfig) Backoff {
	if config.Multiplier <= 0 {
		config.BaseDelay = 2.0
	}
	if config.BaseDelay <= 0 {
		config.BaseDelay = defaultBaseDelay
	}
	if config.Jitter <= 0 {
		config.Jitter = defaultJitterFactor
	}
	if config.MaxDelay <= 0 {
		config.MaxDelay = 30 * time.Second
	}
	return &exponentialBackoff{config: config}
}

type exponentialBackoff struct {
	config ExponentialBackoffConfig
}

func (b *exponentialBackoff) NextDelay(failures int) time.Duration {
	if failures == 0 {
		return b.config.BaseDelay
	}
	backoff, max := float64(b.config.BaseDelay), float64(b.config.MaxDelay)
	for backoff < max && failures > 0 {
		backoff *= b.config.Multiplier
		failures--
	}
	if backoff > max {
		backoff = max
	}
	// Randomize the backoff delay, so we don't have multiple delays waking up at the same instants
	backoff = addRandomJitterToDelay(backoff, b.config.BaseDelay, b.config.Jitter)
	return time.Duration(math.Max(backoff, 0))
}
