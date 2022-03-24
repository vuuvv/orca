package orm

import (
	"github.com/go-sql-driver/mysql"
	"github.com/jackc/pgconn"
	"github.com/vuuvv/errors"
	"strings"
)

func DuplicateMessage(err error, messages map[string]string) error {
	rawErr := err
	if errors.HasStack(rawErr) {
		rawErr = errors.Unwrap(rawErr)
		if rawErr == nil {
			rawErr = err
		}
	}
	switch val := rawErr.(type) {
	case *mysql.MySQLError:
		if val.Number == 1062 {
			for k, v := range messages {
				if strings.Contains(val.Message, k) {
					return errors.Wrap(err, v)
				}
			}
			return errors.Wrap(err, "插入重复数据")
		}

	case *pgconn.PgError:
		if val.Code == "23505" {
			for k, v := range messages {
				if strings.Contains(val.Message, k) {
					return errors.Wrap(err, v)
				}
			}
			return errors.Wrap(err, "插入重复数据")
		}
	}

	return errors.WithStack(err)
}

func IsDuplicateError(err error) bool {
	rawErr := err
	if errors.HasStack(rawErr) {
		rawErr = errors.Unwrap(rawErr)
		if rawErr == nil {
			rawErr = err
		}
	}
	switch val := rawErr.(type) {
	case *mysql.MySQLError:
		if val.Number == 1062 {
			return true
		}

	case *pgconn.PgError:
		if val.Code == "23505" {
			return true
		}
	}
	return false
}
