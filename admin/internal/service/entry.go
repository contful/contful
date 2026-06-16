// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/contful/contful/admin/internal/model"
	"github.com/contful/contful/admin/internal/repository"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// EntryService 条目服务
type EntryService struct {
	entryRepo       *repository.EntryRepository
	contentSchemaRepo *repository.SchemaRepository
	fieldRepo       *repository.FieldRepository
}

// NewEntryService 新建服务
func NewEntryService(
	entryRepo *repository.EntryRepository,
	contentSchemaRepo *repository.SchemaRepository,
	fieldRepo *repository.FieldRepository,
) *EntryService {
	return &EntryService{
		entryRepo:       entryRepo,
		contentSchemaRepo: contentSchemaRepo,
		fieldRepo:       fieldRepo,
	}
}

// Create 创建条目
func (s *EntryService) Create(ctx context.Context, siteID uuid.UUID, userID *uuid.UUID, req *model.EntryCreate, integritySvc *IntegrityService) (*model.Entry, error) {
	// 验证内容模型存在
	contentSchema, err := s.contentSchemaRepo.GetByIDWithFields(ctx, req.ContentSchemaID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrContentSchemaNotFound
		}
		return nil, err
	}

	// 设置默认值
	locale := req.Locale
	if locale == "" {
		locale = "zh-CN"
	}

	// 创建条目
	entry := &model.Entry{
		ID:             uuid.New(),
		ContentSchemaID:  req.ContentSchemaID,
		SiteID:         siteID,
		Locale:         locale,
		Status:         model.EntryStatusDraft,
		Version:        1,
		VersionHistory: model.JSONArray{},
		SortWeight:     req.SortWeight,
		CreatedBy:      userID,
		ContentSchema:    contentSchema,
		ScheduledPublishTime:   req.ScheduledPublishTime,
		ScheduledUnpublishTime: req.ScheduledUnpublishTime,
	}

	// 设置 SEO
	if req.SEOTitle != "" {
		entry.SEOTitle = req.SEOTitle
	}
	if req.SEODescription != "" {
		entry.SEODescription = req.SEODescription
	}
	if len(req.SEOKeywords) > 0 {
		entry.SEOKeywords = req.SEOKeywords
	}

	// 事务: 创建条目 + 字段值 + 签名
	err = s.entryRepo.WithTransaction(func(txRepo *repository.EntryRepository) error {
		// 创建条目
		if err := txRepo.Create(ctx, entry); err != nil {
			return err
		}

		// 解析并验证字段值
		if len(req.Values) > 0 {
			values, err := s.parseAndValidateValues(ctx, contentSchema.Fields, entry.ID, req.Values)
			if err != nil {
				return err
			}
			if err := txRepo.CreateValues(ctx, values); err != nil {
				return err
			}
			entry.Values = values
		}

		// 数据签名
		if integritySvc != nil {
			if err := integritySvc.SignEntry(entry, entry.Values); err != nil {
				return err
			}
			if err := txRepo.Update(ctx, entry); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return entry, nil
}

// GetByID 获取条目
func (s *EntryService) GetByID(ctx context.Context, siteID uuid.UUID, id uuid.UUID) (*model.Entry, error) {
	entry, err := s.entryRepo.GetByIDWithType(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrEntryNotFound
		}
		return nil, err
	}

	// 验证站点权限
	if entry.SiteID != siteID {
		return nil, ErrEntryNotFound
	}

	return entry, nil
}

// List 列出条目
func (s *EntryService) List(ctx context.Context, siteID uuid.UUID, filter *model.EntryListFilter, page, pageSize int) ([]model.Entry, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var entries []model.Entry
	var total int64
	var err error

	if filter != nil && filter.ContentSchemaID != nil {
		entries, total, err = s.entryRepo.ListByContentSchema(ctx, siteID, *filter.ContentSchemaID, filter, page, pageSize)
	} else {
		entries, total, err = s.entryRepo.ListBySite(ctx, siteID, filter, page, pageSize)
	}

	return entries, total, err
}

// Update 更新条目
func (s *EntryService) Update(ctx context.Context, siteID uuid.UUID, userID *uuid.UUID, id uuid.UUID, req *model.EntryUpdate, integritySvc *IntegrityService) (*model.Entry, error) {
	entry, err := s.entryRepo.GetByIDWithValues(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrEntryNotFound
		}
		return nil, err
	}

	// 验证站点权限
	if entry.SiteID != siteID {
		return nil, ErrEntryNotFound
	}

	// 获取内容模型
	contentSchema, err := s.contentSchemaRepo.GetByIDWithFields(ctx, entry.ContentSchemaID)
	if err != nil {
		return nil, err
	}

	// 更新基础字段
	if req.Locale != nil {
		entry.Locale = *req.Locale
	}
	if req.Status != nil {
		entry.Status = *req.Status
	}
	if req.SEOTitle != nil {
		entry.SEOTitle = *req.SEOTitle
	}
	if req.SEODescription != nil {
		entry.SEODescription = *req.SEODescription
	}
	if req.SEOKeywords != nil {
		entry.SEOKeywords = req.SEOKeywords
	}
	if req.SortWeight != nil {
		entry.SortWeight = *req.SortWeight
	}
	if req.ScheduledPublishTime != nil {
		entry.ScheduledPublishTime = req.ScheduledPublishTime
	}
	if req.ScheduledUnpublishTime != nil {
		entry.ScheduledUnpublishTime = req.ScheduledUnpublishTime
	}

	// 事务: 更新条目 + 字段值 + 重新签名
	err = s.entryRepo.WithTransaction(func(txRepo *repository.EntryRepository) error {
		// 更新条目
		if err := txRepo.Update(ctx, entry); err != nil {
			return err
		}

		// 更新字段值
		if len(req.Values) > 0 {
			// 删除旧值
			if err := txRepo.DeleteValues(ctx, entry.ID); err != nil {
				return err
			}

			// 解析并验证新值
			values, err := s.parseAndValidateValues(ctx, contentSchema.Fields, entry.ID, req.Values)
			if err != nil {
				return err
			}
			if err := txRepo.CreateValues(ctx, values); err != nil {
				return err
			}
			entry.Values = values
		}

		// 重新签名
		if integritySvc != nil {
			if err := integritySvc.SignEntry(entry, entry.Values); err != nil {
				return err
			}
			if err := txRepo.Update(ctx, entry); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return entry, nil
}

// Delete 删除条目
func (s *EntryService) Delete(ctx context.Context, siteID uuid.UUID, id uuid.UUID) error {
	entry, err := s.entryRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrEntryNotFound
		}
		return err
	}

	// 验证站点权限
	if entry.SiteID != siteID {
		return ErrEntryNotFound
	}

	return s.entryRepo.Delete(ctx, id)
}

// Publish 发布条目
func (s *EntryService) Publish(ctx context.Context, siteID uuid.UUID, userID *uuid.UUID, id uuid.UUID, req *model.EntryPublish, integritySvc *IntegrityService) (*model.Entry, error) {
	entry, err := s.entryRepo.GetByIDWithValues(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrEntryNotFound
		}
		return nil, err
	}

	// 验证站点权限
	if entry.SiteID != siteID {
		return nil, ErrEntryNotFound
	}

	// 获取内容模型（用于签名检查）
	contentSchema, err := s.contentSchemaRepo.GetByIDWithFields(ctx, entry.ContentSchemaID)
	if err != nil {
		return nil, err
	}
	entry.ContentSchema = contentSchema

	// 更新版本历史
	now := time.Now()
	historyEntry := map[string]interface{}{
		"version":        entry.Version,
		"created_time":     now.Format(time.RFC3339),
		"change_summary": req.ChangeSummary,
	}
	entry.VersionHistory = append(entry.VersionHistory, historyEntry)

	// 更新发布状态
	entry.Status = model.EntryStatusPublished
	entry.Version++
	entry.PublishedTime = &now
	entry.PublishedBy = userID

	if err := s.entryRepo.Update(ctx, entry); err != nil {
		return nil, err
	}

	// 发布时重新签名
	if integritySvc != nil {
		if err := integritySvc.SignEntry(entry, entry.Values); err != nil {
			log.Warn().Err(err).Interface("entry_id", entry.ID).Msg("failed to sign entry")
		}
		if err := s.entryRepo.Update(ctx, entry); err != nil {
			log.Warn().Err(err).Interface("entry_id", entry.ID).Msg("failed to update signed entry")
		}
	}

	return entry, nil
}

