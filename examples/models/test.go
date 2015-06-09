package models

import (
	"reflect"
	"time"

	"github.com/Odinman/ogo"
)

// 以下struct需要跟数据库映射
type Test struct {
	Id       *string    `json:"id,omitempty" db:",pk" filter:",C,G"`
	Name     *string    `json:"name,omitempty" filter:",C"`
	Type     *int       `json:"type,omitempty" filter:",C"` // type不可编辑
	Status   *int       `json:"status,omitempty" db:",logic" filter:",C"`
	Creator  *string    `json:"creator,omitempty" filter:"userid,G,D"`
	Remark   *string    `json:"remark,omitempty" filter:"test"`
	User     *string    `json:"user,omitempty" filter:",G,D,C,PSU"`
	Items    []string   `json:"items,omitempty"`
	Created  *time.Time `json:"created,omitempty" db:",add_now" filter:",G,F,C,TR"`
	Modified *time.Time `json:"modified,omitempty" filter:",G"`
	ogo.BaseModel
}

func init() {
	ogo.NewModel(new(Test)).AddTable("adminwrite")
	ogo.AddTagHook("test", new(Test).Test)
}

//func (this *Test) New() ogo.Model {
//	t := new(Test)
//	t.Tst = "hi"
//	return t
//}

func (this *Test) Test(v reflect.Value) reflect.Value {
	t := "test"
	return reflect.ValueOf(&t)
}
