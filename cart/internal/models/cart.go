package models

import (
	"e-commerce-backend/cart/dbs"
	"e-commerce-backend/shared/utils"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
)

type Cart struct {
	Id          int       `json:"id" gorm:"autoIncrement"`
	UserId      int       `json:"user_id"`
	ProductId   int       `json:"product_id"`
	Quantity    int       `json:"quantity" default:"1"`
	IsProcessed bool      `json:"is_processed" default:"false"`
	CreatedAt   time.Time `json:"-" gorm:"type:datetime;default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"type:datetime;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`
}

func InitCartSchema() {
	db := dbs.DB
	if err := db.AutoMigrate(&Cart{}); err != nil {
		log.Fatalf(utils.DatabaseMigrationError, "Cart", err)
	} else {
		log.Printf(utils.SchemaMigrationSuccess, "Cart")
	}
}

func (c *Cart) NewCart() *Cart {
	return &Cart{
		UserId:      c.UserId,
		ProductId:   c.ProductId,
		Quantity:    c.Quantity,
		IsProcessed: false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

var cartMutex sync.Mutex

type CartService interface {
	AddToCart(db *gorm.DB) (Cart, error)
	GetCartByUserId(db *gorm.DB, userId int) ([]Cart, error)
	RemoveItemFromCart(db *gorm.DB, cartId, quantity int) error
	GetCartByCartId(db *gorm.DB, cartId int) error
}

func (c *Cart) AddToCart(db *gorm.DB) (Cart, error) {
	//avoid race condition
	cartMutex.Lock()
	defer cartMutex.Unlock()
	c.IsProcessed = false

	var cart Cart
	if err := db.First(&cart, "user_id =? and product_id =?", c.UserId, c.ProductId).Error; err != nil {
		if strings.EqualFold(err.Error(), gorm.ErrRecordNotFound.Error()) {
			if c.Quantity == 0 {
				c.Quantity = 1
			}
			if err := db.Create(&c).Error; err != nil {
				cart = *c
				return cart, fmt.Errorf("failed to create cart: %v", err)
			}
			cart = *c
			return cart, nil
		}
		return cart, err
	}

	if (cart.UserId == c.UserId) && (cart.ProductId == c.ProductId) {
		cart.Quantity = cart.Quantity + c.Quantity
		if err := cart.UpdateCart(db); err != nil {
			return cart, err
		}
		return cart, nil
	}

	if err := db.Create(&cart).Error; err != nil {
		return cart, fmt.Errorf("failed to create cart: %v", err)
	}
	return cart, nil
}

func (c *Cart) UpdateCart(db *gorm.DB) error {
	if err := db.Model(&c).Where("id=?", c.Id).Update("quantity", c.Quantity).Error; err != nil {
		return err
	}
	return nil
}

func (c *Cart) GetCartByUserId(db *gorm.DB, userId int) ([]Cart, error) {
	var cart []Cart
	if err := db.Where("user_id =? and is_processed = false", userId).Find(&cart).Error; err != nil {
		return nil, err
	}
	return cart, nil
}

func (c *Cart) GetCartItemByCartAndUserId(db *gorm.DB, userId int, cartId int) error {
	if err := db.Where("id =? and user_id =? and is_processed = false", cartId, userId).First(&c).Error; err != nil {
		return err
	}
	return nil
}

func (c *Cart) GetCartByCartId(db *gorm.DB, cartId int) error {
	if err := db.Where("id=? and is_processed = false", cartId).First(&c).Error; err != nil {
		if strings.EqualFold(err.Error(), gorm.ErrRecordNotFound.Error()) {
			return fmt.Errorf(utils.CartItemNotFoundError, cartId)
		}
		return err
	}
	return nil
}

func (c *Cart) DeleteCartItem(db *gorm.DB) error {
	if err := db.Where("id =? and is_processed = false", c.Id).Delete(&c).Error; err != nil {
		return err
	}
	return nil
}
