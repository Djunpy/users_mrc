package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
	"net/http"
	"time"
	db "users_mrc/db/sqlc"
	"users_mrc/entities"
	"users_mrc/helpers/server"
	"users_mrc/infrastructure/jwt_token"
	"users_mrc/infrastructure/worker"
	"users_mrc/usecase"
)

type AuthController struct {
	usecase         usecase.AuthUsecase
	taskDistributor worker.TaskDistributor
}

func NewAuthController(usecase usecase.AuthUsecase, taskDistributor worker.TaskDistributor) AuthController {
	return AuthController{usecase: usecase, taskDistributor: taskDistributor}
}

func (ac *AuthController) ChangePassword(ctx *gin.Context) {
	var payload *entities.ChangePasswordReq
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, server.Response(err, server.INVALID_DATA_ERR_CODE, nil))
		return
	}
	jwtPayload, exists := jwt_token.GetJWTPayload(ctx)
	if !exists {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, server.Response(nil, server.UNKNOWN_ERROR_CODE, nil))
		return
	}
	errCode, err := ac.usecase.ChangePassword(ctx, payload, jwtPayload.UserId)
	if err != nil {
		server.HandlerErr(ctx, errCode, err)
		return
	}
	ctx.JSON(http.StatusOK, server.Response(nil, errCode, nil))
}

func (ac *AuthController) RefreshAccessToken(ctx *gin.Context) {
	var payload *entities.RefreshTokenReq
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, server.Response(err, server.INVALID_DATA_ERR_CODE, nil))
		return
	}

	accessToken, errCode, err := ac.usecase.RefreshAccessToken(ctx, payload.RefreshToken)
	if err != nil {
		server.HandlerErr(ctx, errCode, err)
		return
	}
	response := gin.H{"access_token": accessToken}
	ctx.JSON(http.StatusOK, server.Response(nil, errCode, response))
}

func (ac *AuthController) SignUpUser(ctx *gin.Context) {
	var payload *db.CreateOrdinaryUserTxParams
	fmt.Print(payload, "==================")
	var err error
	if err = ctx.ShouldBindJSON(&payload); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, server.Response(err, server.INVALID_DATA_ERR_CODE, nil))
		return
	}
	token, errCode, err := ac.usecase.CreateUser(
		ctx,
		payload,
	)
	if err != nil {
		server.HandlerErr(ctx, errCode, err)
		return
	}

	taskPayload := &entities.PayloadSendVerifyEmail{
		Email:    payload.Email,
		LangCode: "ru",
		JWTToken: token,
	}
	opts := []asynq.Option{
		asynq.MaxRetry(10),
		asynq.ProcessIn(10 * time.Second),
		asynq.Queue(worker.QueueCritical),
	}
	err = ac.taskDistributor.DistributeTaskSendVerifyEmail(ctx, taskPayload, opts...)

	if err != nil {
		log.Info().Err(err).Msg(fmt.Sprintf("distribute task send verify email err: %v", err))
	}
	ctx.JSON(http.StatusOK, server.Response(nil, errCode, nil))
}

func (ac *AuthController) EmailConfirmation(ctx *gin.Context) {
	token := ctx.Query("token")
	if token == "" {
		ctx.JSON(http.StatusBadRequest, server.Response(nil, server.INVALID_URL_PARAM_ERR_CODE, nil))
		return
	}
	errCode, err := ac.usecase.EmailConfirmation(ctx, token)
	if err != nil {
		server.HandlerErr(ctx, errCode, err)
		return
	}
	ctx.JSON(http.StatusOK, server.Response(nil, 0, nil))
}

func (ac *AuthController) SignInUser(ctx *gin.Context) {
	var payload *entities.SignInReq

	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, server.Response(err, server.INVALID_DATA_ERR_CODE, nil))
		return
	}
	user, groups, errCode, err := ac.usecase.GetUser(ctx, payload)
	if err != nil {
		server.AuthHandlerErr(ctx, errCode, err)
		return
	}
	accessToken, _, errCode, err := ac.usecase.CreateAccessAndRefreshToken(user, groups, "access")
	if err != nil {
		server.HandlerErr(ctx, errCode, err)
		return
	}
	refreshToken, _, errCode, err := ac.usecase.CreateAccessAndRefreshToken(user, groups, "refresh")
	if err != nil {
		server.HandlerErr(ctx, errCode, err)
		return
	}
	response := gin.H{"access_token": accessToken, "refresh_token": refreshToken}
	ctx.JSON(http.StatusOK, server.Response(nil, errCode, response))
}
