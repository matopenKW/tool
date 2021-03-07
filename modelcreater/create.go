package modelcreater

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"

	_ "github.com/go-sql-driver/mysql"
)

type Creater struct {
	Template *template.Template
	Value    map[string]interface{}
	FuncMap  template.FuncMap
}

type ModelCreater interface {
	GetModel() (Models, error)
	Execute(string, string, map[string]interface{}, template.FuncMap) error
}

type modelcreater struct {
	db      *sql.DB
	tables  []TableName
	Creater *Creater
}

func NewModelCreater(db *sql.DB, ss ...string) (*modelcreater, error) {
	var tables []TableName
	for _, s := range ss {
		tables = append(tables, TableName(s))
	}
	return &modelcreater{db, tables, nil}, nil
}

func (r *modelcreater) GetModel() (Models, error) {
	ts, err := r.getTables()
	if err != nil {
		log.Println("Get Tables Error")
		return nil, err
	}

	acim, err := r.getAllColumnInfosMap(ts)
	if err != nil {
		return nil, err
	}

	fkMap, err := r.getAllForeignMap(acim)
	if err != nil {
		return nil, err
	}

	var ms []*Model
	for t, cis := range acim {
		exists := func(t TableName) bool {
			for _, table := range r.tables {
				if table == t {
					return true
				}
			}
			return false
		}
		if !exists(t) {
			continue
		}

		fs := fkMap[t]
		cs, err := r.getColumnList(cis, fs)
		if err != nil {
			return nil, err
		}

		mn := ConvertSnakeToCamel(string(t), true)
		ms = append(ms, &Model{
			Name:      mn[:len(mn)-1],
			TableName: t,
			Columns:   cs,
		})
	}

	if len(ms) == 0 {
		log.Println(fmt.Sprintf("Not Exsits Table. tableName=%v", r.tables))
	}

	return ms, err
}

func (r *modelcreater) Execute(goName, tplName string, in map[string]interface{}, fm template.FuncMap) error {
	f, err := os.Create(goName)
	if err != nil {
		return err
	}
	defer f.Close()

	return template.Must(template.New(tplName).Funcs(fm).ParseGlob(fmt.Sprintf("./template/%s", tplName))).Execute(f, in)
}

func (r *modelcreater) getAllColumnInfosMap(ts []TableName) (map[TableName]ColumnInfos, error) {
	cisMap := make(map[TableName]ColumnInfos)
	for _, t := range ts {
		values, err := r.selectQuery(fmt.Sprintf("DESCRIBE %s", t))
		if err != nil {
			return nil, err
		}
		if len(values) == 0 {
			return nil, errors.New("Failed get column")
		}

		cis := make(ColumnInfos, 0)
		for _, v := range values {
			if _, ok := v["Field"]; !ok {
				return nil, errors.New("Field is empty")
			} else if _, ok := v["Type"]; !ok {
				return nil, errors.New("Type is empty")
			} else if _, ok := v["Null"]; !ok {
				return nil, errors.New("Null is empty")
			} else if _, ok := v["Key"]; !ok {
				return nil, errors.New("Key is empty")
			}

			cis = append(cis, &ColumnInfo{
				Field: v["Field"],
				Type:  v["Type"],
				Null:  v["Null"],
				Key:   v["Key"],
			})
		}
		cisMap[t] = cis
	}

	return cisMap, nil
}

func (r *modelcreater) getAllForeignMap(cisMap map[TableName]ColumnInfos) (map[TableName]Foreigns, error) {
	query := "SELECT TABLE_NAME, COLUMN_NAME, REFERENCED_TABLE_NAME, REFERENCED_COLUMN_NAME FROM information_schema.key_column_usage WHERE CONSTRAINT_NAME LIKE 'fk_%'"
	values, err := r.selectQuery(query)
	if err != nil {
		return nil, err
	}

	ret := make(map[TableName]Foreigns)
	for _, v := range values {
		tn, ok := v["TABLE_NAME"]
		if !ok {
			return nil, fmt.Errorf("Invalid FK table name. values=%v", v)
		}
		cn, ok := v["COLUMN_NAME"]
		if !ok {
			return nil, fmt.Errorf("Invalid FK column name. values=%v", v)
		}
		rtn, ok := v["REFERENCED_TABLE_NAME"]
		if !ok {
			return nil, fmt.Errorf("Invalid FK referenced table name. values=%v", v)
		}
		rcn, ok := v["REFERENCED_COLUMN_NAME"]
		if !ok {
			return nil, fmt.Errorf("Invalid FK referenced column name. values=%v", v)
		}

		fkColumns, ok := cisMap[TableName(rtn)]
		if !ok {
			return nil, fmt.Errorf("Not Exsits referenced table. table name=%s", rtn)
		}
		var fkType string
		for _, v := range fkColumns {
			if rcn == v.Field {
				fkType = v.Type
				continue
			}
		}

		ts, ok := ret[TableName(tn)]
		if !ok {
			ts = make(Foreigns, 0)
		}

		ret[TableName(tn)] = append(ts, &Foreign{
			TableName:            TableName(tn),
			ColumnName:           ColumnName(cn),
			ReferencedTableName:  TableName(rtn),
			ReferencedColumnName: rcn,
			ReferencedColumnType: fkType,
		})
	}

	return ret, nil
}

