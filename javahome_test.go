package main

import (
	"fmt"
	"io/fs"
	"slices"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
)

type TestCacheEncoder struct {
	cache JavaHomeCache
}

func (encoder *TestCacheEncoder) StoreCache(context *PjvmContext, cache *JavaHomeCache) error {
	encoder.cache = *cache
	return nil
}

func (encoder *TestCacheEncoder) LoadCache(context *PjvmContext) (*JavaHomeCache, error) {
	return &encoder.cache, nil
}

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

func testContext(cache JavaHomeCache, fileSystem fstest.MapFS, basePaths ...string) *PjvmContext {
	handler := TestVolumeFsPathHandler{rootFs: fileSystem}

	return &PjvmContext{
		fileSystemSupplier: func(p string) VolumeFsPathHandler { return handler },
		cache:              cache,
		cacheEncoder:       &TestCacheEncoder{},
		config:             PjvmConfig{BasePaths: basePaths},
	}
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

func TestFindJdksInPaths(t *testing.T) {
	fsys := testFileSystem(
		"dir1/java1.8/bin/javac.exe",
		"dir2/java25/bin/javac.exe",
	)

	handler := TestVolumeFsPathHandler{rootFs: fsys}

	actual, err := findJdksInPaths([]string{"dir1"}, "", func(p string) VolumeFsPathHandler {
		return handler
	})

	assert.Nil(t, err)

	assert.Equal(t, []JavaHome{{JavaHomePath: "dir1/java1.8", JavaVersion: "1.8"}}, actual)
}

func TestFindJdksCacheMiss(t *testing.T) {
	fsys := testFileSystem(
		"dir1/java1.8/bin/javac.exe",
		"dir1/java25/bin/javac.exe",
	)
	cache, _ := buildJavaHomeCache(
		javaHome("17"), javaHome("19"),
	)

	context := *testContext(cache, fsys, "dir1")
	expected := []JavaHome{{JavaHomePath: "dir1/java25", JavaVersion: "25"}}

	result, err := FindJdks(context, "25")
	assert.Nil(t, err, "Expected no error!")

	assert.Equal(t, expected, result, "Expected to find jdk 25 from file system on cache miss")
	v, _ := context.cacheEncoder.(*TestCacheEncoder)
	assert.Equal(t, expected, v.cache.FindJdks("25"), "Expected cache to be stored containing jdk 25")
}

func TestFindJdksCacheHit(t *testing.T) {
	fsys := testFileSystem(
		"dir1/java17/bin/javac.exe",
	)
	cache, jdkMap := buildJavaHomeCache(
		javaHome("17"), javaHome("19"),
	)

	context := *testContext(cache, fsys, "dir1")
	expected := []JavaHome{jdkMap["17"]}

	result, err := FindJdks(context, "17")
	assert.Nil(t, err, "Expected no error!")

	assert.Equal(t, expected, result, "Expected to find jdk 17 from cache")
}

func javaHome(args ...string) JavaHome {
	switch len(args) {
	case 0:
		return JavaHome{}
	case 1:
		return JavaHome{JavaHomePath: args[0], JavaVersion: args[0]}
	default:
		return JavaHome{JavaHomePath: args[1], JavaVersion: args[0]}
	}
}

func buildJavaHomeCache(jdks ...JavaHome) (JavaHomeCache, map[string]JavaHome) {
	cache := JavaHomeCache{}

	javaHomes := map[string]JavaHome{}

	for _, j := range jdks {
		javaHomes[j.JavaHomePath] = j
	}

	cache.SetJdks(jdks)
	return cache, javaHomes
}

func TestJavaHomeCacheSetJdks(t *testing.T) {
	cache, jdks := buildJavaHomeCache(javaHome("17", "17b"), javaHome("24"), javaHome("19"), javaHome("17", "17a"))

	assert.Equal(t, []JavaHome{jdks["17a"], jdks["17b"], jdks["19"], jdks["24"]}, cache.Jdks)
}

func TestJavaHomeCacheFindJdk(t *testing.T) {
	cache, jdks := buildJavaHomeCache(javaHome("1.8"), javaHome("1.8.2"), javaHome("11.0.4"), javaHome("11.0.5"), javaHome("17.0.4"), javaHome("19", "19a"), javaHome("19", "19b"), javaHome("24.0.1"))

	cases := []struct {
		versionSpecifier   string
		expectedResultKeys []string
	}{
		{versionSpecifier: "1.8.2", expectedResultKeys: []string{"1.8.2"}},
		{versionSpecifier: "1.8", expectedResultKeys: []string{"1.8.2", "1.8"}},
		{versionSpecifier: "11", expectedResultKeys: []string{"11.0.4", "11.0.5"}},
		{versionSpecifier: "11.0.4", expectedResultKeys: []string{"11.0.4"}},
		{versionSpecifier: "19", expectedResultKeys: []string{"19a", "19b"}},
		{versionSpecifier: "19.0.0", expectedResultKeys: []string{}},
		{versionSpecifier: "30", expectedResultKeys: []string{}},
		{versionSpecifier: "17", expectedResultKeys: []string{"17.0.4"}},
		{versionSpecifier: "17.0", expectedResultKeys: []string{"17.0.4"}},
		{versionSpecifier: "17.0.4", expectedResultKeys: []string{"17.0.4"}},
		{versionSpecifier: "24", expectedResultKeys: []string{"24.0.1"}},
		{versionSpecifier: "2", expectedResultKeys: []string{}},
	}

	for _, tt := range cases {
		t.Run(tt.versionSpecifier, func(t *testing.T) {
			expected := make([]JavaHome, len(tt.expectedResultKeys))
			for i, key := range tt.expectedResultKeys {
				expected[i] = jdks[key]
			}
			slices.SortFunc(expected, fullJavaHomeCompare)

			assert.Equal(t, expected, cache.FindJdks(tt.versionSpecifier))
		})
	}
}
