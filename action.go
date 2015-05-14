// Ogo

package ogo

import ()

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

	PASSWORD_SALT = "ogo"

	KEY_SKIPAUTH  = "skipauth"
	KEY_SKIPLOGIN = "skiplogin"
	KEY_SKIPPERM  = "skipperm"
)

type Action interface {
	CRUD(i interface{}, flag int) Handler

	Valid(i interface{}) (interface{}, error) //数据验证

	PreGet(i interface{}) (interface{}, error)  //获取前调整
	PostGet(i interface{}) (interface{}, error) //获取后调整

	PreSearch(i interface{}) (interface{}, error)  // 搜索前的检查
	PostSearch(i interface{}) (interface{}, error) // 搜索后的检查

	PreCreate(i interface{}) (interface{}, error)  // 插入前的检查
	PostCreate(i interface{}) (interface{}, error) // 插入后的处理

	PreUpdate(i interface{}) (interface{}, error)  // 更新前的检查
	PostUpdate(i interface{}) (interface{}, error) // 更新前的检查

	PreDelete(i interface{}) (interface{}, error)  // 删除前的检查
	PostDelete(i interface{}) (interface{}, error) // 删除前的检查

	OnAction(i interface{}) (interface{}, error) //触发器
}

//用router实现这个interface
/* {{{ func (bm *BaseModel) Valid(m Model) (nm Model, err error)
 *
 */
func (bm *BaseModel) Valid(i interface{}) (interface{}, error) {
	return i, nil
}

/* }}} */
/* {{{ func (bm *BaseModel) PreGet(m Model, c *RESTContext) (nm Model, err error)
 *
 */
func (bm *BaseModel) PreGet(i interface{}) (interface{}, error) {
	return i, nil
}

/* }}} */
/* {{{ func (bm *BaseModel) PostGet(i interface{}) (nm Model, err error)
 *
 */
func (bm *BaseModel) PostGet(i interface{}) (interface{}, error) {
	return i, nil
}

/* }}} */
/* {{{ func (_ *BaseModel) PreSearch(m Model) (nm Model, err error)
 *
 */
func (_ *BaseModel) PreSearch(i interface{}) (interface{}, error) {
	c := i.(Model).GetCtx()
	c.Debug("herehere")
	if cons := c.GetEnv(ConditionsKey); cons != nil { //从context里面获取参数条件
		i.(Model).SetConditions(i.(Model), cons.(Conditions))
	}
	return i, nil
}

/* }}} */
/* {{{ func (bm *BaseModel) PostSearch(i interface{}) (interface{}, error)
 *
 */
func (bm *BaseModel) PostSearch(i interface{}) (interface{}, error) {
	return i, nil
}

/* }}} */
/* {{{ func (bm *BaseModel) PreCreate(m Model) (nm Model, err error)
 *
 */
func (bm *BaseModel) PreCreate(i interface{}) (interface{}, error) {
	return i, nil
}

/* }}} */
/* {{{ func (bm *BaseModel) PostCreate(m Model) (nm Model, err error)
 *
 */
func (bm *BaseModel) PostCreate(i interface{}) (interface{}, error) {
	return i, nil
}

/* }}} */
/* {{{ func (bm *BaseModel) PreUpdate(m Model) (nm Model, err error)
 *
 */
func (bm *BaseModel) PreUpdate(i interface{}) (interface{}, error) {
	return i, nil
}

/* }}} */
/* {{{ func (bm *BaseModel) PostUpdate(m Model) (nm Model, err error)
 *
 */
func (bm *BaseModel) PostUpdate(i interface{}) (interface{}, error) {
	return i, nil
}

/* }}} */
/* {{{ func (bm *BaseModel) PreDelete(m Model) (nm Model, err error)
 *
 */
func (bm *BaseModel) PreDelete(i interface{}) (interface{}, error) {
	return i, nil
}

/* }}} */
/* {{{ func (bm *BaseModel) PostDelete(m Model) (nm Model, err error)
 *
 */
func (bm *BaseModel) PostDelete(i interface{}) (interface{}, error) {
	return i, nil
}

/* }}} */
/* {{{ func (bm *BaseModel) OnAction(m Model) (err error)
 *
 */
func (bm *BaseModel) OnAction(i interface{}) (interface{}, error) {
	return i, nil
}

/* }}} */

/* {{{ func (bm *BaseModel) CRUD(m Model, flag int) Handler
 * 通用的操作方法, 根据flag返回
 * 必须符合通用的restful风格
 */
func (bm *BaseModel) CRUD(i interface{}, flag int) Handler {
	act := i.(Action)
	get := func(c *RESTContext) {
		m := i.(Model).New(i.(Model), c) // New会把c藏到m里面
		if _, err := act.PreGet(m); err != nil {
			c.RESTBadRequest(err)
		}
		id := c.URLParams["_id_"]
		if obj, err := m.GetRow(m, id); err == nil {
			r, _ := act.PostGet(obj)
			c.RESTOK(r)
		} else if err == ErrNoRecord {
			c.RESTNotFound(err)
		} else {
			c.RESTPanic(err)
		}

		return
	}
	search := func(c *RESTContext) {
		m := i.(Model).New(i.(Model), c) // New会把c藏到m里面
		if _, err := act.PreSearch(m); err != nil {
			c.RESTBadRequest(err)
		}
		if l, err := m.GetRows(m); err == nil {
			rl, _ := act.PostSearch(l)
			c.RESTOK(rl)
		} else if err == ErrNoRecord {
			c.RESTNotFound(err)
		} else {
			c.RESTPanic(err)
		}

		return
	}
	post := func(c *RESTContext) {
	}
	delete := func(c *RESTContext) {
	}
	patch := func(c *RESTContext) { //修改
	}
	//put := func(c *RESTContext) { //重置
	//}
	head := func(c *RESTContext) { //检查字段
	}
	deny := func(c *RESTContext) {
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
