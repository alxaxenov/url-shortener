package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"log"
	"strings"

	"golang.org/x/tools/go/packages"
)

func scanPackages() []PackageData {
	cfg := &packages.Config{
		Mode: packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedFiles | packages.NeedName,
	}
	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		log.Fatalf("loading packages: %v", err)
	}

	result := []PackageData{}
	for _, pkg := range pkgs {
		if len(pkg.Errors) > 0 {
			log.Printf("package %q parsing error\n", pkg.Dir)
			continue
		}
		if len(pkg.GoFiles) == 0 {
			continue
		}
		pckData := PackageData{Name: pkg.Name, Path: pkg.Dir}

		for _, file := range pkg.Syntax {
			for _, decl := range file.Decls {
				genDecl, ok := decl.(*ast.GenDecl)
				if !ok || genDecl.Tok != token.TYPE {
					continue
				}
				for _, spec := range genDecl.Specs {
					typeSpec, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}
					if !hasComment(genDecl, typeSpec) {
						continue
					}
					typeStruct, ok := typeSpec.Type.(*ast.StructType)
					if !ok {
						continue
					}
					resetLines := getReset(typeStruct, pkg)
					if len(resetLines) > 0 {
						pckData.Structs = append(
							pckData.Structs,
							StructData{Name: typeSpec.Name.Name, Body: resetLines},
						)
					}
				}
			}
		}

		if len(pckData.Structs) > 0 {
			result = append(result, pckData)
		}
	}
	return result
}

func hasComment(decl *ast.GenDecl, spec *ast.TypeSpec) bool {
	if decl.Doc != nil && strings.Contains(decl.Doc.Text(), "generate:reset") {
		return true
	}
	if spec.Doc != nil && strings.Contains(spec.Doc.Text(), "generate:reset") {
		return true
	}
	return false
}

func getReset(st *ast.StructType, pkg *packages.Package) string {
	var result []string
	for _, field := range st.Fields.List {
		for _, name := range field.Names {
			t := pkg.TypesInfo.TypeOf(field.Type)
			reset := generateResetForField(name.Name, t)
			result = append(result, reset...)
		}
	}
	return strings.Join(result, "\n")
}

func generateResetForField(fieldName string, t types.Type) []string {
	// Если указатель, работаем с ним
	if ptr, ok := t.(*types.Pointer); ok {
		inner := ptr.Elem()
		switch u := inner.Underlying().(type) {
		case *types.Basic:
			switch u.Kind() {
			case types.Int, types.Int8, types.Int16, types.Int32, types.Int64,
				types.Uint, types.Uint8, types.Uint16, types.Uint32, types.Uint64,
				types.Float32, types.Float64:
				return []string{fmt.Sprintf("if rs.%s != nil { *rs.%s = 0 }", fieldName, fieldName)}
			case types.String:
				return []string{fmt.Sprintf("if rs.%s != nil { *rs.%s = \"\" }", fieldName, fieldName)}
			case types.Bool:
				return []string{fmt.Sprintf("if rs.%s != nil { *rs.%s = false }", fieldName, fieldName)}
			}
		case *types.Struct:
			return []string{
				fmt.Sprintf("if rs.%s != nil {", fieldName),
				fmt.Sprintf("    if resetter, ok := interface{}(rs.%s).(interface{ Reset() }); ok {", fieldName),
				"        resetter.Reset()",
				"    }",
				"}",
			}
		case *types.Slice:
			return []string{fmt.Sprintf("if rs.%s != nil { *rs.%s = (*rs.%s)[:0] }", fieldName, fieldName, fieldName)}
		case *types.Map:
			return []string{fmt.Sprintf("if rs.%s != nil { clear(*rs.%s) }", fieldName, fieldName)}
		default:
			return nil
		}
		return nil
	}

	// Не указатель: работаем с underlying
	u := t.Underlying()
	switch t := u.(type) {
	case *types.Basic:
		switch t.Kind() {
		case types.Int, types.Int8, types.Int16, types.Int32, types.Int64,
			types.Uint, types.Uint8, types.Uint16, types.Uint32, types.Uint64,
			types.Float32, types.Float64:
			return []string{fmt.Sprintf("rs.%s = 0", fieldName)}
		case types.String:
			return []string{fmt.Sprintf("rs.%s = \"\"", fieldName)}
		case types.Bool:
			return []string{fmt.Sprintf("rs.%s = false", fieldName)}
		}
	case *types.Slice:
		return []string{fmt.Sprintf("rs.%s = rs.%s[:0]", fieldName, fieldName)}
	case *types.Map:
		return []string{fmt.Sprintf("clear(rs.%s)", fieldName)}
	case *types.Struct:
		return []string{
			fmt.Sprintf("if resetter, ok := interface{}(&rs.%s).(interface{ Reset() }); ok {", fieldName),
			"    resetter.Reset()",
			"}",
		}
	}
	return nil
}
