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
		LastError() error
		CurrentRetry() int
		StopRetry(stop bool)
	}
	Callback  func(interface{}, error)
	Callback2 func(interface{}, interface{}, error)
	Runnable  func(context.Context, State) error
	Callable  func(context.Context, State) (interface{}, error)
	Callable2 func(context.Context, State) (interface{}, interface{}, error)
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

func Run(ctx context.Context, runnable Runnable, options ...Option) error {
	return NewRetry(options...).Run(ctx, runnable)
}

func Call(ctx context.Context, callable Callable, options ...Option) (interface{}, error) {
	return NewRetry(options...).Call(ctx, callable)
}

func Call2(ctx context.Context, callable2 Callable2, options ...Option) (interface{}, interface{}, error) {
	return NewRetry(options...).Call2(ctx, callable2)
}

// WithTaskName name of the task useful for debugging...
func WithTaskName(name string) Option {
	return optionFunc(func(c *Config) {
		c.TaskName = name
	})
}

// WithMaxAttempts set the max retry attempts before giving up; -1 means retry "forever"
func WithMaxAttempts(attempts int) Option {
	return optionFunc(func(c *Config) {
		c.MaxAttempts = attempts
	})
}

// WithBackoff the backoff algorithm to use
func WithBackoff(backoff Backoff) Option {
	return optionFunc(func(c *Config) {
		c.Backoff = backoff
	})
}

// WithDefaultRecoverableErrorPredicate the backoff algorithm to use
func WithDefaultRecoverableErrorPredicate(predicate func(err error) bool) Option {
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
		config.Backoff = Exponential
	}
	if config.RecoverableErrorPredicate == nil {
		config.RecoverableErrorPredicate = DefaultRecoverableErrorPredicate
	}
	return &retry{config: config}
}

type retry struct {
	stop      bool
	attempts  int
	lastError error
	config    *Config
}

func (r *retry) LastError() error {
	return r.lastError
}

func (r *retry) StopRetry(stop bool) {
	r.stop = stop
}

func (r *retry) CurrentRetry() int {
	return r.attempts
}

func (r *retry) Run(ctx context.Context, runnable Runnable) (err error) {
	if err = runnable(ctx, r); err != nil {
		for r.scheduleRetry(err) {
			if err = runnable(ctx, r); err == nil {
				break
			}
		}
	}
	return
}

func (r *retry) Call(ctx context.Context, callable Callable) (res interface{}, err error) {
	if res, err = callable(ctx, r); err != nil {
		for r.scheduleRetry(err) {
			if res, err = callable(ctx, r); err == nil {
				break
			}
		}
	}
	return
}

func (r *retry) Call2(ctx context.Context, callable2 Callable2) (res1 interface{}, res2 interface{}, err error) {
	if res1, res2, err = callable2(ctx, r); err != nil {
		for r.scheduleRetry(err) {
			if res1, res2, err = callable2(ctx, r); err == nil {
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
		r.StopRetry(true)
	}
	if !r.stop && (r.config.MaxAttempts == -1 || r.attempts <= r.config.MaxAttempts) {
		delay := r.config.Backoff.Next(r.attempts)
		log.Printf("retry for failed task[%s], error[%s], attempts[%d], nextDelayMillis[%d]",
			r.config.TaskName, err, r.attempts, delay.Milliseconds())
		time.Sleep(delay)
		return true
	}
	r.attempts-- // undo
	return false
}
