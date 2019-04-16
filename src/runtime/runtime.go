package runtime

import (
	"fmt"

	"github.com/tanema/mal/src/types"
)

// Eval will take in an AST and evaluate it, executing each command
func Eval(e types.Env, object types.Base) (types.Base, error) {
	var err error
	for {
		object, err = macroExpand(e, object)
		if err != nil {
			return nil, err
		}

		switch tobject := object.(type) {
		case *types.List:
			if len(tobject.Forms) == 0 {
				return tobject, nil
			}
			sym, _ := tobject.Forms[0].(types.Symbol)
			switch sym {
			case "try*":
				object, err = evalTry(e, tobject.Forms[1:]...)
			case "quote":
				if len(tobject.Forms) < 2 {
					return nil, nil
				}
				return tobject.Forms[1], nil
			case "quasiquote":
				object = evalQuasiQuote(e, tobject.Forms[1])
			case "do":
				object, err = evalDo(e, tobject.Forms[1:]...)
			case "if":
				object, err = evalIf(e, tobject.Forms[1:]...)
			case "defmacro!":
				return evalDefMacro(e, tobject.Forms[1:]...)
			case "macroexpand":
				return macroExpand(e, tobject.Forms[1])
			case "fn*":
				return types.NewFunc(e, Eval, tobject.Forms[1:]...)
			case "def!":
				return evalDef(e, tobject.Forms[1:]...)
			case "let*":
				object, e, err = evalLet(e, tobject.Forms[1:]...)
			default:
				lst, err := evalAST(e, tobject)
				if err != nil {
					return nil, err
				}
				list := lst.(*types.List)
				switch fn := list.Forms[0].(type) {
				case *types.StdFunc:
					return fn.Fn(e, list.Forms[1:])
				case *types.ExtFunc:
					newEnv, err := fn.Env.Child(fn.Params, list.Forms[1:])
					if err != nil {
						return nil, err
					}
					object, e = fn.AST, newEnv
				default:
					return nil, fmt.Errorf("attempt to call non-function %v", list.Forms[0])
				}
			}
		default:
			return evalAST(e, tobject)
		}

		if err != nil {
			return nil, err
		}
	}
}

func evalAST(env types.Env, ast types.Base) (types.Base, error) {
	switch tobject := ast.(type) {
	case types.Symbol:
		symVal, err := env.Get(tobject)
		if err != nil {
			return nil, err
		}
		return symVal, nil
	case *types.List:
		lst, err := evalListForms(tobject.Data(), env)
		return &types.List{Forms: lst}, err
	case *types.Vector:
		lst, err := evalListForms(tobject.Forms, env)
		return &types.Vector{Forms: lst}, err
	case *types.Hashmap:
		lst, err := evalListForms(tobject.ToList(), env)
		if err != nil {
			return nil, err
		}
		return types.NewHashmap(lst)
	default:
		return ast, nil
	}
}

func evalTry(e types.Env, args ...types.Base) (types.Base, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("not enough arguments")
	}

	var catch []types.Base
	var catchDefined bool

	if len(args) > 1 {
		catch, catchDefined = isPair(args[1])
		if catchDefined {
			catchSym, isCatchSym := catch[0].(types.Symbol)
			_, isErrSym := catch[1].(types.Symbol)
			if len(catch) < 3 || !isCatchSym || catchSym != "catch*" || !isErrSym {
				return nil, fmt.Errorf("invalid catch declaration")
			}
		}
	}

	val, evalErr := Eval(e, args[0])
	if evalErr != nil && catchDefined {
		binds := []types.Base{catch[1].(types.Symbol)}
		exprs := []types.Base{string(evalErr.Error())}
		if userErr, isUserErr := evalErr.(types.UserError); isUserErr {
			exprs = []types.Base{userErr.Val}
		}
		newEnv, err := e.Child(binds, exprs)
		if err != nil {
			return nil, err
		}
		return Eval(newEnv, catch[2])
	}

	return val, evalErr
}

func evalDefMacro(e types.Env, args ...types.Base) (types.Base, error) {
	val, err := evalDef(e, args...)
	if err != nil {
		return nil, err
	}
	fn, ok := val.(*types.ExtFunc)
	if !ok {
		return nil, fmt.Errorf("non-func value passed to defmacro")
	}
	fn.IsMacro = true
	return fn, nil
}

