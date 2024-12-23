package user

import (
	"encoding/json"
	"fmt"
	"microservices/pkg/shared/constants"
	"microservices/pkg/shared/models"
	"microservices/pkg/shared/payloads"
	"microservices/pkg/shared/utils"
	"net/http"
	"reflect"
	"strings"

	"gorm.io/gorm"
)

type Service struct {
	DB *gorm.DB
}

type User interface {
	AddUser(w http.ResponseWriter, r *http.Request)
	UpdateUser(w http.ResponseWriter, r *http.Request)
	DeleteUser(w http.ResponseWriter, r *http.Request)
	GetAllUsers(w http.ResponseWriter, r *http.Request)
	FetchUserById(w http.ResponseWriter, r *http.Request)
	LoginUser(w http.ResponseWriter, r *http.Request)
	RequestPasswordReset(w http.ResponseWriter, r *http.Request)
	ResetPassword(w http.ResponseWriter, r *http.Request)
	SendResetPasswordEmail(w http.ResponseWriter, r *http.Request)
	VerifyEmail(w http.ResponseWriter, r *http.Request)
	DeActivateUser(w http.ResponseWriter, r *http.Request)
	ActivateUser(w http.ResponseWriter, r *http.Request)
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

func trackUpdatedUserFields(oldData models.User, newData payloads.UserUpdateRequest) map[string]interface{} {
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

func (db *Service) GetAllUsers(w http.ResponseWriter, r *http.Request) {
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

func (db *Service) FetchUserById(w http.ResponseWriter, r *http.Request) {
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

func (db *Service) AddUser(w http.ResponseWriter, r *http.Request) {
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

func (db *Service) UpdateUser(w http.ResponseWriter, r *http.Request) {
	if ok := utils.CheckRequestMethod(w, r, http.MethodPut); !ok {
		return
	}

	id, err := utils.GetUserIdFromPath(r)
	if err != nil {
		utils.JsonError(w, fmt.Sprintf(utils.InvalidUserIDError, id), http.StatusBadRequest, err)
		return
	}

	var newUserData payloads.UserUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&newUserData); err != nil {
		utils.JsonError(w, utils.InvalidUserDataError, http.StatusBadRequest, err)
		return
	}

	var oldUserData models.User
	_, err = oldUserData.FetchUserById(db.DB, id)
	if err != nil {
		utils.JsonError(w, fmt.Sprintf(utils.UserNotFoundError, id), http.StatusNotFound, err)
		return
	}

	if len(newUserData.Email) > 0 {
		_, err := oldUserData.GetUserByEmail(db.DB, newUserData.Email)
		if err == nil && &oldUserData != nil && id != oldUserData.ID {
			utils.JsonError(w, utils.EmailAlreadyExistsError, http.StatusConflict, nil)
			return
		}
	}

	updatedFields := trackUpdatedUserFields(oldUserData, newUserData)
	if len(updatedFields) == 0 {
		utils.JsonResponse(oldUserData, w, fmt.Sprintf(utils.UserNotModified, id), http.StatusNotModified)
		return
	}

	if _, err := oldUserData.UpdateUser(db.DB, id, updatedFields); err != nil {
		utils.JsonResponse(oldUserData, w, fmt.Sprintf(utils.UserUpdateError, id), http.StatusInternalServerError)
		return
	}

	userResponse, err := oldUserData.FetchUserById(db.DB, id)
	if err != nil {
		utils.JsonError(w, fmt.Sprintf(utils.UserNotFoundError, id), http.StatusNotFound, err)
		return
	}

	utils.JsonResponse(userResponse, w, fmt.Sprintf(utils.UserUpdatedSuccessfully, id), http.StatusOK)
}

func (db *Service) DeleteUser(w http.ResponseWriter, r *http.Request) {
	if ok := utils.CheckRequestMethod(w, r, http.MethodDelete); !ok {
		return
	}

	id, err := utils.GetUserIdFromPath(r)
	if err != nil {
		utils.JsonError(w, fmt.Sprintf(utils.InvalidUserIDError, id), http.StatusBadRequest, err)
		return
	}

	var userRequest models.User
	_, err = userRequest.FetchUserById(db.DB, id)
	if err != nil {
		utils.JsonError(w, fmt.Sprintf(utils.UserNotFoundError, id), http.StatusNotFound, err)
		return
	}

	if err := userRequest.DeleteUser(db.DB, id); err != nil {
		utils.JsonError(w, fmt.Sprintf(utils.UserDeletionError, id), http.StatusInternalServerError, err)
		return
	}

	utils.JsonResponse(map[string]interface{}{"user_id": id}, w, fmt.Sprintf(utils.UserDeletedSuccessfully, id), http.StatusOK)
}

func (db *Service) DeActivateUser(w http.ResponseWriter, r *http.Request) {
	if ok := utils.CheckRequestMethod(w, r, http.MethodPost); !ok {
		return
	}

	id, err := utils.GetUserIdFromPath(r)
	if err != nil {
		utils.JsonError(w, fmt.Sprintf(utils.InvalidUserIDError, id), http.StatusBadRequest, err)
		return
	}

	var user models.User
	_, err = user.FetchUserById(db.DB, id)
	if err != nil {
		utils.JsonError(w, fmt.Sprintf(utils.UserNotFoundError, id), http.StatusNotFound, err)
		return
	}

	//check user already de-active or not
	if !user.IsActive {
		utils.JsonResponse(map[string]interface{}{"user_id": id}, w, fmt.Sprintf(utils.UserAlreadyDeactivated, id), http.StatusForbidden)
		return
	}

	if err := user.DeActivateUser(db.DB, id); err != nil {
		utils.JsonError(w, fmt.Sprintf(utils.UserDeActivationFailed, id), http.StatusInternalServerError, err)
		return
	}

	utils.JsonResponse(map[string]interface{}{"user_id": id}, w, utils.UserDeActivationSuccessfully, http.StatusOK)
}

func (db *Service) ActivateUser(w http.ResponseWriter, r *http.Request) {
	if ok := utils.CheckRequestMethod(w, r, http.MethodPost); !ok {
		return
	}

	id, err := utils.GetUserIdFromPath(r)
	if err != nil {
		utils.JsonError(w, fmt.Sprintf(utils.InvalidUserIDError, id), http.StatusBadRequest, err)
		return
	}

	var user models.User
	_, err = user.FetchUserById(db.DB, id)
	if err != nil {
		utils.JsonError(w, fmt.Sprintf(utils.UserNotFoundError, id), http.StatusNotFound, err)
		return
	}

	//check user already active or not
	if user.IsActive {
		utils.JsonResponse(map[string]interface{}{"user_id": id}, w, fmt.Sprintf(utils.UserAlreadyActivated, id), http.StatusConflict)
		return
	}

	if err := user.ActivateUser(db.DB, id); err != nil {
		utils.JsonError(w, fmt.Sprintf(utils.UserReactivationFailed, id), http.StatusInternalServerError, err)
		return
	}

	utils.JsonResponse(user, w, utils.UserReactivationSuccessfully, http.StatusOK)
}

func (db *Service) LoginUser(w http.ResponseWriter, r *http.Request) {
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

	if ok, err := utils.CompareHashedPassword(user.Password, loginData.Password); !ok {
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

func (db *Service) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	if ok := utils.CheckRequestMethod(w, r, http.MethodPost); !ok {
		return
	}

	var request struct {
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		utils.JsonError(w, utils.InvalidUserDataError, http.StatusBadRequest, err)
		return
	}

	var user models.User
	userData, err := user.GetUserByEmail(db.DB, request.Email)
	if err != nil {
		utils.JsonError(w, utils.UnauthorizedError, http.StatusBadRequest, err)
		return
	}

	token, err := userData.GenerateUserToken(db.DB, constants.PasswordReset)
	if err != nil {
		utils.JsonError(w, utils.TokenGenerationError, http.StatusInternalServerError, err)
		return
	}

	// utils.SendEmail(userData.Email, "Password Reset", fmt.Sprintf(utils.ResetTokenValue, token))
	utils.JsonResponse(map[string]interface{}{"user_id": userData.ID, "token": token}, w, utils.ResetPasswordTokenSent, http.StatusOK)
}

func (db *Service) ResetPassword(w http.ResponseWriter, r *http.Request) {
	if ok := utils.CheckRequestMethod(w, r, http.MethodPost); !ok {
		return
	}

	var request struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		utils.JsonError(w, utils.MissingTokenError, http.StatusBadRequest, err)
		return
	}

	//check password security
	if err := utils.CheckPasswordSecurity(request.NewPassword); err != nil {
		utils.JsonErrorWithExtra(w, utils.InvalidPasswordError, http.StatusBadRequest, err)
		return
	}

	var user models.User
	userToken, err := user.ValidateAndUseToken(db.DB, request.Token, constants.PasswordReset)
	if err != nil {
		utils.JsonError(w, utils.InvalidTokenError, http.StatusBadRequest, err)
		return
	}

	if err := user.ResetPassword(db.DB, userToken.UserID, request.NewPassword); err != nil {
		utils.JsonError(w, utils.PasswordResetError, http.StatusInternalServerError, err)
		return
	}

	utils.JsonResponse(map[string]interface{}{"user_id": userToken.UserID}, w, utils.NewPasswordSetSuccessfully, http.StatusCreated)
}

func (db *Service) SendVerificationEmail(w http.ResponseWriter, r *http.Request) {
	if ok := utils.CheckRequestMethod(w, r, http.MethodPost); !ok {
		return
	}

	var request struct {
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		utils.JsonError(w, utils.EmailRequiredError, http.StatusBadRequest, err)
		return
	}

	var user models.User
	userData, err := user.GetUserByEmail(db.DB, request.Email)
	if err != nil {
		utils.JsonError(w, utils.UnexpectedDatabaseError, http.StatusInternalServerError, err)
		return
	}

	if ok := user.CheckUserEmailAlreadyVerified(db.DB, request.Email); ok {
		utils.JsonError(w, utils.EmailAlreadyVerified, http.StatusAlreadyReported, err)
		return
	}

	token, err := userData.GenerateUserToken(db.DB, constants.EmailVerification)
	if err != nil {
		utils.JsonError(w, utils.TokenGenerationError, http.StatusInternalServerError, err)
		return
	}

	verificationURL := fmt.Sprintf("http://localhost:8080/user/verify/%s", token)
	utils.JsonResponse(verificationURL, w, utils.EmailVerificationTokenSent, http.StatusOK)
}

func (db *Service) VerifyUserEmail(w http.ResponseWriter, r *http.Request) {
	token, err := utils.GetTokenFromPath(r)
	if err != nil {
		utils.JsonError(w, utils.InvalidTokenError, http.StatusBadRequest, err)
		return
	}

	var user models.User
	userData, err := user.ValidateAndUseToken(db.DB, token, constants.EmailVerification)
	if err != nil {
		utils.JsonError(w, utils.InvalidTokenError, http.StatusBadRequest, err)
		return
	}

	if err := user.VerifyUserEmail(db.DB, userData.UserID); err != nil {
		utils.JsonError(w, fmt.Sprintf(utils.EmailVerificationFailed, userData.UserID), http.StatusInternalServerError, err)
		return
	}

	utils.JsonResponse(map[string]interface{}{"user_id": userData.UserID}, w, utils.EmailVerifiedSuccessfully, http.StatusOK)
}
