package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConfigFileNotFount(t *testing.T) {
	configFile := "../../config.yaml"
	_, err := NewConfig(&configFile)
	assert.NotNil(t, err)
}

func TestNewConfigSuccess(t *testing.T) {
	configFile := "../config.yaml"
	c, err := NewConfig(&configFile)
	assert.Nil(t, err)
	assert.NotNil(t, c)
}
