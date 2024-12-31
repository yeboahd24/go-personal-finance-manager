package model

import "time"

type SpendingByCategory struct {
	CategoryID   string  `json:"category_id"`
	CategoryName string  `json:"category_name"`
	Amount      float64 `json:"amount"`
	Percentage  float64 `json:"percentage"`
}

type MonthlySpending struct {
	Month    time.Time `json:"month"`
	Amount   float64   `json:"amount"`
	Change   float64   `json:"change"` // Percentage change from previous month
}

type CashFlow struct {
	Income    float64 `json:"income"`
	Expenses  float64 `json:"expenses"`
	NetIncome float64 `json:"net_income"`
}

type MerchantSpending struct {
	MerchantName string  `json:"merchant_name"`
	Amount       float64 `json:"amount"`
	Transactions int     `json:"transactions"`
	Percentage   float64 `json:"percentage"`
}

type FinancialReport struct {
	Period             string               `json:"period"` // month, quarter, year
	CashFlow           CashFlow            `json:"cash_flow"`
	TopCategories      []SpendingByCategory `json:"top_categories"`
	TopMerchants       []MerchantSpending  `json:"top_merchants"`
	MonthlyTrend       []MonthlySpending   `json:"monthly_trend"`
	AverageDailySpend  float64             `json:"average_daily_spend"`
	LargestExpense     Transaction         `json:"largest_expense"`
	AvgMonthlyIncome   float64             `json:"avg_monthly_income"`
	AvgMonthlyExpenses float64             `json:"avg_monthly_expenses"`
	MinMonthlyExpenses float64             `json:"min_monthly_expenses"`
	MaxMonthlyExpenses float64             `json:"max_monthly_expenses"`
	SavingsRate        float64             `json:"savings_rate"`
}

type AnalyticsFilter struct {
	UserID      string
	StartDate   time.Time
	EndDate     time.Time
	CategoryID  string
	Limit       int
}

type BudgetPerformance struct {
	Period           string                        `json:"period"`
	TotalBudget      float64                      `json:"total_budget"`
	TotalSpent       float64                      `json:"total_spent"`
	RemainingBudget  float64                      `json:"remaining_budget"`
	SpendingProgress float64                      `json:"spending_progress"` // Percentage of budget spent
	Categories       []CategoryBudgetPerformance   `json:"categories"`
}

type CategoryBudgetPerformance struct {
	CategoryID      string  `json:"category_id"`
	CategoryName    string  `json:"category_name"`
	BudgetAmount    float64 `json:"budget_amount"`
	SpentAmount     float64 `json:"spent_amount"`
	RemainingAmount float64 `json:"remaining_amount"`
	SpendingProgress float64 `json:"spending_progress"` // Percentage of category budget spent
}

type IncomeVsExpenses struct {
	Period          string    `json:"period"`
	TotalIncome     float64   `json:"total_income"`
	TotalExpenses   float64   `json:"total_expenses"`
	NetAmount       float64   `json:"net_amount"`
	IncomeBreakdown []Income  `json:"income_breakdown"`
	ExpenseBreakdown []Expense `json:"expense_breakdown"`
}

type Income struct {
	Source    string  `json:"source"`
	Amount    float64 `json:"amount"`
	Percentage float64 `json:"percentage"`
}

type Expense struct {
	Category   string  `json:"category"`
	Amount     float64 `json:"amount"`
	Percentage float64 `json:"percentage"`
}
