package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const envPath = "data/.env"

var httpClient = &http.Client{Timeout: 10 * time.Second}
var telegramAPIBase = "https://api.telegram.org"

func loadEnv() {
	data, err := os.ReadFile(envPath)
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		k = strings.TrimSpace(k)
		v = strings.Trim(strings.TrimSpace(v), "\"'")
		if os.Getenv(k) == "" {
			os.Setenv(k, v)
		}
	}
}

func telegramConfigured() bool {
	return os.Getenv("TELEGRAM_BOT_TOKEN") != "" && os.Getenv("TELEGRAM_CHAT_ID") != ""
}

func sendTelegram(message string) error {
	var lastErr error

	for attempt := 0; attempt < 2; attempt++ {
		if attempt > 0 {
			time.Sleep(2 * time.Second)
		}

		lastErr = doSend(message)
		if lastErr == nil {
			return nil
		}
	}

	return lastErr
}

func doSend(message string) error {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	chatID := os.Getenv("TELEGRAM_CHAT_ID")

	apiURL := fmt.Sprintf("%s/bot%s/sendMessage", telegramAPIBase, token)

	resp, err := httpClient.PostForm(apiURL, url.Values{
		"chat_id": {chatID},
		"text":    {message},
	})
	if err != nil {
		return fmt.Errorf("telegram send failed (network error)")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram returned status %d", resp.StatusCode)
	}

	return nil
}
