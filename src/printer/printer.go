package printer

import (
	"fmt"
	"strings"

	"github.com/tanema/mal/src/types"
)

// List nicely prints collection type values
func List(forms []types.Base, pretty bool, pre, post, join string) string {
	strList := make([]string, len(forms))
	for i, e := range forms {
		strList[i] = Print(e, pretty)
	}
	return pre + strings.Join(strList, join) + post
}

// Print takes any value type and formats into a pleasant string. If the value
// type is unrecognized then an error if logged
func Print(object types.Base, pretty bool) string {
	switch tobj := object.(type) {
	case *types.Vector:
		return List(tobj.Forms, pretty, "[", "]", " ")
	case *types.List:
		return List(tobj.Forms, pretty, "(", ")", " ")
	case *types.Hashmap:
		return List(tobj.ToList(), pretty, "{", "}", " ")
	case types.Symbol:
		return string(tobj)
	case types.Keyword:
		return ":" + string(tobj)
	case *types.StdFunc:
		return "#<std::function>"
	case *types.ExtFunc:
		pre := "#<function "
		if tobj.IsMacro {
			pre = "#<macro "
		}
		return pre + List(tobj.Params, pretty, "[", "]", ", ") + Print(tobj.AST, pretty) + ">"
	case *types.Atom:
		return "(atom " + Print(tobj.Val, pretty) + ")"
	case types.UserError:
		return "Exception: " + Print(tobj.Val, pretty)
	case error:
		return "Exception: " + tobj.Error()
	case string:
		if pretty {
			tobj = strings.Replace(tobj, `\`, `\\`, -1)
			tobj = strings.Replace(tobj, `"`, `\"`, -1)
			tobj = strings.Replace(tobj, "\n", `\n`, -1)
			return `"` + tobj + `"`
		}
		return tobj
	case bool:
		if tobj {
			return "true"
		}
		return "false"
	case nil:
		return "nil"
	case float64:
		return fmt.Sprintf("%v", tobj)
	default:
		return "error formatting datatype"
	}
}
