// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/contful/contful/admin/internal/model"
	"github.com/contful/contful/admin/internal/repository"
	"github.com/rs/zerolog/log"
	"github.com/contful/contful/admin/internal/storage"

	"github.com/google/uuid"
)

// 常见 MIME 类型
var mimeTypeMap = map[string]model.AssetType{
	"image/jpeg":              model.AssetTypeImage,
	"image/png":               model.AssetTypeImage,
	"image/gif":               model.AssetTypeImage,
	"image/webp":              model.AssetTypeImage,
	"image/svg+xml":           model.AssetTypeImage,
	"image/bmp":               model.AssetTypeImage,
	"image/tiff":              model.AssetTypeImage,
	"video/mp4":               model.AssetTypeVideo,
	"video/mpeg":              model.AssetTypeVideo,
	"video/webm":              model.AssetTypeVideo,
	"video/quicktime":        model.AssetTypeVideo,
	"video/x-msvideo":         model.AssetTypeVideo,
	"audio/mpeg":              model.AssetTypeAudio,
	"audio/wav":               model.AssetTypeAudio,
	"audio/ogg":               model.AssetTypeAudio,
	"audio/webm":             model.AssetTypeAudio,
	"audio/aac":               model.AssetTypeAudio,
	"application/pdf":         model.AssetTypeDocument,
	"application/msword":      model.AssetTypeDocument,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": model.AssetTypeDocument,
	"application/vnd.ms-excel": model.AssetTypeDocument,
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": model.AssetTypeDocument,
	"text/plain":              model.AssetTypeFile,
	"text/csv":                model.AssetTypeFile,
	"application/json":        model.AssetTypeFile,
	"application/zip":         model.AssetTypeFile,
	"application/x-rar-compressed": model.AssetTypeFile,
}

// 图片扩展名
var imageExtensions = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true,
	".webp": true, ".svg": true, ".bmp": true, ".tiff": true,
}

// 视频扩展名
var videoExtensions = map[string]bool{
	".mp4": true, ".mpeg": true, ".webm": true, ".mov": true, ".avi": true,
}

// 音频扩展名
var audioExtensions = map[string]bool{
	".mp3": true, ".wav": true, ".ogg": true, ".webm": true, ".aac": true, ".m4a": true,
}

// AssetService 资源服务
type AssetService struct {
	assetRepo       *repository.AssetRepository
	storageProvider storage.StorageProvider
	configService   *ConfigService // 可选，用于数据签名
}

// NewAssetService 新建服务（注入全局 StorageProvider）
func NewAssetService(assetRepo *repository.AssetRepository, storageProvider storage.StorageProvider) *AssetService {
	return &AssetService{
		assetRepo:       assetRepo,
		storageProvider: storageProvider,
	}
}

// SetConfigService 设置配置服务（用于数据签名）
func (s *AssetService) SetConfigService(cs *ConfigService) {
	s.configService = cs
}

// ServeFile 提供静态文件服务
func (s *AssetService) ServeFile(ctx context.Context, siteID uuid.UUID, key string) (io.ReadCloser, string, error) {
	// 调用存储驱动的 ServeFile 方法
	if localProvider, ok := s.storageProvider.(*storage.LocalProvider); ok {
		return localProvider.ServeFile(ctx, key)
	}

	// 对于其他存储驱动，使用 Download 方法
	reader, err := s.storageProvider.Download(ctx, key, nil)
	if err != nil {
		return nil, "", err
	}

	// 尝试从 key 推断 MIME 类型
	ext := filepath.Ext(key)
	mimeType := "application/octet-stream"
	if ext != "" {
		// 简单的 MIME 类型映射
		switch strings.ToLower(ext) {
		case ".jpg", ".jpeg":
			mimeType = "image/jpeg"
		case ".png":
			mimeType = "image/png"
		case ".gif":
			mimeType = "image/gif"
		case ".webp":
			mimeType = "image/webp"
		case ".svg":
			mimeType = "image/svg+xml"
		case ".pdf":
			mimeType = "application/pdf"
		case ".mp4":
			mimeType = "video/mp4"
		case ".mp3":
			mimeType = "audio/mpeg"
		}
	}

	return reader, mimeType, nil
}


