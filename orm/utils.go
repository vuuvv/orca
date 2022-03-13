package orm

import (
	"bytes"
	"fmt"
	"github.com/fatih/structs"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	jsoniter "github.com/json-iterator/go"
	"github.com/vuuvv/errors"
	"github.com/vuuvv/orca/utils"
	reflections "github.com/vuuvv/orca/utils/reflects"
	"github.com/vuuvv/orca/utils/replacer"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"reflect"
	"strings"
)

func RecordNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

func ForUpdate(db *gorm.DB) *gorm.DB {
	return db.Clauses(clause.Locking{Strength: "UPDATE"})
}

func Quote(db *gorm.DB, val string) string {
	buf := bytes.Buffer{}
	db.Dialector.QuoteTo(&buf, val)
	return buf.String()
}

func prepare(template string, vars map[string]string, valueMap map[string]interface{}, criteria map[string]*Criteria) (sql string, values []interface{}) {
	sql, names := replacer.New(template).Replace(vars)
	for _, v := range names {
		val := valueMap[v]
		if c, ok := criteria[v]; ok {
			switch c.op {
			case OP_Contain:
				val = fmt.Sprintf("%%%v%%", val)
			case OP_StartsWith:
				val = fmt.Sprintf("%v%%", val)
			case OP_EndsWith:
				val = fmt.Sprintf("%%%v", val)
			}
		}
		if val == true {
			val = 1
		} else if val == false {
			val = 0
		}
		values = append(values, val)
	}

	return sql, values
}

func GetPaginator(q string) *Paginator {
	if q == "" {
		return &Paginator{
			Page:     1,
			PageSize: 20,
		}
	}
	paginator := &Paginator{}
	utils.PanicIf(jsoniter.Unmarshal([]byte(q), paginator))
	return paginator
}

func GetPaginatorFromCtx(ctx *gin.Context) *Paginator {
	return GetPaginator(ctx.Query("q"))
}

func contains(collection []string, value string) bool {
	for _, v := range collection {
		if v == value {
			return true
		}
	}
	return false
}

func NeedUpdateFields(old interface{}, new interface{}, excludeFields ...string) []string {
	var ret []string
	oldStruct := structs.New(old)
	newStruct := structs.New(new)

	oldFields := utils.ExpandFields(oldStruct.Fields())
	newFields := utils.ExpandFields(newStruct.Fields())

	for _, f := range newFields {
		name := f.Name()
		if contains(excludeFields, name) {
			continue
		}
		for _, o := range oldFields {
			if o.Name() == name {
				if o.Value() != f.Value() {
					ret = append(ret, name)
				}
			}
		}
		//if oldField, ok := oldStruct.FieldOk(name); ok {
		//	if oldField.Value() != f.Value() {
		//		ret = append(ret, name)
		//	}
		//}
	}

	return ret
}

//func CheckWhenIdNotEmpty(db *gorm.DB, model interface{}, id int64, message string) {
//	if id == 0 {
//		return
//	}
//
//	if message == "" {
//		message = fmt.Sprintf("项目不存在: %d", id)
//	}
//
//	err := db.First(model, id).Error
//	if err == utils.ErrRecordNotFound {
//		utils.Panicf(message)
//	}
//	utils.PanicIf(err)
//}

