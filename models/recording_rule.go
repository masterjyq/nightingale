package models

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ccfos/nightingale/v6/pkg/ctx"
	"github.com/ccfos/nightingale/v6/pkg/poster"

	"github.com/pkg/errors"
	"github.com/prometheus/common/model"
	"github.com/toolkits/pkg/logger"
)

// A RecordingRule records its vector expression into new timeseries.
type RecordingRule struct {
	Id                int64             `json:"id" gorm:"primaryKey"`
	GroupId           int64             `json:"group_id"` // busi group id
	DatasourceIds     string            `json:"-" gorm:"datasource_ids,omitempty"`
	DatasourceIdsJson []int64           `json:"datasource_ids" gorm:"-"`                             // for open source fe
	DatasourceQueries []DatasourceQuery `json:"datasource_queries,omitempty" gorm:"serializer:json"` // datasource queries
	Cluster           string            `json:"cluster"`                                             // take effect by cluster, seperated by space
	Name              string            `json:"name"`                                                // new metric name
	Disabled          int               `json:"disabled"`                                            // 0: enabled, 1: disabled
	PromQl            string            `json:"prom_ql"`                                             // just one ql for promql
	QueryConfigs      string            `json:"-" gorm:"query_configs"`                              // query_configs
	QueryConfigsJson  []QueryConfig     `json:"query_configs" gorm:"-"`                              // query_configs for fe
	PromEvalInterval  int               `json:"prom_eval_interval"`                                  // unit:s
	CronPattern       string            `json:"cron_pattern"`
	AppendTags        string            `json:"-"`                    // split by space: service=n9e mod=api
	AppendTagsJSON    []string          `json:"append_tags" gorm:"-"` // for fe
	Note              string            `json:"note"`                 // note
	CreateAt          int64             `json:"create_at"`
	CreateBy          string            `json:"create_by"`
	UpdateAt          int64             `json:"update_at"`
	UpdateBy          string            `json:"update_by"`
}

type QueryConfig struct {
	Queries           []Query `json:"queries"`
	NewMetric         string  `json:"new_metric"`
	Exp               string  `json:"exp"`
	WriteDatasourceId int64   `json:"write_datasource_id"`
	Delay             int     `json:"delay"`
}

type Query struct {
	DatasourceIds     []int64           `json:"datasource_ids"`
	DatasourceQueries []DatasourceQuery `json:"datasource_queries"`
	Cate              string            `json:"cate"`
	Config            interface{}       `json:"config"`
}

func (re *RecordingRule) TableName() string {
	return "recording_rule"
}

func (re *RecordingRule) FE2DB() {
	re.AppendTags = strings.Join(re.AppendTagsJSON, " ")
	idsByte, _ := json.Marshal(re.DatasourceIdsJson)
	re.DatasourceIds = string(idsByte)

	queryConfigsByte, _ := json.Marshal(re.QueryConfigsJson)
	re.QueryConfigs = string(queryConfigsByte)
}

func (re *RecordingRule) DB2FE() error {
	re.AppendTagsJSON = strings.Fields(re.AppendTags)
	json.Unmarshal([]byte(re.DatasourceIds), &re.DatasourceIdsJson)

	re.FillDatasourceQueries()

	json.Unmarshal([]byte(re.QueryConfigs), &re.QueryConfigsJson)
	// 存量数据规则不包含 DatasourceQueries 字段，将 DatasourceIds 转换为 DatasourceQueries 字段
	for i := range re.QueryConfigsJson {
		for j := range re.QueryConfigsJson[i].Queries {
			if len(re.QueryConfigsJson[i].Queries[j].DatasourceQueries) == 0 {
				values := make([]interface{}, 0, len(re.QueryConfigsJson[i].Queries[j].DatasourceIds))
				for _, dsID := range re.QueryConfigsJson[i].Queries[j].DatasourceIds {
					values = append(values, dsID)
				}
				re.QueryConfigsJson[i].Queries[j].DatasourceQueries = []DatasourceQuery{
					{
						MatchType: 0,
						Op:        "in",
						Values:    values,
					},
				}
			}
		}
	}

	if re.CronPattern == "" && re.PromEvalInterval != 0 {
		re.CronPattern = fmt.Sprintf("@every %ds", re.PromEvalInterval)
	}

	return nil
}

