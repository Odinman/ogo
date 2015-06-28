// Ogo

package ogo

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/Odinman/gorp"
	//"github.com/Odinman/ogo/utils"
	_ "github.com/go-sql-driver/mysql"
)

var (
	DataAccessor = make(map[string]string) // tablename::{read/write} => tag
)

const (
	DBTAG      string = "db"
	READTAG    string = "read"
	WRITETAG   string = "write"
	base       string = "0000-00-00 00:00:00.0000000"
	timeFormat string = "2006-01-02 15:04:05.999999"
)

//出/入库转换器
type SelfConverter interface {
	ToDb() (interface{}, error)                    //入库
	FromDb(interface{}) (gorp.CustomScanner, bool) //出库
}

type BaseConverter struct{}

/* {{{ func OpenDB(tag,dns string) error
 *
 */
func OpenDB(tag, dns string) (err error) {
	Debug("open mysql: %s,%s", tag, dns)
	gorp.TraceOn(fmt.Sprintf("[%s]", tag), Logger())
	gorp.SetTypeConvert(BaseConverter{})
	if err = gorp.Open(tag, "mysql", dns); err != nil {
		Debug("open error: %s", err)
	}
	return
}

/* }}} */

/* {{{ func (_ BaseConverter) ToDb(val interface{}) (interface{}, error)
 *
 */
func (_ BaseConverter) ToDb(val interface{}) (interface{}, error) {
	switch t := val.(type) {
	case *[]string, []string, map[string]string, *map[string]string, map[string]interface{}, *map[string]interface{}: //转为字符串
		c, _ := json.Marshal(t)
		return string(c), nil
	//case *float64:
	//	ot := utils.ParseFloat(*t)
	//	Info("float: %f", ot)
	//	return ot, nil
	//case float64:
	//	return utils.ParseFloat(t), nil
	default:
		// 自定义的类型,如果实现了SelfConverter接口,则这里自动执行
		if _, ok := val.(SelfConverter); ok {
			Trace("selfconvert todb")
			return val.(SelfConverter).ToDb()
		} else if _, ok := reflect.Indirect(reflect.ValueOf(val)).Interface().(SelfConverter); ok { //如果采用了指针, 则到这里
			Trace("prt selfconvert todb")
			return val.(SelfConverter).ToDb()
		} else {
			Trace("not selfconvert todb")
		}
	}
	return val, nil
}

/* }}} */

/* {{{ func (_ BaseConverter) FromDb(target interface{}) (gorp.CustomScanner, bool)
 * 类型转换, 主要处理标准类型; 自定义类型通过SelfConverter实现
 */
