package utils

import (
	"errors"
	"fmt"
	"os"
	"simple-docker/cgroups/subsystem"
	"simple-docker/container"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

// command.go
// docker init, but cannot be used by user
var InitCommand = cli.Command{
	Name:  "init",
	Usage: "init a container",
	Action: func(context *cli.Context) error {
		logrus.Infof("Start initiating...")
		return container.InitProcess()
	},
}

var RunCommand = cli.Command{
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
		}, &cli.StringFlag{
			Name:  "v",
			Usage: "volume",
		}, &cli.BoolFlag{
			Name:  "d",
			Usage: "detach container",
		}, &cli.StringFlag{
			Name:  "cpuset",
			Usage: "limit the cpuset",
		}, &cli.StringFlag{
			Name:  "name",
			Usage: "container name",
		}, &cli.StringSliceFlag{
			Name:  "e",
			Usage: "set environment",
		},
	},
	Action: func(context *cli.Context) error {
		args := context.Args()
		if len(args) <= 0 {
			return errors.New("run what?")
		}

		// 转化 cli.Args 为 []string
		cmdArray := make([]string, len(args)) // command
		copy(cmdArray, args)

		// check whether type `-it`
		tty := context.Bool("it")   // presudo terminal
		detach := context.Bool("d") // detach container

		if tty && detach {
			return fmt.Errorf("it and d paramter cannot both privided")
		}

		// get the resource config
		resourceConfig := subsystem.ResourceConfig{
			MemoryLimit: context.String("m"),
			CPUShare:    context.String("cpushare"),
			CPUSet:      context.String("cpu"),
		}
		volume := context.String("v")
		containerName := context.String("name")
		envSlice := context.StringSlice("e")
		imageName := cmdArray[0]
		cmdArray = cmdArray[1:]
		Run(tty, cmdArray, &resourceConfig, volume, containerName, imageName,envSlice)

		return nil
	},
}

var ExecCommand = cli.Command{
	Name:  "exec",
	Usage: "exec a command into container",
	Action: func(context *cli.Context) error {
		if os.Getenv(ENV_EXEC_PID) != "" {
			logrus.Infof("pid callback pid %v", os.Getgid())
			return nil
		}
		if len(context.Args()) < 2 {
			return fmt.Errorf("missing container name or command")
		}
		containerName := context.Args()[0]
		cmdArray := make([]string, len(context.Args())-1)
		for i, v := range context.Args().Tail() {
			cmdArray[i] = v
		}
		fmt.Println(context.Args(), cmdArray)
		ExecContainer(containerName, cmdArray)
		return nil
	},
}

var CommitCommand = cli.Command{
	Name:  "commit",
	Usage: "commit a container into image",
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("missing container name")
		}
		containerName := context.Args()[0]
		imageName := context.Args()[1]
		// commitContainer(containerName)
		commitContainer(containerName, imageName)
		return nil
	},
}

var ListCommand = cli.Command{
	Name:  "ps",
	Usage: "list all the containers",
	Action: func(context *cli.Context) error {
		ListContainers()
		return nil
	},
}

var LogCommand = cli.Command{
	Name:  "logs",
	Usage: "print logs of a container",
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("missing container name")
		}
		contianerName := context.Args()[0]
		logContainer(contianerName)
		return nil
	},
}

var StopCommand = cli.Command{
	Name:  "stop",
	Usage: "stop a container",
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("missing container name")
		}
		containerName := context.Args()[0]
		stopContainer(containerName)
		return nil
	},
}

var RemoveCommand = cli.Command{
	Name:  "rm",
	Usage: "remove a container",
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("missing container name")
		}
		containerName := context.Args()[0]
		removeContainer(containerName)
		return nil
	},
}
