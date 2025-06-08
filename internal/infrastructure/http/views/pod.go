package views

import (
	"github.com/inviewteam/fenrir.executor/internal/domain/entity"
)

type Pod struct {
	Name      string                `json:"name"`
	Status    string                `json:"status"`
	Restarts  int                   `json:"restarts"`
	Age       string                `json:"age"`
	Resources []*ContainerResources `json:"resources"`
}

type ContainerResources struct {
	Name         string `json:"name"`
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
		Resources: func() []*ContainerResources {
			res := make([]*ContainerResources, 0, len(e.Resources))
			for _, cr := range e.Resources {
				res = append(res, &ContainerResources{
					Name:         cr.Name,
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
