package emails

import (
	"bytes"
	"e-commerce-backend/shared/utils"
	"fmt"
	"github.com/joho/godotenv"
	"gopkg.in/gomail.v2"
	"html/template"
	"log"
	"os"
	"strconv"
	"time"
)

type EmailTestData struct {
	CustomMessage string
}

type UserCreation struct {
	Email string
	Role  string
}

type OrderInvoice struct {
	OrderID        string
	CustomerName   string
	TotalAmount    string
	LinkAttachment string
	PdfAttachment  string
}

type OrderInvoiceAttachment struct {
	InvoiceID    string
	OrderID      string
	CustomerName string
}

// GeneralEmailTemplate General Format
type GeneralEmailTemplate struct {
	To               string
	Subject          string
	BodyTemplateName string
	Files            []string
}

// NewGeneralEmailTemplate creates a new CommonEmail struct dynamically from a map of fields
func NewGeneralEmailTemplate(to, subject, bodyTemplateName string, files []string) GeneralEmailTemplate {
	emailTemplate := GeneralEmailTemplate{
		To:               to,
		Subject:          subject,
		BodyTemplateName: bodyTemplateName,
		Files:            files,
	}

	return emailTemplate
}

func ParseTemplate(templateName string, values interface{}) string {
	tmpl, err := template.New(templateName).Parse(templateName)
	if err != nil {
		log.Println(fmt.Sprintf(utils.TemplateParsingFailed, err))
		return ""
	}
	var msgBuffer bytes.Buffer
	err = tmpl.Execute(&msgBuffer, values)
	if err != nil {
		log.Println(fmt.Sprintf(utils.TemplateExecuteFailed, err))
		return ""
	}
	return msgBuffer.String()
}

func SendEmails(emailTemplate GeneralEmailTemplate) error {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatal("Error loading .env file from emails.go")
	}

	email := os.Getenv("EMAIL_ACC")
	password := os.Getenv("EMAIL_PASS")
	smtpHost := os.Getenv("EMAIL_SMTP_HOST")
	smtpPort := os.Getenv("EMAIL_PORT")
	smtpPortInt, _ := strconv.Atoi(smtpPort)

	// Create a new email message
	m := gomail.NewMessage()
	m.SetHeader("From", email)
	m.SetHeader("To", emailTemplate.To)
	m.SetHeader("Subject", emailTemplate.Subject)
	m.SetBody("text/html", emailTemplate.BodyTemplateName)
	m.SetHeader("Disposition-Notification-To", email)
	m.SetHeader("Return-Receipt-To", email)

	if emailTemplate.Files != nil {
		for _, file := range emailTemplate.Files {
			m.Attach(file)
		}
	}

	d := gomail.NewDialer(smtpHost, smtpPortInt, email, password)
	// Send the email
	if err := d.DialAndSend(m); err != nil {
		log.Println("Error sending email:", err)
		return err
	}
	go deleteInvoiceFiles(emailTemplate.Files)
	return nil
}

func deleteInvoiceFiles(files []string) {
	timeout := time.After(20 * time.Second)
	go func() {
		// Wait for the timeout signal or other tasks to finish
		<-timeout
		for _, file := range files {
			err := os.Remove(file)
			if err != nil {
				continue
			} else {
				log.Println("Deleted invoice file:", file)
			}
		}
	}()
}
