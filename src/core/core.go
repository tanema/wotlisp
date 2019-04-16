package core

import (
	"errors"
	"fmt"
	"io/ioutil"
	"reflect"
	"strings"
	"time"

	"github.com/tanema/mal/src/printer"
	"github.com/tanema/mal/src/reader"
	"github.com/tanema/mal/src/readline"
	"github.com/tanema/mal/src/types"
)

var namespace = map[types.Symbol]*types.StdFunc{
	"+":           types.Func(add),
	"-":           types.Func(sub),
	"*":           types.Func(mul),
	"/":           types.Func(div),
	"=":           types.Func(equal),
	"<":           types.Func(lessThan),
	"<=":          types.Func(lessThanEqual),
	">":           types.Func(greaterThan),
	">=":          types.Func(greaterThanEqual),
	"prn":         types.Func(prn),
	"println":     types.Func(prnln),
	"pr-str":      types.Func(prnstr),
	"str":         types.Func(str),
	"list":        types.Func(list),
	"list?":       types.Func(islist),
	"empty?":      types.Func(isempty),
	"count":       types.Func(count),
	"read-string": types.Func(readString),
	"slurp":       types.Func(slurp),
	"atom":        types.Func(atom),
	"atom?":       types.Func(isatom),
	"deref":       types.Func(deref),
	"reset!":      types.Func(reset),
	"swap!":       types.Func(swap),
	"cons":        types.Func(cons),
	"concat":      types.Func(concat),
	"nth":         types.Func(nth),
	"first":       types.Func(first),
	"rest":        types.Func(rest),
	"throw":       types.Func(throw),
	"apply":       types.Func(apply),
	"map":         types.Func(mapvals),
	"nil?":        types.Func(isnil),
	"true?":       types.Func(istrue),
	"false?":      types.Func(isfalse),
	"symbol?":     types.Func(issymbol),
	"symbol":      types.Func(makesymbol),
	"keyword?":    types.Func(iskeyword),
	"keyword":     types.Func(makekeyword),
	"vector?":     types.Func(isvector),
	"vector":      types.Func(makevector),
	"map?":        types.Func(ismap),
	"hash-map":    types.Func(makemap),
	"assoc":       types.Func(assoc),
	"dissoc":      types.Func(dissoc),
	"get":         types.Func(get),
	"contains?":   types.Func(contains),
	"keys":        types.Func(keys),
	"vals":        types.Func(vals),
	"sequential?": types.Func(sequential),
	"readline":    types.Func(rdline),
	"meta":        types.Func(meta),
	"with-meta":   types.Func(withmeta),
	"string?":     types.Func(isstring),
	"number?":     types.Func(isnumber),
	"fn?":         types.Func(isfn),
	"macro?":      types.Func(ismacro),
	"conj":        types.Func(conj),
	"seq":         types.Func(seq),
	"time-ms":     types.Func(timems),
}

func timems(e types.Env, a []types.Base) (types.Base, error) {
	return float64(time.Now().UnixNano() / int64(time.Millisecond)), nil
}

func conj(e types.Env, a []types.Base) (types.Base, error) {
	if len(a) < 2 {
		return nil, errors.New("wrong number of arguments")
	}
	switch v := a[0].(type) {
	case *types.List:
		result := []types.Base{}
		for _, v := range a[1:] {
			result = append([]types.Base{v}, result...)
		}
		return types.NewList(append(result, v.Forms...)...), nil
	case *types.Vector:
		return types.NewVect(append(v.Forms, a[1:]...)...), nil
	default:
		return nil, nil
	}
}

func seq(e types.Env, a []types.Base) (types.Base, error) {
	if err := assertArgNum(a, 1); err != nil {
		return nil, err
	}
	switch v := a[0].(type) {
	case types.Collection:
		data := v.Data()
		if len(data) == 0 {
			return nil, nil
		}
		return types.NewList(data...), nil
	case string:
		if v == "" {
			return nil, nil
		}
		result := make([]types.Base, len(v))
		for i, ch := range strings.Split(v, "") {
			result[i] = ch
		}
		return types.NewList(result...), nil
	default:
		return nil, nil
	}
}

func isstring(e types.Env, a []types.Base) (types.Base, error) {
	if err := assertArgNum(a, 1); err != nil {
		return nil, err
	}
	switch a[0].(type) {
	case string:
		return true, nil
	default:
		return false, nil
	}
}

