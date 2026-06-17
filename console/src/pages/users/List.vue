<template>
    <!-- 页面标题 -->
    <PageHeader
      :title="t('users.title')"
      :subtitle="t('users.subtitle')"
      :show-refresh="true"
      @refresh="loadUsers"
    >
      <template #primary-action>
        <t-button theme="primary" @click="openCreateDialog">
          <template #icon><t-icon name="add" /></template>
          {{ t('users.addUser') }}
        </t-button>
      </template>
    </PageHeader>

    <!-- 用户列表 — t-table -->
    <t-table
      :data="users"
      :columns="columns"
      :loading="loading"
      :pagination="{ current: pagination.current, total: pagination.total, pageSize: pagination.pageSize, showPageSize: false }"
      row-key="id"
      @page-change="onPageChange"
      hover
      stripe
      size="medium"
    >
      <template #role="{ row }">
        <t-tag v-if="userRoleMap[row.id]" theme="primary" variant="light" size="small">{{ userRoleMap[row.id] }}</t-tag>
        <t-tag v-else-if="row.is_super_admin" theme="warning" variant="light" size="small">{{ t('users.superAdmin') }}</t-tag>
        <t-tag v-else variant="light" size="small">{{ t('users.normalUser') }}</t-tag>
      </template>
      <template #status="{ row }">
        <t-tag :theme="(getStatusBadge(row.status || 'unknown') as any)" variant="light" size="small">
          {{ row.status === 'active' ? t('users.statusActive') : row.status === 'inactive' ? t('users.statusInactive') : row.status === 'suspended' ? t('users.statusBanned') : row.status }}
        </t-tag>
      </template>
      <template #operations="{ row }">
        <t-space size="small">
          <t-button variant="outline" size="small" @click="openViewDialog(row)">{{ t('common.view') }}</t-button>
          <t-button variant="outline" size="small" @click="openEditDialog(row)">{{ t('common.edit') }}</t-button>
          <t-button variant="outline" size="small" theme="warning" :loading="signingId === row.id" @click="handleSign(row)">{{ t('users.sign') }}</t-button>
          <t-button variant="outline" size="small" theme="success" :loading="verifyingId === row.id" @click="handleVerify(row)">{{ t('users.verify') }}</t-button>
          <t-button v-if="userStore.user?.is_super_admin && row.id !== userStore.user?.id" variant="outline" size="small" @click="openResetPasswordDialog(row)">{{ t('users.resetPassword') }}</t-button>
          <t-button theme="danger" variant="outline" size="small" :disabled="row.is_super_admin" @click="handleDelete(row)">{{ t('common.delete') }}</t-button>
        </t-space>
      </template>
      <template #empty>
        <div class="empty-state">
          <p class="empty-title">{{ t('users.noUsers') }}</p>
          <p class="empty-desc">{{ t('users.noUsersHint') }}</p>
        </div>
      </template>
    </t-table>

    <!-- 创建用户弹窗 — t-dialog + t-form -->
    <t-dialog
      v-model:visible="createVisible"
      :header="t('users.createTitle')"
      :width="480"
      :confirm-btn="{ content: creating ? t('common.creating') : t('common.create'), theme: 'primary' as const, loading: creating }"
      :cancel-btn="{ content: t('common.cancel') }"
      @confirm="handleCreate"
    >
      <t-form :data="createForm" label-align="top">
        <t-form-item :label="`${t('users.email')} *`">
          <t-input v-model="createForm.email" :placeholder="t('users.enterEmail')" clearable />
          <template v-if="createError" #help>
            {{ createError }}
          </template>
        </t-form-item>
        <t-form-item :label="`${t('users.password')} *`">
          <t-input v-model="createForm.password" :placeholder="t('users.enterPassword')" clearable />
          <!-- 密码强度条 -->
          <div class="password-strength">
            <div class="strength-bar">
              <div class="strength-fill" :class="passwordStrength.level" :style="{ width: passwordStrength.width }"></div>
            </div>
            <span class="strength-text" :class="passwordStrength.level">{{ passwordStrength.label }}</span>
          </div>
          <p class="password-hint">{{ t('users.passwordHint') }}</p>
        </t-form-item>
        <t-form-item :label="t('users.nickname')">
          <t-input v-model="createForm.nickname" :placeholder="t('users.enterNickname')" clearable />
        </t-form-item>
        <t-form-item :label="t('users.superAdminSwitch')">
          <t-switch v-model="createForm.is_super_admin" />
        </t-form-item>
        <t-form-item :label="t('users.assignRoles')">
          <t-select
            v-model="createForm.roleIds"
            :placeholder="t('users.selectRoles')"
            multiple
            :options="systemRoles"
            :keys="{ label: 'name', value: 'id' }"
            clearable
          />
        </t-form-item>
        <t-alert v-if="createError && !createError.includes('@')" theme="error" :message="createError" closable @close="createError = ''" />
      </t-form>
    </t-dialog>

    <!-- 编辑用户弹窗 — t-dialog + t-form -->
    <t-dialog
      v-model:visible="editVisible"
      :header="t('users.editTitle')"
      :width="480"
      :confirm-btn="{ content: updating ? t('common.saving') : t('common.save'), theme: 'primary' as const, loading: updating }"
      :cancel-btn="{ content: t('common.cancel') }"
      @confirm="handleUpdate"
    >
      <t-form :data="editForm" label-align="top">
        <t-form-item :label="t('users.email')">
          <t-input v-model="editForm.email" disabled />
        </t-form-item>
        <t-form-item :label="t('users.nickname')">
          <t-input v-model="editForm.nickname" :placeholder="t('users.enterNickname')" clearable />
        </t-form-item>
        <t-form-item :label="t('users.status')">
          <t-select v-model="editForm.status" :options="[
            { label: t('users.statusActive'), value: 'active' },
            { label: t('users.statusInactive'), value: 'inactive' },
            { label: t('users.statusBanned'), value: 'suspended' },
          ]" />
        </t-form-item>
        <t-form-item :label="t('users.superAdminSwitch')">
          <t-switch v-model="editForm.is_super_admin" :disabled="editForm.is_super_admin" />
        </t-form-item>
        <t-form-item :label="t('users.assignRoles')">
          <t-select
            v-model="editForm.roleIds"
            :placeholder="t('users.selectRoles')"
            multiple
            :options="systemRoles"
            :keys="{ label: 'name', value: 'id' }"
            clearable
          />
        </t-form-item>
        <t-alert v-if="editError" theme="error" :message="editError" closable @close="editError = ''" />
      </t-form>
    </t-dialog>

    <!-- 查看用户弹窗 — 只读展示 -->
    <t-dialog
      v-model:visible="viewVisible"
      :header="t('users.viewTitle')"
      :width="480"
      :footer="false"
    >
      <div class="user-detail">
        <div class="user-detail__avatar">
          {{ viewUser?.email?.charAt(0).toUpperCase() }}
        </div>
        <div class="user-detail__info">
          <div class="user-detail__row">
            <span class="user-detail__label">{{ t('users.email') }}</span>
            <span class="user-detail__value">{{ viewUser?.email }}</span>
          </div>
          <div class="user-detail__row">
            <span class="user-detail__label">{{ t('users.nickname') }}</span>
            <span class="user-detail__value">{{ viewUser?.nickname || '\u2014' }}</span>
          </div>
          <div class="user-detail__row">
            <span class="user-detail__label">{{ t('users.status') }}</span>
            <t-tag v-if="viewUser" :theme="getStatusBadge(viewUser.status)" variant="light" size="small">
              {{ statusLabel(viewUser.status) }}
            </t-tag>
          </div>
          <div class="user-detail__row">
            <span class="user-detail__label">{{ t('users.role') }}</span>
            <div class="user-detail__roles">
              <template v-if="viewUserRoles.length > 0">
                <t-tag v-for="r in viewUserRoles" :key="r.id" theme="primary" variant="light" size="small">
                  {{ r.name }}
                </t-tag>
              </template>
              <t-tag v-else-if="viewUser && viewUser.is_super_admin" theme="warning" variant="light" size="small">
                {{ t('users.superAdmin') }}
              </t-tag>
              <t-tag v-else-if="viewUser" variant="light" size="small">{{ t('users.normalUser') }}</t-tag>
            </div>
          </div>
          <div class="user-detail__row">
            <span class="user-detail__label">{{ t('users.createdTime') }}</span>
            <span class="user-detail__value">{{ formatDate(viewUser?.created_time || '') }}</span>
          </div>
        </div>
      </div>
    </t-dialog>

    <!-- 重置密码弹窗 -->
    <t-dialog
      v-model:visible="resetPwdVisible"
      :header="t('users.resetPasswordTitle')"
      :width="480"
      :confirm-btn="{ content: resetting ? t('common.saving') : t('common.confirm'), theme: 'primary' as const, loading: resetting }"
      :cancel-btn="{ content: t('common.cancel') }"
      @confirm="handleResetPassword"
    >
      <t-form label-align="top">
        <t-form-item :label="t('users.newPassword')">
          <t-input v-model="resetPwdForm.newPassword" :placeholder="t('users.enterPassword')" clearable />
          <!-- 密码强度条 -->
          <div class="password-strength">
            <div class="strength-bar">
              <div class="strength-fill" :class="resetPasswordStrength.level" :style="{ width: resetPasswordStrength.width }"></div>
            </div>
            <span class="strength-text" :class="resetPasswordStrength.level">{{ resetPasswordStrength.label }}</span>
          </div>
          <p class="password-hint">{{ t('users.passwordHint') }}</p>
        </t-form-item>
        <t-form-item :label="t('users.confirmPassword')">
          <t-input v-model="resetPwdForm.confirmPassword" :placeholder="t('users.confirmPasswordHint')" clearable />
        </t-form-item>
        <t-alert v-if="resetPwdError" theme="error" :message="resetPwdError" closable @close="resetPwdError = ''" />
      </t-form>
    </t-dialog>
