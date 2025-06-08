package entity

import (
	"context"
	"time"
)

type Pod struct {
	Name      string
	Status    string
	Restarts  int
	Age       time.Duration
	Resources []*ContainerResources
}

type ContainerResources struct {
	Name         string
	CpuUsage     int64
	MemoryUsage  int64
	CpuLimits    int64
	MemoryLimits int64
}

type Deployment struct {
	Name     string
	Replicas int32
}

func NewPod(name, status string, restarts int, age time.Duration, resources []*ContainerResources) *Pod {
	return &Pod{
		Name:      name,
		Status:    status,
		Restarts:  restarts,
		Age:       age,
		Resources: resources,
	}
}

type KubernetesRepository interface {
	List(ctx context.Context, namespace string) ([]*Pod, error)
	GetPodByName(ctx context.Context, namespace, name string) (*Pod, error)
	GetDeploymentByName(ctx context.Context, namespace, name string) (*Deployment, error)
	Delete(ctx context.Context, namespace string, podName string) error
	Scale(ctx context.Context, namespace, deploymentName string, replicas int32) error
}
