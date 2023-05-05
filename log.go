package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"simple-docker/container"

	"github.com/sirupsen/logrus"
)

func logContainer(containerName string) {
	// get the log path
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	logFileLocation := dirURL + "/" + container.ContainerLogFile
	// open log file
	file, err := os.Open(logFileLocation)
	if err != nil {
		logrus.Errorf("log container open file %s error: %v", logFileLocation, err)
		return
	}
	defer file.Close()
	// read log file content
	content, err := ioutil.ReadAll(file)
	if err != nil {
		logrus.Errorf("log container read file %s error: %v", logFileLocation, err)
		return
	}
	// use Fprint to transfer content to stdout
	fmt.Fprint(os.Stdout, string(content))
}
