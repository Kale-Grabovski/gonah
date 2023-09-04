package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/Kale-Grabovski/gonah/src/domain"
)

func QuerySlice[T any, Ptr interface{ *T }](
	ctx context.Context,
	db domain.DB,
	query string,
	binder func(Ptr) []any,
	postScan func(Ptr),
	args ...any,
) ([]T, error) {
	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []T
	for rows.Next() {
		var result T
		cols := binder(&result)
		err = rows.Scan(cols...)
		if err != nil {
			return nil, err
		}
		postScan(&result)
		results = append(results, result)
	}
	return results, nil
}

func joinStrings(ids []string) string {
	var inUserIds []string
	for _, userId := range ids {
		inUserIds = append(inUserIds, fmt.Sprintf("'%s'", userId))
	}
	return strings.Join(inUserIds, ",")
}