func (re *RecordingRule) FillDatasourceQueries() error {
	// 兼容旧逻辑，将 datasourceIds 转换为 datasourceQueries
	if len(re.DatasourceQueries) == 0 && len(re.DatasourceIds) != 0 {
		datasourceQueries := DatasourceQuery{
			MatchType: 0,
			Op:        "in",
			Values:    make([]interface{}, 0),
		}

		var values []int64
		if re.DatasourceIds != "" {
			json.Unmarshal([]byte(re.DatasourceIds), &values)
		}

		for i := range values {
			if values[i] == 0 {
				// 0 表示所有数据源
				datasourceQueries.MatchType = 2
				break
			}
			datasourceQueries.Values = append(datasourceQueries.Values, values[i])
		}

		re.DatasourceQueries = []DatasourceQuery{datasourceQueries}
	}
	return nil
}

func (re *RecordingRule) Verify() error {
	if re.GroupId < 0 {
		return fmt.Errorf("GroupId(%d) invalid", re.GroupId)
	}

	//if IsAllDatasource(re.DatasourceIdsJson) {
	//	re.DatasourceIdsJson = []int64{0}
	//}

	if re.PromQl != "" && !model.MetricNameRE.MatchString(re.Name) {
		return errors.New("Name has invalid chreacters")
	}

	for _, queryConfig := range re.QueryConfigsJson {
		if !model.MetricNameRE.MatchString(queryConfig.NewMetric) {
			return errors.New("Metric Name has invalid chreacters")
		}
	}

	if re.Name == "" && re.PromQl != "" {
		return errors.New("name is blank")
	}

	if re.PromEvalInterval <= 0 {
		re.PromEvalInterval = 60
	}

	if re.CronPattern == "" {
		re.CronPattern = "@every 60s"
	}

	re.AppendTags = strings.TrimSpace(re.AppendTags)
	rer := strings.Fields(re.AppendTags)
	for i := 0; i < len(rer); i++ {
		pair := strings.Split(rer[i], "=")
		if len(pair) != 2 || !model.LabelNameRE.MatchString(pair[0]) {
			return fmt.Errorf("AppendTags(%s) invalid", rer[i])
		}
	}

	return nil
}

func (re *RecordingRule) Add(ctx *ctx.Context) error {
	if err := re.Verify(); err != nil {
		return err
	}

	// 由于实际场景中会出现name重复的recording rule，所以不需要检查重复
	//exists, err := RecordingRuleExists(0, re.GroupId, re.Cluster, re.Name)
	//if err != nil {
	//	return err
	//}
	//
	//if exists {
	//	return errors.New("RecordingRule already exists")
	//}

	now := time.Now().Unix()
	re.CreateAt = now
	re.UpdateAt = now

	return Insert(ctx, re)
}

func (re *RecordingRule) Update(ctx *ctx.Context, ref RecordingRule) error {
	// 由于实际场景中会出现name重复的recording rule，所以不需要检查重复
	//if re.Name != ref.Name {
	//	exists, err := RecordingRuleExists(re.Id, re.GroupId, re.Cluster, ref.Name)
	//	if err != nil {
	//		return err
	//	}
	//	if exists {
	//		return errors.New("RecordingRule already exists")
	//	}
	//}

	ref.FE2DB()
	ref.Id = re.Id
	ref.GroupId = re.GroupId
	ref.CreateAt = re.CreateAt
	ref.CreateBy = re.CreateBy
	ref.UpdateAt = time.Now().Unix()
	err := ref.Verify()
	if err != nil {
		return err
	}
	return DB(ctx).Model(re).Select("*").Updates(ref).Error
}

func (re *RecordingRule) UpdateFieldsMap(ctx *ctx.Context, fields map[string]interface{}) error {
	return DB(ctx).Model(re).Updates(fields).Error
}

func RecordingRuleDels(ctx *ctx.Context, ids []int64, groupId int64) error {
	for i := 0; i < len(ids); i++ {
		ret := DB(ctx).Where("id = ? and group_id=?", ids[i], groupId).Delete(&RecordingRule{})
		if ret.Error != nil {
			return ret.Error
		}
	}

	return nil
}

// func RecordingRuleExists(ctx *ctx.Context, id, groupId int64, cluster, name string) (bool, error) {
// 	session := DB(ctx).Where("id <> ? and group_id = ? and name =? ", id, groupId, name)

// 	var lst []RecordingRule
// 	err := session.Find(&lst).Error
// 	if err != nil {
// 		return false, err
// 	}
// 	if len(lst) == 0 {
// 		return false, nil
// 	}

