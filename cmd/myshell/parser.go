package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path"
	"regexp"
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

type tokenType int

const (
	tokenWord tokenType = iota
	tokenIONumber
	tokenOpGt
)

type token struct {
	typ   tokenType
	cargo string
}

func isDelim(c rune) bool {
	return unicode.IsSpace(c) || c == '>'
}

var allDigits = regexp.MustCompile(`^\d+$`)

func emit(cargo string, delim rune) token {
	// https://pubs.opengroup.org/onlinepubs/9799919799/utilities/V3_chap02.html#tag_19_10_01
	if cargo == ">" {
		return token{typ: tokenOpGt, cargo: cargo}
	} else if allDigits.MatchString(cargo) && delim == '>' {
		return token{typ: tokenIONumber, cargo: cargo}
	} else {
		return token{typ: tokenWord, cargo: cargo}
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
		args := append([]string{name}, suffix...)
		return executableCommand{path: cmd.path, args: args}, nil

	default:
		panic(fmt.Sprintf("unhandled command: %v", cmd.typ))
	}
}

func tokenize(line string) ([]token, error) {
	r := bufio.NewReader(strings.NewReader(line))
	var tokens []token
	var cur []rune
	for {
		// TODO: quoting
		// TODO: comments
		// TODO: assignment
		c, _, err := r.ReadRune()
		if err == io.EOF {
			if len(cur) > 0 {
				tokens = append(tokens, emit(string(cur), 0))
			}
			return tokens, nil
		} else if err != nil {
			return nil, err
		} else if isDelim(c) {
			if len(cur) > 0 {
				tokens = append(tokens, emit(string(cur), c))
				cur = cur[:0]
			}
			if !unicode.IsSpace(c) {
				cur = append(cur, c)
			}
		} else {
			cur = append(cur, c)
		}
	}
}
