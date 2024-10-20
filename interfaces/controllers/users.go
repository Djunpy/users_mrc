package controllers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"users_mrc/helpers/server"
	"users_mrc/infrastructure/jwt_token"
	"users_mrc/usecase"
)

type UsersController struct {
	usecase usecase.UsersUsecase
}

func NewUsersController(usecase usecase.UsersUsecase) UsersController {
	return UsersController{
		usecase: usecase,
	}
}

func (c *UsersController) GetUserDetail(ctx *gin.Context) {
	jwtPayload, exists := jwt_token.GetJWTPayload(ctx)
	if !exists {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, server.Response(nil, server.UNKNOWN_ERROR_CODE, nil))
		return
	}
	accountDetail, errCode, err := c.usecase.GetUserDetail(ctx, *jwtPayload)
	if err != nil {
		server.HandlerErr(ctx, errCode, err)
		return
	}
	ctx.JSON(http.StatusOK, server.Response(nil, errCode, accountDetail))
}
