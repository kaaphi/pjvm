package main

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
)

const ShellCommandStartMarker = "@@@START_SHELL@@@"
const ShellCommandEndMarker = "@@@END_SHELL@@@"
const ShellPjvmExecPlaceholder = "@@@PJVM_EXEC@@@"

type Shell interface {
	SetJavaHome(javaHome string) ([]string, error)
	EnvCommand() string

	// WriteOutput(msg string, args ...string) string
	// WriteError(msg string, args ...string) string
	ConvertPath(p string) string
}

type GenerateShellCommands func(shell Shell) ([]string, error)

func GetShell(shellType string) (Shell, error) {
	switch shellType {
	case "PowerShell":
		return Powershell{}, nil
	case "GitBash":
		return GitBash{}, nil
	default:
		return nil, fmt.Errorf("Unsupported shell %s", shellType)
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

func ScriptRootDir() (string, error) {
	script_home := os.Getenv("PJVM_SCRIPT_HOME")

	if script_home == "" {
		return ExecutableDir()
	} else {
		return script_home, nil
	}
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

type Powershell struct{}

//go:embed powershell_install.ps1
var powershellInstallScript string

func (shell Powershell) EnvCommand() string {
	return powershellInstallScript
}

func (shell Powershell) SetJavaHome(javaHome string) ([]string, error) {
	script_home, err := ScriptRootDir()
	if err != nil {
		return nil, err
	}

	return []string{
		fmt.Sprintf(`$Env:JAVA_HOME="%s"`, javaHome),
		fmt.Sprintf(`$Env:PATH="%s\bin;$(%s\clean_java_path.ps1)"`, javaHome, script_home),
	}, nil
}

func (shell Powershell) ConvertPath(p string) string {
	return p
}

type GitBash struct{}

//go:embed gitbash_install.sh
var gitbashInstallScript string

func (shell GitBash) EnvCommand() string {
	return gitbashInstallScript
}

func (shell GitBash) SetJavaHome(javaHome string) ([]string, error) {
	script_home, err := ScriptRootDir()
	if err != nil {
		return nil, err
	}

	return []string{
		fmt.Sprintf(`PJVM_HOME="%s"`, shell.ConvertPath(script_home)),
		fmt.Sprintf(`export JAVA_HOME="%s"`, shell.ConvertPath(javaHome)),
		fmt.Sprintf(`export PATH="$JAVA_HOME/bin:%s$PJVM_HOME/clean_java_path.sh "$PATH"%s"`, "`", "`"),
	}, nil
}

func (shell GitBash) ConvertPath(p string) string {
	return fmt.Sprintf("`cygpath -u \"%s\"`", p)
}
