package main

import (
	"encoding/json"
	"fmt"
	"github.com/imRohan/go-ps"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/satori/go.uuid"
	"log"
	"net"
	"os/user"
	"strconv"
	"strings"
)

type Process struct {
	Name      string
	CreatedAt string
	Pid       int
	Ppid      int
	Uuid      uuid.UUID
}

func getProcesses(defaultProcesses bool, searchString string) []Process {
	processes, err := ps.Processes()

	if err != nil {
		log.Fatal(err)
	}

	var output []Process

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

	return output
}

func renderJSON(returnedProcesses []Process) {
	json, err := json.Marshal(returnedProcesses)
	if err != nil {
		fmt.Println("Cannot create JSON ", err)
	}

	fmt.Println(string(json))
}

func outputToProcessWindow(processWindow *walk.TextEdit, returnedProcesses []Process) {
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

func getCurrentUser() string {
	user, err := user.Current()

	if err != nil {
		log.Fatal(err)
	}

	username := fmt.Sprintf("%s", user.Username)
	return username
}

func getMacAddress() string {
	interfaces, _ := net.Interfaces()
	macAddress := ""
	for _, singleInterface := range interfaces {
		hardwareName := singleInterface.Name
		if hardwareName == "Ethernet" || hardwareName == "ethernet" {
			macAddress = singleInterface.HardwareAddr.String()
		}
	}
	return macAddress
}

func main() {

	var processWindow, searchField *walk.TextEdit
	var toggleDefaultsCheckBox *walk.CheckBox
	showDefaultProcesses := false
	searchFieldString := ""
	currentUser := getCurrentUser()
	macAddress := getMacAddress()
	windowTitle := fmt.Sprintf("%s - %s", currentUser, macAddress)

	MainWindow{
		Title:   windowTitle,
		MinSize: Size{300, 600},
		Layout:  VBox{},
		Children: []Widget{
			HSplitter{
				Children: []Widget{
					CheckBox{
						AssignTo: &toggleDefaultsCheckBox,
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
							processWindow.SetText("")
							searchFieldString = searchField.Text()
							returnedProcesses := getProcesses(showDefaultProcesses, searchFieldString)
							outputToProcessWindow(processWindow, returnedProcesses)
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
						Text: "Get Processes",
						OnClicked: func() {
							processWindow.SetText("")
							returnedProcesses := getProcesses(showDefaultProcesses, "")
							outputToProcessWindow(processWindow, returnedProcesses)
							renderJSON(returnedProcesses)
						},
					},
				},
			},
		},
	}.Run()

}
