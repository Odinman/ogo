// Ogo

package ogo

import (
	"encoding/json"
	"fmt"
	"net/http"
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

type Action interface {
	CRUD(i interface{}, flag int) Handler

	Valid(i interface{}) (interface{}, error)   //数据验证
	Filter(i interface{}) (interface{}, error)  //数据过滤
	Trigger(i interface{}) (interface{}, error) //触发器

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
}

/* {{{ func (_ *BaseModel) Valid(i interface{}) (interface{}, error)
 *
 */
func (_ *BaseModel) Valid(i interface{}) (interface{}, error) {
	m := i.(Model)
	c := m.GetCtx()
	if err := json.Unmarshal(c.RequestBody, m); err != nil {
		return nil, err
	}
	v := reflect.ValueOf(m)
	if cols := utils.ReadStructColumns(m, true); cols != nil {
		for _, col := range cols {
			fv := utils.FieldByIndex(v, col.Index)
			// server generate,忽略传入的信息
			if col.ExtOptions.Contains(TAG_GENERATE) && fv.IsValid() && !utils.IsEmptyValue(fv) { // 服务器生成
				fv.Set(reflect.Zero(fv.Type()))
			}
			switch col.ExtTag { //根据tag, 会对数据进行预处理
			case "sha1":
				if fv.IsValid() && !utils.IsEmptyValue(fv) { //不能为空
					switch fv.Type().String() {
					case "*string":
						sv := fv.Elem().String()
						h := utils.HashSha1(sv)
						fv.Set(reflect.ValueOf(&h))
						c.Debug("password: %s, encoded: %s", sv, h)
					case "string":
						sv := fv.String()
						h := utils.HashSha1(sv)
						fv.Set(reflect.ValueOf(h))
						c.Debug("password: %s, encoded: %s", sv, h)
					default:
						return nil, fmt.Errorf("field(%s) must be string, not %s", col.Tag, fv.Kind().String())
					}
				}
			default:
				//可自定义,初始化时放到tagHooks里面
				if col.ExtTag != "" && fv.IsValid() && !utils.IsEmptyValue(fv) { //还必须有值
					if hk, ok := DMux.TagHooks[col.ExtTag]; ok {
						fv.Set(hk(fv))
					} else {
						c.Info("cannot find hook for tag: %s", col.ExtTag)
					}
				}
			}
		}
	}
	return i, nil
}

/* }}} */
/* {{{ func (_ *BaseModel) Filter(i interface{}) (interface{}, error)
 * 处理后过滤
 */
func (_ *BaseModel) Filter(i interface{}) (interface{}, error) {
	//c := i.(Model).GetCtx()
	r := i.(Model).New(i.(Model))
	m := reflect.ValueOf(r)
	v := reflect.ValueOf(i)
	if cols := utils.ReadStructColumns(i, true); cols != nil {
		for _, col := range cols {
			fv := utils.FieldByIndex(v, col.Index)
			mv := utils.FieldByIndex(m, col.Index)
			//c.Trace("field:%s; name: %s, kind:%v; type:%s", col.Tag, col.Name, fv.Kind(), fv.Type().String())
			if col.TagOptions.Contains(DBTAG_PK) || col.ExtOptions.Contains(TAG_RETURN) {
				//pk以及定义了返回tag的赋值
				mv.Set(fv)
			}
		}
	}
	return r.(Model), nil
}

/* }}} */
/* {{{ func (_ *BaseModel) Trigger(i interface{}) (interface{}, error)
 *
 */
func (_ *BaseModel) Trigger(i interface{}) (interface{}, error) {
	return i, nil
}

/* }}} */

/* {{{ func (_ *BaseModel) PreGet(i interface{}) (interface{}, error)
 *
 */
func (_ *BaseModel) PreGet(i interface{}) (interface{}, error) {
	m := i.(Model)
	c := m.GetCtx()
	// pk,放入条件
	id := c.URLParams[RowkeyKey]
	pk := m.PKey(m)
	c.Debug("[PreGet][pk: %s, id: %s]", pk, id)
	m.SetConditions(m, NewCondition(CTYPE_IS, pk, id))
	// 从restcontext里获取条件
	if tr := c.GetEnv(TimeRangeKey); tr != nil { //时间段参数
		m.SetConditions(m, NewCondition(CTYPE_IS, TAG_TIMERANGE, tr.(*TimeRange)))
	}
	if cons := c.GetEnv(ConditionsKey); cons != nil { //从context里面获取参数条件
		m.SetConditions(m, cons.([]*Condition)...)
	}
	return i, nil
}

/* }}} */
/* {{{ func (_ *BaseModel) OnGet(i interface{}) (interface{}, error)
 *
 */
func (_ *BaseModel) OnGet(i interface{}) (interface{}, error) {
	var err error
	m := i.(Model)
	//c := m.GetCtx()
	//id := c.URLParams[RowkeyKey]
	m, err = m.GetRow(m)
	return m, err
}

/* }}} */
/* {{{ func (_ *BaseModel) PostGet(i interface{}) (interface{}, error)
 *
 */
func (_ *BaseModel) PostGet(i interface{}) (interface{}, error) {
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

/* {{{ func (_ *BaseModel) PreSearch(i interface{}) (interface{}, error)
 *
 */
func (_ *BaseModel) PreSearch(i interface{}) (interface{}, error) {
	c := i.(Model).GetCtx()
	m := i.(Model)
	// 从restcontext里获取条件
	if p := c.GetEnv(PaginationKey); p != nil { //排序
		m.SetPagination(p.(*Pagination))
	}
	if ob := c.GetEnv(OrderByKey); ob != nil { //排序
		m.SetConditions(m, NewCondition(CTYPE_ORDER, TAG_ORDERBY, ob.(*OrderBy)))
	}
	if tr := c.GetEnv(TimeRangeKey); tr != nil { //时间段参数
		m.SetConditions(m, NewCondition(CTYPE_RANGE, TAG_TIMERANGE, tr.(*TimeRange)))
	}
	if cons := c.GetEnv(ConditionsKey); cons != nil { //从context里面获取参数条件
		m.SetConditions(m, cons.([]*Condition)...)
	}
	return i, nil
}

/* }}} */
/* {{{ func (_ *BaseModel) OnSearch(i interface{}) (interface{}, error)
 *
 */
func (_ *BaseModel) OnSearch(i interface{}) (interface{}, error) {
	m := i.(Model)
	return m.GetRows(m)
}

/* }}} */
/* {{{ func (_ *BaseModel) PostSearch(i interface{}) (interface{}, error)
 *
 */
func (_ *BaseModel) PostSearch(i interface{}) (interface{}, error) {
	return i, nil
}

/* }}} */

/* {{{ func (_ *BaseModel) PreCreate(i interface{}) (interface{}, error)
 *
 */
func (_ *BaseModel) PreCreate(i interface{}) (interface{}, error) {
	act := i.(Action)
	c := i.(Model).GetCtx()
	var m Model
	if mi, err := act.Valid(i); err != nil {
		return nil, err
	} else {
		m = mi.(Model)
	}
	v := reflect.ValueOf(m)
	// existense checker
	eChecker := m.Existense(m)
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
/* {{{ func (_ *BaseModel) OnCreate(i interface{}) (interface{}, error)
 *
 */
func (_ *BaseModel) OnCreate(i interface{}) (interface{}, error) {
	m := i.(Model)
	if r, err := m.CreateRow(m); err != nil {
		return nil, err
	} else {
		return r, nil
	}
}

/* }}} */
/* {{{ func (_ *BaseModel) PostCreate(i interface{}) (interface{}, error)
 *
 */
func (_ *BaseModel) PostCreate(i interface{}) (interface{}, error) {
	act := i.(Action)
	return act.Filter(i)
}

/* }}} */

/* {{{ func (_ *BaseModel) PreUpdate(i interface{}) (interface{}, error)
 *
 */
func (_ *BaseModel) PreUpdate(i interface{}) (interface{}, error) {
	act := i.(Action)
	c := i.(Model).GetCtx()
	var m Model
	if mi, err := act.Valid(i); err != nil {
		return nil, err
	} else {
		m = mi.(Model)
	}

	//var rk string
	//var ok bool
	//if rk, ok = c.URLParams[RowkeyKey]; !ok {
	//	return nil, fmt.Errorf("rowkey empty")
	//}
	// old
	var older Model
	var err error
	if older, err = m.GetRow(m.New(m)); err != nil {
		return nil, err
	}
	v := reflect.ValueOf(m)
	// existense checker
	eChecker := m.Existense(m)
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
/* {{{ func (_ *BaseModel) OnUpdate(i interface{}) (interface{}, error)
 *
 */
func (_ *BaseModel) OnUpdate(i interface{}) (interface{}, error) {
	m := i.(Model)
	c := m.GetCtx()
	rk := c.URLParams[RowkeyKey]
	if affected, err := m.UpdateRow(m, rk); err != nil {
		return nil, err
	} else {
		if affected <= 0 {
			c.Info("OnUpdate not affected any record")
		}
		return m, nil
	}
}

/* }}} */
/* {{{ func (_ *BaseModel) PostUpdate(i interface{}) (interface{}, error)
 *
 */
func (_ *BaseModel) PostUpdate(i interface{}) (interface{}, error) {
	act := i.(Action)
	return act.Filter(i)
}

/* }}} */

/* {{{ func (_ *BaseModel) PreDelete(i interface{}) (interface{}, error)
 *
 */
func (_ *BaseModel) PreDelete(i interface{}) (interface{}, error) {
	return i, nil
}

/* }}} */
/* {{{ func (_ *BaseModel) OnDelete(i interface{}) (interface{}, error)
 *
 */
func (_ *BaseModel) OnDelete(i interface{}) (interface{}, error) {
	m := i.(Model)
	c := m.GetCtx()
	rk := c.URLParams[RowkeyKey]
	if affected, err := m.DeleteRow(m, rk); err != nil {
		return nil, err
	} else {
		if affected <= 0 {
			c.Info("OnDelete not affected any record")
		}
		return m, nil
	}
}

/* }}} */
/* {{{ func (_ *BaseModel) PostDelete(i interface{}) (interface{}, error)
 *
 */
func (_ *BaseModel) PostDelete(i interface{}) (interface{}, error) {
	return i, nil
}

/* }}} */

/* {{{ func (_ *BaseModel) PreCheck(i interface{}) (interface{}, error)
 *
 */
func (_ *BaseModel) PreCheck(i interface{}) (interface{}, error) {
	c := i.(Model).GetCtx()
	m := i.(Model)
	// 从restcontext里获取条件
	if tr := c.GetEnv(TimeRangeKey); tr != nil { //时间段参数
		m.SetConditions(m, NewCondition(CTYPE_IS, TAG_TIMERANGE, tr.(*TimeRange)))
	}
	if cons := c.GetEnv(ConditionsKey); cons != nil { //从context里面获取参数条件
		m.SetConditions(m, cons.([]*Condition)...)
	}
	return i, nil
}

/* }}} */
/* {{{ func (_ *BaseModel) OnCheck(i interface{}) (interface{}, error)
 *
 */
func (_ *BaseModel) OnCheck(i interface{}) (interface{}, error) {
	m := i.(Model)
	return m.GetCount(m)
}

/* }}} */
/* {{{ func (_ *BaseModel) PostCheck(i interface{}) (interface{}, error)
 *
 */
func (_ *BaseModel) PostCheck(i interface{}) (interface{}, error) {
	return i, nil
}

/* }}} */

/* {{{ func (_ *BaseModel) CRUD(m Model, flag int) Handler
 * 通用的操作方法, 根据flag返回
 * 必须符合通用的restful风格
 */
func (_ *BaseModel) CRUD(i interface{}, flag int) Handler {
	act := i.(Action)
	get := func(c *RESTContext) {
		m := i.(Model).New(i.(Model), c) // New会把c藏到m里面

		if _, err := act.PreGet(m); err != nil {
			c.Warn("PreGet error: %s", err)
			c.RESTBadRequest(err)
			return
		}

		var r interface{}
		var err error
		if r, err = act.OnGet(m); err != nil {
			c.Warn("OnGet error: %s", err)
			if err == ErrNoRecord {
				c.RESTNotFound(err)
			} else {
				c.RESTPanic(err)
			}
			return
		}

		if r, err = act.PostGet(r); err != nil {
			c.Warn("PostGet error: %s", err)
			c.RESTNotOK(err)
		} else {
			c.RESTOK(r)
		}

		return
	}
	search := func(c *RESTContext) {
		m := i.(Model).New(i.(Model), c) // New会把c藏到m里面

		if _, err := act.PreSearch(m); err != nil { // presearch准备条件等
			c.Warn("PreSearch error: %s", err)
			c.RESTBadRequest(err)
			return
		}

		if l, err := act.OnSearch(m); err != nil {
			c.Warn("OnSearch error: %s", err)
			if err == ErrNoRecord {
				c.RESTNotFound(err)
			} else {
				c.RESTPanic(err)
			}
		} else {
			if rl, err := act.PostSearch(l); err != nil {
				c.Warn("PostSearch error: %s", err)
				c.RESTNotOK(err)
			} else {
				c.RESTOK(rl)
			}
		}

		return

	}

	post := func(c *RESTContext) {
		m := i.(Model).New(i.(Model), c) // New会把c藏到m里面
		var err error

		if _, err = act.PreCreate(m); err != nil { // presearch准备条件等
			c.Warn("PreCreate error: %s", err)
			c.RESTBadRequest(err)
			return
		}

		var r interface{}
		if r, err = act.OnCreate(m); err != nil {
			c.Warn("OnCreate error: %s", err)
			c.RESTNotOK(err)
			return
		}
		m = r.(Model)

		// 触发器
		r, err = act.Trigger(m)
		if err != nil {
			c.Warn("Trigger error: %s", err)
		}

		// create ok, return
		if r, err = act.PostCreate(r); err != nil {
			c.Warn("postCreate error: %s", err)
		}
		c.RESTOK(r)
		return
	}

	delete := func(c *RESTContext) {
		m := i.(Model).New(i.(Model), c) // New会把c藏到m里面
		var err error

		if _, err = act.PreDelete(m); err != nil { // presearch准备条件等
			c.Warn("PreUpdat error: %s", err)
			c.RESTBadRequest(err)
			return
		}

		if _, err = act.OnDelete(m); err != nil {
			c.Warn("OnUpdat error: %s", err)
			c.RESTNotOK(err)
			return
		}

		// update ok
		var r interface{}
		if r, err = act.PostDelete(m); err != nil {
			c.Warn("postCreate error: %s", err)
		}

		// 触发器
		_, err = act.Trigger(m)
		if err != nil {
			c.Warn("Trigger error: %s", err)
		}
		c.RESTOK(r)
		return
	}

	patch := func(c *RESTContext) { //修改
		m := i.(Model).New(i.(Model), c) // New会把c藏到m里面
		var err error

		if _, err = act.PreUpdate(m); err != nil { // presearch准备条件等
			c.Warn("PreUpdate error: %s", err)
			c.RESTBadRequest(err)
			return
		}

		if _, err = act.OnUpdate(m); err != nil {
			c.Warn("OnUpdat error: %s", err)
			c.RESTNotOK(err)
			return
		}

		// 触发器
		_, err = act.Trigger(m)
		if err != nil {
			c.Warn("Trigger error: %s", err)
		}

		// update ok
		var r interface{}
		if r, err = act.PostUpdate(m); err != nil {
			c.Warn("postCreate error: %s", err)
		}

		c.RESTOK(r)
		return
	}
	//put := func(c *RESTContext) { //重置
	//}
	head := func(c *RESTContext) { //检查字段
		m := i.(Model).New(i.(Model), c) // New会把c藏到m里面

		if _, err := act.PreCheck(m); err != nil { // presearch准备条件等
			c.Warn("PreCheck error: %s", err)
			c.RESTBadRequest(err)
			return
		}

		if cnt, err := act.OnCheck(m); err != nil {
			c.Warn("OnCheck error: %s", err)
			if err == ErrNoRecord {
				c.RESTNotFound(err)
			} else {
				c.RESTPanic(err)
			}
		} else {
			if cnt, _ := act.PostCheck(cnt); cnt.(int64) > 0 {
				c.Warn("PostCheck error: %s", err)
				c.RESTNotOK(nil)
			} else {
				c.RESTOK(nil)
			}
		}

		return
	}
	deny := func(c *RESTContext) {
		c.HTTPError(http.StatusMethodNotAllowed)
	}

	switch flag {
	case GA_GET:
		return get
	case GA_SEARCH:
		return search
	case GA_POST:
		return post
	case GA_DELETE:
		return delete
	case GA_PATCH:
		return patch
	//case GA_PUT:
	//	return put
	case GA_HEAD:
		return head
	default:
		return deny
	}
}

/* }}} */
