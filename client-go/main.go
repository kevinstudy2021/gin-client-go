package main

import (
	"context"
	"flag"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/klog/v2"
	"path/filepath"
)

func main() {
	var kubeConfig *string
	ctx := context.Background()
	if home := homedir.HomeDir();home != ""{
		kubeConfig = flag.String("kubeConfig", filepath.Join(home, ".kube","config"),"absolvet path to the kube config")
		}else {
		kubeConfig = flag.String("kubeConfig", "","absolvet path to the kube config")
	}
	flag.Parse()
	config, err := clientcmd.BuildConfigFromFlags("", *kubeConfig)
	if err != nil {
		klog.Fatal(err)
	}
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Fatal(err)
	}
	namespaceList, err := clientSet.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return
	}
	namespaces := namespaceList.Items
	for _,namespace := range namespaces{
		fmt.Println(namespace.Name)
	}
	podLists, err := clientSet.CoreV1().Pods("default").List(ctx, metav1.ListOptions{})
	if err != nil {
		return
	}
	fmt.Println("====================pod list====================")
	for _, pods := range podLists.Items{
		fmt.Println(pods.Name)
	}
}