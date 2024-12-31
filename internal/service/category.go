package service

import (
	"context"

	"github.com/yeboahd24/personal-finance-manager/internal/errors"
	"github.com/yeboahd24/personal-finance-manager/internal/model"
	"github.com/yeboahd24/personal-finance-manager/internal/repository"
)

type CategoryService struct {
	repo repository.Repository
}

func NewCategoryService(repo repository.Repository) *CategoryService {
	return &CategoryService{
		repo: repo,
	}
}

func (s *CategoryService) CreateCategory(ctx context.Context, category *model.Category) error {
	if category.Name == "" {
		return errors.New("Category name is required", 400)
	}

	if category.Type == "" {
		return errors.New("Category type is required", 400)
	}

	if category.Type != model.CategoryTypeIncome &&
		category.Type != model.CategoryTypeExpense &&
		category.Type != model.CategoryTypeTransfer {
		return errors.New("Invalid category type", 400)
	}

	if category.ParentID != nil {
		// Verify parent category exists
		parent, err := s.repo.GetCategoryByID(ctx, *category.ParentID)
		if err != nil {
			if err == errors.ErrNotFound {
				return errors.New("Parent category not found", 400)
			}
			return err
		}

		// Verify parent and child types match
		if parent.Type != category.Type {
			return errors.New("Child category must have same type as parent", 400)
		}
	}

	return s.repo.CreateCategory(ctx, category)
}

func (s *CategoryService) GetCategoryByID(ctx context.Context, id string) (*model.Category, error) {
	if id == "" {
		return nil, errors.New("Category ID is required", 400)
	}

	return s.repo.GetCategoryByID(ctx, id)
}

func (s *CategoryService) GetCategories(ctx context.Context, userID string) ([]*model.Category, error) {
	return s.repo.GetCategories(ctx, userID)
}

func (s *CategoryService) UpdateCategory(ctx context.Context, category *model.Category) error {
	if category.ID == "" {
		return errors.New("Category ID is required", 400)
	}

	if category.Name == "" {
		return errors.New("Category name is required", 400)
	}

	// Get existing category
	existing, err := s.repo.GetCategoryByID(ctx, category.ID)
	if err != nil {
		return err
	}

	// Don't allow changing type or parent_id as it could break budget calculations
	category.Type = existing.Type
	category.ParentID = existing.ParentID

	return s.repo.UpdateCategory(ctx, category)
}

func (s *CategoryService) DeleteCategory(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("Category ID is required", 400)
	}

	return s.repo.DeleteCategory(ctx, id)
}

func (s *CategoryService) InitializeDefaultCategories(ctx context.Context, userID string) error {
	return s.repo.InitializeDefaultCategories(ctx, userID)
}
