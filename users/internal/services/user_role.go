// Package services for user-related operations.
package services

import (
	"e-commerce-backend/shared/utils"
	"e-commerce-backend/users/internal/models"
	"e-commerce-backend/users/pkg/constants"
	"e-commerce-backend/users/pkg/payloads"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"gorm.io/gorm"
)

type RoleService struct {
	DB *gorm.DB
}

func NewRoleService(db *gorm.DB) *RoleService {
	return &RoleService{
		DB: db,
	}
}

type RoleInterface interface {
	CreateRole(w http.ResponseWriter, r *http.Request)
	GetAllRoles(w http.ResponseWriter, r *http.Request)
	//DeleteRole(w http.ResponseWriter, r *http.Request)
	RequestRoleChange(w http.ResponseWriter, r *http.Request)
	ReviewRoleChange(w http.ResponseWriter, r *http.Request)
}

func (db *RoleService) CreateRole(w http.ResponseWriter, r *http.Request) {
	// Parse the request body for admin input
	var req struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JsonError(w, utils.InvalidRequestBodyError, http.StatusBadRequest, err)
		return
	}
	if req.Role == "" {
		utils.JsonError(w, utils.InvalidRoleError, http.StatusBadRequest, fmt.Errorf("role cannot be empty"))
		return
	}

	var role models.Role
	if err := role.CheckRoleExists(db.DB, req.Role); err != nil {
		if strings.Contains(err.Error(), "record not found") {
			// role.Role = req.Role
		} else {
			utils.JsonError(w, fmt.Sprintf(utils.RoleAlreadyExistsError, req.Role), http.StatusConflict, err)
			return
		}
	}

	// role.Role = req.Role
	if err := role.CreateRole(db.DB); err != nil {
		utils.JsonError(w, utils.RoleCreationError, http.StatusInternalServerError, err)
		return
	}
	utils.JsonResponse(role, w, fmt.Sprintf(utils.RoleCreatedSuccessfully, req.Role), http.StatusOK)
}

func (db *RoleService) GetAllRoles(w http.ResponseWriter, r *http.Request) {
	var role models.Role
	roles, err := role.GetAllRoles(db.DB)
	if err != nil {
		utils.JsonError(w, utils.RoleRetrievalError, http.StatusInternalServerError, err)
		return
	}
	utils.JsonResponse(roles, w, utils.RolesRetrievedSuccessfully, http.StatusOK)
}

func (db *RoleService) RequestRoleChange(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetIDFromPath(r)
	if err != nil {
		utils.JsonError(w, fmt.Sprintf(utils.InvalidUserIDError, userID), http.StatusBadRequest, err)
		return
	}

	// Parse the request body for admin input
	var req struct {
		RequestedRole string `json:"requested_role"`
		Comment       string `json:"comment"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JsonError(w, utils.InvalidRequestBodyError, http.StatusBadRequest, err)
		return
	}
	if req.RequestedRole == "" {
		utils.JsonError(w, utils.InvalidRoleError, http.StatusBadRequest, fmt.Errorf("role cannot be empty"))
		return
	}
	var userRole string
	if strings.ToLower(req.RequestedRole) == "admin" {
		userRole = constants.RoleAdmin
	} else if strings.ToLower(req.RequestedRole) == "seller" {
		userRole = constants.RoleSeller
	} else {
		userRole = constants.RoleUser
	}

	if err := models.RequestRoleChange(db.DB, userID, userRole, req.Comment); err != nil {
		utils.JsonError(w, err.Error(), http.StatusBadRequest, err)
		return
	}

	utils.JsonResponse(nil, w, fmt.Sprintf(utils.RoleChangeRequestCreated, userID, userRole), http.StatusOK)
}

func (db *RoleService) ReviewRoleChange(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from the path
	userID, err := utils.GetIDFromPath(r)
	if err != nil {
		utils.JsonError(w, fmt.Sprintf(utils.InvalidUserIDError, userID), http.StatusBadRequest, err)
		return
	}

	// Parse the request body for admin input
	var req struct {
		RequestedRole string `json:"requested_role"`
		Approve       bool   `json:"approve"`
		Comment       string `json:"comment"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JsonError(w, utils.InvalidRequestBodyError, http.StatusBadRequest, err)
		return
	}

	// Validate the input
	if req.RequestedRole == "" {
		utils.JsonError(w, utils.InvalidRoleError, http.StatusBadRequest, fmt.Errorf("role cannot be empty"))
		return
	}

	adminID := utils.GetUserIDFromContext(r)
	if adminID == 0 {
		utils.JsonError(w, utils.UnauthorizedError, http.StatusUnauthorized, fmt.Errorf("admin authorization required"))
		return
	}

	// Fetch the role change request
	request, err := models.GetUserRequestedRole(db.DB, userID, req.RequestedRole, constants.RoleStatusPending)
	if err != nil {
		if strings.Contains(err.Error(), gorm.ErrRecordNotFound.Error()) {
			utils.JsonError(w, fmt.Sprintf(utils.RoleChangeRequestNotFound, userID, req.RequestedRole), http.StatusNotFound, err)
			return
		}
		utils.JsonError(w, utils.InternalServerError, http.StatusInternalServerError, err)
		return
	}
	_ = request

	// Create the command to process the role change
	cmd := models.RoleChangeCommand{
		UserID:        userID,
		RequestedRole: req.RequestedRole,
		AdminID:       adminID,
		Approve:       req.Approve,
		Comment:       req.Comment,
		DB:            db.DB,
	}

	// Execute the command
	if err := cmd.Execute(); err != nil {
		utils.JsonError(w, utils.RoleChangeFailedError, http.StatusInternalServerError, err)
		return
	}

	// Prepare a success response
	statusMessage := "Role change request processed successfully"
	if !req.Approve {
		statusMessage = "Role change request rejected"
	}
	utils.JsonResponse(map[string]interface{}{
		"user_id":        userID,
		"requested_role": req.RequestedRole,
		"approved":       req.Approve,
	}, w, statusMessage, http.StatusOK)
}

