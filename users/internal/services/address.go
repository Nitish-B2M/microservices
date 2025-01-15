package services

import (
	"e-commerce-backend/shared/utils"
	"e-commerce-backend/users/internal/models"
	"encoding/json"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"strings"
)

type AdrServices struct {
	DB *gorm.DB
}

func NewAdrServices(db *gorm.DB) *AdrServices {
	return &AdrServices{db}
}

type AddressServices interface {
	AddAddress(w http.ResponseWriter, r *http.Request)
	GetAddressByUserId(w http.ResponseWriter, r *http.Request)
	UpdateAddress(w http.ResponseWriter, r *http.Request)
	DeleteAddress(w http.ResponseWriter, r *http.Request)
	SetPrimaryAddress(w http.ResponseWriter, r *http.Request)
}

func toLowerCaseData(adr models.Address) models.Address {
	adr.FullName = strings.ToLower(adr.FullName)
	adr.Area = strings.ToLower(adr.Area)
	adr.City = strings.ToLower(adr.City)
	adr.State = strings.ToLower(adr.State)
	adr.Street = strings.ToLower(adr.Street)
	adr.Country = strings.ToLower(adr.Country)
	adr.PostalCode = strings.ToLower(adr.PostalCode)
	return adr
}

func toTitleCaseData(adr models.Address) models.Address {
	adr.FullName = strings.Title(adr.FullName)
	adr.Area = strings.Title(adr.Area)
	adr.City = strings.Title(adr.City)
	adr.State = strings.Title(adr.State)
	adr.Street = strings.Title(adr.Street)
	adr.Country = strings.Title(adr.Country)
	adr.PostalCode = strings.Title(adr.PostalCode)
	return adr
}

func validateAddress(address models.Address) []error {
	var err []error

	if address.FullName == "" {
		if strings.TrimSpace(address.FullName) == "" {
			err = append(err, errors.New("full name is required"))
		}
	}
	if len(address.FullName) > 40 || len(address.FullName) <= 2 {
		if strings.TrimSpace(address.FullName) == "" {
			err = append(err, errors.New("full name is required"))
		} else {
			err = append(err, errors.New("full name must be between 2 and 40 characters"))
		}
	}
	if address.Area == "" || address.City == "" || address.State == "" {
		if strings.TrimSpace(address.Area) == "" || strings.TrimSpace(address.City) == "" || strings.TrimSpace(address.State) == "" {
			err = append(err, errors.New("area or city or state are required"))
		}
	}
	if address.Street == "" || address.PostalCode == "" || address.Country == "" {
		if strings.TrimSpace(address.Street) == "" || strings.TrimSpace(address.PostalCode) == "" || strings.TrimSpace(address.Country) == "" {
			err = append(err, errors.New("street or zip or country are required"))
		}
	}
	if address.AddressType == "" {
		if strings.TrimSpace(address.AddressType) == "" {
			err = append(err, errors.New("address type are required"))
		}
	}
	postalCode := strings.TrimSpace(address.PostalCode)
	if len(postalCode) != 6 {
		err = append(err, errors.New("postal_code must be length of 6"))
	}
	_, er := strconv.Atoi(postalCode)
	if er != nil {
		err = append(err, errors.New("postal code must be a number"))
	}

	if address.Phone == "" {
		if strings.TrimSpace(address.Phone) == "" {
			err = append(err, errors.New("phone number is required"))
		}
	}
	phone := strings.TrimSpace(address.Phone)
	if len(phone) != 10 {
		err = append(err, errors.New("phone number must be length of 10"))
	}
	_, er = strconv.Atoi(phone)
	if er != nil {
		err = append(err, errors.New("phone number must be a number"))
	}

	return err
}

func (db *AdrServices) GetAddressByUserId(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value(utils.UserIDKey).(int)
	if !ok {
		utils.JsonError(w, "invalid user", http.StatusBadRequest, nil)
		return
	}

	var adr models.Address
	addresses, err := adr.GetAddressByUserId(db.DB, userId)
	if err != nil {
		utils.JsonError(w, err.Error(), http.StatusBadRequest, nil)
		return
	}

	for i := range addresses {
		addresses[i] = toTitleCaseData(addresses[i])
	}

	utils.JsonResponse(addresses, w, "address fetched successfully", http.StatusOK)
}

