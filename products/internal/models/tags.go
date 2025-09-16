package models

import (
	"e-commerce-backend/products/pkg/payloads"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// EntityType constants
const (
	EntityTypeTag      = "tag"
	EntityTypeCategory = "category"
	EntityTypeBrand    = "brand"
)

func CheckLabelExists(db *gorm.DB, label string, labelType string) error {
	var entity Entities
	return db.Where("entity = ? AND entity_type = ?", strings.ToLower(label), labelType).First(&entity).Error
}

func CreateEntity(db *gorm.DB, entity string, entityType string) (uint, error) {
	newEntity := Entities{
		EntityName: strings.ToLower(entity),
		EntityType: entityType,
	}
	if err := db.Create(&newEntity).Error; err != nil {
		return 0, fmt.Errorf("failed to create entity: %v", err)
	}
	return newEntity.EID, nil
}

func StoreProductEntity(db *gorm.DB, entityID uint, entityType string, productID string) (ProductEntity, error) {
	productEntity := ProductEntity{
		EntityID:   entityID,
		ProductID:  productID,
		EntityType: entityType,
	}
	if err := db.Create(&productEntity).Error; err != nil {
		return ProductEntity{}, fmt.Errorf("failed to create product entity: %v", err)
	}
	return productEntity, nil
}

func CheckAndCreateTags(db *gorm.DB, tags []string, productID string) ([]ProductEntity, error) {
	var tagsResp []ProductEntity

	for _, tagName := range tags {
		tagName = strings.ToLower(strings.TrimSpace(tagName))
		if tagName == "" {
			continue
		}

		// Try to find existing tag
		var tag Entities
		err := db.Where("entity = ?", tagName).First(&tag).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				tag = Entities{EntityName: tagName, EntityType: "tag"}
				if err := db.Create(&tag).Error; err != nil {
					return nil, fmt.Errorf("failed to create tag %s: %v", tagName, err)
				}
			} else {
				return nil, fmt.Errorf("error checking tag %s: %v", tagName, err)
			}
		} else {
			if tag.EntityType != "tag" {
				return nil, fmt.Errorf("tag %s is not a tag", tagName)
			}
		}

		// Now insert into the ProductEntity table using tag.EntityID
		productTagEntity := ProductEntity{
			EntityID:   tag.EID,
			EntityType: "tag",
			ProductID:  productID,
		}
		err = db.Create(&productTagEntity).Error
		if err != nil {
			return nil, fmt.Errorf("failed to link tag %s to product: %v", tagName, err)
		}

		tagsResp = append(tagsResp, productTagEntity)
	}

	return tagsResp, nil
}

func CheckAndCreateCategoryOrBrand(db *gorm.DB, entity string, entityType string, productID string) (*ProductEntity, error) {
	if entityType != EntityTypeCategory && entityType != EntityTypeBrand {
		return nil, fmt.Errorf("invalid entity type: %s", entityType)
	}

	entity = strings.ToLower(strings.TrimSpace(entity))
	if entity == "" {
		return nil, nil
	}

	// Try to find existing entity
	var existingEntity Entities
	err := db.Where("entity = ? AND entity_type = ?", entity, entityType).First(&existingEntity).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create new entity
			entityID, err := CreateEntity(db, entity, entityType)
			if err != nil {
				return nil, fmt.Errorf("failed to create %s %s: %v", entityType, entity, err)
			}
			existingEntity.EID = entityID
			existingEntity.EntityName = entity
		} else {
			return nil, fmt.Errorf("error checking %s %s: %v", entityType, entity, err)
		}
	}

	// Create or update product-entity relationship
	productEntity, err := StoreProductEntity(db, existingEntity.EID, entityType, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to link %s %s to product: %v", entityType, entity, err)
	}

	return &productEntity, nil
}

func GetProductEntities(db *gorm.DB, productID string, entityType string) ([]payloads.EntityLabelResponse, error) {
	var entities []struct {
		ProductEntityID uint
		Entity          string
		EntityType      string
	}

	err := db.Table("product_entities").
		Select("product_entities.product_entity_id, entities.entity, entities.entity_type").
		Joins("JOIN entities ON entities.entity_id = product_entities.entity_id").
		Where("product_entities.product_id = ? AND product_entities.entity_type = ?", productID, entityType).
		Scan(&entities).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch entities: %v", err)
	}

	var response []payloads.EntityLabelResponse
	for _, e := range entities {
		response = append(response, payloads.EntityLabelResponse{
			ProductEntityID: e.ProductEntityID,
			Entity:          e.Entity,
			EntityType:      e.EntityType,
		})
	}

	return response, nil
}
