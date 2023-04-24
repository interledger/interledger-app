package main

import (
	"bufio"
	_ "embed"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path"
	"reflect"
	"strings"
	"text/template"
)

//go:embed bckgen.tmpl
var tmpl string

func main() {
	const (
		pkgPrefix  = "package "
		pkgUnknown = "unknown"
	)

	// Determine the package name from the input file's package declaration
	inputDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	inputDir = "/Users/barnarddutoit/workspace/fynbos/go/tools/backendsgen/"
	inputFile, err := os.Open(path.Join(inputDir, "backends.go"))
	if err != nil {
		panic(err)
	}
	defer inputFile.Close()

	outputFile, err := os.Create(path.Join(inputDir, "backends_gen.go"))

	s := bufio.NewScanner(inputFile)
	var packageName string
	for s.Scan() {
		line := s.Text()
		if strings.HasPrefix(line, pkgPrefix) {
			packageName = line[len(pkgPrefix):]
			break
		}
	}
	if err := s.Err(); err != nil {
		panic(err)
	}
	if packageName == "" {
		packageName = pkgUnknown
	}

	// Parse the input file and extract the method signatures
	var fields []struct {
		Name string
		Type string
	}
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, inputFile.Name(), nil, 0)
	if err != nil {
		panic(err)
	}
	for _, decl := range node.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.TYPE {
			for _, spec := range genDecl.Specs {
				if typeSpec, ok := spec.(*ast.TypeSpec); ok && typeSpec.Name.Name == "Backends" {
					if interfaceType, ok := typeSpec.Type.(*ast.InterfaceType); ok {
						for _, method := range interfaceType.Methods.List {
							fmt.Println("TTTTTTT", method, reflect.TypeOf(method), reflect.TypeOf(method.Type))
							if ident, ok := method.Type.(*ast.Ident); ok {
								fmt.Println("XXXXX", ident.Name)
								fields = append(fields, struct {
									Name string
									Type string
								}{method.Names[0].Name, ident.Name})
							} else {
								fmt.Println("NOK")
							}
						}
					}
				}
			}
		} else {
			fmt.Println("FUCk")
		}
	}

	// Generate the list of optional functions and their parameters
	var funcs []struct {
		WithFuncName   string
		WithFuncParam  string
		WithFuncAssign string
	}
	for _, field := range fields {
		name := "With" + field.Name
		param := strings.ToLower(string(field.Name[0])) + field.Name[1:] + " " + field.Type
		assign := "b." + field.Name + " = " + strings.ToLower(string(field.Name[0])) + field.Name[1:]
		funcs = append(funcs, struct {
			WithFuncName   string
			WithFuncParam  string
			WithFuncAssign string
		}{name, param, assign})
	}

	// Generate the template and execute it
	tmpl, err := template.New("backendImpl").Parse(tmpl)
	if err != nil {
		panic(err)
	}

	err = tmpl.Execute(outputFile, struct {
		Package string
		Fields  []struct {
			Name string
			Type string
		}
		Funcs []struct {
			WithFuncName   string
			WithFuncParam  string
			WithFuncAssign string
		}
	}{packageName, fields, funcs})
	if err != nil {
		panic(err)
	}
}
