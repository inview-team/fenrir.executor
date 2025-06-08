package routes

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/inviewteam/fenrir.executor/internal/application"
	"github.com/inviewteam/fenrir.executor/internal/domain/service"
	"github.com/inviewteam/fenrir.executor/internal/infrastructure/http/views"

	log "github.com/sirupsen/logrus"
)

// restartPod godoc
//
//	@Summary		Restart Pod
//	@Description	restart pod
//	@Tags			Pods
//	@Param			namespace	path	string	true	"Name of namespace"
//	@Param			pod_name	path	string	true	"Name of pod"
//	@Success		200
//	@Router			/kubernetes/{namespace}/pods/{pod_name} [delete]
func restartPod(srv *service.Executor) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		errMsg := "failed to restart pod"
		ctx := r.Context()
		namespace := mux.Vars(r)["namespace"]
		podName := mux.Vars(r)["pod_name"]

		err := srv.Restart(ctx, namespace, podName)
		if err != nil {
			if errors.Is(err, service.ErrPodNotFound) {
				log.Info("pod not found")
				http.Error(w, "pod not found", http.StatusNotFound)
			} else {
				log.Error(err.Error())
				http.Error(w, errMsg, http.StatusInternalServerError)
			}
			return
		}
		w.WriteHeader(http.StatusOK)
	})
}

// scaleDeployment godoc
//
//	@Summary		Scale Deployment
//	@Description	Scale Deployment
//	@Tags			Deployments
//	@Param			namespace		path	string	true	"Name of namespace"
//	@Param			deployment_name	path	string	true	"Name of Deployment"
//	@Param			replicas		query	string	true	"Amount of Replicas"
//	@Success		200
//	@Router			/kubernetes/{namespace}/pods/{pod_name} [put]
func scaleDeployment(srv *service.Executor) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		errMsg := "failed to scale deployment"
		ctx := r.Context()
		namespace := mux.Vars(r)["namespace"]
		deploymentName := mux.Vars(r)["deployment_name"]
		replicas := r.URL.Query().Get("replicas")

		targetReplicas, err := strconv.Atoi(replicas)
		if err != nil {
			log.Info("wrong payload")
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		err = srv.Scale(ctx, namespace, deploymentName, int32(targetReplicas))
		if err != nil {
			if errors.Is(err, service.ErrDeploymentNotFound) {
				log.Info("deployment not found")
				http.Error(w, "deployment not found", http.StatusNotFound)
			}
			log.Error(err.Error())
			http.Error(w, errMsg, http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
}

// getPodInformation godoc
//
//	@Summary		Get Pod Information
//	@Description	Get Pod Information
//	@Tags			Pods
//	@Param			namespace	path	string	true	"Name of namespace"
//	@Param			pod_name	path	string	true	"Name of pod"
//	@Success		200			object	views.Pod
//	@Router			/kubernetes/{namespace}/deployments/{pod_name} [get]
func getPodInformation(srv *service.Executor) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		errMsg := "failed to restart pod"
		ctx := r.Context()
		namespace := mux.Vars(r)["namespace"]
		podName := mux.Vars(r)["pod_name"]

		pod, err := srv.GetPodByName(ctx, namespace, podName)
		if err != nil {
			if errors.Is(err, service.ErrPodNotFound) {
				log.Info("pod not found")
				http.Error(w, err.Error(), http.StatusNotFound)
			} else {
				log.Error(err.Error())
				http.Error(w, errMsg, http.StatusInternalServerError)
			}
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(views.NewPod(pod))
	})
}

// getPodInformation godoc
//
//	@Summary		Get Pod Information
//	@Description	Get Pod Information
//	@Tags			Pods
//	@Param			namespace	path	string	true	"Name of namespace"
//	@Param			deployment		query	string	true	"Amount of Replicas"
//	@Success		200			object	views.Pod
//	@Router			/kubernetes/{namespace}/pods [get]
func listPodByDeployment(srv *service.Executor) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		errMsg := "failed to restart pod"
		ctx := r.Context()
		namespace := mux.Vars(r)["namespace"]
		deployment := r.URL.Query().Get("deployment")

		pods, err := srv.ListPodByDeployment(ctx, namespace, deployment)
		if err != nil {
			if errors.Is(err, service.ErrDeploymentNotFound) {
				log.Info("deployment not found")
				http.Error(w, service.ErrDeploymentNotFound.Error(), http.StatusNotFound)
			} else {
				log.Error(err.Error())
				http.Error(w, errMsg, http.StatusInternalServerError)
			}
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(views.NewPods(pods))
	})
}

func makeKubernetesRoutes(r *mux.Router, app *application.Application) {
	path := "/kubernetes"
	serviceRouter := r.PathPrefix(path).Subrouter()
	serviceRouter.Handle("/{namespace}/pods/{pod_name}", getPodInformation(app.ExecutorService)).Methods("GET")
	serviceRouter.Handle("/{namespace}/pods/{pod_name}", restartPod(app.ExecutorService)).Methods("DELETE")
	serviceRouter.Handle("/{namespace}/pods", listPodByDeployment(app.ExecutorService)).Methods("GET")
	serviceRouter.Handle("/{namespace}/deployments/{deployment_name}", scaleDeployment(app.ExecutorService)).Methods("PUT")
}
