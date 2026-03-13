<template>
  <div class="page-container">
    <!-- 筛选器 -->
    <el-card class="filter-card">
      <el-form :inline="true" :model="filterForm" class="filter-form">
        <el-form-item label="执行/迭代">
          <el-select
            v-model="filterForm.execution"
            placeholder="请选择执行/迭代"
            clearable
            style="width: 220px"
          >
            <el-option
              v-for="item in executionOptions"
              :key="item.id"
              :label="item.name"
              :value="item.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="指派人">
          <el-select
            v-model="filterForm.assignedTo"
            placeholder="请选择或输入指派人"
            clearable
            filterable
            style="width: 180px"
          >
            <el-option
              v-for="item in assignedToOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select
            v-model="filterForm.status"
            placeholder="请选择状态"
            clearable
            style="width: 160px"
          >
            <el-option
              v-for="item in statusOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="时间范围">
          <el-date-picker
            v-model="filterForm.dateRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            style="width: 300px"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">
            <el-icon><Search /></el-icon>查询
          </el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 数据表格 -->
    <el-card class="table-card">
      <el-table
        v-loading="loading"
        :data="taskList"
        border
        stripe
        style="width: 100%"
      >
        <el-table-column prop="id" label="ID" width="80" align="center" />
        <el-table-column prop="name" label="标题" min-width="200">
          <template #default="{ row }">
            <a href="javascript:void(0)" @click="openZentaoTask(row.id)" class="task-title" show-overflow-tooltip>{{ row.name }}</a>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">
              {{ getStatusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="指派给" width="120" align="center">
          <template #default="{ row }">
            {{ row.assignedTo?.realname || row.assignedTo?.account || '未指派' }}
          </template>
        </el-table-column>
        <el-table-column prop="estimate" label="预估工时" width="100" align="center">
          <template #default="{ row }">
            {{ row.estimate || 0 }}h
          </template>
        </el-table-column>
        <el-table-column prop="consumed" label="消耗工时" width="100" align="center">
          <template #default="{ row }">
            {{ row.consumed || 0 }}h
          </template>
        </el-table-column>
        <el-table-column label="进度" width="120" align="center">
          <template #default="{ row }">
            <el-progress
              :percentage="getProgress(row.estimate, row.consumed)"
              :status="getProgressStatus(row.estimate, row.consumed)"
              :stroke-width="10"
            />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="100" align="center">
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="openTaskDetail(row)">查看</el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="pagination.total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handlePageChange"
        />
      </div>
    </el-card>
  </div>

  <!-- 任务详情弹窗 -->
  <el-dialog
    v-model="detailDialogVisible"
    :title="`任务详情 - ID: ${currentTask?.id}`"
    width="80%"
    destroy-on-close
  >
    <div v-if="currentTask" class="task-detail">
      <el-descriptions :column="1" border>
        <el-descriptions-item label="ID">{{ currentTask.id }}</el-descriptions-item>
        <el-descriptions-item label="标题">{{ currentTask.name }}</el-descriptions-item>
        <el-descriptions-item label="状态">{{ getStatusLabel(currentTask.status) }}</el-descriptions-item>
        <el-descriptions-item label="指派给">{{ currentTask.assignedTo?.realname || currentTask.assignedTo?.account || '未指派' }}</el-descriptions-item>
        <el-descriptions-item label="预估工时">{{ currentTask.estimate || 0 }}h</el-descriptions-item>
        <el-descriptions-item label="消耗工时">{{ currentTask.consumed || 0 }}h</el-descriptions-item>
        <el-descriptions-item label="进度">{{ getProgress(currentTask.estimate, currentTask.consumed) }}%</el-descriptions-item>
        <el-descriptions-item label="描述" :span="2">
          <div v-html="currentTask.desc || '无'" />
        </el-descriptions-item>
      </el-descriptions>
      <div class="dialog-actions">
        <el-button @click="openZentaoTask(currentTask.id)">在禅道中查看</el-button>
      </div>
    </div>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, inject, watch, computed } from 'vue'
import { Search } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import {
  getExecutions,
  getTasks,
  getTaskStatusOptions,
  getUsers
} from '@/api/zentao'
import type { Task, User, Execution, SelectOption } from '@/types/api'

import * as runtime from '@wailsjs/runtime/runtime'

interface GlobalSelection {
  product: number | null
  project: number | null
  execution: number | null
}

interface FilterForm {
  execution: number | null
  assignedTo: string
  status: string
  dateRange: [string, string] | []
}

interface Pagination {
  page: number
  pageSize: number
  total: number
}

const globalSelection = inject<GlobalSelection>('globalSelection')!

const filterForm = reactive<FilterForm>({
  execution: null,
  assignedTo: '',
  status: '',
  dateRange: []
})

const executionOptions = ref<Execution[]>([])
const statusOptions = ref<SelectOption[]>(getTaskStatusOptions())
const userOptions = ref<User[]>([])

const taskList = ref<Task[]>([])
const loading = ref<boolean>(false)

const detailDialogVisible = ref<boolean>(false)
const currentTask = ref<Task | null>(null)

const pagination = reactive<Pagination>({
  page: 1,
  pageSize: 20,
  total: 0
})

const assignedToOptions = computed(() => {
  const assignees = new Map<string, { value: string; label: string }>()
  userOptions.value.forEach((user: User) => {
    if (user.account) {
      assignees.set(user.account, {
        value: user.account,
        label: user.realname || user.account
      })
    }
  })
  return Array.from(assignees.values()).sort((a, b) => a.label.localeCompare(b.label))
})

const fetchExecutions = async (): Promise<void> => {
  try {
    const params: { projectID?: number; productID?: number } = {}
    if (globalSelection.project) {
      params.projectID = globalSelection.project
    } else if (globalSelection.product) {
      params.productID = globalSelection.product
    }
    const res = await getExecutions(params)
    executionOptions.value = res.data || []
  } catch (error) {
    console.error('获取执行/迭代列表失败:', error)
  }
}

const fetchUsers = async (): Promise<void> => {
  try {
    const users = await getUsers()
    userOptions.value = users || []
  } catch (error) {
    console.error('获取用户列表失败:', error)
    ElMessage.error('获取用户列表失败，请刷新页面重试')
  }
}

watch(() => [globalSelection.product, globalSelection.project], () => {
  filterForm.execution = null
  fetchExecutions()
}, { deep: true })

const fetchTasks = async (): Promise<void> => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      pageSize: pagination.pageSize,
      executionID: filterForm.execution ?? undefined,
      assignedTo: filterForm.assignedTo,
      status: filterForm.status,
      startDate: filterForm.dateRange[0] || '',
      endDate: filterForm.dateRange[1] || ''
    }
    const res = await getTasks(params)
    const data = res.data
    taskList.value = Array.isArray(data) ? data : []
    pagination.total = Array.isArray(data) ? data.length : 0
  } catch (error) {
    console.error('获取任务列表失败:', error)
    ElMessage.error('获取任务列表失败')
  } finally {
    loading.value = false
  }
}

