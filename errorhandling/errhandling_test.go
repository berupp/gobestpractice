package errorhandling_test

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"minimalgo/errorhandling"
	"testing"
)

func TestPreDefinedErrorHandling(t *testing.T) {
	err := errorhandling.ReturnPredefinedError()
	//You can use a switch
	switch err {
	case errorhandling.ConnectionError:
		assert.Equal(t, "connection failed", err.Error())
	default:
		t.Fatal("unexpected error")
	}

	//Or a simple comparison
	if err == errorhandling.ConnectionError {
		assert.Equal(t, "connection failed", err.Error())
	} else {
		t.Fatal("unexpected error")
	}
	//Or using errors.Is(), which examines the entire error chain in case of wrapped errors
	if errors.Is(err, errorhandling.ConnectionError) {
		assert.Equal(t, "connection failed", err.Error())
	} else {
		t.Fatal("unexpected error")
	}
}

func TestCustomErrorHandling(t *testing.T) {
	err := errorhandling.ReturnCustomError()

	switch err.(type) {
	case errorhandling.CustomError:
		customError := err.(errorhandling.CustomError)

		assert.Equal(t, 22, customError.Status)
		assert.Equal(t, "Just cause", customError.Reason)
	default:
		t.Fatal("unexpected error")
	}
	//Or using errors.As(...) which examines the entire error chain to find wrapped errorhandling.CustomError
	var ce errorhandling.CustomError
	if errors.As(err, &ce) {
		assert.Equal(t, 22, ce.Status)
		assert.Equal(t, "Just cause", ce.Reason)
	} else {
		t.Fatal("unexpected error")
	}
}

func TestCustomErrorChainHandling(t *testing.T) {
	err := errorhandling.ReturnCustomError()
	//Wrapping the original errorhandling.CustomError into another error using format directive '%w'
	wrapErr := fmt.Errorf("I caught an error %w", err)

	var ce errorhandling.CustomError
	//errors.As() will find the errorhandling.CustomError in the chain
	if errors.As(wrapErr, &ce) {
		assert.Equal(t, 22, ce.Status)
		assert.Equal(t, "Just cause", ce.Reason)
	} else {
		t.Fatal("unexpected error")
	}
}