// Upload 上传资源
func (s *AssetService) Upload(ctx context.Context, siteID, userID uuid.UUID, file *multipart.FileHeader, folderID *uuid.UUID, alt, title string) (*model.Asset, error) {
	// 打开文件
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}
	defer src.Close()

	// 读取文件内容
	data, err := io.ReadAll(src)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}

	// 计算哈希
	hash := sha256.Sum256(data)
	fileHash := hex.EncodeToString(hash[:])

	// 检查是否已存在相同文件
	existing, err := s.assetRepo.GetByFileHash(ctx, siteID, fileHash)
	if err == nil && existing != nil {
		// 返回已存在的文件
		return existing, nil
	}

	// 确定文件类型
	ext := strings.ToLower(filepath.Ext(file.Filename))
	mimeType := file.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	assetType := s.getAssetType(ext, mimeType)
	if assetType == "" {
		return nil, errors.New("不支持的文件类型")
	}

	// 生成唯一文件名
	filename := fmt.Sprintf("%s_%d%s", uuid.New().String()[:8], time.Now().Unix(), ext)

	// 生成存储 key: {site_id}/{year}/{month}/{day}/{filename}
	now := time.Now()
	storageKey := filepath.Join(
		siteID.String(),
		fmt.Sprintf("%d", now.Year()),
		fmt.Sprintf("%02d", now.Month()),
		fmt.Sprintf("%02d", now.Day()),
		filename,
	)

	// 上传到存储驱动
	objectInfo, err := s.storageProvider.Upload(ctx, storageKey, bytes.NewReader(data), file.Size, &storage.WriteOptions{
		ContentType: mimeType,
	})
	if err != nil {
		return nil, fmt.Errorf("上传文件失败: %w", err)
	}

	// 创建资源记录
	asset := &model.Asset{
		ID:            uuid.New(),
		SiteID:        siteID,
		FolderID:      folderID,
		UUID:          uuid.New().String(),
		Name:          s.generateSlug(filepath.Base(file.Filename), ext),
		OriginalName:  file.Filename,
		Slug:          s.generateSlug(filepath.Base(file.Filename), ext),
		Type:          assetType,
		MimeType:      mimeType,
		Extension:     strings.TrimPrefix(ext, "."),
		Size:          file.Size,
		Path:          storageKey,
		URL:           objectInfo.URL,
		Alt:           alt,
		Title:         title,
		Visibility:    model.AssetVisibilityPrivate,
		FileHash:      fileHash,
		Disk:          s.storageProvider.Name(),
		CreatedBy:     &userID,
	}

	// 如果是图片，尝试获取尺寸
	if assetType == model.AssetTypeImage && imageExtensions[ext] {
		// 简单的 PNG/JPEG 尺寸检测（实际项目中建议使用 imaging 库）
		if width, height, ok := s.getImageDimensions(data, ext); ok {
			asset.Width = &width
			asset.Height = &height
		}
		// 生成缩略图 URL（实际项目中需要实际生成缩略图）
		thumbnailURL := strings.Replace(objectInfo.URL, filepath.Ext(objectInfo.URL), "_thumb"+ext, 1)
		asset.ThumbnailURL = &thumbnailURL
	}

	// 保存到数据库
	if err := s.assetRepo.Create(ctx, asset); err != nil {
		// 删除已上传的文件（回滚）
		if delErr := s.storageProvider.Delete(ctx, storageKey); delErr != nil {
			log.Warn().Err(delErr).Str("key", storageKey).Msg("failed to rollback uploaded file")
		}
		return nil, fmt.Errorf("保存资源记录失败: %w", err)
	}

	// 数据签名（仅当签名密钥已配置时）
	if s.configService != nil {
		signingKey, _ := s.configService.GetAuditSigningKey()
		alg := "HMAC-SHA256" // 默认算法
		intSvc, _ := NewIntegrityService(siteID, signingKey, alg)
		if intSvc != nil && intSvc.IsEnabled() {
			if err := intSvc.SignAsset(asset); err != nil {
				log.Warn().Err(err).Interface("asset_id", asset.ID).Msg("failed to sign asset")
			}
			if err := s.assetRepo.Update(ctx, asset); err != nil {
				log.Warn().Err(err).Interface("asset_id", asset.ID).Msg("failed to update signed asset")
			}
		}
	}

	return asset, nil
}