</template>

<script setup lang="ts">

// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

import { ref, reactive, computed, onMounted, h } from 'vue'
import { useI18n } from 'vue-i18n'
import { DialogPlugin } from 'tdesign-vue-next'
import { useUserStore } from '@/stores/user'
import { showError, showSuccess } from '@/utils/request'

function handleError(err: unknown) {
  if (err instanceof Error) {
    showError(err.message)
  } else {
    showError(String(err))
  }
}
import PageHeader from '@/components/PageHeader.vue'
import { resetPassword } from '@/api/user'
import { signUser, verifyUser } from '@/api/user'
import { listSystemRoles, getUserRoles, assignUserRole, removeUserRole } from '@/api/rbac'
import type { SystemRole } from '@/api/rbac'

const { t } = useI18n()

const userStore = useUserStore()

interface User {
  id: string
  email: string
  nickname?: string
  avatar_url?: string
  status: string
  is_super_admin: boolean
  created_time: string
}

const users = ref<User[]>([])
const loading = ref(false)
const systemRoles = ref<SystemRole[]>([])
// 用户ID → 角色名映射
const userRoleMap = ref<Record<string, string>>({})
// 用户ID → 角色列表映射（用于查看详情）
const userRolesMap = ref<Record<string, SystemRole[]>>({})
const pagination = reactive({
  current: 1,
  pageSize: 20,
  total: 0,
})

