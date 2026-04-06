package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/urfave/cli/v3"
)

func PjvmInstall(ctx context.Context, cmd *cli.Command) error {
	shell, err := GetShell(cmd.String("shell"))
	if err != nil {
		return err
	}

	install := shell.InstallCommand()

	exec, err := os.Executable()

	if err != nil {
		return err
	}

	for _, line := range strings.Split(strings.ReplaceAll(install, ShellPjvmExecPlaceholder, shell.ConvertPath(exec)), "\n") {
		fmt.Println(line)
	}

	return nil
}

func PjvmUse(ctx context.Context, cmd *cli.Command) error {
	version_specifier := cmd.StringArg("version")

	javaHomes, err := findJavaVersions([]string{`c:\tools\java`}, version_specifier, WindowsVolume)

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

func main() {
	//fmt.Println(ExecutableDir())

	cmd := &cli.Command{
		Usage: "A JDK version manager",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "shell",
				Usage:    "the shell to use for the output",
				Hidden:   true,
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "list",
				Usage: "list all available java versions",
				Action: func(context.Context, *cli.Command) error {
					javaHomes, err := findJavaVersions([]string{`c:\tools\java`}, "", WindowsVolume)

					if err != nil {
						return err
					}

					for _, javaHome := range javaHomes {
						fmt.Println(javaHome)
					}
					return nil
				},
			},
			{
				Name:  "use",
				Usage: "use the specific Java version",
				Arguments: []cli.Argument{
					&cli.StringArg{
						Name:      "version",
						UsageText: "version_specifier",
					},
				},
				Action: PjvmUse,
			},
			{
				Name:  "install",
				Usage: "install into the given shell",
				Action: PjvmInstall,
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