func (db *AdrServices) AddAddress(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value(utils.UserIDKey).(int)
	if !ok {
		utils.JsonError(w, "invalid user", http.StatusBadRequest, nil)
		return
	}

	var address models.Address
	if err := json.NewDecoder(r.Body).Decode(&address); err != nil {
		utils.JsonError(w, err.Error(), http.StatusBadRequest, err)
		return
	}

	errs := validateAddress(address)
	if len(errs) > 0 {
		errStr := ""
		for i, err := range errs {
			errStr += err.Error()
			if i != len(errs)-1 {
				errStr += ", "
			}
		}
		utils.JsonErrorWithExtra(w, "invalid address", http.StatusBadRequest, fmt.Errorf(errStr), errs)
		return
	}
	address = toLowerCaseData(address)

	var temp models.Address
	existingAdr, err := temp.GetAddressByUserId(db.DB, userId)
	if err != nil {
		utils.JsonError(w, err.Error(), http.StatusBadRequest, err)
		return
	}

	if len(existingAdr) == 0 {
		address.IsPrimary = true
	} else {
		for _, adr := range existingAdr {
			if strings.EqualFold(adr.AddressType, address.AddressType) {
				utils.JsonError(w, fmt.Sprintf("'%s' address type is already exists", address.AddressType), http.StatusBadRequest, err)
				return
			}
			if address.IsPrimary {
				adr.IsPrimary = false
				if err := adr.UpdateAddress(db.DB, userId); err != nil {
					utils.JsonError(w, "failed to update address", http.StatusBadRequest, err)
					return
				}
			}
		}
	}

	address.UserId = userId
	if err := address.CreateAddress(db.DB); err != nil {
		utils.JsonError(w, "address add failed", http.StatusBadRequest, err)
		return
	}
	address = toTitleCaseData(address)

	utils.JsonResponse(address, w, "Address added successfully", http.StatusOK)
}

func (db *AdrServices) UpdateAddress(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value(utils.UserIDKey).(int)
	if !ok {
		utils.JsonError(w, "invalid user", http.StatusBadRequest, nil)
		return
	}

	adrId, err := utils.GetIDFromPath(r)
	if err != nil {
		utils.JsonError(w, err.Error(), http.StatusBadRequest, nil)
		return
	}

	var partialAddress models.Address
	if err := json.NewDecoder(r.Body).Decode(&partialAddress); err != nil {
		utils.JsonError(w, fmt.Sprintf("Failed to parse address: %v", err), http.StatusBadRequest, err)
		return
	}

	var address models.Address
	if err := address.GetAddressById(db.DB, adrId); err != nil {
		utils.JsonError(w, err.Error(), http.StatusBadRequest, nil)
		return
	}

	// Only update fields that are non-zero (i.e., if they were provided)
	errs := validateAddress(partialAddress)
	if len(errs) > 0 {
		errStr := ""
		for i, err := range errs {
			errStr += err.Error()
			if i != len(errs)-1 {
				errStr += ", "
			}
		}
		utils.JsonErrorWithExtra(w, "invalid address", http.StatusBadRequest, fmt.Errorf(errStr), errs)
		return
	}
	partialAddress = toLowerCaseData(partialAddress)
	if partialAddress.FullName != address.FullName {
		address.FullName = partialAddress.FullName
	}
	if partialAddress.Street != address.Street {
		address.Street = partialAddress.Street
	}
	if partialAddress.Area != address.Area {
		address.Area = partialAddress.Area
	}
	if partialAddress.City != address.City {
		address.City = partialAddress.City
	}
	if partialAddress.State != address.State {
		address.State = partialAddress.State
	}
	if partialAddress.Country != address.Country {
		address.Country = partialAddress.Country
	}
	if partialAddress.PostalCode != address.PostalCode {
		address.PostalCode = partialAddress.PostalCode
	}
	if partialAddress.AddressType != "" {
		if partialAddress.AddressType == "other" {
			if partialAddress.OtherAddressType == "" {
				utils.JsonError(w, "other address type is required", http.StatusBadRequest, nil)
				return
			}
			partialAddress.AddressType = strings.ToLower(partialAddress.OtherAddressType)
		}
		address.AddressType = strings.ToLower(partialAddress.AddressType)
	}
	if partialAddress.Phone != address.Phone {
		address.Phone = partialAddress.Phone
	}

	if err := address.UpdateAddress(db.DB, userId); err != nil {
		utils.JsonError(w, "failed to update address", http.StatusBadRequest, err)
		return
	}

	if address.IsPrimary != partialAddress.IsPrimary {
		var temp models.Address
		addresses, err := temp.GetAddressByUserId(db.DB, userId)
		if err != nil {
			utils.JsonError(w, err.Error(), http.StatusBadRequest, nil)
			return
		}
		if len(addresses) == 0 {
			utils.JsonError(w, "address is empty", http.StatusBadRequest, nil)
			return
		}
		if len(addresses) == 1 {
			if partialAddress.IsPrimary {
				addresses[0].IsPrimary = true
				if err := addresses[0].UpdateAddress(db.DB, userId); err != nil {
					utils.JsonError(w, "failed to update address", http.StatusBadRequest, err)
					return
				}
			} else {
				utils.JsonResponse(addresses, w, "you can't make is address as not primary", http.StatusBadRequest)
				return
			}
		} else {
			if partialAddress.IsPrimary {
				for _, a := range addresses {
					if a.Id != adrId {
						a.IsPrimary = !partialAddress.IsPrimary
					} else {
						a.IsPrimary = partialAddress.IsPrimary
					}
					if err := a.UpdateAddress(db.DB, userId); err != nil {
						utils.JsonError(w, "failed to update address", http.StatusBadRequest, err)
						return
					}
				}
			} else {
				for _, a := range addresses {
					if a.Id != adrId {
						a.IsPrimary = !partialAddress.IsPrimary
						if err := a.UpdateAddress(db.DB, userId); err != nil {
							utils.JsonError(w, "failed to update address", http.StatusBadRequest, err)
							return
						}
						break
					}
				}
				address.IsPrimary = partialAddress.IsPrimary
				if err := address.UpdateAddress(db.DB, userId); err != nil {
					utils.JsonError(w, "failed to update address", http.StatusBadRequest, err)
					return
				}
			}
		}
	}

	if err := address.GetAddressById(db.DB, adrId); err != nil {
		utils.JsonError(w, err.Error(), http.StatusBadRequest, nil)
		return
	}
	address = toTitleCaseData(address)

	utils.JsonResponse(address, w, "Address updated successfully", http.StatusOK)
}

