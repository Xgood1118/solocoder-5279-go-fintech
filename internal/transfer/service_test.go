package transfer

import (
	"context"
	"os"
	"testing"

	"github.com/fintech/core/internal/account"
	"github.com/fintech/core/internal/config"
	"github.com/fintech/core/internal/db"
	"github.com/fintech/core/internal/ledger"
	"github.com/fintech/core/internal/model"
	"github.com/fintech/core/internal/risk"
	"github.com/fintech/core/pkg/money"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestServices(t *testing.T) (*db.DB, *Service, *account.Service, func()) {
	tmpFile, err := os.CreateTemp("", "fintech-test-*.db")
	require.NoError(t, err)
	tmpFile.Close()

	database, err := db.New(tmpFile.Name())
	require.NoError(t, err)

	cfg := config.Default()
	ledgerSvc := ledger.NewService(database)
	accountSvc := account.NewService(database, ledgerSvc, 4)
	riskSvc := risk.NewService(database, &cfg.Risk)
	transferSvc := NewService(database, accountSvc, ledgerSvc, riskSvc)

	cleanup := func() {
		database.Close()
		os.Remove(tmpFile.Name())
	}

	return database, transferSvc, accountSvc, cleanup
}

func createTestAccount(t *testing.T, accountSvc *account.Service, name string, balance int64) string {
	ctx := context.Background()
	acc, err := accountSvc.Create(ctx, account.CreateAccountParams{
		AccountType:    model.AccountTypePersonal,
		Name:           name,
		Password:       "123456",
		InitialBalance: money.NewFromInt(balance, ""),
	})
	require.NoError(t, err)
	return acc.ID
}

func TestIntraBankTransfer(t *testing.T) {
	_, transferSvc, accountSvc, cleanup := setupTestServices(t)
	defer cleanup()

	ctx := context.Background()
	fromID := createTestAccount(t, accountSvc, "付款人", 1000)
	toID := createTestAccount(t, accountSvc, "收款人", 500)

	params := IntraBankTransferParams{
		BizID:         "test-biz-001",
		FromAccountID: fromID,
		ToAccountID:   toID,
		Amount:        money.NewFromInt(300, ""),
		Remark:        "测试转账",
		Password:      "123456",
		IP:            "127.0.0.1",
	}

	transfer, err := transferSvc.IntraBankTransfer(ctx, params)
	require.NoError(t, err)
	require.NotNil(t, transfer)

	assert.Equal(t, model.TransferStatusSuccess, transfer.Status)
	assert.Equal(t, "300.00", transfer.Amount.Amount.StringFixed(2))
	assert.Equal(t, fromID, transfer.FromAccountID)
	assert.Equal(t, toID, transfer.ToAccountID)

	fromBal, _ := accountSvc.GetBalance(ctx, fromID)
	toBal, _ := accountSvc.GetBalance(ctx, toID)
	assert.Equal(t, "700.00", fromBal.Amount.StringFixed(2))
	assert.Equal(t, "800.00", toBal.Amount.StringFixed(2))
}

func TestIntraBankTransferInsufficientBalance(t *testing.T) {
	_, transferSvc, accountSvc, cleanup := setupTestServices(t)
	defer cleanup()

	ctx := context.Background()
	fromID := createTestAccount(t, accountSvc, "付款人", 100)
	toID := createTestAccount(t, accountSvc, "收款人", 500)

	params := IntraBankTransferParams{
		BizID:         "test-biz-002",
		FromAccountID: fromID,
		ToAccountID:   toID,
		Amount:        money.NewFromInt(200, ""),
		Password:      "123456",
	}

	_, err := transferSvc.IntraBankTransfer(ctx, params)
	assert.Error(t, err)
}

func TestIntraBankTransferSameAccount(t *testing.T) {
	_, transferSvc, accountSvc, cleanup := setupTestServices(t)
	defer cleanup()

	ctx := context.Background()
	id := createTestAccount(t, accountSvc, "测试用户", 1000)

	params := IntraBankTransferParams{
		BizID:         "test-biz-003",
		FromAccountID: id,
		ToAccountID:   id,
		Amount:        money.NewFromInt(100, ""),
		Password:      "123456",
	}

	_, err := transferSvc.IntraBankTransfer(ctx, params)
	assert.Error(t, err)
}

func TestIntraBankTransferIdempotent(t *testing.T) {
	_, transferSvc, accountSvc, cleanup := setupTestServices(t)
	defer cleanup()

	ctx := context.Background()
	fromID := createTestAccount(t, accountSvc, "付款人", 1000)
	toID := createTestAccount(t, accountSvc, "收款人", 500)

	params := IntraBankTransferParams{
		BizID:         "test-biz-idempotent",
		FromAccountID: fromID,
		ToAccountID:   toID,
		Amount:        money.NewFromInt(100, ""),
		Password:      "123456",
	}

	t1, err := transferSvc.IntraBankTransfer(ctx, params)
	require.NoError(t, err)

	t2, err := transferSvc.IntraBankTransfer(ctx, params)
	require.NoError(t, err)

	assert.Equal(t, t1.ID, t2.ID)

	fromBal, _ := accountSvc.GetBalance(ctx, fromID)
	assert.Equal(t, "900.00", fromBal.Amount.StringFixed(2))
}

func TestIntraBankTransferInvalidPassword(t *testing.T) {
	_, transferSvc, accountSvc, cleanup := setupTestServices(t)
	defer cleanup()

	ctx := context.Background()
	fromID := createTestAccount(t, accountSvc, "付款人", 1000)
	toID := createTestAccount(t, accountSvc, "收款人", 500)

	params := IntraBankTransferParams{
		BizID:         "test-biz-pwd",
		FromAccountID: fromID,
		ToAccountID:   toID,
		Amount:        money.NewFromInt(100, ""),
		Password:      "wrongpassword",
	}

	_, err := transferSvc.IntraBankTransfer(ctx, params)
	assert.Error(t, err)
}

func TestIntraBankTransferZeroAmount(t *testing.T) {
	_, transferSvc, accountSvc, cleanup := setupTestServices(t)
	defer cleanup()

	ctx := context.Background()
	fromID := createTestAccount(t, accountSvc, "付款人", 1000)
	toID := createTestAccount(t, accountSvc, "收款人", 500)

	params := IntraBankTransferParams{
		BizID:         "test-biz-zero",
		FromAccountID: fromID,
		ToAccountID:   toID,
		Amount:        money.Zero(),
		Password:      "123456",
	}

	_, err := transferSvc.IntraBankTransfer(ctx, params)
	assert.Error(t, err)
}

func TestGetTransfer(t *testing.T) {
	_, transferSvc, accountSvc, cleanup := setupTestServices(t)
	defer cleanup()

	ctx := context.Background()
	fromID := createTestAccount(t, accountSvc, "付款人", 1000)
	toID := createTestAccount(t, accountSvc, "收款人", 500)

	params := IntraBankTransferParams{
		BizID:         "test-biz-get",
		FromAccountID: fromID,
		ToAccountID:   toID,
		Amount:        money.NewFromInt(100, ""),
		Password:      "123456",
	}

	created, _ := transferSvc.IntraBankTransfer(ctx, params)

	found, err := transferSvc.Get(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)
	assert.Equal(t, model.TransferStatusSuccess, found.Status)
}

func TestGetByBizID(t *testing.T) {
	_, transferSvc, accountSvc, cleanup := setupTestServices(t)
	defer cleanup()

	ctx := context.Background()
	fromID := createTestAccount(t, accountSvc, "付款人", 1000)
	toID := createTestAccount(t, accountSvc, "收款人", 500)

	bizID := "test-biz-unique"
	params := IntraBankTransferParams{
		BizID:         bizID,
		FromAccountID: fromID,
		ToAccountID:   toID,
		Amount:        money.NewFromInt(100, ""),
		Password:      "123456",
	}

	transferSvc.IntraBankTransfer(ctx, params)

	found, err := transferSvc.GetByBizID(ctx, bizID)
	require.NoError(t, err)
	assert.Equal(t, bizID, found.BizID)
}

func TestListByAccount(t *testing.T) {
	_, transferSvc, accountSvc, cleanup := setupTestServices(t)
	defer cleanup()

	ctx := context.Background()
	fromID := createTestAccount(t, accountSvc, "付款人", 1000)
	toID := createTestAccount(t, accountSvc, "收款人", 500)

	for i := 0; i < 3; i++ {
		params := IntraBankTransferParams{
			BizID:         "test-biz-list-" + string(rune('0'+i)),
			FromAccountID: fromID,
			ToAccountID:   toID,
			Amount:        money.NewFromInt(50, ""),
			Password:      "123456",
		}
		transferSvc.IntraBankTransfer(ctx, params)
	}

	transfers, err := transferSvc.ListByAccount(ctx, fromID, 10, 0)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(transfers), 3)
}
