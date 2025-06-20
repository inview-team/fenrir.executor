package kuber

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/inviewteam/fenrir.executor/internal/domain/entity"
	"github.com/inviewteam/fenrir.executor/internal/domain/service"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	log "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
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

func (r *Repository) GetPodContainers(ctx context.Context, namespace, podName string) ([]*entity.Container, error) {
	pod, err := r.client.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			return nil, service.ErrPodNotFound
		} else {
			return nil, fmt.Errorf("failed to get pod: %w", err)
		}
	}

	podMetrics, err := r.mClient.MetricsV1beta1().PodMetricses(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod metrics: %w", err)
	}

	metricsMap := make(map[string]v1.ResourceList)
	for _, c := range podMetrics.Containers {
		metricsMap[c.Name] = c.Usage
	}

	containers := make(map[string]*entity.Container, len(pod.Spec.Containers))
	for _, container := range pod.Spec.Containers {
		usage, ok := metricsMap[container.Name]
		var cpuUsage, memUsage int64
		if ok {
			cpuUsage = usage.Cpu().MilliValue() * 1000000 // милликоны -> наносекунды
			memUsage = usage.Memory().Value()
		}

		cpuLimit := int64(0)
		memLimit := int64(0)
		if container.Resources.Limits != nil {
			if cpuQ, ok := container.Resources.Limits[v1.ResourceCPU]; ok {
				cpuLimit = cpuQ.MilliValue() * 1000000
			}
			if memQ, ok := container.Resources.Limits[v1.ResourceMemory]; ok {
				memLimit = memQ.Value()
			}
		}

		containers[container.Name] = &entity.Container{
			Name:         container.Name,
			State:        "",
			CpuUsage:     cpuUsage,
			MemoryUsage:  memUsage,
			CpuLimits:    cpuLimit,
			MemoryLimits: memLimit,
		}
	}

	eContainers := make([]*entity.Container, 0, len(pod.Spec.Containers))
	for _, container := range pod.Status.ContainerStatuses {
		eContainer, ok := containers[container.Name]
		if !ok {
			continue
		}

		if container.State.Running != nil {
			eContainer.State = "Running"
		} else if container.State.Waiting != nil {
			eContainer.State = "Waiting"
		} else if container.State.Terminated != nil {
			eContainer.State = "Terminated"
		}

		eContainers = append(eContainers, eContainer)

	}

	return eContainers, nil
}

func (r *Repository) GetDeploymentByName(ctx context.Context, namespace string, deploymentName string) (*entity.Deployment, error) {
	dpClient := r.client.AppsV1().Deployments(namespace)
	deployment, err := dpClient.Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment: %w", err)
	}
	return &entity.Deployment{Name: deployment.Name, Replicas: *deployment.Spec.Replicas}, nil
}

func (r *Repository) GetPodLogs(ctx context.Context, namespace, podName, containerName string, tailLines int64) (string, error) {
	_, err := r.client.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			return "", service.ErrPodNotFound
		} else {
			return "", fmt.Errorf("failed to get logs: %w", err)
		}
	}

	podLogOpts := v1.PodLogOptions{
		Container: containerName,
		TailLines: &tailLines,
	}
	req := r.client.CoreV1().Pods(namespace).GetLogs(podName, &podLogOpts)
	podLogs, err := req.Stream(ctx)
	if err != nil {
		return "", fmt.Errorf("error in opening stream: %w", err)
	}
	defer podLogs.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		return "", fmt.Errorf("error in copy information from podLogs to buf: %w", err)
	}
	str := buf.String()

	return str, nil
}

func (r *Repository) DescribePod(ctx context.Context, namespace, podName string) (string, error) {
	pod, err := r.client.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			return "", service.ErrPodNotFound
		}
		return "", fmt.Errorf("failed to get pod: %w", err)
	}

	pod.ManagedFields = nil
	y, err := yaml.Marshal(pod)
	if err != nil {
		return "", fmt.Errorf("failed to marshal pod to yaml: %w", err)
	}

	return string(y), nil
}

func (r *Repository) DescribeDeployment(ctx context.Context, namespace, deploymentName string) (string, error) {
	deployment, err := r.client.AppsV1().Deployments(namespace).Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			return "", service.ErrDeploymentNotFound
		}
		return "", fmt.Errorf("failed to get deployment: %w", err)
	}

	deployment.ManagedFields = nil
	y, err := yaml.Marshal(deployment)
	if err != nil {
		return "", fmt.Errorf("failed to marshal deployment to yaml: %w", err)
	}

	return string(y), nil
}

func (r *Repository) Rollback(ctx context.Context, namespace, deploymentName string) error {
	dpClient := r.client.AppsV1().Deployments(namespace)
	deployment, err := dpClient.Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			return service.ErrDeploymentNotFound
		}
		return fmt.Errorf("failed to get deployment: %w", err)
	}

	revisionList, err := r.client.AppsV1().ReplicaSets(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: "app=" + deployment.Spec.Selector.MatchLabels["app"], // Adjust label selector as needed
	})
	if err != nil {
		return fmt.Errorf("failed to list replica sets: %v", err)
	}
	log.Infof("found revisions %v", revisionList)

	// Find the second last revision (previous revision)
	var previousRevision *appsv1.ReplicaSet
	if len(revisionList.Items) > 1 {
		// Sort ReplicaSets by creation timestamp
		sortReplicaSetsByCreationTimestamp(revisionList.Items)
		previousRevision = &revisionList.Items[len(revisionList.Items)-2] // Get the second last
	} else {
		return service.ErrNoPreviousRevisionsFound
	}

	// Get the desired ReplicaSet's template
	if previousRevision == nil {
		return service.ErrNoPreviousRevisionsFound
	}
	oldTemplate := previousRevision.Spec.Template

	// Update the deployment with the old template
	deployment.Spec.Template = oldTemplate

	// Apply the updated deployment
	_, err = r.client.AppsV1().Deployments(namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update deployment: %v", err)
	}

	return nil
}

func sortReplicaSetsByCreationTimestamp(replicaSets []appsv1.ReplicaSet) {
	sort.Slice(replicaSets, func(i, j int) bool {
		return replicaSets[i].CreationTimestamp.Before(&replicaSets[j].CreationTimestamp)
	})
}