func (db *RoleService) SwitchRole(w http.ResponseWriter, r *http.Request) {
	token := utils.GetTokenFromRequestHeader(r)
	if token == "" {
		utils.JsonError(w, utils.UnauthorizedError, http.StatusUnauthorized, errors.New(utils.MissingTokenError))
		return
	}

	userID := utils.GetUserIDFromContext(r)
	if userID == 0 {
		utils.JsonError(w, utils.UserIdNotFoundInCtx, http.StatusBadRequest, errors.New(utils.UserIdNotFoundInCtx))
		return
	}

	currentUsername := utils.GetUserNameIDFromContext(r)
	if currentUsername == "" {
		utils.JsonError(w, utils.UserNameNotFoundInCtx, http.StatusBadRequest, errors.New(utils.UserNameNotFoundInCtx))
		return
	}

	var requestBody struct {
		Username string `json:"username" validate:"required,min=3,max=20"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		utils.JsonError(w, utils.InvalidRequestBodyError, http.StatusBadRequest, err)
		return
	}
	if ok, err := utils.ValidateStructUsingValidators(requestBody); !ok {
		errStr := strings.Join(err, ", ")
		utils.JsonError(w, utils.InvalidUserDataError, http.StatusBadRequest, errors.New(errStr))
		return
	}

	if requestBody.Username == currentUsername {
		utils.JsonError(w, utils.ExpectedAnotherUsername, http.StatusBadRequest, fmt.Errorf(strings.TrimRight(fmt.Sprintf(utils.SameUsernameError, requestBody.Username), ".!")))
		return
	}

	var userService models.User
	userResponse, err := userService.GetUserUsingUsername(db.DB, requestBody.Username)
	if err != nil {
		utils.JsonError(w, fmt.Sprintf(utils.UsernameNotFoundError, requestBody.Username), http.StatusInternalServerError, err)
		return
	}

	expireTime := time.Now().Add(time.Hour * 24).Unix()
	newToken, err := utils.GenerateJWT(userResponse.ID, userResponse.Email, requestBody.Username, expireTime)
	if err != nil {
		utils.JsonError(w, utils.TokenGenerationError, http.StatusInternalServerError, err)
		return
	}

	authResponse := payloads.UserAuthResponse{
		Token:      newToken,
		ExpireTime: expireTime,
		User:       *userResponse,
	}
	utils.JsonResponse(authResponse, w, fmt.Sprintf(utils.UserLoggedInSuccessfully, userResponse.ID), http.StatusOK)
}