// 	// match cluster
// 	for _, r := range lst {
// 		if MatchCluster(r.Cluster, cluster) {
// 			return true, nil
// 		}
// 	}
// 	return false, nil
// }

func RecordingRuleGets(ctx *ctx.Context, groupId int64) ([]RecordingRule, error) {
	session := DB(ctx).Where("group_id=?", groupId).Order("name")

	var lst []RecordingRule
	err := session.Find(&lst).Error
	if err == nil {
		for i := 0; i < len(lst); i++ {
			lst[i].DB2FE()
		}
	}

	return lst, err
}

func RecordingRuleGetsByBGIds(ctx *ctx.Context, bgids []int64) ([]RecordingRule, error) {
	session := DB(ctx)
	if len(bgids) > 0 {
		session = session.Where("group_id in (?)", bgids).Order("name")
	}

	var lst []RecordingRule
	err := session.Find(&lst).Error
	if err == nil {
		for i := 0; i < len(lst); i++ {
			lst[i].DB2FE()
		}
	}

	return lst, err
}

func RecordingRuleGet(ctx *ctx.Context, where string, regs ...interface{}) (*RecordingRule, error) {
	var lst []*RecordingRule
	err := DB(ctx).Where(where, regs...).Find(&lst).Error
	if err != nil {
		return nil, err
	}

	if len(lst) == 0 {
		return nil, nil
	}

	lst[0].DB2FE()

	return lst[0], nil
}

func RecordingRuleGetById(ctx *ctx.Context, id int64) (*RecordingRule, error) {
	return RecordingRuleGet(ctx, "id=?", id)
}

func RecordingRuleEnabledGets(ctx *ctx.Context) ([]*RecordingRule, error) {
	session := DB(ctx)

	var lst []*RecordingRule
	err := session.Where("disabled = ?", 0).Find(&lst).Error
	if err != nil {
		return lst, err
	}

	for i := 0; i < len(lst); i++ {
		lst[i].DB2FE()
	}
	return lst, nil
}

func RecordingRuleGetsByCluster(ctx *ctx.Context) ([]*RecordingRule, error) {
	if !ctx.IsCenter {
		lst, err := poster.GetByUrls[[]*RecordingRule](ctx, "/v1/n9e/recording-rules")
		if err != nil {
			return nil, err
		}
		for i := 0; i < len(lst); i++ {
			lst[i].FE2DB()
		}
		return lst, err
	}

	session := DB(ctx).Where("disabled = ?", 0)

	var lst []*RecordingRule
	err := session.Find(&lst).Error
	if err != nil {
		return lst, err
	}

	if len(lst) == 0 {
		return lst, nil
	}

	for i := 0; i < len(lst); i++ {
		lst[i].DB2FE()
	}
	return lst, nil
}

func RecordingRuleStatistics(ctx *ctx.Context) (*Statistics, error) {
	if !ctx.IsCenter {
		s, err := poster.GetByUrls[*Statistics](ctx, "/v1/n9e/statistic?name=recording_rule")
		return s, err
	}

	session := DB(ctx).Model(&RecordingRule{}).Select("count(*) as total", "max(update_at) as last_updated")

	var stats []*Statistics
	err := session.Find(&stats).Error
	if err != nil {
		return nil, err
	}

	return stats[0], nil
}

func RecordingRuleUpgradeToV6(ctx *ctx.Context, dsm map[string]Datasource) error {
	var lst []*RecordingRule
	err := DB(ctx).Find(&lst).Error
	if err != nil {
		return err
	}

	for i := 0; i < len(lst); i++ {
		var ids []int64
		if lst[i].Cluster == "$all" {
			ids = append(ids, 0)
		} else {
			clusters := strings.Fields(lst[i].Cluster)
			for j := 0; j < len(clusters); j++ {
				if ds, exists := dsm[clusters[j]]; exists {
					ids = append(ids, ds.Id)
				}
			}
		}

		b, err := json.Marshal(ids)
		if err != nil {
			continue
		}
		lst[i].DatasourceIds = string(b)

		err = lst[i].UpdateFieldsMap(ctx, map[string]interface{}{"datasource_ids": lst[i].DatasourceIds})
		if err != nil {
			logger.Errorf("update alert rule:%d datasource ids failed, %v", lst[i].Id, err)
		}
	}
	return nil
}
