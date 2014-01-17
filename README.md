#gore

`gore` is a command-line evaluator for golang code -- a REPL without a loop, if you will. It is a replacement for the [go playground](http://play.golang.org), while making it much easier to interactively try out bits of code: `gore` automatically supplies boiler-plate code such as `import` and `package` declarations and a `main` function wrapper. Also, since it runs on your own computer, no code is rejected on security grounds (unlike go playground's safe sandbox mode).

#Usage

(note: In the examples below, $ is the shell prompt, and the output of the snippet follows "----------------"
#### Code snippet in command line: gore evaluates its first argument
```
$ gore 'println(200*300, math.Log10(1000))'
---------------------------------
60000 +3.000000e+000
```

Note the absence of boiler-plate code like `package main`, `import "math"` and `func main() {}`

#### Default to `stdin` without arguments

```
$ gore
Enter one or more lines and hit ctrl-D
func test() string {return "hello"}
println(test())
^D
---------------------------------
hello
```
#### Alias for convenient printing
The example above can be written more compactly:
```
$ gore 'p 200*300, math.Log10(100)'
---------------------------------
60000
2
```
`p arg1, arg2` pretty-prints each argument by formatting it with `fmt.Printf("%v\n")`
`t` arg1, arg2` prints the type of each argument
#### Command-line arg can be over multiple lines
```
$ gore '
 p "Making a point"
 type Point struct {
    x,y int
 }
 v := Point{10, 100}
 p v
' 
---------------------------------
Making a point
{10 100}
```
#### Import statements are inferred 
Standard go packages are automatically imported. Where there is a clash of names, the more "likely" one is preferred: `math/rand` to `crypto/rand`, `net/http/pprof` to `runtime/pprof` and `text/template` to `html/template`. Of course, you can add import statements of your own (which overrides the default preferences as well)
```
$ gore '
  r := regexp.MustCompile(`(\w+) says (\w+)`)
  match := r.FindStringSubmatch("World says Hello")
  p "0:" + match[0], "1:"+ match[1], "2:" + match[2]
  '
---------------------------------
0:World says Hello
1:World
2:Hello
```


# Install

```
go get github.com/sriram-srinivasan/gore
go test github.com/sriram-srinivasan/gore/eval
```

# The `gore/eval` package

`gore` is a thin command-line wrapper over the `gore/eval` package. Use this for your own REPL.

### How it works

The `eval.Eval` function expands aliases, and scans the snippet for references to packages from the standard Go library. All such references a corresponding `import` statement. The source is then partitioned into global and non-global code, where global refers to `type`, `import` and `func` declarations. The rest is bundled into a `func main() {}` wrapper. This reorganized code is compiled using `go run` and the output (stdout and stderr) collected. If there are compiler errors pointing to incorrectly inferred packages, the corresponding import statements are removed and the code is run once again.

To examine the generated code, set the environment variables TMPDIR or TEMPDIR, and look for $TMPDIR/gore_eval.go

# License

`gore` is available under a liberal MIT style license. See the _LICENSE_ file.