// 创建弹窗
const createVisible = ref(false)
const creating = ref(false)
const createError = ref('')
const createForm = reactive({
  email: '',
  password: '',
  nickname: '',
  is_super_admin: false,
  roleIds: [] as string[],
})

// 密码强度计算
const passwordStrength = computed(() => {
  const pwd = createForm.password
  if (!pwd) return { level: '', width: '0%', label: '' }

  let score = 0
  if (pwd.length >= 8) score++
  if (pwd.length >= 12) score++
  if (/[a-z]/.test(pwd)) score++
  if (/[A-Z]/.test(pwd)) score++
  if (/[0-9]/.test(pwd)) score++
  if (/[^a-zA-Z0-9]/.test(pwd)) score++

  if (score < 3) return { level: 'weak', width: '33%', label: t('users.passwordWeak') }
  if (score < 5) return { level: 'medium', width: '66%', label: t('users.passwordMedium') }
  return { level: 'strong', width: '100%', label: t('users.passwordStrong') }
})

// 编辑弹窗
const editVisible = ref(false)
const updating = ref(false)
const editError = ref('')
const editingUserId = ref('')
const editForm = reactive({
  email: '',
  nickname: '',
  status: 'active' as string,
  is_super_admin: false,
  roleIds: [] as string[],
})

