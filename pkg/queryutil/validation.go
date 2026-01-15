package queryutil

import commonv1 "github.com/belikoooova/hackaton-platform-api/api/common/v1"

func IsValidStringOperation(op commonv1.FilterOperation) bool {
	return op == commonv1.FilterOperation_FILTER_OPERATION_EQUAL ||
		op == commonv1.FilterOperation_FILTER_OPERATION_PREFIX ||
		op == commonv1.FilterOperation_FILTER_OPERATION_CONTAINS
}

func IsValidInOperation(op commonv1.FilterOperation) bool {
	return op == commonv1.FilterOperation_FILTER_OPERATION_IN
}
