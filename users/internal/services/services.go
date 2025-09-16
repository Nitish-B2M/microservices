package services

import (
	"e-commerce-backend/shared/notifications/emails"
	"e-commerce-backend/shared/notifications/emails/templates"
	"e-commerce-backend/shared/utils"
	"e-commerce-backend/users/internal/models"
	"e-commerce-backend/users/pkg/constants"
	"e-commerce-backend/users/pkg/payloads"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/joho/godotenv"

	"gorm.io/gorm"
)

type Service struct {
	DB *gorm.DB
}

func NewUser(db *gorm.DB) *Service {
	return &Service{
		DB: db,
	}
}

type UserInterface interface {
	GetAllUsers(w http.ResponseWriter, r *http.Request)
	GetUserProfile(w http.ResponseWriter, r *http.Request)
	CreateUser(w http.ResponseWriter, r *http.Request)
	UpdateUser(w http.ResponseWriter, r *http.Request)
	DeleteUser(w http.ResponseWriter, r *http.Request)
	DeActivateUser(w http.ResponseWriter, r *http.Request)
	ActivateUser(w http.ResponseWriter, r *http.Request)
	LoginUser(w http.ResponseWriter, r *http.Request)
	RequestPasswordReset(w http.ResponseWriter, r *http.Request)
	ResetPassword(w http.ResponseWriter, r *http.Request)
	SendVerificationEmail(w http.ResponseWriter, r *http.Request)
	VerifyUserEmail(w http.ResponseWriter, r *http.Request)
}

func validateCreateUserRequest(w http.ResponseWriter, data models.User) bool {
	var errorMessages []string

	// Collect validation errors
	if data.FirstName == "" {
		errorMessages = append(errorMessages, utils.FirstNameRequiredError)
	}
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
		utils.ErrorResponseFunc(w, strings.Join(errorMessages, ", "), http.StatusBadRequest, errors.New(strings.Join(errorMessages, ", ")))
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
	var userService models.User
	if db.DB == nil {
		utils.ErrorResponseFunc(w, utils.DatabaseConnectionError, http.StatusInternalServerError, errors.New("database connection is nil"))
		return
	}
	userResponses, err := userService.GetAllUsers(db.DB)
	if err != nil {
		utils.ErrorResponseFunc(w, utils.UserNotFoundError, http.StatusNotFound, err)
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

func (db *Service) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	if ok := utils.CheckRequestMethod(w, r, http.MethodPost); !ok {
		return
	}

	id := utils.GetUserIDFromContext(r)

	var userService models.User
	userResponse, err := userService.GetUserByID(db.DB, id)
	if err != nil {
		if strings.Contains(err.Error(), gorm.ErrRecordNotFound.Error()) {
			utils.ErrorResponseFunc(w, fmt.Sprintf(utils.UserNotFoundError, id), http.StatusNotFound, err)
			return
		}
		utils.ErrorResponseFunc(w, fmt.Sprintf(utils.UserNotFoundError, id), http.StatusNotFound, err)
		return
	}
	if !userResponse.IsActive {
		utils.SuccessResponseFunc(w, utils.RequestUserIsDeactivated, map[string]interface{}{"user_id": id}, http.StatusForbidden)
		return
	}

	userName := utils.GetUserNameIDFromContext(r)
	for _, role := range userResponse.Role.Roles {
		if role.Username == userName {
			userResponse.Role.ActiveRole = role
			break
		}
	}

	utils.SuccessResponseFunc(w, fmt.Sprintf(utils.UserFetchedSuccessfully, id), userResponse, http.StatusOK)
}

func (db *Service) CreateUser(w http.ResponseWriter, r *http.Request) {
	var userRequest models.User
	if err := json.NewDecoder(r.Body).Decode(&userRequest); err != nil {
		utils.ErrorResponseFunc(w, utils.InvalidUserDataError, http.StatusBadRequest, err)
		return
	}

	if ok := validateCreateUserRequest(w, userRequest); !ok {
		return
	}

	existingUser, err := userRequest.GetUserByEmail(db.DB, userRequest.Email)
	if existingUser != nil {
		utils.ErrorResponseFunc(w, utils.EmailAlreadyExistsError, http.StatusConflict, err)
		return
	}
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
		} else {
			utils.ErrorResponseFunc(w, utils.UserCreationError, http.StatusInternalServerError, err)
			return
		}
	}

	hashedPassword, err := utils.HashedPassword(userRequest.Password)
	if err != nil {
		utils.ErrorResponseFunc(w, utils.PasswordHashError, http.StatusInternalServerError, err)
		return
	}

	userRequest.Password = hashedPassword
	id, err := userRequest.CreateUser(db.DB)
	if err != nil {
		utils.ErrorResponseFunc(w, utils.UserCreationError, http.StatusInternalServerError, err)
		return
	}

	userResponse, err := userRequest.GetUserByID(db.DB, id)
	if err != nil {
		if strings.Contains(err.Error(), gorm.ErrRecordNotFound.Error()) {
			utils.ErrorResponseFunc(w, fmt.Sprintf(utils.UserNotFoundError, id), http.StatusNotFound, err)
			return
		}
		utils.ErrorResponseFunc(w, fmt.Sprintf(utils.UserNotFoundError, id), http.StatusNotFound, err)
		return
	}

	//send registration success mail (go routine)
	emailContent := emails.UserCreation{Email: userResponse.Email, Role: "user"} //need to change role as role functionality added
	emails.EmailWorkerWithGoRoutine(userResponse.Email, templates.UserRegistrationSuccessFul, templates.USER_CREATED_TEMPLATE, emailContent, []string{})

	utils.JsonResponse(userResponse, w, fmt.Sprintf(utils.UserCreatedSuccessfully, 1), http.StatusOK)
}

