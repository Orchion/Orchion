package inference

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewService(t *testing.T) {
	service := NewService()

	assert.NotNil(t, service)
	assert.NotNil(t, service.engines)
	assert.Len(t, service.engines, 0)
}

func TestService_BasicFunctionality(t *testing.T) {
	service := NewService()

	// Test that service initializes correctly
	assert.NotNil(t, service)
	assert.NotNil(t, service.engines)
	assert.Len(t, service.engines, 0)
}