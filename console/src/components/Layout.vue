<script setup lang="ts">

// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

import { ref, computed, nextTick, provide, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useUserStore } from '@/stores/user'
import { useSiteStore } from '@/stores/site'
import request, { showError, showSuccess } from '@/utils/request'
import { DialogPlugin } from 'tdesign-vue-next'
import LangSwitcher from './LangSwitcher.vue'

const { t } = useI18n()
import { version } from '../../package.json'
const edition = import.meta.env.VITE_EDITION || 'community'
const editionLabel = edition === 'enterprise' ? t('edition.enterprise') : t('edition.community')

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()
const siteStore = useSiteStore()

// 全局清除所有缓存
const clearingAllCache = ref(false)
const handleClearAllCache = () => {
  const dialog = DialogPlugin.confirm({
    header: t('cache.clearAllTitle'),
    body: t('cache.clearAllBody'),
    confirmBtn: { content: t('cache.clearAllConfirm'), theme: 'danger' },
    cancelBtn: t('common.cancel'),
    onConfirm: async () => {
      clearingAllCache.value = true
      try {
        const res = await request.post('/cache/invalidate/all')
        showSuccess(t('cache.clearAllSuccess', { count: res.data?.deleted || 0 }))
      } catch (error: any) {
        showError(error)
      } finally {
        clearingAllCache.value = false
      }
      dialog.destroy()
    },
  })
}

const sidebarCollapsed = ref(false)

// 初始化状态：子组件可等待此状态变为 true 再发起需要 X-Site-ID 的请求
const initialized = ref(false)

// 提供初始化状态和 siteStore 供子组件使用
// 子组件（如 Dashboard）需要等待 initialized 为 true 后才能安全地发起需要 X-Site-ID 的请求
provide('layoutInitialized', initialized)
provide('siteStore', siteStore)

// 创建站点弹窗
const showCreateSite = ref(false)
const newSiteName = ref('')
const newSiteSlug = ref('')
const newSiteDesc = ref('')
const creating = ref(false)

// 菜单项 → 权限映射
const MENU_PERMISSION_MAP: Record<string, string> = {
  '/': 'dashboard:read',
  '/sites': 'sites:read',
  '/content/schemas': 'schema:read',
  '/content/entries': 'entry:read',
  '/assets': 'asset:read',
  '/users': 'users:read',
  '/tokens': 'tokens:read',
  '/system/roles': 'roles:read',
  '/system/permissions': 'roles:read',
  '/audit/logs': 'audit:read',
  '/system/config': 'settings:read',
}

// 根据权限过滤菜单项
interface MenuGroup {
  label: string
  items: { path: string; icon: string; label: string; name: string; tIcon: string }[]
}

const filteredMenuItems = computed(() => {
  const allGroups: MenuGroup[] = [
    {
      label: t('menu.dashboard'),
      items: [
        { path: '/', icon: 'dashboard', label: t('menu.dashboard'), name: 'Dashboard', tIcon: 'dashboard' },
      ],
    },
    {
      label: t('menu.contentManagement'),
      items: [
        { path: '/sites', icon: 'layers', label: t('menu.sites'), name: 'Sites', tIcon: 'layers' },
        { path: '/content/schemas', icon: 'schema', label: t('menu.contentSchemas'), name: 'ContentSchemas', tIcon: 'server' },
        { path: '/content/entries', icon: 'article', label: t('menu.contentEntries'), name: 'Content', tIcon: 'article' },
        { path: '/assets', icon: 'image', label: t('menu.media'), name: 'Media', tIcon: 'image' },
      ],
    },
    {
      label: t('menu.accessControl'),
      items: [
        { path: '/tokens', icon: 'key', label: t('menu.tokens'), name: 'ApiTokens', tIcon: 'key' },
      ],
    },
    {
      label: t('menu.systemManagement'),
      items: [
        { path: '/system/config', icon: 'setting', label: t('menu.systemConfig'), name: 'SystemConfig', tIcon: 'setting' },
        { path: '/users', icon: 'people', label: t('menu.users'), name: 'Users', tIcon: 'user' },
        { path: '/system/roles', icon: 'shield', label: t('roles.title'), name: 'SystemRoles', tIcon: 'lock-on' },
        { path: '/system/permissions', icon: 'secured', label: t('permissions.title'), name: 'SystemPermissions', tIcon: 'secured' },
        { path: '/system/webhooks', icon: 'notification', label: t('menu.webhooks'), name: 'SystemWebhooks', tIcon: 'notification' },
      ],
    },
    {
      label: t('menu.auditLogs'),
      items: [
        { path: '/audit/logs', icon: 'file-search', label: t('menu.auditLogs'), name: 'AuditLogs', tIcon: 'file-search' },
      ],
    },
  ]

  // super_admin 看到所有菜单
  if (userStore.isSuperAdmin) {
    return allGroups
  }

  // 根据权限过滤每个分组内的菜单项
  return allGroups.map(group => ({
    ...group,
    items: group.items.filter(item => {
      const requiredPermission = MENU_PERMISSION_MAP[item.path]
      if (!requiredPermission) return true
      return userStore.hasPermission(requiredPermission)
    }),
  })).filter(group => group.items.length > 0)
}) // 移除空分组

