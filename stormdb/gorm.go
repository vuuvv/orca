package stormdb

import (
	"github.com/asdine/storm/v3"
	"github.com/vuuvv/errors"
	"go.etcd.io/etcd/client/pkg/v3/fileutil"
	"go.uber.org/zap"
	"os"
	"path/filepath"
)

func NewStorm(config *Config) (db *storm.DB, err error) {
	if config.Path == "" {
		config.Path = "storm.db"
	}
	p, err := filepath.Abs(config.Path)
	if err != nil {
		return nil, errors.Errorf("Initialize storm db error: %s, path: %s", err, config.Path)
	}
	dir := filepath.Dir(p)
	if !fileutil.Exist(dir) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, errors.Errorf("Initialize storm db error: %s, path: %s", err, config.Path)
		}
	}
	db, err = storm.Open(config.Path)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	zap.L().Info("Storm DB initialize success", zap.String("path", config.Path))
	return db, nil
}
