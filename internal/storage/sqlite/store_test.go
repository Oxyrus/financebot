package sqlite

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/Oxyrus/financebot/internal/expense"
)

func TestNewStoreRequiresPath(t *testing.T) {
	if _, err := NewStore(""); err == nil {
		t.Fatal("expected error when database path is empty")
	}
}

func TestSQLiteStoreSaveExpense(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "finance.db")
	store, err := NewStore(path)
	if err != nil {
		t.Fatalf("NewStore error: %v", err)
	}
	defer store.Close()

	item := expense.Item{
		Category:    "Travel",
		Amount:      45.67,
		Description: "Taxi",
	}

	if err := store.SaveExpense(context.Background(), item); err != nil {
		t.Fatalf("SaveExpense error: %v", err)
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open sqlite verify: %v", err)
	}
	defer db.Close()

	var (
		category    string
		amount      float64
		description string
	)

	if err := db.QueryRow(`SELECT category, amount, description FROM expenses LIMIT 1`).Scan(&category, &amount, &description); err != nil {
		t.Fatalf("verify inserted row: %v", err)
	}

	if category != item.Category || amount != item.Amount || description != item.Description {
		t.Fatalf("unexpected row values: got %q, %f, %q", category, amount, description)
	}
}

func TestSQLiteStoreRejectsEmptyDescription(t *testing.T) {
	path := filepath.Join(t.TempDir(), "finance.db")
	store, err := NewStore(path)
	if err != nil {
		t.Fatalf("NewStore error: %v", err)
	}
	defer store.Close()

	err = store.SaveExpense(context.Background(), expense.Item{
		Category: "General",
		Amount:   10,
	})

	if err == nil {
		t.Fatal("expected error for empty description")
	}
}
