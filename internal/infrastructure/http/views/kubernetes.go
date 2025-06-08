package views

import (
	"github.com/inviewteam/fenrir.executor/internal/domain/entity"
)

type Pod struct {
	Name       string       `json:"name"`
	Status     string       `json:"status"`
	Restarts   int          `json:"restarts"`
	Age        string       `json:"age"`
	Containers []*Container `json:"containers"`
}

type Container struct {
	Name         string `json:"name"`
	State        string `json:"state"`
	CpuUsage     int64  `json:"cpuUsage"`
	MemoryUsage  int64  `json:"memoryUsage"`
	CpuLimits    int64  `json:"cpuLimits"`
	MemoryLimits int64  `json:"memoryLimits"`
}

func NewPod(e *entity.Pod) *Pod {
	return &Pod{
		Name:     e.Name,
		Status:   e.Status,
		Restarts: e.Restarts,
		Age:      e.Age.String(),
		Containers: func() []*Container {
			res := make([]*Container, 0, len(e.Containers))
			for _, cr := range e.Containers {
				res = append(res, &Container{
					Name:         cr.Name,
					State:        cr.State,
					CpuUsage:     cr.CpuUsage,
					MemoryUsage:  cr.MemoryUsage,
					CpuLimits:    cr.CpuLimits,
					MemoryLimits: cr.MemoryLimits,
				})
			}
			return res
		}(),
	}
}

type DeploymentPods struct {
	Pods []DeploymentPod `json:"pods"`
}

type DeploymentPod struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

func NewPods(podEntities []*entity.Pod) *DeploymentPods {
	pods := make([]DeploymentPod, 0, len(podEntities))
	for _, p := range podEntities {
		pods = append(pods, DeploymentPod{
			Name:   p.Name,
			Status: p.Status,
		})
	}
	return &DeploymentPods{Pods: pods}
}

type Deployment struct {
	Name     string `json:"name"`
	Replicas int32  `json:"replicas"`
}
