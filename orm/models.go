package orm

import (
	"github.com/vuuvv/orca/id"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
	"strings"
	"time"
)

type EntityType interface {
	GetId() int64
	SetId(value int64)
	IsNull() bool
	TableName() string
	TableTitle() string
}

type Id struct {
	Id int64 `json:"id" gorm:"autoIncrement:false"`
}

func (e *Id) BeforeCreate(_ *gorm.DB) error {
	if e.Id == 0 {
		e.Id = id.Next()
	}
	return nil
}

func (e *Id) TableName() string {
	return ""
}

func (e *Id) TableTitle() string {
	return "未指定"
}

func (e *Id) IsNull() bool {
	return e.Id == 0
}

func (e *Id) GetId() int64 {
	return e.Id
}

func (e *Id) SetId(value int64) {
	e.Id = value
}

type Entity struct {
	CreatedBy int64                 `json:"createdBy" gorm:"comment:创建用户"`
	UpdatedBy int64                 `json:"updatedBy" gorm:"comment:最后更新用户"`
	CreatedAt time.Time             `json:"createdAt" gorm:"comment:创建时间"`
	UpdatedAt time.Time             `json:"updatedAt" gorm:"comment:最后更新时间"`
	Trashed   soft_delete.DeletedAt `json:"trashed" gorm:"softDelete:flag,DeletedAtField:DeletedAt;comment:是否已删除"`
	DeletedAt soft_delete.DeletedAt `json:"deletedAt"  gorm:"comment:删除时间"`
}

type TreeType interface {
	GetId() int64
	SetId(value int64)
	IsNull() bool
	TableName() string
	TableTitle() string

	GetParentId() int64
	SetParentId(value int64)
	GetCode() string
	SetCode(value string)
	GetPath() string
	SetPath(value string)
}

type Tree struct {
	Id
	ParentId int64  `json:"parentId" gorm:"default:0;comment:父Id"`
	Code     string `json:"code" gorm:"comment:用于查询的Id，和Path一起使用"`
	Path     string `json:"path" gorm:"comment:查询路径，包括所有祖先的code，用:分割"`
}

func (e *Tree) GetParentId() int64 {
	return e.ParentId
}

func (e *Tree) SetParentId(value int64) {
	e.ParentId = value
}

func (e *Tree) GetCode() string {
	return e.Code
}

func (e *Tree) SetCode(value string) {
	e.Code = value
}

func (e *Tree) GetPath() string {
	return e.Path
}

func (e *Tree) SetPath(value string) {
	e.Path = value
}

func (e *Tree) Codes() []string {
	return strings.Split(e.Path, ":")
}

func (e *Tree) AncestorsCode() []string {
	codes := e.Codes()
	return codes[0 : len(codes)-1]
}
