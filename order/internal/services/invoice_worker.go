package services

import (
	"e-commerce-backend/order/internal/models"
	"e-commerce-backend/shared/invoices"
	"strconv"
	"strings"
	"time"
)

var InvoiceChannel = make(chan invoices.Invoice, 10)

func GenerateOrderInvoice(order models.Order, userData map[string]interface{}, invoice invoices.Invoice, taxAmount, totalAmount float64) {
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
	invoice.CompanyDetails.CompanyEmail = "nyofficialcc@outlook.com"
	invoice.CompanyDetails.CompanyUrl = "https://nitish-b2m.github.io/myportfolio.github.io/"

	//invoice basic
	invoice.InvoiceId = string(order.OrderID)
	invoice.Title = "Invoice"
	invoice.Date = time.Now().Format("2006-01-02 15:04:05")
	invoice.TaxAmount = strconv.FormatFloat(taxAmount, 'f', -1, 64)
	invoice.TotalAmount = strconv.FormatFloat(totalAmount, 'f', -1, 64)

	//seller data
	invoice.SellerDetails = invoice.UserDetails

	//send it into channel
	InvoiceChannel <- invoice
	go SendInvoice()
}

func SendInvoice() {
	for task := range InvoiceChannel {
		invoices.InvoiceGenerator(task)
	}
}
