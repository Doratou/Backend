package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type constValue struct {
	name  string // e.g. UserRoleUser
	value string // literal source text, e.g. "\"USER\""
}

type enumType struct {
	name     string // e.g. UserRole
	doc      []string
	values   []constValue
	hasValid bool
	// extraMethods holds the source text of any method on this type other
	// than IsValid (e.g. String(), MarshalJSON()), preserved verbatim.
	extraMethods []string
	firstSeen    int // order of first appearance in the file
}

func main() {
	dry := flag.Bool("dry", false, "print the rewritten files instead of writing them")
	flag.Parse()

	pattern := "./internal/domain/enum/*.go"
	if flag.NArg() > 0 {
		pattern = flag.Arg(0)
	}

	matches, err := filepath.Glob(pattern)
	if err != nil {
		fmt.Fprintf(os.Stderr, "bad pattern %q: %v\n", pattern, err)
		os.Exit(1)
	}
	if len(matches) == 0 {
		fmt.Fprintf(os.Stderr, "no files matched %q\n", pattern)
		os.Exit(1)
	}
	sort.Strings(matches)

	for _, path := range matches {
		if err := processFile(path, *dry); err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
			os.Exit(1)
		}
	}
}

func processFile(path string, dry bool) error {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parse: %w", err)
	}

	enums := map[string]*enumType{}
	var order []string        // enum type names, in first-seen order
	var otherDecls []ast.Decl // non-enum, non-IsValid-method decls, kept as-is

	getEnum := func(name string) *enumType {
		e, ok := enums[name]
		if !ok {
			e = &enumType{name: name, firstSeen: len(order)}
			enums[name] = e
			order = append(order, name)
		}
		return e
	}

	// Pass 1: type declarations (`type X string`) define which identifiers
	// are enum types, plus their doc comments.
	for _, decl := range file.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok || gd.Tok != token.TYPE {
			continue
		}
		isEnumBlock := false
		for _, spec := range gd.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			ident, ok := ts.Type.(*ast.Ident)
			if !ok || ident.Name != "string" {
				continue
			}
			isEnumBlock = true
			e := getEnum(ts.Name.Name)
			doc := ts.Doc
			if doc == nil {
				doc = gd.Doc
			}
			if doc != nil {
				for _, c := range doc.List {
					e.doc = append(e.doc, c.Text)
				}
			}
		}
		if !isEnumBlock {
			otherDecls = append(otherDecls, decl)
		}
	}

	// Pass 2: const blocks. Each ValueSpec that's typed as one of our enum
	// names (explicitly, or inherited from the previous spec in the same
	// block) contributes its values in source order.
	for _, decl := range file.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok || gd.Tok != token.CONST {
			continue
		}
		matchedAny := false
		var lastType string
		for _, spec := range gd.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			typeName := lastType
			if vs.Type != nil {
				if ident, ok := vs.Type.(*ast.Ident); ok {
					typeName = ident.Name
				}
			}
			e, known := enums[typeName]
			if !known {
				continue
			}
			lastType = typeName
			matchedAny = true
			for i, nameIdent := range vs.Names {
				lit := ""
				if i < len(vs.Values) {
					lit = exprString(fset, vs.Values[i])
				}
				e.values = append(e.values, constValue{name: nameIdent.Name, value: lit})
			}
		}
		if !matchedAny {
			otherDecls = append(otherDecls, decl)
		}
	}

	// Pass 3: methods. An IsValid method on an enum type is dropped (we
	// regenerate it); any other method on an enum type is preserved
	// verbatim and re-emitted after IsValid. Everything else is kept as-is.
	for _, decl := range file.Decls {
		fd, ok := decl.(*ast.FuncDecl)
		if !ok || fd.Recv == nil || len(fd.Recv.List) != 1 {
			if ok {
				otherDecls = append(otherDecls, decl)
			}
			continue
		}
		recvType := recvTypeName(fd.Recv.List[0].Type)
		e, known := enums[recvType]
		if !known {
			otherDecls = append(otherDecls, decl)
			continue
		}
		if fd.Name.Name == "IsValid" {
			e.hasValid = true
			continue // regenerated below; drop the original source
		}
		e.extraMethods = append(e.extraMethods, declString(fset, fd))
	}

	// fset is also needed for exprString/declString above; keep var used.
	_ = fset

	if len(order) == 0 {
		return fmt.Errorf("no string-backed enum types found")
	}

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "package %s\n\n", file.Name.Name)

	for i, name := range order {
		e := enums[name]
		if i > 0 {
			buf.WriteString("\n")
		}
		writeEnum(&buf, e)
	}

	for _, decl := range otherDecls {
		buf.WriteString("\n")
		buf.WriteString(declString(fset, decl))
		buf.WriteString("\n")
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("gofmt generated source: %w\n--- generated ---\n%s", err, buf.String())
	}

	if dry {
		fmt.Printf("// ===== %s =====\n%s\n", path, formatted)
		return nil
	}

	info, err := os.Stat(path)
	mode := os.FileMode(0644)
	if err == nil {
		mode = info.Mode()
	}
	if err := os.WriteFile(path, formatted, mode); err != nil {
		return fmt.Errorf("write: %w", err)
	}
	fmt.Printf("rewrote %s (%d enum type%s)\n", path, len(order), plural(len(order)))
	return nil
}

