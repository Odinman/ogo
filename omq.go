// Ogo

package ogo

import (
	"fmt"
	"time"
)

/* {{{ func OmqSet(key, value string, expire int) error
 *
 */
func OmqSet(key, value string, expire int) error {
	if requester, e := OmqPool().Get(); e == nil {
		defer requester.Close()
		if reply, e := requester.Do(10000*time.Millisecond, "SET", "redis", key, value, expire); e == nil {
			Debug("Received: %s", reply[0])
			if reply[0] == "OK" {
				return nil
			}
		} else {
			Info("set %s error: %s", key, e)
			return e
		}
	} else {
		return e
	}
	return fmt.Errorf("set failed")
}

/* }}} */

/* {{{ func OmqGet(key string) (string, error)
 *
 */
func OmqGet(key string) (string, error) {
	if requester, e := OmqPool().Get(); e == nil {
		defer requester.Close()
		if reply, e := requester.Do(10000*time.Millisecond, "GET", "redis", key); e == nil {
			Debug("Received: %s", reply)
			if reply[0] == "OK" {
				return reply[1], nil
			}
		} else {
			Info("get %s error: %s", key, e)
			return "", e
		}
	} else {
		return "", e
	}
	return "", fmt.Errorf("get %s failed", key)
}

/* }}} */

/* {{{ func OmqDel(key string) (string, error)
 *
 */
func OmqDel(key string) (string, error) {
	if requester, e := OmqPool().Get(); e == nil {
		defer requester.Close()
		if reply, e := requester.Do(10000*time.Millisecond, "DEL", "redis", key); e == nil {
			Debug("Received: %s", reply[0])
			if reply[0] != "" {
				return reply[0], nil
			}
		} else {
			Info("del %s error: %s", key, e)
			return "", e
		}
	} else {
		return "", e
	}
	return "", fmt.Errorf("del %s failed", key)
}

/* }}} */

/* {{{ func OmqTask(msg ...string) error
 *
 */
func OmqTask(msg ...string) error {
	if requester, e := OmqPool().Get(); e == nil && len(msg) > 1 {
		defer requester.Close()
		key := msg[0]
		values := msg[1:]
		if reply, e := requester.Do(10000*time.Millisecond, "TASK", key, values); e == nil {
			Debug("Received: %s", reply[0])
			if reply[0] == "OK" {
				return nil
			}
		} else {
			Info("task %s error: %s", key, e)
			return e
		}
	} else {
		return e
	}
	return fmt.Errorf("push task failed")
}

/* }}} */

/* {{{ func OmqBlockTask(msg ...string) (string, error)
 *
 */
func OmqBlockTask(msg ...string) (string, error) {
	if requester, e := OmqPool().Get(); e == nil && len(msg) > 1 {
		defer requester.Close()
		key := msg[0]
		values := msg[1:]
		if reply, e := requester.Do(13*time.Second, "BTASK", key, values); e == nil {
			Debug("Received: %s", reply)
			if reply[0] == "OK" {
				return reply[1], nil
			}
		} else {
			Info("task %s error: %s", key, e)
			return "", e
		}
	} else {
		return "", e
	}
	return "", fmt.Errorf("task_failed")
}

/* }}} */

/* {{{ func OmqPop(msg ...string) error
 *
 */
func OmqPop(msg ...string) ([]string, error) {
	if requester, e := OmqPool().Get(); e == nil && len(msg) > 0 {
		defer requester.Close()
		key := msg[0]
		//values := msg[1:]
		if reply, e := requester.Do(10000*time.Millisecond, "POP", key); e == nil {
			Debug("Received: %s", reply[0])
			if reply[0] == "OK" {
				return reply[1:], e
			} else {
				Info("task %s is empty", key)
				return nil, nil
			}
		} else {
			Info("task %s error: %s", key, e)
			return nil, e
		}
	} else {
		return nil, e
	}
	return nil, fmt.Errorf("pop task failed")
}

/* }}} */

func OmqBPop(msg string, timeout int) ([]string, error) {
	if requester, e := OmqPool().Get(); e == nil && len(msg) > 0 {
		defer requester.Close()
		key := msg
		if reply, e := requester.Do(10000*time.Millisecond, "BPOP", key, timeout); e == nil {
			Debug("Received: %s", reply[0])
			if reply[0] == "OK" {
				return reply[1:], e
			} else {
				Info("task %s is empty", key)
				return nil, nil
			}
		} else {
			Info("task %s error: %s", key, e)
			return nil, e
		}
	} else {
		return nil, e
	}
	return nil, fmt.Errorf("bpop task failed")
}
