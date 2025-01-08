package main

import (
	"fmt"
	"os"
	"strings"
	"syscall"
)

type command interface {
	Execute()
}

type echoCommand struct {
	words []string
}

func (cmd echoCommand) Execute() {
	fmt.Println(strings.Join(cmd.words, " "))
}

type exitCommand struct {
	code int
}

func (cmd exitCommand) Execute() {
	os.Exit(cmd.code)
}

type typeCommand struct {
	name string
	typ  commandType
	path string
}

func (cmd typeCommand) Execute() {
	switch cmd.typ {
	case executable:
		fmt.Printf("%s is %s\n", cmd.name, cmd.path)
	default:
		fmt.Printf("%s is a shell builtin\n", cmd.name)
	}
}

type pwdCommand struct {
	path string
}

func (cmd pwdCommand) Execute() {
	fmt.Println(cmd.path)
}

type cdCommand struct {
	path string
}

func (cmd cdCommand) Execute() {
	if err := os.Chdir(cmd.path); err != nil {
		fmt.Fprintf(os.Stderr, "cd: %s: No such file or directory\n", cmd.path)
	}
}

type executableCommand struct {
	path string
	args []string
}

func (cmd executableCommand) Execute() {
	// this is basically os/exec.Command.Run
	cwd, _ := os.Getwd()
	pid, err := syscall.ForkExec(cmd.path, cmd.args, &syscall.ProcAttr{
		Dir: cwd,
		Files: []uintptr{
			os.Stdin.Fd(),
			os.Stdout.Fd(),
			os.Stderr.Fd(),
		},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
	}
	var status syscall.WaitStatus
	if _, err = syscall.Wait4(pid, &status, 0, nil); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
	}
}
