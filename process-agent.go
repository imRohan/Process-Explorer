package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/imRohan/go-ps"
	"github.com/kardianos/service"
	"github.com/satori/go.uuid"
	"log"
	"net"
	"os"
	"os/user"
	"strings"
	"time"
)

var options = struct {
	autoRefresh          bool
	hideDefaultProcesses bool
	refreshTime          time.Duration
}{true, true, 10}

var logger service.Logger

type program struct{}

type UserDetails struct {
	Name string `json:"userName"`
	Mac  string `json:"macAddress"`
}

type Process struct {
	Name      string    `json:"name"`
	CreatedAt string    `json:"createdAt"`
	Pid       int       `json:"pid"`
	Ppid      int       `json:"ppid"`
	Uuid      uuid.UUID `json:"uuid"`
}

type outputStruct struct {
	UserDetails UserDetails `json:"UserDetails"`
	Processes   []Process   `json:"processes"`
}

func getProcesses() (output []Process, err error) {
	defaultProcesses := options.hideDefaultProcesses

	processes, err := ps.Processes()
	if err != nil {
		return output, errors.New("Could Not Get Processes \r\n")
	}

	for _, process := range processes {
		createdAt := process.CreationTime()
		name := strings.Split(process.Executable(), ".exe")[0]
		pid := process.Pid()
		ppid := process.PPid()
		uuid := uuid.NewV1()
		if !defaultProcesses || defaultProcesses && createdAt.Year() > 1 {
			currentProcess := Process{name, createdAt.String(), pid, ppid, uuid}
			output = append(output, currentProcess)
		}
	}

	return output, nil
}

func renderJSON(user UserDetails, returnedProcesses []Process) error {
	outputPackage := outputStruct{user, returnedProcesses}

	json, err := json.Marshal(outputPackage)
	if err != nil {
		return errors.New("Cannot Generate JSON \r\n")
	}

	log.Println(string(json))
	return nil
}

func getCurrentUser() (username string, err error) {
	user, err := user.Current()

	if err != nil {
		return username, errors.New("Cannot Get User Details \r\n")
	}

	username = fmt.Sprintf("%s", user.Username)
	return username, nil
}

func getMacAddress() (macAddress string, err error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return macAddress, errors.New("Cannot Get Interfaces \r\n")
	}

	for _, singleInterface := range interfaces {
		hardwareName := singleInterface.Name
		if hardwareName == "Ethernet" || hardwareName == "ethernet" {
			macAddress = singleInterface.HardwareAddr.String()
		}
	}

	return macAddress, nil
}

func pollProcesses(user UserDetails) {
	for range time.Tick(time.Second * options.refreshTime) {
		if options.autoRefresh {
			returnedProcesses, err := getProcesses()
			if err != nil {
				log.Println(err)
			} else {
				log.Println(len(returnedProcesses), "running processes \r\n")
				renderJSON(user, returnedProcesses)
			}
		}
	}
}

func (p *program) Start(s service.Service) error {
	options.autoRefresh = true

	name, err := getCurrentUser()
	if err != nil {
		return err
	}

	mac, err := getMacAddress()
	if err != nil {
		return err
	}

  user := UserDetails{name, mac}

	logString := fmt.Sprintf("Service Started for user '%s' \r\n"+
		"Options: [Auto Refresh: %v(%v seconds), Hide Defaults: %v] \r\n",
		user.Name, options.autoRefresh, options.refreshTime, options.hideDefaultProcesses)
	log.Println(logString)

	go pollProcesses(user)
	return nil
}

func (p *program) Stop(s service.Service) error {
	options.autoRefresh = false
	log.Println("Service Terminated")
	return nil
}

func main() {
	logFile, err := os.OpenFile("BioconnectProcess.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf(err.Error())
	}
	defer logFile.Close()

	log.SetOutput(logFile)

	svcConfig := &service.Config{
		Name:        "BioConnectProcessAgent",
		DisplayName: "BioConnect Process Agent",
		Description: "Collets data on the processes running",
	}

	prg := &program{}

	processService, err := service.New(prg, svcConfig)
	if err != nil {
		log.Println("Failed to create service: " + err.Error() + "\r\n")
		return
	}

	if len(os.Args) > 1 {
		var err error
		argument := os.Args[1]
		switch argument {
		case "install":
			err = processService.Install()
			if err != nil {
				log.Printf("Failed to install: " + err.Error() + "\r\n")
				return
			}
			log.Println("Service installed \r\n")
		case "start":
			err = processService.Run()
			if err != nil {
				log.Println("Failed to start service: " + err.Error() + "\r\n")
			}
		case "stop":
			err = processService.Stop()
			if err != nil {
				log.Println("Failed to stop service: " + err.Error() + "\r\n")
			}
		}
	} else {
		err = processService.Run()
		if err != nil {
			log.Println("Failed to start service: " + err.Error() + "\r\n")
		}
	}
}
