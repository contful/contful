<script setup lang="ts">

// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { DialogPlugin } from 'tdesign-vue-next'
import { useSiteStore } from '@/stores/site'
import { showError, showSuccess } from '@/utils/request'
import {
  getContentSchemas,
  type ContentSchema,
} from '@/api/schema'
import {
  getEntries,
  getEntry,
  createEntry,
  updateEntry,
  deleteEntry,
  publishEntry,
  unpublishEntry,
  batchDeleteEntries,
  batchPublishEntries,
  batchUnpublishEntries,
  invalidateCache,
  invalidateSchemaCache,
  type Entry,
} from '@/api/entry'
import PageHeader from '@/components/PageHeader.vue'
import TableToolbar from '@/components/common/TableToolbar.vue'
import RichTextEditor from '@/components/RichTextEditor.vue'
import JsonEditor from '@/components/JsonEditor.vue'
import ScheduleCalendar from './ScheduleCalendar.vue'

// 类型守卫：处理 unknown 类型的 error 参数
const handleError = (error: unknown) => showError(error as Parameters<typeof showError>[0])

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const siteStore = useSiteStore()

// 状态
const loading = ref(false)
const submitting = ref(false)
const contentSchemas = ref<ContentSchema[]>([])
const entries = ref<Entry[]>([])
const selectedType = ref<ContentSchema | null>(null)
const showModal = ref(false)
const editingEntry = ref<Entry | null>(null)

// 内容类型条目数缓存
const entryCounts = ref<Record<string, number>>({})

// 分页
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)

// 表单数据
const formData = ref<Record<string, any>>({})

// 排期时间
const scheduledPublishTime = ref<string>('')
const scheduledUnpublishTime = ref<string>('')

// Tab 状态
const activeTab = ref<string>('list')

// 过滤器
const statusFilter = ref<string>('')

// 搜索与排序
const searchKeyword = ref<string>('')
const sortField = ref<string>('updated_time')
const sortOrder = ref<'asc' | 'desc'>('desc')

// 批量选择
const selectedIds = ref<Set<string>>(new Set())
const batchLoading = ref(false)

// 缓存状态
const cacheLoading = ref(false)

// 发布加载状态
const publishLoading = ref<string | null>(null)

// 计算属性
const isAllSelected = computed(() => {
  return entries.value.length > 0 && selectedIds.value.size === entries.value.length
})

const selectedCount = computed(() => selectedIds.value.size)
const hasSelected = computed(() => selectedIds.value.size > 0)

// 加载内容类型
const loadContentSchemas = async () => {
  if (!siteStore.currentSiteId) {
    contentSchemas.value = []
    return
  }
  try {
    const res = await getContentSchemas({ page: 1, page_size: 100 })
    contentSchemas.value = res.data?.items || []
    await loadEntryCounts()
  } catch (error) {
    handleError(error)
  }
}

// 加载每个内容类型的条目数
const loadEntryCounts = async () => {
  const counts: Record<string, number> = {}
  for (const type of contentSchemas.value) {
    try {
      const res = await getEntries({ schema_id: type.id, page: 1, page_size: 1 })
      counts[type.id] = res.data?.total || 0
    } catch {
      counts[type.id] = 0
    }
  }
  entryCounts.value = counts
}

// 加载内容列表
const loadEntries = async () => {
  if (!selectedType.value) return
  loading.value = true
  selectedIds.value.clear()
  try {
    const params: any = {
      page: page.value,
      page_size: pageSize.value,
      schema_id: selectedType.value.id,
      sort_field: sortField.value,
      sort_order: sortOrder.value,
    }
    if (statusFilter.value) {
      params.status = statusFilter.value
    }
    if (searchKeyword.value) {
      params.keyword = searchKeyword.value
    }
    const res = await getEntries(params)
    entries.value = res.data?.items || []
    total.value = res.data?.total || 0
  } catch (error) {
    handleError(error)
  } finally {
    loading.value = false
  }
}

