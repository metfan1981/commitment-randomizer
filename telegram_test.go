package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestDoSend_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	old := telegramAPIBase
	telegramAPIBase = server.URL
	defer func() { telegramAPIBase = old }()

	os.Setenv("TELEGRAM_BOT_TOKEN", "test-token")
	os.Setenv("TELEGRAM_CHAT_ID", "12345")
	defer os.Unsetenv("TELEGRAM_BOT_TOKEN")
	defer os.Unsetenv("TELEGRAM_CHAT_ID")

	err := doSend("hello")
	if err != nil {
		t.Errorf("expected success, got: %v", err)
	}
}

func TestDoSend_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	old := telegramAPIBase
	telegramAPIBase = server.URL
	defer func() { telegramAPIBase = old }()

	os.Setenv("TELEGRAM_BOT_TOKEN", "test-token")
	os.Setenv("TELEGRAM_CHAT_ID", "12345")
	defer os.Unsetenv("TELEGRAM_BOT_TOKEN")
	defer os.Unsetenv("TELEGRAM_CHAT_ID")

	err := doSend("hello")
	if err == nil {
		t.Fatal("expected error for 400 response")
	}
	if !strings.Contains(err.Error(), "400") {
		t.Errorf("error should mention status code, got: %v", err)
	}
}

func TestDoSend_NetworkError(t *testing.T) {
	old := telegramAPIBase
	telegramAPIBase = "http://127.0.0.1:1" // nothing listening
	defer func() { telegramAPIBase = old }()

	os.Setenv("TELEGRAM_BOT_TOKEN", "secret-token-value")
	os.Setenv("TELEGRAM_CHAT_ID", "12345")
	defer os.Unsetenv("TELEGRAM_BOT_TOKEN")
	defer os.Unsetenv("TELEGRAM_CHAT_ID")

	err := doSend("hello")
	if err == nil {
		t.Fatal("expected error for unreachable host")
	}
	if strings.Contains(err.Error(), "secret-token-value") {
		t.Error("error message must not contain the bot token")
	}
	if !strings.Contains(err.Error(), "network error") {
		t.Errorf("expected 'network error' in message, got: %v", err)
	}
}

func TestDoSend_RequestContainsChatAndText(t *testing.T) {
	var gotChatID, gotText string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		gotChatID = r.FormValue("chat_id")
		gotText = r.FormValue("text")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	old := telegramAPIBase
	telegramAPIBase = server.URL
	defer func() { telegramAPIBase = old }()

	os.Setenv("TELEGRAM_BOT_TOKEN", "test-token")
	os.Setenv("TELEGRAM_CHAT_ID", "99999")
	defer os.Unsetenv("TELEGRAM_BOT_TOKEN")
	defer os.Unsetenv("TELEGRAM_CHAT_ID")

	doSend("test message")

	if gotChatID != "99999" {
		t.Errorf("chat_id = %q, want %q", gotChatID, "99999")
	}
	if gotText != "test message" {
		t.Errorf("text = %q, want %q", gotText, "test message")
	}
}
