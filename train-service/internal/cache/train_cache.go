package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"train-service/internal/domain"
)

type TrainCache struct {
	rdb *redis.Client
	ttl time.Duration
}

func NewTrainCache(rdb *redis.Client) *TrainCache {
	return &TrainCache{rdb: rdb, ttl: 30 * time.Second}
}

type SearchResult struct {
	Trains []*domain.Train `json:"trains"`
	Total  int32           `json:"total"`
}

func searchKey(f *domain.SearchFilter) string {
	var after, before string
	if f.DepartureAfter != nil {
		after = f.DepartureAfter.UTC().Format(time.RFC3339)
	}
	if f.DepartureBefore != nil {
		before = f.DepartureBefore.UTC().Format(time.RFC3339)
	}
	return fmt.Sprintf("search:%s:%s:%s:%s:%d:%d", f.Origin, f.Destination, after, before, f.Page, f.PageSize)
}

func (c *TrainCache) GetSearch(ctx context.Context, f *domain.SearchFilter) (*SearchResult, bool) {
	if c == nil || c.rdb == nil {
		return nil, false
	}
	raw, err := c.rdb.Get(ctx, searchKey(f)).Bytes()
	if err != nil {
		return nil, false
	}
	var out SearchResult
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, false
	}
	return &out, true
}

func (c *TrainCache) SetSearch(ctx context.Context, f *domain.SearchFilter, res *SearchResult) {
	if c == nil || c.rdb == nil {
		return
	}
	raw, err := json.Marshal(res)
	if err != nil {
		return
	}
	_ = c.rdb.Set(ctx, searchKey(f), raw, c.ttl).Err()
}

func (c *TrainCache) InvalidateAll(ctx context.Context) {
	if c == nil || c.rdb == nil {
		return
	}
	keys, err := c.rdb.Keys(ctx, "search:*").Result()
	if err != nil || len(keys) == 0 {
		return
	}
	_ = c.rdb.Del(ctx, keys...).Err()
}
