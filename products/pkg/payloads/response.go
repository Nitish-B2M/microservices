package payloads

import (
	"gorm.io/datatypes"
	"time"
)

type ProductQtyUpdateResponse struct {
	ID    string `json:"id"`
	PName string `json:"product_name"`
}

type ProductResponse struct {
	ID            string                `json:"id"`
	SellerID      string                `json:"seller_id"`
	SKU           string                `json:"sku"`
	Name          string                `json:"name"`
	Description   string                `json:"description"`
	ShortDesc     string                `json:"short_desc"`
	Price         float64               `json:"price"`
	ComparePrice  float64               `json:"compare_price"`
	CostPrice     float64               `json:"cost_price"`
	Quantity      int                   `json:"quantity"`
	InStock       bool                  `json:"in_stock"`
	IsFeatured    bool                  `json:"is_featured"`
	Rating        float64               `json:"rating"`
	Tax           float64               `json:"tax"`
	Discount      float64               `json:"discount"`
	DiscountType  string                `json:"discount_type"`
	Category      EntityLabelResponse   `json:"category"`
	Brand         EntityLabelResponse   `json:"brand"`
	Weight        float64               `json:"weight"`
	Length        float64               `json:"length"`
	Width         float64               `json:"width"`
	Height        float64               `json:"height"`
	DimensionUnit string                `json:"dimension_unit"`
	StockStatus   string                `json:"stock_status"`
	MetaTitle     string                `json:"meta_title"`
	MetaDesc      string                `json:"meta_description"`
	SlugURL       string                `json:"slug_url"`
	MainImage     string                `json:"main_image"`
	CreatedAt     time.Time             `json:"created_at"`
	UpdatedAt     time.Time             `json:"updated_at"`
	DeletedAt     *time.Time            `json:"deleted_at"`
	Tags          []EntityLabelResponse `json:"tags"`

	// These fields will be handled by separate tables
	Images     []ProductImageResp     `json:"images"`
	Variants   []ProductVariantResp   `json:"variants"`
	Attributes []ProductAttributeResp `json:"attributes"`
}

// ProductImageResp represents an image associated with a product
type ProductImageResp struct {
	ID        string `json:"id"`
	PID       string `json:"product_id"`
	URL       string `json:"url"`
	IsMain    bool   `json:"is_main"`
	SortOrder int    `json:"sort_order"`
}

// ProductVariantResp represents product variations
type ProductVariantResp struct {
	ID       string         `json:"id"`
	PID      string         `json:"product_id"`
	SKU      string         `json:"sku"`
	Name     string         `json:"name"`
	Price    float64        `json:"price"`
	Quantity int            `json:"quantity"`
	Options  datatypes.JSON `json:"options"`
}

// ProductAttributeResp represents product specifications
type ProductAttributeResp struct {
	ID    string `json:"id"`
	PID   string `json:"product_id"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

type EntityLabelResponse struct {
	ProductEntityID uint   `json:"product_entity_id"`
	Entity          string `json:"entity_name"`
	EntityType      string `json:"entity_type"`
}