func (db *Service) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userID := utils.GetUserIDFromContext(r)
	var newUserData payloads.UserUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&newUserData); err != nil {
		utils.ErrorResponseFunc(w, utils.InvalidUserDataError, http.StatusBadRequest, err)
		return
	}

	var oldUserData models.User
	_, err := oldUserData.GetUserByID(db.DB, userID)
	if err != nil {
		if strings.Contains(err.Error(), gorm.ErrRecordNotFound.Error()) {
			utils.ErrorResponseFunc(w, fmt.Sprintf(utils.UserNotFoundError, userID), http.StatusNotFound, err)
			return
		}
		utils.ErrorResponseFunc(w, fmt.Sprintf(utils.UserNotFoundError, userID), http.StatusNotFound, err)
		return
	}
	if !oldUserData.IsActive {
		utils.JsonResponse(map[string]interface{}{"user_id": userID}, w, utils.RequestUserIsDeactivated, http.StatusForbidden)
		return
	}

	if len(newUserData.Email) > 0 {
		var temp models.User
		_, err := temp.GetUserByEmail(db.DB, newUserData.Email)
		if err == nil && userID != temp.ID {
			utils.ErrorResponseFunc(w, utils.EmailAlreadyExistsError, http.StatusConflict, errors.New(utils.EmailAlreadyExistsError))
			return
		}
	}

	updatedFields := trackUpdatedUserFields(oldUserData, newUserData)
	if len(updatedFields) == 0 {
		utils.JsonResponse(oldUserData, w, fmt.Sprintf(utils.UserNotModified, userID), http.StatusNotModified)
		return
	}

	if _, err := oldUserData.UpdateUser(db.DB, userID, updatedFields); err != nil {
		utils.JsonResponse(oldUserData, w, fmt.Sprintf(utils.UserUpdateError, userID), http.StatusInternalServerError)
		return
	}

	userResponse, err := oldUserData.GetUserByID(db.DB, userID)
	if err != nil {
		if strings.Contains(err.Error(), gorm.ErrRecordNotFound.Error()) {
			utils.ErrorResponseFunc(w, fmt.Sprintf(utils.UserNotFoundError, userID), http.StatusNotFound, err)
			return
		}
		utils.ErrorResponseFunc(w, fmt.Sprintf(utils.UserNotFoundError, userID), http.StatusNotFound, err)
		return
	}

	utils.JsonResponse(userResponse, w, utils.ProfileUpdatedSuccessfully, http.StatusOK)
}

