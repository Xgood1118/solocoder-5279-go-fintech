package config

import (
	"time"

	"github.com/shopspring/decimal"
)

type Config struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	Redis    RedisConfig    `json:"redis"`
	Security SecurityConfig `json:"security"`
	Risk     RiskConfig     `json:"risk"`
	Interest InterestConfig `json:"interest"`
	Report   ReportConfig   `json:"report"`
}

type ServerConfig struct {
	RESTPort int    `json:"rest_port"`
	GRPCPort int    `json:"grpc_port"`
	Mode     string `json:"mode"`
}

type DatabaseConfig struct {
	Path string `json:"path"`
}

type RedisConfig struct {
	Addr     string `json:"addr"`
	Password string `json:"password"`
	DB       int    `json:"db"`
	Enabled  bool   `json:"enabled"`
}

type SecurityConfig struct {
	PasswordHashCost int `json:"password_hash_cost"`
}

type RiskConfig struct {
	SingleLimit   decimal.Decimal `json:"single_limit"`
	DailyLimit    decimal.Decimal `json:"daily_limit"`
	MonthlyLimit  decimal.Decimal `json:"monthly_limit"`
	LargeAmount   decimal.Decimal `json:"large_amount"`
	SuspiciousCount int           `json:"suspicious_count"`
	SuspiciousWindow time.Duration `json:"suspicious_window"`
}

type InterestConfig struct {
	DemandRate    decimal.Decimal `json:"demand_rate"`
	Fixed3MRate   decimal.Decimal `json:"fixed_3m_rate"`
	Fixed6MRate   decimal.Decimal `json:"fixed_6m_rate"`
	Fixed1YRate   decimal.Decimal `json:"fixed_1y_rate"`
	SettleDay     int             `json:"settle_day"`
}

type ReportConfig struct {
	DailyCron   string `json:"daily_cron"`
	MonthlyCron string `json:"monthly_cron"`
}

func Default() *Config {
	return &Config{
		Server: ServerConfig{
			RESTPort: 8080,
			GRPCPort: 9090,
			Mode:     "dev",
		},
		Database: DatabaseConfig{
			Path: "./data/fintech.db",
		},
		Redis: RedisConfig{
			Addr:    "127.0.0.1:6379",
			Enabled: false,
		},
		Security: SecurityConfig{
			PasswordHashCost: 12,
		},
		Risk: RiskConfig{
			SingleLimit:       decimal.NewFromInt(50000),
			DailyLimit:        decimal.NewFromInt(500000),
			MonthlyLimit:      decimal.NewFromInt(5000000),
			LargeAmount:       decimal.NewFromInt(50000),
			SuspiciousCount:   5,
			SuspiciousWindow:  1 * time.Hour,
		},
		Interest: InterestConfig{
			DemandRate:  decimal.NewFromFloat(0.0035),
			Fixed3MRate: decimal.NewFromFloat(0.0115),
			Fixed6MRate: decimal.NewFromFloat(0.0135),
			Fixed1YRate: decimal.NewFromFloat(0.0155),
			SettleDay:   20,
		},
		Report: ReportConfig{
			DailyCron:   "0 0 0 * * *",
			MonthlyCron: "0 0 0 1 * *",
		},
	}
}
