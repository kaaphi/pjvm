package main

import (
	"cmp"
	"encoding/gob"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"
)

const CacheFileName = ".pjvm_cache"

var VersionPattern = regexp.MustCompile(`\d+([.-_]\d+)*`)

type JavaHomeCache struct {
	Jdks []JavaHome
}

type JavaHomeCacheEncoder interface {
	StoreCache(context *PjvmContext, cache *JavaHomeCache) error
	LoadCache(context *PjvmContext) (*JavaHomeCache, error)
}

type FileSystemCacheEncoder struct{}

func fullJavaHomeCompare(a JavaHome, b JavaHome) int {
	if c := cmp.Compare(a.JavaVersion, b.JavaVersion); c == 0 {
		return cmp.Compare(a.JavaHomePath, b.JavaHomePath)
	} else {
		return c
	}
}

func versionJavaHomeCompare(a JavaHome, b JavaHome) int {
	return cmp.Compare(a.JavaVersion, b.JavaVersion)
}

func (cache *JavaHomeCache) SetJdks(jdks []JavaHome) {
	slices.SortFunc(jdks, fullJavaHomeCompare)
	cache.Jdks = jdks
}

func (cache *JavaHomeCache) AddJdk(jdk JavaHome) {
	i, found := slices.BinarySearchFunc(cache.Jdks, jdk, fullJavaHomeCompare)

	if !found {
		cache.Jdks = slices.Insert(cache.Jdks, i, jdk)
	}
}

func (cache *JavaHomeCache) FindJdks(versionSpecifier string) []JavaHome {
	i, _ := slices.BinarySearchFunc(cache.Jdks, JavaHome{JavaVersion: versionSpecifier}, versionJavaHomeCompare)

	result := make([]JavaHome, 0, 2)
	for ; i < len(cache.Jdks); i++ {
		suffix, matches := strings.CutPrefix(cache.Jdks[i].JavaVersion, versionSpecifier)
		if matches {
			first_rune, _ := utf8.DecodeRuneInString(suffix)
			if !unicode.IsDigit(first_rune) {
				result = append(result, cache.Jdks[i])
			} else {
				break
			}
		} else {
			break
		}
	}
	return result
}

func (FileSystemCacheEncoder) StoreCache(context *PjvmContext, cache *JavaHomeCache) error {
	cacheFile := filepath.Join(context.config.ConfigPath, CacheFileName)

	file, err := os.Create(cacheFile)
	if err != nil {
		return err
	}

	defer func() {
		if err := file.Close(); err != nil {
			log.Fatalf("Failed to close %s after storing cache: %s", cacheFile, err)
		}
	}()

	// Create the encoder
	encoder := gob.NewEncoder(file)

	// Encode the object
	if err := encoder.Encode(cache); err != nil {
		return fmt.Errorf("failed to encode cache: %w", err)
	}

	return nil
}

func (FileSystemCacheEncoder) LoadCache(context *PjvmContext) (*JavaHomeCache, error) {
	cacheFile := filepath.Join(context.config.ConfigPath, CacheFileName)

	file, err := os.Open(cacheFile)
	if err != nil {
		if os.IsNotExist(err) {
			return &JavaHomeCache{}, nil
		}
		return nil, fmt.Errorf("could not open file %s: %w", cacheFile, err)
	}

	defer func() {
		if err := file.Close(); err != nil {
			log.Fatalf("Failed to close %s after loading cache: %s", cacheFile, err)
		}
	}()

	decoder := gob.NewDecoder(file)

	cache := JavaHomeCache{}
	if err := decoder.Decode(&cache); err != nil {
		return nil, fmt.Errorf("failed to decode cache from file %s: %w", cacheFile, err)
	}

	return &cache, nil
}

type JavaHome struct {
	JavaHomePath string
	JavaVersion  string
}

type VolumeFsSupplier func(p string) VolumeFsPathHandler

type VolumeFsPathHandler interface {
	// Convert the given platform-specific path that might have a specific volume (e.g. C:\my\path) to a slash path
	// relative to the volume (e.g. my/path) that can be used to reference the path in the RootFS.
	FromVolumePath(path string) (string, error)
	// Convert a slash path to a platform-specific path that includes the volume
	ToVolumePath(path string) (string, error)
	// The file system
	RootFS() fs.StatFS
}

