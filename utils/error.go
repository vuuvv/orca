package utils

import (
	"github.com/go-redis/redis/v8"
	"github.com/vuuvv/errors"
	"gorm.io/gorm"
)

func RecordNotFound(err error) bool {
	//return errors.Is(err, redis.Nil) || errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, errors.Nil)
	return errors.Is(err, redis.Nil) || errors.Is(err, gorm.ErrRecordNotFound)
}

// WrapRecordNotFound 如果是记录不存在错误，生成新的错误
func WrapRecordNotFound(err error, recordNotFoundMsg string) error {
	if err == nil {
		return nil
	}
	if RecordNotFound(err) {
		return errors.Wrapf(err, recordNotFoundMsg)
	}
	return errors.WithStack(err)
}

//func IsDuplicateError(err error) bool {
//	if errors.(err) {
//		err = errors.Unwrap(err)
//	}
//	if mysqlError, ok := err.(*mysql.MySQLError); ok {
//		if mysqlError.Number == 1062 {
//			return true
//		}
//	}
//	return false
//}

//func DuplicateMessage(err error, messages map[string]string) error {
//	if mysqlError, ok := err.(*mysql.MySQLError); ok {
//		if mysqlError.Number == 1062 {
//			for k, v := range messages {
//				if strings.Contains(mysqlError.Message, k) {
//					return errors.WrapWithMessage(err, v)
//				}
//			}
//			return  errors.WrapWithMessage(err, "插入重复数据")
//		}
//	}
//	return errors.Wrap(err)
//}

//func LoggerError(c *gin.Context, err interface{}, start time.Time) {
//	// Check for a broken connection, as it is not really a
//	// condition that warrants a panic stack trace.
//	var brokenPipe bool
//	if ne, ok := err.(*net.OpError); ok {
//		if se, ok := ne.Err.(*os.SyscallError); ok {
//			if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
//				brokenPipe = true
//			}
//		}
//	}
//
//	fields := ZapFieldsFromGinContext(c, start)
//	fields = append(fields, zap.Any("error", err))
//	fields = append(fields, zap.Bool("recover", true))
//	zap.L().Warn("", fields...)
//
//	//httpRequest, _ := httputil.DumpRequest(c.Request, false)
//	if brokenPipe {
//		// If the connection is dead, we can't write a status to it.
//		_ = c.Error(err.(error)) // nolint: errcheck
//		c.Abort()
//		return
//	}
//	//stack := ""
//	//if errors.HasStack(err) {
//	//	stack = fmt.Sprintf("%+v", err)
//	//} else {
//	//	stack = string(debug.Stack())
//	//}
//
//	//zap.S().Error("[Recovery from panic]",
//	//	zap.Error(err.(error)),
//	//	zap.String("request", string(httpRequest)),
//	//	zap.Time("time", time.Now()),
//	//	zap.String("stack", stack),
//	//)
//}

//func ApiResponse(result interface{}, err error) interface{} {
//	if err != nil {
//		return errors.WithStack(err)
//	}
//	return result
//}
