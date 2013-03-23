package eval

/* 
 repl is an attempt to 
 repl provides a single function, Eval, that "evaluates" its argument. See documentation for Eval for more details
*/

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

/* 
 Wraps a chunk of code into a running program.

 All lines that are not type, func or import declarations are moved
 into a main function, and appropriate import statements are added 
 based on a superficial pattern match. (If you use fmt.Println, it'll
 add 'import "fmt"' automatically, if it is not already added.
*/
type strmap map[string]string

var builtinPkgs = map[string]string{
	"adler32":   "hash/adler32",
	"aes":       "crypto/aes",
	"ascii85":   "encoding/ascii85",
	"asn1":      "encoding/asn1",
	"ast":       "go/ast",
	"atomic":    "sync/atomic",
	"base32":    "encoding/base32",
	"base64":    "encoding/base64",
	"big":       "math/big",
	"binary":    "encoding/binary",
	"bufio":     "bufio",
	"build":     "go/build",
	"bytes":     "bytes",
	"bzip2":     "compress/bzip2",
	"cgi":       "net/http/cgi",
	"cgo":       "runtime/cgo",
	"cipher":    "crypto/cipher",
	"cmplx":     "math/cmplx",
	"color":     "image/color",
	"crc32":     "hash/crc32",
	"crc64":     "hash/crc64",
	"crypto":    "crypto",
	"csv":       "encoding/csv",
	"debug":     "runtime/debug",
	"des":       "crypto/des",
	"doc":       "go/doc",
	"draw":      "image/draw",
	"driver":    "database/sql/driver",
	"dsa":       "crypto/dsa",
	"dwarf":     "debug/dwarf",
	"ecdsa":     "crypto/ecdsa",
	"elf":       "debug/elf",
	"elliptic":  "crypto/elliptic",
	"errors":    "errors",
	"exec":      "os/exec",
	"expvar":    "expvar",
	"fcgi":      "net/http/fcgi",
	"filepath":  "path/filepath",
	"flag":      "flag",
	"flate":     "compress/flate",
	"fmt":       "fmt",
	"fnv":       "hash/fnv",
	"gif":       "image/gif",
	"gob":       "encoding/gob",
	"gosym":     "debug/gosym",
	"gzip":      "compress/gzip",
	"hash":      "hash",
	"heap":      "container/heap",
	"hex":       "encoding/hex",
	"hmac":      "crypto/hmac",
	"html":      "html",
	"http":      "net/http",
	"httputil":  "net/http/httputil",
	"image":     "image",
	"io":        "io",
	"ioutil":    "io/ioutil",
	"jpeg":      "image/jpeg",
	"json":      "encoding/json",
	"jsonrpc":   "net/rpc/jsonrpc",
	"list":      "container/list",
	"log":       "log",
	"lzw":       "compress/lzw",
	"macho":     "debug/macho",
	"mail":      "net/mail",
	"math":      "math",
	"md5":       "crypto/md5",
	"mime":      "mime",
	"multipart": "mime/multipart",
	"net":       "net",
	"os":        "os",
	"parse":     "text/template/parse",
	"parser":    "go/parser",
	"path":      "path",
	"pe":        "debug/pe",
	"pem":       "encoding/pem",
	"pkix":      "crypto/x509/pkix",
	"png":       "image/png",
	"pprof":     "net/http/pprof",
	//"pprof": "runtime/pprof",
	"printer": "go/printer",
	//"rand": "crypto/rand",
	"rand":    "math/rand",
	"rc4":     "crypto/rc4",
	"reflect": "reflect",
	"regexp":  "regexp",
	"ring":    "container/ring",
	"rpc":     "net/rpc",
	"rsa":     "crypto/rsa",
	"runtime": "runtime",
	//"scanner": "go/scanner",
	"scanner":     "text/scanner",
	"sha1":        "crypto/sha1",
	"sha256":      "crypto/sha256",
	"sha512":      "crypto/sha512",
	"signal":      "os/signal",
	"smtp":        "net/smtp",
	"sort":        "sort",
	"sql":         "database/sql",
	"strconv":     "strconv",
	"strings":     "strings",
	"subtle":      "crypto/subtle",
	"suffixarray": "index/suffixarray",
	"sync":        "sync",
	"syntax":      "regexp/syntax",
	"syscall":     "syscall",
	"syslog":      "log/syslog",
	"tabwriter":   "text/tabwriter",
	"tar":         "archive/tar",
	//"template": "html/template",
	//"template": "text/template",
	"textproto": "net/textproto",
	"time":      "time",
	"tls":       "crypto/tls",
	"token":     "go/token",
	"unicode":   "unicode",
	"unsafe":    "unsafe",
	"url":       "net/url",
	"user":      "os/user",
	"utf16":     "unicode/utf16",
	"utf8":      "unicode/utf8",
	"x509":      "crypto/x509",
	"xml":       "encoding/xml",
	"zip":       "archive/zip",
	"zlib":      "compress/zlib",
}

