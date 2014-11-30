/* 解析http请求的各种条件 */
package ogo

import (
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
