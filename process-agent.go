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

// Build options, refresh time is in seconds
var options = struct {
	autoRefresh          bool
	hideDefaultProcesses bool
	refreshTime          time.Duration
}{true, true, 10}

// Struct to hold User Details
type UserDetails struct {
	Name string `json:"userName"`
	Mac  string `json:"macAddress"`
}

// Struct to hold process information
type Process struct {
	Name      string    `json:"name"`
	CreatedAt string    `json:"createdAt"`
	Pid       int       `json:"pid"`
	Ppid      int       `json:"ppid"`
	Uuid      uuid.UUID `json:"uuid"`
}

// Struct for renderJson params
type RenderParams struct {
	UserDetails UserDetails
	Processes   []Process
	Callback    func([]byte)
}

type program struct{}

/* pollProcesses
   Continually gets a list of process at a fixed duration (see global vars above)
   and forwards the returned list of processes to a callback
*/
func pollProcesses(user UserDetails) {
	for range time.Tick(time.Second * options.refreshTime) {
		if options.autoRefresh {
			returnedProcesses, err := getProcesses()
			if err != nil {
				log.Println(err)
			} else {
				log.Println(len(returnedProcesses), "running processes \r\n")
				renderJSON(RenderParams{user, returnedProcesses, exportData})
			}
		}
	}
}

/* getProcesses
   Gets all running processes and returns an array of
   Process objects
*/
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

/* renderJSON
   Requires a UserDetails object as well as an array of Processes
   will return valid JSON of all running processes
*/
func renderJSON(params RenderParams) error {
  // Struct which is holds the final JSON
  var outputPackage = struct {
    UserDetails UserDetails `json:"userDetails"`
    Processes   []Process   `json:"processes"`
  }{params.UserDetails, params.Processes}

	json, err := json.Marshal(outputPackage)
	if err != nil {
		return errors.New("Cannot Generate JSON \r\n")
	}

	go params.Callback(json)
	return nil
}

/* exportData
   Placeholder for what to do with the built json...POST?
*/
func exportData(json []byte) {
  jsonString := string(json)
  log.Println("\r\n Returned JSON: " + jsonString + "\r\n")
}

/* getCurrentUser
   Get the username of the current user.
   Note: If run as a service, the current user will always be NT AUTHORITY\SYSTEM.
         However, if run as a standalone application, this will return the
         current logged in user
*/
func getCurrentUser() (username string, err error) {
	user, err := user.Current()

	if err != nil {
		return username, errors.New("Cannot Get User Details \r\n")
	}

	username = fmt.Sprintf("%s", user.Username)
	return username, nil
}

/* getMacAddress
   Returns the mac accress of the machine, assuming the hardware name is a variant of
   ethernet.
*/
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


/* Start
   Grabs the current user details and initializes pollProcesses
*/
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

/* Stop
   Sets the auto refresh flag to false and terminates the application
*/
func (p *program) Stop(s service.Service) error {
	options.autoRefresh = false
	log.Println("Service Terminated")
	return nil
}

/* main
   Initializes the log & service and waits for arguments. If no args present, call Run()
*/
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
