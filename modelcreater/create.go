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

const (
	schemaName = ""
)

type ModelCreater interface {
	Execute() error
}

type modelcreater struct {
	db *sql.DB
}

func NewModelCreater() (*modelcreater, error) {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@/%s", "root", "password", ""))
	if err != nil {
		return nil, err
	}
	return &modelcreater{db}, nil

}

func (r *modelcreater) Execute() (Models, error) {
	defer r.db.Close()
	tableName := TableName("users")
	ts, err := r.getTables(tableName)
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

	if len(ms) != 0 {
		for _, m := range ms {
			log.Println(m)
		}
		log.Println("Success!")
	} else {
		log.Println(fmt.Sprintf("Not Exsits Table. tableName=%s", tableName))
	}

	return nil, err
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

func (r *modelcreater) getTables(t TableName) ([]TableName, error) {
	query := "SHOW TABLES"
	if t != "" {
		query += fmt.Sprintf(" WHERE Tables_in_%s = '%s'", t)
	}
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

func (r *modelcreater) createGofile(goName, tplName string, in interface{}, fm map[string]interface{}) error {
	f, err := os.Create(goName)
	if err != nil {
		return err
	}
	defer f.Close()

	funcMap := r.setBaseFuncMap()
	for k, v := range fm {
		funcMap[k] = v
	}

	err = template.Must(template.New(tplName).Funcs(funcMap).ParseGlob(fmt.Sprintf("./template/%s", tplName))).Execute(f, in)
	if err != nil {
		return err
	}
	return nil
}

func (r *modelcreater) setBaseFuncMap() template.FuncMap {
	return template.FuncMap{}
}
