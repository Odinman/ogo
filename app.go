// Ogo

package ogo

import (
	"time"
)

type Application struct {
	Id        *string    `json:"id,omitempty" db:",pk" filter:",G"`
	Name      *string    `json:"name,omitempty" filter:",C"`
	Type      *int       `json:"type,omitempty" filter:",C"` // type不可编辑
	Status    *int       `json:"status,omitempty" db:",logic" filter:",C"`
	Uuid      *string    `json:"uuid,omitempty" filter:"uuid,G,C,D,RET"`        //短uuid
	AccessKey *string    `json:"access_key,omitempty" filter:"luuid,G,D,C,RET"` //长uuid
	User      *int       `json:"user,omitempty" filter:",C,D"`
	Creator   *string    `json:"creator,omitempty" filter:"userid,G,D"`
	Remark    *string    `json:"remark,omitempty"`
	Version   *string    `json:"version,omitempty" filter:",D,C"`
	Created   *time.Time `json:"created,omitempty" db:",add_now" filter:",G,F"`
	Modified  *time.Time `json:"modified,omitempty" filter:",G"`
	BaseModel
}
