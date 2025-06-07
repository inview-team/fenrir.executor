package kuber

import (
	"context"
	"fmt"

	"github.com/inviewteam/fenrir.executor/internal/domain/entity"
	"github.com/inviewteam/fenrir.executor/internal/domain/service"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Repository struct {
	client *kubernetes.Clientset
}

func New(config *rest.Config) (*Repository, error) {
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
		ePods = append(ePods, entity.NewPod(pod.Name, string(pod.Status.Phase)))
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
		if kerrors.IsNotFound(err) {
			return service.ErrPodNotFound
		} else {
			return fmt.Errorf("failed to delete pod: %v", err)
		}
	}
	return nil
}

func (r *Repository) GetPodByName(ctx context.Context, namespace, podName string) (*entity.Pod, error) {
	pod, err := r.client.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			return nil, service.ErrPodNotFound
		} else {
			return nil, fmt.Errorf("failed to get pod: %v", err)
		}
	}

	return entity.NewPod(pod.Name, string(pod.Status.Phase)), nil
}

func (r *Repository) GetDeploymentByName(ctx context.Context, namespace string, deploymentName string) (*entity.Deployment, error) {
	dpClient := r.client.AppsV1().Deployments(namespace)
	deployment, err := dpClient.Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment: %v", err)
	}
	return &entity.Deployment{Name: deployment.Name, Replicas: *deployment.Spec.Replicas}, nil
}
