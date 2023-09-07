package project_parser

import (
	"gmtc/ast"
	"gmtc/parser"
	"gmtc/project"
	"gmtc/utils"
)

type ProjectParser struct {
	Project project.Project
}

func Tokenize(p project.Project) error {
	pp := ProjectParser{p}

	macros := make([]map[string]parser.Macro, 0)
	errs := make(utils.Errors, 0)

	for _, res := range pp.Project.Resources {
		script, ok := res.(*project.ResGMScript)
		if ok {
			m, e := script.Pretokenize()
			if e != nil {
				errs = errs.AddPrefix(script.GMLPath, e)
				continue
			}
			macros = append(macros, m)
			continue
		}

		obj, ok := res.(*project.ResGMObject)
		if ok {
			for _, ev := range obj.Events {
				m, e := ev.Pretokenize()
				if e != nil {
					errs = errs.AddPrefix(ev.GMLPath, e)
					continue
				}
				macros = append(macros, m)
			}
			continue
		}
	}

	all_macros := utils.MapMerge(macros...)

	for _, res := range pp.Project.Resources {
		script, ok := res.(*project.ResGMScript)
		if ok {
			script.Tokens = script.Tokens.InsertMacros(all_macros).Clean(all_macros)
			continue
		}

		obj, ok := res.(*project.ResGMObject)
		if ok {
			for _, ev := range obj.Events {
				ev.Tokens = ev.Tokens.InsertMacros(all_macros).Clean(all_macros)
			}
			continue
		}
	}

	for _, res := range pp.Project.Resources {
		script, ok := res.(*project.ResGMScript)
		if ok {
			_, e := ast.ParseAST(script.Tokens)
			if e != nil {
				errs = errs.AddPrefix(script.GMLPath, e)
			}
			continue
		}

		obj, ok := res.(*project.ResGMObject)
		if ok {
			for _, ev := range obj.Events {
				_, e := ast.ParseAST(ev.Tokens)
				if e != nil {
					errs = errs.AddPrefix(ev.GMLPath, e)
				}
			}
			continue
		}
	}

	if len(errs) == 0 { return nil }
	return errs.Merge()
}
