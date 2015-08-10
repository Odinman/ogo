// Ogo

package ogo

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/Odinman/gorp"
	"github.com/Odinman/ogo/utils"
)

const (
	// db tag
	DBTAG_PK    = "pk"
	DBTAG_LOGIC = "logic"

	//tag
	TAG_REQUIRED   = "R"     // 必填
	TAG_GENERATE   = "G"     // 服务端生成, 同时不可编辑
	TAG_CONDITION  = "C"     // 可作为查询条件
	TAG_DENY       = "D"     // 不可编辑, 可为空
	TAG_SECRET     = "S"     //保密,一般不见人
	TAG_TIMERANGE  = "TR"    //时间范围条件
	TAG_REPORT     = "RPT"   //报表字段
	TAG_CANGROUP   = "GRP"   //可以group操作
	TAG_ORDERBY    = "O"     //可排序
	TAG_AORDERBY   = "AO"    //正排序(默认DESC)
	TAG_VERIFIABLE = "V"     //验证后可修改
	TAG_RETURN     = "RET"   // 返回,创建后需要返回数值
	TAG_SUM        = "SUM"   // 求和
	TAG_COUNT      = "COUNT" // 计数
)

type List struct {
	Info  ListInfo    `json:"info,omitempty"`
	Total int64       `json:"total"`
	List  interface{} `json:"list"`
}

type ListInfo struct {
	Page    *int        `json:"page,omitempty"`     //当前页面
	PerPage *int        `json:"per_page,omitempty"` //每页元素个数
	Sum     interface{} `json:"sum,omitempty"`      //求和
}

//错误代码
var (
	ErrRequired      = errors.New("field is required")
	ErrNonEditable   = errors.New("field is non-editable")
	ErrNonSearchable = errors.New("field is non-searchable")
	ErrExists        = errors.New("field value exists")
	ErrInvalid       = errors.New("invalid_query")
	ErrNoRecord      = errors.New("No_record")
	ErrNeedField     = errors.New("Need field but missing")
)

type Condition struct {
	Table string
	Field string
	Is    interface{}
	Not   interface{}
	Or    interface{}
	Gt    interface{}
	Lt    interface{}
	Like  interface{}
	Join  interface{}
	Range interface{} //范围条件, btween ? and ?
	Order interface{}
	Page  interface{}
	Raw   string //原始字符串
}

//order by
type OrderBy struct {
	Field string
	Sort  string
}

func NewCondition(typ int, field string, cs ...interface{}) *Condition {
	if field == "" || len(cs) < 1 { //至少1个元素
		return nil
	}
	con := &Condition{Field: field}
	var v interface{}
	if len(cs) == 1 {
		v = cs[0]
	} else {
		v = cs
	}
	switch typ {
	case CTYPE_IS:
		con.Is = v
	case CTYPE_NOT:
		con.Not = v
	case CTYPE_GT:
		con.Gt = v
	case CTYPE_LT:
		con.Lt = v
	case CTYPE_JOIN:
		con.Join = v
	case CTYPE_OR:
		con.Or = v
	case CTYPE_LIKE:
		con.Like = v
	case CTYPE_RANGE:
		con.Range = v
	case CTYPE_ORDER:
		con.Order = v
	case CTYPE_PAGE:
		con.Page = v
	case CTYPE_RAW:
		con.Raw = v.(string)
	default:
	}
	return con
}

/* {{{ func (v *Condition) DoWhere(b *gorp.Builder)
 * 只负责生成部分sql, IS/NOT/LIKE/GT/LT
 */
