package storage

import (
	"context"

	"github.com/Oxyrus/financebot/internal/expense"
)

// ExpenseStore persists categorized expenses.
type ExpenseStore interface {
	SaveExpense(ctx context.Context, item expense.Item) error
	Close() error
}
