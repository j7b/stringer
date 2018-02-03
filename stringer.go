//go:generate go run stringer.go -dir example

package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/build"
	"go/format"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

func init() {
	log.SetFlags(0)
	log.SetPrefix("stringer: ")
}

const Template = `package {{.PackageName}}

import "fmt"

{{range $type,$names := .NameMap}}func ({{firstletter $type}} {{$type}}) String() string {
	{{range $names}}if {{firstletter $type}} == {{.}} {return "{{.}}"}
	{{end}}
	return fmt.Sprintf("{{$type}}(%v)",{{firstletter $type}})
}

{{end}}
`

type Generator struct {
	PackageName string
	NameMap     map[string][]string
}

func firstletter(s string) string {
	return strings.ToLower(s)[:1]
}

func (g *Generator) Generate(w io.Writer) error {
	tmpl := template.Must(template.New("").Funcs(map[string]interface{}{"firstletter": firstletter}).Parse(Template))
	buf := new(bytes.Buffer)
	err := tmpl.Execute(buf, g)
	if err != nil {
		return err
	}
	src, err := format.Source(buf.Bytes())
	if err != nil {
		return err
	}
	_, err = w.Write(src)
	return err
}

func generator() (*Generator, error) {
	pkg, err := build.Default.ImportDir(Directory, 0)
	if err != nil {
		return nil, err
	}
	if pkg.Name == "main" {
		return nil, fmt.Errorf(`Refuse to process package "main"`)
	}
	var filenames []string
	for _, v := range [][]string{pkg.GoFiles, pkg.CgoFiles, pkg.SFiles} {
		filenames = append(filenames, v...)
	}
	dirprefix(Directory, filenames)
	var astFiles []*ast.File
	fs := token.NewFileSet()
	for _, name := range filenames {
		if !strings.HasSuffix(name, ".go") {
			continue
		}
		parsedFile, err := parser.ParseFile(fs, name, nil, parser.ParseComments)
		if err != nil {
			return nil, err
		}
		astFiles = append(astFiles, parsedFile)
	}
	config := types.Config{Importer: importer.Default(), FakeImportC: true}
	_, err = config.Check(Directory, fs, astFiles, nil)
	if err != nil {
		return nil, err
	}
	namemap := make(map[string][]string)
	for _, file := range astFiles {
		if file != nil {
			ast.Inspect(file, func(node ast.Node) bool {
				decl, ok := node.(*ast.GenDecl)
				if !ok || decl.Tok != token.CONST {
					return true
				}
				typ := ""
				for _, spec := range decl.Specs {
					vspec := spec.(*ast.ValueSpec)
					if vspec.Type == nil && len(vspec.Values) > 0 {
						for i, name := range vspec.Names {
							if name.Name == "_" {
								continue
							}
							v := vspec.Values[i]
							if ce, ok := v.(*ast.CallExpr); ok {
								fun, ok := ce.Fun.(*ast.Ident)
								if ok {
									typ = fun.Name
									if TypeSet.Check(typ) {
										name := vspec.Names[i]
										namemap[typ] = append(namemap[typ], name.Name)
									}
								}
							}
						}
						typ = ""
						continue
					}
					if vspec.Type != nil {
						ident, ok := vspec.Type.(*ast.Ident)
						if !ok {
							continue
						}
						typ = ident.Name
					}
					if TypeSet.Check(typ) {
						for _, name := range vspec.Names {
							if name.Name == "_" {
								continue
							}
							namemap[typ] = append(namemap[typ], name.Name)
						}
					}
				}
				return true
			})
		}
	}
	return &Generator{PackageName: pkg.Name, NameMap: namemap}, nil
}

func dirprefix(dir string, slice []string) {
	if dir == `.` {
		return
	}
	for i, v := range slice {
		slice[i] = filepath.Join(dir, v)
	}
}

func main() {
	g, err := generator()
	if err != nil {
		log.Fatal(err)
	}
	TypeSet.Unchecked(log.Printf)
	if Outfile == `-` {
		buf := new(bytes.Buffer)
		if err = g.Generate(buf); err != nil {
			log.Fatal(err)
		}
		io.Copy(os.Stdout, buf)
		return
	}
	outfile := filepath.Join(Directory, Outfile)
	f, err := os.Create(outfile)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if err = g.Generate(f); err != nil {
		log.Fatal(err)
	}
}

var directory = flag.String("dir", "", "Package `directory`, defaults to current.")
var outfile = flag.String("o", "", "Output `filename`, defaults to stringer_gen.go in package directory.  If a literal dash (meaning '-o -') writes to standard output.")
var typenames = flag.String("types", "", "Comma-delimited `list` of names of types to generate String methods for, defaults to all public types with named constants.")

var Directory, Outfile string
var TypeSet typeset

type typeset map[string]bool

func (t typeset) Check(name string) bool {
	if len(name) == 0 {
		return false
	}
	first := name[:1]
	if strings.ToUpper(first) != first {
		return false
	}
	if len(t) == 0 {
		return true
	}
	_, ok := t[name]
	if ok {
		t[name] = true
	}
	return ok
}

func (t typeset) Unchecked(f func(string, ...interface{})) {
	for k, v := range TypeSet {
		if !v {
			f("The type `%s` was specified but no constants were found", k)
		}
	}
}

func init() {
	if !flag.Parsed() {
		flag.Parse()
	}
	t := *typenames
	if len(t) > 0 {
		TypeSet = make(typeset)
		names := strings.Split(t, `,`)
		for i := range names {
			t = strings.TrimSpace(names[i])
			TypeSet[t] = false
		}
	}
	Directory, Outfile = *directory, *outfile
	if len(Outfile) == 0 {
		Outfile = "stringer_gen.go"
	}
	if len(Directory) == 0 {
		Directory = "."
	}
}
