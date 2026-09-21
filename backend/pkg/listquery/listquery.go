// Package listquery 提供列表查询的声明式筛选子句与统一的日期/时间解析。
//
// 设计目标: 每个筛选维度(等值、关键字模糊、时间区间、枚举集合)在全系统只有一份拼装规则,
// 各模块只声明需要的条件并组合, 新增筛选维度时只需加一行子句, 不再多处各改一遍。
package listquery

import (
	"strings"
	"time"

	"gorm.io/gorm"

	"streetlight/internal/apperr"
)

// Clause 是一个筛选子句, 把条件叠加到查询语句上。
// 子句必须满足: 条件值为空白(或零值)时原样返回语句, 不产生任何 WHERE 片段。
type Clause func(*gorm.DB) *gorm.DB

// Apply 按声明顺序依次应用筛选子句。
func Apply(statement *gorm.DB, clauses ...Clause) *gorm.DB {
	for _, clause := range clauses {
		if clause != nil {
			statement = clause(statement)
		}
	}
	return statement
}

// Equal 等值匹配, 值去空白后为空时跳过。
func Equal(column, value string) Clause {
	return func(statement *gorm.DB) *gorm.DB {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return statement.Where(column+" = ?", trimmed)
		}
		return statement
	}
}

// LikeAny 关键字在任一给定列上模糊匹配, 关键字去空白后为空时跳过。
func LikeAny(keyword string, columns ...string) Clause {
	return func(statement *gorm.DB) *gorm.DB {
		trimmed := strings.TrimSpace(keyword)
		if trimmed == "" || len(columns) == 0 {
			return statement
		}
		like := "%" + trimmed + "%"
		conditions := make([]string, 0, len(columns))
		args := make([]any, 0, len(columns))
		for _, column := range columns {
			conditions = append(conditions, column+" LIKE ?")
			args = append(args, like)
		}
		return statement.Where(strings.Join(conditions, " OR "), args...)
	}
}

// InStrings 列值落在给定集合内, 集合为空时跳过。values 只允许来自模块内常量。
func InStrings(column string, values []string) Clause {
	return func(statement *gorm.DB) *gorm.DB {
		if len(values) == 0 {
			return statement
		}
		return statement.Where(column+" IN ?", values)
	}
}

// TimeGTE 时间列大于等于给定值, 值为 nil 时跳过。
func TimeGTE(column string, value *time.Time) Clause {
	return func(statement *gorm.DB) *gorm.DB {
		if value == nil {
			return statement
		}
		return statement.Where(column+" >= ?", *value)
	}
}

// TimeLT 时间列严格小于给定值, 值为 nil 时跳过。
func TimeLT(column string, value *time.Time) Clause {
	return func(statement *gorm.DB) *gorm.DB {
		if value == nil {
			return statement
		}
		return statement.Where(column+" < ?", *value)
	}
}

// TimeLayouts 是接口接受的时间格式, 按顺序尝试解析。
var TimeLayouts = []string{time.RFC3339, "2006-01-02 15:04:05", "2006-01-02T15:04:05", "2006-01-02"}

// ParseTime 按 TimeLayouts 解析时间字符串, 空白值返回 fallback。
// ok 为 false 表示格式无法识别, 由调用方决定错误提示文案。
func ParseTime(value string, fallback time.Time) (result time.Time, ok bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback, true
	}
	for _, layout := range TimeLayouts {
		if parsed, err := time.ParseInLocation(layout, value, time.Local); err == nil {
			return parsed, true
		}
	}
	return time.Time{}, false
}

// ParseDay 解析 YYYY-MM-DD 日期, 返回当天零点(本地时区)。
func ParseDay(value string) (time.Time, error) {
	date, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(value), time.Local)
	if err != nil {
		return time.Time{}, apperr.BadRequest("日期格式应为 YYYY-MM-DD, 当前值: %s", value)
	}
	return date, nil
}

// DateRange 把起止两个 YYYY-MM-DD 字符串解析为左闭右开的时间区间 [from, to):
// 结束日期自动顺延一天以覆盖当天, 任一端为空表示该端不限制。
// 这是全部列表接口日期区间筛选的唯一实现, 保证各页面区间语义一致。
func DateRange(start, end string) (from, to *time.Time, err error) {
	if strings.TrimSpace(start) != "" {
		day, parseErr := ParseDay(start)
		if parseErr != nil {
			return nil, nil, parseErr
		}
		from = &day
	}
	if strings.TrimSpace(end) != "" {
		day, parseErr := ParseDay(end)
		if parseErr != nil {
			return nil, nil, parseErr
		}
		day = day.AddDate(0, 0, 1)
		to = &day
	}
	if from != nil && to != nil && to.Before(*from) {
		return nil, nil, apperr.BadRequest("结束日期不能早于开始日期")
	}
	return from, to, nil
}