func (db *Service) DeleteUser(w http.ResponseWriter, r *http.Request) {
	if ok := utils.CheckRequestMethod(w, r, http.MethodDelete); !ok {
		return
	}

	id, err := utils.GetIDFromPath(r)
	if err != nil {
		utils.ErrorResponseFunc(w, fmt.Sprintf(utils.InvalidUserIDError, id), http.StatusBadRequest, err)
		return
	}

	var userRequest models.User
	_, err = userRequest.GetUserByID(db.DB, id)
	if err != nil {
		if strings.Contains(err.Error(), gorm.ErrRecordNotFound.Error()) {
			utils.ErrorResponseFunc(w, fmt.Sprintf(utils.UserNotFoundError, id), http.StatusNotFound, err)
			return
		}
		utils.ErrorResponseFunc(w, fmt.Sprintf(utils.UserNotFoundError, id), http.StatusNotFound, err)
		return
	}

	if err := userRequest.DeleteUser(db.DB, id); err != nil {
		utils.ErrorResponseFunc(w, fmt.Sprintf(utils.UserDeletionError, id), http.StatusInternalServerError, err)
		return
	}

	utils.JsonResponse(map[string]interface{}{"user_id": id}, w, fmt.Sprintf(utils.UserDeletedSuccessfully, id), http.StatusOK)
}

func (db *Service) DeActivateUser(w http.ResponseWriter, r *http.Request) {
	id, err := utils.GetIDFromPath(r)
	if err != nil {
		utils.ErrorResponseFunc(w, fmt.Sprintf(utils.InvalidUserIDError, id), http.StatusBadRequest, err)
		return
	}

	var user models.User
	_, err = user.GetUserByID(db.DB, id)
	if err != nil {
		if strings.Contains(err.Error(), gorm.ErrRecordNotFound.Error()) {
			utils.ErrorResponseFunc(w, fmt.Sprintf(utils.UserNotFoundError, id), http.StatusNotFound, err)
			return
		}
		utils.ErrorResponseFunc(w, fmt.Sprintf(utils.UserNotFoundError, id), http.StatusNotFound, err)
		return
	}

	//check user already de-active or not
	if !user.IsActive {
		utils.JsonResponse(map[string]interface{}{"user_id": id}, w, fmt.Sprintf(utils.UserAlreadyDeactivated, id), http.StatusForbidden)
		return
	}

	if err := user.DeActivateUser(db.DB, id); err != nil {
		utils.ErrorResponseFunc(w, fmt.Sprintf(utils.UserDeActivationFailed, id), http.StatusInternalServerError, err)
		return
	}

	utils.JsonResponse(map[string]interface{}{"user_id": id}, w, utils.UserDeActivationSuccessfully, http.StatusOK)
}

func (db *Service) ActivateUser(w http.ResponseWriter, r *http.Request) {
	id, err := utils.GetIDFromPath(r)
	if err != nil {
		utils.ErrorResponseFunc(w, fmt.Sprintf(utils.InvalidUserIDError, id), http.StatusBadRequest, err)
		return
	}

	var user models.User
	_, err = user.GetUserByID(db.DB, id)
	if err != nil {
		if strings.Contains(err.Error(), gorm.ErrRecordNotFound.Error()) {
			utils.ErrorResponseFunc(w, fmt.Sprintf(utils.UserNotFoundError, id), http.StatusNotFound, err)
			return
		}
		utils.ErrorResponseFunc(w, fmt.Sprintf(utils.UserNotFoundError, id), http.StatusNotFound, err)
		return
	}

	//check user already active or not
	if user.IsActive {
		utils.JsonResponse(map[string]interface{}{"user_id": id}, w, fmt.Sprintf(utils.UserAlreadyActivated, id), http.StatusConflict)
		return
	}

	if err := user.ActivateUser(db.DB, id); err != nil {
		utils.ErrorResponseFunc(w, fmt.Sprintf(utils.UserReactivationFailed, id), http.StatusInternalServerError, err)
		return
	}

	utils.JsonResponse(map[string]interface{}{"user_id": id}, w, utils.UserReactivationSuccessfully, http.StatusOK)
}

