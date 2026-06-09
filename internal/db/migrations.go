package db

const migration001 = `
CREATE TABLE IF NOT EXISTS accounts (
    id TEXT PRIMARY KEY,
    account_type TEXT NOT NULL,
    account_no TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    id_card_no TEXT,
    password_hash TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'active',
    balance TEXT NOT NULL DEFAULT '0',
    frozen_amount TEXT NOT NULL DEFAULT '0',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    closed_at DATETIME
);

CREATE INDEX IF NOT EXISTS idx_accounts_status ON accounts(status);
CREATE INDEX IF NOT EXISTS idx_accounts_account_no ON accounts(account_no);

CREATE TABLE IF NOT EXISTS account_balances (
    id TEXT PRIMARY KEY,
    account_id TEXT UNIQUE NOT NULL,
    balance TEXT NOT NULL DEFAULT '0',
    frozen_amount TEXT NOT NULL DEFAULT '0',
    version INTEGER NOT NULL DEFAULT 0,
    last_txn_id TEXT,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (account_id) REFERENCES accounts(id)
);
`

const migration002 = `
CREATE TABLE IF NOT EXISTS transfers (
    id TEXT PRIMARY KEY,
    biz_id TEXT UNIQUE NOT NULL,
    from_account_id TEXT NOT NULL,
    to_account_id TEXT,
    to_bank_code TEXT,
    to_account_no TEXT,
    to_account_name TEXT,
    amount TEXT NOT NULL,
    fee TEXT NOT NULL DEFAULT '0',
    remark TEXT,
    status TEXT NOT NULL DEFAULT 'pending',
    is_cross_bank INTEGER NOT NULL DEFAULT 0,
    settlement_id TEXT,
    failure_reason TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at DATETIME
);

CREATE INDEX IF NOT EXISTS idx_transfers_biz_id ON transfers(biz_id);
CREATE INDEX IF NOT EXISTS idx_transfers_from_account ON transfers(from_account_id);
CREATE INDEX IF NOT EXISTS idx_transfers_status ON transfers(status);
CREATE INDEX IF NOT EXISTS idx_transfers_created_at ON transfers(created_at);

CREATE TABLE IF NOT EXISTS ledger_entries (
    id TEXT PRIMARY KEY,
    account_id TEXT NOT NULL,
    transfer_id TEXT,
    biz_id TEXT,
    direction TEXT NOT NULL,
    type TEXT NOT NULL,
    amount TEXT NOT NULL,
    balance_after TEXT NOT NULL,
    counterparty TEXT,
    remark TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_ledger_account ON ledger_entries(account_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_ledger_transfer ON ledger_entries(transfer_id);
CREATE INDEX IF NOT EXISTS idx_ledger_biz_id ON ledger_entries(biz_id);
CREATE INDEX IF NOT EXISTS idx_ledger_created_at ON ledger_entries(created_at);
`

const migration003 = `
CREATE TABLE IF NOT EXISTS audit_logs (
    id TEXT PRIMARY KEY,
    action TEXT NOT NULL,
    account_id TEXT,
    operator_id TEXT,
    ip TEXT,
    user_agent TEXT,
    request_id TEXT,
    before TEXT,
    after TEXT,
    detail TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_audit_action ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_account ON audit_logs(account_id);
CREATE INDEX IF NOT EXISTS idx_audit_created_at ON audit_logs(created_at);
`

const migration004 = `
CREATE TABLE IF NOT EXISTS blacklist (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL,
    value TEXT NOT NULL,
    reason TEXT,
    expires_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_blacklist_type_value ON blacklist(type, value);

CREATE TABLE IF NOT EXISTS suspicious_transactions (
    id TEXT PRIMARY KEY,
    account_id TEXT NOT NULL,
    transfer_id TEXT NOT NULL,
    type TEXT NOT NULL,
    description TEXT,
    risk_score INTEGER NOT NULL DEFAULT 0,
    reported INTEGER NOT NULL DEFAULT 0,
    reported_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_suspicious_account ON suspicious_transactions(account_id);
CREATE INDEX IF NOT EXISTS idx_suspicious_created_at ON suspicious_transactions(created_at);
`

