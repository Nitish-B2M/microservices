// Package utils email-template
package utils

const UserCreatedTemplate = `
Hello {{.Email}},<br><br>
Your account has been successfully created. Welcome to our platform!<br><br>

Email: {{.Email}}<br>
Role: {{.Role}}<br><br>
Thank you for joining us!<br><br>
Best regards,<br>
The Team
`
