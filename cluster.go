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

const (
	TD_TOTAL_FIELD = "_total_"
)

type TemporaryCounter struct {
	Name          string // 计数器名称
	Total         string // 汇总字段
	TotalAmount   float64
	Journal       string  //流水号
	JournalAmount float64 //金额
}

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

/* {{{ func CacheGet(key) (string,error)
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

/* {{{ func (tc *TemporaryCounter) Incr(amount float64) error
 * 临时扣除(就是在TemporaryCounter增加一笔)
 */
func (tc *TemporaryCounter) Incr(amount float64) error {
	if cc := ClusterClient(); cc != nil {
		if tc.Name == "" || tc.Journal == "" || amount == 0 { //必须有名称以及流水号
			return fmt.Errorf("can't incr")
		}
		if tc.Total == "" {
			tc.Total = TD_TOTAL_FIELD
		}
		if cc.HExists(tc.Name, tc.Journal).Val() { // 流水号存在,说明已经处理
			return nil
		} else {
			if cc.HSet(tc.Name, tc.Journal, strconv.FormatFloat(amount, 'f', 6, 64)).Val() { // set 成功
				tc.JournalAmount = amount
				return cc.HIncrByFloat(tc.Name, tc.Total, amount).Err()
			}
		}
	} else {
		return fmt.Errorf("not found cache client")
	}
	return fmt.Errorf("incr failed")
}

/* }}} */

/* {{{ func (tc *TemporaryCounter) GetTotalAmount() (float64, error)
 * 获取临时扣除总额
 */
func (tc *TemporaryCounter) GetTotalAmount() (float64, error) {
	if cc := ClusterClient(); cc != nil {
		if tc.Name == "" { //必须有名称
			return 0, fmt.Errorf("can't get")
		}
		if tc.Total == "" {
			tc.Total = TD_TOTAL_FIELD
		}
		if tas, err := cc.HGet(tc.Name, tc.Total).Result(); err != nil {
			return 0, err
		} else {
			return strconv.ParseFloat(tas, 64)
		}
	} else {
		return 0, fmt.Errorf("not found cache client")
	}
}

/* }}} */

/* {{{ func (tc *TemporaryCounter) clearJournal(journal string) error
 * 清除临时扣除
 */
func (tc *TemporaryCounter) clearJournal(journal string) error {
	if cc := ClusterClient(); cc != nil {
		if tc.Name == "" || journal == "" { //必须有名称以及流水号
			return fmt.Errorf("can't clear")
		}
		if tc.Total == "" {
			tc.Total = TD_TOTAL_FIELD
		}
		if as, err := cc.HGet(tc.Name, journal).Result(); err != nil {
			return err
		} else {
			amount, _ := strconv.ParseFloat(as, 64)
			if err := cc.HDel(tc.Name, journal).Err(); err != nil {
				return err
			} else {
				return cc.HIncrByFloat(tc.Name, tc.Total, 0-amount).Err()
			}
		}
	} else {
		return fmt.Errorf("not found cache client")
	}
	return fmt.Errorf("deduct failed")
}

/* }}} */