// 查看弹窗
const viewVisible = ref(false)
const viewUser = ref<User | null>(null)
const viewUserRoles = ref<SystemRole[]>([])

// 重置密码弹窗
const resetPwdVisible = ref(false)
const resetting = ref(false)
const resetPwdError = ref('')
const resetTargetUserId = ref('')
const signingId = ref('')
const verifyingId = ref('')
const resetPwdForm = reactive({
  newPassword: '',
  confirmPassword: '',
})

// 重置密码对话框的密码强度计算
const resetPasswordStrength = computed(() => {
  const pwd = resetPwdForm.newPassword
  if (!pwd) return { level: '', width: '0%', label: '' }

  let score = 0
  if (pwd.length >= 8) score++
  if (pwd.length >= 12) score++
  if (/[a-z]/.test(pwd)) score++
  if (/[A-Z]/.test(pwd)) score++
  if (/[0-9]/.test(pwd)) score++
  if (/[^a-zA-Z0-9]/.test(pwd)) score++

  if (score < 3) return { level: 'weak', width: '33%', label: t('users.passwordWeak') }
  if (score < 5) return { level: 'medium', width: '66%', label: t('users.passwordMedium') }
  return { level: 'strong', width: '100%', label: t('users.passwordStrong') }
})

/** 状态值 → i18n 标签文本 */
const statusLabel = (status: string) => {
  const map: Record<string, string> = {
    active: t('users.statusActive'),
    inactive: t('users.statusInactive'),
    suspended: t('users.statusBanned'),
  }
  return map[status] || status
}

const loadUsers = async () => {
  loading.value = true
  try {
    const result = await userStore.listUsers(pagination.current, pagination.pageSize)
    if (result) {
      users.value = result.items
      pagination.total = result.total
      // 加载每个用户的系统角色
      for (const u of result.items) {
        try {
          const roles = await getUserRoles(u.id)
          const roleData = roles.data || []
          userRolesMap.value[u.id] = roleData
          if (roleData.length > 0) {
            userRoleMap.value[u.id] = roleData.map(r => r.name).join(', ')
          }
        } catch {
          // 静默失败
        }
      }
    }
  } catch (error) {
    handleError(error)
  } finally {
    loading.value = false
  }
}

// 加载系统角色并构建用户-角色映射
const loadSystemRoles = async () => {
  try {
    const res = await listSystemRoles()
    systemRoles.value = res.data || []
  } catch {
    // 静默失败，角色列降级为仅显示 super_admin 状态
  }
}

const onPageChange = ({ current, pageSize }: { current: number; pageSize: number }) => {
  pagination.current = current
  pagination.pageSize = pageSize
  loadUsers()
}

const formatDate = (dateStr: string) => {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString()
}

const getStatusBadge = (status: string): 'success' | 'warning' | 'danger' | 'default' => {
  const map: Record<string, 'success' | 'warning' | 'danger' | 'default'> = {
    active: 'success',
    inactive: 'warning',
    suspended: 'danger',
  }
  return map[status] || 'default'
}

/**
 * 密码强度检查：至少8位，包含大小写字母与数字
 */
const checkPasswordStrength = (pwd: string): boolean => {
  if (pwd.length < 8) return false
  if (!/[a-z]/.test(pwd)) return false
  if (!/[A-Z]/.test(pwd)) return false
  if (!/[0-9]/.test(pwd)) return false
  return true
}

const openCreateDialog = () => {
  createForm.email = ''
  createForm.password = ''
  createForm.nickname = ''
  createForm.is_super_admin = false
  createForm.roleIds = []
  createError.value = ''
  createVisible.value = true
}

const openViewDialog = async (user: User) => {
  viewUser.value = user
  viewUserRoles.value = userRolesMap.value[user.id] || []
  viewVisible.value = true
}

