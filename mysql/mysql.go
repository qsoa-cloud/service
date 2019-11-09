package mysql

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"sync"

	"github.com/go-sql-driver/mysql"
)

type qDriver struct {
	watchers    map[string]*watcher
	watchersMtx sync.RWMutex
}

func init() {
	drv := &qDriver{
		watchers: map[string]*watcher{},
	}

	sql.Register("qmysql", drv)
}

func (d *qDriver) Open(name string) (driver.Conn, error) {
	instance := d.getWatcher(name).getRandomInstance()
	if instance == nil {
		return nil, fmt.Errorf("no instances were found")
	}

	return mysql.MySQLDriver{}.Open(instance.Value)
}

func (d *qDriver) getWatcher(name string) *watcher {
	d.watchersMtx.RLock()
	w := d.watchers[name]
	d.watchersMtx.RUnlock()

	if w == nil {
		d.watchersMtx.Lock()
		if _, exists := d.watchers[name]; !exists {
			w = newWatcher(name)
			d.watchers[name] = w
		}
		d.watchersMtx.Unlock()
	}

	return w
}
