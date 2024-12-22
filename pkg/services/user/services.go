package user

import (
	"encoding/json"
	"fmt"
	"microservices/pkg/shared/models"
	"microservices/pkg/shared/payloads"
	"microservices/pkg/shared/utils"
	"net/http"
	"reflect"
	"strings"

	"gorm.io/gorm"
)

type UserService struct {
	DB *gorm.DB
}

type User interface {
	AddUser(w http.ResponseWriter, r *http.Request)
	UpdateUser(w http.ResponseWriter, r *http.Request)
	DeleteUser(w http.ResponseWriter, r *http.Request)
	GetAllUsers(w http.ResponseWriter, r *http.Request)
	FetchUserById(w http.ResponseWriter, r *http.Request)
}

func validateAddUserRequest(w http.ResponseWriter, data models.User) bool {
	var errorMessages []string

	// Collect validation errors
	if data.FirstName == "" {
		errorMessages = append(errorMessages, utils.FirstNameRequiredError)
	}
	// if data.LastName == "" {
	// 	errorMessages = append(errorMessages, utils.LastNameRequiredError)
	// }
	if data.Email == "" {
		errorMessages = append(errorMessages, utils.EmailRequiredError)
	} else {
		if err := utils.CheckEmailSecurity(data.Email); err != nil {
			errorMessages = append(errorMessages, err.Error())
		}
	}

	if data.Password == "" {
		errorMessages = append(errorMessages, utils.PasswordRequiredError)
	}

	if len(data.Password) < 6 {
		errorMessages = append(errorMessages, utils.PasswordLengthError)
	} else {
		if err := utils.CheckPasswordSecurity(data.Password); err != nil {
			errorMessages = append(errorMessages, err.Error())
		}
	}

	if len(errorMessages) > 0 {
		utils.JsonError(w, strings.Join(errorMessages, ", "), http.StatusBadRequest, nil)
		return false
	}

	return true
}

func trackUpdatedUserFields(oldData payloads.UserResponse, newData models.User) map[string]interface{} {
	updatedFields := make(map[string]interface{})
	v := reflect.ValueOf(&newData).Elem()

	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		fieldName := field.Name
		fieldValue := v.Field(i)

		if fieldValue.IsZero() || fieldName == "ID" {
			continue
		}

		oldFieldValue := reflect.ValueOf(&oldData).Elem().FieldByName(fieldName)

		if !reflect.DeepEqual(fieldValue.Interface(), oldFieldValue.Interface()) {
			updatedFields[fieldName] = fieldValue.Interface()

			reflect.ValueOf(&oldData).Elem().FieldByName(fieldName).Set(fieldValue)
		}
	}

	return updatedFields
}

func (db *UserService) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	if ok := utils.CheckRequestMethod(w, r, http.MethodGet); !ok {
		return
	}

	var userService models.User
	userResponses, err := userService.GetAllUsers(db.DB)
	if err != nil {
		utils.JsonError(w, utils.UserNotFoundError, http.StatusNotFound, err)
		return
	}

	var output interface{}
	if len(userResponses) == 0 {
		output = []models.User{}
	} else {
		output = userResponses
	}

	utils.JsonResponse(output, w, utils.UsersFetchedSuccessfully, http.StatusOK)
}

func (db *UserService) FetchUserById(w http.ResponseWriter, r *http.Request) {
	if ok := utils.CheckRequestMethod(w, r, http.MethodGet); !ok {
		return
	}

	id, err := utils.GetUserIdFromPath(r)
	if err != nil {
		utils.JsonError(w, fmt.Sprintf(utils.InvalidUserIDError, id), http.StatusBadRequest, err)
		return
	}

	var userService models.User
	userResponse, err := userService.FetchUserById(db.DB, id)
	if err != nil {
		utils.JsonError(w, fmt.Sprintf(utils.UserNotFoundError, id), http.StatusNotFound, err)
		return
	}

	utils.JsonResponse(userResponse, w, fmt.Sprintf(utils.UserFetchedSuccessfully, id), http.StatusOK)
}

func (db *UserService) AddUser(w http.ResponseWriter, r *http.Request) {
	if ok := utils.CheckRequestMethod(w, r, http.MethodPost); !ok {
		return
	}

	var userRequest models.User
	if err := json.NewDecoder(r.Body).Decode(&userRequest); err != nil {
		utils.JsonError(w, utils.InvalidUserDataError, http.StatusBadRequest, err)
		return
	}

	if ok := validateAddUserRequest(w, userRequest); !ok {
		return
	}

	existingUser, err := userRequest.GetUserByEmail(db.DB, userRequest.Email)
	if err == nil && existingUser != nil {
		utils.JsonError(w, utils.EmailAlreadyExistsError, http.StatusConflict, nil)
		return
	}

	hashedPassword, err := utils.HashedPassword(userRequest.Password)
	if err != nil {
		utils.JsonError(w, utils.PasswordHashError, http.StatusInternalServerError, err)
		return
	}

	userRequest.Password = hashedPassword
	id, err := userRequest.AddUser(db.DB)
	if err != nil {
		utils.JsonError(w, utils.UserCreationError, http.StatusInternalServerError, err)
		return
	}

	userResponse, err := userRequest.FetchUserById(db.DB, id)
	if err != nil {
		utils.JsonError(w, fmt.Sprintf(utils.UserNotFoundError, id), http.StatusNotFound, err)
		return
	}

	utils.JsonResponse(userResponse, w, fmt.Sprintf(utils.UserCreatedSuccessfully, id), http.StatusOK)
}

