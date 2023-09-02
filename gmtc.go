package main

import (
	"flag"
	"fmt"
	"gmtc/parser"
	"gmtc/project"
	"io"
	"os"
	"strings"
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

func ParseFile(filepath string) error {
	if strings.HasSuffix(filepath, ".yyp") {
		return project.ReadYYP(filepath)
	}

	bytes, err := os.ReadFile(filepath)
	if err != nil { return err }
	text := string(bytes)
	err = Parse(text)
	if err != nil {
		return fmt.Errorf("%v: %v", filepath, err)
	}
	return nil
}

func ParseStdin() error {
	text, err := ReadStdin()
	if err != nil { return err }
	err = Parse(text)
	if err != nil {
		return fmt.Errorf("Stdin: %v", err)
	}
	return nil
}

func Parse(text string) error {
	tokens, err := parser.Tokenize(text)
	if err != nil { return err }
	fmt.Println(tokens)
	return nil
}

func main() {
	flag.Parse()

	var err error

	if len(*filepath) > 0 && *stdin {
		panic("Cannot read file and stdin at the same time.")
	}

	if len(*filepath) == 0 && !*stdin {
		flag.Usage()
		return
	}

	if len(*filepath) > 0 {
		err = ParseFile(*filepath)
		if err != nil { panic(err) }
		return
	}

	if *stdin {
		err = ParseStdin()
		if err != nil { panic(err) }
	}
}
