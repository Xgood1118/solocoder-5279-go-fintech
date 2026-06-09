package audit

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/fintech/core/internal/db"
	"github.com/fintech/core/internal/model"
	"github.com/fintech/core/pkg/id"
)

type Service struct {
	db      *db.DB
	logChan chan *model.AuditLog
	wg      sync.WaitGroup
	stopCh  chan struct{}
}

type LogParams struct {
	Action     model.AuditAction
	AccountID  string
	OperatorID string
	IP         string
	UserAgent  string
	RequestID  string
	Before     interface{}
	After      interface{}
	Detail     string
}

func NewService(database *db.DB, bufferSize int) *Service {
	if bufferSize <= 0 {
		bufferSize = 1000
	}

	svc := &Service{
		db:      database,
		logChan: make(chan *model.AuditLog, bufferSize),
		stopCh:  make(chan struct{}),
	}

	go svc.worker()
	return svc
}

func (s *Service) worker() {
	s.wg.Add(1)
	defer s.wg.Done()

	for {
		select {
		case logEntry, ok := <-s.logChan:
			if !ok {
				return
			}
			s.writeLog(logEntry)
		case <-s.stopCh:
			for {
				select {
				case logEntry, ok := <-s.logChan:
					if !ok {
						return
					}
					s.writeLog(logEntry)
				default:
					return
				}
			}
		}
	}
}

func (s *Service) writeLog(entry *model.AuditLog) {
	_, err := s.db.ExecContext(context.Background(), `
		INSERT INTO audit_logs (id, action, account_id, operator_id, ip, user_agent, request_id, before, after, detail)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, entry.ID, entry.Action, entry.AccountID, entry.OperatorID, entry.IP,
		entry.UserAgent, entry.RequestID, entry.Before, entry.After, entry.Detail)
	if err != nil {
		log.Printf("Failed to write audit log: %v", err)
	}
}

func (s *Service) Log(ctx context.Context, params LogParams) {
	entry := &model.AuditLog{
		ID:         id.NewAuditID(),
		Action:     params.Action,
		AccountID:  params.AccountID,
		OperatorID: params.OperatorID,
		IP:         params.IP,
		UserAgent:  params.UserAgent,
		RequestID:  params.RequestID,
		Detail:     params.Detail,
	}

	if params.Before != nil {
		if b, err := json.Marshal(params.Before); err == nil {
			entry.Before = string(b)
		}
	}
	if params.After != nil {
		if b, err := json.Marshal(params.After); err == nil {
			entry.After = string(b)
		}
	}

	select {
	case s.logChan <- entry:
	default:
		log.Printf("Audit log channel full, dropping log: %s", entry.ID)
	}
}

func (s *Service) LogSync(ctx context.Context, params LogParams) error {
	entry := &model.AuditLog{
		ID:         id.NewAuditID(),
		Action:     params.Action,
		AccountID:  params.AccountID,
		OperatorID: params.OperatorID,
		IP:         params.IP,
		UserAgent:  params.UserAgent,
		RequestID:  params.RequestID,
		Detail:     params.Detail,
	}

	if params.Before != nil {
		if b, err := json.Marshal(params.Before); err == nil {
			entry.Before = string(b)
		}
	}
	if params.After != nil {
		if b, err := json.Marshal(params.After); err == nil {
			entry.After = string(b)
		}
	}

	s.writeLog(entry)
	return nil
}

func (s *Service) List(ctx context.Context, limit, offset int) ([]*model.AuditLog, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, action, account_id, operator_id, ip, user_agent, request_id, before, after, detail, created_at
		FROM audit_logs ORDER BY created_at DESC LIMIT ? OFFSET ?
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*model.AuditLog
	for rows.Next() {
		l, err := db.ScanAuditLog(rows)
		if err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, nil
}

func (s *Service) ListByAccount(ctx context.Context, accountID string, limit, offset int) ([]*model.AuditLog, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, action, account_id, operator_id, ip, user_agent, request_id, before, after, detail, created_at
		FROM audit_logs WHERE account_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?
	`, accountID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*model.AuditLog
	for rows.Next() {
		l, err := db.ScanAuditLog(rows)
		if err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, nil
}

func (s *Service) Stop() {
	close(s.stopCh)
	s.wg.Wait()
	close(s.logChan)
}
