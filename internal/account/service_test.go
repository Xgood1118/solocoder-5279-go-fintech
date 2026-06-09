package account

import (
	"context"
	"os"
	"testing"

	"github.com/fintech/core/internal/db"
	"github.com/fintech/core/internal/ledger"
	"github.com/fintech/core/internal/model"
	"github.com/fintech/core/pkg/money"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) (*db.DB, func()) {
	tmpFile, err := os.CreateTemp("", "fintech-test-*.db")
	require.NoError(t, err)
	tmpFile.Close()

	database, err := db.New(tmpFile.Name())
	require.NoError(t, err)

	cleanup := func() {
		database.Close()
		os.Remove(tmpFile.Name())
	}

	return database, cleanup
}

func TestCreateAccount(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	ledgerSvc := ledger.NewService(database)
	accountSvc := NewService(database, ledgerSvc, 4)

	params := CreateAccountParams{
		AccountType:    model.AccountTypePersonal,
		Name:           "张三",
		IDCardNo:       "110101199001011234",
		Password:       "password123",
		InitialBalance: money.NewFromInt(1000, ""),
	}

	ctx := context.Background()
	acc, err := accountSvc.Create(ctx, params)
	require.NoError(t, err)
	require.NotNil(t, acc)

	assert.Equal(t, "张三", acc.Name)
	assert.Equal(t, model.AccountStatusActive, acc.Status)
	assert.Equal(t, "1000.00", acc.Balance.Amount.StringFixed(2))
	assert.NotEmpty(t, acc.ID)
	assert.NotEmpty(t, acc.AccountNo)
}

func TestGetAccount(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	ledgerSvc := ledger.NewService(database)
	accountSvc := NewService(database, ledgerSvc, 4)

	ctx := context.Background()
	params := CreateAccountParams{
		AccountType: model.AccountTypePersonal,
		Name:        "李四",
		Password:    "123456",
	}

	acc, err := accountSvc.Create(ctx, params)
	require.NoError(t, err)

	found, err := accountSvc.Get(ctx, acc.ID)
	require.NoError(t, err)
	assert.Equal(t, acc.ID, found.ID)
	assert.Equal(t, "李四", found.Name)
}

func TestFreezeAndUnfreeze(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	ledgerSvc := ledger.NewService(database)
	accountSvc := NewService(database, ledgerSvc, 4)

	ctx := context.Background()
	acc, _ := accountSvc.Create(ctx, CreateAccountParams{
		AccountType: model.AccountTypePersonal,
		Name:        "王五",
		Password:    "123456",
	})

	err := accountSvc.Freeze(ctx, acc.ID)
	require.NoError(t, err)

	frozen, _ := accountSvc.Get(ctx, acc.ID)
	assert.Equal(t, model.AccountStatusFrozen, frozen.Status)

	err = accountSvc.Unfreeze(ctx, acc.ID)
	require.NoError(t, err)

	active, _ := accountSvc.Get(ctx, acc.ID)
	assert.Equal(t, model.AccountStatusActive, active.Status)
}

func TestCloseAccount(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	ledgerSvc := ledger.NewService(database)
	accountSvc := NewService(database, ledgerSvc, 4)

	ctx := context.Background()
	acc, _ := accountSvc.Create(ctx, CreateAccountParams{
		AccountType: model.AccountTypePersonal,
		Name:        "赵六",
		Password:    "123456",
	})

	err := accountSvc.Close(ctx, acc.ID)
	require.NoError(t, err)

	closed, _ := accountSvc.Get(ctx, acc.ID)
	assert.Equal(t, model.AccountStatusClosed, closed.Status)
}

func TestCloseAccountWithBalance(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	ledgerSvc := ledger.NewService(database)
	accountSvc := NewService(database, ledgerSvc, 4)

	ctx := context.Background()
	acc, _ := accountSvc.Create(ctx, CreateAccountParams{
		AccountType:    model.AccountTypePersonal,
		Name:           "钱七",
		Password:       "123456",
		InitialBalance: money.NewFromInt(100, ""),
	})

	err := accountSvc.Close(ctx, acc.ID)
	assert.Error(t, err)
}

func TestChangePassword(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	ledgerSvc := ledger.NewService(database)
	accountSvc := NewService(database, ledgerSvc, 4)

	ctx := context.Background()
	acc, _ := accountSvc.Create(ctx, CreateAccountParams{
		AccountType: model.AccountTypePersonal,
		Name:        "孙八",
		Password:    "oldpass",
	})

	err := accountSvc.ChangePassword(ctx, acc.ID, "oldpass", "newpass123")
	require.NoError(t, err)

	ok, err := accountSvc.VerifyPassword(ctx, acc.ID, "newpass123")
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = accountSvc.VerifyPassword(ctx, acc.ID, "oldpass")
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestVerifyPassword(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	ledgerSvc := ledger.NewService(database)
	accountSvc := NewService(database, ledgerSvc, 4)

	ctx := context.Background()
	acc, _ := accountSvc.Create(ctx, CreateAccountParams{
		AccountType: model.AccountTypePersonal,
		Name:        "周九",
		Password:    "secret123",
	})

	ok, err := accountSvc.VerifyPassword(ctx, acc.ID, "secret123")
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = accountSvc.VerifyPassword(ctx, acc.ID, "wrongpassword")
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestGetBalance(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	ledgerSvc := ledger.NewService(database)
	accountSvc := NewService(database, ledgerSvc, 4)

	ctx := context.Background()
	acc, _ := accountSvc.Create(ctx, CreateAccountParams{
		AccountType:    model.AccountTypePersonal,
		Name:           "吴十",
		Password:       "123456",
		InitialBalance: money.NewFromInt(5000, ""),
	})

	balance, err := accountSvc.GetBalance(ctx, acc.ID)
	require.NoError(t, err)
	assert.Equal(t, "5000.00", balance.Amount.StringFixed(2))
}
