// Ogo

package ogo

import (
	"time"
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