// Unpublish 取消发布
func (s *EntryService) Unpublish(ctx context.Context, siteID uuid.UUID, id uuid.UUID) (*model.Entry, error) {
	entry, err := s.entryRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrEntryNotFound
		}
		return nil, err
	}

	// 验证站点权限
	if entry.SiteID != siteID {
		return nil, ErrEntryNotFound
	}

	entry.Status = model.EntryStatusDraft
	entry.PublishedTime = nil
	entry.PublishedBy = nil

	if err := s.entryRepo.Update(ctx, entry); err != nil {
		return nil, err
	}

	return entry, nil
}

// GetVersions 获取版本历史
func (s *EntryService) GetVersions(ctx context.Context, siteID uuid.UUID, id uuid.UUID) ([]model.EntryVersion, error) {
	entry, err := s.entryRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrEntryNotFound
		}
		return nil, err
	}

	// 验证站点权限
	if entry.SiteID != siteID {
		return nil, ErrEntryNotFound
	}

	return s.entryRepo.GetVersions(ctx, id)
}

// SetSchedule 设置排期时间
func (s *EntryService) SetSchedule(ctx context.Context, siteID uuid.UUID, id uuid.UUID, req *model.ScheduleRequest) (*model.Entry, error) {
	entry, err := s.entryRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrEntryNotFound
		}
		return nil, err
	}

	if entry.SiteID != siteID {
		return nil, ErrEntryNotFound
	}

	entry.ScheduledPublishTime = req.ScheduledPublishTime
	entry.ScheduledUnpublishTime = req.ScheduledUnpublishTime

	if err := s.entryRepo.Update(ctx, entry); err != nil {
		return nil, err
	}
	return entry, nil
}

