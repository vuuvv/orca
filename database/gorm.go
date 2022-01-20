package database

import (
	"github.com/vuuvv/errors"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)

func NewGorm(config *Config) (db *gorm.DB, err error) {
	gConfig := &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
	}
	if config.Debug {
		gConfig.Logger = logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
			SlowThreshold: 200 * time.Millisecond,
			LogLevel:      logger.Info,
			Colorful:      true,
		})
	}
	switch config.Type {
	case "postgres":
		db, err = gorm.Open(postgres.New(postgres.Config{
			DSN:                  config.Dsn,
			PreferSimpleProtocol: true,
		}), gConfig)
	default:
		db, err = gorm.Open(mysql.Open(config.Dsn), gConfig)
	}
	if err != nil {
		return nil, errors.WithStack(err)
	}
	zap.L().Info("gorm客户端初始化成功", zap.String("dns", config.Dsn))
	return db, nil
}
