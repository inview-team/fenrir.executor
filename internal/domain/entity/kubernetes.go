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
	ListPodsByDeployment(ctx context.Context, namespace, deploymentName string) ([]*Pod, error)
	GetPodByName(ctx context.Context, namespace, name string) (*Pod, error)
	GetPodMetrics(ctx context.Context, namespace, name string) ([]*ContainerResources, error)
	GetDeploymentByName(ctx context.Context, namespace, name string) (*Deployment, error)
	Delete(ctx context.Context, namespace string, podName string) error
	Scale(ctx context.Context, namespace, deploymentName string, replicas int32) error
	GetPodLogs(ctx context.Context, namespace, podName, containerName string, tailLines int64) (string, error)
	DescribePod(ctx context.Context, namespace, podName string) (string, error)
	DescribeDeployment(ctx context.Context, namespace, deploymentName string) (string, error)
}
