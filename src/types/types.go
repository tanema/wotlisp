package types

import (
	"errors"
	"fmt"
)

type (
	// Base is a catchall type that all other types will be passed around as.
	Base interface{}
	// Atom hold a single value
	Atom struct{ Val Base }
	// Symbol is a string that is used for looking up definition in the environment
	Symbol string
	// Keyword is like a ruby symbol. It is a simplified string
	Keyword string
)

// Env is the object that is passed around containing the definition of the running environment
type Env interface {
	Child([]Base, []Base) (Env, error)
	Find(Symbol) Env
	Set(Symbol, Base)
	Get(Symbol) (Base, error)
}

// Collection is a general interface used to abstract the differences between Lists and Vectors
type Collection interface {
	Data() []Base
}

// List is a sequential data structure that grows unbound
type List struct {
	Forms []Base
	Meta  Base
}

// NewList will create a new list from variable arguments passed
func NewList(forms ...Base) *List {
	return &List{Forms: forms}
}

// Data satisfies the Collection interface, making it easier to handle in common situations
func (l *List) Data() []Base { return l.Forms }

// Vector is a sequential data structure that does not grow
type Vector struct {
	Forms []Base
	Meta  Base
}

// NewVect will create a new vector from variable arguments passed
func NewVect(forms ...Base) *Vector {
	return &Vector{Forms: forms}
}

// Data satisfies the Collection interface, making it easier to handle in common situations
func (l *Vector) Data() []Base { return l.Forms }

// Hashmap is a data structure that maps key to values
type Hashmap struct {
	Forms map[Base]Base
	Meta  Base
}

// NewHashmap will create a new hashmap using an array of keys and values that are interleaved.
// To exclude some keys from the definition, you can pass in an array of excluded keys
func NewHashmap(values []Base, excludeKeys ...Base) (*Hashmap, error) {
	if len(values)%2 == 1 {
		return nil, errors.New("Odd number of arguments to NewHashMap")
	}
	m := map[Base]Base{}
	for i := 0; i < len(values); i += 2 {
		key := values[i]
		found := false
		for _, exclude := range excludeKeys {
			if key == exclude {
				found = true
				break
			}
		}
		if !found {
			m[key] = values[i+1]
		}
	}
	return &Hashmap{Forms: m}, nil
}

// ToList will interleave keys and values back into an array
func (hm *Hashmap) ToList() []Base {
	values := []Base{}
	for key, val := range hm.Forms {
		values = append(values, key)
		values = append(values, val)
	}
	return values
}

// Keys will return an array of all the keys in the map
func (hm *Hashmap) Keys() []Base {
	keys := make([]Base, 0, len(hm.Forms))
	for k := range hm.Forms {
		keys = append(keys, k)
	}
	return keys
}

// Vals will return an array of all the values in the map
func (hm *Hashmap) Vals() []Base {
	vals := make([]Base, 0, len(hm.Forms))
	for _, v := range hm.Forms {
		vals = append(vals, v)
	}
	return vals
}

// StdFunc wraps a standard library function that does not need closure support
type StdFunc struct {
	Fn   func(Env, []Base) (Base, error)
	Meta Base
}

// Func is a helper to convert a simple function into a StdFunc
func Func(fn func(Env, []Base) (Base, error)) *StdFunc {
	return &StdFunc{Fn: fn}
}

// ExtFunc is a user-space defined function or macro that has closure capabilities
type ExtFunc struct {
	AST     Base
	Params  []Base
	Env     Env
	IsMacro bool
	eval    func(Env, Base) (Base, error)
	Meta    Base
}

// NewFunc will generate a closure environment around a simple function signature
// To be called later.
func NewFunc(env Env, eval func(Env, Base) (Base, error), args ...Base) (*ExtFunc, error) {
	if len(args) < 2 {
		return nil, errors.New("improperly formatted fn* statement")
	}

	var params []Base
	switch tparams := args[0].(type) {
	case Collection:
		params = tparams.Data()
	default:
		return nil, errors.New("invalid fn* param declaration")
	}

	return &ExtFunc{
		AST:    args[1],
		Params: params,
		Env:    env,
		eval:   eval,
	}, nil
}

// Apply will call the defined functions with the passed in arguments
func (fn *ExtFunc) Apply(arguments []Base) (Base, error) {
	newEnv, err := fn.Env.Child(fn.Params, arguments)
	if err != nil {
		return nil, err
	}
	return fn.eval(newEnv, fn.AST)
}

// Clone will generate a copy of the original function so that the original is left unmutated
func (fn *ExtFunc) Clone() *ExtFunc {
	return &ExtFunc{
		AST:     fn.AST,
		Params:  fn.Params,
		Env:     fn.Env,
		eval:    fn.eval,
		IsMacro: fn.IsMacro,
	}
}

// CallFunc will allow either StdFunc or ExtFunc to be passed and called. If it
// is not either then an error will be thrown
func CallFunc(e Env, baseFn Base, arguments []Base) (Base, error) {
	switch fn := baseFn.(type) {
	case *StdFunc:
		return fn.Fn(e, arguments)
	case *ExtFunc:
		return fn.Apply(arguments)
	default:
		return nil, fmt.Errorf("attempt to call non-function %v", baseFn)
	}
}

// UserError wraps values that are thrown by a user. This really can be any value
// but this is used in an error state
type UserError struct {
	Val Base
}

func (err UserError) Error() string {
	return "User Error"
}
