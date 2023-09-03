package project

import (
	"fmt"
	"github.com/tidwall/gjson"
	"gmtc/utils"
	"gmtc/parser"
	"os"
	"path"
	"strings"
)

type Project struct {
	Root      string
	File      string
	Resources []Resource
	Errors    utils.Errors
}

func LoadProject(file_path string) (Project, error) {
	root_dir := path.Dir(file_path)

	b, err := os.ReadFile(file_path)
	if err != nil {
		return Project{}, err
	}

	proj_json := gjson.ParseBytes(b)

	resource_paths := proj_json.Get("resources.#.id.path").Array()
	resources := make([]Resource, 0, len(resource_paths))
	proj_errors := make(utils.Errors, 0)

	for _, res_path := range resource_paths {
		res_fullpath := path.Join(root_dir, res_path.Str)
		resource, load_err := loadResource(res_fullpath)
		if load_err != nil {
			proj_errors = proj_errors.Add(load_err)
			continue
		}
		resources = append(resources, resource)
	}

	return Project{
		Root:      root_dir,
		File:      file_path,
		Resources: resources,
		Errors:    proj_errors,
	}, nil
}

func (p *Project) ErrorCount() int {
	count := len(p.Errors)

	for _, res := range p.Resources {
		count += len(res.GetErrors())
	}

	return count
}

func (p *Project) AllErrors() utils.Errors {
	out := make(utils.Errors, len(p.Errors))
	copy(out, p.Errors)
	for _, res := range p.Resources {
		out = append(out, res.GetErrors()...)
	}
	return out
}

type Resource interface {
	GetPath() string
	GetErrors() utils.Errors
}

type LoadResourceError struct { error }
type UnknownResourceError struct { error }

func loadResource(file_path string) (Resource, error) {
	b, err := os.ReadFile(file_path)
	if err != nil {
		return nil, LoadResourceError{err}
	}

	res_json := gjson.ParseBytes(b)
	rtype := res_json.Map()["resourceType"].Str

	switch rtype {
	case "GMScript":
		scr, err := loadGMScript(file_path)
		if err != nil {
			return nil, err
		}
		return &scr, nil

	case "GMObject":
		obj, err := loadGMObject(file_path, res_json)
		if err != nil {
			return nil, err
		}
		return &obj, nil

	case "GMShader", "GMRoom", "GMSprite":
		return &BaseResource{file_path}, nil
	}

	return nil, UnknownResourceError{fmt.Errorf("Unknown resource type %v", rtype)}
}

type BaseResource struct {
	Path string
}

func (r *BaseResource) GetPath() string         { return r.Path }
func (r *BaseResource) GetErrors() utils.Errors { return nil }

type ResGMScript struct {
	BaseResource
	GMLPath string
	Script  string
	Tokens  parser.Tokens
	Errors  utils.Errors
}

func (r *ResGMScript) GetErrors() utils.Errors { return r.Errors }

func (r *ResGMScript) Pretokenize() map[string]parser.Macro {
	ts, err := parser.Pretokenize(r.Script)
	if err != nil {
		r.Tokens = nil
		r.Errors = r.Errors.AddPrefix(r.GMLPath, err)
		return nil, err
	}
	r.Tokens = ts
	return ts.ExtractMacros()
}

func loadGMScript(path string) (ResGMScript, error) {
	gml_path := strings.TrimSuffix(path, ".yy") + ".gml"

	b, err := os.ReadFile(gml_path)
	if err != nil {
		return ResGMScript{}, LoadResourceError{err}
	}

	return ResGMScript{
		BaseResource: BaseResource{path},
		GMLPath:      gml_path,
		Script:       string(b),
	}, nil
}

type ResGMObject struct {
	BaseResource
	Dir    string
	Name   string
	Events []ResGMEvent
	Errors utils.Errors
}

func (r *ResGMObject) GetErrors() utils.Errors {
	out := make(utils.Errors, len(r.Errors))
	copy(out, r.Errors)
	for _, ev := range r.Events {
		out = append(out, ev.Errors...)
	}
	return out
}

func loadGMObject(file_path string, data gjson.Result) (ResGMObject, error) {
	dir := path.Dir(file_path)

	data_map := data.Map()
	name := data_map["name"].Str

	ev_array := data_map["eventList"].Array()
	events := make([]ResGMEvent, 0, len(ev_array))
	errors := make(utils.Errors, 0)

	for _, ev_json := range ev_array {
		ev, err := loadGMEvent(dir, ev_json)
		if err != nil {
			errors = errors.Add(err)
			continue
		}
		events = append(events, ev)
	}

	return ResGMObject{
		BaseResource: BaseResource{file_path},
		Dir:          dir,
		Name:         name,
		Events:       events,
		Errors:       errors,
	}, nil
}

type GM_EVENT int

//go:generate stringer -type=GM_EVENT -trimprefix=GME_
const (
	GME_Create GM_EVENT = iota
	GME_Destroy
	GME_Alarm
	GME_Step
	GME_Collision
	GME_Keyboard
	GME_Mouse
	GME_Other
	GME_Draw
	GME_KeyPress
	GME_KeyRelease
	GME_CleanUp
	GME_Gesture
	GME_COUNT
)

type ResGMEvent struct {
	ResGMScript
	Type int
	Num  int
}

func getEventScriptPath(dir string, evtype GM_EVENT, evnum int) string {
	if evtype >= GME_COUNT {
		panic(fmt.Sprintf("Unhandled event script type %v (%v)", evtype, evnum))
	}
	return path.Join(dir, evtype.String()+"_"+fmt.Sprint(evnum))
}

func loadGMEvent(dir string, data gjson.Result) (ResGMEvent, error) {
	data_map := data.Map()

	evnum := int(data_map["eventNum"].Num)
	evtype := int(data_map["eventType"].Num)

	path := getEventScriptPath(dir, GM_EVENT(evtype), evnum)

	script, err := loadGMScript(path)
	if err != nil {
		return ResGMEvent{}, err
	}

	return ResGMEvent{
		ResGMScript: script,
		Type:        evtype,
		Num:         evnum,
	}, nil
}
