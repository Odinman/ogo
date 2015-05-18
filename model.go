// Ogo

package ogo

import (
	"bytes"
	"database/sql"
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
	TAG_REQUIRED   = "R"   // 必填
	TAG_GENERATE   = "G"   // 服务端生成, 同时不可编辑
	TAG_CONDITION  = "C"   // 可作为查询条件
	TAG_DENY       = "D"   // 不可编辑, 可为空
	TAG_SECRET     = "S"   //保密,一般不见人
	TAG_TIMERANGE  = "TR"  //时间范围条件
	TAG_REPORT     = "RPT" //报表字段
	TAG_ORDERBY    = "O"   //可排序
	TAG_VERIFIABLE = "V"   //验证后可修改
	TAG_RETURN     = "RET" // 返回,创建后需要返回数值
)

type List struct {
	Total int64       `json:"total"`
	List  interface{} `json:"list"`
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

type Model interface {
	SetCtx(c *RESTContext)
	GetCtx() *RESTContext
	SetConditions(m Model, ocs Conditions) (Conditions, error)
	GetConditions() Conditions

	New(m Model, c ...interface{}) Model
	NewList(m Model) interface{} // 返回一个空结构列表

	AddTable(m Model, tags ...string)
	DBConn(m Model, tag string) *gorp.DbMap // 数据库连接
	TableName(m Model) string               // 返回表名称, 默认结构type名字(小写), 有特别的表名称,则自己implement 这个方法
	PKey(m Model) string                    // key字段
	ReadPrepare(m Model) (*gorp.Builder, error)

	GetRow(m Model, id string) (Model, error)          //获取单条记录
	GetRows(m Model) (*List, error)                    //获取多条记录
	GetCount(m Model) (int64, error)                   //获取多条记录
	CreateRow(m Model) error                           //创建单条记录
	UpdateRow(m Model, id string) (int64, error)       //更新记录
	DeleteRow(m Model, id string) (int64, error)       //更新记录
	Existense(m Model) func(tag string) (Model, error) //检查存在性
}

//基础model,在这里可以实现Model接口, 其余的只需要嵌入这个struct,就可以继承这些方法
type BaseModel struct {
	Error      error        `json:"-" db:"-"`
	ctx        *RESTContext `json:"-" db:"-"`
	conditions Conditions   `json:"-" db:"-"`
}

/* {{{ func (bm *BaseModel) SetCtx(c *RESTContext) Model
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

/* {{{ func (m *BaseModel) GetConditions() Conditions
 *
 */
func (bm *BaseModel) GetConditions() Conditions {
	return bm.conditions
}

/* }}} */

/* {{{ func (_ *BaseModel) New(m Model, c ...interface{}) Model
 * 初始化model, 后面的c选填
 */
func (_ *BaseModel) New(m Model, c ...interface{}) Model {
	//nm := reflect.ValueOf(m).Interface().(Model)
	nm := reflect.New(reflect.Indirect(reflect.ValueOf(m)).Type()).Interface()
	if len(c) > 0 {
		nm.(Model).SetCtx(c[0].(*RESTContext))
	}
	return nm.(Model)
}

/* }}} */

/* {{{ func (m *BaseModel) NewList(m Model) *[]Model
 *
 */
func (_ *BaseModel) NewList(m Model) interface{} {
	//ms := reflect.New(reflect.SliceOf(reflect.TypeOf(m))).Interface().(*[]Model)
	ms := reflect.New(reflect.SliceOf(reflect.TypeOf(m))).Interface()
	return ms
}

/* }}} */

/* {{{ func (m *BaseModel) DBConn(m Model,tag string) *gorp.DbMap
 * 默认数据库连接为admin
 */
func (_ *BaseModel) DBConn(m Model, tag string) *gorp.DbMap {
	tb := m.TableName(m)
	if dt, ok := DataAccessor[tb+"::"+tag]; ok && dt != "" {
		return gorp.Using(dt)
	}
	return gorp.Using(DBTAG)
}

/* }}} */

/* {{{ func (m *BaseModel) TableName(m Model) (n string)
 * 获取表名称, 默认为结构名
 */
func (_ *BaseModel) TableName(m Model) (n string) {
	//默认, struct的名字就是表名, 如果不是请在各自的model里定义
	reflectVal := reflect.ValueOf(m)
	mt := reflect.Indirect(reflectVal).Type()
	n = underscore(strings.TrimSuffix(mt.Name(), "Table"))
	return
}

/* }}} */

/* {{{ func (_ *BaseModel) PKey(m Model) string
 *  通过配置找到pk
 */
func (_ *BaseModel) PKey(m Model) string {
	if cols := utils.ReadStructColumns(m, true); cols != nil {
		for _, col := range cols {
			// check required field
			if col.TagOptions.Contains(DBTAG_PK) {
				return col.Name
			}
		}
	}
	return ""
}

/* }}} */

/* {{{ func (_ *BaseModel) Existense(m Model) func(tag string) (Model, error)
 *
 */
func (_ *BaseModel) Existense(m Model) func(tag string) (Model, error) {
	return func(tag string) (Model, error) {
		return nil, fmt.Errorf("can't check")
	}
}

/* }}} */

/* {{{ func (_ *BaseModel) GetRow(m Model, id string) (Model, error)
 * 根据条件获取一条记录, model为表结构
 */
func (_ *BaseModel) GetRow(m Model, id string) (Model, error) {
	//db := m.DBConn(m, READTAG)
	//if obj, err := db.Get(m, id); err != nil {
	//	//Debug("get error: %s, %v", err, obj)
	//	if err == sql.ErrNoRows {
	//		return nil, ErrNoRecord
	//	} else {
	//		return nil, err
	//	}
	//} else {
	//	return obj.(Model), nil
	//}
	c := m.GetCtx()
	builder, _ := m.ReadPrepare(m)
	ms := m.NewList(m)
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
		c.Debug("len: %d, no record", resultsValue.Len())
		return nil, ErrNoRecord
	}
	return resultsValue.Index(0).Interface().(Model), nil
}

