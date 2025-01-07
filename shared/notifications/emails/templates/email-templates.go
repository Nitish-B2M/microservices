package templates

// Email Subject
const (
	UserRegistrationSuccessFul       = "Congratulation User, For successfully registration with us"
	TestEmailSubject                 = "Testing email service"
	OrderPlacedSuccessfullySubject   = "Your Order Has Been Successfully Placed!"
	OrderConfirmationReceivedSubject = "Order #%s Confirmation - We’ve Received Your Order!"
	OrderShippedSubject              = "Your Order #%s Has Been Shipped!"
	OrderDeliveredSubject            = "Your Order #%s Has Been Delivered!"
	OrderCancelledSubject            = "Your Order #%s Has Been Canceled"
	PaymentFailedSubject             = "Payment Unsuccessful for Your Order"
	OrderRefundProcessedSubject      = "Your Order #%s Refund Has Been Processed"
	OrderUpdatedSubject              = "Your Order #%s Has Been Updated"
	OrderAwaitingPaymentSubject      = "Action Required: Your Order is Awaiting Payment"
	OrderInvoiceAttachedSubject      = "Your Order #%s Invoice File Attached Below"
)

const TEST_EMAIL_TEMPLATE = `
Hello World!<br>
If you received this message, so please ignore this message.<br>
This message for just testing purposes only.<br><br>

{{.CustomMessage}}<br>
The Team
`

const USER_CREATED_TEMPLATE = `
Hello {{.Email}},<br><br>
Your account has been successfully created. Welcome to our platform!<br><br>

Email: {{.Email}}<br>
Role: {{.Role}}<br><br>
Thank you for joining us!<br><br>
Best regards,<br>
The Team
`

const ORDER_SUCCESS_TEMPLATE = `
Hello {{.CustomerName}},<br><br>
Thank you for your order! We are happy to inform you that your order has been successfully placed.<br><br>

Order ID: {{.OrderID}}<br>
Total Amount: ${{.TotalAmount}}<br><br>

Your order will be processed soon, and you will receive a confirmation once it is shipped.<br><br>

If you have any questions, feel free to reach out to our support team at <a href="mailto:support@yourcompany.com">support@yourcompany.com</a>.<br><br>

Thank you for shopping with us!<br><br>
Best regards,<br>
The Team
`

const ORDER_SHIPPED_TEMPLATE = `
Hello {{.CustomerName}},<br><br>
Your order is on its way! We are happy to inform you that your order has been shipped.<br><br>

Order ID: {{.OrderID}}<br>
Total Amount: ${{.TotalAmount}}<br><br>

Shipping Method: {{.ShippingMethod}}<br>
Tracking Number: {{.TrackingNumber}}<br><br>

You can track your shipment using the tracking number above.<br><br>

If you have any questions, feel free to reach out to our support team at <a href="mailto:support@yourcompany.com">support@yourcompany.com</a>.<br><br>

Thank you for shopping with us!<br><br>
Best regards,<br>
The Team
`

const ORDER_PLACED_TEMPLATE = `
Hello {{.CustomerName}},<br><br>
Thank you for your order! We are happy to inform you that your order has been successfully placed.<br><br>

Order ID: {{.OrderID}}<br>
Total Amount: ${{.TotalAmount}}<br><br>

Your order will be processed soon, and you will receive a confirmation once it is shipped.<br><br>

If you have any questions, feel free to reach out to our support team at <a href="mailto:support@yourcompany.com">support@yourcompany.com</a>.<br><br>

Thank you for shopping with us!<br><br>
Best regards,<br>
The Team
`

const ORDER_CONFIRMATION_TEMPLATE = `
Hello {{.CustomerName}},<br><br>
We have received your order and it is currently being processed!<br><br>

Order ID: {{.OrderID}}<br>
Total Amount: ${{.TotalAmount}}<br><br>

You will be notified once your order is ready to ship. Your items are being prepared.<br><br>

If you have any questions, please contact our support team at <a href="mailto:support@yourcompany.com">support@yourcompany.com</a>.<br><br>

Thank you for shopping with us!<br><br>
Best regards,<br>
The Team
`

