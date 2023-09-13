package main

import (
	"flag"
	"io"
	"os"
	"strings"
	"gmtc/project"
)

var filepath = flag.String("path", "", "Path to the file to be parsed")
var stdin = flag.Bool("stdin", false, "Read from stdin")

func ParseProject(p project.Project) error {
	p.Parse()
	return p.AllErrors().Merge()
}

func LoadProject(filepath string) (project.Project, error) {
	p, err := project.LoadProject(filepath)
	if err != nil { return project.Project{}, err }
	return p, nil
}

func ParseFile(filepath string) error {
	if strings.HasSuffix(filepath, ".yyp") {
		p, err := LoadProject(filepath)
		if err != nil { return err }
		return ParseProject(p)
	}

	p := project.SingleFile(filepath)
	return ParseProject(p)
}

func ReadStdin() (string, error) {
	text, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", err
	}
	return string(text), nil
}

func ParseStdin() error {
	text, err := ReadStdin()
	if err != nil {
		return err
	}

	p := project.CodeProject("<stdin>", text)
	return ParseProject(p)
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
		if err != nil {
			panic(err)
		}
		return
	}

	if *stdin {
		err = ParseStdin()
		if err != nil {
			panic(err)
		}
	}
}
