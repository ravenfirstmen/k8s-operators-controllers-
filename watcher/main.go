package main

import (
	"context"
	"flag"
	log "github.com/sirupsen/logrus"
	"k8s-watcher/pkg/processor"
	"k8s-watcher/pkg/watcher"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
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

	ctx := context.Background()
	p := processor.NewEventsProcessor(logger, clientSet)

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		p.Stop(ctx)
	}()

	p.StartWatchers(ctx, []watcher.Watcher{
		watcher.NewDeploymentWatcher(logger, clientSet),
		watcher.NewPodWatcher(logger, clientSet),
	})
	
	return 0
}

func main() {
	logger := initLogger()
	os.Exit(run(logger))
}
