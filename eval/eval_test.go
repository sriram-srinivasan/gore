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

func TestAliases(t *testing.T) {
	// Ensure that using p and t as variables or as function names doesn't incorrectly expand them
	code := `
            p  := 10
            p  p
            func t() int{
               return 100
            }
            t()
            t t()
        `
	check(t, code, "10\nint\n", "")
}

func TestPartitioning(t *testing.T) {
	code := `
          p "TestPartitioning"
          /* 
           */ imported := 10 // trying to fool it with leading string "import"
    /* */ import "strings"
          p strings.TrimLeft("foobar", "fo")
          typed := true  // starts with "type", but shouldn't be fooled into putting it at toplevel
          p typed
          type S struct{
              a int
              b bool
          }
          s := S{a: imported, b: typed}
          p s
         `
	check(t, code, "TestPartitioning\nbar\ntrue\n{a:10 b:true}", "")
}

func TestStrings(t *testing.T) {
	// Inside a double quoted string, it should be ok to have:
	//   1. expressions of the form abc.foo, where abc is not mistakenly interpreted to be a package name
	//   2. escaped double quotes
	//   3. Single quotes
	//   4. Comment-like characters which should not be interpreted as a comment
	//   5. Trailing '{' which shouldn't be interpreted either
	code := `x := "abc.def\"'ghi'//{"` + "\n"
	// test single quoted literal with escaped single quote
	code += `y := '\''` + "\n"
	// multiline raw string with comment-like chars which shouldn't be interpreted as comments
	code += "z := `multiline\n"
	code += "string /**/`; println(x,y,z)\n"

	out :=
		`abc.def"'ghi'//{ 39 multiline
string /**/`
	check(t, code, out, "")
}

// checks that comment chars inside strings are ignored, and that leading and trailing comments don't confuse paren/bracket accounting
func TestComments(t *testing.T) {
	code := `
           a := "/* test string {"
           if a != "" { // ....
               /* multiine
                   */
               println(a)
 /* ... */  } // ...
 /* ... */ import ( // ..
                 "math" 
 /*  */    )  // ...
             math.Log10(100)
           `
	check(t, code, "/* test string {", "")
}

// check that line numbers of compiler errors are not thrown off by multiline comments 
func TestCommentsErr(t *testing.T) {
	code := `
         a := /* {
            dummy comments
         */ xxx.Foo()
         p a
        `
	check(t, code, "", ":4: undefined: xxx")
}

var ts = strings.TrimSpace

func check(t *testing.T, code string, expected_out string, expected_err string) {
	out, err := eval.Eval(code)
	if !(ts(expected_out) == ts(out)) {
		t.Error(fmt.Sprintf("Expected output to be \n%s\nInstead got:\n%s\n", expected_out, out))
	}
	if !(ts(expected_err) == ts(err)) {
		t.Error(fmt.Sprintf("Expected compiler error to be \n%s\n. Instead got:\n%s\n", expected_err, err))
	}
}
