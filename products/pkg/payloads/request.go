package payloads

import (
	"time"
)

type ProductRequest struct {
	ID            string     `json:"id"`
	SellerID      string     `json:"seller_id"`
	SKU           string     `json:"sku" validate:"required"`
	Name          string     `json:"name" validate:"required,min=3,max=50"`
	Description   string     `json:"description" validate:"required,min=10,max=500"`
	ShortDesc     string     `json:"short_desc" validate:"required,max=200"`
	Price         float64    `json:"price" validate:"required,gt=0"`
	ComparePrice  float64    `json:"compare_price" validate:"omitempty,gte=0"`
	CostPrice     float64    `json:"cost_price" validate:"omitempty,gte=0"`
	Quantity      int        `json:"quantity" validate:"required,gt=0"`
	InStock       bool       `json:"in_stock"`
	IsDeleted     bool       `json:"is_deleted"`
	IsActive      bool       `json:"is_active"`
	IsFeatured    bool       `json:"is_featured"`
	Rating        float64    `json:"rating" validate:"omitempty,gte=0"`
	Tax           float64    `json:"tax" validate:"omitempty,gte=0"`
	Discount      float64    `json:"discount" validate:"omitempty,gte=0"`
	DiscountType  string     `json:"discount_type" validate:"omitempty,oneof=percentage fixed"`
	Category      string     `json:"category"`
	Brand         string     `json:"brand"`
	Weight        float64    `json:"weight" validate:"omitempty,gte=0"`
	Length        float64    `json:"length" validate:"omitempty,gte=0"`
	Width         float64    `json:"width" validate:"omitempty,gte=0"`
	Height        float64    `json:"height" validate:"omitempty,gte=0"`
	DimensionUnit string     `json:"dimension_unit" validate:"omitempty,oneof=cm inch meter"`
	StockStatus   string     `json:"stock_status" validate:"omitempty,oneof=in_stock out_of_stock preorder"`
	MetaTitle     string     `json:"meta_title" validate:"omitempty,max=100"`
	MetaDesc      string     `json:"meta_description" validate:"omitempty,max=300"`
	SlugURL       string     `json:"slug_url"`
	MainImage     string     `json:"main_image" validate:"omitempty,url"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	DeletedAt     *time.Time `json:"deleted_at"`
	Tags          []string   `json:"tags"`

	// These fields will be handled by separate tables
	Images     []ProductImageRequest     `json:"images" gorm:"foreignKey:PID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Variants   []ProductVariantRequest   `json:"variants" gorm:"foreignKey:PID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Attributes []ProductAttributeRequest `json:"attributes" gorm:"foreignKey:PID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

// type ProductVariantRequest struct {
// 	ID       string         `json:"id" gorm:"primaryKey;"`
// 	PID      string         `json:"product_id" gorm:"index;type:varchar(36)"`
// 	SKU      string         `json:"sku" gorm:"type:varchar(255)"`
// 	Name     string         `json:"name" gorm:"type:varchar(255)"`
// 	Price    float64        `json:"price" validate:"required,gt=0"`
// 	Quantity int            `json:"quantity" validate:"required,gt=0"`
// 	Options  datatypes.JSON `json:"options" gorm:"type:json" validate:"omitempty"` // Stored as JSON string
// }

type ProductVariantRequest struct {
	ID       string                 `json:"id"`
	PID      string                 `json:"pid"`
	SKU      string                 `json:"sku"`
	Name     string                 `json:"name"`
	Price    float64                `json:"price"`
	Quantity int                    `json:"quantity"`
	Options  []VariantOptionRequest `json:"options"`
}

type VariantOptionRequest struct {
	OptionName  string `json:"option_name"`
	OptionValue string `json:"option_value"`
}

type ProductAttributeRequest struct {
	ID    string `json:"id" gorm:"primaryKey"`
	PID   string `json:"product_id" gorm:"index;type:varchar(36)"`
	Name  string `json:"name" validate:"required"`
	Value string `json:"value" validate:"required"`
}

type ProductImageRequest struct {
	ID        string `json:"id"`
	PID       string `json:"product_id"`
	IsMain    bool   `json:"is_main"`
	URL       string `json:"url"`
	Image     string `json:"image"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	SortOrder int    `json:"sort_order"`
}

type ProductEntityResponse struct {
	PidEid     uint      `json:"product_entity_id"`
	Entity     string    `json:"entity_name"`
	EntityType string    `json:"entity_type"`
	CreatedAt  time.Time `json:"created_at"`
}
