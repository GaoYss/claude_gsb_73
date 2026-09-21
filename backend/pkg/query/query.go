// Package query 提供各列表读模型共用的查询条件与时间解析原语。
//
// 设计目标:
//   - 同一份台账/业务数据在任何页面拼装筛选条件时都走同一套 Condition,
//     杜绝各模块各写一份 WHERE 导致的口径漂移;
//   - 列名只允许来自模块内部的常量/白名单, Condition 不接收外部传入的裸列名;
//   - 日期区间统一为半开区间 [from, to), 结束日期自动 +1 天, 与既有行为保持一致。
package query

import (
	"strings"
	"time"

	"gorm.io/gorm"

	"streetlight/internal/apperr"
)

// Condition 是一个可追加到 gorm 查询上的不可变筛选条件。
// 空值条件必须原样返回 statement, 保证"未填写即不过滤"。
type Condition func(statement *gorm.DB) *gorm.DB

// Apply 按声明顺序拼接全部条件(条件之间为 AND)。
func Apply(statement *gorm.DB, conditions ...Condition) *gorm.DB {
	for _, condition := range conditions {
		if condition != nil {
			statement = condition(statement)
		}
	}
	return statement
}

// Keyword 对多列做同一关键词的模糊匹配(OR), 关键词去空白后为空时不过滤。
func Keyword(keyword string, columns ...string) Condition {
	return func(statement *gorm.DB) *gorm.DB {
		value := strings.TrimSpace(keyword)
		if value == "" || len(columns) == 0 {
			return statement
		}
		like := "%" + value + "%"
		clauses := make([]string, len(columns))
		args := make([]any, len(columns))
		for i, column := range columns {
			clauses[i] = column + " LIKE ?"
			args[i] = like
		}
		return statement.Where(strings.Join(clauses, " OR "), args...)
	}
}

// Eq 字符串精确匹配, 去空白后为空时不过滤。
func Eq(column, value string) Condition {
	return func(statement *gorm.DB) *gorm.DB {
		value = strings.TrimSpace(value)
		if value == "" {
			return statement
		}
		return statement.Where(column+" = ?", value)
	}
}

// EqUint 无符号整数精确匹配, 零值视为未填写不过滤。
func EqUint(column string, value uint) Condition {
	return func(statement *gorm.DB) *gorm.DB {
		if value == 0 {
			return statement
		}
		return statement.Where(column+" = ?", value)
	}
}

// In 集合匹配, 空集合不过滤。values 必须来自内部常量, 不接受外部透传。
func In(column string, values []string) Condition {
	return func(statement *gorm.DB) *gorm.DB {
		if len(values) == 0 {
			return statement
		}
		return statement.Where(column+" IN ?", values)
	}
}

// TimeRange 生成半开区间条件 [from, to), 任一端点为 nil 时该端不限制。
func TimeRange(column string, from, to *time.Time) Condition {
	return func(statement *gorm.DB) *gorm.DB {
		if from != nil {
			statement = statement.Where(column+" >= ?", *from)
		}
		if to != nil {
			statement = statement.Where(column+" < ?", *to)
		}
		return statement
	}
}

// Exists 追加 EXISTS 子查询条件, SQL 文本由调用方(内部代码)提供。
func Exists(sql string, args ...any) Condition {
	return func(statement *gorm.DB) *gorm.DB {
		return statement.Where("EXISTS ("+sql+")", args...)
	}
}

// TimeLayouts 是接口允许用户提交的全部时间格式, 顺序即尝试顺序。
var TimeLayouts = []string{
	time.RFC3339,
	"2006-01-02 15:04:05",
	"2006-01-02T15:04:05",
	"2006-01-02",
}

// ParseDay 解析 YYYY-MM-DD, 返回当天零点(本地时区)。
func ParseDay(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	date, err := time.ParseInLocation("2006-01-02", value, time.Local)
	if err != nil {
		return time.Time{}, apperr.BadRequest("日期格式应为 YYYY-MM-DD, 当前值: %s", value)
	}
	return date, nil
}

// ParseHalfOpenRange 将起止日期(YYYY-MM-DD)解析为半开区间 [from, end+1d)。
// 任一参数留空表示该端不限制; 结束早于开始时返回业务错误。
func ParseHalfOpenRange(start, end string) (from, to *time.Time, err error) {
	if value := strings.TrimSpace(start); value != "" {
		parsed, parseErr := ParseDay(value)
		if parseErr != nil {
			return nil, nil, parseErr
		}
		from = &parsed
	}
	if value := strings.TrimSpace(end); value != "" {
		parsed, parseErr := ParseDay(value)
		if parseErr != nil {
			return nil, nil, parseErr
		}
		parsed = parsed.AddDate(0, 0, 1)
		to = &parsed
	}
	if from != nil && to != nil && to.Before(*from) {
		return nil, nil, apperr.BadRequest("结束日期不能早于开始日期")
	}
	return from, to, nil
}

// ParseFlexibleTime 解析多种常见时间格式; 留空时返回 fallback。
// invalidFormat 为格式非法时的错误文案模板(用 %%s 占位原始值),
// 由调用方传入以保持各模块既有的错误提示不变。
func ParseFlexibleTime(value string, fallback time.Time, invalidFormat string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback, nil
	}
	for _, layout := range TimeLayouts {
		if parsed, err := time.ParseInLocation(layout, value, time.Local); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, apperr.BadRequest(invalidFormat, value)
}
