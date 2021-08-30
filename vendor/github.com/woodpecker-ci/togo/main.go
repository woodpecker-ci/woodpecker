package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "togo"
	app.Usage = "togo provides tools to convert files to go"
	app.Version = "1.0.0"
	app.Author = "bradrydzewski"
	app.Commands = []cli.Command{
		ddlCommand,
		sqlCommand,
		httpCommand,
		httptestCommand,
		tmplCommand,
		i18nCommand,
	}
	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
