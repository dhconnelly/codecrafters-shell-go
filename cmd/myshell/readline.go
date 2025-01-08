package main

import (
	"bufio"
	"fmt"
	"io"
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
	// TODO: use direct syscalls (unbuffered read, pipe, and poll) to replicate
	// bash's sigterm handling
	fmt.Fprint(p.w, p.prompt)
	if p.s.Scan() {
		return p.s.Text(), nil
	}
	if p.s.Err() != nil {
		return "", p.s.Err()
	}
	return "", io.EOF
}
