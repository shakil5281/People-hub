package repository

import "gorm.io/gorm"

type Repository struct {
	Group *GroupRepository
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{
		Group: NewGroupRepository(db),
	}
}