// 选择内容类型
const selectType = (type: ContentSchema) => {
  selectedType.value = type
  page.value = 1
  loadEntries()
}

// 打开创建弹窗
const openCreateModal = () => {
  editingEntry.value = null
  formData.value = {}
  showModal.value = true
}

// 打开编辑弹窗
const openEditModal = async (entry: Entry) => {
  try {
    const res = await getEntry(entry.id)
    editingEntry.value = res.data
    formData.value = { ...res.data.values || {} }
    scheduledPublishTime.value = res.data.scheduled_publish_time || ''
    scheduledUnpublishTime.value = res.data.scheduled_unpublish_time || ''
    showModal.value = true
  } catch (error) {
    handleError(error)
  }
}

// 关闭弹窗
const closeModal = () => {
  showModal.value = false
  editingEntry.value = null
  formData.value = {}
  scheduledPublishTime.value = ''
  scheduledUnpublishTime.value = ''
}

// 提交表单
const handleSubmit = async () => {
  if (!selectedType.value) return
  submitting.value = true
  try {
    if (editingEntry.value) {
      const updateData: any = { values: formData.value }
      updateData.scheduled_publish_time = scheduledPublishTime.value || null
      updateData.scheduled_unpublish_time = scheduledUnpublishTime.value || null
      await updateEntry(editingEntry.value.id, updateData)
      showSuccess(t('content.updateSuccess'))
    } else {
      const createData: any = { schema_id: selectedType.value.id, values: formData.value }
      createData.scheduled_publish_time = scheduledPublishTime.value || undefined
      createData.scheduled_unpublish_time = scheduledUnpublishTime.value || undefined
      await createEntry(createData)
      showSuccess(t('content.createSuccess'))
    }
    closeModal()
    loadEntries()
    loadEntryCounts()
  } catch (error) {
    handleError(error)
  } finally {
    submitting.value = false
  }
}

// 发布/取消发布
const handlePublish = async (entry: Entry) => {
  publishLoading.value = entry.id
  try {
    if (entry.status === 'published') {
      await unpublishEntry(entry.id)
      showSuccess(t('content.unpublishSuccess'))
    } else {
      await publishEntry(entry.id)
      showSuccess(t('content.publishSuccess'))
    }
    loadEntries()
  } catch (error) {
    handleError(error)
  } finally {
    publishLoading.value = null
  }
}

// 删除确认 — 显示具体 Entry 名称
const confirmDelete = (entry: Entry) => {
  const entryTitle = entry.values?.title || entry.id.slice(0, 8)
  const dialog = DialogPlugin.confirm({
    header: t('common.confirmDelete'),
    body: `${t('content.confirmDelete')}「${entryTitle}」？`,
    theme: 'warning',
    onConfirm: async () => {
      try {
        await deleteEntry(entry.id)
        showSuccess(t('content.deleteSuccess'))
        loadEntries()
        loadEntryCounts()
        dialog.destroy()
      } catch (error) {
        handleError(error)
      }
    },
  })
}

// 切换单选
const toggleSelect = (id: string) => {
  if (selectedIds.value.has(id)) {
    selectedIds.value.delete(id)
  } else {
    selectedIds.value.add(id)
  }
  // 触发响应式更新
  selectedIds.value = new Set(selectedIds.value)
}

// 切换全选
const toggleSelectAll = () => {
  if (isAllSelected.value) {
    selectedIds.value.clear()
  } else {
    entries.value.forEach(entry => {
      selectedIds.value.add(entry.id)
    })
  }
  selectedIds.value = new Set(selectedIds.value)
}

