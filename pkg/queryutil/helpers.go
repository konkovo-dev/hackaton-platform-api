package queryutil

import commonv1 "github.com/belikoooova/hackaton-platform-api/api/common/v1"

func GetStringValue(f *commonv1.Filter) string {
	if v, ok := f.Value.(*commonv1.Filter_StringValue); ok {
		return v.StringValue
	}
	return ""
}

func GetStringListValue(f *commonv1.Filter) []string {
	if v, ok := f.Value.(*commonv1.Filter_StringList); ok && v.StringList != nil {
		return v.StringList.Values
	}
	return nil
}

func GetSortDirection(dir commonv1.SortDirection) string {
	if dir == commonv1.SortDirection_SORT_DIRECTION_DESC {
		return "DESC"
	}
	return "ASC"
}
