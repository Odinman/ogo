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

/* {{{ func CacheIncrByFloat(key string, value float64) error
 *
 */
func CacheIncrByFloat(key string, value float64) error {
	if cc := ClusterClient(); cc != nil {
		return cc.IncrByFloat(key, value).Err()
	} else {
		return fmt.Errorf("incr %s failed", key)
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

/* {{{ func GetLock(key string) (string, error)
 * 获取锁
 */
func GetLock(key string) (string, error) {
	if cc := ClusterClient(); cc != nil {
		cs := utils.NewUUID()
		ts := time.Now().Unix()
		val := fmt.Sprint(ts, ",", cs)
		var tried int
		for tried <= 3 {
			tried++
			if err := cc.SetNX(key, val, 600*time.Second).Err(); err != nil { //10分钟自动消失
				// lock exist
				if old, err := cc.Get(key).Result(); err == nil {
					vs := strings.SplitN(old, ",", 2)
					if ots, _ := strconv.Atoi(vs[0]); int(ts)-ots > 60 { //1分钟过期
						//过期了,抢
						if getSet := cc.GetSet(key, val); getSet.Err() == nil {
							gl := getSet.Val()
							if gl == old {
								//抢到了,返回checksum
								return cs, nil
							} else {
								//没抢到,并且还覆盖了人家抢到的锁,(可能会产生问题)
								//cc.Set(key, gl, 600*time.Second)
							}
						} else { //奇怪的情况
							return "", err
						}
					}
				} else { //奇怪的情况
					return "", err
				}
			} else {
				return cs, nil
			}
		}
		time.Sleep(50 * time.Millisecond) //50ms
	}
	return "", fmt.Errorf("can't get lock")
}

/* }}} */

/* {{{ func ReleaseLock(key string) error
 * 释放锁
 */
func ReleaseLock(key string) error {
	if cc := ClusterClient(); cc != nil {
		if cur, err := cc.Get(key).Result(); err == nil {
			vs := strings.SplitN(cur, ",", 2)
			Debug("[lock: %s]", vs[1])
			return cc.Del(key).Err()
		}
	}
	return fmt.Errorf("release wrong")
}

/* }}} */
