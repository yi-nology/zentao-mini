<template>
  <div class="scheduler-page">
    <div class="page-tabs">
      <button class="tab-btn" :class="{ active: activeTab === 'tasks' }" @click="activeTab = 'tasks'">任务列表</button>
      <button class="tab-btn" :class="{ active: activeTab === 'logs' }" @click="activeTab = 'logs'">执行日志</button>
      <div style="flex:1" />
      <button v-if="activeTab === 'tasks'" class="btn btn-primary" @click="openCreateDialog">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:16px;height:16px"><path d="M12 5v14M5 12h14" /></svg>
        新建任务
      </button>
    </div>

    <!-- Task List -->
    <div v-if="activeTab === 'tasks'" class="card">
      <div v-if="loading" class="empty-state">加载中...</div>
      <div v-else-if="tasks.length === 0" class="empty-state">暂无定时任务，点击右上角「新建任务」创建</div>
      <table v-else class="data-table">
        <thead>
          <tr>
            <th>任务名称</th>
            <th>Cron 表达式</th>
            <th>项目</th>
            <th>Webhooks</th>
            <th>状态</th>
            <th>上次执行</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="task in tasks" :key="task.id">
            <td class="task-name">{{ task.name }}</td>
            <td><code class="cron-code">{{ task.cronExpr }}</code></td>
            <td>{{ task.projectName || task.projectId }}</td>
            <td>
              <span v-for="wh in task.webhooks" :key="wh.id" class="webhook-tag" :class="{ disabled: !wh.enabled }">
                {{ wh.name }}
              </span>
              <span v-if="task.webhooks.length === 0" class="text-muted">无</span>
            </td>
            <td>
              <span class="status-badge" :class="task.enabled ? 'status-on' : 'status-off'" @click="handleToggle(task)">
                {{ task.enabled ? '已启用' : '已禁用' }}
              </span>
            </td>
            <td>
              <template v-if="task.lastRunAt">
                <div class="text-sm">{{ formatTime(task.lastRunAt) }}</div>
                <span class="run-status" :class="'run-' + task.lastRunStatus">{{ runStatusLabel(task.lastRunStatus) }}</span>
              </template>
              <span v-else class="text-muted">未执行</span>
            </td>
            <td class="actions-cell">
              <button class="btn btn-sm" @click="handleRunNow(task)" title="立即执行">▶</button>
              <button class="btn btn-sm" @click="openEditDialog(task)" title="编辑">✎</button>
              <button class="btn btn-sm btn-danger" @click="handleDelete(task)" title="删除">✕</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Logs -->
    <div v-if="activeTab === 'logs'" class="card">
      <div v-if="logsLoading" class="empty-state">加载中...</div>
      <div v-else-if="logs.length === 0" class="empty-state">暂无执行日志</div>
      <table v-else class="data-table">
        <thead>
          <tr>
            <th>任务名称</th>
            <th>执行时间</th>
            <th>耗时</th>
            <th>状态</th>
            <th>Bug数</th>
            <th>高级别</th>
            <th>Webhook结果</th>
            <th>错误信息</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="log in logs" :key="log.id">
            <td>{{ log.taskName }}</td>
            <td class="text-sm">{{ formatTime(log.startedAt) }}</td>
            <td class="text-sm">{{ log.finishedAt ? calcDuration(log.startedAt, log.finishedAt) : '-' }}</td>
            <td><span class="run-status" :class="'run-' + log.status">{{ runStatusLabel(log.status) }}</span></td>
            <td>{{ log.bugTotal }}</td>
            <td>{{ log.highSeverity }}</td>
            <td>
              <span v-for="wr in log.webhookResults" :key="wr.webhookId" class="wh-result" :class="wr.success ? 'wh-ok' : 'wh-fail'" :title="wr.webhookName + ': ' + (wr.error || 'OK')">
                {{ wr.success ? '✓' : '✗' }} {{ wr.webhookName }}
              </span>
              <span v-if="log.webhookResults.length === 0" class="text-muted">-</span>
            </td>
            <td class="error-cell">{{ log.error || '-' }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Create/Edit Dialog -->
    <div v-if="dialogVisible" class="dialog-overlay" @click.self="dialogVisible = false">
      <div class="dialog">
        <div class="dialog-header">
          <h3>{{ isEdit ? '编辑任务' : '新建任务' }}</h3>
          <button class="dialog-close" @click="dialogVisible = false">✕</button>
        </div>
        <div class="dialog-body">
          <div class="form-group">
            <label>任务名称 <span class="required">*</span></label>
            <input v-model="form.name" class="form-input" placeholder="如：每日Bug分布报告" />
          </div>
          <div class="form-row">
            <div class="form-group form-half">
              <label>产品 <span class="required">*</span></label>
              <select v-model="form.productId" class="form-input" @change="onProductChange">
                <option value="">请选择产品</option>
                <option v-for="p in products" :key="p.id" :value="p.id">{{ p.name }}</option>
              </select>
            </div>
            <div class="form-group form-half">
              <label>项目 <span class="required">*</span></label>
              <select v-model="form.projectId" class="form-input" @change="onProjectChange">
                <option value="">请选择项目</option>
                <option v-for="p in projects" :key="p.id" :value="p.id">{{ p.name }}</option>
              </select>
            </div>
          </div>
          <div class="form-group">
            <label>Cron 表达式 <span class="required">*</span></label>
            <input v-model="form.cronExpr" class="form-input" placeholder="0 9 * * 1-5" />
            <div class="cron-presets">
              <button v-for="preset in CRON_PRESETS" :key="preset.expr" class="preset-btn" @click="form.cronExpr = preset.expr">
                {{ preset.label }}
              </button>
            </div>
          </div>
          <div class="form-group">
            <label>Bug 状态过滤</label>
            <select v-model="form.statusFilter" class="form-input">
              <option v-for="opt in STATUS_OPTIONS" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
            </select>
          </div>
          <div class="form-group">
            <label>Webhook 列表 <span class="required">*</span></label>
            <div v-for="(wh, idx) in form.webhooks" :key="idx" class="webhook-block">
              <div class="webhook-row">
                <input v-model="wh.name" class="form-input wh-name" placeholder="名称" />
                <select v-model="wh.platform" class="form-input wh-platform">
                  <option value="generic">通用</option>
                  <option value="lanxin">蓝信</option>
                </select>
                <label class="wh-enable">
                  <input type="checkbox" v-model="wh.enabled" />
                  启用
                </label>
                <button class="btn btn-sm btn-danger" @click="form.webhooks.splice(idx, 1)">✕</button>
              </div>
              <div class="webhook-row" style="margin-top:6px">
                <input v-model="wh.url" class="form-input wh-url-full" placeholder="Webhook URL (https://...)" />
              </div>
              <div v-if="wh.platform === 'lanxin'" class="webhook-row" style="margin-top:6px">
                <input v-model="wh.secret" class="form-input wh-url-full" placeholder="蓝信签名密钥 (可选，留空则不加签)" />
              </div>
              <div style="margin-top:4px">
                <button class="btn btn-sm" @click="handleTestWebhook(wh.url)" :disabled="!wh.url">测试</button>
              </div>
            </div>
            <button class="btn btn-sm" @click="form.webhooks.push({ id: '', name: '', url: '', enabled: true, platform: 'generic', secret: '' })">+ 添加 Webhook</button>
            <div v-if="testResult" class="test-result" :class="testResult.success ? 'test-ok' : 'test-fail'">
              {{ testResult.success ? '连接成功' : '连接失败: ' + testResult.error }}
            </div>
          </div>
        </div>
        <div class="dialog-footer">
          <button class="btn" @click="dialogVisible = false">取消</button>
          <button class="btn btn-primary" @click="handleSubmit" :disabled="submitting">
            {{ submitting ? '提交中...' : '保存' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Run Result Dialog -->
    <div v-if="runResultVisible" class="dialog-overlay" @click.self="runResultVisible = false">
      <div class="dialog">
        <div class="dialog-header">
          <h3>手动执行结果</h3>
          <button class="dialog-close" @click="runResultVisible = false">✕</button>
        </div>
        <div class="dialog-body">
          <div v-if="runResult" class="run-result-info">
            <div class="result-row">
              <span class="result-label">状态：</span>
              <span class="run-status" :class="'run-' + runResult.status">{{ runStatusLabel(runResult.status) }}</span>
            </div>
            <div class="result-row">
              <span class="result-label">Bug 总数：</span>
              <span>{{ runResult.bugTotal }}</span>
            </div>
            <div class="result-row">
              <span class="result-label">高级别：</span>
              <span>{{ runResult.highSeverity }}</span>
            </div>
            <div class="result-row">
              <span class="result-label">指派人分组：</span>
              <span>{{ runResult.assigneeCount }} 人</span>
            </div>
            <div v-if="runResult.webhookResults.length > 0" class="result-webhooks">
              <div class="result-label">Webhook 结果：</div>
              <div v-for="wr in runResult.webhookResults" :key="wr.webhookId" class="wh-result-item">
                <span :class="wr.success ? 'wh-ok' : 'wh-fail'">{{ wr.success ? '✓' : '✗' }}</span>
                {{ wr.webhookName }} ({{ wr.webhookUrl }})
                <span v-if="wr.error" class="text-muted">- {{ wr.error }}</span>
              </div>
            </div>
            <div v-if="runResult.error" class="result-error">{{ runResult.error }}</div>
          </div>
        </div>
        <div class="dialog-footer">
          <button class="btn" @click="runResultVisible = false">关闭</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive, watch } from 'vue'
import {
  listTasks, createTask, updateTask, deleteTask, toggleTask,
  runTaskNow, getAllLogs, testWebhook
} from '@/api/scheduler'
import { getProducts, getProjects } from '@/api/zentao'
import type { SchedulerTask, TaskExecutionLog, WebhookResult, WebhookConfig } from '@/types/scheduler'
import { CRON_PRESETS, STATUS_OPTIONS } from '@/types/scheduler'

interface Product {
  id: number
  name: string
}

interface Project {
  id: number
  name: string
}

const activeTab = ref<'tasks' | 'logs'>('tasks')
const tasks = ref<SchedulerTask[]>([])
const logs = ref<TaskExecutionLog[]>([])
const loading = ref(false)
const logsLoading = ref(false)
const products = ref<Product[]>([])
const projects = ref<Project[]>([])

const dialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const testResult = ref<WebhookResult | null>(null)
const runResultVisible = ref(false)
const runResult = ref<TaskExecutionLog | null>(null)

const form = reactive<{
  id: string
  name: string
  productId: number | string
  projectId: number | string
  projectName: string
  productName: string
  cronExpr: string
  statusFilter: string
  webhooks: WebhookConfig[]
}>({
  id: '',
  name: '',
  productId: '',
  projectId: '',
  projectName: '',
  productName: '',
  cronExpr: '0 9 * * 1-5',
  statusFilter: 'active',
  webhooks: [{ id: '', name: '', url: '', enabled: true, platform: 'generic', secret: '' }],
})

onMounted(() => {
  loadTasks()
  loadProducts()
})

const loadTasks = async () => {
  loading.value = true
  try {
    const res = await listTasks()
    tasks.value = res.data || []
  } catch { tasks.value = [] }
  finally { loading.value = false }
}

const loadLogs = async () => {
  logsLoading.value = true
  try {
    const res = await getAllLogs()
    logs.value = res.data || []
  } catch { logs.value = [] }
  finally { logsLoading.value = false }
}

const loadProducts = async () => {
  try {
    const res = await getProducts()
    products.value = (res.data as unknown as Product[]) || []
  } catch { products.value = [] }
}

const onProductChange = async () => {
  form.projectId = ''
  projects.value = []
  if (!form.productId) return
  try {
    const res = await getProjects({ productId: Number(form.productId) })
    const data = res as unknown as { data: Project[] }
    projects.value = data.data || (Array.isArray(data) ? data as unknown as Project[] : [])
  } catch { projects.value = [] }
}

const onProjectChange = () => {
  const proj = projects.value.find(p => p.id === Number(form.projectId))
  if (proj) form.projectName = proj.name
  const prod = products.value.find(p => p.id === Number(form.productId))
  if (prod) form.productName = prod.name
}

const openCreateDialog = () => {
  isEdit.value = false
  form.id = ''
  form.name = ''
  form.productId = ''
  form.projectId = ''
  form.projectName = ''
  form.productName = ''
  form.cronExpr = '0 9 * * 1-5'
  form.statusFilter = 'active'
  form.webhooks = [{ id: '', name: '', url: '', enabled: true, platform: 'generic', secret: '' }]
  testResult.value = null
  dialogVisible.value = true
}

const openEditDialog = (task: SchedulerTask) => {
  isEdit.value = true
  form.id = task.id
  form.name = task.name
  form.productId = task.productId
  form.projectId = task.projectId
  form.projectName = task.projectName
  form.productName = task.productName
  form.cronExpr = task.cronExpr
  form.statusFilter = task.statusFilter
  form.webhooks = task.webhooks.map(w => ({ ...w }))
  testResult.value = null
  dialogVisible.value = true
  onProductChange()
}

const handleSubmit = async () => {
  if (!form.name || !form.cronExpr || !form.projectId || form.webhooks.length === 0) return
  submitting.value = true
  try {
    const payload = {
      ...form,
      productId: Number(form.productId),
      projectId: Number(form.projectId),
    }
    if (isEdit.value) {
      await updateTask(form.id, payload)
    } else {
      await createTask(payload)
    }
    dialogVisible.value = false
    await loadTasks()
  } catch (e) {
    console.error('保存失败', e)
  } finally {
    submitting.value = false
  }
}

const handleToggle = async (task: SchedulerTask) => {
  try {
    await toggleTask(task.id)
    await loadTasks()
  } catch (e) { console.error(e) }
}

const handleRunNow = async (task: SchedulerTask) => {
  try {
    const res = await runTaskNow(task.id)
    runResult.value = res.data
    runResultVisible.value = true
    await loadTasks()
  } catch (e) { console.error(e) }
}

const handleDelete = async (task: SchedulerTask) => {
  if (!confirm(`确定删除任务「${task.name}」？`)) return
  try {
    await deleteTask(task.id)
    await loadTasks()
  } catch (e) { console.error(e) }
}

const handleTestWebhook = async (url: string) => {
  if (!url) return
  try {
    const res = await testWebhook(url)
    testResult.value = res.data
  } catch (e: any) {
    testResult.value = { success: false, error: e.message || '请求失败', webhookId: '', webhookName: '', webhookUrl: url, statusCode: 0 }
  }
}

const formatTime = (t: string | null) => {
  if (!t) return '-'
  return new Date(t).toLocaleString('zh-CN')
}

const calcDuration = (start: string, end: string) => {
  const ms = new Date(end).getTime() - new Date(start).getTime()
  if (ms < 1000) return ms + 'ms'
  return (ms / 1000).toFixed(1) + 's'
}

const runStatusLabel = (s: string) => {
  const m: Record<string, string> = { running: '执行中', success: '成功', partial: '部分成功', failed: '失败' }
  return m[s] || s
}

watch(() => activeTab.value, (tab) => {
  if (tab === 'logs') loadLogs()
})
</script>

<style scoped>
.scheduler-page {
  max-width: 1200px;
}

.page-tabs {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 20px;
}

.tab-btn {
  padding: 8px 20px;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  background: var(--color-bg-card);
  color: var(--color-text-secondary);
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s;
}

.tab-btn:hover {
  border-color: var(--color-primary);
  color: var(--color-primary);
}

.tab-btn.active {
  background: var(--color-primary);
  color: #fff;
  border-color: var(--color-primary);
}

.card {
  background: var(--color-bg-card);
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-sm);
  overflow: hidden;
}

.empty-state {
  padding: 60px 20px;
  text-align: center;
  color: var(--color-text-tertiary);
}

.data-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}

