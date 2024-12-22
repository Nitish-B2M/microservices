package handlers

import (
	"microservices/pkg/services/user"
	"microservices/pkg/shared/dbs"
	"net/http"
)

func UserHandler() {
	userService := &user.Service{DB: dbs.DB}

	http.HandleFunc("/users", userService.GetAllUsers)
	http.HandleFunc("/user/{id}", userService.FetchUserById)
	http.HandleFunc("/user/add", userService.AddUser)
	http.HandleFunc("/user/update/{id}", userService.UpdateUser)
	http.HandleFunc("/user/delete/{id}", userService.DeleteUser)
	http.HandleFunc("/user/login", userService.LoginUser)
	http.HandleFunc("/user/password/reset", userService.ResetPassword)
	http.HandleFunc("/user/password/reset/request", userService.RequestPasswordReset)
	http.HandleFunc("/user/verify/send", userService.SendVerificationEmail)
	http.HandleFunc("/user/verify/{token}", userService.VerifyUserEmail)

	// http.Handle("/user/profile", middlewares.JWTMiddleware(http.HandlerFunc(db.UserService.FetchUserById)))
	// http.Handle("/user/list", middlewares.JWTMiddleware(http.HandlerFunc(db.UserService.GetAllUsers)))
}
