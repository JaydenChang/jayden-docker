package main

import (
	"fmt"
	"simple-docker/cgroups/subsystem"
	"simple-docker/container"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var runCommand = cli.Command{
	Name:  "run",
	Usage: "Create a container with namespace and cgroups limit",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "ti",
			Usage: "enable tty",
		},
		cli.StringFlag{
			Name:  "m",
			Usage: "memory limit",
		},
		cli.StringFlag{
			Name:  "cpushare",
			Usage: "cpushare limit",
		},
		cli.StringFlag{
			Name:  "cpuset",
			Usage: "cpuset limit",
		},
		cli.StringFlag{
			Name:  "v",
			Usage: "docker volume",
		},
		cli.BoolFlag{
			Name:  "d",
			Usage: "detach container",
		},
		cli.StringFlag{
			Name:  "name",
			Usage: "container name",
		},
		cli.StringSliceFlag{
			Name:  "e",
			Usage: "docker env",
		},
		cli.StringFlag{
			Name:  "net",
			Usage: "container network",
		},
		cli.StringSliceFlag{
			Name:  "p",
			Usage: "port mapping",
		},
	},
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("missing container args")
		}
		tty := context.Bool("ti")
		detach := context.Bool("d")
		volume := context.String("v")
		res := &subsystem.ResourceConfig{
			MemoryLimit: context.String("m"),
			CpuSet:      context.String("cpuset"),
			CpuShare:    context.String("cpushare"),
		}
		var cmdArray []string
		for _, arg := range context.Args() {
			cmdArray = append(cmdArray, arg)
		}
		if tty && detach {
			return fmt.Errorf("ti and d paramter can not both privided")
		}
		containerName := context.String("name")
		nets := context.String("net")
		imageName := context.Args().Get(0)
		envs := context.StringSlice("e")
		ports := context.StringSlice("p")

		Run(cmdArray, tty, res, containerName, imageName, volume, nets, envs, ports)
		return nil
	},
}

var initCommand = cli.Command{
	Name:  "init",
	Usage: "Init container process run user's process in container. Do not call it outside",
	Action: func(context *cli.Context) error {
		logrus.Infof("init come on")
		return container.RunContainerInitProcess()
	},
}

var commitCommand = cli.Command{
	Name:  "commit",
	Usage: "docker container process run user's process in container. Do not call it outside",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "c",
			Usage: "export image path",
		},
	},
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("missing container args")
		}
		imageName := context.Args().Get(0)
		imagePath := context.String("c")
		return container.CommitContainer(imageName, imagePath)
	},
}

var networkCommand = cli.Command{
	Name:  "network",
	Usage: "container network commands",
	Subcommands: []cli.Command{
		{
			Name:  "create",
			Usage: "create container network",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "driver",
					Usage: "network driver",
				},
				cli.StringFlag{
					Name:  "subnet",
					Usage: "network cidr",
				},
			},
			Action: func(context *cli.Context) error {
				if len(context.Args()) < 1 {
					return fmt.Errorf("missing network name")
				}
				err := network.Init()
				if err != nil {
					logrus.Errorf("network init failed, err: %v", err)
					return err
				}
				err = network.CreateNework(context.String("driver"), context.String("subnet", context.Args()[0]))
				if err == nil {
					return fmt.Errorf("create network err: %+v", err)
				}
				return nil
			},
		},
		{
			Name:  "list",
			Usage: "list container network",
			Action: func(context *cli.Context) error {
				err := network.Init()
				if err != nil {
					logrus.Errorf("network init failed, err: %v", err)
					return err
				}
				err = network.Init()
				if err != nil {
					logrus.Errorf("network init failed, err: %v", err)
					return err
				}
				err = network.DeleteNetwork(context.Args()[0])
				if err != nil {
					return fmt.Errorf("delete network err: %+v", err)
				}
				return nil
			},
		},
	},
}
