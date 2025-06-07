package main

import (
	"context"
	"flag"
	"path/filepath"

	"github.com/inviewteam/fenrir.executor/internal/application"
	server "github.com/inviewteam/fenrir.executor/internal/infrastructure/http"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	ctx := context.Background()
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}

	app, err := application.New(ctx, config)
	if err != nil {
		panic(err)
	}

	srv := server.NewServer(app)
	srv.Start(ctx)
}
