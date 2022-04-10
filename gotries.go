/*
 * Copyright (C) 2021, Monime Ltd, All Rights Reserved.
 * Unauthorized copy or sharing of this file through
 * any medium is strictly not allowed.
 */

package gotries

import (
	"context"
	"errors"
	"log"
	"time"
)

//goland:noinspection GoUnusedGlobalVariable
var (
	_                                Retry = &retry{}
	_                                State = &retry{}
	DefaultRecoverableErrorPredicate       = func(err error) bool {
		return !(errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded))
	}
)

type (
	State interface {
		Retrying() bool
		LastError() error
		CurrentAttempts() int
		Context() context.Context
		StopNextAttempt(stop bool)
	}
	Callback  func(interface{}, error)
	Callback2 func(interface{}, interface{}, error)
	Runnable  func(State) error
	Callable  func(State) (interface{}, error)
	Callable2 func(State) (interface{}, interface{}, error)
	Retry     interface {
		Run(ctx context.Context, runnable Runnable) error
		Call(ctx context.Context, callable Callable) (interface{}, error)
		Call2(ctx context.Context, callable2 Callable2) (interface{}, interface{}, error)
	}
	Config struct {
		MaxAttempts               int
		TaskName                  string
		Backoff                   Backoff
		RecoverableErrorPredicate func(err error) bool
	}
)

// Run is a syntactic sugar
func Run(ctx context.Context, runnable Runnable, options ...Option) error {
	return NewRetry(options...).Run(ctx, runnable)
}

// Call is a syntactic sugar
func Call(ctx context.Context, callable Callable, options ...Option) (interface{}, error) {
	return NewRetry(options...).Call(ctx, callable)
}

// Call2 is a syntactic sugar
func Call2(ctx context.Context, callable2 Callable2, options ...Option) (interface{}, interface{}, error) {
	return NewRetry(options...).Call2(ctx, callable2)
}

// WithTaskName name of the task useful for debugging...
func WithTaskName(name string) Option {
	return optionFunc(func(c *Config) {
		c.TaskName = name
	})
}

// WithMaxAttempts set the max retry attempts before giving up; < 0 means keep retrying "forever"
// Note, attempt semantics begins after the first execution.
func WithMaxAttempts(attempts int) Option {
	return optionFunc(func(c *Config) {
		c.MaxAttempts = attempts
	})
}

// WithBackoff set the retry backoff algorithm to use. Default is ExponentialBackoff
func WithBackoff(backoff Backoff) Option {
	return optionFunc(func(c *Config) {
		c.Backoff = backoff
	})
}

// WithRecoverableErrorPredicate set the predicate use to test whether an error is recoverable
// or not before a retry is scheduled.
// The default ensures the error is neither context.Canceled nor context.DeadlineExceeded
func WithRecoverableErrorPredicate(predicate func(err error) bool) Option {
	return optionFunc(func(c *Config) {
		c.RecoverableErrorPredicate = predicate
	})
}

type Option interface {
	apply(c *Config)
}

type optionFunc func(c *Config)

func (f optionFunc) apply(c *Config) {
	f(c)
}

func NewRetry(options ...Option) Retry {
	config := &Config{MaxAttempts: 4}
	for _, option := range options {
		option.apply(config)
	}
	if config.TaskName == "" {
		config.TaskName = "default"
	}
	if config.Backoff == nil {
		config.Backoff = ExponentialBackoff
	}
	if config.RecoverableErrorPredicate == nil {
		config.RecoverableErrorPredicate = DefaultRecoverableErrorPredicate
	}
	return &retry{config: config}
}

type retry struct {
	stopNextAttempt bool
	attempts        int
	lastError       error
	config          *Config
	context         context.Context
}

func (r *retry) Retrying() bool {
	return r.attempts > 0
}

func (r *retry) LastError() error {
	return r.lastError
}

func (r *retry) CurrentAttempts() int {
	return r.attempts
}

func (r *retry) Context() context.Context {
	return r.context
}

func (r *retry) StopNextAttempt(stop bool) {
	r.stopNextAttempt = stop
}

func (r *retry) Run(ctx context.Context, runnable Runnable) (err error) {
	r.context = ctx
	if err = runnable(r); err != nil {
		for r.scheduleRetry(err) {
			if err = runnable(r); err == nil {
				break
			}
		}
	}
	return
}

func (r *retry) Call(ctx context.Context, callable Callable) (res interface{}, err error) {
	r.context = ctx
	if res, err = callable(r); err != nil {
		for r.scheduleRetry(err) {
			if res, err = callable(r); err == nil {
				break
			}
		}
	}
	return
}

func (r *retry) Call2(ctx context.Context, callable2 Callable2) (res1 interface{}, res2 interface{}, err error) {
	r.context = ctx
	if res1, res2, err = callable2(r); err != nil {
		for r.scheduleRetry(err) {
			if res1, res2, err = callable2(r); err == nil {
				break
			}
		}
	}
	return
}

func (r *retry) scheduleRetry(err error) bool {
	r.attempts++
	r.lastError = err
	if err != nil && !r.config.RecoverableErrorPredicate(err) {
		r.StopNextAttempt(true)
	}
	if !r.stopNextAttempt && (r.config.MaxAttempts < 0 || r.attempts <= r.config.MaxAttempts) {
		delay := r.config.Backoff.NextDelay(r.attempts)
		log.Printf("retry for failed task[%s], error[%s], attempts[%d], nextDelayMillis[%d]",
			r.config.TaskName, err, r.attempts, delay.Milliseconds())
		if r.sleep(delay) {
			return true
		}
		// context was canceled
	}
	r.attempts-- // undo
	return false
}

func (r *retry) sleep(delay time.Duration) bool {
	timer := time.NewTimer(delay)
	for {
		select {
		case <-r.context.Done():
			if !timer.Stop() {
				<-timer.C
			}
			// context was canceled
			return false
		case <-timer.C:
			return true
		}
	}
}
