// Ogo

package ogo

import (
	"fmt"
	"reflect"

	"github.com/Odinman/ogo/utils"
)

const (
	// generic action const
	GA_GET = 1 << iota
	GA_SEARCH
	GA_POST
	GA_DELETE
	GA_PATCH
	//GA_PUT
	GA_HEAD
	GA_ALL = GA_GET | GA_SEARCH | GA_POST | GA_DELETE | GA_PATCH | GA_HEAD

	KEY_SKIPAUTH  = "skipauth"
	KEY_SKIPLOGIN = "skiplogin"
	KEY_SKIPPERM  = "skipperm"
)

type ActionInterface interface {
	PreGet(i interface{}) (interface{}, error)  //获取前
	OnGet(i interface{}) (interface{}, error)   //获取中
	PostGet(i interface{}) (interface{}, error) //获取后

	PreSearch(i interface{}) (interface{}, error)  // 搜索前的检查
	OnSearch(i interface{}) (interface{}, error)   // 搜索前的检查
	PostSearch(i interface{}) (interface{}, error) // 搜索后的检查

	PreCreate(i interface{}) (interface{}, error)  // 插入前的检查
	OnCreate(i interface{}) (interface{}, error)   // 插入前的检查
	PostCreate(i interface{}) (interface{}, error) // 插入后的处理

	PreUpdate(i interface{}) (interface{}, error)  // 更新前的检查
	OnUpdate(i interface{}) (interface{}, error)   // 更新前的检查
	PostUpdate(i interface{}) (interface{}, error) // 更新后的操作

	PreDelete(i interface{}) (interface{}, error)  // 删除前的检查
	OnDelete(i interface{}) (interface{}, error)   // 删除前的检查
	PostDelete(i interface{}) (interface{}, error) // 删除后的检查

	PreCheck(i interface{}) (interface{}, error)  // 搜索前的检查
	OnCheck(i interface{}) (interface{}, error)   // 搜索前的检查
	PostCheck(i interface{}) (interface{}, error) // 搜索后的检查

	Trigger(i interface{}) (interface{}, error) //触发器
}

/* {{{ func (_ *Router) Trigger(i interface{}) (interface{}, error)
 *
 */
func (_ *Router) Trigger(i interface{}) (interface{}, error) {
	return i, nil
}

/* }}} */

/* {{{ func (_ *Router) PreGet(i interface{}) (interface{}, error)
 *
 */
func (_ *Router) PreGet(i interface{}) (interface{}, error) {
	m := i.(Model)
	c := m.GetCtx()
	// pk,放入条件
	id := c.URLParams[RowkeyKey]
	pk := m.PKey()
	c.Debug("[PreGet][pk: %s, id: %s]", pk, id)
	m.SetConditions(NewCondition(CTYPE_IS, pk, id))
	// 从restcontext里获取条件
	if tr := c.GetEnv(TimeRangeKey); tr != nil { //时间段参数
		m.SetConditions(NewCondition(CTYPE_IS, TAG_TIMERANGE, tr.(*TimeRange)))
	}
	if cons := c.GetEnv(ConditionsKey); cons != nil { //从context里面获取参数条件
		m.SetConditions(cons.([]*Condition)...)
	}
	// fields
	if fs := c.GetEnv(FieldsKey); fs != nil { //从context里面获取参数条件
		m.SetFields(fs.([]string))
	}
	return i, nil
}

/* }}} */
/* {{{ func (_ *Router) OnGet(i interface{}) (interface{}, error)
 *
 */
func (_ *Router) OnGet(i interface{}) (interface{}, error) {
	m := i.(Model)
	//c := m.GetCtx()
	//id := c.URLParams[RowkeyKey]
	return m.GetRow(m)
}

/* }}} */
/* {{{ func (_ *Router) PostGet(i interface{}) (interface{}, error)
 *
 */
func (_ *Router) PostGet(i interface{}) (interface{}, error) {
	if cols := utils.ReadStructColumns(i, true); cols != nil {
		v := reflect.ValueOf(i)
		for _, col := range cols {
			if col.ExtOptions.Contains(TAG_SECRET) { //保密,不对外
				fv := utils.FieldByIndex(v, col.Index)
				fv.Set(reflect.Zero(fv.Type()))
			}
		}
	}
	return i, nil

}

/* }}} */

/* {{{ func (_ *Router) PreSearch(i interface{}) (interface{}, error)
 *
 */
func (_ *Router) PreSearch(i interface{}) (interface{}, error) {
	m := i.(Model)
	c := m.GetCtx()
	// 从restcontext里获取条件
	if p := c.GetEnv(PaginationKey); p != nil { //排序
		m.SetPagination(p.(*Pagination))
	}
	if ob := c.GetEnv(OrderByKey); ob != nil { //排序
		m.SetConditions(NewCondition(CTYPE_ORDER, TAG_ORDERBY, ob.(*OrderBy)))
	}
	if tr := c.GetEnv(TimeRangeKey); tr != nil { //时间段参数
		m.SetConditions(NewCondition(CTYPE_RANGE, TAG_TIMERANGE, tr.(*TimeRange)))
	}
	if cons := c.GetEnv(ConditionsKey); cons != nil { //从context里面获取参数条件
		m.SetConditions(cons.([]*Condition)...)
	}
	// fields
	if fs := c.GetEnv(FieldsKey); fs != nil { //从context里面获取参数条件
		m.SetFields(fs.([]string))
	}
	return i, nil
}

