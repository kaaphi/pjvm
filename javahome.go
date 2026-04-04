package main

import (
	"errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

var VersionPattern = regexp.MustCompile(`\d+([.-_]\d+)*`)

type VolumeFsSupplier func(p string) VolumeFsPathHandler

type VolumeFsPathHandler interface {
	FromVolumePath(path string) (string, error)
	ToVolumePath(path string) string
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

func (handler WindowsFsPathHandler) ToVolumePath(p string) string {
	return filepath.FromSlash(path.Join(handler.vol, p))
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

func versionMatches(javaHome string, versionMatcher string) bool {
	if versionMatcher == "" {
		return true
	}

	javaVersion := VersionPattern.FindString(path.Base(javaHome))
	if javaVersion == "" {
		return false
	}

	suffix, matches := strings.CutPrefix(javaVersion, versionMatcher)
	if matches {
		first_rune, _ := utf8.DecodeRuneInString(suffix)
		return !unicode.IsDigit(first_rune)
	} else {
		return false
	}
}

func findJavaVersions(paths []string, version string, volumeFsSupplier VolumeFsSupplier) []string {
	var estimated_capacity int
	if version == "" {
		estimated_capacity = 2
	} else {
		estimated_capacity = len(paths) * 2
	}
	var all_matches []string = make([]string, 0, estimated_capacity)

	for _, base_path := range paths {
		pathHandler := volumeFsSupplier(base_path)
		fs.WalkDir(pathHandler.RootFS(), pathHandler.ToVolumePath(base_path), func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if isJavaHome(pathHandler.RootFS(), path, d) {
				if versionMatches(path, version) {
					all_matches = append(all_matches, path)
				}
				return fs.SkipDir
			}

			return nil
		})
	}

	return all_matches
}
