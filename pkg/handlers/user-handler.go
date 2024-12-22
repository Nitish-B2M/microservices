package handlers

import (
	"microservices/pkg/services/user"
	"microservices/pkg/shared/dbs"
	"net/http"
)

func UserHandler() {
	userService := &user.UserService{DB: dbs.DB}

	http.HandleFunc("/users", userService.GetAllUsers)
	http.HandleFunc("/users/{id}", userService.FetchUserById)
	http.HandleFunc("/users/add", userService.AddUser)
	http.HandleFunc("/users/update/{id}", userService.UpdateUser)
	http.HandleFunc("/users/delete/{id}", userService.DeleteUser)
	http.HandleFunc("/users/login", userService.LoginUser)
	// http.HandleFunc("/user/password/reset/request", userService.RequestPasswordReset)
	// http.Handle("/user/profile", middlewares.JWTMiddleware(http.HandlerFunc(db.UserService.FetchUserById)))
	// http.Handle("/user/list", middlewares.JWTMiddleware(http.HandlerFunc(db.UserService.GetAllUsers)))
}
