package twilio

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type Twilio struct {
	SID       string
	AuthToken string
	Number    string
	Client    *http.Client
}

func (t *Twilio) Send(number, message string) error {
	v := url.Values{}
	v.Set("Body", message)
	v.Set("From", t.Number)
	v.Set("To", number)
	data := *strings.NewReader(v.Encode())

	req, err := http.NewRequest("POST", "https://api.twilio.com/2010-04-01/Accounts/"+t.SID+"/Messages.json", &data)
	if err != nil {
		return fmt.Errorf("error formatting twilio send text request: %w", err)
	}
	req.SetBasicAuth(t.SID, t.AuthToken)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := t.Client.Do(req)
	if err != nil {
		return fmt.Errorf("error making twilio send text request: %w", err)
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		var data map[string]interface{}
		decoder := json.NewDecoder(resp.Body)
		err := decoder.Decode(&data)
		if err == nil {
			fmt.Println(data["sid"])
		}
		fmt.Println("sent it so hard 🤙🏽")
		return nil
	}

	return fmt.Errorf("error making twilio send text request: %v %v", resp.StatusCode, resp.Status)
}
