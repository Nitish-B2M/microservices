// Package utils
package utils

const (
	UserCreatedSuccessfully = "User created successfully"
	UserDeleted             = "User deleted successfully"
	UserUpdated             = "User updated successfully"
	UserNotFoundError       = "user not found"
	UserNotActiveError      = "user is not active"
	RoleNotExistsError      = "role not exists"
)

const (
	SubjectUserCreated  = "Congratulation User, For successfully registration with us"
	UserCreatedTemplate = `
		<!DOCTYPE html>
		<html lang="en">
		<head>
			<meta charset="UTF-8">
			<title>Welcome to Our Platform</title>
			<style>
				body {
					font-family: Arial, sans-serif;
					color: #333;
					line-height: 1.6;
				}
				.container {
					max-width: 600px;
					margin: auto;
					padding: 20px;
					border: 1px solid #e0e0e0;
					border-radius: 8px;
					background-color: #f9f9f9;
				}
				.footer {
					margin-top: 30px;
					font-size: 12px;
					color: #999;
				}
			</style>
		</head>
		<body>
			<div class="container">
				<h2>Welcome, {{.FullName}}!</h2>
				<p>Hello {{.Email}},</p>
				<p>Your account has been successfully created. We're excited to have you join us!</p>

				<p>Here are a few things you can do next:</p>
				<ul>
					<li>Complete your profile</li>
					<li>Explore our platform</li>
					<li>Reach out to support if you need help</li>
				</ul>

				<p>Best regards,<br>
				The Team</p>

				<div class="footer">
					If you did not create this account, please ignore this email or contact our support.
				</div>
			</div>
		</body>
		</html>
	`
)

const (
	VerificationMailSubject  = "Please verify your email address"
	VerificationMailTemplate = `
		<!DOCTYPE html>
		<html lang="en">
		<head>
			<meta charset="UTF-8" />
			<title>Email Verification</title>
			<style>
				body {
					font-family: Arial, sans-serif;
					background-color: #f7f9fc;
					color: #333;
					padding: 20px;
				}
				.container {
					background: #ffffff;
					max-width: 600px;
					margin: 0 auto;
					padding: 30px;
					border-radius: 8px;
					box-shadow: 0 2px 8px rgba(0,0,0,0.1);
				}
				h2 {
					color: #2c3e50;
				}
				a.button {
					display: inline-block;
					padding: 12px 24px;
					margin-top: 20px;
					background-color: #3498db;
					color: #ffffff;
					text-decoration: none;
					border-radius: 5px;
					font-weight: bold;
				}
				.footer {
					margin-top: 30px;
					font-size: 12px;
					color: #888888;
				}
			</style>
		</head>
		<body>
			<div class="container">
				<h2>Hello {{.FullName}},</h2>
				<p>Thank you for registering with us! Please verify your email address by clicking the button below:</p>

				<p><a href="{{.VerificationURL}}" class="button">Verify Email</a></p>

				<p>If you did not create an account, please ignore this email.</p>

				<p>Best regards,<br/>The Team</p>

				<div class="footer">
					If the button doesn't work, copy and paste this URL into your browser:<br/>
					<a href="{{.VerificationURL}}">{{.VerificationURL}}</a>
				</div>
			</div>
		</body>
		</html>
	`
)
