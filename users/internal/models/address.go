package models

import (
	"e-commerce-backend/shared/utils"
	"e-commerce-backend/users/dbs"
	"log"
	"time"

	"gorm.io/gorm"
)

type Address struct {
	Id               int       `json:"address_id" gorm:"primaryKey;autoIncrement"`
	UserId           int       `json:"user_id" gorm:"not null"`
	FullName         string    `json:"full_name" gorm:"not null;type:varchar(100)"`
	Street           string    `json:"street" gorm:"not null;type:varchar(100)"`
	Area             string    `json:"area" gorm:"not null;type:varchar(60)"`
	City             string    `json:"city" gorm:"not null;type:varchar(100)"`
	State            string    `json:"state" gorm:"not null;type:varchar(25)"`
	PostalCode       string    `json:"postal_code" gorm:"not null;type:varchar(6)"`
	Country          string    `json:"country" gorm:"not null"`
	IsPrimary        bool      `json:"is_primary" gorm:"default:false"`
	Phone            string    `json:"phone_number" gorm:"not null;type:varchar(20)"`
	AddressType      string    `json:"address_type" gorm:"type:varchar(20)"`
	OtherAddressType string    `json:"address_type_other" gorm:"type:varchar(100)"`
	CreatedAt        time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt        time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func InitAddressSchemas() {
	db := dbs.UserDB
	if err := db.AutoMigrate(&Address{}); err != nil {
		log.Printf(utils.DatabaseMigrationError, "Address", err.Error())
	} else {
		log.Printf(utils.SchemaMigrationSuccess, "Address")
	}
}

type AddressInterface interface {
	CreateAddress(db *gorm.DB) error
	UpdateAddress(db *gorm.DB, userId int) error
	GetAddressByUserId(db *gorm.DB, userId int) ([]Address, error)
	GetAddressById(db *gorm.DB, ardId int) error
}

func (adr *Address) GetPrimaryAddress(db *gorm.DB, userId int) error {
	return db.Where("user_id =? and is_primary =?", userId, true).First(adr).Error
}

func (adr *Address) GetAddressByUserId(db *gorm.DB, userId int) ([]Address, error) {
	var adrs []Address
	if err := db.Model(adr).Where("user_id =?", userId).Find(&adrs).Error; err != nil {
		return nil, err
	}
	return adrs, nil
}

func (adr *Address) GetAddressById(db *gorm.DB, adrId int) error {
	return db.Where("id=?", adrId).First(&adr).Error
}

func (adr *Address) CreateAddress(db *gorm.DB) error {
	if err := db.Create(&adr).Error; err != nil {
		return err
	}
	return nil
}

func (adr *Address) UpdateAddress(db *gorm.DB, userId int) error {
	return db.Where("user_id =? and id =?", userId, adr.Id).Save(adr).Error
}

func (adr *Address) DeleteAddress(db *gorm.DB, adrId, userId int) error {
	return db.Where("user_id =? and id =?", userId, adrId).Delete(adr).Error
}

func (adr *Address) SetPrimaryAddress(db *gorm.DB, userId int) error {
	addresses, _ := adr.GetAddressByUserId(db, userId)
	for _, a := range addresses {
		if a.Id == adr.Id {
			a.IsPrimary = true
		} else {
			a.IsPrimary = false
		}
		if err := a.UpdateAddress(db, userId); err != nil {
			return err
		}
	}
	return nil
}
