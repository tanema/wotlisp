package reader

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/tanema/mal/src/types"
)

var (
	// ErrUnderflow is thrown when params are not matched
	ErrUnderflow = errors.New("EOF underflow error: more input expected")

	tokensPattern = regexp.MustCompile(`[\s,]*(~@|[\[\]{}()'` + "`" + `~^@]|"(?:\\.|[^\\"])*"?|;.*|[^\s\[\]{}('"` + "`" + `,;)]*)`)
	numberPattern = regexp.MustCompile(`^-?[0-9]+\.?[0-9]*$`)

	stringEsc = map[string]string{
		`\\`: `\`,
		`\"`: `"`,
		`\n`: "\n",
	}
)

type reader struct {
	tokens []string
}

// ReadString will take in a source code string, tokenize it and then parse it,
// returning a value to be evaluated
func ReadString(in string) (types.Base, error) {
	rdr := &reader{tokens: tokenize(in)}
	return rdr.form()
}

func (rdr *reader) next() (string, bool) {
	var token string
	if len(rdr.tokens) == 0 {
		return token, false
	}
	token, rdr.tokens = rdr.tokens[0], rdr.tokens[1:]
	return token, true
}

func (rdr *reader) peek() (string, bool) {
	if len(rdr.tokens) > 0 {
		return rdr.tokens[0], true
	}
	return "", false
}

func tokenize(in string) []string {
	results := []string{}
	for _, group := range tokensPattern.FindAllStringSubmatch(in, -1) {
		if (group[1] == "") || (group[1][0] == ';') {
			continue
		}
		results = append(results, group[1])
	}
	return results
}

func (rdr *reader) form() (types.Base, error) {
	token, hasNext := rdr.peek()
	if !hasNext {
		return nil, ErrUnderflow
	}

	switch token {
	case `'`:
		return rdr.modifier("quote")
	case "`":
		return rdr.modifier("quasiquote")
	case `~`:
		return rdr.modifier("unquote")
	case `~@`:
		return rdr.modifier("splice-unquote")
	case `^`:
		return rdr.meta()
	case `@`:
		return rdr.modifier("deref")
	case ")":
		return nil, errors.New("unexpected ')'")
	case "(":
		return rdr.list("(", ")")
	case "]":
		return nil, errors.New("unexpected ']'")
	case "[":
		return rdr.vector()
	case "}":
		return nil, errors.New("unexpected '}'")
	case "{":
		return rdr.hashMap()
	default:
		return rdr.atom()
	}
}

func (rdr *reader) modifier(symbol string) (*types.List, error) {
	rdr.next()
	form, err := rdr.form()
	return types.NewList(types.Symbol(symbol), form), err
}

func (rdr *reader) meta() (*types.List, error) {
	rdr.next()
	meta, err := rdr.form()
	if err != nil {
		return nil, err
	}
	form, err := rdr.form()
	return types.NewList(types.Symbol("with-meta"), form, meta), err
}

func (rdr *reader) list(start, end string) (*types.List, error) {
	list := &types.List{Forms: []types.Base{}}
	token, hasNext := rdr.next()
	if !hasNext {
		return list, ErrUnderflow
	}
	if token != start {
		return list, fmt.Errorf("unexpected '%v'", token)
	}
	token, hasNext = rdr.peek()
	for ; token != end && hasNext; token, hasNext = rdr.peek() {
		form, err := rdr.form()
		if err != nil {
			return list, err
		}
		list.Forms = append(list.Forms, form)
	}
	if token, hasNext := rdr.next(); !hasNext {
		return list, ErrUnderflow
	} else if token != end {
		return list, fmt.Errorf("unexpected '%v'", token)
	}
	return list, nil
}

func (rdr *reader) vector() (*types.Vector, error) {
	list, err := rdr.list("[", "]")
	return &types.Vector{Forms: list.Forms}, err
}

func (rdr *reader) hashMap() (*types.Hashmap, error) {
	list, err := rdr.list("{", "}")
	if err != nil {
		return nil, err
	}
	return types.NewHashmap(list.Forms)
}

func (rdr *reader) atom() (types.Base, error) {
	token, hasNext := rdr.next()
	if !hasNext {
		return nil, ErrUnderflow
	}

	if token == "nil" {
		return nil, nil
	} else if token == "true" {
		return true, nil
	} else if token == "false" {
		return false, nil
	} else if token[0] == ':' {
		return types.Keyword(token[1:]), nil
	} else if match := numberPattern.MatchString(token); match {
		num, err := strconv.ParseFloat(token, 64)
		if err != nil {
			err = errors.New("improperly formatted number")
		}
		return num, nil
	} else if token[0] == '"' {
		if token[len(token)-1] != '"' {
			return nil, errors.New("expected '\"', got EOF")
		}
		str := token[1 : len(token)-1]
		// for find, replace := range stringEsc {
		//	str = strings.Replace(str, find, replace, -1)
		// }
		return strings.Replace(
			strings.Replace(
				strings.Replace(
					strings.Replace(str, `\\`, "\u029e", -1),
					`\"`, `"`, -1),
				`\n`, "\n", -1),
			"\u029e", "\\", -1), nil
		//return str, nil
	}

	return types.Symbol(token), nil
}
