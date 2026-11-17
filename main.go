package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

const (
	BotToken = "8638263479:AAEZ6He5F3RayRsS-E8v-SyndkItJV02IME"
	BotID    = "@mybenshop_bot"
)

// بخش بعد از ":" توکن به عنوان مسیر وبهوک استفاده می‌شود
var hook = strings.Split(BotToken, ":")[1]

// ساختار مورد نیاز از آپدیت تلگرام
type Update struct {
	Message *struct {
		Text string `json:"text"`
		Chat struct {
			ID int64 `json:"id"`
		} `json:"chat"`
	} `json:"message"`
}

func main() {
	http.HandleFunc("/", handler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("listening on :" + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
	// دریافت آپدیت از تلگرام
	if r.URL.Path == "/"+hook {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		var update Update
		if err := json.Unmarshal(body, &update); err != nil {
			http.Error(w, "bad json", http.StatusBadRequest)
			return
		}

		// فقط پیام‌های متنی پردازش می‌شوند
		if update.Message != nil && update.Message.Text != "" {
			chatID := update.Message.Chat.ID
			text := update.Message.Text

			if _, err := req("sendMessage", map[string]any{
				"chat_id": chatID,
				"text":    reply(text),
			}); err != nil {
				log.Println("sendMessage error:", err)
			}
		}

		fmt.Fprint(w, "ok")
		return
	}

	// ست شدن وبهوک
	if r.Method == http.MethodGet && r.URL.Path == "/" {
		webhookURL := fmt.Sprintf("%s/%s", origin(r), hook)

		result, err := req("setWebhook", map[string]any{
			"url": webhookURL,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(result)
		return
	}

	fmt.Fprint(w, "ok")
}

// منطق پاسخ‌دهی ربات
func reply(text string) string {
	switch strings.TrimSpace(text) {
	case "سلام":
		return "علیک سلام"
	case "خدافظ", "خداحافظ":
		return "خدافظ✋"
	default:
		return "تو گفتی: " + text
	}
}

// معادل url.origin در جاوااسکریپت
func origin(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	return scheme + "://" + r.Host
}

// تابع ارسال درخواست به API تلگرام
func req(method string, data map[string]any) (map[string]any, error) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/%s", BotToken, method)

	var (
		resp *http.Response
		err  error
	)

	if data != nil {
		payload, mErr := json.Marshal(data)
		if mErr != nil {
			return nil, mErr
		}
		resp, err = http.Post(url, "application/json", bytes.NewReader(payload))
	} else {
		resp, err = http.Get(url)
	}
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}
