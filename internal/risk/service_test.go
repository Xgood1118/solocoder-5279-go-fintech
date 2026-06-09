package risk

import (
	"context"
	"os"
	"testing"

	"github.com/fintech/core/internal/config"
	"github.com/fintech/core/internal/db"
	"github.com/fintech/core/internal/model"
	"github.com/fintech/core/pkg/money"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRisk(t *testing.T) (*db.DB, *Service, func()) {
	tmpFile, err := os.CreateTemp("", "fintech-test-*.db")
	require.NoError(t, err)
	tmpFile.Close()

	database, err := db.New(tmpFile.Name())
	require.NoError(t, err)

	cfg := config.Default()
	svc := NewService(database, &cfg.Risk)

	cleanup := func() {
		database.Close()
		os.Remove(tmpFile.Name())
	}

	return database, svc, cleanup
}

func TestBlacklist(t *testing.T) {
	_, svc, cleanup := setupTestRisk(t)
	defer cleanup()

	ctx := context.Background()

	err := svc.LoadBlacklist(ctx)
	require.NoError(t, err)

	assert.False(t, svc.IsBlacklisted(ctx, "ACC123"))

	err = svc.AddToBlacklist(ctx, model.BlacklistTypeAccount, "ACC123", "测试黑名单", nil)
	require.NoError(t, err)

	assert.True(t, svc.IsBlacklisted(ctx, "ACC123"))
	assert.False(t, svc.IsBlacklisted(ctx, "ACC456"))

	err = svc.RemoveFromBlacklist(ctx, model.BlacklistTypeAccount, "ACC123")
	require.NoError(t, err)

	assert.False(t, svc.IsBlacklisted(ctx, "ACC123"))
}

func TestIPBlacklist(t *testing.T) {
	_, svc, cleanup := setupTestRisk(t)
	defer cleanup()

	ctx := context.Background()

	assert.False(t, svc.IsIPBlacklisted(ctx, "192.168.1.1"))

	err := svc.AddToBlacklist(ctx, model.BlacklistTypeIP, "192.168.1.1", "可疑IP", nil)
	require.NoError(t, err)

	assert.True(t, svc.IsIPBlacklisted(ctx, "192.168.1.1"))
}

func TestCheckTransfer(t *testing.T) {
	_, svc, cleanup := setupTestRisk(t)
	defer cleanup()

	ctx := context.Background()

	params := CheckTransferParams{
		FromAccountID: "ACC001",
		ToAccountID:   "ACC002",
		Amount:        money.NewFromInt(1000, ""),
		IP:            "127.0.0.1",
	}

	err := svc.CheckTransfer(ctx, params)
	assert.NoError(t, err)
}

func TestCheckTransferBlacklisted(t *testing.T) {
	_, svc, cleanup := setupTestRisk(t)
	defer cleanup()

	ctx := context.Background()

	svc.AddToBlacklist(ctx, model.BlacklistTypeAccount, "ACC-BAD", "黑名单账户", nil)

	params := CheckTransferParams{
		FromAccountID: "ACC-BAD",
		ToAccountID:   "ACC-GOOD",
		Amount:        money.NewFromInt(100, ""),
	}

	err := svc.CheckTransfer(ctx, params)
	assert.Error(t, err)
}

func TestCheckTransferSingleLimit(t *testing.T) {
	_, svc, cleanup := setupTestRisk(t)
	defer cleanup()

	ctx := context.Background()

	params := CheckTransferParams{
		FromAccountID: "ACC001",
		ToAccountID:   "ACC002",
		Amount:        money.NewFromInt(100000, ""),
	}

	err := svc.CheckTransfer(ctx, params)
	assert.Error(t, err)
}

func TestListBlacklist(t *testing.T) {
	_, svc, cleanup := setupTestRisk(t)
	defer cleanup()

	ctx := context.Background()

	svc.AddToBlacklist(ctx, model.BlacklistTypeAccount, "ACC1", "原因1", nil)
	svc.AddToBlacklist(ctx, model.BlacklistTypeIP, "1.2.3.4", "原因2", nil)

	entries, err := svc.ListBlacklist(ctx, 10, 0)
	require.NoError(t, err)
	assert.Equal(t, 2, len(entries))
}

func TestDetectSuspicious(t *testing.T) {
	database, svc, cleanup := setupTestRisk(t)
	defer cleanup()

	ctx := context.Background()

	_, err := database.ExecContext(ctx, `
		INSERT INTO accounts (id, account_type, account_no, name, password_hash, status, balance)
		VALUES ('ACC-TEST', 'personal', '123456', '测试用户', 'hash', 'active', '10000')
	`)
	require.NoError(t, err)

	suspicious, err := svc.DetectSuspicious(ctx, "ACC-TEST", money.NewFromInt(60000, ""))
	require.NoError(t, err)
	assert.Greater(t, len(suspicious), 0)
}
