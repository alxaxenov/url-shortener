package main

import (
	"bytes"
	"go/format"
	"log"
	"os"
	"path/filepath"
	"text/template"
)

const templateStr = `
// Code generated; DO NOT EDIT.

package {{.Name}}

{{range .Structs}}
func (rs *{{.Name}}) Reset() {
	if rs == nil {
        return
    }
	{{.Body}}
}

{{end}}
`

var tmpl = template.Must(template.New("enum").Parse(templateStr))

func main() {
	d := scanPackages()
	for _, pkgData := range d {
		var buf bytes.Buffer
		err := tmpl.Execute(&buf, pkgData)
		if err != nil {
			log.Printf("Skipping package %s. Error executing template for package %s: %v", pkgData.Name, pkgData.Path, err)
			continue
		}

		bufFmt, err := format.Source(buf.Bytes())
		if err != nil {
			log.Printf("Skipping package %s. Result formatting error for package %s: %v", pkgData.Name, pkgData.Path, err)
			continue
		}

		path := filepath.Join(pkgData.Path, "reset.gen.go")
		err = os.WriteFile(path, bufFmt, 0644)
		if err != nil {
			log.Printf("Skipping package %s. File writing error for package %s: %v", pkgData.Name, pkgData.Path, err)
			continue
		}

	}
}
