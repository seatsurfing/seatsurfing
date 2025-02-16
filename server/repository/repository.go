package repository

import "database/sql"

type Repository interface {
	RunSchemaUpgrade(curVersion, targetVersion int)
}

func CheckNullString(s NullString) sql.NullString {
	if len(s) == 0 {
		return sql.NullString{}
	}
	return sql.NullString{
		String: string(s),
		Valid:  true,
	}
}

func MaxOf(vars ...int) int {
	max := vars[0]

	for _, i := range vars {
		if max < i {
			max = i
		}
	}

	return max
}