const isActive = (path: string) => {
  if (path === '/') return route.path === '/'
  return route.path.startsWith(path)
}

const handleLogout = async () => {
  await userStore.logout()
  router.push('/login')
}

const user = computed(() => userStore.user)
const userLoading = ref(true)

const goToProfile = () => {
  router.push('/profile')
}

// 生成 slug
const generateSlug = (name: string) => {
  return name
    .toLowerCase()
    .replace(/[^a-z0-9\u4e00-\u9fff]/g, '-')
    .replace(/-+/g, '-')
    .replace(/^-|-$/g, '')
    .slice(0, 100) || 'my-site'
}

// 复制 Site ID 到剪贴板
const copySiteId = async () => {
  if (!siteStore.currentSiteId) return
  try {
    await navigator.clipboard.writeText(siteStore.currentSiteId)
    showSuccess(t('site.siteIdCopied'))
  } catch (e) {
    showError({ response: { data: { msg: t('site.copyFailed') } } } as any)
  }
}

const handleNameInput = (val: string) => {
  const name = String(val || '')
  // 仅自动生成 slug（当 slug 为空或等于自动生成值时才更新）
  if (!newSiteSlug.value || newSiteSlug.value === generateSlug(newSiteSlug.value)) {
    newSiteSlug.value = generateSlug(name)
  }
}

const handleCreateSite = async () => {
  if (!newSiteName.value.trim()) {
    showError({ response: { data: { msg: t('site.enterSiteName') } } } as any)
    return
  }
  const slug = newSiteSlug.value.trim() || generateSlug(newSiteName.value)
  if (!/^[a-z][a-z0-9-]{0,98}[a-z0-9]$/.test(slug)) {
    showError({ response: { data: { msg: t('site.formatError') } } } as any)
    return
  }
  creating.value = true
  const result = await siteStore.createAndSwitch({
    name: newSiteName.value.trim(),
    slug,
    description: newSiteDesc.value.trim() || undefined,
  })
  creating.value = false
  if (result.success) {
    closeCreateSiteDialog()
  }
}

// 关闭创建站点弹窗
const closeCreateSiteDialog = async () => {
  showCreateSite.value = false
  await nextTick()
  newSiteName.value = ''
  newSiteSlug.value = ''
  newSiteDesc.value = ''
}

const siteOptions = computed(() =>
  siteStore.sites.map(s => ({ label: s.name, value: s.id }))
)
// 响应式跟踪 sites 加载状态，替代独立的 siteLoading ref
const siteLoading = computed(() => siteStore.loading || siteStore.sites.length === 0)

// 监听用户数据恢复，加载站点列表和权限
// 用 user（Vue ref）而非 isLoggedIn（computed），确保 Pinia 响应式正常触发
watch(
  () => userStore.user,
  async (u) => {
    if (!u) { userLoading.value = true; return }
    userLoading.value = false

    if (siteStore.sites.length === 0) {
      try {
        await siteStore.fetchSites()
      } catch {
        // 站点列表加载失败不应阻塞页面渲染
      }
    }

    if (!userStore.isSuperAdmin && userStore.permissions.length === 0) {
      try {
        await userStore.fetchPermissions()
      } catch {
        // 权限列表加载失败不应阻塞页面渲染
      }
    }

    initialized.value = true
  },
  { immediate: true }
)
</script>

