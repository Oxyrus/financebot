package sqlite

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

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

func TestSQLiteStoreStats(t *testing.T) {
	path := filepath.Join(t.TempDir(), "finance.db")
	store, err := NewStore(path)
	if err != nil {
		t.Fatalf("NewStore error: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	if err := store.SaveExpense(ctx, expense.Item{Category: "Food", Amount: 10, Description: "Lunch"}); err != nil {
		t.Fatalf("SaveExpense food: %v", err)
	}
	if err := store.SaveExpense(ctx, expense.Item{Category: "Travel", Amount: 15.5, Description: "Taxi"}); err != nil {
		t.Fatalf("SaveExpense travel: %v", err)
	}

	old := time.Now().AddDate(0, 0, -10)
	if _, err := store.db.Exec(`INSERT INTO expenses (category, amount, description, created_at) VALUES (?, ?, ?, ?)`, "Old", 99, "Old expense", old); err != nil {
		t.Fatalf("insert old expense: %v", err)
	}

	summary, err := store.Stats(ctx, time.Now().AddDate(0, 0, -7))
	if err != nil {
		t.Fatalf("Stats error: %v", err)
	}

	if summary.TotalCount != 2 {
		t.Fatalf("expected 2 recent expenses, got %d", summary.TotalCount)
	}
	if summary.TotalAmount < 25.49 || summary.TotalAmount > 25.51 {
		t.Fatalf("unexpected total amount %.2f", summary.TotalAmount)
	}
	if summary.CategoryTotals["Food"] != 10 {
		t.Fatalf("unexpected food total %.2f", summary.CategoryTotals["Food"])
	}
	if summary.CategoryTotals["Travel"] != 15.5 {
		t.Fatalf("unexpected travel total %.2f", summary.CategoryTotals["Travel"])
	}
	if _, ok := summary.CategoryTotals["Old"]; ok {
		t.Fatalf("expected old expenses to be excluded")
	}
}
