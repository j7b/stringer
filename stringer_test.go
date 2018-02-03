package main

import (
	"bytes"
	"testing"
)

func TestGenerate(t *testing.T) {
	g := Generator{
		PackageName: "example",
		NameMap:     map[string][]string{"Int": []string{"One", "Two"}},
	}
	buf := new(bytes.Buffer)
	if err := g.Generate(buf); err != nil {
		t.Fatal(err)
	}
	t.Logf("\n%s\n", buf.String())
}
