package processor

import (
	"context"
	"github.com/sirupsen/logrus"
	"k8s-watcher/pkg/watcher"
	"k8s.io/client-go/kubernetes"
	"sync"
)

type EventsProcessor struct {
	clientSet *kubernetes.Clientset
	logger    *logrus.Logger
	watchers  []watcher.Watcher
}

func NewEventsProcessor(logger *logrus.Logger, clientSet *kubernetes.Clientset) *EventsProcessor {
	return &EventsProcessor{clientSet: clientSet, logger: logger}
}

func (p *EventsProcessor) StartWatchers(ctx context.Context, watchers []watcher.Watcher) {
	logger := p.logger

	listener := func(wg *sync.WaitGroup, ctx context.Context, watcher watcher.Watcher) {
		defer wg.Done()

		logger := logger.WithContext(ctx)
		watch, err := watcher.Watch(ctx)
		if err != nil {
			logger.WithError(err).Error("failed to watch watch")
			return
		}
		defer watch.Stop()

		for {
			select {
			case event := <-watch.ResultChan():
				err := watcher.Process(ctx, event)
				if err != nil {
					logger.WithError(err).Error("failed to process watch")
					return
				}
				break
			case done := <-watcher.Done(ctx):
				if done {
					logger.Infof("stopping watcher %s", watcher.Name())
					return
				}
				break
			}
		}
	}

	var wg sync.WaitGroup
	p.watchers = []watcher.Watcher{}
	for _, w := range watchers {
		p.watchers = append(p.watchers, w)
		wg.Add(1)
		go listener(&wg, ctx, w)
	}
	wg.Wait()
}

func (p *EventsProcessor) Stop(ctx context.Context) {
	logger := p.logger.WithContext(ctx)

	for _, w := range p.watchers {
		err := w.Stop(ctx)
		if err != nil {
			logger.WithError(err).Error("failed to stop watcher")
			return
		}
	}
}
