package main

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"github.com/urfave/cli/v3"
)

type PjvmConfg struct {
	BasePaths []string
}

func PjvmEnv(ctx context.Context, cmd *cli.Command) error {
	shell, err := GetShell(cmd.String("shell"))
	if err != nil {
		return err
	}

	env := shell.EnvCommand()

	exec, err := os.Executable()

	if err != nil {
		return err
	}

	for _, line := range strings.Split(strings.ReplaceAll(env, ShellPjvmExecPlaceholder, shell.ConvertPath(exec)), "\n") {
		fmt.Println(line)
	}

	return nil
}

func PjvmUse(ctx context.Context, cmd *cli.Command) error {
	config, err := loadConfig(cmd)
	if err != nil {
		return err
	}

	version_specifier := cmd.StringArg("version")

	javaHomes, err := findJavaVersions(config.BasePaths, version_specifier, WindowsVolume)

	if err != nil {
		return err
	}

	numMatches := len(javaHomes)
	if numMatches == 0 {
		return fmt.Errorf("No Java versions found matching <%s>", version_specifier)
	} else if numMatches > 1 {
		fmt.Printf("Found %d version matches:\n", numMatches)
		for _, javaHome := range javaHomes {
			fmt.Println(javaHome)
		}
		return nil
	}

	err = ShellCommands(cmd.String("shell"), func(shell Shell) ([]string, error) {
		return shell.SetJavaHome(javaHomes[0].JavaHomePath)
	})
	if err != nil {
		return err
	}
	fmt.Printf("Using Java version %s from %s\n", javaHomes[0].JavaVersion, javaHomes[0].JavaHomePath)

	return nil
}

func PjvmList(ctx context.Context, cmd *cli.Command) error {
	config, err := loadConfig(cmd)

	if err != nil {
		return err
	}

	javaHomes, err := findJavaVersions(config.BasePaths, "", WindowsVolume)

	if err != nil {
		return err
	}

	for _, javaHome := range javaHomes {
		fmt.Printf("%-12s %s\n", javaHome.JavaVersion, javaHome.JavaHomePath)
	}
	return nil
}

func loadConfig(cmd *cli.Command) (PjvmConfg, error) {
	var cfg PjvmConfg
	configFile := cmd.String("config")

	if configFile == "" {
		configFile = os.Getenv("PJVM_CONFIG")
	}

	if configFile == "" {
		userHome, err := os.UserHomeDir()

		if err != nil {
			return cfg, err
		}

		configFile = filepath.Join(userHome, ".pjvm")
	}

	content, err := os.ReadFile(configFile)
	if err != nil {
		return cfg, err
	}

	if err := toml.Unmarshal(content, &cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}

//go:embed version.txt
var version string

func main() {
	cmd := &cli.Command{
		Usage:   "A JDK version manager",
		Version: version,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:   "shell",
				Usage:  "the shell to use for the output",
				Hidden: true,
			},
			&cli.StringFlag{
				Name:  "config",
				Usage: "the location of the config file",
			},
		},
		Commands: []*cli.Command{
			{
				Name:   "list",
				Usage:  "list all available java versions",
				Action: PjvmList,
			},
			{
				Name:  "use",
				Usage: "use the specified Java version",
				Arguments: []cli.Argument{
					&cli.StringArg{
						Name:      "version",
						UsageText: "version_specifier",
					},
				},
				Action: PjvmUse,
			},
			{
				Name:   "env",
				Usage:  "output env to evaluate to setup for the current shell",
				Hidden: true,
				Action: PjvmEnv,
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