<template>
  <div class="app-layout" :class="{ collapsed: sidebarCollapsed }">
    <!-- 顶部 Header -->
    <header class="app-header">
      <div class="header-left">
        <div class="logo">
          <img src="/assets/logo.png" alt="Contful" class="logo-img" />
        </div>
        <t-button
          shape="square" variant="text"
          :aria-label="sidebarCollapsed ? t('a11y.expandSidebar') : t('a11y.collapseSidebar')"
          style="flex-shrink: 0"
          @click="sidebarCollapsed = !sidebarCollapsed"
        >
          <template #icon>
            <Icon :name="sidebarCollapsed ? 'indent-right' : 'indent-left'" />
          </template>
        </t-button>
        <!-- 站点选择器：紧跟折叠按钮，左侧分隔线对齐 -->
        <div class="site-selector">
          <t-select
            v-if="!siteLoading"
            v-model="siteStore.currentSiteId"
            :options="siteOptions"
            :placeholder="t('site.selectSite')"
            :clearable="false"
            style="width: 220px"
            @change="(val: string) => siteStore.setCurrentSite(val)"
          />
          <t-skeleton v-else :loading="true" theme="text" :width="220" />
          <!-- 复制 Site ID 按钮 -->
          <t-tooltip
            v-if="siteStore.currentSiteId"
            :content="t('site.copySiteId') + ': ' + siteStore.currentSiteId"
            placement="bottom"
          >
            <t-button
              shape="square" variant="text"
              size="small"
              @click="copySiteId"
            >
              <template #icon>
                <Icon name="file-copy" />
              </template>
            </t-button>
          </t-tooltip>
        </div>
      </div>
      <!-- 占位，撑开左右两端的间距 -->
      <div class="header-spacer"></div>
      <div class="header-right">
        <a
          href="https://contful.com"
          target="_blank"
          rel="noopener noreferrer"
          class="header-link"
          :title="t('header.officialSite')"
        >
          {{ t('header.officialSite') }}
        </a>
        <LangSwitcher />
        <t-button
          variant="text"
          shape="square"
          :loading="clearingAllCache"
          :title="t('cache.clearAllTitle')"
          @click="handleClearAllCache"
        >
          <template #icon><t-icon name="clear" /></template>
        </t-button>
        <t-dropdown trigger="click">
          <div class="user-trigger">
            <!-- 加载中：显示骨架屏 -->
            <template v-if="userLoading">
              <t-skeleton theme="avatar" :loading="true" size="small" />
              <t-skeleton theme="text" :loading="true" :width="80" style="margin-left: 8px" />
            </template>
            <!-- 加载完成：显示用户信息 -->
            <template v-else-if="user">
              <img
                :src="user.avatar_url || '/assets/default.jpeg'"
                class="avatar"
                :alt="user.nickname || user.email"
                @error="($event.target as HTMLImageElement).style.display='none'"
              />
              <span class="user-name">{{ user.nickname || user.email }}</span>
            </template>
            <!-- 兜底：显示默认头像 -->
            <template v-else>
              <img src="/assets/default.jpeg" class="avatar" alt="User" />
            </template>
            <t-icon name="chevron-down" size="14px" style="color: var(--color-text-secondary)" />
          </div>
          <template #dropdown>
            <t-dropdown-menu>
              <t-dropdown-item class="dropdown-email-item" disabled>
                {{ user?.email }}
              </t-dropdown-item>
              <t-dropdown-item @click="goToProfile">
                <template #prefix-icon><t-icon name="user" /></template>
                {{ t('settings.personalProfile') }}
              </t-dropdown-item>
              <t-divider />
              <t-dropdown-item theme="error" @click="handleLogout">
                <template #prefix-icon><t-icon name="poweroff" /></template>
                {{ t('common.logout') }}
              </t-dropdown-item>
            </t-dropdown-menu>
          </template>
        </t-dropdown>
      </div>
    </header>

    <div class="app-body">
      <!-- 侧边栏 -->
      <aside class="sidebar">
        <nav class="sidebar-nav" aria-label="主导航">
          <template v-for="group in filteredMenuItems" :key="group.label">
            <div class="nav-group-label">{{ group.label }}</div>
            <router-link
              v-for="item in group.items"
              :key="item.path"
              :to="item.path"
              class="nav-item"
              :class="{ active: isActive(item.path) }"
              :aria-current="isActive(item.path) ? 'page' : undefined"
            >
              <span class="nav-icon">
                <Icon :name="item.tIcon" />
              </span>
              <span class="nav-label">{{ item.label }}</span>
            </router-link>
          </template>
        </nav>
        <div class="sidebar-footer">
          <span class="version-text">{{ editionLabel }} v{{ version }}</span>
        </div>
      </aside>

      <!-- 主内容区 -->
      <main id="main-content" class="main-content">
        <div class="content-wrapper">
          <slot />
        </div>
      </main>
    </div>

    <!-- 创建站点弹窗 -->
    <t-dialog
      v-model:visible="showCreateSite"
      :header="t('site.createSite')"
      :close-on-overlay-click="true"
      :close-on-esc-keydown="true"
      :destroy-on-close="false"
      @close="closeCreateSiteDialog"
    >
      <t-form layout="vertical">
        <t-form-item :label="t('site.siteName')" required>
          <t-input
            v-model="newSiteName"
            :placeholder="t('site.siteNamePlaceholder')"
            :maxlength="200"
            @change="handleNameInput"
          />
        </t-form-item>
        <t-form-item :label="t('site.siteSlug')" required>
          <t-input
            v-model="newSiteSlug"
            :placeholder="t('site.siteSlugPlaceholder')"
            :maxlength="100"
          />
        </t-form-item>
        <t-form-item :label="t('site.siteDescription')">
          <t-textarea
            v-model="newSiteDesc"
            :placeholder="t('site.siteDescPlaceholder')"
            :maxlength="2000"
            :autosize="{ minRows: 2, maxRows: 4 }"
          />
        </t-form-item>
      </t-form>
      <template #footer>
        <div class="dialog-footer">
          <t-button variant="outline" @click="closeCreateSiteDialog">{{ t('common.cancel') }}</t-button>
          <t-button theme="primary" :loading="creating" @click="handleCreateSite">{{ t('common.create') }}</t-button>
        </div>
      </template>
    </t-dialog>
  </div>