const handleCreate = async () => {
  if (!createForm.email) { createError.value = t('users.emailRequired'); return }
  if (!checkPasswordStrength(createForm.password)) { createError.value = t('users.passwordWeak'); return }
  creating.value = true
  createError.value = ''
  try {
    const newUser = await userStore.createUser({
      email: createForm.email,
      password: createForm.password,
      nickname: createForm.nickname || undefined,
      is_super_admin: createForm.is_super_admin,
    })
    showSuccess(t('users.createSuccess'))
    // 分配选中的角色
    if (createForm.roleIds.length > 0 && newUser?.data?.id) {
      for (const roleId of createForm.roleIds) {
        try { await assignUserRole(newUser.data.id, roleId) } catch { /* 逐个失败不影响 */ }
      }
    }
    createVisible.value = false
    loadUsers()
  } catch (error: any) {
    const msg = error?.response?.data?.msg || t('users.createFailed')
    createError.value = msg.includes('already exists') ? t('users.emailTaken') : msg
  } finally {
    creating.value = false
  }
}

const openEditDialog = async (user: User) => {
  editingUserId.value = user.id
  editForm.email = user.email
  editForm.nickname = user.nickname || ''
  editForm.status = user.status
  editForm.is_super_admin = user.is_super_admin
  // 加载用户当前角色
  const currentRoles = userRolesMap.value[user.id] || []
  editForm.roleIds = currentRoles.map(r => r.id)
  editError.value = ''
  editVisible.value = true
}

const handleUpdate = async () => {
  updating.value = true
  editError.value = ''
  try {
    await userStore.updateUser(editingUserId.value, {
      nickname: editForm.nickname || undefined,
      status: editForm.status,
      is_super_admin: editForm.is_super_admin,
    })
    // 同步角色：计算差异后增删
    const currentRoles = userRolesMap.value[editingUserId.value] || []
    const currentRoleIds = new Set(currentRoles.map(r => r.id))
    const newRoleIds = new Set(editForm.roleIds)
    // 需要新增的
    for (const roleId of editForm.roleIds) {
      if (!currentRoleIds.has(roleId)) {
        try { await assignUserRole(editingUserId.value, roleId) } catch { /* */ }
      }
    }
    // 需要移除的
    for (const r of currentRoles) {
      if (!newRoleIds.has(r.id)) {
        try { await removeUserRole(editingUserId.value, r.id) } catch { /* */ }
      }
    }
    showSuccess(t('users.updateSuccess'))
    editVisible.value = false
    loadUsers()
  } catch (error: any) {
    editError.value = error?.response?.data?.msg || t('users.updateFailed')
  } finally {
    updating.value = false
  }
}

const handleDelete = (user: User) => {
  const dlg = DialogPlugin.confirm({
    header: t('users.confirmDeleteUser'),
    body: t('users.deleteConfirmMsg', { email: user.email }),
    theme: 'warning',
    onConfirm: async () => {
      try {
        await userStore.deleteUser(user.id)
        showSuccess(t('users.deleteSuccess'))
        loadUsers()
        dlg.destroy()
      } catch (error) {
        handleError(error)
      }
    },
  })
}

const openResetPasswordDialog = (user: User) => {
  resetTargetUserId.value = user.id
  resetPwdForm.newPassword = ''
  resetPwdForm.confirmPassword = ''
  resetPwdError.value = ''
  resetPwdVisible.value = true
}

const handleResetPassword = async () => {
  if (!resetPwdForm.newPassword) { resetPwdError.value = t('users.passwordRequired'); return }
  if (resetPwdForm.newPassword !== resetPwdForm.confirmPassword) { resetPwdError.value = t('users.passwordMismatch'); return }
  if (!checkPasswordStrength(resetPwdForm.newPassword)) { resetPwdError.value = t('users.passwordWeak'); return }

  resetting.value = true
  resetPwdError.value = ''
  try {
    await resetPassword(resetTargetUserId.value, resetPwdForm.newPassword)
    showSuccess(t('users.resetPasswordSuccess'))
    resetPwdVisible.value = false
  } catch (error: any) {
    resetPwdError.value = error?.response?.data?.msg || t('users.resetPasswordFailed')
  } finally {
    resetting.value = false
  }
}

