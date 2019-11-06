package repository

import (
	"database/sql"
)

type Test struct {
	ID        int64
	CreatedAt sql.NullTime
}

type ITestRepository interface {
	CreateTest(test *Test) (*Test, error)
}

func (r *repo) CreateTest(test *Test) (*Test, error) {
	res, err := r.db.Exec("INSERT INTO test(created_at) VALUES(NOW())")

	if err != nil {
		return nil, err
	}

	userID, err := res.LastInsertId()

	if err != nil {
		return nil, err
	}

	test.ID = userID

	return test, nil
}
