package main

import (
	"simple-docker/utils"
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
		utils.InitCommand,
		utils.RunCommand,
		utils.LogCommand,
		utils.ListCommand,
		utils.ExecCommand,
		utils.CommitCommand,
		utils.StopCommand,
		utils.RemoveCommand,
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
