package service

import (
	"context"

	"github.com/yeboahd24/personal-finance-manager/internal/errors"
	"github.com/yeboahd24/personal-finance-manager/internal/model"
	"github.com/yeboahd24/personal-finance-manager/internal/repository"
)

type BudgetService struct {
	repo repository.Repository
}

func NewBudgetService(repo repository.Repository) *BudgetService {
	return &BudgetService{
		repo: repo,
	}
}

func (s *BudgetService) CreateBudget(ctx context.Context, userID string, budget *model.Budget) error {
	if budget.Amount <= 0 {
		return errors.New("Budget amount must be greater than 0", 400)
	}

	if budget.CategoryID == "" {
		return errors.New("Category ID is required", 400)
	}

	if budget.PeriodStart.IsZero() || budget.PeriodEnd.IsZero() {
		return errors.New("Budget period start and end dates are required", 400)
	}

	if budget.PeriodEnd.Before(budget.PeriodStart) {
		return errors.New("Budget period end date must be after start date", 400)
	}

	// Verify category exists and is an expense category
	category, err := s.repo.GetCategoryByID(ctx, budget.CategoryID)
	if err != nil {
		if err == errors.ErrNotFound {
			return errors.New("Category not found", 400)
		}
		return err
	}

	if category.Type != model.CategoryTypeExpense {
		return errors.New("Budgets can only be created for expense categories", 400)
	}

	// Check for overlapping budgets for the same category
	existingBudgets, err := s.repo.GetBudgets(ctx, userID, model.BudgetFilter{
		UserID:      userID,
		CategoryID:  budget.CategoryID,
		PeriodStart: budget.PeriodStart,
		PeriodEnd:   budget.PeriodEnd,
	})
	if err != nil {
		return err
	}

	for _, existing := range existingBudgets {
		if (budget.PeriodStart.Before(existing.PeriodEnd) || budget.PeriodStart.Equal(existing.PeriodEnd)) &&
			(budget.PeriodEnd.After(existing.PeriodStart) || budget.PeriodEnd.Equal(existing.PeriodStart)) {
			return errors.New("Budget period overlaps with existing budget", 400)
		}
	}

	budget.UserID = userID
	return s.repo.CreateBudget(ctx, budget)
}

func (s *BudgetService) GetBudgetByID(ctx context.Context, userID, id string) (*model.Budget, error) {
	budget, err := s.repo.GetBudgetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if budget.UserID != userID {
		return nil, errors.ErrNotFound
	}

	return budget, nil
}

func (s *BudgetService) GetBudgets(ctx context.Context, userID string, filter model.BudgetFilter) ([]*model.Budget, error) {
	filter.UserID = userID
	return s.repo.GetBudgets(ctx, userID, filter)
}

func (s *BudgetService) UpdateBudget(ctx context.Context, userID string, budget *model.Budget) error {
	existing, err := s.repo.GetBudgetByID(ctx, budget.ID)
	if err != nil {
		return err
	}

	if existing.UserID != userID {
		return errors.ErrNotFound
	}

	if budget.Amount <= 0 {
		return errors.New("Budget amount must be greater than 0", 400)
	}

	if budget.PeriodStart.IsZero() || budget.PeriodEnd.IsZero() {
		return errors.New("Budget period start and end dates are required", 400)
	}

	if budget.PeriodEnd.Before(budget.PeriodStart) {
		return errors.New("Budget period end date must be after start date", 400)
	}

	// Check for overlapping budgets
	existingBudgets, err := s.repo.GetBudgets(ctx, userID, model.BudgetFilter{
		UserID:      userID,
		CategoryID:  existing.CategoryID,
		PeriodStart: budget.PeriodStart,
		PeriodEnd:   budget.PeriodEnd,
	})
	if err != nil {
		return err
	}

	for _, other := range existingBudgets {
		if other.ID != budget.ID &&
			(budget.PeriodStart.Before(other.PeriodEnd) || budget.PeriodStart.Equal(other.PeriodEnd)) &&
			(budget.PeriodEnd.After(other.PeriodStart) || budget.PeriodEnd.Equal(other.PeriodStart)) {
			return errors.New("Budget period overlaps with existing budget", 400)
		}
	}

	// Keep original user_id and category_id
	budget.UserID = existing.UserID
	budget.CategoryID = existing.CategoryID

	return s.repo.UpdateBudget(ctx, budget)
}

func (s *BudgetService) DeleteBudget(ctx context.Context, userID, id string) error {
	existing, err := s.repo.GetBudgetByID(ctx, id)
	if err != nil {
		return err
	}

	if existing.UserID != userID {
		return errors.ErrNotFound
	}

	return s.repo.DeleteBudget(ctx, id)
}

func (s *BudgetService) GetBudgetSummary(ctx context.Context, userID string, period string) (*model.BudgetSummary, error) {
	switch period {
	case "month":
	case "year":
	default:
		return nil, errors.New("Invalid period. Must be 'month' or 'year'", 400)
	}

	return s.repo.GetBudgetSummary(ctx, userID, period)
}
