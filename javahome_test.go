package main

import (
	"fmt"
	"io/fs"
	"testing"
	"testing/fstest"

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

func (handler TestVolumeFsPathHandler) ToVolumePath(p string) (string, error) {
	return p, nil
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
		javaHome       JavaHome
		versionMatch   string
		expectedResult bool
	}{
		{javaHome: JavaHome{JavaHomePath: "java1.8", JavaVersion: "1.8"}, versionMatch: "", expectedResult: true},
		{javaHome: JavaHome{JavaHomePath: "java", JavaVersion: ""}, versionMatch: "", expectedResult: false},
		{javaHome: JavaHome{JavaHomePath: "dir/java", JavaVersion: ""}, versionMatch: "", expectedResult: false},
		{javaHome: JavaHome{JavaHomePath: "dir/java1.8", JavaVersion: "1.8"}, versionMatch: "1.8", expectedResult: true},
		{javaHome: JavaHome{JavaHomePath: "dir/java1.8/subdir", JavaVersion: ""}, versionMatch: "1.8", expectedResult: false},
		{javaHome: JavaHome{JavaHomePath: "java1.8", JavaVersion: "1.8"}, versionMatch: "1", expectedResult: true},
		{javaHome: JavaHome{JavaHomePath: "java1.8", JavaVersion: "1.8"}, versionMatch: "1.8", expectedResult: true},
		{javaHome: JavaHome{JavaHomePath: "java1.8other", JavaVersion: "1.8"}, versionMatch: "1.8", expectedResult: true},
		{javaHome: JavaHome{JavaHomePath: "java17.0.4_9", JavaVersion: ""}, versionMatch: "1", expectedResult: false},
		{javaHome: JavaHome{JavaHomePath: "java17.0.4_9", JavaVersion: "17.0.4_9"}, versionMatch: "17", expectedResult: true},
		{javaHome: JavaHome{JavaHomePath: "java17.0.4_9", JavaVersion: "17.0.4_9"}, versionMatch: "17.0", expectedResult: true},
		{javaHome: JavaHome{JavaHomePath: "java17.0.4_9", JavaVersion: "17.0.4_9"}, versionMatch: "17.0.4", expectedResult: true},
		{javaHome: JavaHome{JavaHomePath: "java17.0.4_9", JavaVersion: "17.0.4_9"}, versionMatch: "17.0.4_9", expectedResult: true},
		{javaHome: JavaHome{JavaHomePath: "java17.0.4_9", JavaVersion: "17.0.4_9"}, versionMatch: "17.0.3", expectedResult: false},
		{javaHome: JavaHome{JavaHomePath: "java17.0.4_9", JavaVersion: "17.0.4_9"}, versionMatch: "17.1", expectedResult: false},
	}

	for _, tt := range cases {
		t.Run(fmt.Sprintf("%s, %s", tt.javaHome, tt.versionMatch), func(t *testing.T) {
			version, matches := versionMatches(tt.javaHome.JavaHomePath, tt.versionMatch)

			assert.Equal(t, tt.expectedResult, matches)
			if matches {
				assert.Equal(t, tt.javaHome.JavaVersion, version)
			}
		})
	}
}

func TestFindJavaVersions(t *testing.T) {
	fsys := testFileSystem(
		"dir1/java1.8/bin/javac.exe",
		"dir2/java25/bin/javac.exe",
	)

	handler := TestVolumeFsPathHandler{rootFs: fsys}

	actual, err := findJavaVersions([]string{"dir1"}, "", func(p string) VolumeFsPathHandler {
		return handler
	})

	assert.Nil(t, err)

	assert.Equal(t, []JavaHome{JavaHome{JavaHomePath: "dir1/java1.8", JavaVersion: "1.8"}}, actual)
}
