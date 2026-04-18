package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddJavaHome(t *testing.T) {
	result := updateJavaHomeInPath(
		[]string{
			"/path/1",
			"/path/2",
			"/path/current_java/bin/",
			"/path/3",
		},
		[]string{
			"/path/current_java/bin/java.exe",
		},
		"/javahome",
		GoFilePaths{},
	)

	assert.Equal(t, []string{
		"/javahome/bin",
		"/path/1",
		"/path/2",
		"/path/3",
	}, result)
}
