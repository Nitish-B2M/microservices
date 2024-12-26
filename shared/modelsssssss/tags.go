package models

import (
	"errors"
	"fmt"
	"log"
	"microservices/pkg/shared/dbs"
	"microservices/pkg/shared/payloads"
	"microservices/pkg/shared/utils"

	"gorm.io/gorm"
)

type Tag struct {
	ID   int    `json:"id"`
	Name string `json:"name" gorm:"unique;not null"`
}

func NewTag() *Tag {
	return &Tag{}
}

func InitTagSchema() {
	db := dbs.DB
	if err := db.AutoMigrate(&Tag{}); err != nil {
		log.Fatalf(utils.DatabaseMigrationError, "Tag", err)
	} else {
		log.Printf(utils.SchemaMigrationSuccess, "Tag")
	}
}

func CreateTag(db *gorm.DB, tagName string) (int, error) {
	tag := Tag{Name: tagName}
	if err := db.Create(&tag).Error; err != nil {
		return 0, fmt.Errorf("failed to create tag: %v", err)
	}
	return tag.ID, nil
}

func FetchTagById(db *gorm.DB, tagId int) (*Tag, error) {
	var tag Tag
	if err := db.First(&tag, tagId).Error; err != nil {
		return nil, err
	}
	return &tag, nil
}

func CheckAndCreateProductTags(db *gorm.DB, newProduct payloads.ProductRequest) ([]int, map[string]error) {
	tagErrors := make(map[string]error)
	var tagIds []int

	if len(newProduct.Tags) == 0 || newProduct.Tags == nil {
		return tagIds, tagErrors
	}

	for i := range newProduct.Tags {
		var tag Tag
		tagName := newProduct.Tags[i]
		if err := db.Where("name = ?", tagName).First(&tag).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				utils.SimpleLog("info", fmt.Sprintf(utils.TagNotExist, tagName))
				id, err := CreateTag(db, tagName)
				if err != nil {
					utils.SimpleLog("error", err.Error())
					tagErrors[tagName] = fmt.Errorf(utils.TagCreationFailed, err)
				}
				tagIds = append(tagIds, id)
			} else {
				tagErrors[tagName] = fmt.Errorf(utils.TagExistError, err)
			}
		} else {
			tagIds = append(tagIds, tag.ID)
		}
	}
	return tagIds, tagErrors
}

func AddTagToProduct(db *gorm.DB, tagIds []int, productID int) []error {
	var errs []error
	for _, tagID := range tagIds {
		productTag := ProductTag{
			ProductID: productID,
			TagID:     tagID,
		}
		if err := db.Create(&productTag).Error; err != nil {
			errs = append(errs, fmt.Errorf("failed to create product[%d]-tag[%d] association: %v", tagID, productID, err))
		}
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

func RemoveTagToProduct(db *gorm.DB, tagIds []int, productID int) []error {
	if err := db.Where("product_id = ? AND tag_id IN ?", productID, tagIds).Delete(&ProductTag{}).Error; err != nil {
		return []error{fmt.Errorf("failed to delete product-tag associations: %v", err)}
	}

	return nil
}

func UpdateTagToProduct(db *gorm.DB, newTagIds []int, productID int) []error {
	var errs []error
	var productTags []ProductTag

	if err := db.Where("product_id =?", productID).Find(&productTags).Error; err != nil {

	}

	tagLookup := make(map[int]bool)
	for _, tagID := range newTagIds {
		tagLookup[tagID] = true
	}

	var addTags []int
	var removeTags []int
	for _, pTag := range productTags {
		if tagLookup[pTag.TagID] {
			continue
		} else {
			removeTags = append(removeTags, pTag.TagID)
			tagLookup[pTag.TagID] = true
		}
	}

	for _, tagID := range newTagIds {
		tagFound := false
		for _, pTag := range productTags {
			if pTag.TagID == tagID {
				tagFound = true
				break
			}
		}

		if !tagFound {
			addTags = append(addTags, tagID)
		}
	}

	utils.SimpleLog("info", "tags to remove:", removeTags)
	utils.SimpleLog("info", "tags to add:", addTags)

	if len(addTags) > 0 {
		err := AddTagToProduct(db, addTags, productID)
		errs = append(errs, err...)
	}

	if len(removeTags) > 0 {
		err := RemoveTagToProduct(db, removeTags, productID)
		errs = append(errs, err...)
	}

	return errs
}
