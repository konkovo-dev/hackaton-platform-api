package queryutil

import (
	"fmt"
	"strings"

	"github.com/belikoooova/hackaton-platform-api/pkg/queryutil/sqlbuilder"
)

type QueryBuilder struct {
	baseQuery    string
	whereClauses []string
	args         []interface{}
	argCounter   int
	sortFields   []string
	limit        int
}

func NewQueryBuilder(baseQuery string) *QueryBuilder {
	return &QueryBuilder{
		baseQuery:    baseQuery,
		whereClauses: []string{},
		args:         []interface{}{},
		argCounter:   1,
		sortFields:   []string{},
	}
}

func (qb *QueryBuilder) WithSearch(fields []string, searchQuery string) *QueryBuilder {
	if searchQuery == "" || len(fields) == 0 {
		return qb
	}

	searchPattern := "%" + searchQuery + "%"
	conditions := make([]string, len(fields))

	for i, field := range fields {
		conditions[i] = fmt.Sprintf("%s ILIKE $%d", field, qb.argCounter)
		qb.args = append(qb.args, searchPattern)
		qb.argCounter++
	}

	qb.whereClauses = append(qb.whereClauses,
		"("+strings.Join(conditions, " OR ")+")")

	return qb
}

func (qb *QueryBuilder) WithFilters(filterGroups []*sqlbuilder.FilterGroup, fieldMapping sqlbuilder.FieldMapping) *QueryBuilder {
	if len(filterGroups) == 0 {
		return qb
	}

	wb := sqlbuilder.NewWhereBuilder(fieldMapping)
	wb.SetArgCounter(qb.argCounter)

	whereClause := wb.Build(filterGroups)
	if whereClause.SQL != "" {
		qb.whereClauses = append(qb.whereClauses, whereClause.SQL)
		qb.args = append(qb.args, whereClause.Args...)
		qb.argCounter += len(whereClause.Args)
	}

	return qb
}

func (qb *QueryBuilder) WithCursor(fields []CursorField) *QueryBuilder {
	if len(fields) == 0 {
		return qb
	}

	fieldExprs := make([]string, len(fields))
	placeholders := make([]string, len(fields))

	for i, field := range fields {
		fieldExprs[i] = field.Column
		placeholders[i] = fmt.Sprintf("$%d", qb.argCounter)
		qb.args = append(qb.args, field.Value)
		qb.argCounter++
	}

	operator := ">"
	if len(fields) > 0 && fields[0].Descending {
		operator = "<"
	}

	qb.whereClauses = append(qb.whereClauses,
		fmt.Sprintf("(%s) %s (%s)",
			strings.Join(fieldExprs, ", "),
			operator,
			strings.Join(placeholders, ", ")))

	return qb
}

func (qb *QueryBuilder) WithOrderBy(fields []OrderByField) *QueryBuilder {
	for _, field := range fields {
		direction := "ASC"
		if field.Descending {
			direction = "DESC"
		}
		qb.sortFields = append(qb.sortFields, field.Column+" "+direction)
	}
	return qb
}

func (qb *QueryBuilder) WithCustomWhere(clause string, args ...interface{}) *QueryBuilder {
	if clause == "" {
		return qb
	}

	adjustedClause := clause
	for range args {
		adjustedClause = strings.Replace(adjustedClause, "?", fmt.Sprintf("$%d", qb.argCounter), 1)
		qb.argCounter++
	}

	qb.whereClauses = append(qb.whereClauses, adjustedClause)
	qb.args = append(qb.args, args...)

	return qb
}

func (qb *QueryBuilder) WithLimit(limit int) *QueryBuilder {
	qb.limit = limit
	return qb
}

func (qb *QueryBuilder) Build() (string, []interface{}) {
	query := qb.baseQuery

	if len(qb.whereClauses) > 0 {
		query += " WHERE " + strings.Join(qb.whereClauses, " AND ")
	}

	if len(qb.sortFields) > 0 {
		query += " ORDER BY " + strings.Join(qb.sortFields, ", ")
	}

	if qb.limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", qb.argCounter)
		qb.args = append(qb.args, qb.limit)
	}

	return query, qb.args
}

type CursorField struct {
	Column     string
	Value      interface{}
	Descending bool
}

type OrderByField struct {
	Column     string
	Descending bool
}
