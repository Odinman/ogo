// Ogo

package ogo

import (
	"fmt"
	"sync"
	"time"
)

type OnlineMap struct {
	lock *sync.RWMutex
	olm  map[string]*Online
}

type Online struct {
	Key string
	Ex  time.Time
}

var (
	OnlineCache *OnlineMap
)

func init() {
	OnlineCache = &OnlineMap{
		lock: new(sync.RWMutex),
		olm:  make(map[string]*Online),
	}
}

/* {{{ func GetOnlineKey(id string) (*Online,error)
 * 获取在线的key
 */
func GetOnlineKey(id string) (ol *Online, err error) {
	if ol = OnlineCache.Get(id); ol == nil {
		err = fmt.Errorf("not found")
	}
	return
}

/* }}} */

/* {{{ func SetOnlineKey(id string, ol *Online) bool
 *
 */
func SetOnlineKey(id, key string) bool {
	ol := &Online{
		Key: key,
		Ex:  time.Now().Add(300 * time.Second), //缓存5分钟
	}
	return OnlineCache.Set(id, ol)
}

/* }}} */

// Maps the given key and value. Returns false
// if the key is already in the map and changes nothing.
func (m *OnlineMap) Set(k string, v *Online) bool {
	m.lock.Lock()
	defer m.lock.Unlock()
	if val, ok := m.olm[k]; !ok {
		m.olm[k] = v
	} else if val != v {
		m.olm[k] = v
	} else {
		return false
	}
	return true
}

// Get from maps return the k's value
func (m *OnlineMap) Get(k string) (ol *Online) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	if val, ok := m.olm[k]; ok {
		if time.Now().Before(val.Ex) { //还没过期
			return val
		} else { //过期,删除
			m.Delete(k)
		}
	}
	return nil
}

// Delete the given key and value.
func (m *OnlineMap) Delete(k string) {
	m.lock.Lock()
	defer m.lock.Unlock()
	delete(m.olm, k)
}
