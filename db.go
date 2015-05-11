// Ogo

package ogo

import (
	"bytes"
	"fmt"
	"reflect"

	"github.com/Odinman/gorp"
	"github.com/Odinman/ogo/utils"
)

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
