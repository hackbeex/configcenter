package meta

import (
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/hackbeex/configcenter/discover/watcher"
	"github.com/hackbeex/configcenter/util/log"
	"time"
)

func (t *Table) watch(w watcher.Watcher) {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("[Recovery] %s panic recovered:\n%s", time.Now().String(), err)
		}
		w.Refresh()
		go t.watch(w)
	}()

	for {
		select {
		case <-w.Ctx().Done():
			log.Error(w.Ctx().Err())
			w.Refresh()
			goto Over
		case resp := <-w.GetWatchChan():
			if resp.Canceled {
				log.Error("watch canceled", w.Ctx().Err())
				w.Refresh()
				goto Over
			}
			for _, evt := range resp.Events {
				switch evt.Type {
				case mvccpb.PUT:
					if err := w.Put(evt.Kv, evt.IsCreate()); err != nil {
						log.Error(err)
					}
				case mvccpb.DELETE:
					if err := w.Delete(evt.Kv); err != nil {
						log.Error(err)
					}
				default:
					log.Warnf("unrecognized event type: %d", evt.Type)
				}
			}
		}
	}

Over:
	log.Info("watch task finished")
}
