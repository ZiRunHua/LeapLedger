package accountService

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"strings"
)

var GroupApp = getNewGroup()

type Group struct {
	templates map[string]accountTemplate
	Base      base
}

type accountTemplate struct {
	Name     string
	Category string
}

func getNewGroup() *Group {
	group := &Group{}
	group.initTemplate()
	return group
}

const templatePath = "/data/account/template"

func (g *Group) initTemplate() {
	fmt.Printf("initTemplate path：%s\n", templatePath)
	//遍历文件夹
	err := filepath.WalkDir(
		templatePath, func(path string, fileInfo os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			//忽然文件夹和非yml格式文件
			if fileInfo.IsDir() || (filepath.Ext(path) != ".yaml" || filepath.Ext(path) != ".yml") {
				return nil
			}
			fileName := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
			content, err := os.ReadFile(path)
			if err != nil {
				fmt.Printf("读取文件 %s 时发生错误：%v\n", path, err)
				return nil // 不影响编译
			}

			g.templates[fileName] = accountTemplate{}
			if err = yaml.Unmarshal(content, g.templates[fileName]); err != nil {
				delete(g.templates, fileName)
				fmt.Printf("映射yaml文件 %s 时发生错误：%v\n", path, err)
				return nil
			}

			return nil
		},
	)
	if err != nil {
		fmt.Printf("遍历文件夹时发生错误：%v\n", err)
	}
}
