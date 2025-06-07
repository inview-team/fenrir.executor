package routes

import (
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/inviewteam/fenrir.executor/docs"
	"github.com/inviewteam/fenrir.executor/internal/application"
	"github.com/inviewteam/fenrir.executor/internal/infrastructure/http/handlers"
	httpSwagger "github.com/swaggo/http-swagger"
)

func Make(app *application.Application) http.Handler {
	r := mux.NewRouter()
	r.PathPrefix("/docs/").Handler(httpSwagger.WrapHandler)

	r.MethodNotAllowedHandler = handlers.NotAllowedHandler()
	r.NotFoundHandler = handlers.NotFoundHandler()

	path := "/api"
	apiRouter := r.PathPrefix(path).Subrouter()
	makeKubernetesRoutes(apiRouter, app)
	return r
}
