package emails

import "e-commerce-backend/shared/utils"

func EmailWorkerWithGoRoutine(to, subject, bodyTemplateName string, bodyContent interface{}, files []string) {
	body := ParseTemplate(bodyTemplateName, bodyContent)
	emailTemplate := NewGeneralEmailTemplate(to, subject, body, files)
	go SendEmails(emailTemplate)
}

func EmailWorker(to, subject, bodyTemplateName string, bodyContent interface{}, files []string) {
	body := ParseTemplate(bodyTemplateName, bodyContent)
	emailTemplate := NewGeneralEmailTemplate(to, subject, body, files)
	err := SendEmails(emailTemplate)
	if err != nil {
		utils.LogErrorWithFilename("email_worker.go", err.Error(), map[string]interface{}{"error": err})
		return
	}
}
