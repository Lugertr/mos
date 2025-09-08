package service

import (
	"context"
	"errors"
	"strings"

	"archive"
	"archive/pkg/repository"

	"github.com/go-playground/validator/v10"
)

type DocumentTypesService struct {
	repo repository.DocumentTypes
	v    *validator.Validate
}

func NewDocumentTypesService(repo repository.DocumentTypes) *DocumentTypesService {
	return &DocumentTypesService{
		repo: repo,
		v:    validator.New(),
	}
}

func (s *DocumentTypesService) CreateDocumentType(ctx context.Context, in archive.DocumentTypeCreate) (int64, error) {
	if err := s.v.Struct(in); err != nil {
		return 0, err
	}
	t := archive.DocumentType{
		Name: strings.TrimSpace(in.Name),
	}
	return s.repo.CreateDocumentType(ctx, t)
}

func (s *DocumentTypesService) GetAllDocumentTypes(ctx context.Context) ([]archive.DocumentType, error) {
	return s.repo.GetAllDocumentTypes(ctx)
}

func (s *DocumentTypesService) GetDocumentType(ctx context.Context, id int64) (archive.DocumentType, error) {
	if id <= 0 {
		return archive.DocumentType{}, errors.New("invalid id")
	}
	return s.repo.GetDocumentType(ctx, id)
}

func (s *DocumentTypesService) UpdateDocumentType(ctx context.Context, id int64, in archive.DocumentType) error {
	if id <= 0 {
		return errors.New("invalid id")
	}
	if err := s.v.Struct(in); err != nil {
		return err
	}
	t := archive.DocumentType{
		Name: strings.TrimSpace(in.Name),
	}
	return s.repo.UpdateDocumentType(ctx, id, t)
}

func (s *DocumentTypesService) DeleteDocumentType(ctx context.Context, id int64) error {
	if id <= 0 {
		return errors.New("invalid id")
	}
	return s.repo.DeleteDocumentType(ctx, id)
}
