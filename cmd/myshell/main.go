package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	// implements "Shell Command Language" per the POSIX standard:
	// https://pubs.opengroup.org/onlinepubs/9799919799/utilities/V3_chap02.html

	p := newPrompter("$ ", os.Stdout, os.Stdin)
	for {
		line, err := p.readline()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		if line == "" {
			continue
		}

		toks, err := tokenize(line)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			continue
		}
		if len(toks) == 0 {
			continue
		}

		env := defaultEnv()
		toks, err = applyRedirects(&env, toks)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			continue
		}

		cmd, err := parse(toks)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			continue
		}

		cmd.Execute(env)
	}
}