func (v *Condition) DoWhere(b *gorp.Builder) {
	if v.Raw != "" {
		b.Where(fmt.Sprint("(", v.Raw, ")"))
	}
	if v.Is != nil {
		//Debug("[=][key: %s]%v", v.Field, v)
		switch vt := v.Is.(type) {
		case string:
			b.Where(fmt.Sprintf("T.`%s` = ?", v.Field), vt)
		case []string:
			//Debug("[=][slices][key: %s]%v", v.Field, v)
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
		case []interface{}:
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
		case []interface{}:
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
	if v.Gt != nil {
		//Debug("[>=][key: %s]%v", v.Field, v)
		switch vt := v.Gt.(type) {
		case string:
			b.Where(fmt.Sprintf("T.`%s` >= '%s'", v.Field, vt))
		case []string:
			vs := bytes.Buffer{}
			first := true
			vs.WriteString("(")
			for _, vv := range vt {
				if !first {
					vs.WriteString(" OR ")
				}
				vs.WriteString(fmt.Sprintf("T.`%s` >= '%s'", v.Field, vv))
				first = false
			}
			vs.WriteString(")")
			b.Where(vs.String())
		case []interface{}:
			vs := bytes.Buffer{}
			first := true
			vs.WriteString("(")
			for _, vv := range vt {
				if !first {
					vs.WriteString(" OR ")
				}
				vs.WriteString(fmt.Sprintf("T.`%s` >= '%s'", v.Field, vv))
				first = false
			}
			vs.WriteString(")")
			b.Where(vs.String())
		default:
		}
	}
	if v.Lt != nil {
		//Debug("[<][key: %s]%v", v.Field, v)
		switch vt := v.Lt.(type) {
		case string:
			b.Where(fmt.Sprintf("T.`%s` < '%s'", v.Field, vt))
		case []string:
			vs := bytes.Buffer{}
			first := true
			vs.WriteString("(")
			for _, vv := range vt {
				if !first {
					vs.WriteString(" OR ")
				}
				vs.WriteString(fmt.Sprintf("T.`%s` < '%s'", v.Field, vv))
				first = false
			}
			vs.WriteString(")")
			b.Where(vs.String())
		case []interface{}:
			vs := bytes.Buffer{}
			first := true
			vs.WriteString("(")
			for _, vv := range vt {
				if !first {
					vs.WriteString(" OR ")
				}
				vs.WriteString(fmt.Sprintf("T.`%s` < '%s'", v.Field, vv))
				first = false
			}
			vs.WriteString(")")
			b.Where(vs.String())
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
		case []interface{}:
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

/* }}} */

/* {{{ func (con *Condition) Merge(oc *Condition)
 * 直接覆盖
 */
func (con *Condition) Merge(oc *Condition) {
	if oc == nil {
		return
	}
	if oc.Is != nil {
		con.Is = oc.Is
	}
	if oc.Or != nil {
		con.Or = oc.Or
	}
	if oc.Not != nil {
		con.Not = oc.Not
	}
	if oc.Gt != nil {
		con.Gt = oc.Gt
	}
	if oc.Lt != nil {
		con.Lt = oc.Lt
	}
	if oc.Like != nil {
		con.Like = oc.Like
	}
	if oc.Range != nil {
		con.Range = oc.Range
	}
	if oc.Order != nil {
		con.Order = oc.Order
	}
	if oc.Join != nil {
		con.Join = oc.Join
	}
	if oc.Page != nil {
		con.Page = oc.Page
	}
	if oc.Raw != "" {
		con.Raw = oc.Raw
	}
}

/* }}} */

type Model interface {
	SetModel(m Model) Model
	GetModel() Model
	SetCtx(c *RESTContext)
	GetCtx() *RESTContext
	SetConditions(cs ...*Condition) ([]*Condition, error)
	GetConditions() []*Condition
	SetPagination(p *Pagination)
	GetPagination() *Pagination
	SetFields(fs []string)
	GetFields() []string
	GetChecker() Checker

	New(c ...interface{}) Model
	NewList() interface{} // 返回一个空结构列表

	// db
	AddTable(tags ...string)
	DBConn(tag string) *gorp.DbMap // 数据库连接
	TableName() string             // 返回表名称, 默认结构type名字(小写), 有特别的表名称,则自己implement 这个方法
	PKey() string                  // key字段
	ReadPrepare() (*gorp.Builder, error)

	// data accessor
	GetRow(ext ...interface{}) (Model, error) //获取单条记录
	GetRows() (*List, error)                  //获取多条记录
	GetOlder() Model                          //获取旧记录
	GetSum(d []string) (*List, error)         //获取多条记录
	GetCount() (int64, error)                 //获取多条记录
	CreateRow() (Model, error)                //创建单条记录
	UpdateRow(id string) (int64, error)       //更新记录
	DeleteRow(id string) (int64, error)       //更新记录
	CheckerFactory() Checker                  //检查存在性
	Valid() (Model, error)                    //数据验证
	Filter() (Model, error)                   //数据过滤(创建,更新后)
	Protect() (Model, error)                  //数据保护(获取数据时过滤字段)
}

type Checker func(string) (interface{}, error)

//基础model,在这里可以实现Model接口, 其余的只需要嵌入这个struct,就可以继承这些方法
type BaseModel struct {
	Error      error        `json:"-" db:"-"`
	Locked     []string     `json:"-" db:"-"`
	Model      Model        `json:"-" db:"-"`
	ctx        *RESTContext `json:"-" db:"-"`
	checker    Checker      `json:"-" db:"-"`
	conditions []*Condition `json:"-" db:"-"`
	pagination *Pagination  `json:"-" db:"-"`
	fields     []string     `json:"-" db:"-"`
	base       string       `json:"-" db:"-"` //这个的作用就是判断是否是BaseModel
	older      Model        `json:"-" db:"-"`
}

/* {{{ func NewModel(m Model,c ...interface{}) Model {
 * 第一个参数,model,必须是指针; 第二个参数, *RESTContext
 */
func NewModel(m Model, c ...interface{}) Model {
	//新建一个指针
	nmi := reflect.New(reflect.Indirect(reflect.ValueOf(m)).Type()).Interface().(Model)
	nmi.SetModel(nmi)
	if len(c) > 0 {
		nmi.SetCtx(c[0].(*RESTContext))
	}
	return nmi
}

/* }}} */

/* {{{ func BuildModel(m Model,c ...interface{}) Model {
 * 第一个参数,model,必须是指针; 第二个参数, *RESTContext
 */
func BuildModel(m Model, c ...interface{}) Model {
	//新建一个指针
	m.SetModel(m)
	if len(c) > 0 {
		m.SetCtx(c[0].(*RESTContext))
	}
	return m
}

/* }}} */

/* {{{ func GetCondition(cs []*Condition, k string) (con *Condition, err error)
 *
 */
func GetCondition(cs []*Condition, k string) (con *Condition, err error) {
	if cs == nil {
		err = fmt.Errorf("conditions empty")
	} else {
		for _, c := range cs {
			//Debug("field: %s, key: %s", c.Field, k)
			if c.Field == k {
				return c, nil
			}
		}
	}
	return nil, fmt.Errorf("cannot found condition: %s", k)
}

/* }}} */

/* {{{ func (bm *BaseModel) SetModel(m Model) Model
 *
 */
func (bm *BaseModel) SetModel(m Model) Model {
	bm.Model = m
	return bm
}

/* }}} */

/* {{{ func (bm *BaseModel) GetModel() Model
 *
 */
func (bm *BaseModel) GetModel() Model {
	return bm.Model
}

/* }}} */

/* {{{ func (bm *BaseModel) SetCtx(c *RESTContext)
 *
 */
func (bm *BaseModel) SetCtx(c *RESTContext) {
	bm.ctx = c
}

/* }}} */

/* {{{ func (bm *BaseModel) GetCtx() *RESTContext
 *
 */
func (bm *BaseModel) GetCtx() *RESTContext {
	return bm.ctx
}

/* }}} */

/* {{{ func (bm *BaseModel) SetConditions(cs ...*Condition) (cons []*Condition, err error)
 * 生成条件
 */
func (bm *BaseModel) SetConditions(cs ...*Condition) (cons []*Condition, err error) {
	var m Model
	if m = bm.GetModel(); m == nil {
		err := fmt.Errorf("not found model")
		Info("error: %s", err)
		return nil, err
	}
	if bm.conditions == nil {
		bm.conditions = make([]*Condition, 0)
	}
	if cols := utils.ReadStructColumns(m, true); cols != nil {
		for _, col := range cols {
			// time range
			if col.ExtOptions.Contains(TAG_TIMERANGE) {
				if condition, e := GetCondition(cs, TAG_TIMERANGE); e == nil && condition.Range != nil {
					//Debug("[SetConditions]timerange")
					condition.Field = col.Tag
					bm.conditions = append(bm.conditions, condition)
				} else {
					Trace("get condition failed: %s", e)
				}
			}
			if col.ExtOptions.Contains(TAG_ORDERBY) {
				if condition, e := GetCondition(cs, TAG_ORDERBY); e == nil && condition.Order != nil {
					//Debug("[SetConditions]order")
					condition.Field = col.Tag
					bm.conditions = append(bm.conditions, condition)
				} else {
					Trace("get condition failed: %s", e)
				}
			}
			if col.TagOptions.Contains(DBTAG_PK) || col.ExtOptions.Contains(TAG_CONDITION) { //primary key or conditional
				if condition, e := GetCondition(cs, col.Tag); e == nil && (condition.Is != nil || condition.Not != nil || condition.Gt != nil || condition.Lt != nil || condition.Like != nil || condition.Join != nil || condition.Raw != "" || condition.Or != nil) {
					//Debug("[SetConditions][tag: %s]%v", col.Tag, condition)
					bm.conditions = append(bm.conditions, condition)
				}
			}
		}
	}
	return bm.conditions, nil
}

/* }}} */

/* {{{ func (bm *BaseModel) GetConditions() []*Condition
 *
 */
func (bm *BaseModel) GetConditions() []*Condition {
	return bm.conditions
}

/* }}} */

/* {{{ func (bm *BaseModel) SetPagination(p *Pagination)
 * 生成条件
 */
func (bm *BaseModel) SetPagination(p *Pagination) {
	if bm.conditions == nil {
		bm.pagination = new(Pagination)
	}
	bm.pagination = p
}

/* }}} */

/* {{{ func (bm *BaseModel) GetPagination() *Pagination
 *
 */
func (bm *BaseModel) GetPagination() *Pagination {
	return bm.pagination
}

/* }}} */

/* {{{ func (bm *BaseModel) SetFields(fs []string)
 * 生成条件
 */
func (bm *BaseModel) SetFields(fs []string) {
	if bm.fields == nil {
		bm.fields = make([]string, 0)
	}
	bm.fields = fs
}

/* }}} */

/* {{{ func (bm *BaseModel) GetFields() []string
 *
 */
func (bm *BaseModel) GetFields() []string {
	return bm.fields
}

/* }}} */

/* {{{ func (bm *BaseModel) GetChecker() Checker
 *
 */
func (bm *BaseModel) GetChecker() Checker {
	if bm.checker == nil {
		if m := bm.GetModel(); m != nil {
			bm.checker = m.CheckerFactory()
		}
	}
	return bm.checker
}

/* }}} */

/* {{{ func (bm *BaseModel) New(c ...interface{}) Model
 * 初始化model, 后面的c选填
 */
func (bm *BaseModel) New(c ...interface{}) Model {
	if m := bm.GetModel(); m != nil {
		return NewModel(m, c...)
	} else {
		Info("New:not found model")
		return nil
	}
}

/* }}} */

/* {{{ func (bm *BaseModel) NewList() *[]Model
 *
 */
func (bm *BaseModel) NewList() interface{} {
	var m Model
	if m = bm.GetModel(); m == nil {
		err := fmt.Errorf("not found model")
		Info("error: %s", err)
		return nil
	}
	ms := reflect.New(reflect.SliceOf(reflect.TypeOf(m))).Interface()
	return ms
}

/* }}} */

/* {{{ func (bm *BaseModel) DBConn(tag string) *gorp.DbMap
 * 默认数据库连接为admin
 */
func (bm *BaseModel) DBConn(tag string) *gorp.DbMap {
	tb := bm.TableName()
	if dt, ok := DataAccessor[tb+"::"+tag]; ok && dt != "" {
		return gorp.Using(dt)
	}
	return gorp.Using(DBTAG)
}

/* }}} */

/* {{{ func (bm *BaseModel) TableName() (n string)
 * 获取表名称, 默认为结构名
 */
func (bm *BaseModel) TableName() (n string) {
	//默认, struct的名字就是表名, 如果不是请在各自的model里定义
	var m Model
	if m = bm.GetModel(); m == nil {
		err := fmt.Errorf("not found model")
		Info("error: %s", err)
		return ""
	}
	reflectVal := reflect.ValueOf(m)
	mt := reflect.Indirect(reflectVal).Type()
	n = underscore(strings.TrimSuffix(mt.Name(), "Table"))
	return
}

/* }}} */

/* {{{ func (bm *BaseModel) PKey() string
 *  通过配置找到pk
 */
func (bm *BaseModel) PKey() string {
	var m Model
	if m = bm.GetModel(); m == nil {
		err := fmt.Errorf("not found model")
		Info("error: %s", err)
		return ""
	}
	if cols := utils.ReadStructColumns(m, true); cols != nil {
		for _, col := range cols {
			// check required field
			if col.TagOptions.Contains(DBTAG_PK) {
				return col.Tag
			}
		}
	}
	return ""
}

/* }}} */

/* {{{ func (bm *BaseModel) CheckerFactory() Checker
 *
 */
func (bm *BaseModel) CheckerFactory() Checker {
	return func(tag string) (interface{}, error) {
		return nil, nil
	}
}

/* }}} */

/* {{{ func (bm *BaseModel) Filter() (Model, error)
 * 数据过滤
 */
func (bm *BaseModel) Filter() (Model, error) {
	if m := bm.GetModel(); m != nil {
		//r := m.New()
		r := NewModel(m)
		rv := reflect.ValueOf(r)
		v := reflect.ValueOf(m)
		if cols := utils.ReadStructColumns(m, true); cols != nil {
			for _, col := range cols {
				fv := utils.FieldByIndex(v, col.Index)
				mv := utils.FieldByIndex(rv, col.Index)
				//c.Trace("field:%s; name: %s, kind:%v; type:%s", col.Tag, col.Name, fv.Kind(), fv.Type().String())
				if col.TagOptions.Contains(DBTAG_PK) || col.ExtOptions.Contains(TAG_RETURN) {
					//pk以及定义了返回tag的赋值
					mv.Set(fv)
				}
			}
		}
		return r.(Model), nil
	} else {
		err := fmt.Errorf("not found model")
		Info("error: %s", err)
		return nil, err
	}
}

/* }}} */

/* {{{ func (bm *BaseModel) Valid() (Model, error)
 * 根据条件获取一条记录, model为表结构
 */
func (bm *BaseModel) Valid() (Model, error) {
	var m Model
	if m = bm.GetModel(); m == nil {
		err := fmt.Errorf("not found model")
		Info("error: %s", err)
		return nil, err
	}
	c := m.GetCtx()
	if err := json.Unmarshal(c.RequestBody, m); err != nil {
		return nil, err
	}
	// checker
	checker := m.GetChecker()
	v := reflect.ValueOf(m)
	if cols := utils.ReadStructColumns(m, true); cols != nil {
		for _, col := range cols {
			fv := utils.FieldByIndex(v, col.Index)
			// server generate,忽略传入的信息
			if fv.IsValid() && !utils.IsEmptyValue(fv) { //传入了内容
				if col.ExtOptions.Contains(TAG_GENERATE) { //服务器生成, 忽略传入
					fv.Set(reflect.Zero(fv.Type()))
				} else if c.Route.Updating && col.ExtOptions.Contains(TAG_DENY) { //尝试编辑不可编辑的字段,要报错
					c.Info("%s is uneditable: %v", col.Tag, fv)
					return nil, fmt.Errorf("%s is uneditable", col.Tag) //尝试编辑不可编辑的字段,直接报错
				}
			} else { //空
				if col.ExtOptions.Contains(TAG_REQUIRED) && c.Route.Creating { // 创建时必须传入,但是为空
					c.Debug("field %s required but empty", col.Tag)
					return nil, fmt.Errorf("field %s required but empty", col.Tag)
				}
			}
			switch col.ExtTag { //根据tag, 会对数据进行预处理
			case "sha1":
				if fv.IsValid() && !utils.IsEmptyValue(fv) { //不能为空
					switch fv.Type().String() {
					case "*string":
						sv := fv.Elem().String()
						h := utils.HashSha1(sv)
						fv.Set(reflect.ValueOf(&h))
						//c.Debug("password: %s, encoded: %s", sv, h)
					case "string":
						sv := fv.String()
						h := utils.HashSha1(sv)
						fv.Set(reflect.ValueOf(h))
						//c.Debug("password: %s, encoded: %s", sv, h)
					default:
						return nil, fmt.Errorf("field(%s) must be string, not %s", col.Tag, fv.Kind().String())
					}
				}
			case "userid": //替换为userid
				if c.Route.Creating {
					var userid string
					if uid := c.GetEnv(USERID_KEY); uid == nil {
						userid = "0"
						//c.Debug("userid not exists")
					} else {
						userid = uid.(string)
						//c.Debug("userid: %s", userid)
					}
					switch fv.Type().String() {
					case "*string":
						fv.Set(reflect.ValueOf(&userid))
					case "string":
						fv.Set(reflect.ValueOf(userid))
					default:
						return nil, fmt.Errorf("field(%s) must be string, not %s", col.Tag, fv.Kind().String())
					}
				}
			case "existense": //检查存在性
				if c.Route.Creating { //创建时才检查,这里不够安全(将来改)
					if exValue, err := checker(col.Tag); err != nil {
						return nil, fmt.Errorf("%s existense check failed: %s", col.Tag, err.Error())
					} else if exValue != nil {
						//c.Debug("%s existense: %v", col.Tag, exValue)
						fv.Set(reflect.ValueOf(exValue))
					}
				}
			case "uuid":
				if c.Route.Creating {
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
				}
			case "luuid":
				if c.Route.Creating {
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
				}
			case "forbbiden": //这个字段如果旧记录有值, 则返回错误
				if c.Route.Updating {
					older := m.GetOlder()
					ov := reflect.ValueOf(older)
					fov := utils.FieldByIndex(ov, col.Index)
					if fov.IsValid() && !utils.IsEmptyValue(fov) {
						return nil, fmt.Errorf("field(%s) has value, can't be updated", col.Tag)
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
	return m, nil
}

/* }}} */

/* {{{ func (bm *BaseModel) Protect() (Model, error)
 * 数据过滤
 */
func (bm *BaseModel) Protect() (Model, error) {
	if m := bm.GetModel(); m != nil {
		if cols := utils.ReadStructColumns(m, true); cols != nil {
			v := reflect.ValueOf(m)
			for _, col := range cols {
				if col.ExtOptions.Contains(TAG_SECRET) { //保密,不对外
					fv := utils.FieldByIndex(v, col.Index)
					fv.Set(reflect.Zero(fv.Type()))
				}
			}
		}
		return m, nil
	} else {
		err := fmt.Errorf("not found model")
		Info("error: %s", err)
		return nil, err
	}
}

/* }}} */

/* {{{ func (bm *BaseModel) GetRow(ext ...interface{}) (Model, error)
 * 根据条件获取一条记录, model为表结构
 */
func (bm *BaseModel) GetRow(ext ...interface{}) (Model, error) {
	var m Model
	if m = bm.GetModel(); m == nil {
		err := fmt.Errorf("not found model")
		Info("error: %s", err)
		return nil, err
	}
	c := m.GetCtx()
	if len(ext) > 0 {
		if id, ok := ext[0].(string); ok {
			m.SetConditions(NewCondition(CTYPE_IS, m.PKey(), id))
		}
	}
	builder, _ := m.ReadPrepare()
	ms := m.NewList()
	var err error
	err = builder.Select(GetDbFields(m)).Limit("1").Find(ms)
	if err != nil && err != sql.ErrNoRows {
		//支持出错
		return nil, err
	} else if ms == nil {
		//没找到记录
		return nil, ErrNoRecord
	}

	resultsValue := reflect.Indirect(reflect.ValueOf(ms))
	if resultsValue.Len() <= 0 {
		//c.Debug("len: %d, no record", resultsValue.Len())
		return nil, ErrNoRecord
	}
	return BuildModel(resultsValue.Index(0).Interface().(Model), c), nil
}

/* }}} */

/* {{{ func (bm *BaseModel) CreateRow() (Model, error)
 * 根据条件获取一条记录, model为表结构
 */
func (bm *BaseModel) CreateRow() (Model, error) {
	if m := bm.GetModel(); m != nil {
		db := bm.DBConn(WRITETAG)
		if err := db.Insert(m); err != nil { //Insert会把m换成新的
			return nil, err
		} else {
			return m.SetModel(m), nil
		}
	} else {
		err := fmt.Errorf("not found model")
		Info("error: %s", err)
		return nil, err
	}
}

/* }}} */

/* {{{ func (bm *BaseModel) UpdateRow(id string) (affected int64, err error)
 * 根据条件获取一条记录, model为表结构
 */
func (bm *BaseModel) UpdateRow(id string) (affected int64, err error) {
	if m := bm.GetModel(); m != nil {
		db := bm.DBConn(WRITETAG)
		if id != "" {
			if err = utils.ImportValue(m, map[string]string{DBTAG_PK: id}); err != nil {
				return
			}
		}
		return db.Update(m)
	} else {
		err := fmt.Errorf("not found model")
		Info("error: %s", err)
		return 0, err
	}
}

/* }}} */

/* {{{ func (bm *BaseModel) DeleteRow(id string) (affected int64, err error)
 * 删除记录(逻辑删除)
 */
func (bm *BaseModel) DeleteRow(id string) (affected int64, err error) {
	if m := bm.GetModel(); m != nil {
		db := bm.DBConn(WRITETAG)
		if err = utils.ImportValue(m, map[string]string{DBTAG_PK: id, DBTAG_LOGIC: "-1"}); err != nil {
			return
		}
		return db.Update(m)
	} else {
		err := fmt.Errorf("not found model")
		Info("error: %s", err)
		return 0, err
	}
}

/* }}} */

/* {{{ func (bm *BaseModel) GetRows() (l *List, err error)
 * 获取list, 通用函数
 */
func (bm *BaseModel) GetRows() (l *List, err error) {
	//c := m.GetCtx()
	if m := bm.GetModel(); m != nil {
		l = new(List)
		builder, _ := bm.ReadPrepare()
		count, _ := builder.Count() //结果数
		ms := bm.NewList()
		//p := c.GetEnv(PaginationKey).(*Pagination)
		if p := bm.GetPagination(); p != nil {
			l.Info.Page = &p.Page
			l.Info.PerPage = &p.PerPage
			err = builder.Select(GetDbFields(m, true)).Offset(p.Offset).Limit(p.PerPage).Find(ms)
		} else {
			err = builder.Select(GetDbFields(m, true)).Find(ms)
		}
		if err != nil && err != sql.ErrNoRows {
			//支持出错
			return l, err
		} else if ms == nil {
			//没找到记录
			return l, ErrNoRecord
		}

		l.Total = count
		l.List = ms

		return l, nil
	} else {
		err := fmt.Errorf("not found model")
		Info("error: %s", err)
		return nil, err
	}
}

/* }}} */

/* {{{ func (bm *BaseModel) GetSum(d []string) (l *List, err error)
 * 获取list, 通用函数
 */
func (bm *BaseModel) GetSum(d []string) (l *List, err error) {
	//c := m.GetCtx()
	if m := bm.GetModel(); m != nil {
		builder, _ := bm.ReadPrepare()

		l = new(List)

		group := make([]string, 0)
		ms := bm.NewList()
		if err := builder.Select(GetSumFields(m, group)).Find(ms); err == nil {
			sumValue := reflect.Indirect(reflect.ValueOf(ms))
			if sumValue.Len() > 0 {
				l.Info.Sum = sumValue.Index(0).Interface()
			}
		}

		if len(d) > 0 {
			group = append(group, d...)
		}
		builder.Group(group)

		ms = bm.NewList()

		if err = builder.Select(GetSumFields(m, group)).Find(ms); err != nil {
			return l, err
		} else if ms == nil {
			return l, ErrNoRecord
		}

		listValue := reflect.Indirect(reflect.ValueOf(ms))
		l.Total = int64(listValue.Len())

		l.List = ms

		return
	} else {
		err := fmt.Errorf("not found model")
		Info("error: %s", err)
		return nil, err
	}
}

/* }}} */

/* {{{ func (bm *BaseModel) GetCount() (cnt int64, err error)
 * 获取list, 通用函数
 */
func (bm *BaseModel) GetCount() (cnt int64, err error) {
	builder, _ := bm.ReadPrepare()
	return builder.Count()

}

/* }}} */

/* {{{ func (bm *BaseModel) GetOlder() Model
 * 获取旧记录
 */
func (bm *BaseModel) GetOlder() Model {
	if bm.older == nil {
		if m := bm.GetModel(); m != nil {
			c := m.GetCtx()
			if rk, ok := c.URLParams[RowkeyKey]; ok && c.Route.Updating {
				if older, err := m.GetRow(rk); err == nil {
					bm.older = older
				}
			}
		}
	}
	return bm.older
}

/* }}} */

/* {{{ func (bm *BaseModel) AddTable(tags ...string)
 * 注册表结构
 */
func (bm *BaseModel) AddTable(tags ...string) {
	if m := bm.GetModel(); m != nil {
		reflectVal := reflect.ValueOf(m)
		mv := reflect.Indirect(reflectVal).Interface()
		//Debug("table name: %s", bm.TableName())
		tb := bm.TableName()
		gorp.AddTableWithName(mv, tb).SetKeys(true, bm.PKey())

		//data accessor, 默认都是DBTAG
		DataAccessor[tb+"::"+WRITETAG] = DBTAG
		DataAccessor[tb+"::"+READTAG] = DBTAG
		if len(tags) > 0 {
			writeTag := tags[0]
			if dns := Config().String("data::" + writeTag); dns != "" {
				Info("%s's writer: %s", tb, dns)
				if err := OpenDB(writeTag, dns); err != nil {
					Warn("open db(%s) error: %s", writeTag, err)
				} else {
					DataAccessor[tb+"::"+WRITETAG] = writeTag
				}
			}
		}
		if len(tags) > 1 {
			readTag := tags[1]
			if dns := Config().String("data::" + readTag); dns != "" {
				Info("%s's reader: %s", tb, dns)
				if err := OpenDB(readTag, dns); err != nil {
					Warn("open db(%s) error: %s", readTag, err)
				} else {
					DataAccessor[tb+"::"+READTAG] = readTag
				}
			}
		}
	} else {
		err := fmt.Errorf("not found model")
		Info("error: %s", err)
	}
}

/* }}} */

/* {{{ func (bm *BaseModel) ReadPrepare() (b *gorp.Builder, err error)
 * 查询准备
 */
func (bm *BaseModel) ReadPrepare() (b *gorp.Builder, err error) {
	var m Model
	if m = bm.GetModel(); m == nil {
		err := fmt.Errorf("not found model")
		Info("error: %s", err)
		return nil, err
	}
	db := bm.DBConn(READTAG)
	tb := bm.TableName()
	b = gorp.NewBuilder(db).Table(tb)
	cons := bm.GetConditions()

	// condition
	if len(cons) > 0 {
		//range condition,先搞范围查询
		for _, v := range cons {
			if v.Range != nil {
				//Debug("[perpare]timerange")
				switch vt := v.Range.(type) {
				case *TimeRange: //只支持timerange
					b.Where(fmt.Sprintf("T.`%s` BETWEEN ? AND ?", v.Field), vt.Start, vt.End)
				case TimeRange: //只支持timerange
					b.Where(fmt.Sprintf("T.`%s` BETWEEN ? AND ?", v.Field), vt.Start, vt.End)
				default:
					//nothing
				}
			}
		}
		jc := 0
		orCons := make(map[string][]string)
		for _, v := range cons {
			//Debug("[key: %s]%v", v.Field, v)
			v.DoWhere(b) //已经处理了 raw/is/not/like/gt/lt
			if v.Or != nil {
				//Debug("[OR][key: %s]%v", v.Field, v)
				oc := v.Or.(*Condition)
				orKey := oc.Field
				if orCons[orKey] == nil {
					orCons[orKey] = make([]string, 0)
				}
				//Debug("or condition: %s", orKey)
				switch ot := oc.Is.(type) {
				case string:
					//Debug("or condition: %s, field: %s", orKey, v.Field)
					orCons[orKey] = append(orCons[orKey], fmt.Sprintf("T.`%s` = '%s'", v.Field, ot))
				case []string:
					vs := bytes.Buffer{}
					first := true
					vs.WriteString("(")
					for _, vv := range ot {
						if !first {
							vs.WriteString(",")
						}
						vs.WriteString(fmt.Sprintf("'%s'", vv))
						first = false
					}
					vs.WriteString(")")
					orCons[orKey] = append(orCons[orKey], fmt.Sprintf("T.`%s` IN %s", v.Field, vs.String()))
				case []interface{}:
					vs := bytes.Buffer{}
					first := true
					vs.WriteString("(")
					for _, vv := range ot {
						if !first {
							vs.WriteString(",")
						}
						vs.WriteString(fmt.Sprintf("'%s'", vv))
						first = false
					}
					vs.WriteString(")")
					orCons[orKey] = append(orCons[orKey], fmt.Sprintf("T.`%s` IN %s", v.Field, vs.String()))
				default:
				}
			}
			if v.Join != nil { //关联查询
				switch vt := v.Join.(type) {
				case *Condition:
					//c.Trace("%s will join %s.%s", v.Field, v.Field, vt.Field)
					if vt.Is != nil {
						jt := v.Field
						jf := vt.Field
						var canJoin bool
						if t, ok := gorp.GetTable(jt); ok {
							//c.Trace("table: %s; type name: %s", jt, t.Gotype.Name())
							if cols := utils.ReadStructColumns(reflect.New(t.Gotype).Interface(), true); cols != nil {
								for _, col := range cols {
									//c.Trace("%s %s", jt, col.Tag)
									if col.Tag == jf && col.ExtOptions.Contains(TAG_CONDITION) { //可作为条件
										//c.Trace("%s.%s can join", jt, jf)
										canJoin = true
										break
									}
								}
							}
						}
						if canJoin {
							js := fmt.Sprintf("LEFT JOIN `%s` T%d ON T.`%s` = T%d.`id`", jt, jc, v.Field, jc)
							b.Joins(js)
							b.Where(fmt.Sprintf("T%d.`%s`=?", jc, jf), vt.Is.(string))
							jc++
						}
					}
				default:
					//c.Trace("not support !*Condition: %v", vt)
				}
			}
		}
		if len(orCons) > 0 {
			for _, css := range orCons {
				b.Where("(" + strings.Join(css, " OR ") + ")")
			}
		}
	} else { //没有条件从自身找
		//Debug("find condition from struct")
		if cols := utils.ReadStructColumns(m, true); cols != nil {
			v := reflect.ValueOf(m)
			for _, col := range cols {
				fv := utils.FieldByIndex(v, col.Index)
				if (col.TagOptions.Contains(DBTAG_PK) || col.ExtOptions.Contains(TAG_CONDITION)) && fv.IsValid() && !utils.IsEmptyValue(fv) { //有值
					if fs := utils.GetRealString(fv); fs != "" {
						// 多个字段有值, 用AND
						b.Where(fmt.Sprintf("T.`%s` = ?", col.Tag), fs)
					}
				}
			}
		}
	}

	//order
	ordered := false
	for _, v := range cons {
		if v.Order != nil {
			switch vt := v.Order.(type) {
			case *OrderBy:
				b.Order(fmt.Sprintf("T.`%s` %s", vt.Field, vt.Sort))
				ordered = true
			case OrderBy:
				b.Order(fmt.Sprintf("T.`%s` %s", vt.Field, vt.Sort))
				ordered = true
			default:
				//nothing
			}
		}
	}
	if !ordered {
		//默认排序
		if cols := utils.ReadStructColumns(m, true); cols != nil {
			for _, col := range cols {
				if col.ExtOptions.Contains(TAG_ORDERBY) { // 默认为pk降序
					b.Order(fmt.Sprintf("T.`%s` DESC", col.Tag))
					ordered = true
				} else if col.ExtOptions.Contains(TAG_AORDERBY) { //正排序
					b.Order(fmt.Sprintf("T.`%s` ASC", col.Tag))
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

	//b.Where(SkipLogicDeleted(m))
	return
}

/* }}} */

/* {{{ func underscore(str string) string
 *
 */
func underscore(str string) string {
	buf := bytes.Buffer{}
	for i, s := range str {
		if s <= 'Z' && s >= 'A' {
			if i > 0 {
				buf.WriteString("_")
			}
			buf.WriteString(string(s + 32))
		} else {
			buf.WriteString(string(s))
		}
	}
	return buf.String()
}

/* }}} */

/* {{{ GetDbFields(i interface{}, ops ...interface{}) (s string)
 * 从struct中解析数据库字段以及字段选项
 */
func GetDbFields(i interface{}, ops ...interface{}) (s []string) {
	var readTag bool
	if len(ops) > 0 {
		if st, ok := ops[0].(bool); ok && st == true {
			readTag = true
		}
	}

	fs := i.(Model).GetFields()
	if cols := utils.ReadStructColumns(i, true); cols != nil {
		s = make([]string, 0)
		for _, col := range cols {
			if col.Tag == "-" { //无此字段
				continue
			} else if readTag && col.ExtOptions.Contains(TAG_SECRET) { //默认忽略tag
				continue
			} else if len(fs) > 0 && !col.TagOptions.Contains(DBTAG_PK) && !utils.InSlice(col.Tag, fs) {
				continue
			}
			s = append(s, col.Tag)
		}
	}
	return
}

/* }}} */

/* {{{ func GetSumFields(i interface{}, g []string) (s string)
 * 从struct中解析数据库字段以及字段选项,为了报表
 */
func GetSumFields(i interface{}, g []string) (s string) {
	if cols := utils.ReadStructColumns(i, true); cols != nil {
		bs := bytes.Buffer{}
		first := true
		for _, col := range cols {
			if !col.ExtOptions.Contains(TAG_REPORT) { //不是报表字段,不对外
				continue
			}
			if col.ExtOptions.Contains(TAG_SECRET) { //保密,不对外
				continue
			}
			if col.ExtOptions.Contains(TAG_CANGROUP) && !utils.InSlice(col.Tag, g) {
				continue
			}
			if !first {
				bs.WriteString(",")
			}
			if col.ExtOptions.Contains(TAG_SUM) {
				bs.WriteString(fmt.Sprintf("SUM(T.`%s`) AS `%s`", col.Tag, col.Tag))
			} else if col.ExtOptions.Contains(TAG_COUNT) {
				bs.WriteString(fmt.Sprintf("COUNT(T.`%s`) AS `%s`", col.Tag, col.Tag))
			} else {
				bs.WriteString("T.`" + col.Tag + "`")
			}
			first = false
		}
		s = bs.String()
	}
	return
}

/* }}} */