func isnumber(e types.Env, a []types.Base) (types.Base, error) {
	if err := assertArgNum(a, 1); err != nil {
		return nil, err
	}
	switch a[0].(type) {
	case float64:
		return true, nil
	default:
		return false, nil
	}
}
func isfn(e types.Env, a []types.Base) (types.Base, error) {
	if err := assertArgNum(a, 1); err != nil {
		return nil, err
	}
	switch fn := a[0].(type) {
	case *types.StdFunc:
		return true, nil
	case *types.ExtFunc:
		return !fn.IsMacro, nil
	default:
		return false, nil
	}
}

func ismacro(e types.Env, a []types.Base) (types.Base, error) {
	if err := assertArgNum(a, 1); err != nil {
		return nil, err
	}
	switch fn := a[0].(type) {
	case *types.ExtFunc:
		return fn.IsMacro, nil
	default:
		return false, nil
	}
}

func rdline(e types.Env, a []types.Base) (types.Base, error) {
	prompt := ""
	if len(a) > 0 {
		if p, isstring := a[0].(string); isstring {
			prompt = p
		}
	}
	return readline.Readline(prompt)
}

func meta(e types.Env, a []types.Base) (types.Base, error) {
	if err := assertArgNum(a, 1); err != nil {
		return nil, err
	}
	switch val := a[0].(type) {
	case *types.List:
		return val.Meta, nil
	case *types.Vector:
		return val.Meta, nil
	case *types.Hashmap:
		return val.Meta, nil
	case *types.StdFunc:
		return val.Meta, nil
	case *types.ExtFunc:
		return val.Meta, nil
	default:
		return nil, nil
	}
}

func withmeta(e types.Env, a []types.Base) (types.Base, error) {
	if err := assertArgNum(a, 2); err != nil {
		return nil, err
	}
	switch val := a[0].(type) {
	case *types.List:
		list := types.NewList(val.Forms...)
		list.Meta = a[1]
		return list, nil
	case *types.Vector:
		vect := types.NewVect(val.Forms...)
		vect.Meta = a[1]
		return vect, nil
	case *types.Hashmap:
		hmap, _ := types.NewHashmap(val.ToList())
		hmap.Meta = a[1]
		return hmap, nil
	case *types.StdFunc:
		clonedFn := types.Func(val.Fn)
		clonedFn.Meta = a[1]
		return clonedFn, nil
	case *types.ExtFunc:
		clonedFn := val.Clone()
		clonedFn.Meta = a[1]
		return clonedFn, nil
	default:
		return nil, errors.New("invalid data type for metadata")
	}
}

func assoc(e types.Env, a []types.Base) (types.Base, error) {
	if len(a) < 3 {
		return nil, errors.New("wrong number of arguments")
	}
	hmap, isHmap := a[0].(*types.Hashmap)
	if !isHmap {
		return nil, errors.New("cannot assoc with non-hashmap")
	}
	return types.NewHashmap(append(hmap.ToList(), a[1:]...))
}

func dissoc(e types.Env, a []types.Base) (types.Base, error) {
	if len(a) < 2 {
		return nil, errors.New("wrong number of arguments")
	}
	hmap, isHmap := a[0].(*types.Hashmap)
	if !isHmap {
		return nil, errors.New("cannot dissoc with non-hashmap")
	}
	return types.NewHashmap(hmap.ToList(), a[1:]...)
}

func get(e types.Env, a []types.Base) (types.Base, error) {
	if err := assertArgNum(a, 2); err != nil {
		return nil, err
	}
	hmap, isHmap := a[0].(*types.Hashmap)
	if !isHmap {
		return nil, nil
	}
	return hmap.Forms[a[1]], nil
}

func contains(e types.Env, a []types.Base) (types.Base, error) {
	if err := assertArgNum(a, 2); err != nil {
		return nil, err
	}
	hmap, isHmap := a[0].(*types.Hashmap)
	if !isHmap {
		return nil, errors.New("cannot contains? with non-hashmap")
	}
	_, found := hmap.Forms[a[1]]
	return found, nil
}

func keys(e types.Env, a []types.Base) (types.Base, error) {
	if err := assertArgNum(a, 1); err != nil {
		return nil, err
	}
	hmap, isHmap := a[0].(*types.Hashmap)
	if !isHmap {
		return nil, errors.New("cannot index keys with non-hashmap")
	}
	return types.NewList(hmap.Keys()...), nil
}

func vals(e types.Env, a []types.Base) (types.Base, error) {
	if err := assertArgNum(a, 1); err != nil {
		return nil, err
	}
	hmap, isHmap := a[0].(*types.Hashmap)
	if !isHmap {
		return nil, errors.New("cannot index keys with non-hashmap")
	}
	return types.NewList(hmap.Vals()...), nil
}