// ClearSchedule 清除排期时间
func (s *EntryService) ClearSchedule(ctx context.Context, siteID uuid.UUID, id uuid.UUID) (*model.Entry, error) {
	entry, err := s.entryRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrEntryNotFound
		}
		return nil, err
	}

	if entry.SiteID != siteID {
		return nil, ErrEntryNotFound
	}

	entry.ScheduledPublishTime = nil
	entry.ScheduledUnpublishTime = nil

	if err := s.entryRepo.Update(ctx, entry); err != nil {
		return nil, err
	}
	return entry, nil
}

// ============ 批量操作 ============

// BatchDelete 批量删除
func (s *EntryService) BatchDelete(ctx context.Context, siteID uuid.UUID, ids []uuid.UUID) (*model.BatchResponse, error) {
	// 验证站点权限
	var validIDs []uuid.UUID
	for _, id := range ids {
		entry, err := s.entryRepo.GetByID(ctx, id)
		if err != nil {
			continue // 跳过不存在的
		}
		if entry.SiteID == siteID {
			validIDs = append(validIDs, id)
		}
	}

	if len(validIDs) == 0 {
		return &model.BatchResponse{SuccessCount: 0, FailedCount: len(ids)}, nil
	}

	count, err := s.entryRepo.BatchDelete(ctx, validIDs)
	if err != nil {
		return nil, err
	}

	return &model.BatchResponse{
		SuccessCount: int(count),
		FailedCount:  len(ids) - int(count),
	}, nil
}

// BatchPublish 批量发布
func (s *EntryService) BatchPublish(ctx context.Context, siteID uuid.UUID, ids []uuid.UUID) (*model.BatchResponse, error) {
	// 验证站点权限
	var validIDs []uuid.UUID
	for _, id := range ids {
		entry, err := s.entryRepo.GetByID(ctx, id)
		if err != nil {
			continue
		}
		if entry.SiteID == siteID {
			validIDs = append(validIDs, id)
		}
	}

	if len(validIDs) == 0 {
		return &model.BatchResponse{SuccessCount: 0, FailedCount: len(ids)}, nil
	}

	count, err := s.entryRepo.BatchPublish(ctx, validIDs)
	if err != nil {
		return nil, err
	}

	return &model.BatchResponse{
		SuccessCount: int(count),
		FailedCount:  len(ids) - int(count),
	}, nil
}

