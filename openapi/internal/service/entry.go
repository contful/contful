// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/rs/zerolog/log"

	"github.com/contful/contful/openapi/internal/repository"
	"github.com/google/uuid"
)

// ErrContentSchemaNotFound 内容模型不存在
var ErrContentSchemaNotFound = errors.New("content schema not found")

// ErrEntryNotFound 条目不存在
var ErrEntryNotFound = errors.New("entry not found")

// EntryService Open API 内容读取服务
type EntryService struct {
	entryRepo  *repository.EntryRepository
	csRepo     *repository.ContentSchemaRepository
	cacheSvc   *CacheService
}

// NewEntryService 创建 EntryService
func NewEntryService(entryRepo *repository.EntryRepository, csRepo *repository.ContentSchemaRepository, cacheSvc *CacheService) *EntryService {
	return &EntryService{
		entryRepo: entryRepo,
		csRepo:    csRepo,
		cacheSvc:  cacheSvc,
	}
}

// EntryItem 对外输出的条目结构（扁平化字段值）
type EntryItem struct {
	ID             uuid.UUID              `json:"id"`
	Locale         string                 `json:"locale"`
	Version        int                    `json:"version"`
	PublishedTime  interface{}            `json:"published_time"`
	UpdatedTime    interface{}            `json:"updated_time"`
	Fields         map[string]interface{} `json:"fields"`
}

// EntryListResponse 条目列表响应
type EntryListResponse struct {
	Items    []EntryItem `json:"items"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

// ListBySlug 通过内容类型 slug 列出已发布条目
func (s *EntryService) ListBySlug(ctx context.Context, siteID uuid.UUID, slug string, locale string, sortField, sortOrder string, page, pageSize int) (*EntryListResponse, error) {
	// 1. 尝试从缓存获取
	cacheKey := s.cacheSvc.GetEntryListKey(siteID, slug, locale, sortField, sortOrder, page, pageSize)
	if cachedData, err := s.cacheSvc.Get(ctx, cacheKey); err == nil && cachedData != nil {
		var resp EntryListResponse
		if err := json.Unmarshal(cachedData, &resp); err == nil {
			log.Debug().Str("key", cacheKey).Msg("[Cache] Hit")
			return &resp, nil
		}
	}

	// 2. 通过 slug 找内容类型
	ct, err := s.csRepo.FindBySlug(ctx, siteID, slug)
	if err != nil {
		return nil, ErrContentSchemaNotFound
	}

	// 3. 查询已发布条目
	filter := repository.EntryListFilter{
		Locale:    locale,
		SortField: sortField,
		SortOrder: sortOrder,
	}
	entries, total, err := s.entryRepo.ListPublished(ctx, siteID, ct.ID, filter, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("query entries failed: %w", err)
	}

	// 4. 组装响应
	items := make([]EntryItem, len(entries))
	for i, e := range entries {
		items[i] = flattenEntry(e)
	}

	resp := &EntryListResponse{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}

	// 5. 写入缓存（跳过空结果，避免缓存穿透导致新发布内容不可见）
	if total > 0 {
		if err := s.cacheSvc.Set(ctx, cacheKey, resp); err != nil {
			log.Warn().Err(err).Msg("[Cache] Set failed")
		}
	}

	return resp, nil
}

// GetByID 获取单个已发布条目
func (s *EntryService) GetByID(ctx context.Context, siteID uuid.UUID, slug string, entryID uuid.UUID) (*EntryItem, error) {
	// 1. 尝试从缓存获取
	cacheKey := s.cacheSvc.GetEntryDetailKey(siteID, slug, entryID)
	if cachedData, err := s.cacheSvc.Get(ctx, cacheKey); err == nil && cachedData != nil {
		var item EntryItem
		if err := json.Unmarshal(cachedData, &item); err == nil {
			log.Debug().Str("key", cacheKey).Msg("[Cache] Hit")
			return &item, nil
		}
	}

	// 2. 验证 slug 对应的内容类型存在
	_, err := s.csRepo.FindBySlug(ctx, siteID, slug)
	if err != nil {
		return nil, ErrContentSchemaNotFound
	}

	entry, err := s.entryRepo.GetPublishedByID(ctx, siteID, entryID)
	if err != nil {
		return nil, ErrEntryNotFound
	}

	item := flattenEntry(*entry)

	// 3. 写入缓存
	if err := s.cacheSvc.Set(ctx, cacheKey, item); err != nil {
		log.Warn().Err(err).Msg("[Cache] Set failed")
	}

	return &item, nil
}

// flattenEntry 将条目的 Values 扁平化为 fields map
func flattenEntry(e repository.Entry) EntryItem {
	fields := make(map[string]interface{})
	for _, v := range e.Values {
		if v.Field.Name == "" {
			continue
		}
		fields[v.Field.Name] = extractValue(v.Value)
	}

	var publishedTime interface{}
	if e.PublishedTime != nil {
		publishedTime = e.PublishedTime.Format("2006-01-02T15:04:05Z07:00")
	}

	return EntryItem{
		ID:            e.ID,
		Locale:        e.Locale,
		Version:       e.Version,
		PublishedTime: publishedTime,
		UpdatedTime:   e.UpdatedTime.Format("2006-01-02T15:04:05Z07:00"),
		Fields:        fields,
	}
}

// extractValue 从 JSONB value 中提取实际值
// entry_values.value 存储为 {"value": <实际值>} 结构
func extractValue(v repository.JSONBValue) interface{} {
	if v == nil {
		return nil
	}
	// 如果有 "value" 包装层，解包
	if wrapped, ok := v["value"]; ok {
		return wrapped
	}
	// 如果是原始 JSON 对象（如 rich_text 存完整结构），直接返回
	// 尝试将字符串类型的 JSON 进一步解析
	if raw, ok := v["raw"]; ok {
		return raw
	}
	// 尝试将整个 map 作为字符串返回（兼容旧格式）
	if s, ok := v["string_value"]; ok {
		return s
	}
	// 如果只有一个键值对，取其值
	if len(v) == 1 {
		for _, val := range v {
			return val
		}
	}
	// 返回完整 map
	b, _ := json.Marshal(v)
	return string(b)
}

// ParsePage 解析分页参数，返回 page 和 pageSize
func ParsePage(pageStr, pageSizeStr string) (int, int) {
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(pageSizeStr)
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}
	return page, pageSize
}
