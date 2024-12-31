package service

import (
	"context"
	"github.com/yeboahd24/personal-finance-manager/internal/model"
	"github.com/yeboahd24/personal-finance-manager/internal/repository"
	"time"
)

// MetricsService handles metrics-related operations
type MetricsService struct {
	repo repository.Repository
}

// NewMetricsService creates a new metrics service instance
func NewMetricsService(repo repository.Repository) *MetricsService {
	return &MetricsService{
		repo: repo,
	}
}

// GetSystemMetrics retrieves various system-wide metrics
func (s *MetricsService) GetSystemMetrics(ctx context.Context) (map[string]interface{}, error) {
	metrics := make(map[string]interface{})

	userCount, err := s.repo.GetUserCount(ctx)
	if err != nil {
		return nil, err
	}
	metrics["total_users"] = userCount

	return metrics, nil
}

// GetNetWorth calculates the total net worth for a user
func (s *MetricsService) GetNetWorth(ctx context.Context, userID string) (float64, error) {
	// Get total assets for user
	assets, err := s.repo.(repository.AccountRepository).GetTotalAssetsByUser(ctx, userID)
	if err != nil {
		return 0, err
	}

	// Get total liabilities for user
	liabilities, err := s.repo.(repository.AccountRepository).GetTotalLiabilitiesByUser(ctx, userID)
	if err != nil {
		return 0, err
	}

	return assets - liabilities, nil
}

// GetSavingsRate calculates the savings rate as a percentage for a user
func (s *MetricsService) GetSavingsRate(ctx context.Context, userID string) (float64, error) {
	filter := model.AnalyticsFilter{
		UserID:    userID,
		StartDate: time.Now().AddDate(0, -1, 0), // Last month
		EndDate:   time.Now(),
	}
	
	incomeVsExpenses, err := s.repo.(repository.AnalyticsRepository).GetIncomeVsExpenses(ctx, filter)
	if err != nil {
		return 0, err
	}

	if incomeVsExpenses.TotalIncome == 0 {
		return 0, nil
	}

	return ((incomeVsExpenses.TotalIncome - incomeVsExpenses.TotalExpenses) / incomeVsExpenses.TotalIncome) * 100, nil
}

// GetDebtToIncomeRatio calculates the debt-to-income ratio for a user
func (s *MetricsService) GetDebtToIncomeRatio(ctx context.Context, userID string) (float64, error) {
	monthlyDebt, err := s.repo.GetMonthlyDebtPaymentsByUser(ctx, userID)
	if err != nil {
		return 0, err
	}

	// Get income using GetIncomeVsExpenses for the current month
	filter := model.AnalyticsFilter{
		UserID:    userID,
		StartDate: time.Now().AddDate(0, 0, -30), // Last 30 days
		EndDate:   time.Now(),
	}
	incomeVsExpenses, err := s.repo.GetIncomeVsExpenses(ctx, filter)
	if err != nil {
		return 0, err
	}

	monthlyIncome := incomeVsExpenses.TotalIncome
	if monthlyIncome == 0 {
		return 0, nil
	}

	return (monthlyDebt / monthlyIncome) * 100, nil
}

// GetEmergencyFundCoverage calculates how many months of expenses are covered by emergency fund for a user
func (s *MetricsService) GetEmergencyFundCoverage(ctx context.Context, userID string) (float64, error) {
	emergencyFund, err := s.repo.GetEmergencyFundBalanceByUser(ctx, userID)
	if err != nil {
		return 0, err
	}

	monthlyExpenses, err := s.repo.GetAverageMonthlyExpensesByUser(ctx, userID)
	if err != nil {
		return 0, err
	}

	if monthlyExpenses == 0 {
		return 0, nil
	}

	return emergencyFund / monthlyExpenses, nil
}
