package notifications

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

var emailClient = &http.Client{Timeout: 10 * time.Second}

func SendSubmissionResult(to, name, status, challengeTitle string, credits int) error {
	apiKey := os.Getenv("SENDGRID_API_KEY")
	from := os.Getenv("EMAIL_FROM")
	if apiKey == "" || from == "" || to == "" {
		return nil
	}

	var subject, body string
	switch status {
	case "done":
		subject = fmt.Sprintf("Your submission for \"%s\" was accepted!", challengeTitle)
		body = fmt.Sprintf("Hi %s,\n\nGreat news! Your robot data submission for \"%s\" has been processed successfully. You earned %d credits.\n\nKeep it up!\nThe HumanLoop Team", name, challengeTitle, credits)
	case "failed":
		subject = fmt.Sprintf("Your submission for \"%s\" could not be processed", challengeTitle)
		body = fmt.Sprintf("Hi %s,\n\nUnfortunately we could not process your submission for \"%s\". Please try again with a clearer video.\n\nThe HumanLoop Team", name, challengeTitle)
	case "synthetic":
		subject = "Submission rejected — synthetic video detected"
		body = fmt.Sprintf("Hi %s,\n\nYour submission was rejected because it appears to be a synthetic or AI-generated video. Only real robot footage is accepted.\n\nThe HumanLoop Team", name)
	default:
		return nil
	}

	payload, err := json.Marshal(map[string]any{
		"personalizations": []map[string]any{
			{"to": []map[string]string{{"email": to, "name": name}}},
		},
		"from":    map[string]string{"email": from},
		"subject": subject,
		"content": []map[string]string{{"type": "text/plain", "value": body}},
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, "https://api.sendgrid.com/v3/mail/send", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := emailClient.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}