func (r *modelcreater) getColumnList(cis ColumnInfos, fs Foreigns) (Columns, error) {
	ignore := func(s ColumnName) bool {
		if s == "created_at" || s == "updated_at" {
			return true
		}
		return false
	}

	fkMap := make(map[ColumnName]*Foreign)
	for _, f := range fs {
		fkMap[f.ColumnName] = f
	}

	cs := make(Columns, 0)
	for _, v := range cis {
		cn := ColumnName(v.Field)
		if ignore(cn) {
			continue
		}
		cs = append(cs, &Column{
			Name:        cn,
			Type:        v.Type,
			IsPk:        v.Key == "PRI",
			NullAllowed: v.Null == "YES",
			Foreign:     fkMap[cn],
		})
	}
	return cs, nil
}

func (r *modelcreater) getColumnType(typeStr string) string {
	switch typeStr {
	case "text":
		return "string"
	case "datetime", "timestamp":
		return "time.Time"
	case "double":
		return "float64"
	default:
		if strings.Index(typeStr, "varchar") > -1 {
			return "string"
		} else if strings.Index(typeStr, "int") > -1 {
			return "int"
		} else if strings.Index(typeStr, "tinyint") > -1 {
			return "bool"
		}
		return ""
	}
}

func (r *modelcreater) getTables() ([]TableName, error) {
	query := "SHOW TABLES"
	values, err := r.selectQuery(query)
	if err != nil {
		return nil, err
	}

	tables := make([]TableName, 0)
	for _, m := range values {
		for _, v := range m {
			tables = append(tables, TableName(v))
		}
	}
	return tables, nil
}

func (r *modelcreater) getColumnInfos(t TableName) (ColumnInfos, error) {
	values, err := r.selectQuery(fmt.Sprintf("DESCRIBE %s", t))
	if err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return nil, errors.New("Failed get column")
	}

	cs := make(ColumnInfos, 0)
	for _, v := range values {
		if _, ok := v["Field"]; !ok {
			return nil, errors.New("Field is empty")
		} else if _, ok := v["Type"]; !ok {
			return nil, errors.New("Type is empty")
		} else if _, ok := v["Null"]; !ok {
			return nil, errors.New("Null is empty")
		} else if _, ok := v["Key"]; !ok {
			return nil, errors.New("Key is empty")
		}

		cs = append(cs, &ColumnInfo{
			Field: v["Field"],
			Type:  v["Type"],
			Null:  v["Null"],
			Key:   v["Key"],
		})
	}

	return cs, nil
}

func (r *modelcreater) selectQuery(query string) ([]map[string]string, error) {
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	values := make([]sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	ret := make([]map[string]string, 0)
	for rows.Next() {
		row := make(map[string]string)
		err = rows.Scan(scanArgs...)
		if err != nil {
			return nil, err
		}

		var value string
		for i, col := range values {
			if col == nil {
				value = "NULL"
			} else {
				value = string(col)
			}
			row[columns[i]] = value
		}
		ret = append(ret, row)
	}

	return ret, nil
}

func (r *modelcreater) setModelTemp(m *Model) error {
	in := map[string]interface{}{
		"ModelName": m.Name,
		"TableName": m.TableName,
		"Columns":   m.Columns,
	}
	fm := template.FuncMap{
		"Func": func(cs *Column, index int) string {
			return fmt.Sprintf("%s %s %s", cs.Name, cs.Type, "")
		},
	}

	modelTemp := "model.go.tpl"
	temp, err := template.New(modelTemp).Funcs(fm).ParseGlob(fmt.Sprintf("./template/%s", modelTemp))
	if err != nil {
		return err
	}

	r.Creater = &Creater{
		Template: temp,
		Value:    in,
		FuncMap:  fm,
	}
	return nil
}
