package utils

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"sync"
	"sync/atomic"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
)

// Exemption writes in the framework provider can remove SDK-owned tag
// configurations. Share invalidation, but keep response data per API client.
var metricTagReadGeneration atomic.Uint64

type tagIndexingRulesReadCache struct {
	mu         sync.Mutex
	generation uint64
	loaded     bool
	rules      []datadogV2.TagIndexingRuleData
	byID       map[string]datadogV2.TagIndexingRuleData
	err        error
}

type metricTagConfigurationsReadCache struct {
	mu         sync.Mutex
	generation uint64
	loaded     bool
	configs    map[string]*datadogV2.MetricTagConfiguration
}

// InvalidateMetricTagReadCaches also invalidates snapshots in the other
// implementation of the muxed provider, without sharing credentials or data.
func (i *ApiInstances) InvalidateMetricTagReadCaches() {
	metricTagReadGeneration.Add(1)
}

// ListTagIndexingRulesSnapshot returns a complete, ordered snapshot. The lock
// coalesces parallel refreshes; callers receive their own slice to sort safely.
func (i *ApiInstances) ListTagIndexingRulesSnapshot(ctx context.Context) ([]datadogV2.TagIndexingRuleData, error) {
	cache := &i.tagIndexingRulesReadCache
	cache.mu.Lock()
	defer cache.mu.Unlock()
	if err := i.loadTagIndexingRulesSnapshot(ctx, cache); err != nil {
		return nil, err
	}
	return append([]datadogV2.TagIndexingRuleData(nil), cache.rules...), nil
}

func (i *ApiInstances) loadTagIndexingRulesSnapshot(ctx context.Context, cache *tagIndexingRulesReadCache) error {
	generation := metricTagReadGeneration.Load()
	if !cache.loaded || cache.generation != generation {
		cache.rules, cache.err = i.listAllTagIndexingRules(ctx)
		cache.byID = make(map[string]datadogV2.TagIndexingRuleData, len(cache.rules))
		for _, rule := range cache.rules {
			cache.byID[rule.GetId()] = rule
		}
		cache.generation = generation
		cache.loaded = metricTagReadGeneration.Load() == generation
		if !cache.loaded {
			return fmt.Errorf("tag indexing rules changed while loading the snapshot")
		}
	}
	return cache.err
}

func (i *ApiInstances) listAllTagIndexingRules(ctx context.Context) ([]datadogV2.TagIndexingRuleData, error) {
	const pageLimit int64 = 1000
	var rules []datadogV2.TagIndexingRuleData
	seen := make(map[string]bool)
	total := int64(-1)
	for {
		options := datadogV2.NewListTagIndexingRulesOptionalParameters().WithPageLimit(pageLimit).WithPageOffset(int64(len(rules)))
		response, _, err := i.GetMetricsApiV2().ListTagIndexingRules(ctx, *options)
		if err != nil {
			return nil, err
		}
		if response.UnparsedObject != nil {
			return nil, fmt.Errorf("unparsed tag indexing rules response")
		}
		meta, ok := response.GetMetaOk()
		if !ok || meta.UnparsedObject != nil {
			return nil, fmt.Errorf("missing or unparsed tag indexing rules pagination metadata")
		}
		pageTotal, ok := meta.GetTotalOk()
		if !ok || *pageTotal < 0 || (total >= 0 && total != *pageTotal) {
			return nil, fmt.Errorf("inconsistent tag indexing rules total")
		}
		total = *pageTotal
		page, ok := response.GetDataOk()
		if !ok || (len(*page) == 0 && int64(len(rules)) < total) {
			return nil, fmt.Errorf("incomplete tag indexing rules page")
		}
		for _, rule := range *page {
			if err := CheckForUnparsed(rule); err != nil {
				return nil, fmt.Errorf("unparsed tag indexing rule in snapshot")
			}
			id := rule.GetId()
			attributes, ok := rule.GetAttributesOk()
			if id == "" || seen[id] || !ok {
				return nil, fmt.Errorf("missing or duplicate tag indexing rule identity")
			}
			if _, ok := attributes.GetRuleOrderOk(); !ok {
				return nil, fmt.Errorf("missing tag indexing rule order")
			}
			seen[id] = true
			rules = append(rules, rule)
		}
		if int64(len(rules)) > total {
			return nil, fmt.Errorf("tag indexing rules page exceeds reported total")
		}
		if int64(len(rules)) == total {
			break
		}
	}
	sort.Slice(rules, func(a, b int) bool {
		left, right := rules[a].GetAttributes(), rules[b].GetAttributes()
		return left.GetRuleOrder() < right.GetRuleOrder()
	})
	return rules, nil
}

