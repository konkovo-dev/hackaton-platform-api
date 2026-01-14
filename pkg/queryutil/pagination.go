package queryutil

import commonv1 "github.com/belikoooova/hackaton-platform-api/api/common/v1"

const (
	DefaultPageSize = 20
	MaxPageSize     = 100
)

type PaginationParams struct {
	PageSize    int
	PageToken   string
	DefaultSize int
	MaxSize     int
}

func ParsePagination(page *commonv1.PageRequest, defaultSize, maxSize int) PaginationParams {
	if defaultSize == 0 {
		defaultSize = DefaultPageSize
	}
	if maxSize == 0 {
		maxSize = MaxPageSize
	}

	pageSize := defaultSize
	pageToken := ""

	if page != nil {
		if page.PageSize > 0 {
			pageSize = int(page.PageSize)
			if pageSize > maxSize {
				pageSize = maxSize
			}
		}
		pageToken = page.PageToken
	}

	return PaginationParams{
		PageSize:    pageSize,
		PageToken:   pageToken,
		DefaultSize: defaultSize,
		MaxSize:     maxSize,
	}
}

func ShouldFetchMore(pageSize int) int {
	return pageSize + 1
}

func HasMoreResults(resultsCount, pageSize int) bool {
	return resultsCount > pageSize
}

func TrimResults[T any](results []T, pageSize int) []T {
	if len(results) > pageSize {
		return results[:pageSize]
	}
	return results
}
