package options_test

import (
	"github.com/stretchr/testify/assert"
	"minimalgo/options"
	"testing"
)

func TestOptionsPattern(t *testing.T) {
	{
		connection := options.New(options.WithIP("localhost"))

		assert.Equal(t, "ip: localhost, port: 0", connection.ToString())
	}
	{
		connection := options.New(
			options.WithIP("localhost"),
			options.WithPort(90008))

		assert.Equal(t, "ip: localhost, port: 90008", connection.ToString())
	}
}
