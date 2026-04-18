package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v3"
)

func FakeJava(ctx context.Context, cmd *cli.Command) error {
	var versionPath string

	exec := cmd.String("exe")
	if exec == "" {
		execPath, err := os.Executable()
		if err != nil {
			return err
		}
		exec = filepath.Base(execPath)
		versionPath = filepath.Dir(filepath.Dir(execPath))
	} else {
		workingDir, err := os.Getwd()
		if err != nil {
			return err
		}
		versionPath = workingDir
	}

	if cmd.Bool("version") {
		switch exec {
		case "java.exe":
			versionFile := filepath.Join(versionPath, "version.txt")

			data, err := os.ReadFile(versionFile)
			if err != nil {
				return fmt.Errorf("error reading version text file: %w", err)
			}

			content := string(data)
			fmt.Print(content)
		case "javac.exe":
			fmt.Println("This is the compiler")
		default:
			return fmt.Errorf("not a java: <%s>", exec)
		}
	} else {
		fmt.Printf("Fake %s in %s\n", exec, filepath.Base(versionPath))
	}

	return nil
}

func Install(ctx context.Context, cmd *cli.Command) error {
	dir, err := filepath.Abs(cmd.Args().First())
	if err != nil {
		return err
	}

	exec, err := os.Executable()
	if err != nil {
		return err
	}
	execSrc, err := os.Open(exec)
	if err != nil {
		return err
	}
	defer closeFile(exec, execSrc)

	return filepath.WalkDir(dir, func(filePath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if isFakeJavaHome(filePath, d) {
			fmt.Printf("Found fake java path <%s>\n", filePath)
			binDir := filepath.Join(filePath, "bin")
			err := os.MkdirAll(binDir, os.ModePerm)
			if err != nil {
				return err
			}
			if err = copyFakeJava(execSrc, binDir, "java.exe", "javac.exe"); err != nil {
				return err
			}

			return fs.SkipDir
		}
		return nil
	})
}

func copyFakeJava(src *os.File, dstDir string, dstNames ...string) error {
	for _, dstName := range dstNames {
		dstPath := filepath.Join(dstDir, dstName)
		dst, err := os.Create(dstPath)
		defer closeFile(dstPath, dst)
		if err != nil {
			return err
		}

		_, err = io.Copy(dst, src)
		if err != nil {
			return err
		}
		if _, err := src.Seek(0, io.SeekStart); err != nil {
			return err
		}
	}
	return nil
}

func closeFile(path string, file *os.File) {
	if err := file.Close(); err != nil {
		log.Fatalf("Failed to close %s in defer: %s", path, err)
	}
}

func isFakeJavaHome(filePath string, d fs.DirEntry) bool {
	if !d.IsDir() {
		return false
	}

	_, error := os.Stat(filepath.Join(filePath, "version.txt"))

	return !errors.Is(error, fs.ErrNotExist)
}

func main() {
	cmd := &cli.Command{
		Name:  "fakejava",
		Usage: "pretend to be java",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "exe",
				Usage: "Override the exe name",
			},
			&cli.BoolFlag{
				Name:  "version",
				Usage: "Show the version",
			},
		},
		Action: FakeJava,
		Commands: []*cli.Command{
			{
				Name:   "install",
				Usage:  "install into all the subdirectories of the provided path",
				Action: Install,
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
