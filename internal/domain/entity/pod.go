package entity

type Pod struct {
	ID     string
	Name   string
	Status string
}

func NewPod(id, name, status string) *Pod {
	return &Pod{
		ID:     id,
		Name:   name,
		Status: status,
	}
}

type PodRepository interface {
	List(namespace string) ([]*Pod, error)
	Get(namespace, id string) (*Pod, error)
	Update(namespace string, pod *Pod) error
}
