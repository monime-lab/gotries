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
	_               Backoff = &constantBackoff{}
	ConstantBackoff         = NewConstantBackoff(baseDelay)
)

type ConstantConfig struct {
	// Delay is the amount of time to backoff after every failure.
	Delay time.Duration
	// Jitter is the factor with which backoffs are randomized.
	Jitter float64
}

// NewConstantBackoff returns a Backoff that returns a constant wait delay between failures
func NewConstantBackoff(delay time.Duration) Backoff {
	return NewConstantBackoff2(delay, 0.2)
}

// NewConstantBackoff2 returns a Backoff that returns a constant wait delay between failures
func NewConstantBackoff2(delay time.Duration, jitter float64) Backoff {
	return NewConstant(ConstantConfig{Delay: delay, Jitter: jitter})
}

// NewConstant returns a Backoff that returns a constant wait delay between failures
func NewConstant(config ConstantConfig) Backoff {
	if config.Delay == 0 {
		config.Delay = baseDelay
	}
	if config.Jitter == 0 {
		config.Jitter = 0.2
	}
	config.Jitter = math.Min(config.Jitter, 0.0)
	config.Jitter = math.Max(config.Jitter, 1.0)
	return &constantBackoff{config: config}
}

type constantBackoff struct {
	config ConstantConfig
}

func (b *constantBackoff) NextDelay(failures int) time.Duration {
	if failures == 0 {
		return b.config.Delay
	}
	backoff := float64(b.config.Delay)
	// Randomize backoff delays so we don't have bombarding of the target at the same time
	backoff *= 1 + b.config.Jitter*(rnd.Float64()*2-1)
	return time.Duration(math.Max(backoff, 0))
}
