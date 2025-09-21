package services

import (
	"context"
	"e-commerce-backend/shared/notifications/emails"
	"e-commerce-backend/shared/notifications/emails/templates"
	"e-commerce-backend/shared/utils"
	"e-commerce-backend/users/internal/models"
	"e-commerce-backend/users/internal/repository"
	"e-commerce-backend/users/pkg/constants"
	"e-commerce-backend/users/pkg/payloads"
	pkg_utils "e-commerce-backend/users/utils"
	"e-commerce-backend/users/validator"
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

type Services struct {
	UserRepo *repository.UserRepo
}

func NewService(userRepo *repository.UserRepo) *Services {
	return &Services{
		UserRepo: userRepo,
	}
}

type Service struct {
	DB *gorm.DB
}

func NewUser(db *gorm.DB) *Service {
	return &Service{
		DB: db,
	}
}

type UersInterface interface {
	CreateUser(db *gorm.DB) error
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

	utils.SuccessResponseFunc(w, utils.UsersFetchedSuccessfully, output, http.StatusOK)
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

	utils.SuccessResponseFunc(w, fmt.Sprintf(utils.UserCreatedSuccessfully, id), userResponse, http.StatusOK)
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
		utils.SuccessResponseFunc(w, utils.RequestUserIsDeactivated, map[string]interface{}{"user_id": userID}, http.StatusForbidden)
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
		utils.SuccessResponseFunc(w, fmt.Sprintf(utils.UserNotModified, userID), oldUserData, http.StatusNotModified)
		return
	}

	if _, err := oldUserData.UpdateUser(db.DB, userID, updatedFields); err != nil {
		utils.SuccessResponseFunc(w, fmt.Sprintf(utils.UserUpdateError, userID), oldUserData, http.StatusInternalServerError)
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

	utils.SuccessResponseFunc(w, utils.ProfileUpdatedSuccessfully, userResponse, http.StatusOK)
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

	utils.SuccessResponseFunc(w, fmt.Sprintf(utils.UserDeletedSuccessfully, id), map[string]interface{}{"user_id": id}, http.StatusOK)
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
		utils.SuccessResponseFunc(w, fmt.Sprintf(utils.UserAlreadyDeactivated, id), map[string]interface{}{"user_id": id}, http.StatusForbidden)
		return
	}

	if err := user.DeActivateUser(db.DB, id); err != nil {
		utils.ErrorResponseFunc(w, fmt.Sprintf(utils.UserDeActivationFailed, id), http.StatusInternalServerError, err)
		return
	}

	utils.SuccessResponseFunc(w, utils.UserDeActivationSuccessfully, map[string]interface{}{"user_id": id}, http.StatusOK)
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
		utils.SuccessResponseFunc(w, fmt.Sprintf(utils.UserAlreadyActivated, id), map[string]interface{}{"user_id": id}, http.StatusConflict)
		return
	}

	if err := user.ActivateUser(db.DB, id); err != nil {
		utils.ErrorResponseFunc(w, fmt.Sprintf(utils.UserReactivationFailed, id), http.StatusInternalServerError, err)
		return
	}

	utils.SuccessResponseFunc(w, utils.UserReactivationSuccessfully, map[string]interface{}{"user_id": id}, http.StatusOK)
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

		if err := utils.CompareHashedPassword(user.Password, loginData.Password); err != nil {
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

		utils.SuccessResponseFunc(w, fmt.Sprintf(utils.UserLoggedInSuccessfully, userResponse.ID), authResponse, http.StatusOK)
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

	token := utils.GenerateRandomToken()

	userRepo := repository.NewUserRepo(db.DB)
	if err := userRepo.StoreToken(userData.ID, token, constants.PasswordReset); err != nil {
		utils.ErrorResponseFunc(w, utils.TokenGenerationError, http.StatusInternalServerError, err)
		return
	}

	clientURL := os.Getenv("CLIENT_URL") + "/new-password?token=" + token
	bodyContent := map[string]interface{}{
		"CustomerName":      userData.FirstName + " " + userData.LastName,
		"ResetPasswordLink": clientURL,
	}
	log.Println(bodyContent)
	emails.EmailWorker(request.Email, templates.PasswordResetSubject, templates.PASSWORD_RESET_TEMPLATE, bodyContent, []string{})
	utils.SuccessResponseFunc(w, utils.ResetPasswordTokenSent, map[string]interface{}{"user_id": userData.ID, "reset_password_link": clientURL}, http.StatusOK)
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

	utils.SuccessResponseFunc(w, utils.NewPasswordSetSuccessfully, map[string]interface{}{"user_id": userToken.UserID}, http.StatusCreated)
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

	userRepo := repository.NewUserRepo(db.DB)
	token := utils.GenerateRandomToken()
	if err := userRepo.StoreToken(userData.ID, token, constants.EmailVerificationType); err != nil {
		utils.ErrorResponseFunc(w, utils.TokenGenerationError, http.StatusInternalServerError, err)
		return
	}

	verificationURL := fmt.Sprintf("http://localhost:8080/user/verify/%s", token)
	utils.SuccessResponseFunc(w, utils.EmailVerificationTokenSent, verificationURL, http.StatusOK)
}

