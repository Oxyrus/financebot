package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Oxyrus/financebot/internal/expense"
	"github.com/Oxyrus/financebot/internal/storage"

	_ "modernc.org/sqlite" // pure Go SQLite driver
)

const (
	defaultMaxOpenConns = 1
	expenseInsert       = `INSERT INTO expenses (category, amount, description, created_at) VALUES (?, ?, ?, ?)`
	expenseSchema       = `CREATE TABLE IF NOT EXISTS expenses (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		category TEXT NOT NULL,
		amount REAL NOT NULL,
		description TEXT NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);`
)

// Store persists expenses in a local SQLite database file.
type Store struct {
	db           *sql.DB
	insertStmt   *sql.Stmt
	databasePath string
}

var _ storage.ExpenseStore = (*Store)(nil)

// NewStore opens (or creates) the SQLite database at the provided path.
func NewStore(databasePath string) (*Store, error) {
	if databasePath == "" {
		return nil, errors.New("sqlite: database path is required")
	}

	if err := ensureDir(databasePath); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", databasePath)
	if err != nil {
		return nil, fmt.Errorf("sqlite: open: %w", err)
	}

	db.SetMaxOpenConns(defaultMaxOpenConns)
	db.SetMaxIdleConns(defaultMaxOpenConns)
	db.SetConnMaxLifetime(0)

	if err := migrate(db); err != nil {
		db.Close()
		return nil, err
	}

	stmt, err := db.Prepare(expenseInsert)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("sqlite: prepare insert: %w", err)
	}

	return &Store{
		db:           db,
		insertStmt:   stmt,
		databasePath: databasePath,
	}, nil
}

// SaveExpense writes a new expense row to the database.
func (s *Store) SaveExpense(ctx context.Context, item expense.Item) error {
	if item.Description == "" {
		return errors.New("sqlite: expense description cannot be empty")
	}
	if _, err := s.insertStmt.ExecContext(ctx, item.Category, item.Amount, item.Description, time.Now().UTC()); err != nil {
		return fmt.Errorf("sqlite: insert expense: %w", err)
	}
	return nil
}

// Close flushes prepared statements and closes the underlying database connection.
func (s *Store) Close() error {
	if s.insertStmt != nil {
		s.insertStmt.Close()
	}
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// Stats aggregates spending grouped by category since the provided time.
func (s *Store) Stats(ctx context.Context, since time.Time) (storage.Summary, error) {
	summary := storage.Summary{CategoryTotals: make(map[string]float64)}

	rows, err := s.db.QueryContext(ctx, `
		SELECT category, COUNT(*), COALESCE(SUM(amount), 0) 
		FROM expenses
		WHERE created_at >= ?
			AND category IS NOT NULL
			AND category != ''
		GROUP BY category`, since.UTC())
	if err != nil {
		return summary, fmt.Errorf("sqlite: query stats: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			category string
			count    int64
			total    float64
		)
		if err := rows.Scan(&category, &count, &total); err != nil {
			return summary, fmt.Errorf("sqlite: scan stats: %w", err)
		}
		summary.TotalCount += int(count)
		summary.TotalAmount += total
		summary.CategoryTotals[category] = total
	}

	if err := rows.Err(); err != nil {
		return summary, fmt.Errorf("sqlite: stats rows: %w", err)
	}

	return summary, nil
}

func migrate(db *sql.DB) error {
	if _, err := db.Exec(expenseSchema); err != nil {
		return fmt.Errorf("sqlite: migrate schema: %w", err)
	}
	return nil
}

func ensureDir(databasePath string) error {
	dir := filepath.Dir(databasePath)
	if dir == "." {
		return nil
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("sqlite: ensure dir: %w", err)
	}
	return nil
}
