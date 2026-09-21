package query

import (
	"context"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"streetlight/internal/apperr"
)

type sampleRow struct {
	ID        uint   `gorm:"primaryKey"`
	Code      string `gorm:"size:32"`
	Road      string `gorm:"size:32"`
	Status    string `gorm:"size:32"`
	CreatedAt time.Time
}

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{
		Logger:         logger.Default.LogMode(logger.Silent),
		NamingStrategy: schema.NamingStrategy{SingularTable: true},
	})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	require.NoError(t, db.AutoMigrate(&sampleRow{}))
	return db
}

func seed(t *testing.T, db *gorm.DB) {
	t.Helper()
	base := time.Date(2026, 9, 1, 0, 0, 0, 0, time.Local)
	rows := []sampleRow{
		{Code: "LD-001", Road: "中山路", Status: "normal", CreatedAt: base},
		{Code: "LD-002", Road: "中山路", Status: "fault", CreatedAt: base.Add(25 * time.Hour)},
		{Code: "LD-003", Road: "解放路", Status: "normal", CreatedAt: base.Add(48 * time.Hour)},
	}
	require.NoError(t, db.Create(&rows).Error)
}

func TestKeywordConditionIsSharedOrMatch(t *testing.T) {
	db := newTestDB(t)
	seed(t, db)
	ctx := context.Background()

	var codes []string
	err := Apply(db.WithContext(ctx).Model(&sampleRow{}),
		Keyword("LD", "code", "road"),
	).Order("id ASC").Pluck("code", &codes).Error
	require.NoError(t, err)
	require.Equal(t, []string{"LD-001", "LD-002", "LD-003"}, codes)

	// 空白关键词不过滤, 返回全部。
	codes = codes[:0]
	err = Apply(db.WithContext(ctx).Model(&sampleRow{}),
		Keyword("  ", "code"),
	).Pluck("code", &codes).Error
	require.NoError(t, err)
	require.Len(t, codes, 3)
}

func TestEqAndInConditions(t *testing.T) {
	db := newTestDB(t)
	seed(t, db)
	ctx := context.Background()

	var count int64
	require.NoError(t, Apply(db.WithContext(ctx).Model(&sampleRow{}), Eq("road", "中山路")).Count(&count).Error)
	require.Equal(t, int64(2), count)

	require.NoError(t, Apply(db.WithContext(ctx).Model(&sampleRow{}),
		Eq("road", ""), In("status", []string{"fault"})).Count(&count).Error)
	require.Equal(t, int64(1), count)
}

func TestHalfOpenRangeIncludesEndDay(t *testing.T) {
	db := newTestDB(t)
	seed(t, db)
	ctx := context.Background()

	// 起止同一天 [09-01, 09-02), 只命中 09-01 的一条。
	from, to, err := ParseHalfOpenRange("2026-09-01", "2026-09-01")
	require.NoError(t, err)
	var count int64
	require.NoError(t, Apply(db.WithContext(ctx).Model(&sampleRow{}),
		TimeRange("created_at", from, to)).Count(&count).Error)
	require.Equal(t, int64(1), count)

	// [09-01, 09-03) 覆盖到 09-02 结束, 命中前两条。
	from, to, err = ParseHalfOpenRange("2026-09-01", "2026-09-02")
	require.NoError(t, err)
	require.NoError(t, Apply(db.WithContext(ctx).Model(&sampleRow{}),
		TimeRange("created_at", from, to)).Count(&count).Error)
	require.Equal(t, int64(2), count)
}

func TestHalfOpenRangeRejectsReversedAndInvalid(t *testing.T) {
	_, _, err := ParseHalfOpenRange("2026-09-10", "2026-09-01")
	require.Error(t, err)
	businessErr, ok := apperr.As(err)
	require.True(t, ok)
	require.Equal(t, "结束日期不能早于开始日期", businessErr.Message)

	_, _, err = ParseHalfOpenRange("not-a-date", "")
	require.Error(t, err)
	_, ok = apperr.As(err)
	require.True(t, ok)
}

func TestParseFlexibleTimeFormatsAndEmpty(t *testing.T) {
	fallback := time.Date(2026, 1, 1, 8, 0, 0, 0, time.Local)

	got, err := ParseFlexibleTime("", fallback, "bad %s")
	require.NoError(t, err)
	require.Equal(t, fallback, got)

	got, err = ParseFlexibleTime("2026-09-01 10:30:00", fallback, "bad %s")
	require.NoError(t, err)
	require.Equal(t, 10, got.Hour())
	require.Equal(t, 30, got.Minute())

	_, err = ParseFlexibleTime("xxxx", fallback, "时间格式不正确, 建议使用 YYYY-MM-DD HH:mm:ss: %s")
	require.Error(t, err)
	require.Contains(t, err.Error(), "时间格式不正确")
	require.Contains(t, err.Error(), "xxxx")
}
