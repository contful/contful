// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// CacheTTL 内容缓存 TTL（30 分钟）
// 平衡数据新鲜度和数据库压力
const CacheTTL = 30 * time.Minute

// CacheKeyPrefix 缓存键前缀（Admin 与 OpenAPI 共用规范）
// 格式: contful:content:{siteID}:{slug}:{locale}:{sort}:{order}:{page}:{size}
//       contful:content:{siteID}:{slug}:{entryID}
// 列表与详情通过冒号分隔的段数区分（列表 8 段 vs 详情 6 段）
const CacheKeyPrefix = "contful:content:"

// CacheService Redis 缓存服务
type CacheService struct {
	rdb *redis.Client
}

// NewCacheService 创建 CacheService
func NewCacheService(rdb *redis.Client) *CacheService {
	return &CacheService{rdb: rdb}
}

// GetEntryListKey 生成内容列表缓存键
// 格式: contful:content:{siteID}:{slug}:{locale}:{sort}:{order}:{page}:{size}
func (s *CacheService) GetEntryListKey(siteID uuid.UUID, slug, locale, sortField, sortOrder string, page, pageSize int) string {
	return fmt.Sprintf("%s%s:%s:%s:%s:%s:%d:%d",
		CacheKeyPrefix, siteID.String(), slug, locale, sortField, sortOrder, page, pageSize)
}

// GetEntryDetailKey 生成内容详情缓存键
// 格式: contful:content:{siteID}:{slug}:{entryID}
func (s *CacheService) GetEntryDetailKey(siteID uuid.UUID, slug string, entryID uuid.UUID) string {
	return fmt.Sprintf("%s%s:%s:%s", CacheKeyPrefix, siteID.String(), slug, entryID.String())
}

// GetSitePattern 生成站点内容缓存匹配模式（用于清除该站点所有内容缓存）
func (s *CacheService) GetSitePattern(siteID uuid.UUID) string {
	return fmt.Sprintf("%s%s:*", CacheKeyPrefix, siteID.String())
}

// Get 获取缓存
func (s *CacheService) Get(ctx context.Context, key string) ([]byte, error) {
	data, err := s.rdb.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil // 缓存未命中
	}
	if err != nil {
		return nil, fmt.Errorf("redis get failed: %w", err)
	}
	return data, nil
}

// Set 设置缓存
func (s *CacheService) Set(ctx context.Context, key string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("json marshal failed: %w", err)
	}
	if err := s.rdb.Set(ctx, key, jsonData, CacheTTL).Err(); err != nil {
		return fmt.Errorf("redis set failed: %w", err)
	}
	return nil
}

// Delete 删除指定缓存
func (s *CacheService) Delete(ctx context.Context, key string) error {
	if err := s.rdb.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("redis del failed: %w", err)
	}
	return nil
}

// InvalidateSite 清除指定站点的所有内容缓存
// 当内容发布/取消发布时调用
func (s *CacheService) InvalidateSite(ctx context.Context, siteID uuid.UUID) (int64, error) {
	pattern := s.GetSitePattern(siteID)
	
	var cursor uint64
	var deletedCount int64
	
	for {
		keys, nextCursor, err := s.rdb.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return deletedCount, fmt.Errorf("redis scan failed: %w", err)
		}
		
		if len(keys) > 0 {
			deleted, err := s.rdb.Del(ctx, keys...).Result()
			if err != nil {
				return deletedCount, fmt.Errorf("redis del failed: %w", err)
			}
			deletedCount += deleted
		}
		
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	
	return deletedCount, nil
}

// InvalidateContentSchema 清除指定内容模型的所有缓存
func (s *CacheService) InvalidateContentSchema(ctx context.Context, siteID uuid.UUID, slug string) (int64, error) {
	pattern := fmt.Sprintf("%s%s:%s:*", CacheKeyPrefix, siteID.String(), slug)
	
	var cursor uint64
	var deletedCount int64
	
	for {
		keys, nextCursor, err := s.rdb.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return deletedCount, fmt.Errorf("redis scan failed: %w", err)
		}
		
		if len(keys) > 0 {
			deleted, err := s.rdb.Del(ctx, keys...).Result()
			if err != nil {
				return deletedCount, fmt.Errorf("redis del failed: %w", err)
			}
			deletedCount += deleted
		}
		
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	
	return deletedCount, nil
}