func (_ BaseConverter) FromDb(target interface{}) (gorp.CustomScanner, bool) {
	switch t := target.(type) {
	case **time.Time:
		binder := func(holder, target interface{}) error {
			var err error
			if holder.(*sql.NullString).Valid {
				var dt time.Time
				str := holder.(*sql.NullString).String
				switch len(str) {
				case 10, 19, 21, 22, 23, 24, 25, 26: // up to "YYYY-MM-DD HH:MM:SS.MMMMMM"
					if str == base[:len(str)] {
						return nil
					}
					dt, err = time.ParseInLocation(timeFormat[:len(str)], str, time.Local)
				default:
					err = fmt.Errorf("Invalid Time-String: %s", str)
					return err
				}
				if err != nil {
					return err
				}
				//dt = dt.UTC()
				dt = dt.Local()
				*(target.(**time.Time)) = &dt
				return nil
			}
			return nil
		}
		return gorp.CustomScanner{new(sql.NullString), target, binder}, true
	case *time.Time:
		binder := func(holder, target interface{}) error {
			var err error
			if holder.(*sql.NullString).Valid {
				var dt time.Time
				str := holder.(*sql.NullString).String
				switch len(str) {
				case 10, 19, 21, 22, 23, 24, 25, 26: // up to "YYYY-MM-DD HH:MM:SS.MMMMMM"
					if str == base[:len(str)] {
						return nil
					}
					dt, err = time.ParseInLocation(timeFormat[:len(str)], str, time.Local)
				default:
					err = fmt.Errorf("Invalid Time-String: %s", str)
					return err
				}
				if err != nil {
					return err
				}
				//dt = dt.UTC()
				dt = dt.Local()
				*(target.(*time.Time)) = dt
				return nil
			}
			return nil
		}
		return gorp.CustomScanner{new(sql.NullString), target, binder}, true
	case **[]string:
		binder := func(holder, target interface{}) error {
			if holder.(*sql.NullString).Valid {
				var st []string
				str := holder.(*sql.NullString).String
				//ogo.Debug("str: %s", str)
				if err := json.Unmarshal([]byte(str), &st); err != nil {
					return err
				}
				*(target.(**[]string)) = &st
			}
			return nil
		}
		return gorp.CustomScanner{new(sql.NullString), target, binder}, true
	case *[]string:
		binder := func(holder, target interface{}) error {
			if holder.(*sql.NullString).Valid {
				var st []string
				str := holder.(*sql.NullString).String
				//ogo.Debug("str: %s", str)
				if err := json.Unmarshal([]byte(str), &st); err != nil {
					return err
				}
				*(target.(*[]string)) = st
			}
			return nil
		}
		return gorp.CustomScanner{new(sql.NullString), target, binder}, true
	case **map[string]string:
		binder := func(holder, target interface{}) error {
			if holder.(*sql.NullString).Valid {
				var st map[string]string
				str := holder.(*sql.NullString).String
				//ogo.Debug("str: %s", str)
				if err := json.Unmarshal([]byte(str), &st); err != nil {
					return err
				}
				*(target.(**map[string]string)) = &st
			}
			return nil
		}
		return gorp.CustomScanner{new(sql.NullString), target, binder}, true
	case *map[string]string:
		binder := func(holder, target interface{}) error {
			if holder.(*sql.NullString).Valid {
				var st map[string]string
				str := holder.(*sql.NullString).String
				//ogo.Debug("str: %s", str)
				if err := json.Unmarshal([]byte(str), &st); err != nil {
					return err
				}
				*(target.(*map[string]string)) = st
			}
			return nil
		}
		return gorp.CustomScanner{new(sql.NullString), target, binder}, true
	case **map[string]interface{}:
		binder := func(holder, target interface{}) error {
			if holder.(*sql.NullString).Valid {
				var st map[string]interface{}
				str := holder.(*sql.NullString).String
				//ogo.Debug("str: %s", str)
				if err := json.Unmarshal([]byte(str), &st); err != nil {
					return err
				}
				*(target.(**map[string]interface{})) = &st
			}
			return nil
		}
		return gorp.CustomScanner{new(sql.NullString), target, binder}, true
	case *map[string]interface{}:
		binder := func(holder, target interface{}) error {
			if holder.(*sql.NullString).Valid {
				var st map[string]interface{}
				str := holder.(*sql.NullString).String
				//ogo.Debug("str: %s", str)
				if err := json.Unmarshal([]byte(str), &st); err != nil {
					return err
				}
				*(target.(*map[string]interface{})) = st
			}
			return nil
		}
		return gorp.CustomScanner{new(sql.NullString), target, binder}, true
	case **string:
		binder := func(holder, target interface{}) error {
			*t = &holder.(*sql.NullString).String
			return nil
		}
		return gorp.CustomScanner{new(sql.NullString), target, binder}, true
	case **float64:
		binder := func(holder, target interface{}) error {
			*t = &holder.(*sql.NullFloat64).Float64
			return nil
		}
		return gorp.CustomScanner{new(sql.NullFloat64), target, binder}, true

	case **int64:
		binder := func(holder, target interface{}) error {
			*t = &holder.(*sql.NullInt64).Int64
			return nil
		}
		return gorp.CustomScanner{new(sql.NullInt64), target, binder}, true
	default:
		// 自定义的类型,如果实现了SelfConverter接口,则这里自动执行
		if t, ok := target.(SelfConverter); ok {
			Trace("selfconvert begin(value)")
			return t.FromDb(target)
		} else if t, ok := reflect.Indirect(reflect.ValueOf(target)).Interface().(SelfConverter); ok { //如果采用了指针, 则到这里
			Trace("ptr converter: %s", target)
			return t.FromDb(target)
		} else {
			Trace("no converter: %s", target)
		}
	}
	return gorp.CustomScanner{}, false
}

/* }}} */
