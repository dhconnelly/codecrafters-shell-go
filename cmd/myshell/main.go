package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
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

type command interface{}

func parse(line string) (command, error) {
	return nil, fmt.Errorf("%s: %w", line, ErrCommandNotFound)
}

func main() {
	p := newPrompter("$ ", os.Stdout, os.Stdin)
	for {
		line, err := p.readline()
		if err == io.EOF {
			break
		}
		if line == "" {
			continue
		}
		if err != nil {
			panic(err)
		}
		if _, err := parse(line); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
		}
	}
}
