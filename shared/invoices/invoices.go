package invoices

import (
	"e-commerce-backend/shared/notifications/emails"
	"e-commerce-backend/shared/notifications/emails/templates"
	"fmt"
	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/line"
	"github.com/johnfercher/maroto/v2/pkg/components/list"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/border"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/consts/linestyle"
	"github.com/johnfercher/maroto/v2/pkg/consts/orientation"
	"github.com/johnfercher/maroto/v2/pkg/consts/pagesize"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"
	"log"
	"strconv"
	"strings"
	"time"
)

func InvoiceGeneratorWithSendMail(invoice Invoice) {
	filePath, _ := InvoiceGenerator(invoice)
	body := emails.OrderInvoiceAttachment{OrderID: invoice.InvoiceId, InvoiceID: invoice.InvoiceId, CustomerName: invoice.UserDetails.Name}
	emails.EmailWorkerWithGoRoutine(invoice.UserDetails.Email, fmt.Sprintf(templates.OrderInvoiceAttachedSubject, invoice.InvoiceId), templates.ORDER_INVOICE_TEMPLATE, body, []string{filePath})
}

func InvoiceGenerator(invoice Invoice) (string, error) {
	cfg := config.NewBuilder().
		WithOrientation(orientation.Vertical).
		WithPageSize(pagesize.A4).
		WithLeftMargin(5).
		WithTopMargin(15).
		WithRightMargin(5).
		WithBottomMargin(15).
		Build()
	m := maroto.New(cfg)

	// 1. Header
	addHeader(m, invoice)
	// 2. Invoice Number
	addInvoiceDetails(m, invoice)
	// 3. Item List
	addItemList(m, invoice.InvoiceItemList)
	addLine(m, 0, 0.3, "dash")
	// 4. Footer - Signature and QR code
	invoice.addFooter(m)

	// Save the PDF file
	document, err := m.Generate()
	if err != nil {
		log.Fatal(err.Error())
		return "", err
	}

	outFilePath := fmt.Sprintf("D:/data/golang/microservices/e-commerce-nyoffical/shared/invoices/uploads/%s_invoice_%v.pdf", invoice.InvoiceId, time.Now().Unix())
	err = document.Save(outFilePath)
	if err != nil {
		log.Fatal(err.Error())
		return "", err
	}
	log.Println("PDF saved successfully. Filepath is", outFilePath)
	return outFilePath, nil
}

func addHeader(m core.Maroto, inv Invoice) {
	m.AddRow(20,
		col.New(3).Add(
			text.New(strings.Title(inv.UserDetails.Name), props.Text{
				Size:  8,
				Align: align.Right,
				Color: getRedColor(),
			}),
			text.New(inv.UserDetails.Email, props.Text{
				Top:   4,
				Style: fontstyle.BoldItalic,
				Size:  8,
				Align: align.Right,
				Color: getBlueColor(),
			}),
			text.New(inv.UserDetails.Address, props.Text{
				Top:   8,
				Style: fontstyle.BoldItalic,
				Size:  8,
				Align: align.Right,
				Color: getBlueColor(),
			}),
		),
		col.New(6),
		col.New(3).Add(
			text.New(strings.Join([]string{inv.CompanyDetails.CompanyName, inv.CompanyDetails.CompanyAddress}, ", "), props.Text{
				Size:  8,
				Align: align.Right,
				Color: getRedColor(),
			}),
			text.New(inv.CompanyDetails.CompanyEmail, props.Text{
				Top:   8,
				Style: fontstyle.BoldItalic,
				Size:  8,
				Align: align.Right,
				Color: getBlueColor(),
			}),
			text.New(inv.CompanyDetails.CompanyUrl, props.Text{
				Top:   10,
				Style: fontstyle.BoldItalic,
				Size:  8,
				Align: align.Right,
				Color: getBlueColor(),
			}),
		),
	)
	m.AddRow(20,
		text.NewCol(12, "Order Invoice", props.Text{
			Top:   5,
			Style: fontstyle.Bold,
			Align: align.Center,
			Size:  12,
		}),
	)
}

// Adds invoice details
func addInvoiceDetails(m core.Maroto, inv Invoice) {
	m.AddRow(10,
		text.NewCol(6, "Date: "+time.Now().Format("02 Jan 2006"), props.Text{
			Align: align.Left,
			Size:  10,
		}),
		text.NewCol(6, fmt.Sprintf("Invoice #%s", inv.InvoiceId), props.Text{
			Align: align.Right,
			Size:  10,
		}),
	)
	//m.AddRow(10, line.NewCol(12)) one way to add line
	addLine(m, 10, 0.4, "solid")
}

