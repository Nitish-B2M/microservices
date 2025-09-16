// Package models for product
package models

import (
	"time"

	"gorm.io/datatypes"
)

type Product struct {
	ID            string             `json:"id" gorm:"primaryKey"`
	SellerID      string             `json:"seller_id" gorm:"not null"`
	SKU           string             `json:"sku" validate:"required" gorm:"uniqueIndex;type:varchar(255)"`
	Name          string             `json:"name" validate:"required,min=3,max=50"`
	Description   string             `json:"description" validate:"required,min=10,max=500"`
	ShortDesc     string             `json:"short_desc" validate:"required,max=200"`
	Price         float64            `json:"price" validate:"required,gt=0"`
	ComparePrice  float64            `json:"compare_price" validate:"omitempty,gte=0"`
	CostPrice     float64            `json:"cost_price" validate:"omitempty,gte=0"`
	Quantity      int                `json:"quantity" validate:"required,gt=0"`
	InStock       bool               `json:"in_stock" gorm:"default:true"`
	IsDeleted     bool               `json:"is_deleted" gorm:"default:false"`
	IsActive      bool               `json:"is_active" gorm:"default:true"`
	IsFeatured    bool               `json:"is_featured" gorm:"default:false"`
	Rating        float64            `json:"rating" validate:"omitempty,gte=0"`
	Tax           float64            `json:"tax" validate:"omitempty,gte=0"`
	Discount      float64            `json:"discount" validate:"omitempty,gte=0"`
	DiscountType  string             `json:"discount_type" validate:"omitempty,oneof=percentage fixed"`
	Weight        float64            `json:"weight" validate:"omitempty,gte=0"`
	Length        float64            `json:"length" validate:"omitempty,gte=0"`
	Width         float64            `json:"width" validate:"omitempty,gte=0"`
	Height        float64            `json:"height" validate:"omitempty,gte=0"`
	DimensionUnit string             `json:"dimension_unit" validate:"omitempty,oneof=cm inch meter"`
	StockStatus   string             `json:"stock_status" validate:"omitempty,oneof=in_stock out_of_stock preorder"`
	MetaTitle     string             `json:"meta_title" validate:"omitempty,max=100"`
	MetaDesc      string             `json:"meta_description" validate:"omitempty,max=300"`
	SlugURL       string             `json:"slug_url" gorm:"uniqueIndex;type:varchar(255)"`
	MainImage     string             `json:"main_image" validate:"omitempty,url"`
	CreatedAt     time.Time          `json:"created_at" gorm:"type:datetime;default:CURRENT_TIMESTAMP"`
	UpdatedAt     time.Time          `json:"updated_at" gorm:"type:datetime;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`
	DeletedAt     *time.Time         `json:"deleted_at" gorm:"index"`
	Category      ProductEntity      `json:"category" gorm:"foreignKey:ProductID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Brand         ProductEntity      `json:"brand" gorm:"foreignKey:ProductID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Tags          []ProductEntity    `json:"tags" gorm:"foreignKey:ProductID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Images        []ProductImage     `json:"images" gorm:"foreignKey:PID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Variants      []ProductVariant   `json:"variants" gorm:"foreignKey:PID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Attributes    []ProductAttribute `json:"attributes" gorm:"foreignKey:PID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type ProductEntity struct {
	ID         uint      `json:"product_entity_id" gorm:"primaryKey;autoIncrement"`
	EntityID   uint      `json:"-"`
	ProductID  string    `json:"-" gorm:"type:varchar(36);index"`
	EntityType string    `json:"-" gorm:"type:varchar(50)"`
	CreatedAt  time.Time `json:"-" gorm:"type:datetime;default:CURRENT_TIMESTAMP"`
	Entity     Entities  `json:"entity" gorm:"foreignKey:EntityID"`
}

type Entities struct {
	EID        uint      `json:"entity_id" gorm:"primaryKey;autoIncrement"`
	EntityName string    `json:"entity_name" gorm:"type:varchar(255);uniqueIndex:idx_entity_type"`
	EntityType string    `json:"entity_type" gorm:"type:varchar(50);uniqueIndex:idx_entity_type"`
	CreatedAt  time.Time `json:"created_at" gorm:"type:datetime;default:CURRENT_TIMESTAMP"`
}

type ProductAttribute struct {
	ID    string `json:"id" gorm:"primaryKey"`
	PID   string `json:"product_id" gorm:"index;type:varchar(36)"`
	Name  string `json:"name" validate:"required"`
	Value string `json:"value" validate:"required"`
}

type ProductImage struct {
	ID        string `json:"id" gorm:"primaryKey"`
	PID       string `json:"product_id" gorm:"index;type:varchar(36)"`
	URL       string `json:"url" validate:"required,url"`
	IsMain    bool   `json:"is_main"`
	SortOrder int    `json:"sort_order" validate:"gte=0"`
}

type ProductVariant struct {
	ID       string         `json:"id" gorm:"primaryKey;"`
	PID      string         `json:"product_id" gorm:"index;type:varchar(36)"`
	SKU      string         `json:"sku" gorm:"type:varchar(255)"`
	Name     string         `json:"name" gorm:"type:varchar(255)"`
	Price    float64        `json:"price" validate:"required,gt=0"`
	Quantity int            `json:"quantity" validate:"required,gt=0"`
	Options  datatypes.JSON `json:"options" gorm:"type:json" validate:"omitempty"` // Stored as JSON string
}
