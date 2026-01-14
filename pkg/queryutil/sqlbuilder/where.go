package sqlbuilder

import (
	"fmt"
	"strings"

	commonv1 "github.com/belikoooova/hackaton-platform-api/api/common/v1"
)

type WhereBuilder struct {
	fieldMapping FieldMapping
	argCounter   int
	args         []interface{}
}

func NewWhereBuilder(fieldMapping FieldMapping) *WhereBuilder {
	return &WhereBuilder{
		fieldMapping: fieldMapping,
		argCounter:   1,
		args:         []interface{}{},
	}
}

func (wb *WhereBuilder) SetArgCounter(counter int) {
	wb.argCounter = counter
}

func (wb *WhereBuilder) Build(filterGroups []*FilterGroup) WhereClause {
	if len(filterGroups) == 0 {
		return WhereClause{SQL: "", Args: []interface{}{}}
	}

	groupClauses := []string{}

	for _, group := range filterGroups {
		if len(group.Filters) == 0 {
			continue
		}

		filterClauses := []string{}
		for _, filter := range group.Filters {
			clause := wb.buildFilterClause(filter)
			if clause != "" {
				filterClauses = append(filterClauses, clause)
			}
		}

		if len(filterClauses) > 0 {
			groupClauses = append(groupClauses, "("+strings.Join(filterClauses, " AND ")+")")
		}
	}

	if len(groupClauses) == 0 {
		return WhereClause{SQL: "", Args: []interface{}{}}
	}

	return WhereClause{
		SQL:  strings.Join(groupClauses, " OR "),
		Args: wb.args,
	}
}

func (wb *WhereBuilder) buildFilterClause(filter *Filter) string {
	column := wb.getColumnName(filter.Field)
	if column == "" {
		return ""
	}

	switch filter.Operation {
	case commonv1.FilterOperation_FILTER_OPERATION_EQUAL:
		return wb.buildEqual(column, filter.Value)

	case commonv1.FilterOperation_FILTER_OPERATION_IN:
		return wb.buildIn(column, filter.Value)

	case commonv1.FilterOperation_FILTER_OPERATION_CONTAINS:
		return wb.buildContains(column, filter.Value)

	case commonv1.FilterOperation_FILTER_OPERATION_PREFIX:
		return wb.buildPrefix(column, filter.Value)

	default:
		return ""
	}
}

func (wb *WhereBuilder) buildEqual(column string, value interface{}) string {
	wb.args = append(wb.args, value)
	placeholder := wb.nextPlaceholder()
	return fmt.Sprintf("%s = %s", column, placeholder)
}

func (wb *WhereBuilder) buildIn(column string, value interface{}) string {
	list, ok := value.([]interface{})
	if !ok {
		return ""
	}

	if len(list) == 0 {
		return ""
	}

	placeholders := []string{}
	for _, item := range list {
		wb.args = append(wb.args, item)
		placeholders = append(placeholders, wb.nextPlaceholder())
	}

	return fmt.Sprintf("%s IN (%s)", column, strings.Join(placeholders, ", "))
}

func (wb *WhereBuilder) buildContains(column string, value interface{}) string {
	strValue, ok := value.(string)
	if !ok {
		return ""
	}

	wb.args = append(wb.args, "%"+strValue+"%")
	placeholder := wb.nextPlaceholder()
	return fmt.Sprintf("%s ILIKE %s", column, placeholder)
}

func (wb *WhereBuilder) buildPrefix(column string, value interface{}) string {
	strValue, ok := value.(string)
	if !ok {
		return ""
	}

	wb.args = append(wb.args, strValue+"%")
	placeholder := wb.nextPlaceholder()
	return fmt.Sprintf("%s ILIKE %s", column, placeholder)
}

func (wb *WhereBuilder) getColumnName(field string) string {
	if column, ok := wb.fieldMapping[field]; ok {
		return column
	}
	return ""
}

func (wb *WhereBuilder) nextPlaceholder() string {
	placeholder := fmt.Sprintf("$%d", wb.argCounter)
	wb.argCounter++
	return placeholder
}
