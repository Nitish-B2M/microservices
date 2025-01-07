package services

import (
	"e-commerce-backend/order/internal/models"
	"e-commerce-backend/shared/invoices"
	"e-commerce-backend/shared/notifications/emails"
	"e-commerce-backend/shared/notifications/emails/templates"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

var InvoiceChannel = make(chan invoices.Invoice, 10)

func GenerateOrderInvoice(order models.Order, userData map[string]interface{}, invoice invoices.Invoice) {
	taxAmount := order.TaxAmount
	subTotal := order.SubTotal
	discountAmt := order.DiscountAmount
	totalAmount := order.TotalAmount
	//user data
	invoice.UserDetails.Name = strings.Join([]string{userData["first_name"].(string), userData["last_name"].(string)}, " ")
	invoice.UserDetails.Email = userData["email"].(string)
	invoice.UserDetails.Address = "XYZ, city, state, country"

	//seller data
	invoice.SellerDetails.Name = strings.Join([]string{userData["first_name"].(string), userData["last_name"].(string)}, " ")
	invoice.SellerDetails.Email = userData["email"].(string)
	invoice.SellerDetails.Address = "XYZ, city, state, country"

	//company data (hard-coded temporary data)
	invoice.CompanyDetails.CompanyId = "123"
	invoice.CompanyDetails.CompanyName = "NY Official Company"
	invoice.CompanyDetails.CompanyAddress = "XYZ, city, state, country"
	invoice.CompanyDetails.CompanyEmail = "nyofficialcc@abc.com"
	invoice.CompanyDetails.CompanyUrl = "https://nitish-b2m.github.io/myportfolio.github.io/"

	//invoice basic
	invoice.InvoiceId = strconv.Itoa(order.OrderID)
	invoice.Title = "Invoice"
	invoice.Date = time.Now().Format("2006-01-02 15:04:05")
	invoice.TaxAmount = strconv.FormatFloat(taxAmount, 'f', -1, 64)
	invoice.SubTotal = strconv.FormatFloat(subTotal, 'f', -1, 64)
	invoice.TotalAmount = strconv.FormatFloat(totalAmount, 'f', -1, 64)
	invoice.TotalDiscount = strconv.FormatFloat(discountAmt, 'f', -1, 64)

	//seller data
	invoice.SellerDetails = invoice.UserDetails

	//send it into channel
	InvoiceChannel <- invoice
}

func SendInvoice() {
	var wg sync.WaitGroup
	for task := range InvoiceChannel {
		wg.Add(1)
		go func(invoice invoices.Invoice) {
			defer wg.Done()
			SendOrderSuccessMail(invoice)
			invoices.InvoiceGeneratorWithSendMail(invoice)
		}(task)
	}
	wg.Wait()
}

func SendOrderSuccessMail(invoice invoices.Invoice) {
	emailContent := emails.OrderInvoice{OrderID: invoice.InvoiceId, TotalAmount: invoice.TotalAmount, CustomerName: invoice.UserDetails.Name}
	emails.EmailWorkerWithGoRoutine(invoice.UserDetails.Email, fmt.Sprintf(templates.OrderConfirmationReceivedSubject, invoice.InvoiceId), templates.ORDER_CONFIRMATION_TEMPLATE, emailContent, []string{})
}
