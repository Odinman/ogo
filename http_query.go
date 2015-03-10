/* 解析http请求的各种条件 */
package ogo

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	_DEF_PAGE     = 1 //1-base
	_DEF_PER_PAGE = 25
	_MAX_PER_PAGE = 100 //每页最大个数

	//time
	_DATE_FORM  = "2006-01-02"
	_DATE_FORM1 = "20060102"
)

//时间段
type TimeRange struct {
	Start time.Time
	End   time.Time
}

//order by
type OrderBy struct {
	Field string
	Sort  string
}

// 分页信息
type Pagination struct {
	Page    int
	PerPage int
	Offset  int
}

// 条件信息
type Condition struct {
	Field string
	Is    interface{}
	Not   interface{}
	Like  interface{}
	Join  interface{}
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

/* {{{ func (rc *RESTContext) setTimeRangeFromDate(p []string) {
 * 时间段信息
 */
func (rc *RESTContext) setTimeRangeFromDate(p []string) {
	tr := new(TimeRange)
	rc.SetEnv(TimeRangeKey, tr)

	var s, e, format string
	if len(p) > 1 { //有多个,第一个是start, 第二个是end, 其余忽略
		s, e = p[0], p[1]
	} else if len(p) > 0 { //只有一个, 可通过 "{start},{end}"方式传
		pieces := strings.SplitN(p[0], ",", 2)
		s = pieces[0]
		if len(pieces) > 1 {
			e = pieces[1]
		}
	}
	if len(s) == len(_DATE_FORM) {
		format = _DATE_FORM
	} else if len(s) == len(_DATE_FORM1) {
		format = _DATE_FORM1
	}
	if ts, err := time.ParseInLocation(format, s, Env().Location); err == nil {
		tr.Start = ts
		dura, _ := time.ParseDuration("86399s") // 一天少一秒
		tr.End = ts.Add(dura)                   //当天的最后一秒
		//只有成功获取了start, end才有意义
		if t, err := time.ParseInLocation(format, e, Env().Location); err == nil {
			te := t.Add(dura)
			if te.After(ts) { //必须比开始大
				tr.End = te
			}
		}
	}

	return
}

/* }}} */

/* {{{ func (rc *RESTContext) setTimeRangeFromStartEnd() {
 * 时间段信息
 */
func (rc *RESTContext) setTimeRangeFromStartEnd() {
	var sp, ep []string
	var ok bool
	r := rc.Request
	if sp, ok = r.Form[_PARAM_START]; !ok {
		//没有传入start,do nothing
		return
	}
	//删除
	r.Form.Del(_PARAM_START)

	if ep, ok = r.Form[_PARAM_END]; !ok {
		//没有传入end,do nothing
		return
	}
	//删除
	r.Form.Del(_PARAM_END)

	s, e := sp[0], ep[0]

	if len(s) != len(e) {
		//长度不一致,返回
		return
	}

	var format string
	if len(s) == len(_DATE_FORM) {
		format = _DATE_FORM
	} else if len(s) == len(_DATE_FORM1) {
		format = _DATE_FORM1
	}
	tr := new(TimeRange)
	if ts, err := time.ParseInLocation(format, s, Env().Location); err != nil {
		return
	} else {
		tr.Start = ts
	}
	if te, err := time.ParseInLocation(format, e, Env().Location); err != nil {
		return
	} else {
		tr.End = te
	}

	rc.SetEnv(TimeRangeKey, tr)

	return
}

/* }}} */

/* {{{ func (rc *RESTContext) setOrderBy(p string) {
 * 时间段信息
 */
func (rc *RESTContext) setOrderBy(p []string) {
	ob := new(OrderBy)
	rc.SetEnv(OrderByKey, ob)
	if len(p) > 0 { //只有一个, 可通过 "{start},{end}"方式传
		pieces := strings.SplitN(p[0], ",", 2)
		ob.Field = pieces[0]
		ob.Sort = "DESC" //默认降序
		if len(pieces) > 1 {
			if strings.ToUpper(pieces[1]) == "ASC" {
				ob.Sort = "ASC"
			}
		}
	}

	return
}

/* }}} */
