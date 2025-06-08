package service

import (
	"context"
	"fmt"
	"time"

	"github.com/inviewteam/fenrir.executor/internal/domain/entity"
	log "github.com/sirupsen/logrus"
)

type Executor struct {
	kubeRepo entity.KubernetesRepository
}

func New(pRepo entity.KubernetesRepository) *Executor {
	return &Executor{
		kubeRepo: pRepo,
	}
}

func (s *Executor) Restart(ctx context.Context, namespace, podName string) error {
	log.Infof("Restart pod %s", podName)
	err := s.kubeRepo.Delete(ctx, namespace, podName)
	if err != nil {
		return err
	}

	for {
		_, err := s.kubeRepo.GetPodByName(ctx, namespace, podName)
		if err != nil {
			if err == ErrPodNotFound {
				break
			}
			return err
		}
		log.Infof("Wait when pod %s restart", podName)
		time.Sleep(5 * time.Second)
	}
	return nil
}

func (s *Executor) Scale(ctx context.Context, namespace, deploymentName string, targetReplicas int32) error {
	log.Infof("Scale deployment %s to replicas %d", deploymentName, targetReplicas)
	deployment, err := s.kubeRepo.GetDeploymentByName(ctx, namespace, deploymentName)
	if err != nil {
		return fmt.Errorf("failed to scale: %w", err)
	}
	if deployment == nil {
		return fmt.Errorf("failed to scale: %w", ErrDeploymentNotFound)
	}
	err = s.kubeRepo.Scale(ctx, namespace, deploymentName, targetReplicas)
	if err != nil {
		return fmt.Errorf("failed to scale: %w", err)
	}

	for {
		deployment, err := s.kubeRepo.GetDeploymentByName(ctx, namespace, deploymentName)
		if err != nil {
			return fmt.Errorf("failed to scale: %w", err)
		}

		if deployment.Replicas == targetReplicas {
			break
		}

		log.Infof("Wait until deployment %s end scalling", deploymentName)
		time.Sleep(5 * time.Second)
	}
	return nil
}

func (s *Executor) GetPodByName(ctx context.Context, namespace string, podName string) (*entity.Pod, error) {
	pod, err := s.kubeRepo.GetPodByName(ctx, namespace, podName)
	if err != nil {
		return nil, fmt.Errorf("failed to get pod by name: %w", err)
	}
	return pod, nil
}
