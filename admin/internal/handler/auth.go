// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/contful/contful/admin/internal/crypto"
	"github.com/contful/contful/admin/internal/middleware"
	"github.com/contful/contful/admin/internal/model"
	"github.com/contful/contful/admin/internal/repository"
	"github.com/contful/contful/admin/internal/service"
)

type AuthHandler struct {
	authService   *service.AuthService
	asymCrypter   crypto.AsymmetricCrypter // 非对称加密器（RSA 或 SM2）
	asymPubKey    string                   // 公钥 PEM（RSA 或 SM2）
	asymPrivKey   string                   // 私钥 PEM（RSA 或 SM2）
	cryptoMode    string                   // 加密模式（rsa 或 sm）
	sm2PubKeyHex  string                   // SM2 原始公钥 hex（仅 SM2 模式，供前端 sm-crypto 使用）
}

// setRefreshTokenCookie 将 refresh_token 写入 HttpOnly Cookie
// Secure 属性根据当前请求协议动态决定：HTTPS=true，HTTP=false（兼容本地开发环境）
func setRefreshTokenCookie(c *gin.Context, token string, maxAge int) {
	secure := c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https"
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("refresh_token", token, maxAge, "/", "", secure, true) // HttpOnly + 动态 Secure + SameSite=Lax
}

// clearRefreshTokenCookie 清除 refresh_token Cookie（清除时 Secure 也需要匹配）
func clearRefreshTokenCookie(c *gin.Context) {
	secure := c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https"
	c.SetCookie("refresh_token", "", -1, "/", "", secure, true)
}

func NewAuthHandler(authService *service.AuthService, asymCrypter crypto.AsymmetricCrypter, pubKeyPEM, privKeyPEM, cryptoMode string) *AuthHandler {
	h := &AuthHandler{
		authService: authService,
		asymCrypter: asymCrypter,
		asymPubKey:  pubKeyPEM,
		asymPrivKey: privKeyPEM,
		cryptoMode:  cryptoMode,
	}

	// SM2 模式：提取原始公钥 hex（供前端 sm-crypto 使用）
	if cryptoMode == crypto.ModeSM {
		if hex, err := crypto.SM2PubKeyToHex(pubKeyPEM); err == nil {
			h.sm2PubKeyHex = hex
		}
	}

	return h
}

// PublicKey 获取非对称加密公钥（RSA 或 SM2）和一次性 Anti-Replay Token
// GET /admin/api/v1/auth/public/key
func (h *AuthHandler) PublicKey(c *gin.Context) {
	token, tokenID, err := h.authService.GenerateRSAToken(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, "failed to generate token"))
		return
	}

	data := map[string]string{
		"public_key":  h.asymPubKey,
		"crypto_mode": h.cryptoMode,
		"token_id":    tokenID,
		"token":       token,
	}

	// SM2 模式：额外返回原始公钥 hex（前端 sm-crypto 需要此格式）
	if h.cryptoMode == crypto.ModeSM && h.sm2PubKeyHex != "" {
		data["sm2_public_key_hex"] = h.sm2PubKeyHex
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(data))
}

// Register 注册
// POST /admin/v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, "invalid request: "+err.Error()))
		return
	}

	ip := c.ClientIP()
	user, err := h.authService.Register(c.Request.Context(), &req, ip)
	if err != nil {
		if errors.Is(err, repository.ErrUserAlreadyExists) {
			c.JSON(http.StatusConflict, model.NewErrorResponse(model.CodeConflict, "email already exists"))
			return
		}
		// P1-004: 密码强度错误
		if errors.Is(err, service.ErrPasswordTooShort) ||
			errors.Is(err, service.ErrPasswordNoUppercase) ||
			errors.Is(err, service.ErrPasswordNoLowercase) ||
			errors.Is(err, service.ErrPasswordNoDigit) ||
			errors.Is(err, service.ErrPasswordNoSpecialChar) {
			c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, err.Error()))
			return
		}
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, "internal error"))
		return
	}

	c.JSON(http.StatusCreated, model.NewSuccessResponse(user))
}

