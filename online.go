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
	Key  string
	Ex   time.Time
	Item interface{}
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

/* {{{ func SetOnlineKey(id string, ol *Online,p interface{}) bool
 *
 */
func SetOnlineKey(id, key string, p interface{}) bool {
	ol := &Online{
		Key:  key,
		Ex:   time.Now().Add(300 * time.Second), //缓存5分钟
		Item: p,
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
	//defer m.lock.RUnlock()
	if val, ok := m.olm[k]; ok {
		if time.Now().Before(val.Ex) { //还没过期
			m.lock.RUnlock()
			return val
		} else { //过期,删除
			m.lock.RUnlock()
			m.Delete(k)
			return nil
		}
	}
	m.lock.RUnlock()
	return nil
}

// Delete the given key and value.
func (m *OnlineMap) Delete(k string) {
	m.lock.Lock()
	defer m.lock.Unlock()
	delete(m.olm, k)
}