func sequential(e types.Env, a []types.Base) (types.Base, error) {
	if err := assertArgNum(a, 1); err != nil {
		return nil, err
	}

	switch a[0].(type) {
	case types.Collection:
		return true, nil
	default:
		return false, nil
	}
}

func throw(e types.Env, a []types.Base) (types.Base, error) {
	if len(a) == 0 {
		return nil, fmt.Errorf("standard error")
	}
	return nil, types.UserError{Val: a[0]}
}

func isnil(e types.Env, a []types.Base) (types.Base, error) {
	if err := assertArgNum(a, 1); err != nil {
		return nil, err
	}
	return a[0] == nil, nil
}

func istrue(e types.Env, a []types.Base) (types.Base, error) {
	if err := assertArgNum(a, 1); err != nil {
		return nil, err
	}
	val, isbool := a[0].(bool)
	return isbool && val, nil
}

func isfalse(e types.Env, a []types.Base) (types.Base, error) {
	if err := assertArgNum(a, 1); err != nil {
		return nil, err
	}
	val, isbool := a[0].(bool)
	return !isbool || !val, nil
}

func issymbol(e types.Env, a []types.Base) (types.Base, error) {
	if err := assertArgNum(a, 1); err != nil {
		return nil, err
	}
	_, isSymbol := a[0].(types.Symbol)
	return isSymbol, nil
}

func makesymbol(e types.Env, a []types.Base) (types.Base, error) {
	if err := assertArgNum(a, 1); err != nil {
		return nil, err
	}
	val, isString := a[0].(string)
	if !isString {
		return nil, errors.New("cannot create symbol with non-string")
	}
	return types.Symbol(val), nil
}

func iskeyword(e types.Env, a []types.Base) (types.Base, error) {
	if err := assertArgNum(a, 1); err != nil {
		return nil, err
	}
	_, isKey := a[0].(types.Keyword)
	return isKey, nil
}

func makekeyword(e types.Env, a []types.Base) (types.Base, error) {
	if err := assertArgNum(a, 1); err != nil {
		return nil, err
	}
	val, isString := a[0].(string)
	if !isString {
		return nil, errors.New("cannot create keyword with non-string")
	}
	return types.Keyword(val), nil
}

func isvector(e types.Env, a []types.Base) (types.Base, error) {
	if err := assertArgNum(a, 1); err != nil {
		return nil, err
	}
	_, isVect := a[0].(*types.Vector)
	return isVect, nil
}

func makevector(e types.Env, a []types.Base) (types.Base, error) {
	return types.NewVect(a...), nil
}

func ismap(e types.Env, a []types.Base) (types.Base, error) {
	if err := assertArgNum(a, 1); err != nil {
		return nil, err
	}
	_, isMap := a[0].(*types.Hashmap)
	return isMap, nil
}

func makemap(e types.Env, a []types.Base) (types.Base, error) {
	return types.NewHashmap(a)
}

func apply(e types.Env, a []types.Base) (types.Base, error) {
	if len(a) < 2 {
		return nil, fmt.Errorf("not enough arugments")
	}
	final := []types.Base{}
	for _, val := range a[1:] {
		if col, isCol := val.(types.Collection); isCol {
			final = append(final, col.Data()...)
		} else {
			final = append(final, val)
		}
	}
	return types.CallFunc(e, a[0], final)
}

func mapvals(e types.Env, a []types.Base) (types.Base, error) {
	if err := assertArgNum(a, 2); err != nil {
		return nil, err
	}
	col, ok := a[1].(types.Collection)
	if !ok {
		return nil, fmt.Errorf("invalid collection")
	}

	final := []types.Base{}
	for _, val := range col.Data() {
		val, err := types.CallFunc(e, a[0], []types.Base{val})
		if err != nil {
			return nil, err
		}
		final = append(final, val)
	}

	return types.NewList(final...), nil
}

func nth(e types.Env, a []types.Base) (types.Base, error) {
	if err := assertArgNum(a, 2); err != nil {
		return nil, err
	}
	col, ok := a[0].(types.Collection)
	if !ok {
		return nil, fmt.Errorf("cannot get the nth part of non collection")
	}
	n, ok := a[1].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid value to index on collection")
	}
	data := col.Data()
	if len(data) <= int(n) {
		return nil, fmt.Errorf("index out of bounds")
	}
	return data[int(n)], nil
}

func first(e types.Env, a []types.Base) (types.Base, error) {
	if err := assertArgNum(a, 1); err != nil {
		return nil, nil
	}
	col, ok := a[0].(types.Collection)
	if !ok {
		return nil, nil
	}
	data := col.Data()
	if len(data) == 0 {
		return nil, nil
	}
	return data[0], nil
}

