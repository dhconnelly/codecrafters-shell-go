package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
	"syscall"
	"unicode"
)

type prompter struct {
	prompt string
	w      io.Writer
	s      *bufio.Scanner
}

func newPrompter(prompt string, w io.Writer, r io.Reader) *prompter {
	s := bufio.NewScanner(r)
	s.Split(bufio.ScanLines)
	return &prompter{prompt: prompt, w: w, s: s}
}

func (p *prompter) readline() (string, error) {
	fmt.Fprint(p.w, p.prompt)
	if p.s.Scan() {
		return p.s.Text(), nil
	}
	if p.s.Err() != nil {
		return "", p.s.Err()
	}
	return "", io.EOF
}

type echoCommand struct {
	words []string
}

type exitCommand struct {
	code int
}

type typeCommand struct {
	name string
	typ  commandType
	path string
}

type pwdCommand struct {
	path string
}

type cdCommand struct {
	path string
}

type executableCommand struct {
	args []string
}

type commandType int

const (
	exit commandType = iota
	echo
	typ
	pwd
	cd
	executable
)

type command struct {
	typ   commandType
	cargo interface{}
}

func tokenize(line string) ([]string, error) {
	r := bufio.NewReader(strings.NewReader(line))
	var tokens []string
	var cur []rune
	for {
		// TODO: operators
		// TODO: quoting
		// TODO: comments
		// TODO: assignment
		c, _, err := r.ReadRune()
		if err == io.EOF {
			if len(cur) > 0 {
				tokens = append(tokens, string(cur))
			}
			return tokens, nil
		} else if err != nil {
			return nil, err
		} else if unicode.IsSpace(c) {
			if len(cur) > 0 {
				tokens = append(tokens, string(cur))
				cur = cur[:0]
			}
		} else {
			cur = append(cur, c)
		}
	}
}

var builtIns = map[string]commandType{
	"exit": exit,
	"echo": echo,
	"type": typ,
	"pwd":  pwd,
	"cd":   cd,
}

type commandInfo struct {
	typ  commandType
	path string
}

func locateExecutable(name string) (string, bool) {
	searchPaths := strings.Split(os.Getenv("PATH"), ":")
	for _, searchPath := range searchPaths {
		path := path.Join(searchPath, name)
		stat, err := os.Stat(path)
		if err != nil {
			// doesn't matter what it is if we can't execute it
			continue
		}
		if stat.Mode().Perm()&0100 > 0 {
			return path, true
		}
	}
	return "", false
}

func resolveCommand(name string) (commandInfo, bool) {
	builtIn, ok := builtIns[name]
	if ok {
		return commandInfo{typ: builtIn}, true
	}
	path, ok := locateExecutable(name)
	if ok {
		return commandInfo{typ: executable, path: path}, true
	}
	return commandInfo{}, false
}

func parse(toks []string) (command, error) {
	name, suffix := toks[0], toks[1:]
	cmd, ok := resolveCommand(name)
	if !ok {
		return command{}, fmt.Errorf("%s: command not found", name)
	}

	switch cmd.typ {
	case exit:
		if len(suffix) != 1 {
			return command{}, fmt.Errorf("usage: exit <code>")
		}
		code, err := strconv.Atoi(suffix[0])
		if err != nil {
			return command{}, fmt.Errorf("exit: invalid code")
		}
		return command{typ: exit, cargo: exitCommand{code: code}}, nil

	case echo:
		return command{typ: echo, cargo: echoCommand{words: suffix}}, nil

	case typ:
		if len(suffix) != 1 {
			return command{}, fmt.Errorf("usage: type <command>")
		}
		cmd2, ok := resolveCommand(suffix[0])
		if !ok {
			return command{}, fmt.Errorf("%s: not found", suffix[0])
		}
		return command{
			typ:   typ,
			cargo: typeCommand{name: suffix[0], typ: cmd2.typ, path: cmd2.path},
		}, nil

	case executable:
		args := append([]string{cmd.path}, suffix...)
		return command{
			typ:   executable,
			cargo: executableCommand{args: args},
		}, nil

	case pwd:
		wd, err := os.Getwd()
		if err != nil {
			return command{}, err
		}
		return command{typ: pwd, cargo: pwdCommand{path: wd}}, nil

	case cd:
		if len(suffix) != 1 {
			return command{}, fmt.Errorf("usage: cd <path>")
		}
		path := suffix[0]
		if path == "~" {
			path = os.Getenv("HOME")
		}
		return command{typ: cd, cargo: cdCommand{path: path}}, nil

	default:
		panic(fmt.Sprintf("unhandled command: %v", cmd.typ))
	}
}

func execute(args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	pid, err := syscall.ForkExec(args[0], args, &syscall.ProcAttr{
		Dir: cwd,
		Files: []uintptr{
			os.Stdin.Fd(),
			os.Stdout.Fd(),
			os.Stderr.Fd(),
		},
	})
	if err != nil {
		return err
	}
	var status syscall.WaitStatus
	_, err = syscall.Wait4(pid, &status, 0, nil)
	return err
}

func main() {
	// implements "Shell Command Language" per the POSIX standard:
	// https://pubs.opengroup.org/onlinepubs/9799919799/utilities/V3_chap02.html

	// TODO: intercept signals
	p := newPrompter("$ ", os.Stdout, os.Stdin)
	for {
		line, err := p.readline()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}

		toks, err := tokenize(line)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			continue
		}
		if len(toks) == 0 {
			continue
		}

		cmd, err := parse(toks)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			continue
		}

		switch cmd.typ {
		case exit:
			os.Exit(cmd.cargo.(exitCommand).code)
		case echo:
			fmt.Println(strings.Join(cmd.cargo.(echoCommand).words, " "))
		case typ:
			typeCmd := cmd.cargo.(typeCommand)
			switch typeCmd.typ {
			case executable:
				fmt.Printf("%s is %s\n", typeCmd.name, typeCmd.path)
			default:
				fmt.Printf("%s is a shell builtin\n", typeCmd.name)
			}
		case executable:
			bin := cmd.cargo.(executableCommand)
			if err := execute(bin.args); err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err)
			}
		case pwd:
			fmt.Println(cmd.cargo.(pwdCommand).path)
		case cd:
			path := cmd.cargo.(cdCommand).path
			if err := os.Chdir(path); err != nil {
				fmt.Fprintf(os.Stderr, "cd: %s: No such file or directory\n", path)
			}
		default:
			panic(fmt.Sprintf("unhandled command: %v", cmd.typ))
		}
	}
}
