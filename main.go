package main

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

const usage = `Usage`

func main() {
	app := cli.NewApp()
	app.Name = "simple-docker"
	app.Usage = usage
	app.Commands = []cli.Command{
		runCommand,
		initCommand,
		commitCommand,
		listCommand,
	}
	app.Before = func(context *cli.Context) error {
		logrus.SetFormatter(&logrus.JSONFormatter{})
		logrus.SetOutput(os.Stdout)
		return nil
	}
	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}
