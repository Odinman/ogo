// Ogo

package ogo

import (
	"fmt"
	"time"

	//"gopkg.in/redis.v3"
)

/* {{{ func CacheSet(key, value string, expire int) error
 *
 */
func CacheSet(key, value string, expire int) error {
	if cc := ClusterClient(); cc != nil {
		return cc.Set(key, value, time.Duration(expire)*time.Second).Err()
	} else {
		return fmt.Errorf("not found cache client")
	}
}

/* }}} */

/* {{{ func CacheGet(key) error
 *
 */
func CacheGet(key string) (string, error) {
	if cc := ClusterClient(); cc != nil {
		return cc.Get(key).Result()
	} else {
		return "", fmt.Errorf("not found cache client")
	}
}

/* }}} */

/* {{{ func GetLock(tag string) string
 * 获取锁
 */
func GetLock(key string) string {
	if cc := ClusterClient(); cc != nil {
		multi := client.Multi()
		defer multi.Close()
		val, err := multi.Watch(key).Result()
	}
	return ""
}

/* }}} */
