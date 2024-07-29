package main

import (
	"flag"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"os"
	"path/filepath"
	"time"
)

func initLogger() *log.Logger {
	logger := log.New()
	logger.SetFormatter(&log.TextFormatter{})
	logger.SetOutput(os.Stdout)
	logger.SetLevel(log.DebugLevel)

	return logger
}

func run(logger *log.Logger) int {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		logger.Error(err)
		return 1
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		logger.Error(err)
		return 1
	}

	informerFactory := informers.NewSharedInformerFactory(clientSet, 15*time.Second)
	podInformer := informerFactory.Core().V1().Pods().Informer()

	_, err = podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod := obj.(*v1.Pod)
			logger.Infof("ADD: Pod: %s", pod.Name)
		},
		UpdateFunc: func(old, new interface{}) {
			oldPod := old.(*v1.Pod)
			newPod := new.(*v1.Pod)
			logger.Infof("UPDATE: Pod %s - Old pod: %s", newPod.Name, oldPod.Name)
		},
		DeleteFunc: func(obj interface{}) {
			pod := obj.(*v1.Pod)
			logger.Infof("DELETE: Pod: %s", pod.Name)
		},
	})
	if err != nil {
		return 0
	}
	podInformer.Run(make(<-chan struct{}))

	return 0
}

func main() {
	logger := initLogger()
	os.Exit(run(logger))
}
