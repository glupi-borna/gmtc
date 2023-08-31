package main

import (
	"flag"
	"fmt"
	"gmtc/parser"
	"io"
	"os"
)

var filepath = flag.String("path", "", "Path to the file to be parsed")
var stdin = flag.Bool("stdin", false, "Read from stdin")

func ReadStdin() (string, error) {
	text, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", err
	}
	return string(text), nil
}

func main() {
	flag.Parse()

	var err error
	text := ""

	if len(*filepath) > 0 && *stdin {
		panic("Cannot read file and stdin at the same time.")
	}

	if len(*filepath) == 0 && !*stdin {
		flag.Usage()
		return
	}

	if len(*filepath) > 0 {
		bytes, err := os.ReadFile(*filepath)
		if err != nil {
			panic(err)
		}
		text = string(bytes)
	}

	if *stdin {
		text, err = ReadStdin()
		if err != nil {
			panic(err)
		}
	}

	tokens, err := parser.Tokenize(&parser.Scanner{Text: text})

	fmt.Println(tokens)
	fmt.Println(err)
	fmt.Println()
	for _, tok := range tokens {
		fmt.Print(tok.Value, " ")
	}
}
