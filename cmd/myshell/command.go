package main

import (
	"fmt"
	"os"
	"strings"
	"syscall"
)

type environment struct {
	stdout *os.File
	stderr *os.File
}

func defaultEnv() environment {
	return environment{
		stdout: os.Stdout,
		stderr: os.Stderr,
	}
}

type command interface {
	Execute(env environment)
}

type echoCommand struct {
	words []string
}

func (cmd echoCommand) Execute(env environment) {
	fmt.Fprintln(env.stdout, strings.Join(cmd.words, " "))
}

type exitCommand struct {
	code int
}

func (cmd exitCommand) Execute(env environment) {
	os.Exit(cmd.code)
}

type typeCommand struct {
	name string
	typ  commandType
	path string
}

func (cmd typeCommand) Execute(env environment) {
	switch cmd.typ {
	case executable:
		fmt.Fprintf(env.stdout, "%s is %s\n", cmd.name, cmd.path)
	default:
		fmt.Fprintf(env.stdout, "%s is a shell builtin\n", cmd.name)
	}
}

type pwdCommand struct {
	path string
}

func (cmd pwdCommand) Execute(env environment) {
	fmt.Fprintln(env.stdout, cmd.path)
}

type cdCommand struct {
	path string
}

func (cmd cdCommand) Execute(env environment) {
	if err := os.Chdir(cmd.path); err != nil {
		fmt.Fprintf(env.stderr, "cd: %s: No such file or directory\n", cmd.path)
	}
}

type executableCommand struct {
	path string
	args []string
}

func (cmd executableCommand) Execute(env environment) {
	// this is basically os/exec.Command.Run
	cwd, _ := os.Getwd()
	pid, err := syscall.ForkExec(cmd.path, cmd.args, &syscall.ProcAttr{
		Dir: cwd,
		Files: []uintptr{
			os.Stdin.Fd(),
			env.stdout.Fd(),
			env.stderr.Fd(),
		},
	})
	if err != nil {
		panic(err)
	}
	var status syscall.WaitStatus
	if _, err = syscall.Wait4(pid, &status, 0, nil); err != nil {
		panic(err)
	}
}
