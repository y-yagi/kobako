package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

// injectable functions for easier testing
var lookPath = exec.LookPath
var getwd = os.Getwd
var execCommand = exec.Command

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "Usage: kobako <command> [args...] (mounts current directory into container by default)")
		return 2
	}

	dockerPath, err := lookPath("docker")
	if err != nil {
		fmt.Fprintln(stderr, "docker not found in PATH")
		return 127
	}

	image := os.Getenv("KOBAKO_IMAGE")
	if image == "" {
		cmdName := args[0]
		switch cmdName {
		case "go", "gofmt", "golangci-lint":
			image = "golang:alpine"
		case "python", "python3", "pip":
			image = "python:alpine"
		case "npx":
			image = "node:alpine"
		default:
			image = "alpine:latest"
		}
	}

	// Mount current directory into the container at /work by default.
	hostDir := os.Getenv("KOBAKO_HOST_DIR")
	if hostDir == "" {
		cwd, err := getwd()
		if err != nil {
			fmt.Fprintln(stderr, "failed to get working directory:", err)
			return 1
		}
		hostDir = cwd
	}

	workdir := os.Getenv("KOBAKO_WORKDIR")
	if workdir == "" {
		workdir = "/work"
	}

	dockerArgs := []string{"run", "--rm", "-i", "-v", hostDir + ":" + workdir, "-w", workdir, image}
	dockerArgs = append(dockerArgs, args...)

	cmd := execCommand(dockerPath, dockerArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	// Forward common termination signals to the child process
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for s := range sigs {
			if cmd.Process != nil {
				_ = cmd.Process.Signal(s)
			}
		}
	}()

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				return status.ExitStatus()
			}
		}
		fmt.Fprintln(stderr, "failed to run docker:", err)
		return 1
	}
	return 0
}
