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
