package kuber

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/inviewteam/fenrir.executor/internal/domain/entity"
	"github.com/inviewteam/fenrir.executor/internal/domain/service"
	v1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
)

type Repository struct {
	client  *kubernetes.Clientset
	mClient *metrics.Clientset
}

func New(config *rest.Config) (*Repository, error) {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	metricsClient, err := metrics.NewForConfig(config)
	if err != nil {
		log.Fatalf("Failed to create metrics client: %w", err)
	}
	return &Repository{client: clientset, mClient: metricsClient}, nil
}

func (r *Repository) ListPodsByDeployment(ctx context.Context, namespace string, deploymentName string) ([]*entity.Pod, error) {
	dpClient := r.client.AppsV1().Deployments(namespace)
	deployment, err := dpClient.Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods of deployment: %w", err)
	}

	selector := deployment.Spec.Selector

	// Convert selector to a string
	labelSelector := metav1.FormatLabelSelector(selector)

	pods, err := r.client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, err
	}
	var ePods []*entity.Pod
	for _, pod := range pods.Items {
		ePods = append(ePods, entity.NewPod(pod.Name, string(pod.Status.Phase), 0, 0, nil))
	}
	return ePods, nil
}

func (r *Repository) Scale(ctx context.Context, namespace, deploymentName string, replicas int32) error {
	dpClient := r.client.AppsV1().Deployments(namespace)
	deployment, err := dpClient.Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to scale deployment: %w", err)
	}
	deployment.Spec.Replicas = &replicas
	_, err = dpClient.Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to scale deployment: %w", err)
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, namespace, podName string) error {
	err := r.client.CoreV1().Pods(namespace).Delete(ctx, podName, metav1.DeleteOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			return service.ErrPodNotFound
		} else {
			return fmt.Errorf("failed to delete pod: %w", err)
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
			return nil, fmt.Errorf("failed to get pod: %w", err)
		}
	}

	totalRestarts := int32(0)
	for _, containerStatus := range pod.Status.ContainerStatuses {
		totalRestarts += containerStatus.RestartCount
	}

	return entity.NewPod(
		pod.Name,
		string(pod.Status.Phase),
		int(totalRestarts),
		time.Since(pod.CreationTimestamp.Time),
		nil), nil
}

func (r *Repository) GetPodMetrics(ctx context.Context, namespace, podName string) ([]*entity.ContainerResources, error) {
	pod, err := r.client.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			return nil, service.ErrPodNotFound
		} else {
			return nil, fmt.Errorf("failed to get pod: %w", err)
		}
	}

	podMetrics, err := r.mClient.MetricsV1beta1().PodMetricses(namespace).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod metrics: %w", err)
	}

	metricsMap := make(map[string]v1.ResourceList)
	for _, c := range podMetrics.Containers {
		metricsMap[c.Name] = c.Usage
	}

	var cResources []*entity.ContainerResources
	fmt.Printf("Метрики и лимиты пода %s/%s:\n", namespace, podName)
	for _, container := range pod.Spec.Containers {
		usage, ok := metricsMap[container.Name]
		if !ok {
			fmt.Printf("  Контейнер %s: метрики не найдены\n", container.Name)
			continue
		}

		cpuLimit := resource.NewQuantity(0, resource.DecimalSI)
		memLimit := resource.NewQuantity(0, resource.BinarySI)

		if container.Resources.Limits != nil {
			if val, ok := container.Resources.Limits[v1.ResourceCPU]; ok {
				cpuLimit = &val
			}
			if val, ok := container.Resources.Limits[v1.ResourceMemory]; ok {
				memLimit = &val
			}
		}

		cResources = append(cResources, &entity.ContainerResources{
			Name:         container.Name,
			CpuUsage:     usage.Cpu().Value(),
			MemoryUsage:  usage.Memory().Value(),
			CpuLimits:    cpuLimit.Value(),
			MemoryLimits: memLimit.Value(),
		})
	}
	return cResources, nil
}

func (r *Repository) GetDeploymentByName(ctx context.Context, namespace string, deploymentName string) (*entity.Deployment, error) {
	dpClient := r.client.AppsV1().Deployments(namespace)
	deployment, err := dpClient.Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment: %w", err)
	}
	return &entity.Deployment{Name: deployment.Name, Replicas: *deployment.Spec.Replicas}, nil
}
