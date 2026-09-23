package model

type MonthlyExpense struct {
	Month     string  `json:"month"` // Format: YYYY-MM
	TotalCost float64 `json:"total_cost"`
}

type CategoryExpense struct {
	Category  string  `json:"category"`
	TotalCost float64 `json:"total_cost"`
}

type ExpenseAnalytics struct {
	MonthlyExpenses  []MonthlyExpense  `json:"monthly_expenses"`
	CategoryExpenses []CategoryExpense `json:"category_expenses"`
}