func rest(e types.Env, a []types.Base) (types.Base, error) {
	if err := assertArgNum(a, 1); err != nil {
		return nil, err
	}
	col, ok := a[0].(types.Collection)
	if !ok {
		return types.NewList(), nil
	}
	data := col.Data()
	if len(data) == 0 {
		return types.NewList(), nil
	}
	return types.NewList(data[1:]...), nil
}

func cons(e types.Env, a []types.Base) (types.Base, error) {
	if err := assertArgNum(a, 2); err != nil {
		return nil, err
	}
	col, ok := a[1].(types.Collection)
	if !ok {
		return nil, fmt.Errorf("cannot cons a non list")
	}
	return types.NewList(append([]types.Base{a[0]}, col.Data()...)...), nil
}

func concat(e types.Env, a []types.Base) (types.Base, error) {
	final := []types.Base{}
	for _, elm := range a {
		col, ok := elm.(types.Collection)
		if !ok {
			return nil, fmt.Errorf("cannot cons a non list")
		}
		final = append(final, col.Data()...)
	}
	return types.NewList(final...), nil
}

func atom(e types.Env, a []types.Base) (types.Base, error) {
	if err := assertArgNum(a, 1); err != nil {
		return nil, err
	}
	return &types.Atom{Val: a[0]}, nil
}

func isatom(e types.Env, a []types.Base) (types.Base, error) {
	if err := assertArgNum(a, 1); err != nil {
		return nil, err
	}
	_, is := a[0].(*types.Atom)
	return is, nil
}

func deref(e types.Env, a []types.Base) (types.Base, error) {
	if err := assertArgNum(a, 1); err != nil {
		return nil, err
	}
	atom, ok := a[0].(*types.Atom)
	if !ok {
		return nil, errors.New("value is not atom")
	}
	return atom.Val, nil
}

func reset(e types.Env, a []types.Base) (types.Base, error) {
	if err := assertArgNum(a, 2); err != nil {
		return nil, err
	}
	atom, ok := a[0].(*types.Atom)
	if !ok {
		return nil, errors.New("value is not atom")
	}
	atom.Val = a[1]
	return atom.Val, nil
}

func swap(e types.Env, a []types.Base) (types.Base, error) {
	if len(a) < 2 {
		return nil, errors.New("wrong number of arguments")
	}
	atom, ok := a[0].(*types.Atom)
	if !ok {
		return nil, errors.New("value is not atom")
	}
	arguments := append([]types.Base{atom.Val}, a[2:]...)
	value, err := types.CallFunc(e, a[1], arguments)
	atom.Val = value
	return value, err
}

func readString(e types.Env, a []types.Base) (types.Base, error) {
	if err := assertArgNum(a, 1); err != nil {
		return nil, err
	}
	source, ok := a[0].(string)
	if !ok {
		return nil, errors.New("cannot read source from non-string")
	}
	return reader.ReadString(source)
}

func slurp(e types.Env, a []types.Base) (types.Base, error) {
	path, ok := a[0].(string)
	if !ok {
		return nil, errors.New("cannot read source from non-string path")
	}
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("problem reading source file: %v", err)
	}
	return string(b), err
}

func prn(e types.Env, a []types.Base) (types.Base, error) {
	fmt.Println(printer.List(a, true, "", "", " "))
	return nil, nil
}

func prnln(e types.Env, a []types.Base) (types.Base, error) {
	fmt.Println(printer.List(a, false, "", "", " "))
	return nil, nil
}

func prnstr(e types.Env, a []types.Base) (types.Base, error) {
	return printer.List(a, true, "", "", " "), nil
}

func str(e types.Env, a []types.Base) (types.Base, error) {
	return printer.List(a, false, "", "", ""), nil
}

func list(e types.Env, a []types.Base) (types.Base, error) {
	return types.NewList(a...), nil
}

func islist(e types.Env, a []types.Base) (types.Base, error) {
	if len(a) == 0 {
		return false, errors.New("not enough arguments to list?")
	}
	_, lst := a[0].(*types.List)
	return lst, nil
}

func isempty(e types.Env, a []types.Base) (types.Base, error) {
	if len(a) == 0 {
		return false, errors.New("not enough arguments to empty?")
	}
	switch data := a[0].(type) {
	case types.Collection:
		return len(data.Data()) == 0, nil
	case *types.Hashmap:
		return len(data.Forms) == 0, nil
	case nil:
		return true, nil
	default:
		return false, errors.New("invalid data type passed to empty?")
	}
}

