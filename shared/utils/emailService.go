package utils

import (
	"bytes"
	"fmt"
	"github.com/joho/godotenv"
	"gopkg.in/gomail.v2"
	"html/template"
	"log"
	"os"
)

type UserCreation struct {
	Email string
	Role  string
}

func SendEmail(to string, subject string, body string) error {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	email := os.Getenv("EMAIL_ACC")
	password := os.Getenv("EMAIL_PASS")

	smtpHost := "smtp.gmail.com"
	smtpPort := 587

	// Create a new email message
	m := gomail.NewMessage()
	m.SetHeader("From", email)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)
	m.SetHeader("Disposition-Notification-To", email)
	m.SetHeader("Return-Receipt-To", email)

	d := gomail.NewDialer(smtpHost, smtpPort, email, password)
	// Send the email
	if err := d.DialAndSend(m); err != nil {
		return err
	}
	return nil
}

func GenerateUserCreationMessage(user UserCreation) (string, error) {
	tmpl, err := template.New("user_created").Parse(USER_CREATED_TEMPLATE)
	if err != nil {
		return "", fmt.Errorf(TemplateParsingFailed, err)
	}

	var msgBuffer bytes.Buffer
	err = tmpl.Execute(&msgBuffer, user)
	if err != nil {
		return "", fmt.Errorf(TemplateExecuteFailed, err)
	}

	return msgBuffer.String(), nil
}