// 批量操作 — DialogPlugin.confirm
const confirmBatchAction = (action: 'delete' | 'publish' | 'unpublish') => {
  const labels: Record<string, string> = {
    delete: t('content.batchDelete'),
    publish: t('content.batchPublish'),
    unpublish: t('content.batchUnpublish'),
  }
  const count = selectedIds.value.size
  let bodyText: string
  if (action === 'delete') {
    bodyText = t('content.batchDeleteMsg', { count })
  } else {
    bodyText = t('content.batchActionMsg', { action: labels[action], count })
  }

  const dlg = DialogPlugin.confirm({
    header: t('common.confirmAction', { action: labels[action] }),
    body: bodyText,
    theme: action === 'delete' ? 'warning' : 'info',
    onConfirm: async () => {
      const ids = Array.from(selectedIds.value)
      if (ids.length === 0) return
      batchLoading.value = true
      try {
        switch (action) {
          case 'delete':
            await batchDeleteEntries(ids)
            showSuccess(t('content.deleted', { count: ids.length }))
            break
          case 'publish':
            await batchPublishEntries(ids)
            showSuccess(t('content.publishedCount', { count: ids.length }))
            break
          case 'unpublish':
            await batchUnpublishEntries(ids)
            showSuccess(t('content.unpublishedCount', { count: ids.length }))
            break
        }
        selectedIds.value.clear()
        selectedIds.value = new Set()
        loadEntries()
        loadEntryCounts()
        dlg.destroy()
      } catch (error) {
        handleError(error)
      } finally {
        batchLoading.value = false
      }
    },
  })
}

// 在新窗口打开链接
const openLink = (url: string) => {
  window.open(url, '_blank')
}

// 清除当前模型缓存
const handleClearCache = () => {
  if (!selectedType.value) return
  const dialog = DialogPlugin.confirm({
    header: t('content.clearSchemaCacheTitle'),
    body: t('content.clearSchemaCacheBody', { schema: selectedType.value.name }),
    confirmBtn: { content: t('common.confirm'), theme: 'warning' },
    cancelBtn: t('common.cancel'),
    onConfirm: async () => {
      cacheLoading.value = true
      try {
        const res = await invalidateSchemaCache(selectedType.value!.slug)
        showSuccess(t('content.cacheCleared', { count: res.data?.deleted || 0 }))
      } catch (error) {
        handleError(error)
      } finally {
        cacheLoading.value = false
      }
      dialog.destroy()
    },
  })
}

// 清除本站点所有内容缓存
const siteCacheLoading = ref(false)
const handleClearSiteCache = () => {
  const dialog = DialogPlugin.confirm({
    header: t('content.clearSiteCacheConfirmTitle'),
    body: t('content.clearSiteCacheConfirmBody'),
    confirmBtn: { content: t('content.clearSiteCacheConfirm'), theme: 'danger' },
    cancelBtn: t('common.cancel'),
    onConfirm: async () => {
      siteCacheLoading.value = true
      try {
        const res = await invalidateCache()
        showSuccess(t('content.cacheCleared', { count: res.data?.deleted || 0 }))
      } catch (error) {
        handleError(error)
      } finally {
        siteCacheLoading.value = false
      }
      dialog.destroy()
    },
  })
}

// 清除搜索
const clearSearch = () => {
  searchKeyword.value = ''
  page.value = 1
  loadEntries()
}

// 状态样式 → t-tag theme
const getStatusTheme = (entry: Entry): string => {
  if (entry.scheduled_publish_time && entry.status === 'draft') return 'primary'
  if (entry.scheduled_unpublish_time && entry.status === 'published') return 'warning'
  const map: Record<string, string> = {
    published: 'success',
    draft: 'warning',
    archived: 'default',
  }
  return map[entry.status] || 'default'
}

const getStatusLabel = (entry: Entry) => {
  if (entry.scheduled_publish_time && entry.status === 'draft') return t('schedule.pendingPublish')
  if (entry.scheduled_unpublish_time && entry.status === 'published') return t('schedule.pendingUnpublish')
  const map: Record<string, string> = {
    published: t('content.published'),
    draft: t('content.draft'),
    archived: t('content.archived'),
  }
  return map[entry.status] || entry.status
}

// 格式化字段值
const formatFieldValue = (value: any): string => {
  if (value === null || value === undefined) return '-'
  if (typeof value === 'object') {
    if ('value' in value) return String(value.value)
    return JSON.stringify(value)
  }
  return String(value)
}

