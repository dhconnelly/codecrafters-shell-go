package main

import (
	"fmt"
	"os"
	"strconv"
)

func applyRedirect(env *environment, sourceFd int, dest *os.File) error {
	switch sourceFd {
	case 1:
		env.stdout = dest
		return nil
	case 2:
		env.stderr = dest
		return nil
	default:
		return fmt.Errorf("error: redirecting fd %d not supported", sourceFd)
	}
}

func applyRedirects(env *environment, toks []token) ([]token, error) {
	var out []token
	var err error
	sourceFd := 1
	i := 0

	for i < len(toks) {
		tok := toks[i]

		switch tok.typ {
		case tokenIONumber:
			sourceFd, err = strconv.Atoi(tok.cargo)
			if err != nil {
				panic(err) // should be caught in the tokenizer
			}
			i++

		case tokenOpGt:
			if i+1 >= len(toks) {
				return nil, fmt.Errorf("syntax error: missing redirection target")
			}
			target := toks[i+1].cargo
			perms := 0644 // TODO
			f, err := os.OpenFile(target, os.O_WRONLY|os.O_TRUNC, os.FileMode(perms))
			if err != nil {
				return nil, err
			}
			if err = applyRedirect(env, sourceFd, f); err != nil {
				return nil, err
			}
			sourceFd = 1
			i += 2

		case tokenWord:
			out = append(out, tok)
			i++
		}
	}

	return out, nil
}
