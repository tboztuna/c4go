package ast

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/Konstantin8105/c4go/util"
)

func formatMultiLine(o interface{}) string {
	s := fmt.Sprintf("%#v", o)
	s = strings.Replace(s, "{", "{\n", -1)
	s = strings.Replace(s, ", ", "\n", -1)

	return s
}

func runNodeTests(t *testing.T, tests map[string]Node) {
	i := 1
	for line, expected := range tests {
		testName := fmt.Sprintf("Example%d", i)
		i++

		t.Run(testName, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("Panic for : %v, %v\n%v", testName, line, r)
				}
			}()
			// Append the name of the struct onto the front. This would make the
			// complete line it would normally be parsing.
			name := reflect.TypeOf(expected).Elem().Name()
			actual, err := Parse(name + " " + line)

			if !reflect.DeepEqual(expected, actual) {
				t.Errorf("%s", util.ShowDiff(formatMultiLine(expected),
					formatMultiLine(actual)))
			}
			if err != nil {
				t.Errorf("Error parsing %v", err)
			}
			if int64(actual.Address()) == 0 {
				t.Errorf("Address for test cannot be nil. %v", actual.Address())
			}
			if len(actual.Children()) != 0 {
				t.Errorf("Amount of children cannot be more 0")
			}
			actual.AddChild(nil)
			if len(actual.Children()) != 1 {
				t.Errorf("Amount of children must be 1")
			}
			if actual.Children()[0] != nil {
				t.Errorf("Children must bee nil")
			}
			pos := actual.Position()
			if pos.Line == 0 && pos.Column == 0 && pos.LineEnd == 0 &&
				pos.ColumnEnd == 0 && pos.File == "" {
				t.Log("Please try to change Position for test. " +
					"Better to test not zero position")
			}
			if pos.Line < 0 || pos.Column < 0 || pos.LineEnd < 0 ||
				pos.ColumnEnd < 0 {
				t.Errorf("Not acceptable negative position")
			}
			var posC Position
			posC.Line = -1
			if pos.Line == -1 {
				t.Fatalf("Not correct default line position")
			}
			setPosition(actual, posC)
			if pos.Line != -1 {
				t.Log("Cannot change position")
			}
		})
	}
}

func TestPrint(t *testing.T) {
	cond := &ConditionalOperator{}
	cond.AddChild(&ImplicitCastExpr{})
	cond.AddChild(&ImplicitCastExpr{})
	s := Atos(cond)
	if len(s) == 0 {
		t.Fatalf("Cannot convert AST tree : %#v", cond)
	}
	lines := strings.Split(s, "\n")
	var amount int
	for _, l := range lines {
		if strings.Contains(l, "ImplicitCastExpr") {
			amount++
		}
	}
	if amount != 2 {
		t.Error("Not correct design of output")
	}
}

func TestPanicCheck(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("panic for parsing string line is not acceptable. %v", r)
		}
	}()
	_, err := Parse("Some strange line")
	if err == nil {
		t.Errorf("Haven`t error for strange string line not acceptable")
	}
	// Correct node of AST:
	// GotoStmt 0x7fb9cc1994d8 <line:18893:9, col:14>  'end_getDigits' 0x7fb9cc199490
	// Modify for panic in ast regexp
	//
	n, err := Parse("GotoStmt 0x7fb9cc1994d8 <lin8893:9, col:14> ts' 99490")
	if err == nil {
		t.Errorf("Haven`t error for guarantee panic line not acceptable\n%v",
			Atos(n))
	}
}

func TestNullStmt(t *testing.T) {
	n, err := Parse("NullStmt")
	if n != nil || err != nil {
		t.Errorf("Not acceptable for NullStmt")
	}
}

type Visitor interface {
	Visit(node Node) (w Visitor)
}

type Founder struct{}

var nodesFromAst []string

func (f Founder) Visit(node ast.Node) (w ast.Visitor) {
	if cs, ok := node.(*ast.CaseClause); ok {
		nodesFromAst = append(nodesFromAst, cs.List[0].(*ast.BasicLit).Value)
	}
	return f
}

func TestAstNodes(t *testing.T) {
	fset := token.NewFileSet() // positions are relative to fset
	f, err := parser.ParseFile(fset, "ast.go", nil, parser.DeclarationErrors)
	if err != nil {
		t.Fatalf("%v", err)
	}

	var fr Founder
	ast.Walk(fr, f.Decls[4])

	/*
		108: *ast.CaseClause {
		.  Case: -
		.  List: []ast.Expr (len = 1) {
		.  .  0: *ast.BasicLit {
		.  .  .  ValuePos: -
		.  .  .  Kind: STRING
		.  .  .  Value: "\"WhileStmt\""
		.  .  }
		.  }
	*/

	nodesFromAst = append(nodesFromAst, "")

	for _, c := range nodesFromAst {
		t.Run(fmt.Sprintf("%v", c), func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("Cannot parse")
				}
			}()
			Parse(c)
		})
	}

	dat, err := ioutil.ReadFile("position.go")

	index := bytes.Index(dat, []byte("setPosition"))
	if index < 0 {
		t.Fatalf("cannot found function")
	}
	dat = dat[index:]

	for _, c := range nodesFromAst {
		if c == "" || c == "\"NullStmt\"" {
			continue
		}
		t.Run(fmt.Sprintf("%v", c), func(t *testing.T) {
			c, err := strconv.Unquote(c)
			if err != nil {
				t.Fatalf("Unquote invalid : %v", err)
			}
			index := bytes.Index(dat, []byte(c))
			if index < 0 {
				t.Fatalf("cannot found type : %v", c)
			}
		})
	}
}
