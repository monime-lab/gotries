/*
 * Copyright (C) 2021, Monime Ltd, All Rights Reserved.
 * Unauthorized copy or sharing of this file through
 * any medium is strictly not allowed.
 */

package gotries

import (
	"math"
	"time"
)

//goland:noinspection GoUnusedGlobalVariable
var (
	_           Backoff = &exponentialBackoff{}
	Exponential         = NewExponential(ExponentialConfig{
		Multiplier: 2.0,
		Jitter:     0.2,
		BaseDelay:  baseDelay,
		MaxDelay:   30 * time.Second,
	})
)

type ExponentialConfig struct {
	// BaseDelay is the amount of time to backoff after the first failure.
	BaseDelay time.Duration
	// Multiplier is the factor with which to multiply backoffs after a
	// failed retry. Should ideally be greater than 1.
	Multiplier float64
	// Jitter is the factor with which backoffs are randomized.
	Jitter float64
	// MaxDelay is the upper bound of backoff delay.
	MaxDelay time.Duration
}

// NewExponential returns a Backoff that returns exponential wait delays between failures
func NewExponential(config ExponentialConfig) Backoff {
	return &exponentialBackoff{
		config: config,
	}
}

type exponentialBackoff struct {
	config ExponentialConfig
}

func (b *exponentialBackoff) Next(failures int) time.Duration {
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
	// Randomize backoff delays so we don't have bombarding of the target at the same time
	backoff *= 1 + b.config.Jitter*(rnd.Float64()*2-1)
	return time.Duration(math.Max(backoff, 0))
}
