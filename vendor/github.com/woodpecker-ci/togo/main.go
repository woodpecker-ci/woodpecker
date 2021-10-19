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
	app.Version = "dev"
	app.Author = "Woodpecker Authors"
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
