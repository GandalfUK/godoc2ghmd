// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Godoc example code extraction and comment -> Markdown formatting.
// Modified from the original example_textFunc in the godoc package.

package main

import (
	"bytes"
	"go/printer"
	"go/token"
	"io"
	"log"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/tools/godoc"
)

func startsWithUppercase(s string) bool {
	r, _ := utf8.DecodeRuneInString(s)
	return unicode.IsUpper(r)
}

// stripExampleSuffix strips lowercase braz in Foo_braz or Foo_Bar_braz from name
// while keeping uppercase Braz in Foo_Braz.
func stripExampleSuffix(name string) string {
	if i := strings.LastIndex(name, "_"); i != -1 {
		if i < len(name)-1 && !startsWithUppercase(name[i+1:]) {
			name = name[:i]
		}
	}
	return name
}

type myPres godoc.Presentation

// Write an AST node to w.
func (p *myPres) writeNode(w io.Writer, fset *token.FileSet, x interface{}) {
	// convert trailing tabs into spaces using a tconv filter
	// to ensure a good outcome in most browsers (there may still
	// be tabs in comments and strings, but converting those into
	// the right number of spaces is much harder)
	//
	// TODO(gri) rethink printer flags - perhaps tconv can be eliminated
	//           with an another printer mode (which is more efficiently
	//           implemented in the printer than here with another layer)
	mode := printer.TabIndent | printer.UseSpaces
	err := (&printer.Config{Mode: mode, Tabwidth: p.TabWidth}).Fprint(&tconv{p: (*godoc.Presentation)(p), output: w}, fset, x)
	if err != nil {
		log.Print(err)
	}
}

func (p *myPres) exampleMDFunc(info *godoc.PageInfo, funcName, indent string) string {
	if !p.ShowExamples {
		return ""
	}

	var buf bytes.Buffer
	first := true
	for _, eg := range info.Examples {
		name := stripExampleSuffix(eg.Name)
		if name != funcName {
			continue
		}

		if !first {
			buf.WriteString("\n")
		}
		first = false

		// print code
		cnode := &printer.CommentedNode{Node: eg.Code, Comments: eg.Comments}
		var buf1 bytes.Buffer
		p.writeNode(&buf1, info.FSet, cnode)
		code := buf1.String()
		// Additional formatting if this is a function body.
		if n := len(code); n >= 2 && code[0] == '{' && code[n-1] == '}' {
			// remove surrounding braces
			code = code[1 : n-1]
			// unindent
			code = strings.Replace(code, "\n    ", "\n", -1)
		}
		code = strings.Trim(code, "\n")

		buf.WriteString(indent)
		buf.WriteString("Example:\n\n")
		buf.WriteString("<details>\n")
		buf.WriteString("<summary>Click to expand code.</summary>\n\n")
		buf.WriteString("```go\n")
		buf.WriteString(code)
		buf.WriteString("\n```\n\n")
		buf.WriteString("</details>\n")
	}
	return buf.String()
}
