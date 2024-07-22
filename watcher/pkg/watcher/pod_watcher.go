package watcher

import (
	"context"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

type PodWatcher struct {
	clientSet *kubernetes.Clientset
	logger    *logrus.Logger

	//
	done chan bool
}

func NewPodWatcher(logger *logrus.Logger, clientSet *kubernetes.Clientset) *PodWatcher {
	return &PodWatcher{
		logger:    logger,
		clientSet: clientSet,
		done:      make(chan bool, 1),
	}
}

func (p *PodWatcher) Name() string {
	return "PodWatcher"
}

func (p *PodWatcher) Watch(ctx context.Context) (watch.Interface, error) {
	logger := p.logger.WithContext(ctx)

	watcher, err := p.clientSet.CoreV1().Pods("").Watch(ctx, metav1.ListOptions{})
	if err != nil {
		logger.WithError(err).Error("failed to subscribe to pods")
		return nil, err
	}

	return watcher, nil
}

func (p *PodWatcher) Process(ctx context.Context, event watch.Event) error {
	pod := event.Object.(*v1.Pod)
	logger := p.logger.
		WithContext(ctx).
		WithFields(logrus.Fields{
			"watcher":    p.Name(),
			"object":     pod.Name,
			"event type": event.Type,
		})

	switch event.Type {
	case watch.Added:
		logger.Info("pod added")
		break
	case watch.Modified:
		logger.Info("pod modified")
		break
	case watch.Deleted:
		logger.Warn("pod watcher received event for a deleted pod")
		break
	case watch.Error:
		logger.Warn("pod watcher received an error")
		break
	}

	return nil
}

func (p *PodWatcher) Done(ctx context.Context) <-chan bool {
	return p.done
}

func (p *PodWatcher) Stop(ctx context.Context) error {
	p.done <- true
	return nil
}
