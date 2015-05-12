// Ogo

package ogo

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/Odinman/gorp"
	"github.com/Odinman/ogo/utils"
)

const (
	base       string = "0000-00-00 00:00:00.0000000"
	timeFormat string = "2006-01-02 15:04:05.999999"
)

type List struct {
	Total int64       `json:"total"`
	List  interface{} `json:"list"`
}

//出/入库转换器
type SelfConverter interface {
	ToDb() (interface{}, error)                    //入库
	FromDb(interface{}) (gorp.CustomScanner, bool) //出库
}

type BaseConverter struct{}

/* {{{ func (_ BaseConverter) ToDb(val interface{}) (interface{}, error)
 *
 */
func (_ BaseConverter) ToDb(val interface{}) (interface{}, error) {
	switch t := val.(type) {
	case *[]string, []string, map[string]string, *map[string]string: //转为字符串
		c, _ := json.Marshal(t)
		return string(c), nil
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

/* {{{ func ReadPrepare(m Model, cons Conditions, c *RESTContext) (b *gorp.Builder, err error)
 * 查询准备
 */
func ReadPrepare(m Model, cons Conditions, c *RESTContext) (b *gorp.Builder, err error) {
	db := m.DBConn("read")
	tb := m.TableName(m)
	b = gorp.NewBuilder(db).Table(tb)

	// time range, 凡有time range的表都应该加上索引
	if tr := c.GetEnv(TimeRangeKey); tr != nil {
		//存在timerange条件
		if cols := utils.ReadStructColumns(m, true); cols != nil {
			for _, col := range cols {
				if col.ExtOptions.Contains(TAG_TIMERANGE) { //时间范围
					c.Debug("time range field: %s, start: %s, end: %s", col.Tag, tr.(*TimeRange).Start, tr.(*TimeRange).End)
					b.Where(fmt.Sprintf("T.`%s` BETWEEN ? AND ?", col.Tag), tr.(*TimeRange).Start, tr.(*TimeRange).End)
				}
			}
		}
	}

	ordered := false
	if ob := c.GetEnv(OrderByKey); ob != nil {
		if cols := utils.ReadStructColumns(m, true); cols != nil {
			for _, col := range cols {
				if col.ExtOptions.Contains(TAG_ORDERBY) && col.Tag == ob.(*OrderBy).Field { //排序
					c.Debug("order by field: %s, sort: %s", ob.(*OrderBy).Field, ob.(*OrderBy).Sort)
					b.Order(fmt.Sprintf("T.`%s` %s", ob.(*OrderBy).Field, ob.(*OrderBy).Sort))
					ordered = true
				}
			}
		}
	}
	if !ordered {
		//默认排序
		if cols := utils.ReadStructColumns(m, true); cols != nil {
			for _, col := range cols {
				if col.TagOptions.Contains(DBTAG_PK) { // 默认为pk降序
					b.Order(fmt.Sprintf("T.`%s` DESC", col.Tag))
				}
			}
		}
	}

	// permission
	//b.Where(PermCondition(c, m))

	// condition
	if len(cons) > 0 {
		jc := 0
		for _, v := range cons {
			c.Trace("cons value: %v", v)
			if v.Is != nil {
				switch vt := v.Is.(type) {
				case string:
					b.Where(fmt.Sprintf("T.`%s` = ?", v.Field), vt)
				case []string:
					vs := bytes.Buffer{}
					first := true
					vs.WriteString("(")
					for _, vv := range vt {
						if !first {
							vs.WriteString(",")
						}
						vs.WriteString(fmt.Sprintf("'%s'", vv))
						first = false
					}
					vs.WriteString(")")
					b.Where(fmt.Sprintf("T.`%s` IN %s", v.Field, vs.String()))
				default:
				}
			}
			if v.Not != nil {
				switch vt := v.Not.(type) {
				case string:
					b.Where(fmt.Sprintf("T.`%s` != ?", v.Field), vt)
				case []string:
					vs := bytes.Buffer{}
					first := true
					vs.WriteString("(")
					for _, vv := range vt {
						if !first {
							vs.WriteString(",")
						}
						vs.WriteString(fmt.Sprintf("'%s'", vv))
						first = false
					}
					vs.WriteString(")")
					b.Where(fmt.Sprintf("T.`%s` NOT IN %s", v.Field, vs.String()))
				default:
				}
			}
			if v.Like != nil {
				switch vt := v.Like.(type) {
				case string:
					b.Where(fmt.Sprintf("T.`%s` LIKE '%%%s%%'", v.Field, vt))
				case []string:
					vs := bytes.Buffer{}
					first := true
					vs.WriteString("(")
					for _, vv := range vt {
						if !first {
							vs.WriteString(" OR ")
						}
						vs.WriteString(fmt.Sprintf("T.`%s` LIKE '%%%s%%'", v.Field, vv))
						first = false
					}
					vs.WriteString(")")
					b.Where(vs.String())
				default:
				}
			}
			if v.Join != nil { //关联查询
				switch vt := v.Join.(type) {
				case *Condition:
					c.Trace("%s will join %s.%s", v.Field, v.Field, vt.Field)
					if vt.Is != nil {
						jt := v.Field
						jf := vt.Field
						var canJoin bool
						if t, ok := gorp.GetTable(jt); ok {
							c.Trace("table: %s; type name: %s", jt, t.Gotype.Name())
							if cols := utils.ReadStructColumns(reflect.New(t.Gotype).Interface(), true); cols != nil {
								for _, col := range cols {
									c.Trace("%s %s", jt, col.Tag)
									if col.Tag == jf && col.ExtOptions.Contains(TAG_CONDITION) { //可作为条件
										c.Trace("%s.%s can join", jt, jf)
										canJoin = true
										break
									}
								}
							}
						} else {
							c.Info("unknown table %s", jt)
						}
						if canJoin {
							js := fmt.Sprintf("LEFT JOIN `%s` T%d ON T.`%s` = T%d.`id`", jt, jc, v.Field, jc)
							b.Joins(js)
							b.Where(fmt.Sprintf("T%d.`%s`=?", jc, jf), vt.Is.(string))
							jc++
						} else {
							c.Trace("%s.%s can't join", jt, jf)
						}
					}
				default:
					c.Trace("not support !*Condition: %v", vt)
				}
			}
		}
	}
	//b.Where(SkipLogicDeleted(m))
	return
}

/* }}} */

/* {{{ func ReportPrepare(model interface{}, cons ogo.Conditions, c *ogo.RESTContext) (b *gorp.Builder, err error)
 * 报表查询准备
 */
func ReportPrepare(m Model, cons Conditions, c *RESTContext) (b *gorp.Builder, err error) {
	db := m.DBConn("read")
	tb := m.TableName(m.(Model))
	b = gorp.NewBuilder(db).Table(tb)

	// time range, 凡可进行time range的字段都应该加上索引
	if tr := c.GetEnv(TimeRangeKey); tr != nil {
		//存在timerange条件
		if cols := utils.ReadStructColumns(m, true); cols != nil {
			for _, col := range cols {
				if col.ExtOptions.Contains(TAG_TIMERANGE) { //时间范围
					c.Debug("time range field: %s, start: %s, end: %s", col.Tag, tr.(*TimeRange).Start, tr.(*TimeRange).End)
					b.Where(fmt.Sprintf("T.`%s` BETWEEN ? AND ?", col.Tag), tr.(*TimeRange).Start, tr.(*TimeRange).End)
				}
			}
		}
	}
	// order by
	if ob := c.GetEnv(OrderByKey); ob != nil {
		if cols := utils.ReadStructColumns(m, true); cols != nil {
			for _, col := range cols {
				if col.ExtOptions.Contains(TAG_ORDERBY) && col.Tag == ob.(*OrderBy).Field { //排序
					c.Debug("order by field: %s, sort: %s", ob.(*OrderBy).Field, ob.(*OrderBy).Sort)
					b.Order(fmt.Sprintf("T.`%s` %s", ob.(*OrderBy).Field, ob.(*OrderBy).Sort))
				}
			}
		}
	}

	// permission
	//b.Where(PermCondition(c, m))

	// condition
	if len(cons) > 0 {
		for _, v := range cons {
			c.Trace("cons value: %v", v)
			if v.Is != nil {
				switch vt := v.Is.(type) {
				case string:
					b.Where(fmt.Sprintf("T.`%s` = ?", v.Field), vt)
				case []string:
					vs := bytes.Buffer{}
					first := true
					vs.WriteString("(")
					for _, vv := range vt {
						if !first {
							vs.WriteString(",")
						}
						vs.WriteString(fmt.Sprintf("'%s'", vv))
						first = false
					}
					vs.WriteString(")")
					b.Where(fmt.Sprintf("T.`%s` IN %s", v.Field, vs.String()))
				default:
				}
			}
			if v.Not != nil {
				switch vt := v.Not.(type) {
				case string:
					b.Where(fmt.Sprintf("T.`%s` != ?", v.Field), vt)
				case []string:
					vs := bytes.Buffer{}
					first := true
					vs.WriteString("(")
					for _, vv := range vt {
						if !first {
							vs.WriteString(",")
						}
						vs.WriteString(fmt.Sprintf("'%s'", vv))
						first = false
					}
					vs.WriteString(")")
					b.Where(fmt.Sprintf("T.`%s` NOT IN %s", v.Field, vs.String()))
				default:
				}
			}
			if v.Like != nil {
				switch vt := v.Like.(type) {
				case string:
					b.Where(fmt.Sprintf("T.`%s` LIKE '%%%s%%'", v.Field, vt))
				case []string:
					vs := bytes.Buffer{}
					first := true
					vs.WriteString("(")
					for _, vv := range vt {
						if !first {
							vs.WriteString(" OR ")
						}
						vs.WriteString(fmt.Sprintf("T.`%s` LIKE '%%%s%%'", v.Field, vv))
						first = false
					}
					vs.WriteString(")")
					b.Where(vs.String())
				default:
				}
			}
		}
	}

	//权限相关,todo
	return
}

/* }}} */

/* {{{ GetDbFields(i interface{}) (s string)
 * 从struct中解析数据库字段以及字段选项
 */
func GetDbFields(i interface{}) (s []string) {
	if cols := utils.ReadStructColumns(i, true); cols != nil {
		s = make([]string, 0)
		for _, col := range cols {
			if col.ExtOptions.Contains(TAG_SECRET) { //保密,不对外
				continue
			}
			s = append(s, col.Tag)
		}
	}
	return
}

/* }}} */