const migration005 = `
CREATE TABLE IF NOT EXISTS compliance_reports (
    id TEXT PRIMARY KEY,
    report_type TEXT NOT NULL,
    account_id TEXT,
    transfer_id TEXT,
    report_content TEXT,
    status TEXT NOT NULL DEFAULT 'pending',
    retry_count INTEGER NOT NULL DEFAULT 0,
    report_batch_no TEXT,
    reported_at DATETIME,
    failure_reason TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_compliance_status ON compliance_reports(status);
CREATE INDEX IF NOT EXISTS idx_compliance_type ON compliance_reports(report_type);
CREATE INDEX IF NOT EXISTS idx_compliance_created_at ON compliance_reports(created_at);
`

const migration006 = `
CREATE TABLE IF NOT EXISTS daily_account_reports (
    id TEXT PRIMARY KEY,
    account_id TEXT NOT NULL,
    report_date TEXT NOT NULL,
    begin_balance TEXT NOT NULL,
    end_balance TEXT NOT NULL,
    total_income TEXT NOT NULL,
    total_expense TEXT NOT NULL,
    txn_count INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_daily_report_account_date ON daily_account_reports(account_id, report_date);

CREATE TABLE IF NOT EXISTS monthly_account_reports (
    id TEXT PRIMARY KEY,
    account_id TEXT NOT NULL,
    report_month TEXT NOT NULL,
    begin_balance TEXT NOT NULL,
    end_balance TEXT NOT NULL,
    total_income TEXT NOT NULL,
    total_expense TEXT NOT NULL,
    interest TEXT NOT NULL DEFAULT '0',
    txn_count INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_monthly_report_account_month ON monthly_account_reports(account_id, report_month);

CREATE TABLE IF NOT EXISTS reconciliation_diffs (
    id TEXT PRIMARY KEY,
    account_id TEXT NOT NULL,
    check_date TEXT NOT NULL,
    balance_amount TEXT NOT NULL,
    ledger_amount TEXT NOT NULL,
    diff_amount TEXT NOT NULL,
    is_fixed INTEGER NOT NULL DEFAULT 0,
    fixed_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_recon_date ON reconciliation_diffs(check_date);
CREATE INDEX IF NOT EXISTS idx_recon_account ON reconciliation_diffs(account_id);
`

const migration007 = `
CREATE TABLE IF NOT EXISTS fixed_deposits (
    id TEXT PRIMARY KEY,
    account_id TEXT NOT NULL,
    principal TEXT NOT NULL,
    interest_type TEXT NOT NULL,
    annual_rate TEXT NOT NULL,
    interest_amount TEXT NOT NULL,
    start_date DATETIME NOT NULL,
    maturity_date DATETIME NOT NULL,
    auto_renew INTEGER NOT NULL DEFAULT 1,
    is_redeemed INTEGER NOT NULL DEFAULT 0,
    redeemed_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_fixed_account ON fixed_deposits(account_id);
CREATE INDEX IF NOT EXISTS idx_fixed_maturity ON fixed_deposits(maturity_date);

CREATE TABLE IF NOT EXISTS interest_records (
    id TEXT PRIMARY KEY,
    account_id TEXT NOT NULL,
    ledger_id TEXT NOT NULL,
    interest_type TEXT NOT NULL,
    principal TEXT NOT NULL,
    rate TEXT NOT NULL,
    days INTEGER NOT NULL,
    interest TEXT NOT NULL,
    period_start DATETIME NOT NULL,
    period_end DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_interest_account ON interest_records(account_id);
CREATE INDEX IF NOT EXISTS idx_interest_period ON interest_records(period_start, period_end);

CREATE TABLE IF NOT EXISTS interest_accruals (
    id TEXT PRIMARY KEY,
    account_id TEXT NOT NULL,
    accrual_date TEXT NOT NULL,
    principal TEXT NOT NULL,
    accrued TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_accrual_account_date ON interest_accruals(account_id, accrual_date);
`
