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
