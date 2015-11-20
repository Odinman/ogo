// Ogo

package ogo

import (
	"errors"
)

var (
	notOwner = errors.New("Not_Owner")
)

type Lock struct {
	Key      string
	checksum string
	locked   bool //是否已锁
}

/* {{{ func NewLock(key string) (lk *Lock)
 *
 */
func NewLock(key string) (lk *Lock) {
	lk = &Lock{Key: key}
	return
}

/* }}} */

/* {{{ func (lk *Lock) Get() error
 * 获取锁
 */
func (lk *Lock) Get() (err error) {
	if !lk.locked { //如果已锁说明本请求是锁的owner
		if lk.checksum, err = GetLock(lk.Key); err == nil {
			lk.locked = true
		}
	}
	return
}

/* }}} */

/* {{{ func (lk *Lock) Release() error
 * 释放锁
 */
func (lk *Lock) Release() (err error) {
	if !lk.locked {
		err = notOwner
	} else {
		if err = ReleaseLock(lk.Key); err == nil {
			lk.locked = false
		}
	}
	return
}

/* }}} */
