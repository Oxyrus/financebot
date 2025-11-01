package memory

import (
	"context"
	"sync"
	"time"

	"github.com/Oxyrus/financebot/internal/expense"
	"github.com/Oxyrus/financebot/internal/storage"
)

// Store keeps expenses in-memory; useful for development and testing.
type Store struct {
	mu      sync.Mutex
	records []record
}

type record struct {
	item      expense.Item
	createdAt time.Time
}

var _ storage.ExpenseStore = (*Store)(nil)

// NewStore creates an empty in-memory store.
func NewStore() *Store {
	return &Store{records: make([]record, 0)}
}

// SaveExpense appends a new expense to memory.
func (s *Store) SaveExpense(_ context.Context, item expense.Item) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.records = append(s.records, record{item: item, createdAt: time.Now().UTC()})
	return nil
}

// Items returns a defensive copy of all stored expenses; primarily for tests.
func (s *Store) Items() []expense.Item {
	s.mu.Lock()
	defer s.mu.Unlock()

	cp := make([]expense.Item, len(s.records))
	for i, rec := range s.records {
		cp[i] = rec.item
	}
	return cp
}

// Close satisfies the ExpenseStore interface; no cleanup required.
func (s *Store) Close() error {
	return nil
}

// Stats aggregates expenses since the provided time.
func (s *Store) Stats(_ context.Context, since time.Time) (storage.Summary, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	summary := storage.Summary{CategoryTotals: make(map[string]float64)}

	for _, rec := range s.records {
		if rec.createdAt.Before(since) {
			continue
		}
		summary.TotalCount++
		summary.TotalAmount += rec.item.Amount
		summary.CategoryTotals[rec.item.Category] += rec.item.Amount
	}

	return summary, nil
}
