package kuber

import (
	"context"
	"fmt"

	"github.com/inviewteam/fenrir.executor/cmd/internal/domain/entity"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Repository struct {
	client *kubernetes.Clientset
}

func New(path *string) (*Repository, error) {
	config, err := clientcmd.BuildConfigFromFlags("", *path)
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &Repository{client: clientset}, nil
}

func (r *Repository) List(ctx context.Context, namespace string) ([]*entity.Pod, error) {
	pods, err := r.client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	var ePods []*entity.Pod
	for _, pod := range pods.Items {
		ePods = append(ePods, entity.NewPod(string(pod.UID), pod.Name, pod.Status.Message))
	}
	return ePods, nil
}

func (r *Repository) Scale(ctx context.Context, namespace, deploymentName string, replicas int32) error {
	dpClient := r.client.AppsV1().Deployments(namespace)
	deployment, err := dpClient.Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to scale deployment: %v", err)
	}
	deployment.Spec.Replicas = &replicas
	_, err = dpClient.Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to scale deployment: %v", err)
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, namespace, podName string) error {
	err := r.client.CoreV1().Pods(namespace).Delete(ctx, podName, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete pod: %v", err)
	}
	return nil
}