</template>

<style scoped>
.app-layout {
  display: flex;
  flex-direction: column;
  height: 100vh;
  overflow: hidden;
  background: var(--color-bg);
}

/* 顶部 Header */
.app-header {
  height: 56px;
  display: flex;
  align-items: center;
  padding: 0 20px;
  background: var(--color-header-bg);
  border-bottom: 1px solid var(--color-border);
  position: sticky;
  top: 0;
  z-index: 100;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.header-spacer {
  flex: 1;
}

.logo {
  display: flex;
  align-items: center;
  gap: 10px;
}

.logo-img {
  height: 32px;
  width: auto;
}

.site-selector {
  display: flex;
  align-items: center;
  gap: 4px;
  padding-left: 12px;
  border-left: 1px solid var(--color-border);
}

.site-selector :deep(.t-select) {
  font-size: 14px;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-shrink: 0;
}

.header-link {
  display: flex;
  align-items: center;
  padding: 6px 12px;
  font-size: 14px;
  color: var(--color-text-secondary);
  text-decoration: none;
  border-radius: 6px;
  transition: all 0.2s;
}

.header-link:hover {
  color: var(--color-primary);
  background: var(--color-hover);
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

.user-menu {
  display: flex;
  align-items: center;
  gap: 10px;
}

.avatar {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  object-fit: cover;
  flex-shrink: 0;
}

.user-name {
  font-size: 14px;
  font-weight: 500;
  color: var(--color-text);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 120px;
}

.user-trigger {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 6px;
  transition: background 0.2s;
}

.user-trigger:hover {
  background: var(--color-hover);
}

:deep(.dropdown-email-item) {
  font-size: 12px;
  color: var(--color-text-secondary);
  cursor: default;
  pointer-events: none;
}

/* 主体区域 */
.app-body {
  display: flex;
  flex: 1;
  overflow: hidden;
}

/* 侧边栏 */
.sidebar {
  width: 220px;
  background: var(--color-sidebar);
  display: flex;
  flex-direction: column;
  transition: width var(--transition-normal);
  flex-shrink: 0;
}

.app-layout.collapsed .sidebar {
  width: 64px;
}

.sidebar-nav {
  flex: 1;
  padding: var(--space-4) var(--space-3);
  overflow-y: auto;
}

.nav-group-label {
  font-size: 11px;
  font-weight: 600;
  color: var(--color-sidebar-text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  padding: var(--space-4) 14px var(--space-2);
  white-space: nowrap;
  overflow: hidden;
  transition: opacity var(--transition-normal);
}

.nav-group-label:first-child {
  padding-top: 0;
}

.app-layout.collapsed .nav-group-label {
  opacity: 0;
}

.sidebar-footer {
  padding: var(--space-3) 14px;
  border-top: 1px solid var(--color-sidebar-hover);
  text-align: center;
}

.version-text {
  font-size: 12px;
  color: var(--color-sidebar-text-secondary);
}

.nav-item {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: 10px 14px;
  margin-bottom: var(--space-1);
  border-radius: var(--radius-md);
  color: var(--color-sidebar-text);
  text-decoration: none;
  transition: all var(--transition-fast);
}

.nav-item:hover {
  background: var(--color-sidebar-hover);
  color: #fff;
}

.nav-item.active {
  background: var(--color-primary);
  color: #fff;
}

.nav-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  width: 20px;
  height: 20px;
}

.nav-label {
  font-size: 14px;
  font-weight: 500;
  white-space: nowrap;
  overflow: hidden;
  transition: opacity var(--transition-normal);
}

.app-layout.collapsed .nav-label {
  opacity: 0;
  width: 0;
  overflow: hidden;
}

/* 主内容区 */
.main-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: var(--color-bg);
}

.content-wrapper {
  flex: 1;
  padding: 24px;
  overflow-y: auto;
}
</style>
