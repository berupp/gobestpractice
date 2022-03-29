package variadictests_test

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestToLowerCase(t *testing.T) {
	//Test 'suit' set up here
	// ... setup ...
	defer func() {
		//Optimal test 'suit' tear down here:
		// ... tear down ...
	}()

	var tests = []struct {
		Name           string
		Input          string
		ExpectedOutput string
	}{
		{
			Name:           "Uppercase",
			Input:          "UPPERCASE",
			ExpectedOutput: "uppercase",
		},
		{
			Name:           "CamelCase",
			Input:          "CamelCase",
			ExpectedOutput: "camelcase",
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			//per-test setup
			lower := strings.ToLower(test.Input)
			assert.Equal(t, test.ExpectedOutput, lower)
			//per-test tear down
		})
	}
	//Optional test 'suit' tear down here
	// ... tear down that may not execute on test failure ...
}

func TestToLowerCaseComplexObject(t *testing.T) {
	var tests = []struct {
		Name           string
		Input          func() []*string //func() provides a nice way to set p more complex inputs
		ExpectedOutput []string
	}{
		{
			Name: "Uppercase",
			Input: func() []*string {
				a := "UPPERCASE"
				return []*string{&a}
			},
			ExpectedOutput: []string{"uppercase"},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			input := test.Input() //Use the input
			assert.Equal(t, 1, len(input))
		})
	}
}
