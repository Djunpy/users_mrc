package worker

import (
	"context"
	"github.com/hibiken/asynq"
	"users_mrc/entities"
)

type TaskDistributor interface {
	DistributeTaskSendVerifyEmail(
		ctx context.Context,
		payload *entities.PayloadSendVerifyEmail,
		opts ...asynq.Option,
	) error
}

type TaskProcessor interface {
	Start() error
	Shutdown()
	ProcessTaskSendVerifyEmail(ctx context.Context, task *asynq.Task) error
}