// 格式化日期
const formatDate = (date: string) => {
  const d = new Date(date)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

// 分页变化
const onPageChange = ({ current, pageSize: ps }: { current: number; pageSize: number }) => {
  page.value = current
  pageSize.value = ps
  loadEntries()
}

// 监听内容类型变化
watch(() => route.query.type, (newType) => {
  if (newType) {
    const type = contentSchemas.value.find(t => t.id === newType)
    if (type) selectType(type)
  }
}, { immediate: true })

onMounted(() => {
  loadContentSchemas()
})
</script>

<template>
    <PageHeader
      :title="t('content.title')"
      :subtitle="t('content.subtitle')"
    >
      <template #primary-action>
        <t-button
          theme="danger"
          variant="outline"
          :disabled="siteCacheLoading"
          :loading="siteCacheLoading"
          @click="handleClearSiteCache"
        >
          <template #icon><t-icon name="clear" /></template> {{ t('content.clearSiteCache') }}
        </t-button>
      </template>
    </PageHeader>

    <div class="content-layout">
      <!-- 无站点提示 -->
      <div v-if="!siteStore.currentSiteId" class="no-site-container">
        <div class="no-site-card">
          <t-icon name="home" size="64px" style="opacity:0.3" />
          <h3>{{ t('site.noSiteTitle') || '\u6682\u65e0\u7ad9\u70b9' }}</h3>
          <p>{{ t('site.noSiteHint') || '\u8bf7\u5148\u521b\u5efa\u4e00\u4e2a\u7ad9\u70b9\uff0c\u624d\u80fd\u7ba1\u7406\u5185\u5bb9' }}</p>
          <t-button theme="primary" @click="router.push('/')">{{ t('site.goToCreate') || '\u8fd4\u56de\u9996\u9875\u521b\u5efa\u7ad9\u70b9' }}</t-button>
        </div>
      </div>

      <!-- 有站点时正常显示 -->
      <template v-else>
      <!-- 侧边：内容类型列表 -->
      <aside class="type-sidebar">
        <div class="sidebar-header">
          <h3>{{ t('contentSchemas.title') }}</h3>
        </div>
        <div class="type-list">
          <button
            v-for="type in contentSchemas"
            :key="type.id"
            class="type-item"
            :class="{ active: selectedType?.id === type.id }"
            @click="selectType(type)"
          >
            <t-icon name="file" size="16px" class="type-icon" />
            <span class="type-name">{{ type.name }}</span>
            <span class="type-count">{{ entryCounts[type.id] || 0 }}</span>
          </button>
          <div v-if="contentSchemas.length === 0" class="empty-tip">
            {{ t('content.noContentSchemas') }}<router-link to="/content/schemas">{{ t('content.goToCreate') }}</router-link>
          </div>
        </div>
      </aside>

      <!-- 主内容区 -->
      <main class="content-main">
        <!-- Tab 切换 -->
        <t-tabs v-model="activeTab" class="schedule-tabs">
          <t-tab-panel value="list" :label="t('schedule.listView')">
            <template v-if="selectedType">
          <TableToolbar>
            <template #left>
              <h2>{{ selectedType.name }}</h2>
              <t-select
                v-model="statusFilter"
                :placeholder="t('content.allStatus')"
                size="small"
                style="width: 120px"
                clearable
                @change="loadEntries"
              >
                <t-option value="draft" :label="t('content.draft')" />
                <t-option value="published" :label="t('content.published')" />
                <t-option value="archived" :label="t('content.archived')" />
              </t-select>

              <!-- 搜索框 — t-input with prefix icon -->
              <t-input
                v-model="searchKeyword"
                :placeholder="t('content.searchContent')"
                clearable
                size="small"
                style="width: 200px"
                @enter="loadEntries"
                @clear="clearSearch"
              >
                <template #prefixIcon><t-icon name="search" /></template>
              </t-input>

              <!-- 排序 -->
              <t-select
                v-model="sortField"
                size="small"
                style="width: 130px"
                @change="loadEntries"
              >
                <t-option value="updated_time" :label="t('content.sortByUpdated')" />
                <t-option value="created_time" :label="t('content.sortByCreated')" />
                <t-option value="published_time" :label="t('content.sortByPublished')" />
                <t-option value="sort_weight" :label="t('content.sortByWeight')" />
              </t-select>
              <t-button variant="outline" @click="sortOrder = sortOrder === 'asc' ? 'desc' : 'asc'; loadEntries()">
                {{ sortOrder === 'asc' ? t('content.asc') : t('content.desc') }}
              </t-button>
            </template>

            <template #right>
              <t-button variant="outline" @click="loadEntries">
                <template #icon><t-icon name="refresh" /></template>
                {{ t('common.refresh') }}
              </t-button>
              <t-button
                variant="outline"
                :disabled="cacheLoading"
                :title="t('content.clearCacheHint')"
                :loading="cacheLoading"
                @click="handleClearCache"
              >
                <template #icon><t-icon name="clear" /></template>
                {{ t('common.clearCache') }}
              </t-button>
              <t-button theme="primary" @click="openCreateModal">
                <template #icon><t-icon name="add" /></template>
                {{ t('content.createEntry') }}
              </t-button>
              <!-- 批量操作 -->
              <div v-if="hasSelected" class="batch-actions">
                <span class="selected-count">{{ t('common.selectedCount', { count: selectedCount }) }}</span>
                <t-button variant="outline" :disabled="batchLoading" @click="confirmBatchAction('publish')">
                  {{ t('content.batchPublish') }}
                </t-button>
                <t-button variant="outline" :disabled="batchLoading" @click="confirmBatchAction('unpublish')">
                  {{ t('content.batchUnpublish') }}
                </t-button>
                <t-button theme="danger" variant="outline" :disabled="batchLoading" @click="confirmBatchAction('delete')">
                  {{ t('content.batchDelete') }}
                </t-button>
              </div>
            </template>
          </TableToolbar>

          <!-- 表格 — 保留原生 table（动态列复杂）但优化样式 -->
          <div class="card table-wrap">
            <table class="table">
              <thead>
                <tr>
                  <th class="checkbox-col">
                    <input
                      type="checkbox"
                      :checked="isAllSelected"
                      :indeterminate.prop="hasSelected && !isAllSelected"
                      @change="toggleSelectAll"
                    />
                  </th>
                  <th>{{ t('common.id') }}</th>
                  <th v-for="field in selectedType.fields?.slice(0, 1)" :key="field.id">
                    {{ field.label }}
                  </th>
                  <th>{{ t('common.status') }}</th>
                  <th>{{ t('common.updatedAt') }}</th>
                  <th>{{ t('common.actions') }}</th>
                </tr>
              </thead>
              <tbody>
                <!-- Loading 行 -->
                <tr v-if="loading">
                  <td colspan="7" class="loading-state">
                    <t-loading size="small" />
                    <span>{{ t('common.loading') }}</span>
                  </td>
                </tr>
                <!-- 空状态 -->
                <tr v-else-if="entries.length === 0">
                  <td colspan="7" class="empty-state">
                    <p class="empty-title">{{ t('content.noContent') }}</p>
                    <p class="empty-desc">{{ t('content.createFirst') }}</p>
                    <t-button theme="primary" @click="openCreateModal">
                      {{ t('content.createFirstEntry') }}
                    </t-button>
                  </td>
                </tr>
                <!-- 数据行 -->
                <tr v-else v-for="entry in entries" :key="entry.id" :class="{ selected: selectedIds.has(entry.id) }">
                  <td class="checkbox-col">
                    <input
                      type="checkbox"
                      :checked="selectedIds.has(entry.id)"
                      @change="toggleSelect(entry.id)"
                    />
                  </td>
                  <td class="id-cell">{{ entry.id }}</td>
                  <td v-for="field in selectedType.fields?.slice(0, 1)" :key="field.id">
                    {{ formatFieldValue(entry.values?.[field.name]) }}
                  </td>
                  <td>
                    <t-tag :theme="(getStatusTheme(entry) as any)" variant="light" size="small">
                      {{ getStatusLabel(entry) }}
                    </t-tag>
                  </td>
                  <td>{{ formatDate(entry.updated_time) }}</td>
                  <td class="actions-cell">
                    <t-button variant="outline" size="small" :disabled="publishLoading === entry.id" @click="openEditModal(entry)">{{ t('common.edit') }}</t-button>
                    <t-button
                      :theme="entry.status === 'published' ? 'default' : 'primary'"
                      size="small"
                      :disabled="publishLoading === entry.id"
                      :loading="publishLoading === entry.id"
                      @click="handlePublish(entry)"
                    >
                      {{ entry.status === 'published' ? t('content.unpublish') : t('content.publish') }}
                    </t-button>
                    <t-button theme="danger" variant="outline" size="small" @click="confirmDelete(entry)">{{ t('common.delete') }}</t-button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <!-- 分页 — t-pagination -->
          <div class="pagination-bar">
            <t-pagination
              v-model:current="page"
              v-model:pageSize="pageSize"
              :total="total"
              :show-page-size="true"
              :page-size-options="[10, 20, 50, 100]"
              size="small"
              @change="onPageChange"
            />
          </div>
        </template>

        <div v-else class="empty-state">
          <t-icon name="file" size="64px" style="opacity:0.3" />
          <h3>{{ t('content.selectType') }}</h3>
          <p>{{ t('content.selectTypeHint') }}</p>
        </div>
          </t-tab-panel>
          <t-tab-panel value="calendar" :label="t('schedule.calendarView')">
            <ScheduleCalendar
              :selected-type="selectedType"
              :content-schemas="contentSchemas"
              @edit-entry="openEditModal"
            />
          </t-tab-panel>
        </t-tabs>
      </main>
      </template>
    </div>

    <!-- 创建/编辑弹窗 — t-dialog + 动态 t-form 字段 -->
    <t-dialog
      v-model:visible="showModal"
      :header="editingEntry ? t('content.editEntry') : t('content.createEntry')"
      :width="640"
      :confirm-btn="{ content: submitting ? t('common.processing') : (editingEntry ? t('common.save') : t('common.create')), theme: 'primary' as const, loading: submitting }"
      :cancel-btn="{ content: t('common.cancel') }"
      :confirm-on-enter="!editingEntry"
      @confirm="handleSubmit"
      @close="closeModal"
    >
      <t-form :data="formData" label-align="top">
        <template v-for="field in selectedType?.fields" :key="field.id">
          <!-- text / email / url -->
          <t-form-item
            v-if="['text','email','url'].includes(field.field_type)"
            :label="`${field.label}${((field as any).required ? ' *' : '')}`"
          >
            <t-input
              v-model="formData[field.name]"
              :placeholder="t('content.enterField', { fieldName: field.label })"
              clearable
            />
          </t-form-item>

          <!-- rich_text -->
          <t-form-item
            v-else-if="field.field_type === 'rich_text'"
            :label="field.label"
          >
            <RichTextEditor
              v-model="formData[field.name]"
              :placeholder="t('content.enterField', { fieldName: field.label })"
            />
          </t-form-item>

          <!-- json -->
          <t-form-item
            v-else-if="field.field_type === 'json'"
            :label="field.label"
          >
            <JsonEditor
              v-model="formData[field.name]"
              :placeholder="t('content.enterField', { fieldName: field.label })"
            />
          </t-form-item>

          <!-- number -->
          <t-form-item
            v-else-if="field.field_type === 'number'"
            :label="`${field.label}${((field as any).required ? ' *' : '')}`"
          >
            <t-input-number
              v-model="formData[field.name]"
              theme="normal"
              :placeholder="t('content.enterNumber')"
            />
          </t-form-item>

          <!-- date -->
          <t-form-item
            v-else-if="field.field_type === 'date'"
            :label="field.label"
          >
            <t-date-picker
              v-model="formData[field.name]"
              :enable-time-picker="false"
            />
          </t-form-item>

          <!-- datetime -->
          <t-form-item
            v-else-if="field.field_type === 'datetime'"
            :label="field.label"
          >
            <t-date-picker
              v-model="formData[field.name]"
              enable-time-picker
            />
          </t-form-item>

          <!-- boolean -->
          <t-form-item
            v-else-if="field.field_type === 'boolean'"
            :label="field.label"
          >
            <t-switch v-model="formData[field.name]" />
          </t-form-item>

          <!-- enum -->
          <t-form-item
            v-else-if="field.field_type === 'enum'"
            :label="`${field.label}${((field as any).required ? ' *' : '')}`"
          >
            <t-select
              v-model="formData[field.name]"
              :placeholder="t('content.select')"
              :options="((field as any).options || field.config?.options || []).map((opt: string) => ({ label: opt, value: opt })) as any"
              clearable
            />
          </t-form-item>

          <!-- email -->
          <t-form-item
            v-else-if="field.field_type === 'email'"
            :label="`${field.label}${((field as any).required ? ' *' : '')}`"
          >
            <t-input
              v-model="formData[field.name]"
              type="email"
              :placeholder="t('content.enterEmail')"
              clearable
            />
          </t-form-item>

          <!-- url -->
          <t-form-item
            v-else-if="field.field_type === 'url'"
            :label="`${field.label}${((field as any).required ? ' *' : '')}`"
          >
            <t-input
              v-model="formData[field.name]"
              :placeholder="t('content.enterUrl')"
              clearable
            >
              <template #suffix>
                <t-button
                  v-if="formData[field.name]"
                  variant="text"
                  size="small"
                  @click="openLink(formData[field.name])"
                >
                  <t-icon name="link" />
                </t-button>
              </template>
            </t-input>
          </t-form-item>

          <!-- fallback: text input -->
          <t-form-item
            v-else
            :label="`${field.label}${((field as any).required ? ' *' : '')}`"
          >
            <t-input
              v-model="formData[field.name]"
              :placeholder="t('content.enterField', { fieldName: field.label })"
              clearable
            />
          </t-form-item>
        </template>
      </t-form>

      <!-- 排期设置 -->
      <t-divider>{{ t('schedule.title') }}</t-divider>
      <t-row :gutter="16">
        <t-col :span="6">
          <t-form-item :label="t('schedule.publishTime')">
            <t-date-picker
              v-model="scheduledPublishTime"
              enable-time-picker
              :placeholder="t('schedule.publishTime')"
              clearable
              style="width: 100%"
            >
              <template #suffixIcon>
                <t-button
                  v-if="scheduledPublishTime"
                  variant="text"
                  size="small"
                  @click="scheduledPublishTime = ''"
                >{{ t('schedule.clearPublish') }}</t-button>
              </template>
            </t-date-picker>
          </t-form-item>
        </t-col>
        <t-col :span="6">
          <t-form-item :label="t('schedule.unpublishTime')">
            <t-date-picker
              v-model="scheduledUnpublishTime"
              enable-time-picker
              :placeholder="t('schedule.unpublishTime')"
              clearable
              style="width: 100%"
            >
              <template #suffixIcon>
                <t-button
                  v-if="scheduledUnpublishTime"
                  variant="text"
                  size="small"
                  @click="scheduledUnpublishTime = ''"
                >{{ t('schedule.clearUnpublish') }}</t-button>
              </template>
            </t-date-picker>
          </t-form-item>
        </t-col>
      </t-row>
    </t-dialog>
</template>

<style scoped>
.content-management {
  height: 100%;
}

.content-layout {
  display: flex;
  gap: 20px;
  height: calc(100vh - 160px);
}

/* === Sidebar === */
.type-sidebar {
  width: 240px;
  background: var(--color-card);
  border: 1px solid var(--color-border);
  border-radius: 12px;
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
}

.sidebar-header {
  padding: 16px;
  border-bottom: 1px solid var(--color-border);
}

.sidebar-header h3 {
  font-size: 14px;
  font-weight: 600;
  color: var(--color-text-secondary);
}

.type-list {
  flex: 1;
  padding: 8px;
  overflow-y: auto;
}

.type-item {
  width: 100%;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  background: transparent;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  text-align: left;
  transition: all 0.2s;
  color: var(--color-text);
}

.type-item:hover {
  background: var(--color-hover);
}

.type-item.active {
  background: var(--color-primary-light);
  color: var(--color-primary);
}

.type-icon {
  display: flex;
  color: var(--color-text-secondary);
}

.type-item.active .type-icon {
  color: var(--color-primary);
}

.type-name {
  flex: 1;
  font-size: 14px;
  font-weight: 500;
}

.type-count {
  font-size: 12px;
  color: var(--color-text-secondary);
  background: var(--color-hover);
  padding: 2px 8px;
  border-radius: 10px;
}

.empty-tip {
  padding: 20px;
  text-align: center;
  font-size: 13px;
  color: var(--color-text-secondary);
}

.empty-tip a {
  color: var(--color-primary);
}

/* === Main content === */
.content-main {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
}

.schedule-tabs {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.schedule-tabs :deep(.t-tabs__content) {
  flex: 1;
  overflow: auto;
  margin: 16px;
}

.content-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
  flex-wrap: wrap;
  gap: 12px;
}

/* toolbar-left/right 已提取到 common.css */
.toolbar-left h2 {
  font-size: 18px;
  font-weight: 600;
  color: var(--color-text);
}

/* batch-actions — toolbar-right 已提取到 common.css */
.batch-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 12px;
  background: var(--color-primary-light);
  border-radius: 8px;
}

