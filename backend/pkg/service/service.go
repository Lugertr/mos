package service

import (
	"archive/pkg/repository"
)

// Service агрегирует все сервисы
type Service struct {
	Authorization Authorization
	Authors       Authors
	DocumentTypes DocumentTypes
	Tags          Tags
	Document      Document
	Admin         Admin
}

func NewService(repos *repository.Repository) *Service {
	return &Service{
		Authorization: NewAuthService(repos.Authorization),
		Authors:       NewAuthorsService(repos.Authors),
		DocumentTypes: NewDocumentTypesService(repos.DocumentTypes),
		Tags:          NewTagsService(repos.Tags),
		Document:      NewDocumentService(repos.Document),
		Admin:         NewAdminService(repos.Admin),
	}
}
