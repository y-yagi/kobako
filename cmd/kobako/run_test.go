package main

import (
	"bytes"
	"errors"
	"os/exec"
	"testing"
)

func TestRun_Usage(t *testing.T) {
	var out, errb bytes.Buffer
	status := run([]string{}, &out, &errb)
	if status != 2 {
		t.Fatalf("expected status 2, got %d", status)
	}
	if errb.Len() == 0 {
		t.Fatalf("expected usage on stderr")
	}
}

func TestRun_NoDocker(t *testing.T) {
	origLook := lookPath
	defer func() { lookPath = origLook }()

	lookPath = func(file string) (string, error) {
		return "", errors.New("not found")
	}

	var out, errb bytes.Buffer
	status := run([]string{"echo", "hi"}, &out, &errb)
	if status != 127 {
		t.Fatalf("expected 127 when docker missing, got %d", status)
	}
}

func TestRun_SuccessFakeExec(t *testing.T) {
	// fake exec.Command to record args and pretend success
	origExec := execCommand
	defer func() { execCommand = origExec }()

	var recorded []string
	execCommand = func(name string, args ...string) *exec.Cmd {
		recorded = append([]string{name}, args...)
		// use a real command that exits 0: `true`
		return exec.Command("true")
	}

	var out, errb bytes.Buffer
	status := run([]string{"echo", "hello"}, &out, &errb)
	if status != 0 {
		t.Fatalf("expected 0, got %d", status)
	}
	if len(recorded) == 0 {
		t.Fatalf("expected docker command to be recorded")
	}
}
