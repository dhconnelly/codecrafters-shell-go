package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
	"unicode"
)

type commandType int

const (
	exitBuiltIn commandType = iota
	echoBuiltIn
	typeBuiltIn
	pwdBuiltIn
	cdBuiltIn
	executable
)

var builtIns = map[string]commandType{
	"exit": exitBuiltIn,
	"echo": echoBuiltIn,
	"type": typeBuiltIn,
	"pwd":  pwdBuiltIn,
	"cd":   cdBuiltIn,
}

type commandInfo struct {
	typ  commandType
	path string
}

func resolveExecutable(name string) (string, bool) {
	// this is basically os/exec.LookPath
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
	path, ok := resolveExecutable(name)
	if ok {
		return commandInfo{typ: executable, path: path}, true
	}
	return commandInfo{}, false
}

func resolvePath(path string) string {
	// TODO: this only happens the specific codecrafters case
	if path == "~" {
		return os.Getenv("HOME")
	}
	return path
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

func parse(toks []string) (command, error) {
	name, suffix := toks[0], toks[1:]
	cmd, ok := resolveCommand(name)
	if !ok {
		return nil, fmt.Errorf("%s: command not found", name)
	}

	switch cmd.typ {
	case exitBuiltIn:
		if len(suffix) != 1 {
			return nil, fmt.Errorf("usage: exit <code>")
		}
		code, err := strconv.Atoi(suffix[0])
		if err != nil {
			return nil, fmt.Errorf("exit: invalid code")
		}
		return exitCommand{code: code}, nil

	case echoBuiltIn:
		return echoCommand{words: suffix}, nil

	case typeBuiltIn:
		if len(suffix) != 1 {
			return nil, fmt.Errorf("usage: type <command>")
		}
		cmd2, ok := resolveCommand(suffix[0])
		if !ok {
			return nil, fmt.Errorf("%s: not found", suffix[0])
		}
		return typeCommand{name: suffix[0], typ: cmd2.typ, path: cmd2.path}, nil

	case pwdBuiltIn:
		wd, _ := os.Getwd()
		return pwdCommand{path: wd}, nil

	case cdBuiltIn:
		if len(suffix) != 1 {
			return nil, fmt.Errorf("usage: cd <path>")
		}
		path := resolvePath(suffix[0])
		return cdCommand{path: path}, nil

	case executable:
		return executableCommand{name: cmd.path, args: suffix}, nil

	default:
		panic(fmt.Sprintf("unhandled command: %v", cmd.typ))
	}
}
