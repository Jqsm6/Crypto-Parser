package tools

import (
	"fmt"
	"net/http"
	"log"
)

func SendTelegramNotify(tgbotapi, chatid, text string) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage?chat_id=%s&text=%s", tgbotapi, chatid, text)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println("Error creating HTTP request:", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error sending HTTP request:", err)
	}
	defer resp.Body.Close()
}