package main

import (
	"os"
	"simple-docker/utils"

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
		utils.ExecCommand,
		utils.LogCommand,
		utils.ListCommand,
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