/* }}} */
/* {{{ func (_ *Router) OnSearch(i interface{}) (interface{}, error)
 *
 */
func (_ *Router) OnSearch(i interface{}) (interface{}, error) {
	m := i.(Model)
	return m.GetRows()
}

/* }}} */
/* {{{ func (_ *Router) PostSearch(i interface{}) (interface{}, error)
 *
 */
func (_ *Router) PostSearch(i interface{}) (interface{}, error) {
	return i, nil
}

/* }}} */

/* {{{ func (_ *Router) PreCreate(i interface{}) (interface{}, error)
 *
 */
func (_ *Router) PreCreate(i interface{}) (interface{}, error) {
	m := i.(Model)
	c := m.GetCtx()
	var err error
	if m, err = m.Valid(); err != nil {
		return nil, err
	}
	v := reflect.ValueOf(m)
	// existense checker
	eChecker := m.Existense()
	if cols := utils.ReadStructColumns(m, true); cols != nil {
		for _, col := range cols {
			fv := utils.FieldByIndex(v, col.Index)
			//c.Trace("name:%s; kind:%v; type:%s; extag: %s", col.Tag, fv.Kind(), fv.Type().String(), col.ExtTag)
			// check required field(when post)
			if col.ExtOptions.Contains(TAG_REQUIRED) && (!fv.IsValid() || utils.IsEmptyValue(fv)) {
				c.Debug("field %s required but empty", col.Tag)
				return nil, fmt.Errorf("field %s required but empty", col.Tag)
			}
			switch col.ExtTag { //根据tag, 会对数据进行预处理
			case "userid": //替换为userid
				var userid string
				if uid := c.GetEnv(USERID_KEY); uid == nil {
					userid = "0"
					c.Debug("userid not exists")
				} else {
					userid = uid.(string)
					c.Debug("userid: %s", userid)
				}
				switch fv.Type().String() {
				case "*string":
					fv.Set(reflect.ValueOf(&userid))
				case "string":
					fv.Set(reflect.ValueOf(userid))
				default:
					return nil, fmt.Errorf("field(%s) must be string, not %s", col.Tag, fv.Kind().String())
				}
			case "existense": //检查存在性
				if exValue, err := eChecker(col.Tag); err != nil {
					return nil, fmt.Errorf("%s existense check failed: %s", col.Tag, err.Error())
				} else {
					c.Debug("%s existense: %v", col.Tag, exValue)
					fv.Set(reflect.ValueOf(exValue))
				}
			case "uuid":
				switch fv.Type().String() {
				case "*string":
					h := utils.NewShortUUID()
					fv.Set(reflect.ValueOf(&h))
				case "string":
					h := utils.NewShortUUID()
					fv.Set(reflect.ValueOf(h))
				default:
					return nil, fmt.Errorf("field(%s) must be string, not %s", col.Tag, fv.Kind().String())
				}
			case "luuid":
				switch fv.Type().String() {
				case "*string":
					h := utils.NewUUID()
					fv.Set(reflect.ValueOf(&h))
				case "string":
					h := utils.NewUUID()
					fv.Set(reflect.ValueOf(h))
				default:
					return nil, fmt.Errorf("field(%s) must be string, not %s", col.Tag, fv.Kind().String())
				}
			default:
			}
		}
	}
	return m, nil
}

/* }}} */
/* {{{ func (_ *Router) OnCreate(i interface{}) (interface{}, error)
 *
 */
func (_ *Router) OnCreate(i interface{}) (interface{}, error) {
	m := i.(Model)
	if r, err := m.CreateRow(); err != nil {
		return nil, err
	} else {
		return r, nil
	}
}

/* }}} */
/* {{{ func (_ *Router) PostCreate(i interface{}) (interface{}, error)
 *
 */
func (_ *Router) PostCreate(i interface{}) (interface{}, error) {
	m := i.(Model)
	return m.Filter()
}

/* }}} */

/* {{{ func (_ *Router) PreUpdate(i interface{}) (interface{}, error)
 *
 */
