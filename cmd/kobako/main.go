package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
)

// injectable functions for easier testing
var lookPath = exec.LookPath
var getwd = os.Getwd
var execCommand = exec.Command

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func containsShellOperators(s string) bool {
	ops := []string{"&&", "||", ";", "|", "$", "`", "`", "("}
	for _, o := range ops {
		if strings.Contains(s, o) {
			return true
		}
	}
	return false
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

	useShell := false
	if len(args) == 1 {
		a := args[0]
		if containsShellOperators(a) {
			useShell = true
		}
	}
	if len(args) > 0 && (args[0] == "--shell" || args[0] == "-s") {
		useShell = true
		// remove the flag
		args = args[1:]
		if len(args) == 0 {
			fmt.Fprintln(stderr, "--shell requires a command string")
			return 2
		}
	}

	dockerArgs := []string{"run", "--rm", "-i", "-v", hostDir + ":" + workdir, "-w", workdir, image}
	if useShell {
		shellCmd := ""
		for i, p := range args {
			if i > 0 {
				shellCmd += " "
			}
			shellCmd += p
		}
		dockerArgs = append(dockerArgs, "sh", "-c", shellCmd)
	} else {
		dockerArgs = append(dockerArgs, args...)
	}

	cmd := execCommand(dockerPath, dockerArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr

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
