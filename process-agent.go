package main

import (
	"fmt"
	"github.com/imRohan/go-ps"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/satori/go.uuid"
	"log"
	"strconv"
	"strings"
	"time"
)

const (
	defaultProcessDuration = "2562047h47m16.854775807s"
)

type Process struct {
	name     string
	duration time.Duration
	pid      int
	ppid     int
	uuid     uuid.UUID
}

func getProcesses(defaultProcesses bool, searchString string) []Process {
	processes, err := ps.Processes()

	if err != nil {
		log.Fatal(err)
	}

	var output []Process

	for _, process := range processes {
		duration := processDuration(process)
		name := strings.Split(process.Executable(), ".exe")[0]
		pid := process.Pid()
		ppid := process.PPid()
		uuid := uuid.NewV1()
		if !defaultProcesses || defaultProcesses && duration.String() != defaultProcessDuration {
			if len(searchString) == 0 || len(searchString) != 0 && name == searchString {
				currentProcess := Process{name, duration, pid, ppid, uuid}
				output = append(output, currentProcess)
			}
		}
	}

	return output
}

func processDuration(process ps.Process) time.Duration {
	_processCreationTime := process.CreationTime()
	_duration := time.Since(_processCreationTime)

	return _duration
}

func outputToProcessWindow(processWindow *walk.TextEdit, returnedProcesses []Process) {
	fmt.Println(returnedProcesses)
	for _, singleProcess := range returnedProcesses {
		name := singleProcess.name
		duration := singleProcess.duration
		timeString := durationSplit(duration)
		pid := singleProcess.pid
		uuid := singleProcess.uuid
		outputString := fmt.Sprintf("%s - %s (%d) [%s] \r", timeString, name, pid, uuid)
		for _, applicationString := range strings.Split(outputString, "\n") {
			processWindow.AppendText(applicationString + "\r\n")
		}
	}
}

func durationSplit(duration time.Duration) string {
	durationString := duration.String()
	splitM := strings.Split(durationString, "m")
	splitH := strings.Split(string(splitM[0]), "h")
	output := ""
	if len(splitH) == 2 {
		h := string(splitH[0])
		m := string(splitH[1])
		s := string(splitM[1])
		output = fmt.Sprintf("%s:%s:%s", h, m, s[0:2])
	} else {
		if len(splitM) == 2 {
			m := string(splitH[0])
			s := string(splitM[1])
			output = fmt.Sprintf("00:%s:%s", m, s[0:2])
		} else {
			s := string(splitM[0])
			output = fmt.Sprintf("00:00:%s", s[0:1])
		}
	}
	return output
}

func main() {

	var processWindow, searchField *walk.TextEdit
	var toggleDefaultsCheckBox *walk.CheckBox
	showDefaultProcesses := false
	searchFieldString := ""

	MainWindow{
		Title:   "Go Look At Processes!",
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
						},
					},
				},
			},
		},
	}.Run()

}