func (db *Service) VerifyUserEmail(w http.ResponseWriter, r *http.Request) {
	EmailVerificationToken, err := utils.GetTokenFromPath(r)
	if err != nil {
		utils.ErrorResponseFunc(w, utils.InvalidTokenError, http.StatusBadRequest, err)
		return
	}

	var user models.User
	userData, err := user.ValidateAndUseToken(db.DB, EmailVerificationToken, constants.EmailVerificationType)
	if err != nil {
		utils.ErrorResponseFunc(w, utils.InvalidTokenError, http.StatusBadRequest, err)
		return
	}

	if err := user.VerifyUserEmail(db.DB, userData.UserID); err != nil {
		utils.ErrorResponseFunc(w, fmt.Sprintf(utils.EmailVerificationFailed, userData.UserID), http.StatusInternalServerError, err)
		return
	}

	utils.SuccessResponseFunc(w, utils.EmailVerifiedSuccessfully, map[string]interface{}{"user_id": userData.UserID}, http.StatusOK)
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

func (s *Services) CreateUser(ctx context.Context, data validator.CreateUserValidator) (*models.User, error) {
	// Check if the email already exists
	// exists, err := s.IsEmailExists(ctx, data.Email)
	// if err != nil {
	// 	return nil, err
	// }
	// if exists {
	// 	return nil, errors.New(utils.EmailAlreadyExistsError)
	// }

	if err := utils.CheckPasswordSecurity(data.Password); err != nil {
		return nil, err
	}

	hashedPassword, err := utils.HashedPassword(data.Password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		FirstName: data.FirstName,
		LastName:  data.LastName,
		Email:     data.Email,
		Password:  hashedPassword,
	}

	// createdUser, err := s.UserRepo.CreateUser(user)
	// if err != nil {
	// 	return nil, err
	// }

	// Send email
	templateData := pkg_utils.UserCreationTD{
		Email:    user.Email,
		FullName: user.FirstName + " " + user.LastName,
	}

	template, err := pkg_utils.GetUserCreatedTemplate(templateData)
	if err != nil {
		return nil, err
	}

	if err := utils.SendEmail(user.Email, pkg_utils.SubjectUserCreated, template); err != nil {
		return nil, err
	}

	// Send verification email
	if err := s.SendVerificationEmail(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Services) IsEmailExists(ctx context.Context, email string) (bool, error) {
	return s.UserRepo.IsEmailExists(email)
}

func (s *Services) SendVerificationEmail(user *models.User) error {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	token := utils.GenerateRandomToken()
	if err := s.UserRepo.StoreToken(user.ID, token, constants.EmailVerificationType); err != nil {
		return err
	}

	url := os.Getenv("CLIENT_URL")
	template, err := pkg_utils.SendVerificationEmail(pkg_utils.VerificationMailTD{
		VerificationURL: url + "/verify?token=" + token,
		FullName:        user.FirstName + " " + user.LastName,
		Email:           user.Email,
	})
	if err != nil {
		return err
	}

	if err := utils.SendEmail(user.Email, pkg_utils.VerificationMailSubject, template); err != nil {
		return err
	}

	return nil
}

func (s *Services) VerifyToken(token string, tokenType int) (*models.UserToken, error) {
	userToken, err := s.UserRepo.VerifyToken(token, constants.EmailVerificationType)
	if err != nil {
		return nil, err
	}

	return userToken, nil
}

func (s *Services) LoginUser(ctx context.Context, data validator.LoginUserValidator) (*payloads.UserLoginResponse, error) {
	var user models.LoginUser

	user.Email = data.Email
	user.Password = data.Password

	dbUser, err := s.UserRepo.LoginUser(user)
	if err != nil {
		return nil, err
	}
	if err := utils.CompareHashedPassword(dbUser.Password, data.Password); err != nil {
		return nil, err
	}

	expireTime := time.Now().Add(time.Hour * 24).Unix()
	token, err := utils.GenerateJWT(dbUser.ID, user.Email, "user.Username", expireTime)
	if err != nil {
		return nil, err
	}

	authResponse := payloads.UserLoginResponse{
		Token:      token,
		ExpireTime: expireTime,
		User: payloads.UserProfileResp{
			ID:         dbUser.ID,
			FirstName:  dbUser.FirstName,
			LastName:   dbUser.LastName,
			FullName:   dbUser.FirstName + " " + dbUser.LastName,
			Email:      dbUser.Email,
			Username:   dbUser.Username,
			IsVerified: dbUser.IsVerified,
			IsActive:   dbUser.IsActive,
			Gender:     dbUser.Gender,
			BaseModel: payloads.BaseModel{
				CreatedAt: dbUser.CreatedAt,
				UpdatedAt: dbUser.UpdatedAt,
			},
		},
	}
	return &authResponse, nil
}

func (s *Services) FetchUserProfileByID(userID int) (*payloads.UserProfileResp, error) {
	user, err := s.UserRepo.FetchUserProfileByID(userID)
	if err != nil {
		return nil, err
	}
	return user, nil
}
