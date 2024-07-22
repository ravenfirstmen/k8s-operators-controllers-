package watcher

import (
	"context"
	"k8s.io/apimachinery/pkg/watch"
)

type Watcher interface {
	Name() string
	Watch(ctx context.Context) (watch.Interface, error)
	Process(ctx context.Context, event watch.Event) error
	Done(ctx context.Context) <-chan bool
	Stop(ctx context.Context) error
}
