package main

import (
	"testing"
)

func TestMainParseHelp(t *testing.T) {
	_, msg, status, err := ParseArgs([]string{})
	if err != nil {
		t.Error("err not nil")
	}
	if msg != usage {
		t.Error("no usage message")
	}
	if status == 0 {
		t.Error("no error status")
	}

	_, msg, status, err = ParseArgs([]string{"-h"})
	if err != nil {
		t.Error("err not nil")
	}
	if msg != usage {
		t.Error("no usage message")
	}
	if status != 0 {
		t.Error("unexpected error status")
	}

	_, msg, status, err = ParseArgs([]string{"-v"})
	if err != nil {
		t.Error("err not nil")
	}
	if msg != version {
		t.Error("no version message")
	}
	if status != 0 {
		t.Error("unexpected error status")
	}
}

func TestMainParseCommand(t *testing.T) {
	job, _, _, err := ParseArgs([]string{"-c", "true"})
	if job == nil {
		t.Error("job nil")
	}
	if err != nil {
		t.Error("err not nil")
	}
	if job.Command == nil {
		t.Error("command nil")
	}
	if job.Script != nil {
		t.Error("script not nil")
	}
}

func TestMainParseScriptNotExistent(t *testing.T) {
	job, msg, _, _ := ParseArgs([]string{"-s", "examples/notfound.sh"})
	if msg == "" {
		t.Error("no message")
	}
	if job != nil {
		t.Error("job not nil")
	}
}

func TestMainParseScript(t *testing.T) {
	job, _, _, err := ParseArgs([]string{"-s", "examples/hello.sh"})
	if err != nil {
		t.Error("err not nil")
	}
	if job == nil {
		t.Error("job nil")
		return
	}
	if job.Script == nil {
		t.Error("script nil")
	}
	if job.Command != nil {
		t.Error("command not nil")
	}
	if job.Script.IsDirMode() {
		t.Error("unexpected dirmode")
	}
}

func TestMainParseScriptDirMode(t *testing.T) {
	job, _, _, err := ParseArgs([]string{"-s", "examples/bootstrap"})
	if err != nil {
		t.Error("err not nil")
	}
	if job == nil {
		t.Error("job nil")
		return
	}
	if job.Script == nil {
		t.Error("script nil")
	}
	if job.Command != nil {
		t.Error("command not nil")
	}
	if !job.Script.IsDirMode() {
		t.Error("expected dirmode")
	}
}