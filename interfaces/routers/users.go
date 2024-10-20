package routers

import (
	"github.com/gin-gonic/gin"
	"users_mrc/interfaces/controllers"
)

type UsersRouter struct {
	controller controllers.UsersController
}

func NewUsersRouter(controller controllers.UsersController) *UsersRouter {
	return &UsersRouter{controller: controller}
}

func (r *UsersRouter) InitAuthRouter(public *gin.RouterGroup, private *gin.RouterGroup) {
	private.GET("/user-detail", r.controller.GetUserDetail)
}