func (db *UserService) UpdateUser(w http.ResponseWriter, r *http.Request) {
	if ok := utils.CheckRequestMethod(w, r, http.MethodPut); !ok {
		return
	}

	id, err := utils.GetUserIdFromPath(r)
	if err != nil {
		utils.JsonError(w, fmt.Sprintf(utils.InvalidUserIDError, id), http.StatusBadRequest, err)
		return
	}

	var newUserData models.User
	if err := json.NewDecoder(r.Body).Decode(&newUserData); err != nil {
		utils.JsonError(w, utils.InvalidUserDataError, http.StatusBadRequest, err)
		return
	}

	if len(newUserData.Email) > 0 {
		existingUser, err := newUserData.GetUserByEmail(db.DB, newUserData.Email)
		if err == nil && existingUser != nil && id != existingUser.ID {
			utils.JsonError(w, utils.EmailAlreadyExistsError, http.StatusConflict, nil)
			return
		}
	}

	var temp models.User
	oldUserData, err := temp.FetchUserById(db.DB, id)
	if err != nil {
		utils.JsonError(w, fmt.Sprintf(utils.UserNotFoundError, id), http.StatusNotFound, err)
		return
	}

	updatedFields := trackUpdatedUserFields(*oldUserData, newUserData)
	if len(updatedFields) == 0 {
		utils.JsonResponse(oldUserData, w, fmt.Sprintf(utils.UserNotModified, id), http.StatusNotModified)
		return
	}

	if _, err := newUserData.UpdateUser(db.DB, id, updatedFields); err != nil {
		utils.JsonResponse(oldUserData, w, fmt.Sprintf(utils.UserUpdateError, id), http.StatusInternalServerError)
		return
	}

	userResponse, err := temp.FetchUserById(db.DB, id)
	if err != nil {
		utils.JsonError(w, fmt.Sprintf(utils.UserNotFoundError, id), http.StatusNotFound, err)
		return
	}

	utils.JsonResponse(userResponse, w, fmt.Sprintf(utils.UserUpdatedSuccessfully, id), http.StatusOK)
}

func (db *UserService) DeleteUser(w http.ResponseWriter, r *http.Request) {
	if ok := utils.CheckRequestMethod(w, r, http.MethodDelete); !ok {
		return
	}

	id, err := utils.GetUserIdFromPath(r)
	if err != nil {
		utils.JsonError(w, fmt.Sprintf(utils.InvalidUserIDError, id), http.StatusBadRequest, err)
		return
	}

	var userRequest models.User
	userResponse, err := userRequest.FetchUserById(db.DB, id)
	if err != nil {
		utils.JsonError(w, fmt.Sprintf(utils.UserNotFoundError, id), http.StatusNotFound, err)
		return
	}

	if err := userRequest.DeleteUser(db.DB, id); err != nil {
		utils.JsonError(w, fmt.Sprintf(utils.UserDeletionError, id), http.StatusInternalServerError, err)
		return
	}

	utils.JsonResponse(userResponse, w, fmt.Sprintf(utils.UserDeletedSuccessfully, id), http.StatusOK)
}

func (db *UserService) LoginUser(w http.ResponseWriter, r *http.Request) {
	if ok := utils.CheckRequestMethod(w, r, http.MethodPost); !ok {
		return
	}

	var loginData models.LoginUser
	if err := json.NewDecoder(r.Body).Decode(&loginData); err != nil {
		utils.JsonError(w, utils.InvalidUserDataError, http.StatusBadRequest, err)
		return
	}

	var userRequest models.User
	user, err := userRequest.GetUserByEmail(db.DB, loginData.Email)
	if err != nil {
		utils.JsonError(w, utils.UnauthorizedError, http.StatusUnauthorized, err)
		return
	}

	token, err := utils.GenerateJWT(user.ID, user.Email)
	if err != nil {
		utils.JsonError(w, utils.TokenGenerationError, http.StatusInternalServerError, err)
		return
	}

	userResponse := models.CopyUserToUserResponse(user)
	authResponse := models.UserAuthResponse{
		Token: token,
		User:  *userResponse,
	}

	utils.JsonResponse(authResponse, w, fmt.Sprintf(utils.UserLoggedInSuccessfully, userResponse.ID), http.StatusOK)
}
