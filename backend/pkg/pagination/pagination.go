package pagination

import (
	"strings"

	"gorm.io/gorm"
)

const (
	// DefaultPage 默认页码。
	DefaultPage = 1
	// DefaultPageSize 默认每页条数。
	DefaultPageSize = 20
	// MaxPageSize 单页最大条数, 防止一次拉取过多数据。
	MaxPageSize = 200
)

// Params 是列表接口通用的分页/排序入参, 各模块以结构体内嵌方式复用。
type Params struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	SortBy   string `form:"sort_by"`
	Order    string `form:"order"`
}

// SortSpec 定义排序白名单: 请求参数名 -> 数据库列名, 以此避免 SQL 注入。
type SortSpec struct {
	Allowed map[string]string
	Default string
}

// Query 是解析后的分页与排序条件。
type Query struct {
	Page       int
	PageSize   int
	SortColumn string
	Descending bool
}

// Parse 归一化分页参数并校验排序字段。
func Parse(p Params, spec SortSpec) Query {
	q := Query{Page: p.Page, PageSize: p.PageSize}
	if q.Page < 1 {
		q.Page = DefaultPage
	}
	if q.PageSize < 1 {
		q.PageSize = DefaultPageSize
	}
	if q.PageSize > MaxPageSize {
		q.PageSize = MaxPageSize
	}

	column := spec.Default
	if key := strings.ToLower(strings.TrimSpace(p.SortBy)); key != "" {
		if allowed, ok := spec.Allowed[key]; ok {
			column = allowed
		}
	}
	q.SortColumn = column
	q.Descending = !strings.EqualFold(strings.TrimSpace(p.Order), "asc")
	return q
}

// Offset 返回 SQL OFFSET。
func (q Query) Offset() int {
	return (q.Page - 1) * q.PageSize
}

// Limit 返回 SQL LIMIT。
func (q Query) Limit() int {
	return q.PageSize
}

// OrderClause 返回可安全拼接的排序片段(列名来自白名单), 并追加 id 兜底保证分页结果稳定。
func (q Query) OrderClause() string {
	column := q.SortColumn
	if column == "" {
		column = "id"
	}
	if q.Descending {
		return column + " DESC, id DESC"
	}
	return column + " ASC, id ASC"
}

// FindPage 基于"同一个条件构造函数"执行总数统计与分页查询,
// 从机制上保证任何读模型的 total 与 items 口径完全一致。
//
// scoped 接收一个基础会话(通常是 db.Model(&Entity{})), 在其上追加筛选条件;
// 统计总数与查询列表两次调用得到的条件必须相同, 排序/分页只附加在列表查询上。
// 数据库错误原样返回, 由调用方按模块语义包装文案。
func FindPage[T any](ctxSession *gorm.DB, page Query, scoped func(*gorm.DB) *gorm.DB) ([]T, int64, error) {
	var total int64
	if err := scoped(ctxSession).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	items := make([]T, 0)
	if err := scoped(ctxSession).
		Order(page.OrderClause()).
		Offset(page.Offset()).
		Limit(page.Limit()).
		Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}
