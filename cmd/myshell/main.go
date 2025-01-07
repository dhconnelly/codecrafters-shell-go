package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"unicode"
)

var (
	ErrCommandNotFound = errors.New("command not found")
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
}

type commandType int

const (
	badCommand commandType = iota
	noop
	exit
	echo
	typ
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

var builtIns = map[string]commandType{"exit": exit, "echo": echo, "type": typ}

func parse(line string) (command, error) {
	toks, err := tokenize(line)
	if err != nil {
		return command{}, err
	}
	if len(toks) == 0 {
		return command{typ: noop}, nil
	}

	name, suffix := toks[0], toks[1:]
	switch name {
	case "exit":
		if len(suffix) != 1 {
			return command{}, fmt.Errorf("usage: exit <code>")
		}
		code, err := strconv.Atoi(suffix[0])
		if err != nil {
			return command{}, fmt.Errorf("exit: invalid code")
		}
		return command{typ: exit, cargo: exitCommand{code: code}}, nil

	case "echo":
		return command{typ: echo, cargo: echoCommand{words: suffix}}, nil

	case "type":
		if len(suffix) != 1 {
			return command{}, fmt.Errorf("usage: type <command>")
		}
		arg := suffix[0]
		builtIn, ok := builtIns[arg]
		if ok {
			return command{typ: typ, cargo: typeCommand{name: arg, typ: builtIn}}, nil
		}
		return command{typ: typ, cargo: typeCommand{name: arg, typ: badCommand}}, nil

	default:
		return command{}, fmt.Errorf("%s: %w", name, ErrCommandNotFound)
	}
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
		cmd, err := parse(line)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			continue
		}

		switch cmd.typ {
		case noop:
			continue
		case exit:
			os.Exit(cmd.cargo.(exitCommand).code)
		case echo:
			fmt.Println(strings.Join(cmd.cargo.(echoCommand).words, " "))
		case typ:
			typeCmd := cmd.cargo.(typeCommand)
			switch typeCmd.typ {
			case badCommand:
				fmt.Printf("%s: not found\n", typeCmd.name)
			default:
				fmt.Printf("%s is a shell builtin\n", typeCmd.name)
			}
		default:
			panic(fmt.Sprintf("unhandled command: %v", cmd.typ))
		}
	}
}
