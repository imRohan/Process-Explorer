package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/imRohan/go-ps"
	"github.com/satori/go.uuid"
  "github.com/kardianos/service"
	"net"
  //"os"
	"os/user"
	"strings"
	"time"
)

var autoRefresh bool = false
var showDefaultProcesses bool = false
var logger service.Logger

type Process struct {
	Name      string    `json:"name"`
	CreatedAt string    `json:"createdAt"`
	Pid       int       `json:"pid"`
	Ppid      int       `json:"ppid"`
	Uuid      uuid.UUID `json:"uuid"`
}

type outputStruct struct {
  Username   string    `json:"username"`
  MacAddress string    `json:"macAddress"`
  Processes  []Process `json:"processes"`
}


func getProcesses() (output []Process, err error) {
	defaultProcesses := showDefaultProcesses

	processes, err := ps.Processes()
	if err != nil {
		return output, errors.New("Could Not Get Processes")
	}

	for _, process := range processes {
		createdAt := process.CreationTime()
		name := strings.Split(process.Executable(), ".exe")[0]
		pid := process.Pid()
		ppid := process.PPid()
		uuid := uuid.NewV1()
		if !defaultProcesses || defaultProcesses && createdAt.Year() != -0001 {
      currentProcess := Process{name, createdAt.String(), pid, ppid, uuid}
      output = append(output, currentProcess)
		}
	}

	return output, nil
}

func renderJSON(returnedProcesses []Process) error {
	currentUser, err := getCurrentUser()
	if err != nil {
		return err
	}
	macAddress, err := getMacAddress()
	if err != nil {
		return err
	}


	outputPackage := outputStruct{currentUser, macAddress, returnedProcesses}

	json, err := json.Marshal(outputPackage)
	if err != nil {
		return errors.New("Cannot Generate JSON")
	}

	fmt.Println(string(json))
	return nil
}

func getCurrentUser() (username string, err error) {
	user, err := user.Current()

	if err != nil {
		return username, errors.New("Cannot Get User Details")
	}

	username = fmt.Sprintf("%s", user.Username)
	return username, nil
}

func getMacAddress() (macAddress string, err error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return macAddress, errors.New("Cannot Get Interfaces")
	}

	for _, singleInterface := range interfaces {
		hardwareName := singleInterface.Name
		if hardwareName == "Ethernet" || hardwareName == "ethernet" {
			macAddress = singleInterface.HardwareAddr.String()
		}
	}

	return macAddress, nil
}

func initAutoRefresh() {
	for range time.Tick(time.Second * 10) {
		status := autoRefresh
		if status {
			fmt.Println("Refresh Processes")
			returnedProcesses, err := getProcesses()
			if err != nil {
				fmt.Println(err)
			} else {
        renderJSON(returnedProcesses)
			}
		}
	}
}

func main() {
}
