package handlers

import (
	"e-commerce-backend/shared/middlewares"
	"e-commerce-backend/users/dbs"
	"e-commerce-backend/users/internal/services"
	"net/http"

	"github.com/gorilla/mux"
)

func UserHandler(r *mux.Router) {
	userService := services.NewUser(dbs.UserDB)
	addressService := services.NewAdrServices(dbs.UserDB)

	r.HandleFunc("/user/signup", userService.CreateUser).Methods(http.MethodPost)
	//here need to add role authentication
	r.Handle("/users", middlewares.AuthMiddleware(http.HandlerFunc(userService.GetAllUsers))).Methods(http.MethodGet)
	r.Handle("/user/profile", middlewares.AuthMiddleware(http.HandlerFunc(userService.GetUserProfile))).Methods(http.MethodGet)
	r.Handle("/user/{id}", middlewares.AuthMiddleware(http.HandlerFunc(userService.GetUserProfile))).Methods(http.MethodGet)
	r.Handle("/user/profile/update", middlewares.AuthMiddleware(http.HandlerFunc(userService.UpdateUser))).Methods(http.MethodPut)
	r.Handle("/user/delete/{id}", middlewares.AuthMiddleware(http.HandlerFunc(userService.DeleteUser))).Methods(http.MethodDelete)
	r.Handle("/user/activate/{id}", middlewares.AuthMiddleware(http.HandlerFunc(userService.ActivateUser))).Methods(http.MethodPut)
	r.Handle("/user/deactivate/{id}", middlewares.AuthMiddleware(http.HandlerFunc(userService.DeActivateUser))).Methods(http.MethodPut)
	r.HandleFunc("/user/password/reset", userService.ResetPassword).Methods(http.MethodPost)
	r.HandleFunc("/user/password/reset/request", userService.RequestPasswordReset).Methods(http.MethodPost)
	r.HandleFunc("/user/login", userService.LoginUser).Methods(http.MethodPost)
	r.HandleFunc("/user/verify/send", userService.SendVerificationEmail).Methods(http.MethodPost)
	r.HandleFunc("/user/verify/{token}", userService.VerifyUserEmail).Methods(http.MethodGet)
	r.Handle("/user/address/add", middlewares.AuthMiddleware(http.HandlerFunc(addressService.AddAddress))).Methods(http.MethodPost)
	r.Handle("/user/address/all", middlewares.AuthMiddleware(http.HandlerFunc(addressService.GetAddressByUserId))).Methods(http.MethodGet)
	r.Handle("/user/address/delete/{id}", middlewares.AuthMiddleware(http.HandlerFunc(addressService.DeleteAddress))).Methods(http.MethodDelete)
	r.Handle("/user/address/update/{id}", middlewares.AuthMiddleware(http.HandlerFunc(addressService.UpdateAddress))).Methods(http.MethodPatch)
}
