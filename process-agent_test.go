package main

import (
	"testing"
)

func TestGetProcesses(t *testing.T){
	_, err := getProcesses()
	if err != nil {
		t.Error("Could Not Get Processes")
	}
}

func TestJSONOutput(t *testing.T){
	output, err := getProcesses()
	if err != nil {
		t.Error("Could Not Get Processes")
	}

	err = renderJSON(output)
	if err != nil {
		t.Error("Could Not Render JSON")
	}
}

func TestGetUsername(t *testing.T){
	_, err := getCurrentUser()
	if err != nil {
		t.Error("Could Not Get Current User")
	}
}


func TestGetMACAddr(t *testing.T){
	_, err := getMacAddress()
	if err != nil {
		t.Error("Could Not Get MacAddress")
	}
}
