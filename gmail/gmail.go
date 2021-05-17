package gmail

import (
	"fmt"
	"log"
	"net/smtp"
)

const (
	endpoint = "https://api.twilio.com/2010-04-01/Accounts/%v/Messages"
)

type Gmail struct {
	Email     string
	AuthToken string
}

func (t *Gmail) Send(email, subject, message string) error {
	msg := fmt.Sprintf("From: %v\nTo: %v\nSubject: %v \n\n%v", t.Email, email, subject, message)
	err := smtp.SendMail("smtp.gmail.com:587",
		smtp.PlainAuth(
			"",
			t.Email,
			t.AuthToken,
			"smtp.gmail.com"),
		t.Email,
		[]string{email},
		[]byte(msg),
	)
	fmt.Println(msg)

	if err != nil {
		return fmt.Errorf("smtp error: %s", err)
	}

	log.Println("Sent it so hard 🤙🏽")
	return nil
}
