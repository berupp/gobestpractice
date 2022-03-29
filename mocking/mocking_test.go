package mocking_test

import (
	"github.com/stretchr/testify/assert"
	"minimalgo/mocking"
	"testing"
)

func TestToUpperCaseFromDatabase(t *testing.T) {
	original := mocking.GetStringFromDatabase //Store original to restore behavior after test
	defer func() {
		mocking.GetStringFromDatabase = original //Make sure behavior is restored after test
	}()

	mocking.GetStringFromDatabase = func(string) string { //Mock
		return "mock"
	}

	result := mocking.ToUpperCaseFromDatabase("abc")
	assert.Equal(t, "MOCK", result)
}
