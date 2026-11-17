package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

const (
	BotToken = "8638263479:AAEZ6He5F3RayRsS-E8v-SyndkItJV02IME"
	BotID    = "@mybenshop_bot"
)

var hook string

func init() {
	// معادل: const HOOK = BOT_TOKEN.split(":")[1]
	parts := strings.Split(BotToken, ":")
	if len(parts) > 1 {
		hook = parts[1]
	} else {
		hook = BotToken
	}
}

// ساختارهای داده برای دریافت آپدیت‌های تلگرام
type User struct {
	ID int64 `json:"id"`
}

type Message struct {
	From User `json:"from"`
}

type CallbackQuery struct {
	ID   string `json:"id"`
	From User   `json:"from"`
}

type Update struct {
	Message       *Message       `json:"message"`
	CallbackQuery *CallbackQuery `json:"callback_query"`
}

func main() {
	// مسیر دریافت وب‌هوک
	http.HandleFunc("/"+hook, handleWebhook)
	// مسیر تنظیم وب‌هوک (GET روی ریشه /)
	http.HandleFunc("/", handleSetWebhook)

	fmt.Println("Server is running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// مدیریت درخواست‌های وب‌هوک تلگرام
func handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var update Update
	err := json.NewDecoder(r.Body).Decode(&update)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var chatID int64

	// قسمت message
	if update.Message != nil {
		chatID = update.Message.From.ID

		updateBytes, _ := json.Marshal(update)
		reqTelegram("sendMessage", map[string]interface{}{
			"chat_id": chatID,
			"text":    string(updateBytes),
		})
	}

	// قسمت callback
	if update.CallbackQuery != nil {
		chatID = update.CallbackQuery.From.ID

		reqTelegram("answerCallbackQuery", map[string]interface{}{
			"callback_query_id": update.CallbackQuery.ID,
			"text":              "دکمه کلیک شد!",
		})

		updateBytes, _ := json.Marshal(update)
		reqTelegram("sendMessage", map[string]interface{}{
			"chat_id": chatID,
			"text":    string(updateBytes),
		})
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

// تنظیم کردن وب‌هوک تلگرام
func handleSetWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet || r.URL.Path != "/" {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
		return
	}

	// تشخیص پروتکل و هاست (در محیط لوکال یا سرور)
	scheme := "https"
	if r.TLS == nil && !strings.Contains(r.Host, "localhost") {
		// اگر پشت پراکسی مثل Nginx یا Cloudflare هستید
		if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
			scheme = proto
		}
	} else if strings.Contains(r.Host, "localhost") {
		scheme = "http"
	}

	webhookURL := fmt.Sprintf("%s://%s/%s", scheme, r.Host, hook)

	res := reqTelegram("setWebhook", map[string]interface{}{
		"url": webhookURL,
	})

	w.Header().Set("Content-Type", "application/json")
	w.Write(res)
}

// تابع کمکی برای ارسال درخواست به تلگرام (معادل تابع req در جاوااسکریپت)
func reqTelegram(method string, data map[string]interface{}) []byte {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/%s", BotToken, method)

	var bodyReader io.Reader
	httpMethod := http.MethodGet

	if data != nil {
		httpMethod = http.MethodPost
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil
		}
		bodyReader = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(httpMethod, apiURL, bodyReader)
	if err != nil {
		return nil
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	return respBody
}
