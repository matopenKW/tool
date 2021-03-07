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
	workDir  = "work/"
	modelDir = "model/"
)

func init() {
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

	ms, err := mc.Execute()
	if err != nil {
		panic(err.Error())
	}

	for _, m := range ms {
		log.Println(m.Name)
		log.Println(m.TableName)
		for _, c := range m.Columns {
			log.Println(c)
		}

		err = createGofile(m, mc)
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
		"Columns":   m.Name,
	}
	fm := template.FuncMap{
		"Func": func(cs *modelcreater.Column, index int) string {
			return fmt.Sprintf("%s %s %s", cs.Name, cs.Type, "")
		},
	}
	return mc.CreateGofile(
		fmt.Sprintf("./%s%s%s.go", workDir, modelDir, m.TableName[0:len(m.TableName)-1]),
		"tmp.go.tpl",
		in,
		fm,
	)
}
