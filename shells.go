package main

import (
	_ "embed"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/hairyhenderson/go-which"
)

const ShellCommandStartMarker = "@@@START_SHELL@@@"
const ShellCommandEndMarker = "@@@END_SHELL@@@"
const ShellPjvmExecPlaceholder = "@@@PJVM_EXEC@@@"

type Shell interface {
	// Returns a slice of stings containing the correct shell commands to evaluate in order to update JAVA_HOME and PATH environment variables for the given javaHome location.
	SetJavaHome(javaHome string) ([]string, error)
	// Returns a string containing the correct shell commands to evaluate in order to set up the pjvm tool in the shell.
	EnvCommand() string
	// Converts the given os-native path to the correct type of path string to use in the shell.
	ConvertPath(p string) string
}

type GenerateShellCommands func(shell Shell) ([]string, error)

func GetShell(shellType string) (Shell, error) {
	switch shellType {
	case "PowerShell":
		return PowerShell{}, nil
	case "GitBash":
		return GitBash{}, nil
	default:
		return nil, fmt.Errorf("Unsupported shell <%s>", shellType)
	}
}

func ShellCommands(shellType string, commands GenerateShellCommands) error {
	shell, err := GetShell(shellType)
	if err != nil {
		return err
	}

	cmds, err := commands(shell)
	if err != nil {
		return err
	}

	fmt.Println(ShellCommandStartMarker)
	for _, cmd := range cmds {
		fmt.Println(cmd)
	}
	fmt.Println(ShellCommandEndMarker)

	return nil
}

func ExecutableDir() (string, error) {
	exec, err := os.Executable()
	if err != nil {
		return "", err
	}

	exec, err = filepath.EvalSymlinks(exec)
	if err != nil {
		return "", err
	}

	return filepath.Dir(exec), nil

}

type PowerShell struct{}

//go:embed powershell_install.ps1
var powershellInstallScript string

func (shell PowerShell) EnvCommand() string {
	return powershellInstallScript
}

func (shell PowerShell) SetJavaHome(javaHome string) ([]string, error) {
	pathEnv := UpdateJavaHomeInPath(javaHome)
	pathString := strings.Join(pathEnv, string(os.PathListSeparator))

	return []string{
		fmt.Sprintf(`$Env:JAVA_HOME="%s"`, javaHome),
		fmt.Sprintf(`$Env:PATH="%s"`, pathString),
	}, nil
}

func (shell PowerShell) ConvertPath(p string) string {
	return p
}

type GitBash struct{}

//go:embed gitbash_install.sh
var gitbashInstallScript string

func (shell GitBash) EnvCommand() string {
	return gitbashInstallScript
}

func (shell GitBash) SetJavaHome(javaHome string) ([]string, error) {
	pathEnv := UpdateJavaHomeInPath(javaHome)
	for i, p := range pathEnv {
		pathEnv[i] = shell.ConvertPath(p)
	}
	pathString := strings.Join(pathEnv, ":")

	return []string{
		fmt.Sprintf(`export JAVA_HOME="%s"`, javaHome),
		fmt.Sprintf(`export PATH="%s"`, pathString),
	}, nil
}

// Converts the path from a windows-style path string to the types of paths used by Git Bash (similar to what the "cygpath -u" command does)
func (shell GitBash) ConvertPath(p string) string {
	rawV := filepath.VolumeName(p)
	v, isDrive := strings.CutSuffix(rawV, ":")
	if isDrive {
		v = `/` + strings.ToLower(v)
		rawV = rawV + `\`
	} else {
		v = filepath.ToSlash(rawV)
	}

	rel, err := filepath.Rel(rawV, p)
	if err != nil {
		log.Fatalf("Failed to convert path: %s", err)
	}
	rel = filepath.ToSlash(rel)
	return v + `/` + rel
}

// An interface for manipulating file system paths
type FilePaths interface {
	Join(elem ...string) string
	Dir(path string) string
	Clean(path string) string
}

// Manipulate file system paths using the path/filepath system library (uses os-native semantics)
type SystemFilePaths struct{}

func (SystemFilePaths) Join(elem ...string) string {
	return filepath.Join(elem...)
}

func (SystemFilePaths) Dir(path string) string {
	return filepath.Dir(path)
}

func (SystemFilePaths) Clean(path string) string {
	return filepath.Clean(path)
}

// Manipulate file system paths using the path system library (uses Go-style forward slash paths)
type GoFilePaths struct{}

func (GoFilePaths) Join(elem ...string) string {
	return path.Join(elem...)
}

func (GoFilePaths) Dir(p string) string {
	return path.Dir(p)
}

func (GoFilePaths) Clean(p string) string {
	return path.Clean(p)
}

// Returns a slice containing updated paths for the PATH environment variable such that the given javaHome is on the path and any other existing java paths have been removed
func UpdateJavaHomeInPath(javaHome string) []string {
	rawPath := os.Getenv("PATH")
	pathSep := string(os.PathListSeparator)
	newEnvPath := updateJavaHomeInPath(strings.Split(rawPath, pathSep), which.All("java.exe"), javaHome, SystemFilePaths{})

	return newEnvPath
}

// Returns a slice containing the same paths as envPath except all paths containing the javasOnPath paths have been removed and newJavaHome has been added
func updateJavaHomeInPath(envPath []string, javasOnPath []string, newJavaHome string, filepaths FilePaths) []string {
	newEnvPath := make([]string, 1, len(envPath)-len(javasOnPath)+1)
	newEnvPath[0] = filepaths.Join(newJavaHome, "bin")

	for _, p := range envPath {
		for _, j := range javasOnPath {
			if filepaths.Dir(j) != filepaths.Clean(p) {
				newEnvPath = append(newEnvPath, p)
			}
		}
	}

	return newEnvPath
}