const handleSign = async (user: User) => {
  signingId.value = user.id
  try {
    await signUser(user.id)
    showSuccess(t('users.signSuccess'))
  } catch (error: any) {
    showError(error?.response?.data?.msg || t('users.signFailed'))
  } finally {
    signingId.value = ''
  }
}

const handleVerify = async (user: User) => {
  verifyingId.value = user.id
  try {
    const res: any = await verifyUser(user.id)
    const result = res.data
    DialogPlugin.alert({
      header: t('users.verifyResult'),
      body: () => h('div', { style: 'line-height:2' }, [
        h('p', [
          h('strong', t('users.verifyStatus') + '：'),
          h('span', { style: `color:${result.valid ? 'var(--td-success-color)' : 'var(--td-error-color)'};font-weight:bold` },
            result.valid ? t('users.verifyPass') : t('users.verifyFail')),
        ]),
        h('p', [h('strong', t('users.verifyAlgorithm') + '：'), result.algorithm || '-']),
        h('p', [h('strong', t('users.verifySignature') + '：'), h('code', { style: 'font-size:12px;word-break:break-all' }, result.signature || '-')]),
        result.reason ? h('p', { style: 'color:var(--td-error-color)' }, result.reason) : null,
      ]),
      confirmBtn: t('common.close'),
    })
  } catch (error: any) {
    showError(error?.response?.data?.msg || t('users.verifyFailed'))
  } finally {
    verifyingId.value = ''
  }
}

// t-table columns
const columns = computed(() => [
  {
    colKey: 'user',
    title: t('users.user'),
    cell: (_h: any, { row }: { row: User }) => h('div', { class: 'user-info' }, [
      h('div', { class: 'user-avatar' }, row.email?.charAt(0).toUpperCase()),
      h('span', { class: 'user-name' }, row.nickname || '\u2014'),
    ]),
  },
  { colKey: 'email', title: t('users.email') },
  {
    colKey: 'role',
    title: t('users.role'),
    width: 250,
  },
  {
    colKey: 'status',
    title: t('users.status'),
    width: 80,
  },
  { colKey: 'created_time', title: t('users.createdTime'), width: 170, cell: (_h: any, { row }: { row: User }) => formatDate(row.created_time) },
  {
    colKey: 'operations',
    title: t('users.actions'),
    width: 440,
  },
])

onMounted(() => {
  loadUsers()
  loadSystemRoles()
})
</script>

<!-- 非 scoped：供 h() 渲染的表格单元格使用 -->
<style>
.user-info {
  display: flex;
  align-items: center;
  gap: 10px;
}

.user-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: var(--color-primary);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  font-weight: 600;
  flex-shrink: 0;
}

.user-name {
  font-weight: 500;
  color: var(--color-text);
}
</style>

<style scoped>
/* 页面特有样式：密码强度 */

/* === Password strength === */
.password-strength {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 6px;
}

.strength-bar {
  flex: 1;
  height: 4px;
  background: var(--color-border);
  border-radius: 2px;
  overflow: hidden;
}

.strength-fill {
  height: 100%;
  transition: width 0.3s, background-color 0.3s;
}

.strength-fill.weak { background: #f85149; }
.strength-fill.medium { background: #e3b341; }
.strength-fill.strong { background: #3fb950; }

.strength-text {
  font-size: 12px;
  min-width: 36px;
}

.strength-text.weak { color: #f85149; }
.strength-text.medium { color: #e3b341; }
.strength-text.strong { color: #3fb950; }

.password-hint {
  margin-top: 4px;
  font-size: 12px;
  color: var(--color-text-secondary);
}

/* === User detail view === */
.user-detail {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 20px;
}

.user-detail__avatar {
  width: 64px;
  height: 64px;
  border-radius: 50%;
  background: var(--color-primary);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  font-weight: 600;
  flex-shrink: 0;
}

.user-detail__info {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.user-detail__row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.user-detail__label {
  width: 80px;
  flex-shrink: 0;
  font-size: 13px;
  color: var(--color-text-secondary);
}

.user-detail__value {
  font-size: 13px;
  color: var(--color-text);
  font-weight: 500;
}
</style>
