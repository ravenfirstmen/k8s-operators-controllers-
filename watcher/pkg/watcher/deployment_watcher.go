package watcher

import (
	"context"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

type DeploymentWatcher struct {
	clientSet *kubernetes.Clientset
	logger    *logrus.Logger

	//
	done chan bool
}

func NewDeploymentWatcher(logger *logrus.Logger, clientSet *kubernetes.Clientset) *DeploymentWatcher {
	return &DeploymentWatcher{
		clientSet: clientSet,
		logger:    logger,
		done:      make(chan bool, 1),
	}
}

func (d *DeploymentWatcher) Name() string {
	return "DeploymentWatcher"
}

func (d *DeploymentWatcher) Watch(ctx context.Context) (watch.Interface, error) {
	logger := d.logger.WithContext(ctx)

	watcher, err := d.clientSet.AppsV1().Deployments("").Watch(ctx, metav1.ListOptions{})
	if err != nil {
		logger.WithError(err).Error("Failed to create deployment watcher")
		return nil, err
	}

	return watcher, nil
}

func (d *DeploymentWatcher) Process(ctx context.Context, event watch.Event) error {
	dep := event.Object.(*v1.Deployment)

	logger := d.logger.
		WithContext(ctx).
		WithFields(logrus.Fields{
			"watcher":    d.Name(),
			"object":     dep.Name,
			"event type": event.Type,
		})

	switch event.Type {
	case watch.Added:
		logger.Info("deployment added")
		break
	case watch.Modified:
		logger.Info("deployment modified")
		break
	case watch.Deleted:
		logger.Warn("deployment watcher received event for a deleted pod")
		break
	case watch.Error:
		logger.Warn("deployment watcher received an error")
		break
	}

	return nil
}

func (d *DeploymentWatcher) Done(ctx context.Context) <-chan bool {
	return d.done
}

func (d *DeploymentWatcher) Stop(ctx context.Context) error {
	d.done <- true
	return nil
}