func Eval(code string) (out string, err string) {
	// No additional wrapping if it has a package declaration already
	if ok, _ := regexp.MatchString("^ *package ", code); ok {
		out, err = run(code)
		return out, err
	}

	code = expandAliases(code)
	pkgsToImport := inferPackages(code)
	code = embedLineNumbers(code)
	global, nonGlobal := partition(code)
	return buildAndExec(global, nonGlobal, pkgsToImport)
}

func expandAliases(code string) string {
	// Expand "p foo(), 2*3"   to fmt.Println(foo(), 2*3)
	r := regexp.MustCompile(`(?m)^\s*p +(.*)$`)
	return string(r.ReplaceAll([]byte(code), []byte(" fmt.Println($1)")))
}

// Each line of the original source is tagged with a line number at the end like so: /*#10#*/
// Since the wrapping process adds import statements and rearranges global and non-global 
// statements (see partition), this embedding permits us to map compiler error numbers back
// to the original source
func embedLineNumbers(code string) string {
	lineNum := 0
	r := regexp.MustCompile("\n")
	return r.ReplaceAllStringFunc(code,
		func(string) string {
			lineNum++
			return fmt.Sprintf("//#%d\n", lineNum)
		})
}

// split code into global and non-global chunks. non-global chunks belong inside
// a main function, and global chunks refer to type, func and import declarations
func partition(code string) (global string, nonGlobal string) {
	r := regexp.MustCompile("^ *(func|type|import)")
	pos := 0 // Always maintained as the position from where to restart search
	for {
		chunk := nextChunk(code[pos:])
		if len(chunk) == 0 {
			break
		}
		if r.FindString(chunk) == "" { // not import, type or func decl. 
			nonGlobal += chunk
		} else {
			global += chunk
		}
		pos += len(chunk)
	}
	return
}

var pkgPattern = regexp.MustCompile(`[a-z]\w+\.`)

func inferPackages(chunk string) (pkgsToImport map[string]bool) {
	pkgsToImport = make(map[string]bool) // used as a set
	pkgs := pkgPattern.FindAllString(chunk, 100000)
	for _, pkg := range pkgs {
		pkg = pkg[:len(pkg)-1] // remove trailing '.'
		if importPkg, ok := builtinPkgs[pkg]; ok {
			pkgsToImport[importPkg] = true
		}
	}
	return pkgsToImport
}

func buildAndExec(global string, nonGlobal string, pkgsToImport map[string]bool) (out string, err string) {
	src := buildMain(global, nonGlobal, pkgsToImport)
	out, err = run(src)
	if err != "" {
		if repairImports(err, pkgsToImport) {
			src = buildMain(global, nonGlobal, pkgsToImport)
			out, err = run(src)
		}
	}
	return out, err
}

