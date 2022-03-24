package orm

import (
	"fmt"
	"github.com/vuuvv/errors"
	"github.com/vuuvv/orca/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Sequence struct {
	Key   string `json:"key" gorm:"size:60;primaryKey"`
	Value int    `json:"value"`
}

func (*Sequence) TableName() string {
	return "t_sequence"
}

func (*Sequence) TableTitle() string {
	return "序列号"
}

type sequenceService struct {
}

func (s *sequenceService) NextId(db *gorm.DB, key string) (value int, err error) {
	err = db.Transaction(func(tx *gorm.DB) error {
		seq := &Sequence{}

		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(seq, fmt.Sprintf("%s = ?", Quote(tx, "key")), key).Error
		if err == gorm.ErrRecordNotFound {
			// 新的记录则插入
			seq.Key = key
			seq.Value = 1
			err = tx.Save(seq).Error
			if err != nil {
				return err
			}
			value = 1
			return nil
		}
		if err != nil {
			return err
		}
		// 旧的记录则更新
		seq.Value++
		value = seq.Value
		return tx.Model(seq).Update("value", seq.Value).Error
	})
	if err != nil {
		return 0, errors.WithStack(err)
	}
	return value, nil
}

func (s *sequenceService) NextCode(db *gorm.DB, key string) (code string, err error) {
	id, err := s.NextId(db, key)
	code, err = utils.EncodeIntToBase64Like(id)
	if err != nil {
		return "", errors.WithStack(err)
	}
	return code, nil
}

var SequenceService = &sequenceService{}
