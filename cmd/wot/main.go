package main

import (
	"fmt"
	"os"

	"github.com/tanema/mal/src/core"
	"github.com/tanema/mal/src/env"
	"github.com/tanema/mal/src/printer"
	"github.com/tanema/mal/src/reader"
	"github.com/tanema/mal/src/readline"
	"github.com/tanema/mal/src/runtime"
	"github.com/tanema/mal/src/types"
)

func main() {
	defaultEnv := core.DefaultNamespace()
	if len(os.Args) > 1 {
		runFile(defaultEnv, os.Args[1], os.Args[2:]...)
	} else {
		runREPL(defaultEnv)
	}
}

func runFile(e *env.Env, path string, argv ...string) {
	targv := make([]types.Base, len(argv))
	for i, arg := range argv {
		targv[i] = types.Base(arg)
	}
	e.Set("*ARGV*", types.NewList(targv...))
	if _, err := rep(`(load-file "`+os.Args[1]+`")`, e); err != nil {
		fmt.Println(printer.Print(err, true))
	}
}

func runREPL(env *env.Env) error {
	keepRunning := true
	env.Set("*ARGV*", types.NewList())
	env.Set("exit", types.Func(func(e types.Env, a []types.Base) (types.Base, error) {
		keepRunning = false
		return nil, nil
	}))
	for keepRunning {
		text, err := readline.Readline("user> ")
		if err != nil {
			fmt.Println(printer.Print(err, true))
			continue
		}
		valStr, err := rep(text, env)
		if err != nil {
			fmt.Println(printer.Print(err, true))
			continue
		}
		fmt.Println(valStr)
	}
	return nil
}

func rep(in string, e *env.Env) (string, error) {
	ast, parseErr := reader.ReadString(in)
	if parseErr != nil {
		return "", parseErr
	}
	val, evalErr := runtime.Eval(e, ast)
	if evalErr != nil {
		return "", evalErr
	}
	return printer.Print(val, true), nil
}
