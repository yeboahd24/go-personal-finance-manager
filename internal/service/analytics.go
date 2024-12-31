package service

import (
	"context"
	"fmt"
	"time"

	"github.com/yeboahd24/personal-finance-manager/internal/model"
	"github.com/yeboahd24/personal-finance-manager/internal/repository"
)

type AnalyticsService struct {
	repo repository.Repository
}

func NewAnalyticsService(repo repository.Repository) *AnalyticsService {
	return &AnalyticsService{
		repo: repo,
	}
}

func (s *AnalyticsService) GetSpendingByCategory(ctx context.Context, userID string, period string) ([]model.SpendingByCategory, error) {
	start, end := getPeriodDates(period)
	filter := model.AnalyticsFilter{
		UserID:    userID,
		StartDate: start,
		EndDate:   end,
		Limit:     10,
	}

	result, err := s.repo.GetSpendingByCategory(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("analytics service - GetSpendingByCategory: %w", err)
	}
	return result, nil
}

func (s *AnalyticsService) GetMonthlySpending(ctx context.Context, userID string, months int) ([]model.MonthlySpending, error) {
	if months <= 0 {
		months = 12
	}

	end := time.Now()
	start := end.AddDate(0, -months+1, 0).UTC()
	start = time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, time.UTC)
	end = time.Date(end.Year(), end.Month()+1, 0, 23, 59, 59, 999999999, time.UTC)

	filter := model.AnalyticsFilter{
		UserID:    userID,
		StartDate: start,
		EndDate:   end,
		Limit:     months,
	}

	result, err := s.repo.GetMonthlySpending(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("analytics service - GetMonthlySpending: %w", err)
	}
	return result, nil
}

func (s *AnalyticsService) GetCashFlow(ctx context.Context, userID string, period string) (*model.CashFlow, error) {
	start, end := getPeriodDates(period)
	filter := model.AnalyticsFilter{
		UserID:    userID,
		StartDate: start,
		EndDate:   end,
	}

	result, err := s.repo.GetCashFlow(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("analytics service - GetCashFlow: %w", err)
	}
	return result, nil
}

func (s *AnalyticsService) GetTopMerchants(ctx context.Context, userID string, period string, limit int) ([]model.MerchantSpending, error) {
	if limit <= 0 {
		limit = 10
	}

	start, end := getPeriodDates(period)
	filter := model.AnalyticsFilter{
		UserID:    userID,
		StartDate: start,
		EndDate:   end,
		Limit:     limit,
	}

	result, err := s.repo.GetTopMerchants(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("analytics service - GetTopMerchants: %w", err)
	}
	return result, nil
}

func (s *AnalyticsService) GetFinancialReport(ctx context.Context, userID string, period string) (*model.FinancialReport, error) {
	start, end := getPeriodDates(period)
	filter := model.AnalyticsFilter{
		UserID:    userID,
		StartDate: start,
		EndDate:   end,
	}

	result, err := s.repo.GetFinancialReport(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("analytics service - GetFinancialReport: %w", err)
	}
	return result, nil
}

func (s *AnalyticsService) GetIncomeVsExpenses(ctx context.Context, userID string, period string) (*model.IncomeVsExpenses, error) {
	start, end := getPeriodDates(period)
	filter := model.AnalyticsFilter{
		UserID:    userID,
		StartDate: start,
		EndDate:   end,
	}
	
	result, err := s.repo.GetIncomeVsExpenses(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("analytics service - GetIncomeVsExpenses: %w", err)
	}
	return result, nil
}

func (s *AnalyticsService) GetBudgetPerformance(ctx context.Context, userID string, period string) (*model.BudgetPerformance, error) {
	start, end := getPeriodDates(period)
	filter := model.AnalyticsFilter{
		UserID:    userID,
		StartDate: start,
		EndDate:   end,
	}

	result, err := s.repo.GetBudgetPerformance(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("analytics service - GetBudgetPerformance: %w", err)
	}
	return result, nil
}

func getPeriodDates(period string) (time.Time, time.Time) {
	now := time.Now().UTC()
	end := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, time.UTC)
	var start time.Time

	switch period {
	case "week":
		start = end.AddDate(0, 0, -7)
	case "month":
		start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		end = time.Date(now.Year(), now.Month()+1, 0, 23, 59, 59, 999999999, time.UTC)
	case "quarter":
		quarter := (int(now.Month())-1)/3 + 1
		start = time.Date(now.Year(), time.Month((quarter-1)*3+1), 1, 0, 0, 0, 0, time.UTC)
		end = time.Date(now.Year(), time.Month(quarter*3+1), 0, 23, 59, 59, 999999999, time.UTC)
	case "year":
		start = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
		end = time.Date(now.Year(), 12, 31, 23, 59, 59, 999999999, time.UTC)
	case "ytd":
		start = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
	default: // last 30 days
		start = end.AddDate(0, 0, -30)
	}

	return start, end
}
