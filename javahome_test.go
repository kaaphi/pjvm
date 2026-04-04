package main

import (
	"fmt"
	"testing"
	"testing/fstest"
	"io/fs"

	"github.com/stretchr/testify/assert"
)

type TestVolumeFsPathHandler struct {
	rootFs fstest.MapFS
}

func (handler TestVolumeFsPathHandler) RootFS() fs.StatFS {
	return handler.rootFs
}

func (handler TestVolumeFsPathHandler) FromVolumePath(p string) (string, error) {
	return p, nil
}

func (handler TestVolumeFsPathHandler) ToVolumePath(p string) (string) {
	return p
}

func testFileSystem(paths ...string) fstest.MapFS {
	fsys := fstest.MapFS{
		"dir1/not_java/bin/file1":  {},
		"dir1/not_java2/file1":     {},
		"dir2/not_java2/bin/file1": {},
		"dir2/not_java2/file1":     {},
		"dir3/file.txt":            {},
		"dir3/file.go":             {},
		"dir3/subdir/x.go":         {},
	}

	for _, path := range paths {
		fsys[path] = &fstest.MapFile{}
	}

	return fsys
}

func TestVersionMatch(t *testing.T) {
	cases := []struct {
		javaHome       string
		versionMatch   string
		expectedResult bool
	}{
		{javaHome: "java1.8", versionMatch: "", expectedResult: true},
		{javaHome: "java", versionMatch: "", expectedResult: true},
		{javaHome: "dir/java", versionMatch: "", expectedResult: true},
		{javaHome: "dir/java1.8", versionMatch: "1.8", expectedResult: true},
		{javaHome: "dir/java1.8/subdir", versionMatch: "1.8", expectedResult: false},
		{javaHome: "java1.8", versionMatch: "1", expectedResult: true},
		{javaHome: "java1.8", versionMatch: "1.8", expectedResult: true},
		{javaHome: "java1.8other", versionMatch: "1.8", expectedResult: true},
		{javaHome: "java17.0.4_9", versionMatch: "1", expectedResult: false},
		{javaHome: "java17.0.4_9", versionMatch: "17", expectedResult: true},
		{javaHome: "java17.0.4_9", versionMatch: "17.0", expectedResult: true},
		{javaHome: "java17.0.4_9", versionMatch: "17.0.4", expectedResult: true},
		{javaHome: "java17.0.4_9", versionMatch: "17.0.4_9", expectedResult: true},
		{javaHome: "java17.0.4_9", versionMatch: "17.0.3", expectedResult: false},
		{javaHome: "java17.0.4_9", versionMatch: "17.1", expectedResult: false},
	}

	assert := assert.New(t)

	for _, tt := range cases {
		t.Run(fmt.Sprintf("%s, %s", tt.javaHome, tt.versionMatch), func(t *testing.T) {
			assert.Equal(tt.expectedResult, versionMatches(tt.javaHome, tt.versionMatch))
		})
	}
}

func TestFindJavaVersions(t *testing.T) {
	fsys := testFileSystem(
		"dir1/java1.8/bin/javac.exe",
		"dir2/java25/bin/javac.exe",
	)

	handler := TestVolumeFsPathHandler{rootFs: fsys}

	actual := findJavaVersions([]string{"dir1"}, "", func(p string) VolumeFsPathHandler {
		return handler
	})

	assert.Equal(t, []string{"dir1/java1.8"}, actual)
}
