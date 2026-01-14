package sqlbuilder

import (
	"fmt"
	"strings"
)

func BuildOrderBy(sortFields []SortField, fieldMapping FieldMapping) string {
	if len(sortFields) == 0 {
		return ""
	}

	clauses := []string{}

	for _, sf := range sortFields {
		column := fieldMapping[sf.Field]
		if column == "" {
			continue
		}

		direction := "ASC"
		if sf.Direction == "DESC" {
			direction = "DESC"
		}

		clauses = append(clauses, fmt.Sprintf("%s %s", column, direction))
	}

	if len(clauses) == 0 {
		return ""
	}

	return strings.Join(clauses, ", ")
}
