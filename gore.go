package main

import (
	"bufio"
	"fmt"
	"github.com/shurcooL/gore/eval"
	"io"
	"os"
)

func main() {
	var src string
	if len(os.Args) > 1 {
		src = os.Args[1]
	} else {
		fmt.Println("Enter one or more lines and press Ctrl-D.")
		src = readStdin()
	}

	out, err := eval.Eval(src)
	if err == "" {
		//fmt.Fprintln(os.Stderr, "------------------------------")
		fmt.Print(out)
	} else {
		fmt.Println("===== Error =====")
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