func (db *Service) LoginUser(w http.ResponseWriter, r *http.Request) {
	var loginData payloads.LoginUser
	if err := json.NewDecoder(r.Body).Decode(&loginData); err != nil {
		utils.ErrorResponseFunc(w, utils.InvalidUserEmailPassword, http.StatusBadRequest, err)
		return
	}

	if ok, errs := utils.ValidateStructUsingValidators(loginData); !ok {
		errStr := strings.Join(errs, ", ")
		utils.ErrorResponseFunc(w, utils.InvalidUserEmailPassword, http.StatusBadRequest, errors.New(errStr))
		return
	}

	var userRequest models.User
	if loginData.Username != "" { //temporary
		//new way to login
		user, err := models.LoggedInUser(db.DB, loginData.Username, loginData.Password)
		if err != nil {
			utils.ErrorResponseFunc(w, fmt.Sprintf(utils.UsernameNotFoundError, loginData.Username), http.StatusNotFound, err)
			return
		}
		if !user.IsActive {
			if err := user.ActivateUser(db.DB, user.ID); err != nil {
				utils.ErrorResponseFunc(w, utils.InternalServerError, http.StatusInternalServerError, err)
				return
			}
		}
		userWithRoles, err := userRequest.GetUserUsingUsername(db.DB, loginData.Username)
		if err != nil {
			utils.ErrorResponseFunc(w, fmt.Sprintf(utils.UsernameNotFoundError, loginData.Username), http.StatusNotFound, err)
			return
		}
		expireTime := time.Now().Add(time.Hour * 24).Unix()
		token, err := utils.GenerateJWT(user.ID, user.Email, loginData.Username, expireTime)
		if err != nil {
			utils.ErrorResponseFunc(w, utils.TokenGenerationError, http.StatusInternalServerError, err)
			return
		}

		authResponse := payloads.UserAuthResponse{
			Token:      token,
			ExpireTime: expireTime,
			User:       *userWithRoles,
		}
		utils.SuccessResponseFunc(w, fmt.Sprintf(utils.UserLoggedInSuccessfully, userWithRoles.ID), authResponse, http.StatusOK)
		return
	} else {
		user, err := userRequest.GetUserByEmail(db.DB, loginData.Email)
		if err != nil {
			if strings.Contains(err.Error(), utils.RequestUserIsDeactivated) {
			} else if strings.Contains(err.Error(), utils.UserIsNotVerifiedError) {
			} else if ok := customEmailErrorMessage(w, err, loginData.Email); !ok {
				return
			}
		}

		if ok, err := utils.CompareHashedPassword(user.Password, loginData.Password); !ok {
			utils.ErrorResponseFunc(w, "Invalid email or password", http.StatusUnauthorized, err)
			return
		}
		expireTime := time.Now().Add(time.Hour * 24).Unix()
		token, err := utils.GenerateJWT(user.ID, user.Email, "user.Username", expireTime)
		if err != nil {
			utils.ErrorResponseFunc(w, utils.TokenGenerationError, http.StatusInternalServerError, err)
			return
		}
		userResponse := models.CopyUserToUserResponse(user)
		authResponse := payloads.UserAuthResponse{
			Token:      token,
			ExpireTime: expireTime,
			User:       *userResponse,
		}

		utils.JsonResponse(authResponse, w, fmt.Sprintf(utils.UserLoggedInSuccessfully, userResponse.ID), http.StatusOK)
		return
	}
}

func (db *Service) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	if ok := utils.CheckRequestMethod(w, r, http.MethodPost); !ok {
		return
	}
	if err := godotenv.Load("../../.env"); err != nil {
		log.Fatal(".env file not found from main.go")
	}

	var request struct {
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		utils.ErrorResponseFunc(w, utils.InvalidUserDataError, http.StatusBadRequest, err)
		return
	}

	var user models.User
	userData, err := user.GetUserByEmail(db.DB, request.Email)
	if err != nil {
		if strings.Contains(err.Error(), utils.UserIsNotVerifiedError) {
		} else if strings.Contains(err.Error(), utils.RequestUserIsDeactivated) {
		} else if ok := customEmailErrorMessage(w, err, request.Email); !ok {
			return
		}
	}

	token, err := userData.GenerateUserToken(db.DB, constants.PasswordReset)
	if err != nil {
		utils.ErrorResponseFunc(w, utils.TokenGenerationError, http.StatusInternalServerError, err)
		return
	}

	clientURL := os.Getenv("CLIENT_URL")
	clientURL += os.Getenv("CLIENT_PORT") + "/new-password?token=" + token
	bodyContent := map[string]interface{}{
		"CustomerName":      userData.FirstName + " " + userData.LastName,
		"ResetPasswordLink": clientURL,
	}
	log.Println(bodyContent)
	emails.EmailWorker(request.Email, templates.PasswordResetSubject, templates.PASSWORD_RESET_TEMPLATE, bodyContent, []string{})
	utils.JsonResponse(map[string]interface{}{"user_id": userData.ID, "reset_password_link": clientURL}, w, utils.ResetPasswordTokenSent, http.StatusOK)
}

