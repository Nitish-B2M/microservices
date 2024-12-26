package handlers

import (
	"e-commerce-backend/users/dbs"
	"e-commerce-backend/users/internal/services/user"
	"net/http"

	"github.com/gorilla/mux"
)

func UserHandler(r *mux.Router) {
	userService := user.NewUser(dbs.UserDB)

	r.HandleFunc("/users", userService.GetAllUsers).Methods(http.MethodGet)
	r.HandleFunc("/user/{id}", userService.FetchUserById).Methods(http.MethodGet)
	r.HandleFunc("/user/add", userService.AddUser).Methods(http.MethodPost)
	r.HandleFunc("/user/update/{id}", userService.UpdateUser).Methods(http.MethodPut)
	r.HandleFunc("/user/delete/{id}", userService.DeleteUser).Methods(http.MethodDelete)
	r.HandleFunc("/user/activate/{id}", userService.ActivateUser).Methods(http.MethodPut)
	r.HandleFunc("/user/deactivate/{id}", userService.DeActivateUser).Methods(http.MethodPut)
	r.HandleFunc("/user/login", userService.LoginUser).Methods(http.MethodPost)
	r.HandleFunc("/user/password/reset", userService.ResetPassword).Methods(http.MethodPost)
	r.HandleFunc("/user/password/reset/request", userService.RequestPasswordReset).Methods(http.MethodPost)
	r.HandleFunc("/user/verify/send", userService.SendVerificationEmail).Methods(http.MethodPost)
	r.HandleFunc("/user/verify/{token}", userService.VerifyUserEmail).Methods(http.MethodGet)
}
