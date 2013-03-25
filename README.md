# `gore`

`gore` is a command-line evaluator of golang code -- a REPL without a loop, if you will -- meant for interactively trying out snippets of go code. It is an offline replacement to the 'go playground', and in addition, automatically supplies boiler-plate code such as `import` and `package` declarations and a `main` function wrapper. 

#Usage

(note: In the examples below, $ is the shell prompt, and the output of the snippet follows "----------------"
#### Code snippet in command line: gore evaluates its first argument
```
$ gore 'println(200*300, math.Log10(1000))'
---------------------------------
60000 +3.000000e+000
```

Note the absence of boiler-plate code like `package main`, `import "math"` and `func main() {}`


#### p alias for convenient printing
The example above can be written more compactly:
```
$ gore 'p 200*300, math.Log10(100)'
---------------------------------
60000
2
```
p pretty-prints each argument by formatting it with `fmt.Printf("%v\n")`

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

# The `gore/eval` package

`gore` is a thin command-line wrapper over the `gore/eval` package. Use this for your own REPL.

### How it works

The `eval.Eval` function expands aliases (currently only the `p` command), and scans the snippet for references to packages from the standard Go library. All such references a corresponding `import` statement. The source is then partitioned into global and non-global code, where global refers to `type`, `import` and `func` declarations. The rest is bundled into a `func main() {}` wrapper. This reorganized code is compiled using `go run` and the output (stdout and stderr) collected. If there are compiler errors pointing to incorrectly inferred packages, the corresponding import statements are removed and the code is run once again.