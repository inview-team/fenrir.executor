package application

import (
	"context"

	"github.com/inviewteam/fenrir.executor/internal/domain/service"
	"github.com/inviewteam/fenrir.executor/internal/infrastructure/kuber"
	"k8s.io/client-go/rest"
)

type Application struct {
	ExecutorService *service.Executor
}

func New(ctx context.Context, kubeConfig *rest.Config) (*Application, error) {
	kRepo, err := kuber.New(kubeConfig)
	if err != nil {
		return nil, err
	}
	return &Application{
		service.New(kRepo),
	}, nil
}
