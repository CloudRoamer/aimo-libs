package config

import (
	"errors"
	"testing"
)

func TestSourceError(t *testing.T) {
	baseErr := errors.New("boom")
	err := &SourceError{Source: "consul", Err: baseErr}

	if err.Error() == "" {
		t.Fatal("Error() should return a message")
	}
	if !errors.Is(err, baseErr) {
		t.Fatal("errors.Is should match wrapped error")
	}
	if err.Unwrap() != baseErr {
		t.Fatal("Unwrap() should return original error")
	}
}