func writeEnum(buf *bytes.Buffer, e *enumType) {
	// 1. comment
	if len(e.doc) > 0 {
		for _, line := range e.doc {
			buf.WriteString(line)
			buf.WriteString("\n")
		}
	} else {
		names := make([]string, len(e.values))
		for i, v := range e.values {
			names[i] = v.value
		}
		fmt.Fprintf(buf, "// %s enum values: %s.\n", e.name, strings.Join(names, ", "))
	}

	// 2. enum (type) definition
	fmt.Fprintf(buf, "type %s string\n\n", e.name)

	// 3. values
	buf.WriteString("const (\n")
	for _, v := range e.values {
		fmt.Fprintf(buf, "\t%s %s = %s\n", v.name, e.name, v.value)
	}
	buf.WriteString(")\n")

	// 4. IsValid
	recv := receiverFor(e.name)
	buf.WriteString("\n")
	fmt.Fprintf(buf, "func (%s %s) IsValid() bool {\n", recv, e.name)
	fmt.Fprintf(buf, "\tswitch %s {\n", recv)
	buf.WriteString("\tcase ")
	for i, v := range e.values {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(v.name)
	}
	buf.WriteString(":\n")
	buf.WriteString("\t\treturn true\n")
	buf.WriteString("\tdefault:\n")
	buf.WriteString("\t\treturn false\n")
	buf.WriteString("\t}\n")
	buf.WriteString("}\n")

	// any other preserved methods
	for _, m := range e.extraMethods {
		buf.WriteString("\n")
		buf.WriteString(m)
		buf.WriteString("\n")
	}
}

// receiverFor picks a short receiver name: the lowercase first letter of
// the type name, falling back to "x" if that happens to be empty.
func receiverFor(typeName string) string {
	if typeName == "" {
		return "x"
	}
	r := strings.ToLower(typeName[:1])
	if r == "" {
		return "x"
	}
	return r
}

func recvTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return recvTypeName(t.X)
	default:
		return ""
	}
}

func exprString(fset *token.FileSet, e ast.Expr) string {
	var buf bytes.Buffer
	if err := format.Node(&buf, fset, e); err != nil {
		return fmt.Sprintf("%v", e)
	}
	return buf.String()
}

func declString(fset *token.FileSet, d ast.Decl) string {
	var buf bytes.Buffer
	if err := format.Node(&buf, fset, d); err != nil {
		return fmt.Sprintf("%v", d)
	}
	return buf.String()
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
