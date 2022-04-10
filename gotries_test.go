//nolint
package gotries

import (
	"context"
	"errors"
	"testing"
)

func TestRunSuccess(t *testing.T) {
	err := Run(context.TODO(), func(state State) error {
		t.Logf("Retries: %d", state.CurrentAttempts())
		return nil
	})
	if err != nil {
		t.Fatalf("after TestRunSuccess, err: %v", err)
	}
}

func TestRunFailure(t *testing.T) {
	err := Run(context.TODO(), func(state State) error {
		t.Logf("Retries: %d", state.CurrentAttempts())
		return errors.New("failure with TestRunFailure")
	})
	if err == nil || err.Error() != "failure with TestRunFailure" {
		t.Fatalf("after TestRunFailure, err: %v", err)
	}
}

// More coming....
