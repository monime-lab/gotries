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
	_                Backoff = &fibonacciBackoff{}
	FibonacciBackoff         = NewFibonacciBackoff2(FibonacciConfig{
		Delay:    defaultBaseDelay,
		MaxDelay: 30 * time.Second,
		Jitter:   defaultJitterFactor,
	})
)

type FibonacciConfig struct {
	// Delay is the amount of time to backoff after every failure.
	Delay time.Duration
	// Jitter is the factor with which the delays are randomized.
	Jitter float64
	// MaxDelay is the upper bound of backoff delay.
	MaxDelay time.Duration
}

// NewFibonacciBackoff2 returns a Backoff that returns FibonacciBackoff sequenced wait delays between failures
func NewFibonacciBackoff2(config FibonacciConfig) Backoff {
	if config.Delay <= 0 {
		config.Delay = defaultBaseDelay
	}
	if config.Jitter <= 0 {
		config.Jitter = defaultJitterFactor
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

func (b *fibonacciBackoff) NextDelay(failures int) time.Duration {
	if failures == 0 {
		return b.config.Delay
	}
	max := float64(b.config.MaxDelay)
	backoff := float64(b.config.Delay) * float64(b.nextDelay(failures))
	if backoff > max {
		backoff = max
	}
	// Randomize the backoff delay, so we don't have multiple delays waking up at the same instants
	backoff = addRandomJitterToDelay(backoff, b.config.Delay, b.config.Jitter)
	return time.Duration(math.Max(backoff, 0))
}