func evalDef(e types.Env, args ...types.Base) (types.Base, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("not enough arguments")
	}

	name, ok := args[0].(types.Symbol)
	if !ok {
		return nil, fmt.Errorf("non-symbol bind value")
	}
	value, err := Eval(e, args[1])
	if err == nil {
		e.Set(name, value)
	}
	return value, err
}

func evalLet(e types.Env, args ...types.Base) (types.Base, types.Env, error) {
	if len(args) < 2 {
		return nil, nil, fmt.Errorf("not enough arguments for  let* call")
	}
	newEnv, err := e.Child(nil, nil)
	if err != nil {
		return nil, nil, err
	}

	var definitions []types.Base
	switch lst := args[0].(type) {
	case types.Collection:
		definitions = lst.Data()
	default:
		return nil, nil, fmt.Errorf("invalid let* environment definition")
	}

	for i := 0; i < len(definitions); i += 2 {
		if _, err := evalDef(newEnv, definitions[i:]...); err != nil {
			return nil, nil, err
		}
	}

	return args[1], newEnv, nil
}

func evalDo(e types.Env, args ...types.Base) (types.Base, error) {
	_, err := evalAST(e, types.NewList(args[:len(args)-1]...))
	if err != nil {
		return nil, err
	}
	return args[len(args)-1], nil
}

func evalIf(e types.Env, args ...types.Base) (types.Base, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("improperly formatted if statement")
	}
	if condition, err := evalBool(e, args[0]); err != nil {
		return nil, err
	} else if condition {
		return args[1], nil
	} else if len(args) > 2 {
		return args[2], nil
	}
	return nil, nil
}

func evalBool(e types.Env, condition types.Base) (bool, error) {
	value, err := Eval(e, condition)
	if err != nil {
		return false, err
	}
	switch tVal := value.(type) {
	case bool:
		return tVal, nil
	case nil:
		return false, nil
	default:
		return true, nil
	}
}

func evalQuasiQuote(e types.Env, object types.Base) types.Base {
	pair, isp := isPair(object)
	if !isp {
		return types.NewList(types.Symbol("quote"), object)
	}

	if sym, ok := pair[0].(types.Symbol); ok && sym == "unquote" {
		return pair[1]
	} else if nextPair, isp := isPair(pair[0]); isp {
		if sym, ok := nextPair[0].(types.Symbol); ok && sym == "splice-unquote" {
			return types.NewList(types.Symbol("concat"), nextPair[1], evalQuasiQuote(e, types.NewList(pair[1:]...)))
		}
	}
	return types.NewList(types.Symbol("cons"), evalQuasiQuote(e, pair[0]), evalQuasiQuote(e, types.NewList(pair[1:]...)))
}

func evalListForms(values []types.Base, env types.Env) ([]types.Base, error) {
	var err error
	forms := make([]types.Base, len(values))
	for i, form := range values {
		forms[i], err = Eval(env, form)
		if err != nil {
			return forms, err
		}
	}
	return forms, nil
}

func isPair(val types.Base) ([]types.Base, bool) {
	lst, isList := val.(types.Collection)
	if !isList {
		return []types.Base{}, false
	}
	data := lst.Data()
	return data, len(data) > 0
}

func isMacroCall(e types.Env, ast types.Base) (*types.List, bool) {
	lst, isList := ast.(*types.List)
	if !isList || len(lst.Forms) == 0 {
		return lst, false
	}

	sym, ok := lst.Forms[0].(types.Symbol)
	if !ok {
		return lst, false
	}

	envVal, err := e.Get(sym)
	if err != nil {
		return lst, false
	}

	fn, ok := envVal.(*types.ExtFunc)
	if !ok {
		return lst, false
	}

	return lst, fn.IsMacro
}

func macroExpand(e types.Env, ast types.Base) (types.Base, error) {
	list, is := isMacroCall(e, ast)
	for ; is; list, is = isMacroCall(e, ast) {
		var err error
		sym := list.Forms[0].(types.Symbol)
		val, _ := e.Get(sym)
		ast, err = val.(*types.ExtFunc).Apply(list.Forms[1:])
		if err != nil {
			return nil, err
		}
	}
	return ast, nil
}
