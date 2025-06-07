package service

import "github.com/inviewteam/fenrir.executor/cmd/internal/domain/entity"

type Executor struct {
	podRepo entity.PodRepository
}

func New(pRepo *entity.PodRepository) *Executor {
	return &Executor{
		podRepo: *pRepo,
	}
}