func (_ *Router) PreUpdate(i interface{}) (interface{}, error) {
	m := i.(Model)
	c := m.GetCtx()
	var err error
	if m, err = m.Valid(); err != nil {
		return nil, err
	}

	var rk string
	var ok bool
	if rk, ok = c.URLParams[RowkeyKey]; !ok {
		return nil, fmt.Errorf("rowkey empty")
	}
	// old
	var older Model
	if older, err = m.GetRow(rk); err != nil {
		return nil, err
	}
	v := reflect.ValueOf(m)
	// existense checker
	eChecker := m.Existense()
	if cols := utils.ReadStructColumns(m, true); cols != nil {
		for _, col := range cols {
			fv := utils.FieldByIndex(v, col.Index)
			//c.Trace("name:%s; kind:%v; type:%s; extag: %s", col.Tag, fv.Kind(), fv.Type().String(), col.ExtTag)
			// check required field(when post)
			if fv.IsValid() && !utils.IsEmptyValue(fv) {
				if col.ExtOptions.Contains(TAG_DENY) { //尝试编辑不可编辑的字段,要报错
					c.Info("%s is uneditable: %v", col.Tag, fv)
					return nil, fmt.Errorf("%s is uneditable", col.Tag) //尝试编辑不可编辑的字段,直接报错
				} else { //忽略
					continue
				}
			}
			// server generate,忽略传入的信息
			switch col.ExtTag { //根据tag, 会对数据进行预处理
			case "userid": //替换为userid
				var userid string
				if uid := c.GetEnv(USERID_KEY); uid == nil {
					userid = "0"
					c.Debug("userid not exists")
				} else {
					userid = uid.(string)
					c.Debug("userid: %s", userid)
				}
				switch fv.Type().String() {
				case "*string":
					fv.Set(reflect.ValueOf(&userid))
				case "string":
					fv.Set(reflect.ValueOf(userid))
				default:
					return nil, fmt.Errorf("field(%s) must be string, not %s", col.Tag, fv.Kind().String())
				}
			case "existense": //检查存在性
				if fv.IsValid() && !utils.IsEmptyValue(fv) { //update时,传入才检查
					if exValue, err := eChecker(col.Tag); err != nil {
						return nil, fmt.Errorf("%s existense check failed: %s", col.Tag, err.Error())
					} else {
						c.Debug("%s existense: %v", col.Tag, exValue)
						fv.Set(reflect.ValueOf(exValue))
					}
				}
			case "forbbiden": //这个字段如果旧记录有值, 则返回错误
				ov := reflect.ValueOf(older)
				fov := utils.FieldByIndex(ov, col.Index)
				if fov.IsValid() && !utils.IsEmptyValue(fov) {
					return nil, fmt.Errorf("field(%s) has value, can't be updated", col.Tag)
				}
			default:
			}
		}
	}
	return m, nil
}

/* }}} */
/* {{{ func (_ *Router) OnUpdate(i interface{}) (interface{}, error)
 *
 */
func (_ *Router) OnUpdate(i interface{}) (interface{}, error) {
	m := i.(Model)
	c := m.GetCtx()
	rk := c.URLParams[RowkeyKey]
	if affected, err := m.UpdateRow(rk); err != nil {
		return nil, err
	} else {
		if affected <= 0 {
			c.Info("OnUpdate not affected any record")
		}
		return m, nil
	}
}

/* }}} */
/* {{{ func (_ *Router) PostUpdate(i interface{}) (interface{}, error)
 *
 */
func (_ *Router) PostUpdate(i interface{}) (interface{}, error) {
	m := i.(Model)
	return m.Filter()
}

/* }}} */

/* {{{ func (_ *Router) PreDelete(i interface{}) (interface{}, error)
 *
 */
func (_ *Router) PreDelete(i interface{}) (interface{}, error) {
	return i, nil
}

/* }}} */
/* {{{ func (_ *Router) OnDelete(i interface{}) (interface{}, error)
 *
 */
func (_ *Router) OnDelete(i interface{}) (interface{}, error) {
	m := i.(Model)
	c := m.GetCtx()
	rk := c.URLParams[RowkeyKey]
	if affected, err := m.DeleteRow(rk); err != nil {
		return nil, err
	} else {
		if affected <= 0 {
			c.Info("OnDelete not affected any record")
		}
		return m, nil
	}
}

/* }}} */
/* {{{ func (_ *Router) PostDelete(i interface{}) (interface{}, error)
 *
 */
func (_ *Router) PostDelete(i interface{}) (interface{}, error) {
	return i, nil
}

/* }}} */

/* {{{ func (_ *Router) PreCheck(i interface{}) (interface{}, error)
 *
 */
func (_ *Router) PreCheck(i interface{}) (interface{}, error) {
	c := i.(Model).GetCtx()
	m := i.(Model)
	// 从restcontext里获取条件
	if tr := c.GetEnv(TimeRangeKey); tr != nil { //时间段参数
		m.SetConditions(NewCondition(CTYPE_IS, TAG_TIMERANGE, tr.(*TimeRange)))
	}
	if cons := c.GetEnv(ConditionsKey); cons != nil { //从context里面获取参数条件
		m.SetConditions(cons.([]*Condition)...)
	}
	return i, nil
}

/* }}} */
/* {{{ func (_ *Router) OnCheck(i interface{}) (interface{}, error)
 *
 */
func (_ *Router) OnCheck(i interface{}) (interface{}, error) {
	m := i.(Model)
	return m.GetCount()
}

/* }}} */
/* {{{ func (_ *Router) PostCheck(i interface{}) (interface{}, error)
 *
 */
func (_ *Router) PostCheck(i interface{}) (interface{}, error) {
	return i, nil
}

/* }}} */
