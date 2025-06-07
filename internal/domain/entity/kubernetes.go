package entity

import "context"

type Pod struct {
	Name   string
	Status string
}

type Deployment struct {
	Name     string
	Replicas int32
}

func NewPod(name, status string) *Pod {
	return &Pod{
		Name:   name,
		Status: status,
	}
}

type KubernetesRepository interface {
	List(ctx context.Context, namespace string) ([]*Pod, error)
	GetPodByName(ctx context.Context, namespace, name string) (*Pod, error)
	GetDeploymentByName(ctx context.Context, namespace, name string) (*Deployment, error)
	Delete(ctx context.Context, namespace string, podName string) error
	Scale(ctx context.Context, namespace, deploymentName string, replicas int32) error
}
