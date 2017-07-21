package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConfiguration(t *testing.T) {

	//test
	config := new(configuration)

	//assert
	assert.Implements(t, (*Interface)(nil), config, "should implement the config interface")
}