const handleSearch = (): void => {
  if (!filterForm.execution) {
    ElMessage.warning('请先选择执行/迭代')
    return
  }
  pagination.page = 1
  fetchTasks()
}

const handleReset = (): void => {
  filterForm.execution = null
  filterForm.assignedTo = ''
  filterForm.status = ''
  filterForm.dateRange = []
  pagination.page = 1
  taskList.value = []
  pagination.total = 0
}

const handleSizeChange = (size: number): void => {
  if (!filterForm.execution) {
    return
  }
  pagination.pageSize = size
  pagination.page = 1
  fetchTasks()
}

const handlePageChange = (page: number): void => {
  if (!filterForm.execution) {
    return
  }
  pagination.page = page
  fetchTasks()
}

const getStatusType = (status: string): string => {
  const types: Record<string, string> = {
    wait: 'info',
    doing: 'primary',
    done: 'success',
    pause: 'warning',
    cancel: 'info',
    closed: 'info'
  }
  return types[status] || 'info'
}

const getStatusLabel = (status: string): string => {
  const labels: Record<string, string> = {
    wait: '未开始',
    doing: '进行中',
    done: '已完成',
    pause: '已暂停',
    cancel: '已取消',
    closed: '已关闭'
  }
  return labels[status] || status
}

const getProgress = (estimate: number, consumed: number): number => {
  if (!estimate || estimate === 0) return 0
  const progress = Math.min(Math.round((consumed / estimate) * 100), 100)
  return progress
}

const getProgressStatus = (estimate: number, consumed: number): string => {
  if (!estimate || estimate === 0) return ''
  const ratio = consumed / estimate
  if (ratio > 1) return 'exception'
  if (ratio >= 0.8) return 'warning'
  return ''
}

const openTaskDetail = (task: Task): void => {
  currentTask.value = task
  detailDialogVisible.value = true
}

const openZentaoTask = (taskId: number): void => {
  const zentaoUrl = `https://pm.kylin.com/task-view-${taskId}.html`
  try {
    const w = window as unknown as { runtime?: { BrowserOpenURL?: (url: string) => void } }
    if (w.runtime && w.runtime.BrowserOpenURL) {
      runtime.BrowserOpenURL(zentaoUrl)
    } else {
      window.open(zentaoUrl, '_blank', 'noopener,noreferrer')
    }
  } catch (error) {
    console.error('打开链接失败:', error)
    window.open(zentaoUrl, '_blank', 'noopener,noreferrer')
  }
}

onMounted(() => {
  fetchExecutions()
  fetchUsers()
})
</script>

<style scoped>
/* 页面特定样式 */
.task-title {
  color: #409eff;
  text-decoration: none;
  cursor: pointer;
}

.task-title:hover {
  text-decoration: underline;
}

.task-detail {
  line-height: 1.6;
  padding: 16px;
}

.task-detail .el-descriptions__content {
  word-break: break-word;
  line-height: 1.8;
}

.task-detail .el-descriptions__label {
  font-weight: 600;
  color: #2c3e50;
}

.dialog-actions {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}
</style>