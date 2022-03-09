package sequence

import (
	"fmt"
	"github.com/vuuvv/errors"
	"github.com/vuuvv/orca"
	"github.com/vuuvv/orca/orm"
	"github.com/vuuvv/orca/utils"
	"gorm.io/gorm"
)

type service struct {
}

func (s *service) NextId(key string) (value int, err error) {
	db := orca.Database()

	err = db.Transaction(func(tx *gorm.DB) error {
		seq := &Sequence{}

		err := orm.ForUpdate(tx).First(seq, fmt.Sprintf("%s = ?", orm.Quote(tx, "key")), key).Error
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

func (s *service) NextCode(key string) (code string, err error) {
	id, err := s.NextId(key)
	code, err = utils.EncodeIntToBase64Like(id)
	if err != nil {
		return "", errors.WithStack(err)
	}
	return code, nil
}

var Service = &service{}
