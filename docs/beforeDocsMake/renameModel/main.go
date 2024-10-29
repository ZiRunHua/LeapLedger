package main

import (
	"bufio"
	"github.com/ZiRunHua/LeapLedger/global/constant"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	handleDir(constant.WORK_PATH + "/api/request/")
	handleDir(constant.WORK_PATH + "/api/response/")
}
func handleDir(path string) {
	_ = filepath.Walk(
		path, func(path string, info fs.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}
			handleFile(path)
			return nil
		},
	)
	return
}
func handleFile(path string) {
	file, err := os.OpenFile(path, os.O_RDWR, 0666)
	defer file.Close()
	if err != nil {
		panic(err)
	}
	reader := bufio.NewReader(file)
	var lines []string
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			panic(err)
		}
		lines = append(lines, line)
	}
	fileSet := token.NewFileSet()
	astFile, err := parser.ParseFile(fileSet, path, nil, parser.AllErrors)
	if err != nil {
		panic(err)
	}

	for _, decl := range astFile.Decls {
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
			lineNumber := fileSet.Position(structType.End()).Line - 1
			if !strings.Contains(lines[lineNumber], "// @name") {
				lines[lineNumber] = strings.TrimSpace(lines[lineNumber]) + " // @name " + structSpec.Name.Name + "\n"
			}
		}
	}
	file, err = os.OpenFile(path, os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	for _, line := range lines {
		_, err = writer.WriteString(line)
		if err != nil {
			panic(err)
		}
	}
	err = writer.Flush()
	if err != nil {
		panic(err)
	}
}
