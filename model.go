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
	TAG_HASH       = "H"   // 需要hash转化
	TAG_SECRET     = "S"   //保密,一般不见人
	TAG_TIMERANGE  = "TR"  //时间范围条件
	TAG_REPORT     = "RPT" //报表字段
	TAG_ORDERBY    = "O"   //可排序
	TAG_VERIFIABLE = "V"   //验证后可修改
)

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
	DBConn(tag string) *gorp.DbMap                                     // 数据库连接
	TableName(m Model) string                                          // 返回表名称, 默认结构type名字(小写), 有特别的表名称,则自己implement 这个方法
	PKey(m Model) string                                               // key字段
	New() Model                                                        // 返回一个空的model结构
	List() interface{}                                                 // 返回一个空结构列表
	OrderBy() string                                                   // 排序
	PostGet(m Model, c *RESTContext) (Model, error)                    //获取后调整
	PreCreate(m Model, c *RESTContext) (Model, error)                  // 插入前的检查
	PostCreate(m Model, c *RESTContext) (Model, error)                 // 插入后的处理
	PreUpdate(m Model, c *RESTContext) (Model, error)                  // 更新前的检查
	PostUpdate(m Model, c *RESTContext) (Model, error)                 // 更新前的检查
	PreDelete(m Model, c *RESTContext) (Model, error)                  // 删除前的检查
	OnAction(m Model, c *RESTContext) error                            //触发器
	Existense(m Model, c *RESTContext) func(tag string) (Model, error) //检查存在性
}

//基础model,在这里可以实现Model接口, 其余的只需要嵌入这个struct,就可以继承这些方法
type BaseModel struct {
}

/* {{{ func (b *BaseModel) DBConn(tag string) *gorp.DbMap
 * 默认数据库连接为admin
 */
func (b *BaseModel) DBConn(tag string) *gorp.DbMap {
	return gorp.Using(tag)
}

/* }}} */

/* {{{ func (b *BaseModel) TableName(m Model) (n string)
 * 获取表名称, 默认为结构名
 */
func (b *BaseModel) TableName(m Model) (n string) {
	//默认, struct的名字就是表名, 如果不是请在各自的model里定义
	reflectVal := reflect.ValueOf(m)
	mt := reflect.Indirect(reflectVal).Type()
	n = underscore(strings.TrimSuffix(mt.Name(), "Table"))
	return
}

/* }}} */

/* {{{ func (b *BaseModel) PKey(m Model) string
 *  通过配置找到pk
 */
func (b *BaseModel) PKey(m Model) string {
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

/* {{{ func (b *BaseModel) New() Model
 *
 */
func (b *BaseModel) New() Model {
	return new(BaseModel)
}

/* }}} */

/* {{{ func (b *BaseModel) List() interface{}
 *
 */
func (b *BaseModel) List() interface{} {
	return &[]BaseModel{}
}

/* }}} */

/* {{{ func (b *BaseModel) OrderBy() string
 *
 */
func (b *BaseModel) OrderBy() string {
	return ""
}

/* }}} */

/* {{{ func (b *BaseModel) PostGet(m Model, c *RESTContext) (Model, error)
 * 创建后检查
 */
func (b *BaseModel) PostGet(m Model, c *RESTContext) (Model, error) {
	return m, nil
}

/* }}} */

/* {{{ func (b *BaseModel) PreCreate(m Model, c *RESTContext) (Model, error)
 * 创建前检查
 */
func (b *BaseModel) PreCreate(m Model, c *RESTContext) (Model, error) {
	return m, nil
}

/* }}} */

/* {{{ func (b *BaseModel) PostCreate(m Model, c *RESTContext) (Model, error)
 * 创建后检查
 */
func (b *BaseModel) PostCreate(m Model, c *RESTContext) (Model, error) {
	return m, nil
}

/* }}} */

/* {{{ func (b *BaseModel) PreUpdate(m Model, c *RESTContext) (Model, error)
 * 更新前检查
 */
func (b *BaseModel) PreUpdate(m Model, c *RESTContext) (Model, error) {
	return m, nil
}

/* }}} */

/* {{{ func (b *BaseModel) PreDelete(m Model, c *RESTContext) (Model, error)
 * 删除前检查
 */
func (b *BaseModel) PreDelete(m Model, c *RESTContext) (Model, error) {
	return m, nil
}

/* }}} */

/* {{{ func (b *BaseModel) PostUpdate(m Model, c *RESTContext) (Model, error)
 * 创建后检查
 */
func (b *BaseModel) PostUpdate(m Model, c *RESTContext) (Model, error) {
	return m, nil
}

/* }}} */

/* {{{ func (b *BaseModel) Existense(m Model, c *RESTContext) func(tag string) (Model, error)
 *
 */
func (b *BaseModel) Existense(m Model, c *RESTContext) func(tag string) (Model, error) {
	return func(tag string) (Model, error) {
		return nil, fmt.Errorf("can't check")
	}
}

/* }}} */

/* {{{ func (b *BaseModel) OnAction(m Model, c *RESTContext) error
 * 触发器, 一般是状态达到某个特定值之后的动作, 原则上这个动作执行失败, 别的程序可以补救
 */
func (b *BaseModel) OnAction(m Model, c *RESTContext) error {
	return fmt.Errorf("nothing to do")
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

/* {{{ func (b *BaseModel) GetRow(m Model, id string, c *RESTContext, exts ...interface{}) (Model, error)
 * 根据条件获取一条记录, model为表结构
 */
func (b *BaseModel) GetRow(m Model, id string, c *RESTContext, exts ...interface{}) (Model, error) {
	if id == "" && len(exts) <= 0 {
		return m, ErrInvalid
	}
	cons := make(Conditions)
	if id != "" {
		pk := m.PKey(m)
		cons[pk] = &Condition{
			Field: pk,
			Is:    id,
		}
	} else if len(exts) > 0 { //条件
		if exCons, ok := exts[0].(map[string]string); ok {
			for k, v := range exCons {
				cons[k] = &Condition{
					Field: k,
					Is:    v,
				}
			}
		}
	}

	builder, _ := ReadPrepare(m, cons, c)

	i := m.New().(Model).List()
	err := builder.Select(GetDbFields(m)).Find(i)
	if err != nil && err != sql.ErrNoRows {
		//支持出错
		return nil, err
	} else if i == nil {
		//没找到记录
		return nil, ErrNoRecord
	}
	//return reflect.ValueOf(m).Index(0), nil
	resultsValue := reflect.Indirect(reflect.ValueOf(i))
	if resultsValue.Len() <= 0 {
		return nil, ErrNoRecord
	}
	return resultsValue.Index(0).Interface().(Model), nil
}

/* }}} */
