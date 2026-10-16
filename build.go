//go:build ignore

package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	fmt.Println("Building Go backend plugin...")
	cmd := exec.Command("go", "build", "-o", "dist/gpx_custom_datasource_linux_amd64", "./pkg")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Build failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Build successful! Binary placed in dist/")
}
