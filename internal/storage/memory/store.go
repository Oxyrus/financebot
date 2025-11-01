package memory

import (
	"context"
	"sync"

	"github.com/Oxyrus/financebot/internal/expense"
	"github.com/Oxyrus/financebot/internal/storage"
)

// Store keeps expenses in-memory; useful for development and testing.
type Store struct {
	mu       sync.Mutex
	expenses []expense.Item
}

var _ storage.ExpenseStore = (*Store)(nil)

// NewStore creates an empty in-memory store.
func NewStore() *Store {
	return &Store{expenses: make([]expense.Item, 0)}
}

// SaveExpense appends a new expense to memory.
func (s *Store) SaveExpense(_ context.Context, item expense.Item) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.expenses = append(s.expenses, item)
	return nil
}

// Items returns a defensive copy of all stored expenses; primarily for tests.
func (s *Store) Items() []expense.Item {
	s.mu.Lock()
	defer s.mu.Unlock()

	cp := make([]expense.Item, len(s.expenses))
	copy(cp, s.expenses)
	return cp
}

// Close satisfies the ExpenseStore interface; no cleanup required.
func (s *Store) Close() error {
	return nil
}
