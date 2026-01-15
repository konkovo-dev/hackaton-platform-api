package users

import "github.com/belikoooova/hackaton-platform-api/pkg/queryutil/sqlbuilder"

// Field names for users queries
const (
	FieldUserID    = "user_id"
	FieldUsername  = "username"
	FieldFirstName = "first_name"
	FieldLastName  = "last_name"
	FieldSkills    = "skills"
)

// Database column names
const (
	ColumnUserID    = "u.id"
	ColumnUsername  = "u.username"
	ColumnFirstName = "u.first_name"
	ColumnLastName  = "u.last_name"
)

// GetFieldMapping returns the mapping from field names to database columns
func GetFieldMapping() sqlbuilder.FieldMapping {
	return sqlbuilder.FieldMapping{
		FieldUsername:  ColumnUsername,
		FieldUserID:    ColumnUserID,
		FieldFirstName: ColumnFirstName,
		FieldLastName:  ColumnLastName,
	}
}