type Invoice struct {
	InvoiceId       string         `json:"invoice_id"`
	Date            string         `json:"date"`
	Title           string         `json:"title"`
	UserDetails     InvUserDetails `json:"user_details"`
	SellerDetails   InvUserDetails `json:"seller_details"`
	InvoiceItemList []InvoiceItem  `json:"invoice_item_list"`
	CompanyDetails  CompanyDetails `json:"company_details"`
	TaxAmount       string         `json:"tax_amount"`
	SubTotal        string         `json:"sub_total"`
	TotalDiscount   string         `json:"total_discount"`
	TotalAmount     string         `json:"total_amount"`
}

type InvUserDetails struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Address string `json:"address"`
}

type InvoiceItem struct {
	Item            string
	Description     string
	Quantity        string
	Price           string
	DiscountedPrice string
	Total           string
}

type CompanyDetails struct {
	CompanyId      string `json:"company_id"`
	CompanyName    string `json:"company_name"`
	CompanyUrl     string `json:"company_url"`
	CompanyAddress string `json:"company_address"`
	CompanyEmail   string `json:"company_email"`
}

func rowHeaderProperties() props.Text {
	return props.Text{
		Style: fontstyle.Bold,
		Top:   3,
		Left:  5,
	}
}

func rowProperties() props.Text {
	return props.Text{
		Top:  2,
		Left: 5,
	}
}

func (o InvoiceItem) GetHeader() core.Row {
	cRow := row.New(10).Add(
		text.NewCol(1, "Sr.", rowHeaderProperties()).WithStyle(&props.Cell{
			BorderColor:     &props.Color{Red: 255, Green: 0, Blue: 108},
			BorderType:      border.Left, // Left border
			BorderThickness: 0.3,
		}),
		text.NewCol(5, "Description", rowHeaderProperties()).WithStyle(&props.Cell{
			BorderColor:     &props.Color{Red: 0, Green: 0, Blue: 0},
			BorderType:      border.Full, // Add left and right borders
			BorderThickness: 0.6,         // Border thickness
		}),
		text.NewCol(1, "Qty.", rowHeaderProperties()),
		text.NewCol(1, "Price", rowHeaderProperties()),
		text.NewCol(2, "Discount", rowHeaderProperties()),
		text.NewCol(2, "Sub Tot.", rowHeaderProperties()),
	)

	cRow.WithStyle(&props.Cell{
		BackgroundColor: &props.Color{Red: 175, Green: 238, Blue: 238},
		BorderColor:     &props.Color{Red: 93, Green: 138, Blue: 168},
	})

	return cRow
}

func splitTextIntoWords(text string) []string {
	return strings.Fields(text)
}