.data-table th {
  padding: 12px 16px;
  text-align: left;
  font-weight: 600;
  color: var(--color-text-secondary);
  background: var(--color-bg-hover);
  border-bottom: 1px solid var(--color-border-light);
  white-space: nowrap;
}

.data-table td {
  padding: 10px 16px;
  border-bottom: 1px solid var(--color-border-light);
  color: var(--color-text-primary);
  vertical-align: middle;
}

.data-table tr:last-child td {
  border-bottom: none;
}

.task-name {
  font-weight: 500;
}

.cron-code {
  background: var(--color-bg-hover);
  padding: 2px 8px;
  border-radius: 4px;
  font-family: monospace;
  font-size: 12px;
}

.webhook-tag {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 11px;
  margin-right: 4px;
  background: var(--color-primary-light);
  color: var(--color-primary);
}

.webhook-tag.disabled {
  background: var(--color-bg-hover);
  color: var(--color-text-tertiary);
  opacity: 0.6;
}

.status-badge {
  display: inline-block;
  padding: 2px 10px;
  border-radius: 100px;
  font-size: 12px;
  cursor: pointer;
  transition: opacity 0.2s;
}

.status-badge:hover { opacity: 0.8; }
.status-on { background: #dcfce7; color: #16a34a; }
.status-off { background: #fee2e2; color: #dc2626; }

.run-status {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 500;
}

.run-success { background: #dcfce7; color: #16a34a; }
.run-failed { background: #fee2e2; color: #dc2626; }
.run-partial { background: #fef3c7; color: #d97706; }
.run-running { background: #dbeafe; color: #2563eb; }

.actions-cell {
  white-space: nowrap;
}

.text-sm { font-size: 12px; }
.text-muted { color: var(--color-text-tertiary); font-size: 12px; }

.wh-result {
  display: inline-block;
  margin-right: 6px;
  font-size: 11px;
}

.wh-ok { color: #16a34a; }
.wh-fail { color: #dc2626; }

.error-cell {
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--color-text-tertiary);
  font-size: 12px;
}

/* Dialog */
.dialog-overlay {
  position: fixed;
  top: 0; left: 0; right: 0; bottom: 0;
  background: rgba(0,0,0,0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.dialog {
  background: var(--color-bg-card);
  border-radius: var(--radius-md);
  width: 640px;
  max-height: 85vh;
  display: flex;
  flex-direction: column;
  box-shadow: var(--shadow-lg);
}

.dialog-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  border-bottom: 1px solid var(--color-border-light);
}

.dialog-header h3 {
  font-size: 16px;
  font-weight: 600;
  margin: 0;
}

.dialog-close {
  background: none;
  border: none;
  font-size: 18px;
  cursor: pointer;
  color: var(--color-text-tertiary);
}

.dialog-body {
  padding: 20px;
  overflow-y: auto;
  flex: 1;
}

.dialog-footer {
  padding: 12px 20px;
  border-top: 1px solid var(--color-border-light);
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

.form-group {
  margin-bottom: 16px;
}

.form-group label {
  display: block;
  font-size: 13px;
  font-weight: 500;
  color: var(--color-text-secondary);
  margin-bottom: 6px;
}

.required { color: #dc2626; }

.form-input {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  font-size: 13px;
  color: var(--color-text-primary);
  background: var(--color-bg);
  outline: none;
  transition: border-color 0.2s;
}

.form-input:focus {
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px rgba(79, 107, 246, 0.12);
}

.form-row {
  display: flex;
  gap: 12px;
}

.form-half { flex: 1; }

.cron-presets {
  display: flex;
  gap: 6px;
  margin-top: 6px;
  flex-wrap: wrap;
}

.preset-btn {
  padding: 3px 10px;
  border: 1px solid var(--color-border);
  border-radius: 4px;
  background: var(--color-bg);
  color: var(--color-text-secondary);
  font-size: 11px;
  cursor: pointer;
  transition: all 0.2s;
}

.preset-btn:hover {
  border-color: var(--color-primary);
  color: var(--color-primary);
}

.webhook-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 6px;
}

.webhook-block {
  border: 1px solid var(--color-border-light);
  border-radius: var(--radius-sm);
  padding: 10px;
  margin-bottom: 10px;
  background: var(--color-bg);
}

.wh-name { width: 120px !important; flex-shrink: 0; }
.wh-platform { width: 100px !important; flex-shrink: 0; }
.wh-url { flex: 1 !important; }
.wh-url-full { width: 100% !important; }

.wh-enable {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: var(--color-text-secondary);
  white-space: nowrap;
  cursor: pointer;
}

.test-result {
  margin-top: 8px;
  padding: 6px 12px;
  border-radius: 4px;
  font-size: 12px;
}

.test-ok { background: #dcfce7; color: #16a34a; }
.test-fail { background: #fee2e2; color: #dc2626; }

.btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  background: var(--color-bg-card);
  color: var(--color-text-secondary);
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
}

.btn:hover { border-color: var(--color-primary); color: var(--color-primary); }

.btn-primary {
  background: var(--color-primary);
  color: #fff;
  border-color: var(--color-primary);
}

.btn-primary:hover { opacity: 0.9; }

.btn-sm {
  padding: 4px 10px;
  font-size: 12px;
}

.btn-danger { color: #dc2626; }
.btn-danger:hover { background: #fee2e2; }

.btn:disabled { opacity: 0.4; cursor: not-allowed; }

/* Run result dialog */
.run-result-info {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.result-row {
  font-size: 14px;
}

.result-label {
  font-weight: 500;
  color: var(--color-text-secondary);
}

.result-webhooks {
  margin-top: 8px;
}

.wh-result-item {
  font-size: 13px;
  padding: 2px 0;
}

.result-error {
  margin-top: 8px;
  padding: 8px 12px;
  background: #fee2e2;
  color: #dc2626;
  border-radius: 4px;
  font-size: 13px;
}
</style>
