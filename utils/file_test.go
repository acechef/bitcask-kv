package utils

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestDirSize(t *testing.T) {
	dir, _ := os.Getwd()
	dirSize, err := DirSize(dir)
	assert.Nil(t, err)
	assert.True(t, dirSize > 0)
}

func TestAvailableDishSize(t *testing.T) {
	size, err := AvailableDishSize()
	assert.Nil(t, err)
	assert.True(t, size > 0)
}
