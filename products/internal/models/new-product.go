// Package models
package models

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel includes soft delete and timestamps
type BaseModel struct {
	CreatedAt time.Time      `json:"created_at" gorm:"type:datetime;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"type:datetime;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// Product is the main product model
type Product struct {
	ID            string  `json:"id" gorm:"primaryKey;type:varchar(36)"`
	SellerID      string  `json:"seller_id" gorm:"not null;type:varchar(36)"`
	SKU           string  `json:"sku" validate:"required" gorm:"uniqueIndex;type:varchar(255)"`
	Name          string  `json:"name" validate:"required,min=3,max=50"`
	Description   string  `json:"description" validate:"required,min=10,max=500"`
	ShortDesc     string  `json:"short_desc" validate:"required,max=200"`
	SlugURL       string  `json:"slug_url" gorm:"uniqueIndex;type:varchar(255)"`
	MainImage     string  `json:"main_image" validate:"omitempty,url"`
	Price         float64 `json:"price" validate:"required,gt=0"`
	ComparePrice  float64 `json:"compare_price,omitempty" validate:"omitempty,gte=0"`
	CostPrice     float64 `json:"cost_price,omitempty" validate:"omitempty,gte=0"`
	Tax           float64 `json:"tax,omitempty" validate:"omitempty,gte=0"`
	Discount      float64 `json:"discount,omitempty" validate:"omitempty,gte=0"`
	DiscountType  string  `json:"discount_type,omitempty" validate:"omitempty,oneof=percentage fixed"`
	Quantity      int     `json:"quantity" validate:"required,gt=0"`
	IsActive      bool    `json:"is_active" gorm:"default:true"`
	IsFeatured    bool    `json:"is_featured" gorm:"default:false"`
	Rating        float64 `json:"rating,omitempty" validate:"omitempty,gte=0"`
	Weight        float64 `json:"weight,omitempty" validate:"omitempty,gte=0"`
	Length        float64 `json:"length,omitempty" validate:"omitempty,gte=0"`
	Width         float64 `json:"width,omitempty" validate:"omitempty,gte=0"`
	Height        float64 `json:"height,omitempty" validate:"omitempty,gte=0"`
	DimensionUnit string  `json:"dimension_unit,omitempty" validate:"omitempty,oneof=cm inch meter"`
	MetaTitle     string  `json:"meta_title,omitempty" validate:"omitempty,max=100"`
	MetaDesc      string  `json:"meta_description,omitempty" validate:"omitempty,max=300"`

	Entities Entities `json:"category" gorm:"foreignKey:ProductID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	// Images     []ProductImage     `json:"images" gorm:"foreignKey:PID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	// Variants   []ProductVariant   `json:"variants" gorm:"foreignKey:PID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	// Attributes []ProductAttribute `json:"attributes" gorm:"foreignKey:PID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`

	BaseModel
}

// Entities links a product to an external entity like brand, category, or tag
type Entities struct {
	ID         uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	ProductID  string `json:"product_id" gorm:"type:varchar(36);index"`
	EntityType string `json:"etype" gorm:"type:varchar(50)"`
	EntityName string `json:"ename" gorm:"type:varchar(255)"`
	BaseModel
}

// ProductImage represents images associated with a product
type ProductImage struct {
	ID        string `json:"id" gorm:"primaryKey;type:varchar(36)"`
	PID       string `json:"product_id" gorm:"index;type:varchar(36)"`
	URL       string `json:"url" validate:"required,url"`
	IsMain    bool   `json:"is_main"`
	SortOrder int    `json:"sort_order" validate:"gte=0"`
	AltText   string `json:"alt_text,omitempty"`

	BaseModel
}

// ProductVariant represents a product variation (e.g., size, color)
type ProductVariant struct {
	ID       string          `json:"id" gorm:"primaryKey;type:varchar(36)"`
	PID      string          `json:"product_id" gorm:"index;type:varchar(36)"`
	SKU      string          `json:"sku" gorm:"type:varchar(255)"`
	Name     string          `json:"name" gorm:"type:varchar(255)"`
	Price    float64         `json:"price" validate:"required,gt=0"`
	Quantity int             `json:"quantity" validate:"required,gt=0"`
	Options  []VariantOption `json:"options" gorm:"foreignKey:VariantID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

// VariantOption stores name-value pair for a variant (e.g., Size: M)
type VariantOption struct {
	ID          string `json:"id" gorm:"primaryKey;type:varchar(36)"`
	VariantID   string `json:"variant_id" gorm:"index;type:varchar(36)"`
	OptionName  string `json:"option_name" validate:"required"`
	OptionValue string `json:"option_value" validate:"required"`
}

// ProductAttribute represents custom attributes (e.g., Material, Origin)
type ProductAttribute struct {
	ID    string `json:"id" gorm:"primaryKey;type:varchar(36)"`
	PID   string `json:"product_id" gorm:"index;type:varchar(36)"`
	Name  string `json:"name" validate:"required"`
	Value string `json:"value" validate:"required"`
}