// CachedTagIndexingRule falls back at the resource level on a snapshot error or
// miss. Only the individual endpoint can establish that a resource was deleted.
func (i *ApiInstances) CachedTagIndexingRule(ctx context.Context, id string) (*datadogV2.TagIndexingRuleResponse, bool) {
	cache := &i.tagIndexingRulesReadCache
	cache.mu.Lock()
	defer cache.mu.Unlock()
	if err := i.loadTagIndexingRulesSnapshot(ctx, cache); err != nil {
		return nil, false
	}
	if rule, ok := cache.byID[id]; ok {
		attributes := rule.GetAttributes()
		_, namePresent := attributes.GetNameOk()
		_, matchesPresent := attributes.GetMetricNameMatchesOk()
		_, tagsPresent := attributes.GetTagsOk()
		_, exclusionPresent := attributes.GetExcludeTagsModeOk()
		if !namePresent || !matchesPresent || !tagsPresent || !exclusionPresent {
			return nil, false
		}
		return &datadogV2.TagIndexingRuleResponse{Data: &rule}, true
	}
	return nil, false
}

// ReadTagIndexingRule uses a bulk snapshot when available and preserves the
// single-resource endpoint's errors and deletion semantics on a miss.
func (i *ApiInstances) ReadTagIndexingRule(ctx context.Context, id string) (datadogV2.TagIndexingRuleResponse, *http.Response, error) {
	if response, ok := i.CachedTagIndexingRule(ctx, id); ok {
		return *response, nil, nil
	}
	return i.GetMetricsApiV2().GetTagIndexingRule(ctx, id)
}

// CachedMetricTagConfiguration never infers deletion from absence: the list
// endpoint only includes metrics reporting within the selected time window.
func (i *ApiInstances) CachedMetricTagConfiguration(ctx context.Context, name string) (*datadogV2.MetricTagConfiguration, bool) {
	cache := &i.metricTagConfigurationsReadCache
	cache.mu.Lock()
	defer cache.mu.Unlock()
	generation := metricTagReadGeneration.Load()
	if !cache.loaded || cache.generation != generation {
		cache.configs = i.listMetricTagConfigurations(ctx)
		cache.generation = generation
		cache.loaded = metricTagReadGeneration.Load() == generation
		if !cache.loaded {
			return nil, false
		}
	}
	config, ok := cache.configs[name]
	return config, ok
}

func (i *ApiInstances) listMetricTagConfigurations(ctx context.Context) map[string]*datadogV2.MetricTagConfiguration {
	options := datadogV2.NewListTagConfigurationsOptionalParameters().WithFilterConfigured(true).WithWindowSeconds(604800).WithPageSize(10000)
	configs := make(map[string]*datadogV2.MetricTagConfiguration)
	seenCursors := make(map[string]bool)
	for {
		response, _, err := i.GetMetricsApiV2().ListTagConfigurations(ctx, *options)
		if err != nil || response.UnparsedObject != nil {
			return nil
		}
		page, ok := response.GetDataOk()
		if !ok {
			return nil
		}
		for _, item := range *page {
			// Configured metric types may outpace the pinned enum.
			// One unsupported metric must not disable batching for the entire page.
			if item.MetricTagConfiguration == nil || CheckForUnparsed(item) != nil {
				continue
			}
			config := item.MetricTagConfiguration
			attributes, ok := config.GetAttributesOk()
			if !ok || config.GetId() == "" {
				continue
			}
			_, tagsPresent := attributes.GetTagsOk()
			metricType, typePresent := attributes.GetMetricTypeOk()
			_, exclusionPresent := attributes.GetExcludeTagsModeOk()
			if !tagsPresent || !typePresent || !exclusionPresent || !metricType.IsValid() {
				continue
			}
			if *metricType == datadogV2.METRICTAGCONFIGURATIONMETRICTYPES_DISTRIBUTION {
				if _, ok := attributes.GetIncludePercentilesOk(); !ok {
					continue
				}
			}
			if _, duplicate := configs[config.GetId()]; duplicate {
				return nil
			}
			configs[config.GetId()] = config
		}
		meta, ok := response.GetMetaOk()
		if !ok || meta.UnparsedObject != nil {
			return nil
		}
		pagination, ok := meta.GetPaginationOk()
		if !ok || pagination.UnparsedObject != nil {
			return nil
		}
		next, ok := pagination.GetNextCursorOk()
		if !ok || next == nil || *next == "" {
			return configs
		}
		if seenCursors[*next] {
			return nil
		}
		seenCursors[*next] = true
		options.WithPageCursor(*next)
	}
}
