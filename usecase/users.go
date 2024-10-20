package usecase

import (
	"context"
	db "users_mrc/db/sqlc"
	"users_mrc/entities"
	"users_mrc/helpers/server"
	"users_mrc/helpers/users"
	"users_mrc/infrastructure/jwt_token"
)

type UsersUsecase struct {
	store db.Store
}

func NewUsersUsecase(store db.Store) UsersUsecase {
	return UsersUsecase{store}
}

func (au *UsersUsecase) GetUserDetail(
	ctx context.Context, jwtPayload jwt_token.Payload) (accountDetail entities.UserDetail, statusCode int32, err error) {
	user, err := au.store.GetUserAndGroupsByEmail(ctx, jwtPayload.Email)
	if err != nil {
		return accountDetail, db.ErrorCode(err), err
	}
	var groups []string
	groups, err = users.ExtractGroups(user.Groups)
	if err != nil {
		return accountDetail, db.ErrorCode(err), err

	}
	roles := users.RemoveSAtEnd(groups)
	phone, err := au.store.GetUserPhoneByUserId(ctx, user.ID)
	if err != nil {
		return accountDetail, db.ErrorCode(err), err
	}
	accountDetail = entities.NewUserDetailResponse(user, phone, roles)
	return accountDetail, server.SUCCESS_CODE, nil
}
