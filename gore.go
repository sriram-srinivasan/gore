package main

import (
	"bufio"
	"fmt"
	"github.com/sriram-srinivasan/gore/eval"
	"io"
	"os"
)

func main() {
	var src string
	if len(os.Args) > 1 {
		src = os.Args[1]
	} else {
		fmt.Println("Enter one or more lines and hit ctrl-D")
		src = readStdin()
	}

	out, err := eval.Eval(src)
	if err == "" {
		println("---------------------------------")
		println(out)
	} else {
		fmt.Println("== Error ========")
		fmt.Println(err)
	}
}

func readStdin() (buf string) {
	r := bufio.NewReader(os.Stdin)
	for {
		if line, err := r.ReadString('\n'); err != nil {
			if err == io.EOF {
				buf += line
			}
			break
		} else {
			buf += line
		}
	}
	return buf
}