// getAssetType 根据扩展名和 MIME 类型获取资源类型
func (s *AssetService) getAssetType(ext, mimeType string) model.AssetType {
	// 先检查 MIME 类型
	if t, ok := mimeTypeMap[mimeType]; ok {
		return t
	}

	// 再检查扩展名
	ext = strings.ToLower(ext)
	if imageExtensions[ext] {
		return model.AssetTypeImage
	}
	if videoExtensions[ext] {
		return model.AssetTypeVideo
	}
	if audioExtensions[ext] {
		return model.AssetTypeAudio
	}

	// 检查是否为文档
	if ext == ".pdf" || ext == ".doc" || ext == ".docx" || ext == ".xls" || ext == ".xlsx" {
		return model.AssetTypeDocument
	}

	return model.AssetTypeFile
}

// generateSlug 生成 slug
func (s *AssetService) generateSlug(name, ext string) string {
	// 去除扩展名
	name = strings.TrimSuffix(name, ext)

	// 转换为小写
	name = strings.ToLower(name)

	// 替换特殊字符为空格
	reg := regexp.MustCompile(fmt.Sprintf("[^a-z0-9\u4e00-\u9fa5]+"))
	name = reg.ReplaceAllString(name, "-")

	// 去除首尾的连字符
	name = strings.Trim(name, "-")

	// 如果为空，生成随机名称
	if name == "" {
		name = fmt.Sprintf("file-%s", uuid.New().String()[:8])
	}

	return name
}

// getImageDimensions 获取图片尺寸（简化实现）
func (s *AssetService) getImageDimensions(data []byte, ext string) (width, height int, ok bool) {
	// 这里使用简化实现，实际项目中建议使用 imaging 库
	// PNG: 前 24 字节包含 IHDR
	if ext == ".png" && len(data) > 24 {
		width = int(data[16])<<24 | int(data[17])<<16 | int(data[18])<<8 | int(data[19])
		height = int(data[20])<<24 | int(data[21])<<16 | int(data[22])<<8 | int(data[23])
		return width, height, width > 0 && height > 0 && width < 65536 && height < 65536
	}

	// JPEG: 读取 SOF0 段
	if (ext == ".jpg" || ext == ".jpeg") && len(data) > 2 {
		// 简化实现：检查 JPEG 标记
		if data[0] == 0xFF && data[1] == 0xD8 {
			// 查找 SOF0 标记 (0xFF 0xC0)
			for i := 2; i < len(data)-9; i++ {
				if data[i] == 0xFF && data[i+1] == 0xC0 {
					height = int(data[i+5])<<8 | int(data[i+6])
					width = int(data[i+7])<<8 | int(data[i+8])
					return width, height, width > 0 && height > 0
				}
			}
		}
	}

	return 0, 0, false
}

// Get 获取资源
func (s *AssetService) Get(ctx context.Context, id uuid.UUID) (*model.Asset, error) {
	return s.assetRepo.GetByID(ctx, id)
}