// BatchUnpublish 批量取消发布
func (s *EntryService) BatchUnpublish(ctx context.Context, siteID uuid.UUID, ids []uuid.UUID) (*model.BatchResponse, error) {
	// 验证站点权限
	var validIDs []uuid.UUID
	for _, id := range ids {
		entry, err := s.entryRepo.GetByID(ctx, id)
		if err != nil {
			continue
		}
		if entry.SiteID == siteID {
			validIDs = append(validIDs, id)
		}
	}

	if len(validIDs) == 0 {
		return &model.BatchResponse{SuccessCount: 0, FailedCount: len(ids)}, nil
	}

	count, err := s.entryRepo.BatchUnpublish(ctx, validIDs)
	if err != nil {
		return nil, err
	}

	return &model.BatchResponse{
		SuccessCount: int(count),
		FailedCount:  len(ids) - int(count),
	}, nil
}

// ============ 辅助方法 ============

// parseAndValidateValues 解析并验证字段值
func (s *EntryService) parseAndValidateValues(ctx context.Context, fields []model.Field, entryID uuid.UUID, values map[string]interface{}) ([]model.EntryValue, error) {
	result := make([]model.EntryValue, 0, len(values))

	// 构建字段映射
	fieldMap := make(map[string]model.Field)
	for _, f := range fields {
		fieldMap[f.Name] = f
	}

	for name, value := range values {
		field, ok := fieldMap[name]
		if !ok {
			continue // 忽略未知字段
		}

		entryValue := model.EntryValue{
			ID:      uuid.New(),
			EntryID: entryID,
			FieldID: field.ID,
		}

		// 设置 JSONB 值：直接存储原始值，不再包装为 { "value": ... }
		if value == nil {
			entryValue.Value = nil
		} else if str, ok := value.(string); ok {
			// 如果是 JSON 字符串，尝试解析后直接存储
			var jsonValue map[string]interface{}
			if err := json.Unmarshal([]byte(str), &jsonValue); err == nil {
				entryValue.Value = jsonValue
			} else {
				// 无法解析为 JSON，包装为 JSONB
				entryValue.Value = map[string]interface{}{"value": str}
			}
		} else {
			// number/bool/array/object 等非 string 类型，包装为 JSONB
			if v, ok := value.(map[string]interface{}); ok {
				entryValue.Value = v
			} else {
				entryValue.Value = map[string]interface{}{"value": value}
			}
		}

		// 设置辅助列（用于索引和查询）
		s.setAuxiliaryValue(&entryValue, &field, value)

		result = append(result, entryValue)
	}

	return result, nil
}

// setAuxiliaryValue 设置辅助列值
func (s *EntryService) setAuxiliaryValue(entryValue *model.EntryValue, field *model.Field, value interface{}) {
	if value == nil {
		return
	}

	switch field.FieldType {
	case "text", "email", "url", "rich_text":
		if str, ok := value.(string); ok {
			entryValue.TextValue = &str
		}
	case "number":
		switch v := value.(type) {
		case float64:
			entryValue.NumberValue = &v
		case int:
			f := float64(v)
			entryValue.NumberValue = &f
		}
	case "boolean":
		if b, ok := value.(bool); ok {
			entryValue.BoolValue = &b
		}
	case "date":
		if str, ok := value.(string); ok {
			if t, err := time.Parse("2006-01-02", str); err == nil {
				entryValue.DateValue = &t
			}
		}
	case "datetime":
		if str, ok := value.(string); ok {
			if t, err := time.Parse(time.RFC3339, str); err == nil {
				entryValue.DatetimeValue = &t
			}
		}
	}
}

// 错误定义
var (
	ErrEntryNotFound = errors.New("entry not found")
	// ErrContentSchemaNotFound 定义在 content_type.go 中
)
