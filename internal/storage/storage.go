package storage

import (
	"context"
	"time"

	"github.com/Oxyrus/financebot/internal/expense"
)

// ExpenseStore persists categorized expenses.
type ExpenseStore interface {
	SaveExpense(ctx context.Context, item expense.Item) error
	Close() error
	Stats(ctx context.Context, since time.Time) (Summary, error)
}

// Summary describes aggregate expense data over a period.
type Summary struct {
	TotalCount     int
	TotalAmount    float64
	CategoryTotals map[string]float64
}
