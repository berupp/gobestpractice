package errorhandling_test

import (
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
}
