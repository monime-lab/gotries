/*
 * Copyright 2022 Monime Ltd, licensed under the
 * Apache License, Version 2.0 (the "License");
 */

package gotries

import (
	"math/rand"
	"time"
)

const (
	defaultJitterFactor = 0.2
	defaultBaseDelay    = 50 * time.Millisecond
)

var (
	rnd = rand.New(rand.NewSource(time.Now().UnixNano()))
)

type (
	// Backoff defines the backoff abstraction,
	// NB: Implementations must be safe for concurrent use
	Backoff interface {
		// NextDelay returns the amount of time to wait before the next
		// retry given the specified number of consecutive failures.
		NextDelay(failures int) time.Duration
	}
)

func addRandomJitterToDelay(delay float64, baseDelay time.Duration, jitterFactor float64) float64 {
	return delay + (float64(baseDelay) * jitterFactor * rnd.Float64())
}