func GetById(db *gorm.DB, model EntityType, id int64) (err error) {
	if id == 0 {
		return errors.New(fmt.Sprintf("未指定%sid", model.TableTitle()))
	}
	err = db.First(model, id).Error
	if utils.RecordNotFound(err) {
		return errors.Wrapf(err, fmt.Sprintf("%s不存在", model.TableTitle()))
	}
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func LockById(tx *gorm.DB, model EntityType, id int64) (err error) {
	if id == 0 {
		return errors.New(fmt.Sprintf("未指定%sid", model.TableTitle()))
	}
	err = ForUpdate(tx).First(model, id).Error
	if utils.RecordNotFound(err) {
		return errors.Wrapf(err, fmt.Sprintf("%s不存在", model.TableTitle()))
	}
	return errors.WithStack(err)
}

func GetBy(db *gorm.DB, model EntityType, name string, value interface{}) (err error) {
	err = db.First(model, fmt.Sprintf("%s=?", Quote(db, name)), value).Error
	if utils.RecordNotFound(err) {
		return errors.Wrapf(err, fmt.Sprintf("%s不存在", model.TableTitle()))
	}
	return errors.WithStack(err)
}

func GetByMap(db *gorm.DB, model EntityType, criteria map[string]interface{}) (err error) {
	if criteria == nil {
		return errors.New("请传入正确参数")
	}
	builder := strings.Builder{}
	var values []interface{}
	for k, v := range criteria {
		builder.WriteString(fmt.Sprintf(" AND %s=?", Quote(db, k)))
		values = append(values, v)
	}
	params := append([]interface{}{builder.String()[5:]}, values...)
	err = db.First(model, params...).Error
	if utils.RecordNotFound(err) {
		return errors.Wrapf(err, fmt.Sprintf("%s不存在", model.TableTitle()))
	}
	return errors.WithStack(err)
}

func LockBy(tx *gorm.DB, model EntityType, name string, value interface{}) (err error) {
	err = ForUpdate(tx).First(model, fmt.Sprintf("`%s`=?", name), value).Error
	if utils.RecordNotFound(err) {
		return errors.Wrapf(err, fmt.Sprintf("%s不存在", model.TableTitle()))
	}
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func AllBy(db *gorm.DB, model EntityType, name string, value interface{}) (err error) {
	err = db.Find(model, fmt.Sprintf("`%s`=?", name), value).Error
	if utils.RecordNotFound(err) {
		return errors.Wrapf(err, fmt.Sprintf("%s不存在", model.TableTitle()))
	}
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func Updates(db *gorm.DB, model interface{}, fields ...string) (err error) {
	if len(fields) == 0 {
		return
	}

	fMap, err := reflections.GetFields(model, fields...)
	if err != nil {
		return errors.WithStack(err)
	}

	err = db.Model(model).Omit(clause.Associations).Updates(fMap).Error
	return errors.WithStack(err)
}

func Update(db *gorm.DB, model EntityType, form EntityType, excludeFields ...string) (err error) {
	err = db.First(model, form.GetId()).Error
	if err != nil {
		if utils.RecordNotFound(err) {
			return errors.Wrapf(err, fmt.Sprintf("%s不存在", model.TableTitle()))
		}
		return errors.WithStack(err)
	}
	fields := NeedUpdateFields(model, form, excludeFields...)
	if len(fields) > 0 {
		err = copier.Copy(model, form)
		if err != nil {
			return errors.WithStack(err)
		}
		err = db.Model(model).Omit(clause.Associations).Select(fields).Updates(model).Error
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

// CreateTree 须放入事务中
func CreateTree(db *gorm.DB, model TreeType, parent TreeType) (err error) {
	var code string
	if parent == nil || parent.IsNull() {
		code, err = SequenceService.NextCode(db, model.TableName())
		if err != nil {
			return errors.WithStack(err)
		}
		model.SetCode(code)
		model.SetPath(code)
	} else {
		code, err = SequenceService.NextCode(db, fmt.Sprintf("%s:%s", model.TableName(), parent.GetPath()))
		if err != nil {
			return errors.WithStack(err)
		}
		model.SetPath(fmt.Sprintf("%s:%s", parent.GetPath(), code))
	}
	return errors.WithStack(db.Create(model).Error)
}

// UpdateTree 须放入事务中
func UpdateTree(db *gorm.DB, model TreeType, form TreeType) (err error) {
	err = GetById(db, model, form.GetId())
	if err != nil {
		return errors.WithStack(err)
	}

	if form.GetId() == form.GetParentId() {
		return errors.New("不可选自己的作为父节点")
	}

	parentChanged := model.GetParentId() != form.GetParentId()
	if parentChanged {
		return errors.Errorf("不可更改[上级%s]", model.TableTitle())
	}
	oldPath := model.GetPath()
	fields := NeedUpdateFields(model, form, "Code", "Path", "Tree")
	if len(fields) > 0 {
		if parentChanged {
			if form.GetParentId() != 0 {
				parent := &Tree{}
				err = db.Table(model.TableName()).
					Where("id=?", form.GetParentId()).
					Select("code", "path").
					First(parent).Error
				if utils.RecordNotFound(err) {
					return errors.Wrapf(err, model.TableTitle()+"不存在："+reflect.TypeOf(model).Name())
				}
				if err != nil {
					return errors.WithStack(err)
				}
				form.SetPath(fmt.Sprintf("%s:%s", parent.GetPath(), model.GetCode()))
			} else {
				form.SetPath(model.GetCode())
			}
			fields = append(fields, "Path")
		}
		err = copier.Copy(model, form)
		if err != nil {
			return errors.WithStack(err)
		}
		err = db.Model(model).Select(fields).Updates(model).Error
		if err != nil {
			return errors.WithStack(err)
		}

		if parentChanged {
			err := db.Exec(
				fmt.Sprintf("update %s set path=REGEXP_REPLACE(path, ?, ?) where path like ?", Quote(db, model.TableName())),
				"^"+oldPath+":", model.GetPath()+":",
				model.GetPath()+":%",
			).Error
			if err != nil {
				return errors.WithStack(err)
			}
		}
	}
	return nil
}

func DeleteTree(db *gorm.DB, model TreeType, ids []int64) (err error) {
	for _, id := range ids {
		err = db.First(model, id).Error
		if utils.RecordNotFound(err) {
			continue
		}
		if err != nil {
			return errors.WithStack(err)
		}
		model.SetId(0)
		err = db.Delete(model, "path like ?", model.GetPath()+"%").Error
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}