// Login 登录
// POST /admin/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, "invalid request: "+err.Error()))
		return
	}

	// 非对称加密密码处理（RSA 或 SM2，根据 crypto_mode 自动选择）
	if req.EncryptedPassword != "" && req.TokenID != "" && req.RSAToken != "" {
		// 验证 Anti-Replay Token
		if !h.authService.ValidateAndConsumeRSAToken(c.Request.Context(), req.TokenID, req.RSAToken) {
			c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, "invalid or expired security token"))
			return
		}
		// 解密密码（RSA 或 SM2）
		if h.asymCrypter != nil {
			plaintext, err := h.asymCrypter.AsymDecrypt([]byte(h.asymPrivKey), []byte(req.EncryptedPassword))
			if err != nil {
				c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, "failed to decrypt password"))
				return
			}
			// 提取密码（格式：password@@token）
			parts := strings.SplitN(string(plaintext), "@@", 2)
			if len(parts) == 2 && parts[1] == req.RSAToken {
				req.Password = parts[0]
			} else {
				c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, "token mismatch"))
				return
			}
		}
	}

	if req.Password == "" {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, "password is required"))
		return
	}

	ip := c.ClientIP()
	resp, err := h.authService.Login(c.Request.Context(), &req, ip)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrInvalidPassword):
			c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "invalid email or password"))
		case errors.Is(err, repository.ErrUserInactive):
			c.JSON(http.StatusForbidden, model.NewErrorResponse(model.CodeForbidden, "user is inactive"))
		case errors.Is(err, repository.ErrUserSuspended):
			c.JSON(http.StatusForbidden, model.NewErrorResponse(model.CodeForbidden, "user is suspended"))
		case errors.Is(err, service.ErrAccountLocked):
			c.JSON(http.StatusLocked, model.NewErrorResponse(model.CodeLocked, err.Error()))
		default:
			c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, "internal error"))
		}
		return
	}

	// 如果返回的是 MFA Required 响应，不需要设置 Cookie
	if mfaResp, ok := resp.(*model.MFARequiredResponse); ok && mfaResp.MFARequired {
		c.JSON(http.StatusOK, model.NewSuccessResponse(mfaResp))
		return
	}

	// 设置 RefreshToken 到 HttpOnly Cookie（安全增强）
	if loginResp, ok := resp.(*model.LoginResponse); ok && loginResp.RefreshToken != "" {
		setRefreshTokenCookie(c, loginResp.RefreshToken, 604800)
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(resp))
}
// Refresh 刷新 Token
// POST /admin/v1/auth/refresh
// 优先从 HttpOnly Cookie 读取 refresh_token（安全），无 Cookie 时从 Authorization Header 兼容旧版
func (h *AuthHandler) Refresh(c *gin.Context) {
	var refreshToken string

	// 优先从 HttpOnly Cookie 读取（Login 时写入）
	cookie, err := c.Cookie("refresh_token")
	log.Debug().Str("cookie", cookie).Err(err).Msg("[Refresh] reading refresh_token cookie")
	if err == nil && cookie != "" {
		refreshToken = cookie
	} else {
		// 兜底：从 Authorization Header 读取（Barear accessToken.refreshToken 格式）
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "missing refresh token"))
			return
		}
		refreshToken = strings.TrimPrefix(authHeader, "Bearer ")
	}

	newAccessToken, newRefreshToken, err := h.authService.Refresh(c.Request.Context(), refreshToken)
	if err != nil {
		// 刷新失败时清除 Cookie
		clearRefreshTokenCookie(c)
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "invalid refresh token"))
		return
	}

	// 新 RefreshToken 也写入 HttpOnly Cookie（Token 轮换）
	setRefreshTokenCookie(c, newRefreshToken, 604800)

	c.JSON(http.StatusOK, model.NewSuccessResponse(model.RefreshResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}))
}

// Logout 登出
// POST /admin/v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	// 优先从 Cookie 删除 refresh token
	if cookie, err := c.Cookie("refresh_token"); err == nil && cookie != "" {
		ip := c.ClientIP()
		h.authService.Logout(c.Request.Context(), cookie, ip) // 忽略错误，不阻断
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(nil))
}

// Me 获取当前用户
// GET /admin/v1/users/me
func (h *AuthHandler) Me(c *gin.Context) {
	userIDVal, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "unauthorized"))
		return
	}

	userIDStr, ok := userIDVal.(string)
	if !ok {
		// 尝试从 uuid.UUID 获取
		if uid, ok := userIDVal.(uuid.UUID); ok {
			userIDStr = uid.String()
		} else {
			c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "invalid user id"))
			return
		}
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "invalid user id"))
		return
	}

	user, err := h.authService.GetUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, "internal error"))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(user))
}

// GetClaims 实现 middleware.claimsGetter 接口，供 JWT 中间件使用
func (h *AuthHandler) GetClaims(token string) (*middleware.Claims, error) {
	// JWT 格式: header.payload.signature，直接解析即可
	claims, err := h.authService.ParseAccessTokenInternal(token)
	if err != nil {
		return nil, err
	}

	return &middleware.Claims{
		UserID:           claims.UserID,
		Email:            claims.Email,
		IsSuperAdmin:     claims.IsSuperAdmin,
		MFASetupRequired: claims.MFASetupRequired,
	}, nil
}

// ListUsers 获取用户列表
// GET /admin/v1/users
func (h *AuthHandler) ListUsers(c *gin.Context) {
	var req model.PageRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, "invalid request"))
		return
	}

	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}

	page, err := h.authService.ListUsers(c.Request.Context(), req.Page, req.PageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, "internal error"))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(page))
}
