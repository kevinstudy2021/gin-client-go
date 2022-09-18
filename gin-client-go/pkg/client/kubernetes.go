package client

import (
	"errors"
	"flag"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/klog/v2"
	"path/filepath"
	"sync"
)

var onceClient sync.Once
var onceConfig sync.Once
var KubeConfig *rest.Config
var KubeClientSet *kubernetes.Clientset

func GetK8SClientSet() (*kubernetes.Clientset,error) {
	onceClient.Do(func() {
		config, err := GetRestConfig()
		if err != nil {
			return
		}
		KubeClientSet, err = kubernetes.NewForConfig(config)
		if err != nil {
			klog.Fatal(err)
			return
		}
	})

	return KubeClientSet, nil
}

func GetRestConfig()(*rest.Config, error)  {
	onceConfig.Do(func() {
		var kubeConfig *string
		var err error
		if home := homedir.HomeDir();home != ""{
			kubeConfig = flag.String("kubeConfig", filepath.Join(home, ".kube","config"),"absolvet path to the kube config")
		}else {
			klog.Fatal("read config error , config is empty")
			err = errors.New("read config error,config is empty")
			return
		}
		flag.Parse()
		KubeConfig, err = clientcmd.BuildConfigFromFlags("", *kubeConfig)
		if err != nil {
			klog.Fatal(err)
			return
		}
		return
	})

	return KubeConfig, nil
}