package orm

import (
	"fmt"
	"github.com/vuuvv/errors"
	"github.com/vuuvv/orca/utils"
	"gorm.io/gorm"
	"strings"
)

type Criteria struct {
	table string
	field string
	op    string
}

func (c *Criteria) String(db *gorm.DB) string {
	if c.table == "" {
		return fmt.Sprintf("%s %s ?", Quote(db, c.field), c.op)
	}
	operator := c.op
	switch operator {
	case OP_Contain:
		operator = OP_Like
	case OP_StartsWith:
		operator = OP_Like
	case OP_EndsWith:
		operator = OP_Like
	}
	return fmt.Sprintf("%s.%s %s ?", Quote(db, c.table), Quote(db, c.field), operator)
}

type SortBy struct {
	table string
	field string
	asc   bool
}

type Paginator struct {
	Page      int                    `json:"page"`
	PageSize  int                    `json:"pageSize"`
	Filters   map[string]interface{} `json:"filters"`
	Sort      []SortBy               `json:"sort"`
	NoCount   bool                   `json:"noCount"`
	UseOffset bool                   `json:"useOffset"`
	Offset    int                    `json:"offset"`
}

func (p *Paginator) HasFilter(key string) (ok bool) {
	if p.Filters == nil {
		return false
	}
	_, ok = p.Filters[key]
	return
}

func (p *Paginator) Filter(key string, value interface{}) {
	if p.Filters == nil {
		p.Filters = make(map[string]interface{})
	}
	p.Filters[key] = value
}

type Page struct {
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
	Total    int         `json:"total"`
	Items    interface{} `json:"items"`
}

type OrderBy struct {
	Table string `json:"table"`
	Field string `json:"field"`
	ASC   bool   `json:"asc"`
}

func (this *OrderBy) Sql(db *gorm.DB) string {
	asc := "ASC"
	if !this.ASC {
		asc = "DESC"
	}
	if this.Table != "" {
		return fmt.Sprintf("%s.%s %s", Quote(db, this.Table), Quote(db, this.Field), asc)
	}
	return fmt.Sprintf("%s %s", this.Field, asc)
}

type PageExecutor struct {
	countSql string
	sql      string
	shareSql string
	orderBy  []*OrderBy
	criteria map[string]*Criteria
}

func NewPage() *PageExecutor {
	return &PageExecutor{}
}

func (p *PageExecutor) Count(sql string) *PageExecutor {
	p.countSql = sql
	return p
}

func (p *PageExecutor) Select(sql string) *PageExecutor {
	p.sql = sql
	return p
}

func (p *PageExecutor) Criteria(name string, table string, field string, op string) *PageExecutor {
	if p.criteria == nil {
		p.criteria = map[string]*Criteria{}
	}
	p.criteria[name] = &Criteria{
		table: table,
		field: field,
		op:    op,
	}
	return p
}

func (p *PageExecutor) Join(sql string) *PageExecutor {
	p.shareSql = sql
	return p
}

func (p *PageExecutor) OrderBy(table string, field string, asc bool) *PageExecutor {
	p.orderBy = append(p.orderBy, &OrderBy{
		Table: table,
		Field: field,
		ASC:   asc,
	})
	return p
}

func (p *PageExecutor) Query(db *gorm.DB, page *Paginator, items interface{}) (*Page, error) {
	vars := map[string]string{}

	for k := range page.Filters {
		if c, ok := p.criteria[k]; ok {
			vars[k] = c.String(db)
		}
	}
	ret := &Page{
		Page:     page.Page,
		PageSize: page.PageSize,
	}

	if ret.Page < 1 {
		ret.Page = 1
	}

	if ret.PageSize == 0 {
		ret.PageSize = 20
	}

	// 获取数量
	if !page.NoCount || !page.UseOffset {
		countSql, values := prepare(utils.LineJoin(p.countSql, p.shareSql), vars, page.Filters, p.criteria)
		rows, err := db.Raw(countSql, values...).Rows()
		if err != nil {
			return ret, err
		}
		ret.Total = 0
		for rows.Next() {
			var c int
			err := db.ScanRows(rows, &c)
			if err != nil {
				return nil, err
			}
			ret.Total += c
		}
	}

	// 获取值
	sql, values := prepare(utils.LineJoin(p.sql, p.shareSql), vars, page.Filters, p.criteria)
	var orderBySql []string
	for _, o := range p.orderBy {
		orderBySql = append(orderBySql, o.Sql(db))
	}
	orderBy := strings.Join(orderBySql, ",")
	if orderBy != "" {
		sql = utils.LineJoin(sql, fmt.Sprintf("ORDER BY %s", orderBy))
	}
	// limit
	if page.UseOffset {
		sql = utils.LineJoin(sql, fmt.Sprintf("LIMIT %d offset %d", ret.PageSize, page.Offset))
	} else {
		sql = utils.LineJoin(sql, fmt.Sprintf("LIMIT %d offset %d", ret.PageSize, ret.PageSize*(ret.Page-1)))
	}
	var err error
	if len(values) > 0 {
		err = errors.WithStack(db.Raw(sql, values...).Scan(items).Error)
	} else {
		err = errors.WithStack(db.Raw(sql).Scan(items).Error)
	}

	if err != nil {
		return nil, err
	}

	ret.Items = items

	return ret, nil
}
