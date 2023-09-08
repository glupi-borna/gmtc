package main

import (
	"flag"
	"fmt"
	"gmtc/ast"
	"gmtc/parser"
	"gmtc/project"
	pp "gmtc/project_parser"
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

func ParseProject(filepath string) error {
	p, err := project.LoadProject(filepath)
	if err != nil {
		return err
	}

	errs := pp.Tokenize(p)
	fmt.Println(errs)
	// 	count := 0
	//
	// 	parse_resource := func(res project.Resource) {
	// 		script, ok := res.(*project.ResGMScript)
	// 		if ok {
	// 			count++
	// 			parse_error := Parse(script.Script)
	// 			if parse_error != nil {
	// 				script.Errors = script.Errors.AddPrefix(script.GMLPath, parse_error)
	// 			}
	// 			return
	// 		}
	//
	// 		obj, ok := res.(*project.ResGMObject)
	// 		if ok {
	// 			for _, ev := range obj.Events {
	// 				count++
	// 				parse_error := Parse(ev.Script)
	// 				if parse_error != nil {
	// 					ev.Errors = ev.Errors.AddPrefix(ev.GMLPath, parse_error)
	// 				}
	// 			}
	// 			return
	// 		}
	// 	}
	//
	// 	for _, res := range p.Resources {
	// 		parse_resource(res)
	// 	}
	//
	// 	all_errors := p.AllErrors()
	// 	err_count := len(all_errors)
	// 	fmt.Printf("Parsing project finished\n")
	// 	fmt.Printf("Parsed %v files\n", count)
	// 	fmt.Printf("%v errors\n", err_count)
	// 	if err_count > 0 {
	// 		load_errors := utils.ErrorCount[project.LoadResourceError](all_errors)
	// 		fmt.Printf("Failed to find %v files\n", load_errors)
	// 		for _, e := range all_errors {
	// 			if _, ok := e.(project.LoadResourceError) ; ok { continue }
	// 			fmt.Println(e)
	// 		}
	// 	}

	return nil
}

func ParseFile(filepath string) error {
	if strings.HasSuffix(filepath, ".yyp") {
		return ParseProject(filepath)
	}

	bytes, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}

	text := string(bytes)
	err = Parse(text)
	if err != nil {
		return fmt.Errorf("%v: %v", filepath, err)
	}

	return nil
}

func ParseStdin() error {
	text, err := ReadStdin()
	if err != nil {
		return err
	}
	err = Parse(text)
	if err != nil {
		return fmt.Errorf("Stdin: %v", err)
	}
	return nil
}

func Parse(text string) error {
	ts, err := parser.TokenizeString(text)
	if err != nil {
		return err
	}

	sn, err := ast.ParseAST(ts)
	if err != nil {
		return err
	}

	nb := ast.NodeBuilder{}
	sn.Render(&nb)
	fmt.Println(nb.String())

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
