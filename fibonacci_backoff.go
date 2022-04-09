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
	_         Backoff = &fibonacciBackoff{}
	Fibonacci         = NewFibonacci(FibonacciConfig{
		Delay:    baseDelay,
		MaxDelay: 30 * time.Second,
		Jitter:   0.2,
	})
)

type FibonacciConfig struct {
	// Delay is the amount of time to backoff after every failure.
	Delay time.Duration
	// Jitter is the factor with which backoffs are randomized.
	Jitter float64
	// MaxDelay is the upper bound of backoff delay.
	MaxDelay time.Duration
}

// NewFibonacci returns a Backoff that returns Fibonacci sequenced wait delays between failures
func NewFibonacci(config FibonacciConfig) Backoff {
	if config.Delay == 0 {
		config.Delay = 500 * time.Millisecond
	}
	if config.Jitter == 0 {
		config.Jitter = 0.2
	}
	return &fibonacciBackoff{config: config}
}

type fibonacciBackoff struct {
	config FibonacciConfig
}

func (b *fibonacciBackoff) nextDelay(n int) uint32 {
	// The 48th fibonacci number overflows uint32, hence we cap at 47.
	// In practice, no one will wait for that amount of time, 🤔
	n = int(math.Min(float64(n), 47))
	if n <= 1 {
		return uint32(n)
	}
	n1 := uint32(1)
	n2 := uint32(0)
	for i := 2; i < n; i++ {
		n2 = n1
		n1 += n2
	}
	return n1 + n2
}

func (b *fibonacciBackoff) Next(failures int) time.Duration {
	if failures == 0 {
		return b.config.Delay
	}
	max := float64(b.config.MaxDelay)
	backoff := float64(b.config.Delay) * float64(b.nextDelay(failures))
	if backoff > max {
		backoff = max
	}
	// Randomize backoff delays so we don't have bombarding of the target at the same time
	backoff *= 1 + b.config.Jitter*(rnd.Float64()*2-1)
	return time.Duration(math.Max(backoff, 0))
}
