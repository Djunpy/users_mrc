package routers

import (
	"github.com/gin-gonic/gin"
	"users_mrc/interfaces/controllers"
)

type AuthRouter struct {
	controller controllers.AuthController
}

func NewAuthRouter(controller controllers.AuthController) *AuthRouter {
	return &AuthRouter{controller: controller}
}

func (r *AuthRouter) InitAuthRouter(public *gin.RouterGroup, private *gin.RouterGroup) {
	public.POST("/sign-up", r.controller.SignUpUser)
	public.GET("/email-confirmation", r.controller.EmailConfirmation)
	public.POST("/sign-in", r.controller.SignInUser)
	public.POST("/refresh-token", r.controller.RefreshAccessToken)
}
