package search

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"hr-management-system/internal/config"

	"github.com/olivere/elastic/v7"
)

type ElasticSearch struct {
	client *elastic.Client
	index  string
}

var es *ElasticSearch

func NewElasticSearch(cfg *config.ElasticConfig) (*ElasticSearch, error) {
	opts := []elastic.ClientOptionFunc{
		elastic.SetURL(cfg.URLs...),
		elastic.SetSniff(false),
		elastic.SetHealthcheck(true),
		elastic.SetHealthcheckInterval(30 * time.Second),
	}

	if cfg.Username != "" && cfg.Password != "" {
		opts = append(opts, elastic.SetBasicAuth(cfg.Username, cfg.Password))
	}

	client, err := elastic.NewClient(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create elasticsearch client: %w", err)
	}

	es = &ElasticSearch{
		client: client,
		index:  cfg.Index,
	}

	return es, nil
}

func GetElasticSearch() *ElasticSearch {
	return es
}

func (e *ElasticSearch) Client() *elastic.Client {
	return e.client
}

func (e *ElasticSearch) InitIndices(ctx context.Context) error {
	indices := map[string]string{
		"employees":   employeesMapping,
		"departments": departmentsMapping,
		"attendances": attendancesMapping,
		"payslips":    payslipsMapping,
		"audit_logs":  auditLogsMapping,
	}

	for name, mapping := range indices {
		indexName := fmt.Sprintf("%s_%s", e.index, name)
		exists, err := e.client.IndexExists(indexName).Do(ctx)
		if err != nil {
			return fmt.Errorf("failed to check index %s: %w", indexName, err)
		}

		if !exists {
			_, err := e.client.CreateIndex(indexName).BodyString(mapping).Do(ctx)
			if err != nil {
				return fmt.Errorf("failed to create index %s: %w", indexName, err)
			}
		}
	}

	return nil
}

func (e *ElasticSearch) Index(ctx context.Context, indexName, id string, doc interface{}) error {
	fullIndex := fmt.Sprintf("%s_%s", e.index, indexName)
	_, err := e.client.Index().
		Index(fullIndex).
		Id(id).
		BodyJson(doc).
		Refresh("true").
		Do(ctx)
	return err
}

func (e *ElasticSearch) BulkIndex(ctx context.Context, indexName string, docs map[string]interface{}) error {
	fullIndex := fmt.Sprintf("%s_%s", e.index, indexName)
	bulk := e.client.Bulk()

	for id, doc := range docs {
		req := elastic.NewBulkIndexRequest().
			Index(fullIndex).
			Id(id).
			Doc(doc)
		bulk.Add(req)
	}

	_, err := bulk.Do(ctx)
	return err
}

func (e *ElasticSearch) Get(ctx context.Context, indexName, id string) (*elastic.GetResult, error) {
	fullIndex := fmt.Sprintf("%s_%s", e.index, indexName)
	return e.client.Get().Index(fullIndex).Id(id).Do(ctx)
}

func (e *ElasticSearch) Update(ctx context.Context, indexName, id string, doc interface{}) error {
	fullIndex := fmt.Sprintf("%s_%s", e.index, indexName)
	_, err := e.client.Update().Index(fullIndex).Id(id).Doc(doc).Refresh("true").Do(ctx)
	return err
}

func (e *ElasticSearch) Delete(ctx context.Context, indexName, id string) error {
	fullIndex := fmt.Sprintf("%s_%s", e.index, indexName)
	_, err := e.client.Delete().Index(fullIndex).Id(id).Refresh("true").Do(ctx)
	return err
}

type SearchParams struct {
	Query        string
	Filters      map[string]interface{}
	From         int
	Size         int
	Sort         []string
	Highlight    []string
	Aggregations map[string]elastic.Aggregation
}

type SearchResult struct {
	Total    int64                    `json:"total"`
	Hits     []map[string]interface{} `json:"hits"`
	Aggs     map[string]interface{}   `json:"aggregations,omitempty"`
	MaxScore float64                  `json:"max_score"`
}

func (e *ElasticSearch) Search(ctx context.Context, indexName string, params SearchParams) (*SearchResult, error) {
	fullIndex := fmt.Sprintf("%s_%s", e.index, indexName)
	query := elastic.NewBoolQuery()

	if params.Query != "" {
		query.Must(elastic.NewMultiMatchQuery(params.Query).Type("best_fields").Fuzziness("AUTO"))
	}

	for field, value := range params.Filters {
		switch v := value.(type) {
		case string:
			query.Filter(elastic.NewTermQuery(field, v))
		case []string:
			query.Filter(elastic.NewTermsQueryFromStrings(field, v...))
		case map[string]interface{}:
			if gte, ok := v["gte"]; ok {
				if lte, ok := v["lte"]; ok {
					query.Filter(elastic.NewRangeQuery(field).Gte(gte).Lte(lte))
				} else {
					query.Filter(elastic.NewRangeQuery(field).Gte(gte))
				}
			} else if lte, ok := v["lte"]; ok {
				query.Filter(elastic.NewRangeQuery(field).Lte(lte))
			}
		default:
			query.Filter(elastic.NewTermQuery(field, v))
		}
	}

	search := e.client.Search().
		Index(fullIndex).
		Query(query).
		From(params.From).
		Size(params.Size).
		TrackTotalHits(true)

	for _, s := range params.Sort {
		asc := true
		field := s
		if strings.HasPrefix(s, "-") {
			asc = false
			field = s[1:]
		}
		search.Sort(field, asc)
	}

	if len(params.Highlight) > 0 {
		hl := elastic.NewHighlight()
		for _, field := range params.Highlight {
			hl.Field(field)
		}
		search.Highlight(hl)
	}

	for name, agg := range params.Aggregations {
		search.Aggregation(name, agg)
	}

	result, err := search.Do(ctx)
	if err != nil {
		return nil, err
	}

	searchResult := &SearchResult{
		Total: result.TotalHits(),
		Hits:  make([]map[string]interface{}, 0, len(result.Hits.Hits)),
	}

	if result.Hits.MaxScore != nil {
		searchResult.MaxScore = *result.Hits.MaxScore
	}

	for _, hit := range result.Hits.Hits {
		var doc map[string]interface{}
		if err := json.Unmarshal(hit.Source, &doc); err != nil {
			continue
		}
		doc["_id"] = hit.Id
		if hit.Score != nil {
			doc["_score"] = *hit.Score
		}
		if hit.Highlight != nil {
			doc["_highlight"] = hit.Highlight
		}
		searchResult.Hits = append(searchResult.Hits, doc)
	}

	return searchResult, nil
}

