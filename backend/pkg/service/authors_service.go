package service

import (
	"archive"
	"archive/pkg/repository"
	"context"
	"errors"
	"strings"

	"github.com/go-playground/validator/v10"
)

// AuthorsService реализует Authors interface.
// Конструктор принимает repository.Authors (sub-repo).
type AuthorsService struct {
	repo repository.Authors
	v    *validator.Validate
}

func NewAuthorsService(repo repository.Authors) *AuthorsService {
	return &AuthorsService{
		repo: repo,
		v:    validator.New(),
	}
}

func (s *AuthorsService) CreateAuthor(ctx context.Context, in archive.AuthorCreate) (int64, error) {
	if err := s.v.Struct(in); err != nil {
		return 0, err
	}
	a := archive.Author{
		FullName: strings.TrimSpace(in.FullName),
	}
	return s.repo.CreateAuthor(ctx, a)
}

func (s *AuthorsService) GetAllAuthors(ctx context.Context) ([]archive.Author, error) {
	return s.repo.GetAllAuthors(ctx)
}

func (s *AuthorsService) GetAuthor(ctx context.Context, id int64) (archive.Author, error) {
	if id <= 0 {
		return archive.Author{}, errors.New("invalid id")
	}
	return s.repo.GetAuthor(ctx, id)
}

func (s *AuthorsService) UpdateAuthor(ctx context.Context, id int64, in archive.Author) error {
	if id <= 0 {
		return errors.New("invalid id")
	}
	if err := s.v.Struct(in); err != nil {
		return err
	}
	a := archive.Author{
		FullName: strings.TrimSpace(in.FullName),
	}
	return s.repo.UpdateAuthor(ctx, id, a)
}

func (s *AuthorsService) DeleteAuthor(ctx context.Context, id int64) error {
	if id <= 0 {
		return errors.New("invalid id")
	}
	return s.repo.DeleteAuthor(ctx, id)
}
