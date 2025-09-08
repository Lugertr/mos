package service

import (
	"context"
	"errors"

	"archive"
	"archive/pkg/repository"
)

type DocumentService struct {
	repo repository.Document
}

func NewDocumentService(repo repository.Document) *DocumentService {
	return &DocumentService{repo: repo}
}

func (s *DocumentService) CreateDocument(ctx context.Context, in archive.DocumentCreateInput) (int64, error) {
	if in.Title == "" {
		return 0, errors.New("title required")
	}
	// default privacy if empty
	if in.Privacy == "" {
		in.Privacy = archive.PrivacyPublic
	}
	return s.repo.CreateDocument(ctx, in)
}

func (s *DocumentService) SearchDocumentsByTag(ctx context.Context, filter archive.DocumentSearchFilter) ([]archive.DocumentSecure, error) {
	return s.repo.SearchDocumentsByTag(ctx, filter)
}

func (s *DocumentService) GetDocumentByID(ctx context.Context, id int64) (archive.DocumentSecure, error) {
	if id <= 0 {
		return archive.DocumentSecure{}, errors.New("invalid id")
	}
	return s.repo.GetDocumentByID(ctx, id)
}

func (s *DocumentService) UpdateDocument(ctx context.Context, id int64, in archive.DocumentUpdateInput) error {
	if id <= 0 {
		return errors.New("invalid id")
	}
	return s.repo.UpdateDocument(ctx, id, in)
}

func (s *DocumentService) DeleteDocument(ctx context.Context, id int64) error {
	if id <= 0 {
		return errors.New("invalid id")
	}
	return s.repo.DeleteDocument(ctx, id)
}

func (s *DocumentService) SetDocumentPermission(ctx context.Context, docID int64, p archive.DocumentPermission) error {
	if docID <= 0 || p.UserID <= 0 {
		return errors.New("invalid input")
	}
	p.DocumentID = docID
	return s.repo.SetDocumentPermission(ctx, docID, p)
}

func (s *DocumentService) RemoveDocumentPermission(ctx context.Context, docID int64, targetUserID int64) error {
	if docID <= 0 || targetUserID <= 0 {
		return errors.New("invalid input")
	}
	return s.repo.RemoveDocumentPermission(ctx, docID, targetUserID)
}
