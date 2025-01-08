package main

import (
	"fmt"
	"os"
	"strconv"
)

func applyRedirect(env *environment, sourceFd int, target string, flag int) error {
	perms := 0644 // TODO
	f, err := os.OpenFile(target, flag, os.FileMode(perms))
	if err != nil {
		return err
	}
	switch sourceFd {
	case 1:
		env.stdout = f
		return nil
	case 2:
		env.stderr = f
		return nil
	default:
		return fmt.Errorf("error: redirecting fd %d not supported", sourceFd)
	}
}

func flagForOp(typ tokenType) int {
	if typ == tokenOpGt {
		return os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	} else if typ == tokenOpGtGt {
		return os.O_CREATE | os.O_WRONLY | os.O_APPEND
	} else {
		panic("unhandled token type")
	}
}

func applyRedirects(env *environment, toks []token) ([]string, error) {
	var out []string
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
			fallthrough
		case tokenOpGtGt:
			if i+1 >= len(toks) {
				return nil, fmt.Errorf("syntax error: missing redirection target")
			}
			target := toks[i+1].cargo
			flag := flagForOp(tok.typ)
			if err = applyRedirect(env, sourceFd, target, flag); err != nil {
				return nil, err
			}
			sourceFd = 1
			i += 2

		case tokenWord:
			out = append(out, tok.cargo)
			i++
		}
	}

	return out, nil
}
