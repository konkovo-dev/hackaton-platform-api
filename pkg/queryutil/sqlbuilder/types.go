package sqlbuilder

import commonv1 "github.com/belikoooova/hackaton-platform-api/api/common/v1"

type FilterGroup struct {
	Filters []*Filter
}

type Filter struct {
	Field     string
	Operation commonv1.FilterOperation
	Value     interface{}
}

type SortField struct {
	Field     string
	Direction string
}

type FieldMapping map[string]string

type WhereClause struct {
	SQL  string
	Args []interface{}
}
