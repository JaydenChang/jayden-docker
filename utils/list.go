package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"simple-docker/container"
	"text/tabwriter"

	"github.com/sirupsen/logrus"
)

func ListContainers() {
	// get the path that store the info of the container
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, "")
	dirURL = dirURL[:len(dirURL)-1]
	// read all the files in the directory
	files, err := ioutil.ReadDir(dirURL)
	if err != nil {
		logrus.Errorf("read dir %s error %v", dirURL, err)
		return
	}
	var containers []*container.ContainerInfo
	for _, file := range files {
		tmpContainer, err := getContainerInfo(file)
		if err != nil {
			logrus.Errorf("get container info error %v", err)
			continue
		}
		containers = append(containers, tmpContainer)
	}
	// use tabwriter.NewWriter to print the containerInfo
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	fmt.Fprintf(w, "ID\tNAME\tPID\tSTATUS\tCOMMAND\tCREATED\n")
	for _, item := range containers {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			item.Id, item.Name, item.Pid, item.Status, item.Command, item.CreatedTime)
	}
	// refresh stdout
	if err := w.Flush(); err != nil {
		logrus.Errorf("flush stdout error %v", err)
		return
	}
}

func getContainerInfo(file os.FileInfo) (*container.ContainerInfo, error) {
	containerName := file.Name()
	// create the absolute path
	configFileDir := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	configFileDir = configFileDir + "/" + container.ConfigName
	fmt.Println(configFileDir)
	// read config.json
	content, err := ioutil.ReadFile(configFileDir)
	if err != nil {
		logrus.Errorf("read file %s error %v", configFileDir, err)
		return nil, err
	}
	var containerInfo container.ContainerInfo
	// turn json to containerInfo
	if err := json.Unmarshal(content, &containerInfo); err != nil {
		logrus.Errorf("unmarshal json error %v", err)
		return nil, err
	}
	return &containerInfo, nil
}
