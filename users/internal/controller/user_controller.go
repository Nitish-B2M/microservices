// Package controller provides HTTP handler functions for user-related endpoints.
package controller

import (
	"e-commerce-backend/shared/utils"
	"e-commerce-backend/users/internal/services"
	"e-commerce-backend/users/pkg/constants"
	"e-commerce-backend/users/pkg/payloads"
	pkg_utils "e-commerce-backend/users/utils"
	"e-commerce-backend/users/validator"
	"net/http"
)

type UserController struct {
	Service *services.Services
}

func NewUserController(svc *services.Services) *UserController {
	return &UserController{Service: svc}
}

type UserRoleController struct {
	RoleService *services.RoleServices
}

func NewUserRoleController(svc *services.RoleServices) *UserRoleController {
	return &UserRoleController{RoleService: svc}
}

func (s *UserController) CreateUser(w http.ResponseWriter, r *http.Request) {
	userInput, err := validator.GetValidatedStruct[validator.CreateUserValidator](r.Context())
	if err != nil {
		utils.ErrorResponseFunc(w, err.Error(), http.StatusBadRequest, err)
		return
	}

	user, err := s.Service.CreateUser(r.Context(), *userInput)
	if err != nil {
		utils.ErrorResponseFunc(w, err.Error(), http.StatusInternalServerError, err)
		return
	}

	userResponse := payloads.UserCreatedResponse{UserID: user.ID}

	utils.SuccessResponseFunc(w, "User created successfully", userResponse, http.StatusCreated)
}

func (s *UserController) LoginUser(w http.ResponseWriter, r *http.Request) {
	userInput, err := validator.GetValidatedStruct[validator.LoginUserValidator](r.Context())
	if err != nil {
		utils.ErrorResponseFunc(w, err.Error(), http.StatusBadRequest, err)
		return
	}

	user, err := s.Service.LoginUser(r.Context(), *userInput)
	if err != nil {
		utils.ErrorResponseFunc(w, err.Error(), http.StatusInternalServerError, err)
		return
	}

	utils.SuccessResponseFunc(w, pkg_utils.UserCreatedSuccessfully, user, http.StatusOK)
}

func (s *UserController) SendVerificationEmail(w http.ResponseWriter, r *http.Request) {
	userInput, err := validator.GetValidatedStruct[validator.SendVerificationEmailValidator](r.Context())
	if err != nil {
		utils.ErrorResponseFunc(w, err.Error(), http.StatusBadRequest, err)
		return
	}

	err = s.Service.SendEmailVerificationMail(userInput.Email)
	if err != nil {
		utils.ErrorResponseFunc(w, err.Error(), http.StatusInternalServerError, err)
		return
	}

	utils.SuccessResponseFunc(w, "Verification email sent successfully", nil, http.StatusOK)
}

func (s *UserController) VerifyUserEmail(w http.ResponseWriter, r *http.Request) {
	token, err := validator.GetValidatedStruct[validator.EmailVerificationValidator](r.Context())
	if err != nil {
		utils.ErrorResponseFunc(w, err.Error(), http.StatusBadRequest, err)
		return
	}

	tokenString := token.Token
	user, err := s.Service.VerifyToken(tokenString, constants.EmailVerificationType)
	if err != nil {
		utils.ErrorResponseFunc(w, err.Error(), http.StatusBadRequest, err)
		return
	}

	utils.SuccessResponseFunc(w, "Token validated successfully", user.ID, http.StatusOK)
}

func (s *UserController) FetchUserProfileByID(w http.ResponseWriter, r *http.Request) {
	userID, err := validator.GetValidatedStruct[validator.FetchUserProfileValidator](r.Context())
	if err != nil {
		utils.ErrorResponseFunc(w, err.Error(), http.StatusBadRequest, err)
		return
	}

	user, err := s.Service.FetchUserProfileByID(userID.ID)
	if err != nil {
		utils.ErrorResponseFunc(w, err.Error(), http.StatusInternalServerError, err)
		return
	}

	utils.SuccessResponseFunc(w, "User profile fetched successfully", user, http.StatusOK)
}

func (ur *UserRoleController) AddAdminRole(w http.ResponseWriter, r *http.Request) {
	userID, err := validator.GetValidatedStruct[validator.FetchUserProfileValidator](r.Context())
	if err != nil {
		utils.ErrorResponseFunc(w, err.Error(), http.StatusBadRequest, err)
		return
	}
	if err := ur.RoleService.AddAdminRole(userID.ID); err != nil {
		utils.ErrorResponseFunc(w, err.Error(), http.StatusInternalServerError, err)
		return
	}

	utils.SuccessResponseFunc(w, "Admin role added successfully", nil, http.StatusOK)
}

func (ur *UserRoleController) AddSellerRole(w http.ResponseWriter, r *http.Request) {
	userID, err := validator.GetValidatedStruct[validator.FetchUserProfileValidator](r.Context())
	if err != nil {
		utils.ErrorResponseFunc(w, err.Error(), http.StatusBadRequest, err)
		return
	}
	if err := ur.RoleService.AddSellerRole(userID.ID); err != nil {
		utils.ErrorResponseFunc(w, err.Error(), http.StatusInternalServerError, err)
		return
	}

	utils.SuccessResponseFunc(w, "Seller role added successfully", nil, http.StatusOK)
}

func (ur *UserRoleController) SwitchRole(w http.ResponseWriter, r *http.Request) {
	userID, err := validator.GetValidatedStruct[validator.SwitchRoleValidator](r.Context())
	if err != nil {
		utils.ErrorResponseFunc(w, err.Error(), http.StatusBadRequest, err)
		return
	}
	userResp, err := ur.RoleService.SwitchRole(userID.UserID, userID.RoleID)
	if err != nil {
		utils.ErrorResponseFunc(w, err.Error(), http.StatusInternalServerError, err)
		return
	}

	utils.SuccessResponseFunc(w, "Role switched successfully", userResp, http.StatusOK)
}
