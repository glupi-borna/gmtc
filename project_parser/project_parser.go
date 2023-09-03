package project_parser

import (
	"gmtc/utils"
	"gmtc/project"
	"gmtc/parser"
)

type ProjectParser struct {
	Project project.Project
}

func Tokenize(p project.Project) {
	pp := ProjectParser{p}

	macros := make([]map[string]parser.Macro, 0)

	for _, res := range pp.Project.Resources {
		script, ok := res.(*project.ResGMScript)
		if ok {
			macros = append(macros, script.Pretokenize())
			continue
		}

		obj, ok := res.(*project.ResGMObject)
		if ok {
			for _, ev := range obj.Events {
				macros = append(macros, ev.Pretokenize())
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
}
