package service

import (
	"context"
	"errors"
	"strings"

	"archive"
	"archive/pkg/repository"

	"github.com/go-playground/validator/v10"
)

type TagsService struct {
	repo repository.Tags
	v    *validator.Validate
}

func NewTagsService(repo repository.Tags) *TagsService {
	return &TagsService{
		repo: repo,
		v:    validator.New(),
	}
}

func (s *TagsService) CreateTag(ctx context.Context, in archive.TagCreate) (int64, error) {
	if err := s.v.Struct(in); err != nil {
		return 0, err
	}
	t := archive.Tag{
		Name: strings.TrimSpace(in.Name),
	}
	return s.repo.CreateTag(ctx, t)
}

func (s *TagsService) GetAllTags(ctx context.Context) ([]archive.Tag, error) {
	return s.repo.GetAllTags(ctx)
}

func (s *TagsService) GetTag(ctx context.Context, id int64) (archive.Tag, error) {
	if id <= 0 {
		return archive.Tag{}, errors.New("invalid id")
	}
	return s.repo.GetTag(ctx, id)
}

func (s *TagsService) UpdateTag(ctx context.Context, id int64, in archive.Tag) error {
	if id <= 0 {
		return errors.New("invalid id")
	}
	if err := s.v.Struct(in); err != nil {
		return err
	}
	t := archive.Tag{
		Name: strings.TrimSpace(in.Name),
	}
	return s.repo.UpdateTag(ctx, id, t)
}

func (s *TagsService) DeleteTag(ctx context.Context, id int64) error {
	if id <= 0 {
		return errors.New("invalid id")
	}
	return s.repo.DeleteTag(ctx, id)
}
