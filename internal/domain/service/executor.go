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

func (s *Executor) ListPodByDeployment(ctx context.Context, namespace, deploymentName string) ([]*entity.Pod, error) {
	pods, err := s.kubeRepo.ListPodsByDeployment(ctx, namespace, deploymentName)
	if err != nil {
		return nil, fmt.Errorf("failed to list pods by deployment: %w", err)
	}

	return pods, nil
}

func (s *Executor) GetPodByName(ctx context.Context, namespace string, podName string) (*entity.Pod, error) {
	pod, err := s.kubeRepo.GetPodByName(ctx, namespace, podName)
	if err != nil {
		return nil, fmt.Errorf("failed to get pod by name: %w", err)
	}
	containers, err := s.kubeRepo.GetPodContainers(ctx, namespace, podName)
	if err != nil {
		log.Errorf("failed to get pod metrics: %w", err)
	}
	pod.Containers = containers
	return pod, nil
}

func (s *Executor) GetDeploymentByName(ctx context.Context, namespace, deploymentName string) (*entity.Deployment, error) {
	deployment, err := s.kubeRepo.GetDeploymentByName(ctx, namespace, deploymentName)
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment: %w", err)
	}
	return deployment, nil
}

func (s *Executor) GetPodLogs(ctx context.Context, namespace, podName, containerName string, tailLines int64) (string, error) {
	logs, err := s.kubeRepo.GetPodLogs(ctx, namespace, podName, containerName, tailLines)
	if err != nil {
		return "", fmt.Errorf("failed to get pod logs: %w", err)
	}
	return logs, nil
}

func (s *Executor) DescribePod(ctx context.Context, namespace, podName string) (string, error) {
	desc, err := s.kubeRepo.DescribePod(ctx, namespace, podName)
	if err != nil {
		return "", fmt.Errorf("failed to describe pod: %w", err)
	}
	return desc, nil
}

func (s *Executor) DescribeDeployment(ctx context.Context, namespace, deploymentName string) (string, error) {
	desc, err := s.kubeRepo.DescribeDeployment(ctx, namespace, deploymentName)
	if err != nil {
		return "", fmt.Errorf("failed to describe deployment: %w", err)
	}
	return desc, nil
}
