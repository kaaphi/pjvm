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

type PjvmConfig struct {
	BasePaths  []string
	ConfigPath string
}

type PjvmContext struct {
	config             PjvmConfig
	fileSystemSupplier VolumeFsSupplier
	cache              JavaHomeCache
	cacheEncoder       JavaHomeCacheEncoder
}

func (context *PjvmContext) StoreCache() error {
	return context.cacheEncoder.StoreCache(context, &context.cache)
}

func (context *PjvmContext) LoadCache() (*JavaHomeCache, error) {
	cache, err := context.cacheEncoder.LoadCache(context)
	if err == nil {
		if cache == nil {
			return nil, fmt.Errorf("encoder loaded nil cache")
		}
		context.cache = *cache
	}
	return cache, err
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

func PjvmUse(_ context.Context, cmd *cli.Command) error {
	context, err := loadContext(cmd)
	if err != nil {
		return err
	}

	versionSpecifier := cmd.StringArg("version")
	javaHomes, err := FindJdks(context, versionSpecifier)
	if err != nil {
		return err
	}

	numMatches := len(javaHomes)
	if numMatches == 0 {
		return fmt.Errorf("no Java versions found matching <%s>", versionSpecifier)
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
	context, err := loadContext(cmd)
	if err != nil {
		return err
	}

	javaHomes, err := FindAllJdks(context)
	if err != nil {
		return err
	}

	for _, javaHome := range javaHomes {
		fmt.Printf("%-12s %s\n", javaHome.JavaVersion, javaHome.JavaHomePath)
	}
	return nil
}

func loadContext(cmd *cli.Command) (PjvmContext, error) {
	context := PjvmContext{
		fileSystemSupplier: WindowsVolume,
		cacheEncoder:       FileSystemCacheEncoder{},
	}
	cfg, err := loadConfig(cmd)
	if err != nil {
		return context, fmt.Errorf("failed to load config: %w", err)
	}

	context.config = cfg

	if _, err := context.LoadCache(); err != nil {
		return context, fmt.Errorf("failed to load cache: %w", err)
	}

	return context, nil
}

func loadConfig(cmd *cli.Command) (PjvmConfig, error) {
	var cfg PjvmConfig
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

	if cfg.ConfigPath == "" {
		cfg.ConfigPath = filepath.Dir(configFile)
	}

	for i, p := range cfg.BasePaths {
		cfg.BasePaths[i], err = filepath.Abs(p)
		if err != nil {
			return cfg, fmt.Errorf("failed to make base path <%s> absolute: %w", p, err)
		}
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