// List 列出资源
func (s *AssetService) List(ctx context.Context, siteID uuid.UUID, filter *model.AssetListFilter, page, pageSize int) (*model.AssetListResponse, error) {
	assets, total, err := s.assetRepo.List(ctx, siteID, filter, page, pageSize)
	if err != nil {
		return nil, err
	}

	items := make([]model.AssetResponse, len(assets))
	for i, asset := range assets {
		items[i] = asset.ToResponse()
	}

	return &model.AssetListResponse{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// Update 更新资源
func (s *AssetService) Update(ctx context.Context, id uuid.UUID, req *model.AssetUpdate) (*model.Asset, error) {
	asset, err := s.assetRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.FolderID != nil {
		asset.FolderID = req.FolderID
	}
	if req.Name != nil {
		asset.Name = *req.Name
	}
	if req.Alt != nil {
		asset.Alt = *req.Alt
	}
	if req.Title != nil {
		asset.Title = *req.Title
	}
	if req.Caption != nil {
		asset.Caption = *req.Caption
	}
	if req.AltText != nil {
		asset.AltText = *req.AltText
	}
	if req.Description != nil {
		asset.Description = *req.Description
	}
	if req.Tags != nil {
		asset.Tags = req.Tags
	}
	if req.Visibility != nil {
		asset.Visibility = *req.Visibility
	}

	if err := s.assetRepo.Update(ctx, asset); err != nil {
		return nil, err
	}

	return asset, nil
}

// Delete 删除资源
func (s *AssetService) Delete(ctx context.Context, id uuid.UUID) error {
	asset, err := s.assetRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// 删除存储文件（不阻塞 DB 删除）
	_ = s.storageProvider.Delete(ctx, asset.Path)

	// 删除数据库记录
	return s.assetRepo.Delete(ctx, id)
}

// BatchDelete 批量删除
func (s *AssetService) BatchDelete(ctx context.Context, ids []uuid.UUID) error {
	// 先查所有资产获取路径
	assets, _ := s.assetRepo.GetByIDs(ctx, ids)
	var allPaths []string
	for _, a := range assets {
		allPaths = append(allPaths, a.Path)
	}
	if len(allPaths) > 0 {
		_ = s.storageProvider.DeleteMulti(ctx, allPaths)
	}

	return s.assetRepo.BatchDelete(ctx, ids)
}

// ============ Folder 操作 ============

// CreateFolder 创建文件夹
func (s *AssetService) CreateFolder(ctx context.Context, siteID, userID uuid.UUID, req *model.FolderCreate) (*model.AssetFolder, error) {
	slug := s.generateSlug(req.Name, "")

	// 构建路径
	var path string
	if req.ParentID != nil {
		parent, err := s.assetRepo.GetFolderByID(ctx, *req.ParentID)
		if err != nil {
			return nil, fmt.Errorf("父文件夹不存在: %w", err)
		}
		path = filepath.Join(parent.Path, slug)
	} else {
		path = "/" + slug
	}

	folder := &model.AssetFolder{
		ID:        uuid.New(),
		SiteID:    siteID,
		ParentID:  req.ParentID,
		Name:      req.Name,
		Slug:      slug,
		Path:      path,
		SortOrder: req.SortOrder,
		CreatedBy: &userID,
	}

	if err := s.assetRepo.CreateFolder(ctx, folder); err != nil {
		return nil, err
	}

	return folder, nil
}

// GetFolder 获取文件夹
func (s *AssetService) GetFolder(ctx context.Context, id uuid.UUID) (*model.AssetFolder, error) {
	return s.assetRepo.GetFolderByID(ctx, id)
}

// ListFolders 列出文件夹
func (s *AssetService) ListFolders(ctx context.Context, siteID uuid.UUID, parentID *uuid.UUID) ([]model.FolderResponse, error) {
	folders, err := s.assetRepo.ListFolders(ctx, siteID, parentID)
	if err != nil {
		return nil, err
	}

	responses := make([]model.FolderResponse, len(folders))
	for i, folder := range folders {
		responses[i] = folder.ToFolderResponse()
	}

	return responses, nil
}

// GetFolderTree 获取文件夹树
func (s *AssetService) GetFolderTree(ctx context.Context, siteID uuid.UUID) ([]model.FolderResponse, error) {
	folders, err := s.assetRepo.GetFolderTree(ctx, siteID)
	if err != nil {
		return nil, err
	}

	responses := make([]model.FolderResponse, len(folders))
	for i, folder := range folders {
		responses[i] = folder.ToFolderResponse()
	}

	return responses, nil
}

// UpdateFolder 更新文件夹
func (s *AssetService) UpdateFolder(ctx context.Context, id uuid.UUID, req *model.FolderUpdate) (*model.AssetFolder, error) {
	folder, err := s.assetRepo.GetFolderByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.ParentID != nil {
		// 防止将自己设为父文件夹
		if *req.ParentID == id {
			return nil, errors.New("不能将自己设为父文件夹")
		}
		folder.ParentID = req.ParentID
	}
	if req.Name != nil {
		folder.Name = *req.Name
		folder.Slug = s.generateSlug(*req.Name, "")
	}
	if req.SortOrder != nil {
		folder.SortOrder = *req.SortOrder
	}

	if err := s.assetRepo.UpdateFolder(ctx, folder); err != nil {
		return nil, err
	}

	return folder, nil
}

// DeleteFolder 删除文件夹
func (s *AssetService) DeleteFolder(ctx context.Context, id uuid.UUID) error {
	return s.assetRepo.DeleteFolder(ctx, id)
}
