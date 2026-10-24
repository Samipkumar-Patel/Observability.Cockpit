//go:build ignore

package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	fmt.Println("Building Go backend plugin...")
	if err := os.MkdirAll("dist", 0755); err != nil {
		fmt.Printf("Failed to create dist dir: %v\n", err)
		os.Exit(1)
	}

	cmd := exec.Command("go", "build", "-o", "dist/samipkumar-llm-datasource_linux_amd64", "./pkg")
	cmd.Env = append(os.Environ(), "GOOS=linux", "GOARCH=amd64", "CGO_ENABLED=0")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Build failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Backend built into dist/. Run 'npm run build' to build the frontend into dist/module.js.")
}
