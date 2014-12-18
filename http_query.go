/* 解析http请求的各种条件 */
package ogo

import (
	"fmt"
	"strconv"
)

const (
	_DEF_PAGE     = 1 //1-base
	_DEF_PER_PAGE = 25
	_MAX_PER_PAGE = 100 //每页最大个数
)

// 分页信息
type Pagination struct {
	Page    int
	PerPage int
	Offset  int
}

// 条件信息
type Condition struct {
	Is   interface{}
	Not  interface{}
	Like interface{}
}

type Conditions map[string]*Condition

/* {{{ func NewPagation(page, perPage string) (p *Pagination)
 */
func NewPagination(page, perPage string) (p *Pagination) {
	var pageNum, offset, perNum int
	if page == "" {
		pageNum = _DEF_PAGE
	} else {
		pageNum, _ = strconv.Atoi(page)
		if pageNum < 1 {
			pageNum = _DEF_PAGE
		}
	}
	if perPage == "" {
		perNum = _DEF_PER_PAGE
	} else {
		perNum, _ = strconv.Atoi(perPage)
		if perNum > _MAX_PER_PAGE {
			perNum = _MAX_PER_PAGE
		}
	}
	offset = (pageNum - 1) * perNum
	p = &Pagination{
		Page:    pageNum,
		PerPage: perNum,
		Offset:  offset,
	}
	return
}

/* }}} */

/* {{{ func (rc *RESTContext) GetCondition(k string) (con *Condition, err error)
 * 设置参数查询条件
 */
func (rc *RESTContext) GetCondition(k string) (con *Condition, err error) {
	if cs, ok := rc.Env[ConditionsKey]; !ok {
		//没有conditions,自动初始化
		rc.SetEnv(ConditionsKey, make(Conditions))
		return nil, fmt.Errorf("Not found conditions! %s", ConditionsKey)
	} else {
		conditions := cs.(Conditions)
		if _, ok := conditions[k]; !ok {
			return nil, fmt.Errorf("Not found condition: %s", k)
		} else {
			con = conditions[k]
		}
	}
	return
}

/* }}} */

/* {{{ func (rc *RESTContext) setCondition(k string, con *Condition) (err error) {
	return
 *
*/
func (rc *RESTContext) setCondition(k string, con *Condition) {
	if k != "" {
		if _, ok := rc.Env[ConditionsKey]; !ok {
			//没有conditions,自动初始化
			rc.SetEnv(ConditionsKey, make(Conditions))
		}
		rc.Env[ConditionsKey].(Conditions)[k] = con
	}
}

/* }}} */
