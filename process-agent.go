package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/imRohan/go-ps"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/satori/go.uuid"
	"net"
	"os/user"
	"strconv"
	"strings"
	"time"
)

type Process struct {
  Name      string    `json:"name"`
  CreatedAt string    `json:"createdAt"`
  Pid       int       `json:"pid"`
  Ppid      int       `json:"ppid"`
  Uuid      uuid.UUID `json:"uuid"`
}

var autoRefresh bool = false
var processWindow, searchField *walk.TextEdit
var autoRefreshCheckbox, toggleDefaultsCheckbox *walk.CheckBox
var searchFieldString string
var showDefaultProcesses bool = false

func getProcesses() (output []Process, err error) {
	defaultProcesses := showDefaultProcesses
	searchString := searchFieldString

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
			if len(searchString) == 0 || len(searchString) != 0 && name == searchString {
				currentProcess := Process{name, createdAt.String(), pid, ppid, uuid}
				output = append(output, currentProcess)
			}
		}
	}

	return output, nil
}

func renderJSON(returnedProcesses []Process) error {
	currentUser, err := getCurrentUser()
	if err != nil {
		fmt.Println(err)
	}
	macAddress, err := getMacAddress()
	if err != nil {
		fmt.Println(err)
	}

	type outputStruct struct {
		Username   string
		MacAddress string
		Processes  []Process
	}

	outputPackage := outputStruct{currentUser, macAddress, returnedProcesses}

	json, err := json.Marshal(outputPackage)
	if err != nil {
		return errors.New("Cannot Generate JSON")
	}

	fmt.Println(string(json))
	return nil
}

func outputToProcessWindow(returnedProcesses []Process) {
	go renderJSON(returnedProcesses)

	processWindow.SetText("")

	for _, singleProcess := range returnedProcesses {
		createdAt := singleProcess.CreatedAt
		name := singleProcess.Name
		pid := singleProcess.Pid
		uuid := singleProcess.Uuid
		outputString := fmt.Sprintf("%s - %s (%d) [%s] \r", createdAt, name, pid, uuid)
		for _, applicationString := range strings.Split(outputString, "\n") {
			processWindow.AppendText(applicationString + "\r\n")
		}
	}
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
			processWindow.SetText("")
			searchFieldString = searchField.Text()
			returnedProcesses, err := getProcesses()
			if err != nil {
				fmt.Println(err)
			} else {
				outputToProcessWindow(returnedProcesses)
			}
		}
	}
}

func main() {
	go initAutoRefresh()

	MainWindow{
		Title:   "Process Agent",
		MinSize: Size{500, 500},
		Layout:  VBox{},
		Children: []Widget{
			HSplitter{
				Children: []Widget{
					CheckBox{
						AssignTo: &autoRefreshCheckbox,
						Text:     "Auto Refresh",
						Checked:  false,
						OnCheckStateChanged: func() {
							autoRefresh = !autoRefresh
							checkboxValue := strconv.FormatBool(autoRefresh)
							checkboxOutput := fmt.Sprintf("Auto Refresh: %s \n", checkboxValue)
							processWindow.AppendText(checkboxOutput)
						},
					},
					CheckBox{
						AssignTo: &toggleDefaultsCheckbox,
						Text:     "Hide Defaults",
						Checked:  false,
						OnCheckStateChanged: func() {
							showDefaultProcesses = !showDefaultProcesses
							checkboxValue := strconv.FormatBool(showDefaultProcesses)
							checkboxOutput := fmt.Sprintf("Hide System Processes: %s \n", checkboxValue)
							processWindow.AppendText(checkboxOutput)
						},
					},
					TextEdit{
						AssignTo: &searchField,
					},
					PushButton{
						Text: "Filter",
						OnClicked: func() {
							searchFieldString = searchField.Text()
							returnedProcesses, err := getProcesses()
							if err != nil {
								fmt.Println(err)
							} else {
								outputToProcessWindow(returnedProcesses)
							}
						},
					},
				},
			},
			HSplitter{
				MinSize: Size{300, 570},
				Children: []Widget{
					TextEdit{AssignTo: &processWindow, ReadOnly: true},
				},
			},
			HSplitter{
				Children: []Widget{
					PushButton{
						Text: "Get All Processes",
						OnClicked: func() {
							searchFieldString = ""
							returnedProcesses, err := getProcesses()
							if err != nil {
								fmt.Println(err)
							} else {
								outputToProcessWindow(returnedProcesses)
							}
						},
					},
				},
			},
		},
	}.Run()
}
