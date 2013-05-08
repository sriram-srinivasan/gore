package eval_test

import (
	"fmt"
	"github.com/sriram-srinivasan/gore/eval"
	"strings"
	"testing"
)

func TestSimple(t *testing.T) {
	code := `fmt.Println("gore test")`
	check(t, code, "gore test", "")
}
func TestSimpleErr(t *testing.T) {
	code := `mt.Println("gore test")`
	check(t, code, "", ":1: undefined: mt")
}
func TestMultiline(t *testing.T) {
	code := `
              import (
                  "regexp"                        // Testing explicit import
              )
              foo := 10
              bar := math.Log10(100)             // Testing implicit import
              type Point struct {
                    x, y int
              }
              pt := Point{y: 1000, x: 100}
              if (pt.x <= 100) {
                  p "Case 1", foo, bar, pt
              } else {
                                                  // Testing nested curlies
                  if (pt.y > 1000)  {
                       p "Case 2", foo, bar, pt
                  }
              }
              // Check type
              t Point{10,100}, 0.34343
              regexp.MustCompile("foobar")
        `
	check(t, code, "Case 1\n10\n2\n{x:100 y:1000}\nmain.Point\nfloat64", "")
}

func TestMultilineRawString(t *testing.T) {
	code := "fmt.Println(`gore raw string\nmultiline test`)"
	check(t, code, "gore raw string\nmultiline test", "")
}

func TestMultilineError(t *testing.T) {
	code := `
             foo := 10
             math.log(100) // Using log instead of Log to provoke error
        `
	check(t, code, "", ":3: cannot refer to unexported name math.log")
}

func TestImportRepair(t *testing.T) {
	// Using a var name ('math' here) that is identical to a standard package name. Eval should
	// import the "math" package and retry if the compiler complains about duplicate packages
	code := `
           type M struct{
               x int
           }
        math := M{100}
        p math.x
        `
	check(t, code, "100", "")
}

var ts = strings.TrimSpace

func check(t *testing.T, code string, expected_out string, expected_err string) {
	out, err := eval.Eval(code)
	if !(ts(expected_out) == ts(out)) {
		t.Error(fmt.Sprintf("Expected output to be:\n%s\n\nInstead got:\n%s\n", expected_out, out))
	}
	if !(ts(expected_err) == ts(err)) {
		t.Error(fmt.Sprintf("Expected compiler error to be:\n%s\n\nInstead got:\n%s\n", expected_err, err))
	}
}