type WindowsFsPathHandler struct {
	vol   string
	volFs fs.StatFS
}

func (handler WindowsFsPathHandler) FromVolumePath(p string) (string, error) {
	rel, err := filepath.Rel(handler.vol+`\`, p)
	if err != nil {
		return "", err
	}
	return filepath.ToSlash(rel), nil
}

func (handler WindowsFsPathHandler) ToVolumePath(p string) (string, error) {
	return filepath.FromSlash(path.Join(handler.vol, p)), nil
}

func (handler WindowsFsPathHandler) RootFS() fs.StatFS {
	return handler.volFs
}

func WindowsVolume(path string) VolumeFsPathHandler {
	vol := filepath.VolumeName(path)
	fsys, _ := os.DirFS(vol).(fs.StatFS)
	handler := WindowsFsPathHandler{vol: vol, volFs: fsys}
	return handler
}

func isJavaHome(fsys fs.StatFS, filePath string, d fs.DirEntry) bool {
	if !d.IsDir() {
		return false
	}

	_, error := fsys.Stat(path.Join(filePath, "bin", "javac.exe"))

	return !errors.Is(error, fs.ErrNotExist)
}

func versionMatches(javaHome string, versionMatcher string) (string, bool) {
	javaVersion := VersionPattern.FindString(path.Base(javaHome))
	if javaVersion == "" {
		return "", false
	}

	if versionMatcher == "" {
		return javaVersion, true
	}

	suffix, matches := strings.CutPrefix(javaVersion, versionMatcher)
	if matches {
		first_rune, _ := utf8.DecodeRuneInString(suffix)
		return javaVersion, !unicode.IsDigit(first_rune)
	} else {
		return "", false
	}
}

func FindAllJdks(context PjvmContext) ([]JavaHome, error) {
	matches, err := findJdksInPaths(context.config.BasePaths, "", context.fileSystemSupplier)
	if err != nil {
		return nil, err
	}

	for _, m := range matches {
		context.cache.AddJdk(m)
	}
	if len(matches) > 0 {
		if err := context.StoreCache(); err != nil {
			log.Fatalf("Failed to store cache: %s", err)
		}
	}

	return context.cache.Jdks, nil
}

func FindJdks(context PjvmContext, versionSpecifier string) ([]JavaHome, error) {
	matches := context.cache.FindJdks(versionSpecifier)

	if len(matches) == 0 {
		matches, err := findJdksInPaths(context.config.BasePaths, versionSpecifier, context.fileSystemSupplier)
		if err != nil {
			return nil, err
		}

		for _, m := range matches {
			context.cache.AddJdk(m)
		}
		if len(matches) > 0 {
			if err := context.StoreCache(); err != nil {
				log.Fatalf("Failed to store cache: %s", err)
			}
		}

		return matches, nil
	} else {
		return matches, nil
	}
}

func findJdksInPaths(paths []string, version string, volumeFsSupplier VolumeFsSupplier) ([]JavaHome, error) {
	var estimatedCapacity int
	if version == "" {
		estimatedCapacity = 2
	} else {
		estimatedCapacity = len(paths) * 2
	}
	allMatches := make([]JavaHome, 0, estimatedCapacity)

	for _, basePath := range paths {
		pathHandler := volumeFsSupplier(basePath)
		basePath, err := pathHandler.FromVolumePath(basePath)

		if err != nil {
			return nil, err
		}

		err = fs.WalkDir(pathHandler.RootFS(), basePath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if isJavaHome(pathHandler.RootFS(), path, d) {
				if javaVersion, matches := versionMatches(path, version); matches {
					volumePath, err := pathHandler.ToVolumePath(path)

					if err != nil {
						return err
					}

					allMatches = append(allMatches, JavaHome{JavaHomePath: volumePath, JavaVersion: javaVersion})
				}
				return fs.SkipDir
			}

			return nil
		})

		if err != nil {
			return nil, err
		}
	}

	return allMatches, nil
}
