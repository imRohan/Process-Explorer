package main

import (
	"testing"
)

var test_user UserDetails

func TestGetProcesses(t *testing.T) {
	_, err := getProcesses()
	if err != nil {
		t.Error("Could Not Get Processes")
	}
}

func TestGetUsername(t *testing.T) {
	name, err := getCurrentUser()
	if err != nil {
		t.Error("Could Not Get Current User")
	}
  test_user.Name = name
}

func TestGetMACAddr(t *testing.T) {
	mac, err := getMacAddress()
	if err != nil {
		t.Error("Could Not Get MacAddress")
	}

  test_user.Mac = mac
}

func TestJSONOutput(t *testing.T) {
	output,_ := getProcesses()

	err := renderJSON(RenderParams{test_user, output, exportData})
	if err != nil {
		t.Error("Could Not Render JSON")
	}
}