func wrapText(text string, lineWidth int) []string {
	var lines []string
	words := splitTextIntoWords(text)
	var currentLine string

	for i, word := range words {
		if len(currentLine)+len(word)+1 <= lineWidth {
			if i == 0 {
				currentLine = word
			} else {
				currentLine += " " + word
			}
		} else {
			lines = append(lines, currentLine)
			currentLine = word
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}

func (o InvoiceItem) GetContent(i int) core.Row {
	descriptionLines := wrapText(o.Description, 20) // Wrap text to fit within the desired width (adjust accordingly)

	// Dynamically calculate row height based on the number of lines in the description
	rowHeight := float64(9 + len(descriptionLines) + 1)
	r := row.New(rowHeight).Add(
		text.NewCol(1, strconv.Itoa(i+1), rowHeaderProperties()),
	)

	// Add each wrapped line of the description
	descriptionText := ""
	for _, line := range descriptionLines {
		descriptionText += line + "\n" // Add a newline after each line
	}
	descriptionText = strings.TrimSpace(descriptionText)

	// Add the description text as a single column
	r.Add(text.NewCol(5, descriptionText, rowProperties()))

	r.Add(
		text.NewCol(1, o.Quantity, rowProperties()),
		text.NewCol(1, strToFloatToStr(o.Price), rowProperties()),
		text.NewCol(2, o.DiscountedPrice+"%", rowProperties()),
		text.NewCol(2, strToFloatToStr(o.Total), rowProperties()),
	)

	if i%2 == 0 {
		r.WithStyle(&props.Cell{
			BackgroundColor: &props.Color{Red: 240, Green: 240, Blue: 240},
		})
	} else {
		r.WithStyle(&props.Cell{
			BackgroundColor: &props.Color{Red: 250, Green: 250, Blue: 250},
		})
	}

	return r
}

// Adds a list of items to the invoice
func addItemList(m core.Maroto, items []InvoiceItem) {
	rows, err := list.Build[InvoiceItem](items)
	if err != nil {
		log.Fatal(err.Error())
	}
	m.AddRows(rows...)
}

func strToFloatToStr(s string) string {
	v, _ := strconv.ParseFloat(s, 64)
	return fmt.Sprintf("%.2f", v)
}

// Adds a footer with total and signature
func (inv Invoice) addFooter(m core.Maroto) {
	taxAmtStr := strToFloatToStr(inv.TaxAmount)
	subTotalAmtStr := strToFloatToStr(inv.SubTotal)
	totalAmtStr := strToFloatToStr(inv.TotalAmount)
	discountAmtStr := strToFloatToStr(inv.TotalDiscount)

	m.AddRow(8,
		text.NewCol(8, ""),
		text.NewCol(2, "Discount ", props.Text{
			Top:   2,
			Style: fontstyle.Bold,
			Size:  10,
			Align: align.Right,
		}, rowHeaderProperties()).WithStyle(&props.Cell{
			BackgroundColor: &props.Color{Red: 240, Green: 240, Blue: 240},
		}),
		text.NewCol(2, discountAmtStr, props.Text{
			Top:   2,
			Style: fontstyle.Bold,
			Size:  10,
			Align: align.Center,
		}, rowProperties()).WithStyle(&props.Cell{
			BackgroundColor: &props.Color{Red: 240, Green: 240, Blue: 240},
		}),
	)

	m.AddRow(8,
		text.NewCol(8, ""),
		text.NewCol(2, "Tax(18%) ", props.Text{
			Top:   2,
			Style: fontstyle.Bold,
			Size:  10,
			Align: align.Right,
		}, rowHeaderProperties()).WithStyle(&props.Cell{
			BackgroundColor: &props.Color{Red: 240, Green: 240, Blue: 240},
		}),
		text.NewCol(2, taxAmtStr, props.Text{
			Top:   2,
			Style: fontstyle.Bold,
			Size:  10,
			Align: align.Center,
		}, rowProperties()).WithStyle(&props.Cell{
			BackgroundColor: &props.Color{Red: 240, Green: 240, Blue: 240},
		}),
	)

	m.AddRow(8,
		text.NewCol(8, ""),
		text.NewCol(2, "Sub Total ", props.Text{
			Top:   2,
			Style: fontstyle.Bold,
			Size:  10,
			Align: align.Right,
		}, rowHeaderProperties()).WithStyle(&props.Cell{
			BackgroundColor: &props.Color{Red: 240, Green: 240, Blue: 240},
		}),
		text.NewCol(2, subTotalAmtStr, props.Text{
			Top:   2,
			Style: fontstyle.Bold,
			Size:  10,
			Align: align.Center,
		}, rowProperties()).WithStyle(&props.Cell{
			BackgroundColor: &props.Color{Red: 240, Green: 240, Blue: 240},
		}),
	)

	m.AddRow(8,
		text.NewCol(8, ""),
		text.NewCol(2, "Total Amount  ", props.Text{
			Top:   2,
			Style: fontstyle.Bold,
			Size:  10,
			Align: align.Right,
			Color: &props.WhiteColor,
		}, rowHeaderProperties()).WithStyle(&props.Cell{
			BackgroundColor: getDarkGrayColor(),
		}),
		text.NewCol(2, totalAmtStr, props.Text{
			Top:   2,
			Style: fontstyle.Bold,
			Size:  10,
			Align: align.Center,
			Color: &props.WhiteColor,
		}, rowProperties()).WithStyle(&props.Cell{
			BackgroundColor: getDarkGrayColor(),
		}),
	)

	//m.AddRow(40,
	//	signature.NewCol(6, "Authorized Signatory", props.Signature{FontFamily: fontfamily.Courier}),
	//	code.NewQrCol(6, "https://codeheim.io", props.Rect{
	//		Percent: 75,
	//		Center:  true,
	//	}),
	//)
}

func addLine(m core.Maroto, rowHeight, thickness float64, lineType string) {
	// Create a custom dashed line
	var lineStyle linestyle.Type
	if lineType == "" {
		return
	} else if strings.EqualFold(lineType, "dash") {
		lineStyle = linestyle.Dashed
	} else if strings.EqualFold(lineType, "solid") {
		lineStyle = linestyle.Solid
	}
	m.AddRow(rowHeight, line.NewCol(12, props.Line{
		Style:     lineStyle,
		Thickness: thickness,
		Color:     &props.Color{Red: 0, Green: 0, Blue: 0},
	}))
}

func getRedColor() *props.Color {
	return &props.Color{
		Red:   150,
		Green: 10,
		Blue:  10,
	}
}

func getBlueColor() *props.Color {
	return &props.Color{
		Red:   10,
		Green: 10,
		Blue:  150,
	}
}

func getGrayColor() *props.Color {
	return &props.Color{
		Red:   200,
		Green: 200,
		Blue:  200,
	}
}

func getDarkGrayColor() *props.Color {
	return &props.Color{
		Red:   55,
		Green: 55,
		Blue:  55,
	}
}
