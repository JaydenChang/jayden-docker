package main

import (
	"errors"
	"simple-docker/cgroup/subsystem"
	"simple-docker/container"
	"simple-docker/dockerCommand"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

// command.go
// docker init, but cannot be used by user
var initCommand = cli.Command{
	Name:  "init",
	Usage: "init a container",
	Action: func(context *cli.Context) error {
		logrus.Infof("Start initiating...")
		return container.InitProcess()
	},
}

// docker run
var runCommand = cli.Command{
	Name:  "run",
	Usage: "Create a container",
	Flags: []cli.Flag{
		// integrate -i and -t for convenience
		&cli.BoolFlag{
			Name:  "it",
			Usage: "open an interactive tty(pseudo terminal)",
		},
		&cli.StringFlag{
			Name:  "m",
			Usage: "limit the memory",
		}, &cli.StringFlag{
			Name:  "cpu",
			Usage: "limit the cpu amount",
		}, &cli.StringFlag{
			Name:  "cpushare",
			Usage: "limit the cpu share",
		},
	},
	Action: func(context *cli.Context) error {
		args := context.Args()
		if len(args) == 0 {
			return errors.New("Run what?")
		}

		// 转化 cli.Args 为 []string
		containerCmd := make([]string, len(args)) // command
		copy(containerCmd, args)

		// check whether type `-it`
		tty := context.Bool("it") // presudo terminal

		// get the resource config
		resourceConfig := subsystem.ResourceConfig{
			MemoryLimit: context.String("m"),
			CPUShare:    context.String("cpushare"),
			CPUSet:      context.String("cpu"),
		}
		dockerCommand.Run(tty, containerCmd, &resourceConfig)

		return nil
	},
}