func (e *ElasticSearch) SearchEmployees(ctx context.Context, query string, filters map[string]interface{}, page, size int) (*SearchResult, error) {
	return e.Search(ctx, "employees", SearchParams{
		Query:     query,
		Filters:   filters,
		From:      (page - 1) * size,
		Size:      size,
		Sort:      []string{"full_name.keyword"},
		Highlight: []string{"full_name", "email", "employee_code"},
	})
}

func (e *ElasticSearch) SearchAuditLogs(ctx context.Context, filters map[string]interface{}, page, size int) (*SearchResult, error) {
	return e.Search(ctx, "audit_logs", SearchParams{
		Filters: filters,
		From:    (page - 1) * size,
		Size:    size,
		Sort:    []string{"-created_at"},
	})
}

func (e *ElasticSearch) HealthCheck(ctx context.Context) error {
	_, _, err := e.client.Ping(e.client.String()).Do(ctx)
	return err
}

var employeesMapping = `{
	"settings": {
		"number_of_shards": 3,
		"number_of_replicas": 1,
		"analysis": {
			"analyzer": {
				"vietnamese": {
					"type": "custom",
					"tokenizer": "standard",
					"filter": ["lowercase", "asciifolding"]
				}
			}
		}
	},
	"mappings": {
		"properties": {
			"id": {"type": "keyword"},
			"employee_code": {"type": "keyword"},
			"full_name": {"type": "text", "analyzer": "vietnamese", "fields": {"keyword": {"type": "keyword"}}},
			"email": {"type": "keyword"},
			"phone": {"type": "keyword"},
			"department_id": {"type": "keyword"},
			"department_name": {"type": "text", "fields": {"keyword": {"type": "keyword"}}},
			"position_id": {"type": "keyword"},
			"position_name": {"type": "text", "fields": {"keyword": {"type": "keyword"}}},
			"employment_status": {"type": "keyword"},
			"employment_type": {"type": "keyword"},
			"join_date": {"type": "date"},
			"created_at": {"type": "date"},
			"updated_at": {"type": "date"}
		}
	}
}`

var departmentsMapping = `{
	"settings": {"number_of_shards": 1, "number_of_replicas": 1},
	"mappings": {
		"properties": {
			"id": {"type": "keyword"},
			"name": {"type": "text", "fields": {"keyword": {"type": "keyword"}}},
			"code": {"type": "keyword"},
			"parent_id": {"type": "keyword"},
			"manager_id": {"type": "keyword"},
			"status": {"type": "keyword"},
			"created_at": {"type": "date"}
		}
	}
}`

var attendancesMapping = `{
	"settings": {"number_of_shards": 3, "number_of_replicas": 1},
	"mappings": {
		"properties": {
			"id": {"type": "keyword"},
			"employee_id": {"type": "keyword"},
			"employee_name": {"type": "text"},
			"date": {"type": "date"},
			"check_in": {"type": "date"},
			"check_out": {"type": "date"},
			"status": {"type": "keyword"},
			"working_hours": {"type": "float"},
			"overtime_hours": {"type": "float"},
			"created_at": {"type": "date"}
		}
	}
}`

var payslipsMapping = `{
	"settings": {"number_of_shards": 3, "number_of_replicas": 1},
	"mappings": {
		"properties": {
			"id": {"type": "keyword"},
			"employee_id": {"type": "keyword"},
			"employee_name": {"type": "text"},
			"period_id": {"type": "keyword"},
			"year": {"type": "integer"},
			"month": {"type": "integer"},
			"gross_earnings": {"type": "float"},
			"total_deductions": {"type": "float"},
			"net_salary": {"type": "float"},
			"status": {"type": "keyword"},
			"created_at": {"type": "date"}
		}
	}
}`

var auditLogsMapping = `{
	"settings": {"number_of_shards": 3, "number_of_replicas": 1},
	"mappings": {
		"properties": {
			"id": {"type": "keyword"},
			"user_id": {"type": "keyword"},
			"action": {"type": "keyword"},
			"table_name": {"type": "keyword"},
			"record_id": {"type": "keyword"},
			"old_values": {"type": "object", "enabled": false},
			"new_values": {"type": "object", "enabled": false},
			"ip_address": {"type": "ip"},
			"user_agent": {"type": "text"},
			"created_at": {"type": "date"}
		}
	}
}`