func count(e types.Env, a []types.Base) (types.Base, error) {
	if len(a) == 0 {
		return false, errors.New("nothing to count")
	}
	switch data := a[0].(type) {
	case types.Collection:
		return float64(len(data.Data())), nil
	case *types.Hashmap:
		return float64(len(data.Forms)), nil
	case string:
		return float64(len(data)), nil
	case nil:
		return float64(0), nil
	default:
		return false, errors.New("invalid data type passed to count")
	}
}

func equal(e types.Env, a []types.Base) (types.Base, error) {
	if len(a) < 2 {
		return false, errors.New("not enough arguments to equal")
	}
	return checkEquality(a[0], a[1])
}

func checkEquality(val1, val2 types.Base) (bool, error) {
	switch data := val1.(type) {
	case types.Collection:
		other, ok := val2.(types.Collection)
		if !ok {
			return false, nil
		}
		return equalLists(data.Data(), other.Data())
	default:
	}

	if reflect.TypeOf(val1) != reflect.TypeOf(val2) {
		return false, nil
	}

	switch data := val1.(type) {
	case *types.Hashmap:
		other := val2.(*types.Hashmap)
		return equalMaps(data.Forms, other.Forms)
	case types.Symbol:
		other := val2.(types.Symbol)
		return data == other, nil
	case types.Keyword:
		other := val2.(types.Keyword)
		return data == other, nil
	case string:
		other := val2.(string)
		return data == other, nil
	case float64:
		other := val2.(float64)
		return data == other, nil
	case bool:
		other := val2.(bool)
		return data == other, nil
	case nil:
		return true, nil
	default:
		return false, errors.New("invalid data type passed to equal")
	}
}

func equalLists(lst1, lst2 []types.Base) (bool, error) {
	if len(lst1) != len(lst2) {
		return false, nil
	}

	for i, elm := range lst1 {
		if equal, err := checkEquality(elm, lst2[i]); err != nil {
			return false, err
		} else if !equal {
			return false, nil
		}
	}

	return true, nil
}

func equalMaps(m1, m2 map[types.Base]types.Base) (bool, error) {
	if len(m1) != len(m2) {
		return false, nil
	}

	for key, val := range m1 {
		other, found := m2[key]
		if !found {
			return false, nil
		}

		if equal, err := checkEquality(val, other); err != nil {
			return false, err
		} else if !equal {
			return false, nil
		}
	}

	return true, nil
}

func prepareCompare(args []types.Base) (float64, float64, error) {
	if len(args) < 2 {
		return 0, 0, errors.New("not enough arguments to equal")
	}
	val1, ok := args[0].(float64)
	if !ok {
		return 0, 0, errors.New("cannot compare non-number values")
	}
	val2, ok := args[1].(float64)
	if !ok {
		return 0, 0, errors.New("cannot compare non-number values")
	}
	return val1, val2, nil
}

func lessThan(e types.Env, a []types.Base) (types.Base, error) {
	v1, v2, err := prepareCompare(a)
	if err != nil {
		return nil, err
	}
	return v1 < v2, nil
}

func lessThanEqual(e types.Env, a []types.Base) (types.Base, error) {
	v1, v2, err := prepareCompare(a)
	if err != nil {
		return nil, err
	}
	return v1 <= v2, nil
}

func greaterThan(e types.Env, a []types.Base) (types.Base, error) {
	v1, v2, err := prepareCompare(a)
	if err != nil {
		return nil, err
	}
	return v1 > v2, nil
}

func greaterThanEqual(e types.Env, a []types.Base) (types.Base, error) {
	v1, v2, err := prepareCompare(a)
	if err != nil {
		return nil, err
	}
	return v1 >= v2, nil
}

func add(e types.Env, a []types.Base) (types.Base, error) {
	if e := assertArgNum(a, 2); e != nil {
		return nil, e
	}
	return a[0].(float64) + a[1].(float64), nil
}

func sub(e types.Env, a []types.Base) (types.Base, error) {
	if e := assertArgNum(a, 2); e != nil {
		return nil, e
	}
	return a[0].(float64) - a[1].(float64), nil
}

func mul(e types.Env, a []types.Base) (types.Base, error) {
	if e := assertArgNum(a, 2); e != nil {
		return nil, e
	}
	return a[0].(float64) * a[1].(float64), nil
}

func div(e types.Env, a []types.Base) (types.Base, error) {
	if e := assertArgNum(a, 2); e != nil {
		return nil, e
	}
	return a[0].(float64) / a[1].(float64), nil
}

func assertArgNum(a []types.Base, expectedLen int) error {
	if len(a) != expectedLen {
		return errors.New("wrong number of arguments")
	}
	return nil
}