.selected-count {
  font-size: 13px;
  color: var(--color-primary);
  font-weight: 500;
  margin-right: 4px;
}

/* === Table === */
.table-wrap {
  padding: 0 !important;
  overflow: hidden;
}

.checkbox-col {
  width: 40px;
  text-align: center;
}

.checkbox-col input[type="checkbox"] {
  width: 16px;
  height: 16px;
  cursor: pointer;
  accent-color: var(--color-primary);
}

tr.selected {
  background: var(--color-primary-light) !important;
}

.id-cell {
  font-family: monospace;
  font-size: 13px;
  color: var(--color-text-secondary);
}

.actions-cell {
  display: flex;
  gap: 6px;
}

.loading-state {
  text-align: center;
  padding: 40px !important;
  color: var(--color-text-secondary);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
}

.empty-state {
  text-align: center;
  padding: 48px 24px !important;
}

.empty-title {
  font-size: 16px;
  font-weight: 500;
  color: var(--color-text);
  margin: 0 0 8px;
}

.empty-desc {
  font-size: 14px;
  color: var(--color-text-secondary);
  margin: 0 0 20px;
}
/* pagination-bar 已提取到 common.css */

/* === Form textarea === */
.entry-textarea {
  width: 100%;
  padding: 8px 10px;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  font-size: 14px;
  line-height: 1.5;
  color: var(--color-text);
  background: var(--color-bg-white);
  resize: vertical;
  box-sizing: border-box;
  font-family: inherit;
}

.entry-textarea:focus {
  outline: none;
  border-color: var(--color-primary);
  box-shadow: 0 0 0 2px rgba(var(--color-primary-rgb, 22, 119, 255), 0.15);
}

/* === No site === */
.no-site-container {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 48px;
}

.no-site-card {
  text-align: center;
  max-width: 400px;
  padding: 40px;
  background: var(--color-card);
  border: 1px solid var(--color-border);
  border-radius: 12px;
}

.no-site-card h3 {
  font-size: 18px;
  font-weight: 600;
  color: var(--color-text);
  margin: 16px 0 8px;
}

.no-site-card p {
  font-size: 14px;
  color: var(--color-text-secondary);
  margin: 0 0 24px;
}
</style>
