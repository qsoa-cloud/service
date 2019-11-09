package mysql

import (
	"context"
	"log"
	"math/rand"

	"gopkg.qsoa.cloud/service/discovery"
)

type watcher struct {
	instances []discovery.Instance
	init      bool
	initCh    chan struct{}
}

func newWatcher(name string) *watcher {
	w := &watcher{
		initCh: make(chan struct{}),
	}

	go w.watch(name)

	return w
}

func (w *watcher) watch(name string) {
	for {
		if err := discovery.Watch(context.Background(), discovery.TypeMySql, name, func(instances []discovery.Instance) {
			w.instances = instances
			if !w.init {
				w.init = true
				close(w.initCh)
			}
		}); err != nil {
			log.Printf("Cannot watch mysql discovery: %v", err)
		}
	}
}

func (w *watcher) getInstances() []discovery.Instance {
	<-w.initCh
	return w.instances
}

func (w *watcher) getRandomInstance() *discovery.Instance {
	instances := w.getInstances()

	if len(instances) == 0 {
		return nil
	}

	return &instances[rand.Intn(len(instances))]
}
