package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStepsAsString(t *testing.T) {

	jobs := []interface{}{
		map[string]interface{}{
			"conclusion": "success",
			"status":     "completed",
			"name":       "basic (author)",
		},
		map[string]interface{}{
			"conclusion": "failed",
			"status":     "completed",
			"name":       "basic (checkstyle)",
		},
	}

	assert.Equal(t, "_c", stepsAsString(jobs))

}
