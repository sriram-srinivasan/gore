package main

import "gore/eval"
import "fmt"
import "bufio"
import "os"
import "io"

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
		fmt.Println(out)
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