const ORDER_DELIVERED_TEMPLATE = `
	Hello {{.CustomerName}},<br><br>
	Good news! Your order has been successfully delivered.<br><br>
	
	Order ID: {{.OrderID}}<br>
	Total Amount: ${{.TotalAmount}}<br><br>
	
	Delivery Date: {{.DeliveryDate}}<br>
	Shipping Address: {{.ShippingAddress}}<br><br>
	
	We hope you enjoy your purchase! If you need any assistance, please don’t hesitate to contact our support team.<br><br>
	
	If you have any feedback or questions, please reach out to us at <a href="mailto:support@yourcompany.com">support@yourcompany.com</a>.<br><br>
	
	Thank you for shopping with us!<br><br>
	Best regards,<br>
	The Team
`

const ORDER_CANCELLED_TEMPLATE = `
	Hello {{.CustomerName}},<br><br>
	We regret to inform you that your order has been canceled.<br><br>
	
	Order ID: {{.OrderID}}<br>
	Total Amount: ${{.TotalAmount}}<br><br>
	
	Reason for cancellation: {{.CancellationReason}}<br><br>
	
	If you believe this was a mistake or if you need further assistance, please contact our support team at <a href="mailto:support@yourcompany.com">support@yourcompany.com</a>.<br><br>
	
	We apologize for any inconvenience caused. Thank you for your understanding.<br><br>
	Best regards,<br>
	The Team
`

const ORDER_REFUND_TEMPLATE = `
	Hello {{.CustomerName}},<br><br>
	We want to inform you that a refund has been processed for your order.<br><br>
	
	Order ID: {{.OrderID}}<br>
	Refund Amount: ${{.RefundAmount}}<br><br>
	
	Refund Method: {{.RefundMethod}}<br><br>
	
	If you have any questions or need further assistance, please contact us at <a href="mailto:support@yourcompany.com">support@yourcompany.com</a>.<br><br>
	
	Thank you for your understanding.<br><br>
	Best regards,<br>
	The Team
`

const ORDER_UPDATE_TEMPLATE = `
	Hello {{.CustomerName}},<br><br>
	We wanted to inform you that there has been an update to your order.<br><br>
	
	Order ID: {{.OrderID}}<br>
	Total Amount: ${{.TotalAmount}}<br><br>
	
	Changes to your order: {{.OrderChanges}}<br><br>
	
	If you have any questions or need further assistance, please reach out to our support team at <a href="mailto:support@yourcompany.com">support@yourcompany.com</a>.<br><br>
	
	We appreciate your business!<br><br>
	Best regards,<br>
	The Team
`

const ORDER_INVOICE_TEMPLATE = `
Hello {{.CustomerName}},<br><br>
Thank you for your order! We are pleased to send you the invoice for your recent purchase.<br><br>

<b>Invoice Number:</b> {{.InvoiceID}}<br>
<b>Order ID:</b> {{.OrderID}}<br>

Your invoice is attached to this email as a PDF file. You can download and review it anytime.<br><br>

If you have any questions or concerns, feel free to contact our support team.<br><br>

Thank you for your purchase!<br><br>
Best regards,<br>
The Team
`

// ************payment email template*********
const PAYMENT_FAILED_TEMPLATE = `
	Hello {{.CustomerName}},<br><br>
	Unfortunately, your payment for order #{{.OrderID}} was unsuccessful.<br><br>
	
	Order ID: {{.OrderID}}<br>
	Total Amount: ${{.TotalAmount}}<br><br>
	
	Reason for failure: {{.PaymentFailureReason}}<br><br>
	
	Please review your payment details and try again. If you have any questions, feel free to contact our support team at <a href="mailto:support@yourcompany.com">support@yourcompany.com</a>.<br><br>
	
	We apologize for the inconvenience and hope to resolve this issue promptly.<br><br>
	Best regards,<br>
	The Team
`
const ORDER_AWAITING_PAYMENT_TEMPLATE = `
	Hello {{.CustomerName}},<br><br>
	We noticed that your order is currently awaiting payment.<br><br>
	
	Order ID: {{.OrderID}}<br>
	Total Amount: ${{.TotalAmount}}<br><br>
	
	Please complete your payment to proceed with the order.<br><br>
	
	If you need any assistance with payment, feel free to contact us at <a href="mailto:support@yourcompany.com">support@yourcompany.com</a>.<br><br>
	
	Thank you for your prompt attention.<br><br>
	Best regards,<br>
	The Team
`
