package main

import (
	"KeepAccount/global/constant"
	"bufio"
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	err := handleDir(constant.WORK_PATH + "/api/request/")
	if err != nil {
		panic(err)
	}
	// err = handleDir(constant.WORK_PATH + "/api/response/")
	// if err != nil {
	// 	panic(err)
	// }
}
func handleDir(path string) error {
	return filepath.Walk(
		path, func(path string, info fs.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}

			return handleFileV1(path)
		},
	)
}
func handleFileV1(path string) error {
	var output bytes.Buffer
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.AllErrors)
	if err != nil {
		return err
	}
	var lastOffset int
	// 打印语法树
	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		if genDecl.Tok != token.TYPE {
			continue
		}
		for _, spec := range genDecl.Specs {
			structSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			structType, ok := structSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}
			line := fset.Position(structType.End())
			output.Write(content[lastOffset:line.Offset])
			if(output.WriteTo())
			output.Write([]byte(" // @name " + structSpec.Name.Name))
			lastOffset = line.Offset

		}

	}
	output.WriteTo()
	err = os.WriteFile(constant.WORK_PATH+"/docs/test/"+filepath.Base(path), output.Bytes(), 0777)
	if err != nil {
		panic(err)
	}
	return nil
}

func handleFile(path string) error {
	log.Println(path)
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, content, parser.ParseComments)
	if err != nil {
		fmt.Println("Error parsing file:", err)
		return err
	}
	var output bytes.Buffer
	writer := bufio.NewWriter(&output)
	for _, decl := range file.Decls {
		fmt.Println(decl.Pos())
		// 检查是否是类型声明
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}
		// 遍历所有类型声明
		for _, spec := range genDecl.Specs {
			fmt.Println(genDecl.Specs)
			panic("")
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			// 检查是否是结构体类型
			_, ok = typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			// 检查是否已经存在 @name 注释
			name := typeSpec.Name.Name
			hasAnnotation := false
			if genDecl.Doc == nil {
				continue
			}
			for _, comment := range genDecl.Doc.List {
				if comment == nil {
					continue
				}
				fmt.Println(comment.Text)
				if strings.Contains(comment.Text, "// @name "+name) {

					hasAnnotation = true
					break
				} else {

					writer.WriteString(comment.Text)
				}
			}

			// 如果没有注释，则添加
			if !hasAnnotation {

				end := genDecl.Doc.List[len(genDecl.Doc.List)-1]
				if strings.Contains(end.Text, "//") {
					continue
				}
				writer.WriteString(end.Text + " // @name " + name)
				fmt.Println(end.Text + " // @name " + name)
				panic("")
				// // 在文件内容中查找结构体的最后一行，添加注释
				// pos := fset.Position(genDecl.End())
				// fmt.Println(pos)
				// lineNumber := pos.Line
				// fmt.Println(output.String())
				// lines := strings.Split(output.String(), "\n")
				// fmt.Println(lines)
				// if lineNumber <= len(lines) {
				// 	lines[lineNumber-1] += " // @name " + name
				// } else {
				// 	lines = append(lines, "// @name "+name)
				// }
				// fmt.Println(lines)
				// for _, line := range lines {
				// 	fmt.Fprintln(writer, line)
				// }
				// panic("")
			}
		}
	}

	output.Write(content)
	writer.Flush()

	return os.WriteFile(constant.WORK_PATH+"/docs/test/"+filepath.Base(path), output.Bytes(), 0644)
}
