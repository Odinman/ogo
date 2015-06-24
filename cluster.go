// Ogo

package ogo

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Odinman/ogo/utils"
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

/* {{{ func GetLock(key string) string
 * 获取锁
 */
func GetLock(key string) string {
	if cc := ClusterClient(); cc != nil {
		cs := utils.NewUUID()
		ts := time.Now().Unix()
		val := fmt.Sprint(ts, ",", cs)
		if err := cc.SetNX(key, val, 600*time.Second).Err(); err == nil { //10分钟自动消失
			// lock exist
			if old, err := cc.Get(key).Result(); err == nil {
				vs := strings.SplitN(old, ",", 2)
				if ots, err := strconv.Atoi(vs[0]); err == nil && int(ts)-ots > 60 { //1分钟过期
					//过期了,抢
					if getSet := cc.GetSet(key, val); getSet.Err() == nil {
						gl := getSet.Val()
						if gl == old {
							//抢到了,返回checksum
							return cs
						} else {
							//没抢到,并且还覆盖了人家抢到的锁,(可能会产生问题)
							//cc.Set(key, gl, 600*time.Second)
						}
					}
				}
			}
		}
	}
	return ""
}

/* }}} */

/* {{{ func ReleaseLock(key,cs string) error
 * 释放锁
 */
func ReleaseLock(key, cs string) error {
	if cc := ClusterClient(); cc != nil {
		if cur, err := cc.Get(key).Result(); err == nil {
			vs := strings.SplitN(cur, ",", 2)
			if vs[1] != cs {
				Warn("[self: %s][lock: %s][diff]", cs, vs[1])
			}
			return cc.Del(key).Err()
		}
	}
	return fmt.Errorf("release wrong")
}

/* }}} */
