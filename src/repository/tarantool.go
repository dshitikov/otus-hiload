package repository

import (
	"errors"
	"github.com/tarantool/go-tarantool"
	"log"
)

type tarantoolRepo struct {
	conn *tarantool.Connection
}

type ITarantoolRepository interface {
	FindByNamePrefix(prefix string, limit int, minId int64) ([]*User, error)
	FindByNamePrefixSQL(prefix string, limit int, minId int64) ([]*User, error)
}

func NewTarantoolRepository(hostPort string) ITarantoolRepository {
	opts := tarantool.Opts{User: "guest"}
	conn, err := tarantool.Connect(hostPort, opts)
	if err != nil {
		log.Fatal("Connection refused:", err)
	}
	return &tarantoolRepo{conn: conn}
}

func (r *tarantoolRepo) FindByNamePrefixSQL(prefix string, limit int, minId int64) ([]*User, error) {
	res, err := r.conn.Call17("find_users_sql", []interface{}{prefix, minId, limit})
	if err != nil {
		return nil, err
	}

	if len(res.Data) != 1 {
		return nil, errors.New("invalid response Data length")
	}
	mm, ok := res.Data[0].(map[interface{}]interface{})
	if !ok {
		return nil, errors.New("conversion error")
	}
	// meta := mm["metadata"]
	rows := mm["rows"]
	rowsArr, ok := rows.([]interface{})
	if !ok {
		return nil, errors.New("conversion error")
	}
	var users = make([]*User, 0, len(rowsArr))
	for _, v := range rowsArr {
		valsArr, ok := v.([]interface{})
		if !ok {
			return nil, errors.New("conversion error")
		}
		if len(valsArr) != 3 {
			return nil, errors.New("conversion error")
		}
		var user = new(User)
		user.ID = int64(valsArr[0].(uint64))
		user.Name = valsArr[1].(string)
		user.LastName = valsArr[2].(string)
		users = append(users, user)
	}

	return users, nil
}

func (r *tarantoolRepo) FindByNamePrefix(prefix string, limit int, minId int64) ([]*User, error) {
	res, err := r.conn.Call17("find_users", []interface{}{prefix, minId, limit})
	if err != nil {
		return nil, err
	}

	if len(res.Data) != 1 {
		return nil, errors.New("invalid response Data length")
	}
	// log.Printf("%#+v", res.Data[0])
	// mm, ok := res.Data[0].(map[interface{}]interface{})
	// if !ok {
	// 	return nil, errors.New("conversion error")
	// }
	// // meta := mm["metadata"]
	// rows := mm["rows"]
	rowsArr, ok := res.Data[0].([]interface{})
	if !ok {
		return nil, errors.New("conversion error")
	}
	var users = make([]*User, 0, len(rowsArr))
	for _, v := range rowsArr {
		valsArr, ok := v.([]interface{})
		if !ok {
			return nil, errors.New("conversion error")
		}
		if len(valsArr) != 3 {
			return nil, errors.New("conversion error")
		}
		var user = new(User)
		user.ID = int64(valsArr[0].(uint64))
		user.Name = valsArr[1].(string)
		user.LastName = valsArr[2].(string)
		users = append(users, user)
	}

	return users, nil
}
