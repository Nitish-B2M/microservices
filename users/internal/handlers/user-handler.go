// Package handlers provides HTTP handler functions for user-related endpoints.
package handlers

import (
	mw "e-commerce-backend/shared/middlewares"
	"e-commerce-backend/users/dbs"
	"e-commerce-backend/users/internal/controller"
	"e-commerce-backend/users/internal/repository"
	"e-commerce-backend/users/internal/services"
	"net/http"

	"fmt"

	"github.com/gorilla/mux"
)

func UserHandler(r *mux.Router) {

	//r.Handle("/user/role/{id}", mw.AuthMiddleware(http.HandlerFunc(roleService.GetRole))).Methods(http.MethodGet)
	//r.Handle("/user/role/update/{id}", mw.AuthMiddleware(http.HandlerFunc(roleService.UpdateRole))).Methods(http.MethodPut)
	//r.Handle("/user/role/delete/{id}", mw.AuthMiddleware(http.HandlerFunc(roleService.DeleteRole))).Methods(http.MethodDelete)
}

func AuthRoutes(r *mux.Router) {
	userService := services.NewUser(dbs.UserDB)
	userRepo := repository.NewUserRepo(dbs.UserDB)
	svc := services.NewService(userRepo)
	userController := controller.NewUserController(svc)

	mw.HandlePost(r, "/login", "LoginUserValidator", userController.LoginUser)
	mw.HandlePost(r, "/register", "CreateUserValidator", userController.CreateUser)
	mw.HandlePost(r, "/email-verification", "EmailVerificationValidator", userController.VerifyUserEmail)
	mw.HandlePost(r, "/email-verification/request", "SendVerificationEmailValidator", userController.SendVerificationEmail)
	mw.HandlePost(r, "/password-reset", "", userService.ResetPassword)
	mw.HandlePost(r, "/password-reset-request", "", userService.RequestPasswordReset)
	// mw.HandlePost(r, "/email-verification/request", "", userService.SendVerificationEmail)
	mw.HandleGet(r, "/email-verification/{token}", "", userService.VerifyUserEmail)
	mw.HandleDelete(r, "/delete-account", "", userService.DeleteUser, mw.AuthMiddleware)
	mw.HandlePut(r, "/activate-account/{id}", "", userService.ActivateUser, mw.AuthMiddleware)
	mw.HandlePut(r, "/deactivate-account/{id}", "", userService.DeActivateUser, mw.AuthMiddleware)
}

func ProfileRoutes(r *mux.Router) {
	userService := services.NewUser(dbs.UserDB)
	userRepo := repository.NewUserRepo(dbs.UserDB)
	svc := services.NewService(userRepo)
	userController := controller.NewUserController(svc)

	mw.HandlePost(r, "", "FetchUserProfileValidator", userController.FetchUserProfileByID, mw.AuthMiddleware)
	mw.HandlePost(r, "/user/{id}", "", userService.GetUserProfile, mw.AuthMiddleware)
	mw.HandlePut(r, "/update", "", userService.UpdateUser, mw.AuthMiddleware)
}

func AddressRoutes(r *mux.Router) {
	addressService := services.NewAdrServices(dbs.UserDB)

	mw.HandlePost(r, "/add", "", addressService.AddAddress, mw.AuthMiddleware)
	mw.HandlePost(r, "/list", "", addressService.GetAddressByUserID, mw.AuthMiddleware)
	mw.HandlePut(r, "/update/{id}", "", addressService.UpdateAddress, mw.AuthMiddleware)
	mw.HandlePut(r, "/set-primary/{id}", "", addressService.SetPrimaryAddress, mw.AuthMiddleware)
	mw.HandleDelete(r, "/delete/{id}", "", addressService.DeleteAddress, mw.AuthMiddleware)
}

func RoleRoutes(r *mux.Router) {
	roleService := services.NewRoleService(dbs.UserDB)
	userRole := repository.NewUserRole(dbs.UserDB)
	roleSvc := services.NewRolesService(userRole)
	userRoleController := controller.NewUserRoleController(roleSvc)

	mw.HandleGet(r, "/list", "", roleService.GetAllRoles, mw.AuthMiddleware)
	mw.HandlePost(r, "/add", "", roleService.CreateRole, mw.AuthMiddleware)
	mw.HandlePost(r, "/switch-role", "SwitchRoleValidator", userRoleController.SwitchRole, mw.AuthMiddleware)
	mw.HandlePost(r, "/review-role", "", roleService.ReviewRoleChange, mw.AuthMiddleware)
	mw.HandlePost(r, "/request-role", "", roleService.RequestRoleChange, mw.AuthMiddleware)
	mw.HandlePost(r, "/add-admin", "FetchUserProfileValidator", userRoleController.AddAdminRole, mw.AuthMiddleware, mw.RoleMiddleware(dbs.UserDB, "admin"))
	mw.HandlePost(r, "/add-seller", "FetchUserProfileValidator", userRoleController.AddSellerRole, mw.AuthMiddleware, mw.RoleMiddleware(dbs.UserDB, "admin"))
}

func AdminRoutes(r *mux.Router) {
	userService := services.NewUser(dbs.UserDB)

	mw.HandleGet(r, "/admin", "", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Admin")
	}, mw.AuthMiddleware, mw.RoleMiddleware(dbs.UserDB, "admin"))
	mw.HandleGet(r, "/users/all", "", userService.GetAllUsers, mw.AuthMiddleware, mw.RoleMiddleware(dbs.UserDB, "admin"))
}
