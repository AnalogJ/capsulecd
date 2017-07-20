package config

import (
	"testing"
	"github.com/stretchr/testify/assert"
)


func TestConfiguration(t *testing.T) {

	//test
	config := new(configuration)

	//assert
	assert.Implements(t, (*Interface)(nil), config, "should implement the config interface")
}
