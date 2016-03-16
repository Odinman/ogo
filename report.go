// Ogo

package ogo

import (
	"reflect"
	"strconv"
	"time"

	"github.com/Odinman/ogo/utils"
)

type Aggregation struct {
	Key    string  `json:"key,omitempty"`
	Value  string  `json:"value,omitempty"`
	Count  int     `json:"count"`
	Amount float64 `json:"amount"`
}
type Aggregations map[string][]Aggregation
type Report struct {
	ReportInfo   *ReportInfo  `json:"report_info,omitempty"`
	Aggregations Aggregations `json:"aggregations,omitempty"` //聚合
	List         *List        `json:"list,omitempty"`
}
type ReportInfo struct {
	Dimension []string   `json:"dimension,omitempty"` // 维度
	Type      string     `json:"type,omitempty"`      // 报表类型
	Tz        string     `json:"tz,omitempty"`        // 时区
	Currency  string     `json:"currency,omitempty"`  // 货币
	Start     *time.Time `json:"start,omitempty"`     // 开始时间
	End       *time.Time `json:"end,omitempty"`       // 开始时间
}

/* {{{ func (rpt *Report) WithDefaults() *Report
 *
 */
func (rpt *Report) WithDefaults() *Report {
	// info
	rpt.ReportInfo = new(ReportInfo)
	rpt.ReportInfo.Type = "simple"
	rpt.ReportInfo.Currency = "CNY"
	rpt.ReportInfo.Tz = Env().Location.String()

	return rpt
}

/* }}} */

/* {{{ func UpdateAggregation(as []Aggregation, a Aggregation) []Aggregation
 *
 */
func UpdateAggregation(as []Aggregation, a Aggregation) []Aggregation {
	if len(as) > 0 {
		for i, ta := range as {
			if ta.Key == a.Key {
				as[i].Count += a.Count
				as[i].Amount += a.Amount
				return as
			}
		}
	}
	//Debug("[append: %s]", a.Key)
	return append(as, a)
}

/* }}} */

/* {{{ func BuildAggregationsFromList(l *List, groups ...[]string) Aggregations
 * 从list生成Aggregations, 传入groups
 */
func BuildAggregationsFromList(l *List, groups ...[]string) (as Aggregations) {
	listValue := reflect.Indirect(reflect.ValueOf(l.List))
	as = make(Aggregations)
	var gs []string
	if len(groups) > 0 {
		gs = groups[0]
	}
	for i := 0; i < listValue.Len(); i++ {
		row := listValue.Index(i).Interface().(Model)
		if cols := utils.ReadStructColumns(row, true); cols != nil {
			rv := reflect.ValueOf(row)
			for _, col := range cols {
				frv := utils.FieldByIndex(rv, col.Index)
				if frv.IsValid() && !utils.IsEmptyValue(frv) {
					if col.ExtOptions.Contains(TAG_AGGREGATION) && (gs == nil || utils.InSlice(col.Tag, gs)) { //聚合元素
						var key string
						switch frv.Type().String() {
						case "*string":
							key = frv.Elem().String()
						case "string":
							key = frv.String()
						case "*int":
							key = strconv.Itoa(int(frv.Elem().Int()))
						default:
							continue
						}
						if _, ok := as[col.Tag]; !ok {
							as[col.Tag] = make([]Aggregation, 0)
						}
						cnt, sum := row.GetCountNSum()
						as[col.Tag] = UpdateAggregation(as[col.Tag], Aggregation{Key: key, Count: int(cnt), Amount: sum})
					}
				}
			}
		}
	}
	return as
}

/* }}} */
