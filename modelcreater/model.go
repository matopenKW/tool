package modelcreater

type TableName string
type ColumnName string

type Model struct {
	Name      string
	TableName TableName
	Columns   Columns
}
type Models []*Model

type ColumnInfo struct {
	Field string
	Type  string
	Null  string
	Key   string
}

type ColumnInfos []*ColumnInfo

type Column struct {
	Name        ColumnName
	Type        string
	IsPk        bool
	NullAllowed bool
	Foreign     *Foreign
}

type Columns []*Column

func (c Columns) GetPKList() Columns {
	var s Columns
	for _, v := range c {
		if v.IsPk {
			s = append(s, v)
		}
	}
	return s
}

type Foreign struct {
	TableName            TableName
	ColumnName           ColumnName
	ReferencedTableName  TableName
	ReferencedColumnName string
	ReferencedColumnType string
}

type Foreigns []*Foreign
