package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Config captures runtime settings needed by the bot.
type Config struct {
	TelegramToken string
	OpenAIKey     string
	allowedUsers  map[string]struct{}
}

// Load reads environment variables (optionally via .env) and validates them.
func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, reading environment variables directly")
	}

	cfg := &Config{
		TelegramToken: os.Getenv("TELEGRAM_TOKEN"),
		OpenAIKey:     os.Getenv("OPENAI_API_KEY"),
		allowedUsers:  parseAllowedUsers(os.Getenv("AUTHORIZED_USERS")),
	}

	if cfg.TelegramToken == "" || cfg.OpenAIKey == "" {
		return nil, fmt.Errorf("TELEGRAM_TOKEN or OPENAI_API_KEY not set")
	}

	if len(cfg.allowedUsers) == 0 {
		cfg.allowedUsers = map[string]struct{}{"iamoxyrus": {}}
	}

	return cfg, nil
}

// IsUserAllowed checks whether the provided Telegram handle is authorized.
func (c *Config) IsUserAllowed(username string) bool {
	if len(c.allowedUsers) == 0 {
		return true
	}
	_, ok := c.allowedUsers[username]
	return ok
}

func parseAllowedUsers(raw string) map[string]struct{} {
	users := make(map[string]struct{})
	for _, user := range strings.Split(raw, ",") {
		user = strings.TrimSpace(user)
		if user == "" {
			continue
		}
		users[user] = struct{}{}
	}
	return users
}
