package project

import (
	"os"
	"fmt"
	"path"
	"strings"
	"github.com/tidwall/gjson"
)

func ReadYYP(file_path string) error {
	root_dir := path.Dir(file_path)

	b, err := os.ReadFile(file_path)
	if err != nil { return err }
	res := gjson.ParseBytes(b)

	resources := res.Map()["resources"].Array()
	for _, res := range resources {
		p := res.Get("id.path").Str
		ReadYY(path.Join(root_dir, p))
	}

	return nil
}

func ReadYY(file_path string) error {
	b, err := os.ReadFile(file_path)
	if err != nil { return err }
	res := gjson.ParseBytes(b)

	rtype := res.Map()["resourceType"].Str
	switch rtype {
		case "GMScript":
			b, err := os.ReadFile(strings.TrimSuffix(file_path, ".yy") + ".gml")
			if err != nil { return err }
			text := string(b)
			fmt.Println(text)

		default: fmt.Printf("ReadYY: Unhandled resourceType: %v\n", rtype)
	}

	return nil
}
