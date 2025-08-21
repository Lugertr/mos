package service

import (
	"archive"
	"archive/pkg/repository"
	"context"
	"time"
)

type AdminService struct {
	repo repository.Admin
}

func NewAdminService(repo repository.Admin) *AdminService {
	return &AdminService{repo: repo}
}

func (s *AdminService) GetLogsByUser(ctx context.Context, adminID int64, targetUserID int64, start *time.Time, end *time.Time) ([]archive.LogRecord, error) {
	return s.repo.GetLogsByUser(ctx, adminID, targetUserID, start, end)
}

func (s *AdminService) GetLogsByTable(ctx context.Context, adminID int64, tableName string, start *time.Time, end *time.Time) ([]archive.LogRecord, error) {
	return s.repo.GetLogsByTable(ctx, adminID, tableName, start, end)
}

func (s *AdminService) GetLogsByDate(ctx context.Context, adminID int64, start time.Time, end time.Time) ([]archive.LogRecord, error) {
	return s.repo.GetLogsByDate(ctx, adminID, start, end)
}
