// Ogo

package ogo

import ()

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
	Defer(i interface{})                        //触发器
}

/* {{{ func (_ *Router) Trigger(i interface{}) (interface{}, error)
 *
 */
func (_ *Router) Trigger(i interface{}) (interface{}, error) {
	return i, nil
}

/* }}} */
/* {{{ func (_ *Router) Defer(i interface{})
 *
 */
func (_ *Router) Defer(i interface{}) {
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
	pk, _ := m.PKey()
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
	//m := i.(Model)
	//return m.Protect()
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
	var err error
	if m, err = m.Valid(); err != nil {
		return nil, err
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
	var err error
	if m, err = m.Valid(); err != nil {
		return nil, err
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