func repairImports(err string, pkgsToImport map[string]bool) (dupsDetected bool) {
	// Look for compile errors of the form
	// "test.go:10: xxx redeclared as imported package name"
	// and remove 'xxx' from pkgsToImport
	dupsDetected = false
	r := regexp.MustCompile(`(\w+) redeclared as imported package name`)
	for _, match := range r.FindAllStringSubmatch(err, 100000) {
		pkg := match[1]
		//fmt.Println("===== Removing " + pkg + " and retrying ....")
		delete(pkgsToImport, pkg)
		dupsDetected = true
	}
	return dupsDetected
}

func run(src string) (output string, err string) {
	src, newToOldLineNums := extractLineNumbers(src)
	tmpfile := save(src)
	cmd := exec.Command("go", "run", tmpfile)
	out, e := cmd.CombinedOutput()

	if e != nil {
		err = string(out)
		return "", remapCompileErrorLines(err, newToOldLineNums)
	} else {
		return string(out), ""
	}
	return "", ""
}

func remapCompileErrorLines(err string, newToOldLineNums map[int]int) string {
	ret := ""
	r := regexp.MustCompile(`^.*?:(\d+):`)
	for _, line := range strings.Split(err, "\n") {
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		if m := r.FindStringSubmatchIndex(line); m != nil {
			newLine, err := strconv.Atoi(line[m[2]:m[3]]) // The $1 slice
			if err != nil {
				panic("Unable to convert " + line[m[2]:m[3]])
			}
			oldLine := newToOldLineNums[newLine]
			if oldLine == 0 {
				panic(fmt.Sprintf("map to %d not found", newLine))
			}
			ret += fmt.Sprintf("%d:%s", oldLine, line[(m[3]+1):])
		} else {
			ret += line
		}
	}
	return ret
}

func extractLineNumbers(src string) (srcNoLineNums string, newToOldLineNums map[int]int) {
	newToOldLineNums = make(map[int]int)
	r := regexp.MustCompile(`//#(\d+)$`)
	for newLineNum, line := range strings.Split(src, "\n") {
		if m := r.FindStringSubmatch(line); m != nil {
			oldLineNum, _ := strconv.Atoi(m[1])
			newToOldLineNums[newLineNum+1] = oldLineNum // compiler errors are 1-based
		}
	}
	srcNoLineNums = r.ReplaceAllString(src, "") // remove line number annotations
	return
}

func save(src string) (tmpfile string) {
	tmpfile = "/tmp/test.go"
	fh, err := os.OpenFile(tmpfile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		panic(err)
	}
	fh.WriteString(src)
	fh.Close()
	return tmpfile
}

func buildMain(global string, nonGlobal string, pkgsToImport map[string]bool) string {
	imports := ""
	for k, _ := range pkgsToImport {
		imports += `import "` + k + "\"\n"
	}
	template := `
package main
%s
%s
func main() {
     %s
}
`
	return fmt.Sprintf(template, imports, global, nonGlobal)
}

// if line ends with '{' or '(', then consume until the corresponding '}' or ')'. Else return the next line.
func nextChunk(code string) (chunk string) {
	// get earliest of '{', '(' or '\n'
	var ch, closech rune
	var i int
	for i, ch = range code {
		if ch == '{' || ch == '(' || ch == '\n' {
			break
		}
	}
	pos := i + 1 // next scan always starts at pos
	if ch == '\n' {
		return code[:pos]
	} else if ch == '{' {
		closech = '}'
	} else if ch == '(' {
		closech = ')'
	} else {
		return code[:]
	}
	// Search for closing ch
	startch := ch
	count := 1
	for i, ch = range code[pos:] {
		if ch == startch {
			count++
		} else if ch == closech {
			count--
			if count == 0 {
				break
			}
		}
	}
	pos += i + 1
	if count != 0 {
		panic(fmt.Sprintf("Mismatched parentheses or brackets: %d; Extracted: <<<\n%s>>>\n", count, code[:pos]))
	}
	// consume trailing whitespace
	for i, ch = range code[pos:] {
		if ch == ' ' && ch == '\t' {
			pos++
		} else {
			if ch == '\n' {
				pos++
			}
			break
		}

	}
	return code[:pos]
}
