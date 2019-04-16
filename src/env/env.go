package env

import (
	"fmt"

	"github.com/tanema/mal/src/types"
)

// Env captures all definitions and their values
type Env struct {
	data  map[string]types.Base
	outer types.Env
}

// New creates a new env, binds and exprs allow for parameter binding
func New(outer types.Env, binds, exprs []types.Base) (*Env, error) {
	env := &Env{data: map[string]types.Base{}, outer: outer}
	for i, bind := range binds {
		key, ok := bind.(types.Symbol)
		if !ok {
			return nil, fmt.Errorf("non-symbol bind value")
		}
		if key == "&" {
			env.Set(binds[i+1].(types.Symbol), types.NewList(exprs[i:]...))
			break
		}
		env.Set(key, exprs[i])
	}
	return env, nil
}

// Child creates a new Env that inherits this one. By adding this method we can
// pass around env without importing env
func (e *Env) Child(binds, exprs []types.Base) (types.Env, error) {
	return New(e, binds, exprs)
}

// Find will find the env with the definition available. It will return nil otherwise
func (e *Env) Find(key types.Symbol) types.Env {
	if _, ok := e.data[string(key)]; ok {
		return e
	} else if e.outer != nil {
		return e.outer.Find(key)
	}
	return nil
}

// Set will set the definition of a symbol on the current env
func (e *Env) Set(key types.Symbol, value types.Base) {
	e.data[string(key)] = value
}

// Get will retreive the value of a symbol recursively up the parentage of this env
func (e *Env) Get(key types.Symbol) (types.Base, error) {
	env := e.Find(key)
	if env == nil {
		return nil, fmt.Errorf("'%v' not found", key)
	}
	return env.(*Env).data[string(key)], nil
}
