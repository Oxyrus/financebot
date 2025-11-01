package expense

import "fmt"

// Item represents a single categorized expense produced by the extractor.
type Item struct {
	Category    string  `json:"category"`
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
}

// ReplyMessage formats a Telegram-friendly confirmation string.
func (e Item) ReplyMessage() string {
	return fmt.Sprintf(
		"Recorded\nDescription: %s\nCategory: %s\nAmount: $%.2f",
		e.Description,
		e.Category,
		e.Amount,
	)
}