/* }}} */

/* {{{ func (_ *BaseModel) CreateRow(m Model) error
 * 根据条件获取一条记录, model为表结构
 */
func (_ *BaseModel) CreateRow(m Model) error {
	db := m.DBConn(m, WRITETAG)
	return db.Insert(m)
}

/* }}} */

/* {{{ func (_ *BaseModel) UpdateRow(m Model, id string) (affected int64, err error)
 * 根据条件获取一条记录, model为表结构
 */
func (_ *BaseModel) UpdateRow(m Model, id string) (affected int64, err error) {
	db := m.DBConn(m, WRITETAG)
	if id != "" {
		if err = utils.ImportValue(m, map[string]string{DBTAG_PK: id}); err != nil {
			return
		}
	}
	return db.Update(m)
}

/* }}} */

/* {{{ func (_ *BaseModel) DeleteRow(m Model, id string) (affected int64, err error)
 * 删除记录(逻辑删除)
 */
func (_ *BaseModel) DeleteRow(m Model, id string) (affected int64, err error) {
	db := m.DBConn(m, WRITETAG)
	if err = utils.ImportValue(m, map[string]string{DBTAG_PK: id, DBTAG_LOGIC: "-1"}); err != nil {
		return
	}
	return db.Update(m)
}

/* }}} */

/* {{{ func (bm *BaseModel) SetConditions(m Model, ocs Conditions) (cons Conditions, err error)
 * 生成条件
 */
func (bm *BaseModel) SetConditions(m Model, ocs Conditions) (cons Conditions, err error) {
	if bm.conditions == nil {
		bm.conditions = make(Conditions)
	}
	if cols := utils.ReadStructColumns(m, true); cols != nil {
		for _, col := range cols {
			// check required field
			//c.Trace("tag: %s, name: %s, exops: %s", col.Tag, col.Name, col.ExtOptions)
			if col.ExtOptions.Contains(TAG_CONDITION) { //可作为条件
				//c.Trace("tag: %s, name: %s", col.Tag, col.Name)
				if condition, e := ocs.Get(col.Tag); e == nil {
					//Trace("condition: %v", condition)
					bm.conditions[col.Tag] = condition
				} else {
					Trace("get condition failed: %s", e)
				}
			} else {
				//Trace("not condition, tag: %s, name: %s, exttag: %s", col.Tag, col.Name, col.ExtOptions)
			}
		}
	}
	return bm.conditions, nil
}

/* }}} */

/* {{{ func (_ *BaseModel) GetRows(m Model) (l *List, err error)
 * 获取list, 通用函数
 */
func (_ *BaseModel) GetRows(m Model) (l *List, err error) {
	c := m.GetCtx()
	builder, _ := m.ReadPrepare(m)
	count, _ := builder.Count() //结果数
	ms := m.NewList(m)
	p := c.GetEnv(PaginationKey).(*Pagination)
	err = builder.Select(GetDbFields(m)).Offset(p.Offset).Limit(p.PerPage).Find(ms)
	if err != nil && err != sql.ErrNoRows {
		//支持出错
		return l, err
	} else if ms == nil {
		//没找到记录
		return l, ErrNoRecord
	}

	l = &List{
		Total: count,
		List:  ms,
	}

	return l, nil
}

/* }}} */

/* {{{ func (_ *BaseModel) GetCount(m Model) (cnt int64, err error)
 * 获取list, 通用函数
 */
func (_ *BaseModel) GetCount(m Model) (cnt int64, err error) {
	builder, _ := m.ReadPrepare(m)
	return builder.Count()

}

/* }}} */

/* {{{ func (_ *BaseModel) AddTable(m Model, tags ...string)
 * 注册表结构
 */
func (_ *BaseModel) AddTable(m Model, tags ...string) {
	reflectVal := reflect.ValueOf(m)
	mv := reflect.Indirect(reflectVal).Interface()
	Debug("table name: %s", m.TableName(m))
	tb := m.TableName(m)
	gorp.AddTableWithName(mv, tb).SetKeys(true, m.PKey(m))

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
}

/* }}} */

/* {{{ func (_ *BaseModel) ReadPrepare(m Model) (b *gorp.Builder, err error)
 * 查询准备
 */
func (_ *BaseModel) ReadPrepare(m Model) (b *gorp.Builder, err error) {
	db := m.DBConn(m, READTAG)
	tb := m.TableName(m)
	b = gorp.NewBuilder(db).Table(tb)
	c := m.GetCtx()
	cons := m.GetConditions()

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

/* {{{ GetDbFields(i interface{}) (s string)
 * 从struct中解析数据库字段以及字段选项
 */
func GetDbFields(i interface{}) (s []string) {
	if cols := utils.ReadStructColumns(i, true); cols != nil {
		s = make([]string, 0)
		for _, col := range cols {
			if col.Tag == "-" || col.ExtOptions.Contains(TAG_SECRET) { //保密,不对外
				continue
			}
			s = append(s, col.Tag)
		}
	}
	return
}

/* }}} */

/* {{{ func ReportPrepare(model interface{}, cons ogo.Conditions, c *ogo.RESTContext) (b *gorp.Builder, err error)
 * 报表查询准备
 */
func ReportPrepare(m Model, cons Conditions, c *RESTContext) (b *gorp.Builder, err error) {
	db := m.DBConn(m, READTAG)
	tb := m.TableName(m)
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
