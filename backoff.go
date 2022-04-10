/*
 * Copyright (C) 2021, Monime Ltd, All Rights Reserved.
 * Unauthorized copy or sharing of this file through
 * any medium is strictly not allowed.
 */

package gotries

import (
	"math/rand"
	"time"
)

const (
	baseDelay = 200 * time.Millisecond
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
