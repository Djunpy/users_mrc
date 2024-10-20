package usecase

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgtype"
	db "users_mrc/db/sqlc"
	"users_mrc/entities"
	"users_mrc/helpers/crypto"
	"users_mrc/helpers/server"
	"users_mrc/infrastructure/jwt_token"
)

type AuthUsecase struct {
	store      db.Store
	tokenMaker jwt_token.Maker
}

func NewAuthUsecase(store db.Store, tokenMaker jwt_token.Maker) AuthUsecase {
	return AuthUsecase{store: store, tokenMaker: tokenMaker}
}

func GetUserRoles(groups []db.GetGroupsByUserIdRow) []string {
	var result []string
	for _, group := range groups {
		result = append(result, group.GroupName)
	}
	return result
}

func (uc *AuthUsecase) EmailConfirmation(ctx context.Context, token string) (int32, error) {
	tokenPayload, err := uc.tokenMaker.VerifyToken(token)
	if err != nil {
		return uc.tokenMaker.GetErrorCode(err), err
	}
	updateParam := &db.UpdateUserByEmailParams{
		VerifiedEmail: pgtype.Bool{Bool: true, Valid: true},
		Email:         tokenPayload.Email,
	}
	_, err = uc.store.UpdateUserByEmail(ctx, *updateParam)
	if err != nil {
		return db.ErrorCode(err), err
	}
	return server.SUCCESS_CODE, nil
}

func (uc *AuthUsecase) CreateUser(ctx context.Context, args *db.CreateOrdinaryUserTxParams) (token string, statusCode int32, err error) {
	userExists, err := uc.store.UserExists(ctx, args.Email)
	if err != nil {
		errCode := db.ErrorCode(err)
		return token, errCode, err
	}
	if userExists {
		return token, server.USER_EXISTS_ERR_CODE, fmt.Errorf("user with email %s already exists", args.Email)
	}
	err = uc.store.TxCreateUser(ctx, args)
	if err != nil {
		errCode := db.ErrorCode(err)
		return token, errCode, err
	}
	user := entities.UserResponse{
		Email:  args.Email,
		UserId: -1,
	}
	token, _, err = uc.tokenMaker.CreateToken(user, "access")
	return token, server.SUCCESS_CODE, nil
}

func (uc *AuthUsecase) CreateAccessAndRefreshToken(user db.User, groups []db.GetGroupsByUserIdRow, tokenType string) (string, *jwt_token.Payload, int32, error) {
	var payload *jwt_token.Payload
	var err error
	var tokenStr string
	groupsNames := GetUserRoles(groups)
	userType := user.UserType.UserTypes
	userResp := entities.UserResponse{
		UserId:   user.ID,
		Email:    user.Email,
		Groups:   groupsNames,
		UserType: string(userType),
	}
	tokenStr, payload, err = uc.tokenMaker.CreateToken(userResp, tokenType)
	if err != nil {
		return tokenStr, payload, uc.tokenMaker.GetErrorCode(err), err

	}
	return tokenStr, payload, server.SUCCESS_CODE, nil
}

func (uc *AuthUsecase) GetUser(
	ctx context.Context, args *entities.SignInReq) (user db.User, groups []db.GetGroupsByUserIdRow, statusCode int32, err error) {
	user, err = uc.store.GetUserByEmail(ctx, args.Login)
	if err != nil {
		return user, groups, db.ErrorCode(err), err
	}
	groups, err = uc.store.GetGroupsByUserId(ctx, user.ID)
	if err != nil {
		return user, groups, db.ErrorCode(err), err
	}

	err = crypto.ComparePassword(user.Password, args.Password)
	if err != nil {
		return user, groups, server.INCORRECT_PASSWORD_ERR_CODE, err
	}

	return user, groups, server.SUCCESS_CODE, nil
}

func (uc *AuthUsecase) RefreshAccessToken(
	ctx context.Context, refreshToken string) (accessToken string, statusCode int32, err error) {
	sub, err := uc.tokenMaker.VerifyToken(refreshToken)
	var groupsNames []string
	if err != nil {
		return "", server.TOKEN_VALIDATION_ERR_CODE, err
	}
	user, err := uc.store.GetUserByEmail(ctx, sub.Email)
	if err != nil {
		statusCode = db.ErrorCode(err)
		return "", statusCode, err
	}
	groups, err := uc.store.GetGroupsByUserId(ctx, user.ID)
	if err != nil {
		statusCode = db.ErrorCode(err)
		return "", statusCode, err
	}
	for _, group := range groups {
		groupsNames = append(groupsNames, group.GroupName)
	}
	userResp := entities.UserResponse{
		UserId: user.ID,
		Email:  user.Email,
		Groups: groupsNames,
	}
	accessToken, _, err = uc.tokenMaker.CreateToken(userResp, "access")
	if err != nil {
		return "", server.GENERATE_JWT_TOKEN_ERR_CODE, err
	}
	err = uc.store.LastTokenUpdate(ctx, user.ID)
	if err != nil {
		return "", server.GENERATE_JWT_TOKEN_ERR_CODE, err
	}
	return accessToken, server.SUCCESS_CODE, nil
}

func (uc *AuthUsecase) ChangePassword(
	ctx context.Context,
	payload *entities.ChangePasswordReq,
	userId int64,
) (statusCode int32, err error) {
	user, err := uc.store.GetUserById(ctx, userId)
	if err != nil {
		statusCode = db.ErrorCode(err)
		return statusCode, err
	}
	err = crypto.ComparePassword(user.Password, payload.OldPassword)
	if err != nil {
		return server.INCORRECT_PASSWORD_ERR_CODE, err
	}
	hashPass := crypto.HashPassword(payload.NewPassword)

	changePassArgs := db.ChangePasswordParams{
		Password: pgtype.Text{String: hashPass, Valid: true},
		ID:       userId,
	}
	err = uc.store.ChangePassword(ctx, changePassArgs)
	if err != nil {
		statusCode = db.ErrorCode(err)
		return statusCode, err
	}
	return server.SUCCESS_CODE, nil
}
