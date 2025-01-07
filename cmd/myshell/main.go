package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
)

type prompter struct {
	prompt string
	w      io.Writer
	r      *bufio.Reader
}

func newPrompter(prompt string, w io.Writer, r io.Reader) *prompter {
	return &prompter{prompt: prompt, w: w, r: bufio.NewReader(r)}
}

func (p *prompter) readline() (string, error) {
	fmt.Fprint(p.w, p.prompt)
	line, err := p.r.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	return line, nil
}

func main() {
	p := newPrompter("$ ", os.Stdout, os.Stdin)
	if _, err := p.readline(); err != nil {
		panic(err)
	}
}
