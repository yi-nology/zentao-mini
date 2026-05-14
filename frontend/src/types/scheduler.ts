export interface WebhookConfig {
  id: string
  name: string
  url: string
  enabled: boolean
}

export interface SchedulerTask {
  id: string
  name: string
  enabled: boolean
  cronExpr: string
  webhooks: WebhookConfig[]
  projectId: number
  productId: number
  projectName: string
  productName: string
  statusFilter: string
  lastRunAt: string | null
  lastRunStatus: string
  createdAt: string
  updatedAt: string
}

export interface WebhookResult {
  webhookId: string
  webhookName: string
  webhookUrl: string
  success: boolean
  statusCode: number
  error: string
}

export interface TaskExecutionLog {
  id: string
  taskId: string
  taskName: string
  startedAt: string
  finishedAt: string | null
  status: 'running' | 'success' | 'partial' | 'failed'
  bugTotal: number
  highSeverity: number
  assigneeCount: number
  webhookResults: WebhookResult[]
  error: string
}

export interface AssigneeBugStats {
  assignee: string
  account: string
  total: number
  highSeverity: number
  fatal: number
  serious: number
  moderate: number
  minor: number
}

export interface BugReport {
  title: string
  timestamp: string
  projectName: string
  total: number
  highSeverity: number
  statusBreakdown: Record<string, number>
  details: AssigneeBugStats[]
  message: string
}

export const CRON_PRESETS = [
  { label: '每天 9:00', expr: '0 9 * * *' },
  { label: '工作日 9:00', expr: '0 9 * * 1-5' },
  { label: '每周一 9:00', expr: '0 9 * * 1' },
  { label: '每8小时', expr: '0 */8 * * *' },
  { label: '每30分钟', expr: '*/30 * * * *' },
] as const

export const STATUS_OPTIONS = [
  { label: '活跃 (active)', value: 'active' },
  { label: '已解决 (resolved)', value: 'resolved' },
  { label: '已关闭 (closed)', value: 'closed' },
  { label: '全部', value: 'all' },
] as const
