package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"text/template"

	_ "github.com/go-sql-driver/mysql"

	"github.com/matopenKW/tool/modelcreater"
)

const (
	workDir       = "work/"
	modelDir      = "model/"
	repositoryDir = "repository/"
)

func init() {
	dirs := []string{
		fmt.Sprintf("./%s%s", workDir, modelDir),
		fmt.Sprintf("./%s%s", workDir, repositoryDir),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0777); err != nil {
			fmt.Println(err)
		}
	}
}

func main() {
	schemaName := os.Getenv("SCHEMA_NAME")
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@/%s", "root", "password", schemaName))
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	mc, err := modelcreater.NewModelCreater(db, "users", "user_accounts")
	if err != nil {
		panic(err.Error())
	}

	ms, err := mc.GetModels()
	if err != nil {
		panic(err.Error())
	}

	for _, m := range ms {
		// err = createGofile(m, mc)
		// if err != nil {
		// 	panic(err.Error())
		// }

		output := fmt.Sprintf("./%s%s%s.go", workDir, modelDir, m.TableName[0:len(m.TableName)-1])
		err = mc.ModelCreate(output)
		if err != nil {
			panic(err.Error())
		}

		err = createRepository(m, mc)
		if err != nil {
			panic(err.Error())
		}
	}

	log.Println("Success!")
}

func createGofile(m *modelcreater.Model, mc modelcreater.ModelCreater) error {

	in := map[string]interface{}{
		"ModelName": m.Name,
		"TableName": m.TableName,
		"Columns":   m.Columns,
	}
	fm := template.FuncMap{
		"Func": func(cs *modelcreater.Column, index int) string {
			return fmt.Sprintf("%s %s %s", cs.Name, cs.Type, "")
		},
	}

	return mc.Execute(
		fmt.Sprintf("./%s%s%s.go", workDir, modelDir, m.TableName[0:len(m.TableName)-1]),
		"tmp.go.tpl",
		in,
		fm,
	)
}

func createRepository(m *modelcreater.Model, mc modelcreater.ModelCreater) error {

	in := map[string]interface{}{
		"ModelName": m.Name,
		"TableName": m.TableName,
		"Columns":   m.Columns,
	}
	fm := template.FuncMap{
		"FindArgs": func(cs modelcreater.Columns) string {
			var args string
			for _, c := range cs {
				if c.IsPk {
					name := string(c.Name)
					args += fmt.Sprintf(", %s model.%s", modelcreater.ConvertSnakeToCamel(name, false), modelcreater.ConvertSnakeToCamel(name, true))
				}
			}
			if len(args) != 0 {
				args = args[2:]
			}

			return args
		},
	}

	return mc.Execute(
		fmt.Sprintf("./%s%s%s.go", workDir, repositoryDir, m.TableName[0:len(m.TableName)-1]),
		"database.go.tpl",
		in,
		fm,
	)
}
