// Ogo

package ogo

import (
	"time"
)

type Aggregation struct {
	Key    string `json:"key"`
	Value  string `json:"value"`
	Count  string `json:"count"`
	Amount string `json:"amount"`
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
