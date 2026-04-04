package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v3"
)

func main() {

	// fsys, path := WindowsFsProvider(`c:\tools\java`)
	// fmt.Println(path)
	// info, err := fsys.Stat(path)
	// fmt.Println(info)
	// fmt.Println(err)
	// fmt.Println(os.Executable())

	// os.Exit(0)

	cmd := &cli.Command{
		Name:  "list",
		Usage: "list all available java versions",
		Action: func(context.Context, *cli.Command) error {
			javaHomes := findJavaVersions([]string{`c:\tools\java`}, "", WindowsVolume)
			for _, javaHome := range javaHomes {
				fmt.Println(javaHome)
			}
			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
