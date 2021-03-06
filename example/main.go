//nolint
package main

import (
	"context"
	"errors"
	"github.com/monime-lab/gotries"
	"log"
)

func main() {
	//exampleOne()
	exampleTwo()
	//exampleThree()
	//customDefaultOptions()
}

func exampleOne() {
	err := gotries.Run(context.TODO(), func(state gotries.State) error {
		if state.CurrentAttempts() <= 2 {
			return errors.New("some error occurred")
		}
		log.Printf("Task completed")
		return nil
	})
	if err != nil {
		panic(err)
	}
}

func exampleTwo() {
	// the library is cancel or timeout aware on the Context during scheduling or on an error
	resp, err := gotries.Call(context.TODO(), func(state gotries.State) (interface{}, error) {
		if state.CurrentAttempts() == 6 {
			//return "It's a success!!!", nil
		}
		// if for_some_condition {
		// 	 state.StopNextAttempt(true)
		// }
		return nil, errors.New("something wen wrong")
	}, gotries.WithMaxAttempts(10))
	if err != nil {
		panic(err)
	}
	log.Printf("Response: %s", resp)
}

func exampleThree() {
	resp, err := gotries.Call(context.TODO(), // there is a 'Call2' also
		func(state gotries.State) (interface{}, error) {
			return getName(state.Context())
		},
		gotries.WithMaxAttempts(5),
		gotries.WithTaskName("getName"), // for debugging
		//gotries.WithBackoff(gotries.LinearBackoff),
		//gotries.WithBackoff(gotries.ConstantBackoff),
		//gotries.WithBackoff(gotries.FibonacciBackoff),
		gotries.WithBackoff(gotries.ExponentialBackoff),
		//gotries.WithBackoff(gotries.NewConstantBackoff(1*time.Second)),
		//gotries.WithBackoff(gotries.NewExponentialBackoff(gotries.ExponentialBackoffConfig{
		//	Multiplier: 2.0,
		//	Jitter:     0.2,
		//	BaseDelay:  500 * time.Millisecond,
		//	MaxDelay:   1 * time.Minute,
		//})),
		gotries.WithRecoverableErrorPredicate(func(err error) bool {
			// same as gotries.DefaultRecoverableErrorPredicate
			return !(errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded))
		}),
	)
	if err != nil {
		panic(err)
	}
	log.Printf("Response: %s", resp)
}

func getName(ctx context.Context) (string, error) {
	return "John Doe", nil
}

// func customDefaultOptions() {
//	gotries.SetDefaultOptions(
//		gotries.WithLogger(func(template string, args ...interface{}) {
//			zap.S().Infof(template, args...)
//		}),
//		gotries.WithBackoff(gotries.ConstantBackoff),
//	)
//}
