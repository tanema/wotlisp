package core

import (
	"github.com/tanema/mal/src/env"
	"github.com/tanema/mal/src/runtime"
	"github.com/tanema/mal/src/types"
)

// DefaultNamespace generate an evironment with the core function and variable declarations defined
func DefaultNamespace() *env.Env {
	defaultEnv, _ := env.New(nil, nil, nil)
	for method, fn := range namespace {
		defaultEnv.Set(method, fn)
	}
	defaultEnv.Set("eval", eval(defaultEnv))
	defaultEnv.Set("*host-language*", "wot")
	ev(defaultEnv, "(def! not (fn* (a) (if a false true)))")
	ev(defaultEnv, `(def! load-file (fn* (f) (eval (read-string (str "(do " (slurp f) ")")))))`)
	ev(defaultEnv, `(defmacro! cond (fn* (& xs) (if (> (count xs) 0) (list 'if (first xs) (if (> (count xs) 1) (nth xs 1) (throw "odd number of forms to cond")) (cons 'cond (rest (rest xs)))))))`)
	ev(defaultEnv, "(def! *gensym-counter* (atom 0))")
	ev(defaultEnv, "(def! gensym (fn* [] (symbol (str \"G__\" (swap! *gensym-counter* (fn* [x] (+ 1 x)))))))")
	ev(defaultEnv, "(defmacro! or (fn* (& xs) (if (empty? xs) nil (if (= 1 (count xs)) (first xs) (let* (condvar (gensym)) `(let* (~condvar ~(first xs)) (if ~condvar ~condvar (or ~@(rest xs)))))))))")
	return defaultEnv
}

func ev(e *env.Env, source string) {
	ast, parseErr := readString(e, []types.Base{source})
	if parseErr != nil {
		panic(parseErr)
	}
	evalFn, err := e.Get("eval")
	if err != nil {
		panic(err)
	}
	if _, err := evalFn.(*types.StdFunc).Fn(e, []types.Base{ast}); err != nil {
		panic(err)
	}
}

func eval(defaultEnv types.Env) *types.StdFunc {
	return types.Func(func(e types.Env, a []types.Base) (types.Base, error) {
		if len(a) < 1 {
			return nil, nil
		}
		return runtime.Eval(defaultEnv, a[0])
	})
}
