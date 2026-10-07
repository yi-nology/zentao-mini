import api from './api'
import type { ApiResponse } from '@/types/api'
import type {
  SchedulerTask,
  TaskExecutionLog,
  WebhookResult,
  RequirementReport,
  TaskProgressReport,
  BugReport,
  BugAgingReport,
  DailyReportCheckReport,
} from '@/types/scheduler'

export const listTasks = (): Promise<ApiResponse<SchedulerTask[]>> =>
  api.get('/scheduler/tasks')

export const createTask = (task: Partial<SchedulerTask>): Promise<ApiResponse<SchedulerTask>> =>
  api.post('/scheduler/tasks', task)

export const updateTask = (id: string, task: Partial<SchedulerTask>): Promise<ApiResponse<SchedulerTask>> =>
  api.put(`/scheduler/tasks/${id}`, task)

export const deleteTask = (id: string): Promise<ApiResponse<null>> =>
  api.delete(`/scheduler/tasks/${id}`)

export const toggleTask = (id: string): Promise<ApiResponse<SchedulerTask>> =>
  api.patch(`/scheduler/tasks/${id}/toggle`)

export const runTaskNow = (id: string): Promise<ApiResponse<TaskExecutionLog>> =>
  api.post(`/scheduler/tasks/${id}/run`)

export const getTaskLogs = (id: string): Promise<ApiResponse<TaskExecutionLog[]>> =>
  api.get(`/scheduler/tasks/${id}/logs`)

export const getAllLogs = (): Promise<ApiResponse<TaskExecutionLog[]>> =>
  api.get('/scheduler/logs')

export const testWebhook = (url: string): Promise<ApiResponse<WebhookResult>> =>
  api.post('/scheduler/test-webhook', { url })

export interface PreviewParams {
  reportType: string
  productId: number
  projectId: number
  projectName: string
  productName: string
  statusFilter: string
  agingDays?: number
  checkHours?: number
  /** 日报检查的检查周期截止月（YYYY-MM），为空则检查最近周期（上月16日～本月15日） */
  period?: string
  keyword: string
  externalInfo: string
  messageHeader?: string
  priorityAssignees?: string[]
  viewURL?: string
}

export const previewReport = (params: PreviewParams): Promise<ApiResponse<RequirementReport | TaskProgressReport | BugReport | BugAgingReport | DailyReportCheckReport>> =>
  api.post('/scheduler/preview', params)