func (db *AdrServices) SetPrimaryAddress(w http.ResponseWriter, r *http.Request) {
	userId := utils.GetUserIdFromContext(r)

	adrId, err := utils.GetIDFromPath(r)
	if err != nil {
		utils.JsonError(w, err.Error(), http.StatusBadRequest, nil)
		return
	}

	var adr models.Address
	if err := adr.GetAddressById(db.DB, adrId); err != nil {
		utils.JsonError(w, err.Error(), http.StatusBadRequest, nil)
		return
	}

	if err := adr.SetPrimaryAddress(db.DB, userId); err != nil {
		utils.JsonError(w, "failed to update address", http.StatusInternalServerError, nil)
		return
	}
	adr = toTitleCaseData(adr)

	utils.JsonResponse(adr, w, "Address set as primary successfully", http.StatusOK)
}

func (db *AdrServices) DeleteAddress(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value(utils.UserIDKey).(int)
	if !ok {
		utils.JsonError(w, "invalid user", http.StatusBadRequest, nil)
		return
	}

	adrId, err := utils.GetIDFromPath(r)
	if err != nil {
		utils.JsonError(w, err.Error(), http.StatusBadRequest, nil)
		return
	}

	var adr models.Address
	if err := adr.GetAddressById(db.DB, adrId); err != nil {
		utils.JsonError(w, err.Error(), http.StatusBadRequest, nil)
		return
	}

	if adr.IsPrimary {
		addresses, err := adr.GetAddressByUserId(db.DB, userId)
		if err != nil {
			utils.JsonError(w, err.Error(), http.StatusBadRequest, nil)
			return
		}
		if len(addresses) == 0 {
			utils.JsonError(w, "address is empty", http.StatusBadRequest, nil)
			return
		}
		if len(addresses) == 1 {
		} else {
			for _, a := range addresses {
				if a.Id != adrId {
					a.IsPrimary = true
					if err := a.UpdateAddress(db.DB, userId); err != nil {
						utils.JsonError(w, "failed to update address", http.StatusBadRequest, err)
						return
					}
					break
				}
			}
		}
	}

	if err := adr.DeleteAddress(db.DB, adrId, userId); err != nil {
		utils.JsonError(w, err.Error(), http.StatusBadRequest, nil)
		return
	}

	utils.JsonResponse(nil, w, "Address deleted successfully", http.StatusOK)
}
