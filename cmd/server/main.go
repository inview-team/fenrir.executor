package main

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"

	"github.com/inviewteam/fenrir.executor/cmd/internal/infrastructure/kuber"
	"k8s.io/client-go/util/homedir"
)

func main() {
	ctx := context.TODO()
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	repo, err := kuber.New(kubeconfig)
	if err != nil {
		panic(err)
	}

	pods, _ := repo.List(ctx, "guestbook")
	for _, pod := range pods {
		fmt.Println(pod)
	}

	// err = repo.Scale(ctx, "guestbook", "frontend", 6)
	// if err != nil {
	// 	fmt.Print(err)
	// }

	err = repo.Delete(ctx, "guestbook", "frontend-795b566649-d482j")
	if err != nil {
		fmt.Println(err)
	}

	pods, _ = repo.List(ctx, "guestbook")
	for _, pod := range pods {
		fmt.Println(pod)
	}
}
