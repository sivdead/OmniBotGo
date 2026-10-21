package persistent

import (
	"testing"
	"time"

	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/sivdead/OmniBotGo/internal/usecase/port"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

const (
	filterDirection = "direction"
	filterStartTime = "start_time"
	filterEndTime   = "end_time"
)

func TestBuildQuery_MessageStatusAndDirectionEnums(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	repo := &BaseRepo{}
	status := entity.MessageStatusFailed
	dir := entity.MessageDirectionInbound

	query, err := repo.buildQuery(db.Model(&entity.Message{}), port.ListParams{
		Filters: map[string]interface{}{
			filterMessageStatus: status,
			filterDirection:     dir,
		},
	})
	require.NoError(t, err)

	sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Find(&[]entity.Message{})
	})
	require.Contains(t, sql, filterMessageStatus)
	require.Contains(t, sql, filterDirection)
}

func TestBuildQuery_StartEndTimeAreRanges(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	repo := &BaseRepo{}
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)

	query, err := repo.buildQuery(db.Model(&entity.Message{}), port.ListParams{
		Filters: map[string]interface{}{
			filterStartTime: start,
			filterEndTime:   end,
		},
	})
	require.NoError(t, err)

	sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Find(&[]entity.Message{})
	})
	require.Contains(t, sql, "created_at >=")
	require.Contains(t, sql, "created_at <=")
	require.NotContains(t, sql, "start_time =")
	require.NotContains(t, sql, "end_time =")
}

func TestBuildQuery_UnknownFilterErrors(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	repo := &BaseRepo{}

	_, err := repo.buildQuery(db.Model(&entity.Message{}), port.ListParams{
		Filters: map[string]interface{}{
			"not_a_real_column": "x",
		},
	})
	require.Error(t, err)
	require.ErrorIs(t, err, errUnknownFilter)
}

func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	require.NoError(t, err)

	return db
}