func (db *Service) ResetPassword(w http.ResponseWriter, r *http.Request) {
	if ok := utils.CheckRequestMethod(w, r, http.MethodPost); !ok {
		return
	}
	resetPasswordToken := r.URL.Query().Get("token")
	if resetPasswordToken == "" {
		utils.ErrorResponseFunc(w, utils.InvalidTokenError, http.StatusBadRequest, errors.New(utils.MissingTokenError))
		return
	}

	var request struct {
		NewPassword string `json:"new_password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		utils.ErrorResponseFunc(w, utils.MissingTokenError, http.StatusBadRequest, err)
		return
	}

	//check password security
	if err := utils.CheckPasswordSecurity(request.NewPassword); err != nil {
		utils.JsonErrorWithExtra(w, utils.InvalidPasswordError, http.StatusBadRequest, err, map[string]interface{}{"function": "ResetPassword"})
		return
	}

	var user models.User
	userToken, err := user.ValidateAndUseToken(db.DB, resetPasswordToken, constants.PasswordReset)
	if err != nil {
		utils.ErrorResponseFunc(w, utils.PassResetTokenExpired, http.StatusBadRequest, err)
		return
	}

	if err := user.ResetPassword(db.DB, userToken.UserID, request.NewPassword); err != nil {
		utils.ErrorResponseFunc(w, utils.PasswordResetError, http.StatusInternalServerError, err)
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
		utils.ErrorResponseFunc(w, utils.EmailRequiredError, http.StatusBadRequest, err)
		return
	}

	if request.Email == "" {
		utils.ErrorResponseFunc(w, utils.EmailRequiredError, http.StatusBadRequest, errors.New(utils.EmailRequiredError))
		return
	}

	var user models.User
	userData, err := user.GetUserByEmail(db.DB, request.Email)
	if err != nil {
		if strings.Contains(err.Error(), utils.UserIsNotVerifiedError) {
		} else if ok := customEmailErrorMessage(w, err, request.Email); !ok {
			return
		}
	}

	if ok := user.CheckUserEmailAlreadyVerified(db.DB, request.Email); ok {
		utils.ErrorResponseFunc(w, utils.EmailAlreadyVerified, http.StatusAlreadyReported, err)
		return
	}

	token, err := userData.GenerateUserToken(db.DB, constants.EmailVerification)
	if err != nil {
		utils.ErrorResponseFunc(w, utils.TokenGenerationError, http.StatusInternalServerError, err)
		return
	}

	verificationURL := fmt.Sprintf("http://localhost:8080/user/verify/%s", token)
	utils.JsonResponse(verificationURL, w, utils.EmailVerificationTokenSent, http.StatusOK)
}

func (db *Service) VerifyUserEmail(w http.ResponseWriter, r *http.Request) {
	emailVerificationToken, err := utils.GetTokenFromPath(r)
	if err != nil {
		utils.ErrorResponseFunc(w, utils.InvalidTokenError, http.StatusBadRequest, err)
		return
	}

	var user models.User
	userData, err := user.ValidateAndUseToken(db.DB, emailVerificationToken, constants.EmailVerification)
	if err != nil {
		utils.ErrorResponseFunc(w, utils.InvalidTokenError, http.StatusBadRequest, err)
		return
	}

	if err := user.VerifyUserEmail(db.DB, userData.UserID); err != nil {
		utils.ErrorResponseFunc(w, fmt.Sprintf(utils.EmailVerificationFailed, userData.UserID), http.StatusInternalServerError, err)
		return
	}

	utils.JsonResponse(map[string]interface{}{"user_id": userData.UserID}, w, utils.EmailVerifiedSuccessfully, http.StatusOK)
}

func customEmailErrorMessage(w http.ResponseWriter, err error, email string) bool {
	if err == nil {
		return false
	}
	if strings.Contains(err.Error(), "record not found") {
		utils.ErrorResponseFunc(w, fmt.Sprintf(utils.UserNotFoundWithEmailError, email), http.StatusNotFound, err)
		return false
	} else if strings.Contains(err.Error(), "not verified") {
		utils.ErrorResponseFunc(w, strings.Join([]string{fmt.Sprintf(utils.EmailNotVerifiedError, email), utils.PleaseVerifyEmail, "click on this link: http://localhost:8080/user/verify/send"}, ", "), http.StatusUnauthorized, err)
		return false
	}
	utils.ErrorResponseFunc(w, utils.UnexpectedDatabaseError, http.StatusInternalServerError, err)
	return false
}
