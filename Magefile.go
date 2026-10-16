// +build mage

package main

import (
	"fmt"
	"os"
	"os/exec"
)

// Default target to build the plugin backend
func Default() error {
	return BuildBackend()
}

// BuildBackend compiles the Go backend plugin binary
func BuildBackend() error {
	fmt.Println("Building Go backend plugin...")
	cmd := exec.Command("go", "build", "-o", "dist/gpx_custom_datasource_linux_amd64", "./pkg")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Clean removes the dist directory
func Clean() error {
	fmt.Println("Cleaning dist directory...")
	return os.RemoveAll("dist")
}
