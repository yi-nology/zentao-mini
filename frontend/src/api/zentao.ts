import type {
  Product,
  Project,
  Execution,
  Bug,
  Story,
  Task,
  User,
  TimelogAnalysis,
  TimelogDashboard,
  TimelogEffort,
  SelectOption,
  PaginatedResponse,
  ApiResponse,
  DashboardData,
  SearchResult
} from '@/types/api'
import api from './api'

interface BugParams {
  productId?: number
  projectId?: number
  assignedTo?: string
  status?: string
  version?: string
  startDate?: string
  endDate?: string
  specificDate?: string
  page?: number
  pageSize?: number
}

interface BuildParams {
  projectId?: number
  executionId?: number
}

export interface Build {
  id: number
  project: number
  product: number
  name: string
  date: string
}

interface StoryParams {
  productId?: number
  projectId?: number
  executionId?: number
  assignedTo?: string
  status?: string
  page?: number
  pageSize?: number
}

interface TaskParams {
  productId?: number
  executionId?: number
  assignedTo?: string
  status?: string
  startDate?: string
  endDate?: string
  page?: number
  pageSize?: number
}

interface TimelogParams {
  productId?: number
  projectId?: number
  executionId?: number
  assignedTo?: string
  dateFrom?: string
  dateTo?: string
}

interface CacheItem<T> {
  data: T
  timestamp: number
}

interface UserCache {
  data: Record<string, CacheItem<unknown>>
  set: <T>(key: string, data: T) => void
  get: <T>(key: string) => T | null
}

export const getProducts = (): Promise<ApiResponse<Product[]>> => {
  return api.get('/products')
}

export const getProjects = (params: Record<string, unknown> = {}): Promise<ApiResponse<Project[]>> => {
  return api.get('/projects', { params })
}

export const getExecutions = (params: Record<string, unknown> = {}): Promise<ApiResponse<Execution[]>> => {
  return api.get('/executions', { params })
}

export const getBugs = (params: BugParams = {}): Promise<ApiResponse<PaginatedResponse<Bug>>> => {
  const apiParams: Record<string, unknown> = {}
  if (params.productId) apiParams.productId = params.productId
  if (params.projectId) apiParams.projectId = params.projectId
  if (params.assignedTo) apiParams.assignedTo = params.assignedTo
  if (params.status) apiParams.status = params.status
  if (params.version) apiParams.version = params.version
  if (params.startDate) apiParams.startDate = params.startDate
  if (params.endDate) apiParams.endDate = params.endDate
  if (params.specificDate) apiParams.specificDate = params.specificDate
  if (params.page) apiParams.page = params.page
  if (params.pageSize) apiParams.pageSize = params.pageSize
  return api.get('/bugs', { params: apiParams })
}

export const getBuildsByProject = (projectId: number): Promise<ApiResponse<Build[]>> => {
  return api.get('/builds/project', { params: { projectId } })
}

export const getBuildsByExecution = (executionId: number): Promise<ApiResponse<Build[]>> => {
  return api.get('/builds/execution', { params: { executionId } })
}

export const getStories = (params: StoryParams = {}): Promise<ApiResponse<PaginatedResponse<Story>>> => {
  const apiParams: Record<string, unknown> = {}
  if (params.productId) apiParams.productId = params.productId
  if (params.projectId) apiParams.projectId = params.projectId
  if (params.executionId) apiParams.executionId = params.executionId
  if (params.assignedTo) apiParams.assignedTo = params.assignedTo
  if (params.status) apiParams.status = params.status
  if (params.page) apiParams.page = params.page
  if (params.pageSize) apiParams.pageSize = params.pageSize
  return api.get('/stories', { params: apiParams })
}

export const getTasks = (params: TaskParams = {}): Promise<ApiResponse<PaginatedResponse<Task>>> => {
  const apiParams: Record<string, unknown> = {}
  if (params.productId) apiParams.productId = params.productId
  if (params.executionId) apiParams.executionId = params.executionId
  if (params.assignedTo) apiParams.assignedTo = params.assignedTo
  if (params.status) apiParams.status = params.status
  if (params.startDate) apiParams.startDate = params.startDate
  if (params.endDate) apiParams.endDate = params.endDate
  if (params.page) apiParams.page = params.page
  if (params.pageSize) apiParams.pageSize = params.pageSize
  return api.get('/tasks', { params: apiParams })
}

export const getBugStatusOptions = (): SelectOption[] => {
  return [
    { label: '激活', value: 'active' },
    { label: '已解决', value: 'resolved' },
    { label: '已关闭', value: 'closed' }
  ]
}

export const getBugSeverityOptions = (): SelectOption[] => {
  return [
    { label: '1', value: 1 },
    { label: '2', value: 2 },
    { label: '3', value: 3 },
    { label: '4', value: 4 }
  ]
}

export const getStoryStatusOptions = (): SelectOption[] => {
  return [
    { label: '草稿', value: 'draft' },
    { label: '激活', value: 'active' },
    { label: '已变更', value: 'changed' },
    { label: '已关闭', value: 'closed' }
  ]
}

export const getStoryStageOptions = (): SelectOption[] => {
  return [
    { label: '等待', value: 'wait' },
    { label: '已计划', value: 'planned' },
    { label: '已立项', value: 'projected' },
    { label: '研发中', value: 'developing' },
    { label: '研发完毕', value: 'developed' },
    { label: '测试中', value: 'testing' },
    { label: '测试完毕', value: 'tested' },
    { label: '已验收', value: 'verified' },
    { label: '已发布', value: 'released' }
  ]
}

export const getTaskStatusOptions = (): SelectOption[] => {
  return [
    { label: '未开始', value: 'wait' },
    { label: '进行中', value: 'doing' },
    { label: '已完成', value: 'done' },
    { label: '已暂停', value: 'pause' },
    { label: '已取消', value: 'cancel' },
    { label: '已关闭', value: 'closed' }
  ]
}

const userCache: UserCache = {
  data: {},
  set<T>(key: string, data: T): void {
    this.data[key] = {
      data,
      timestamp: Date.now()
    }
  },
  get<T>(key: string): T | null {
    const cached = this.data[key] as CacheItem<T> | undefined
    if (cached) {
      if (Date.now() - cached.timestamp < 5 * 60 * 1000) {
        return cached.data
      } else {
        delete this.data[key]
      }
    }
    return null
  }
}

export const getUsers = (_params: Record<string, unknown> = {}): Promise<User[]> => {
  const cacheKey = 'users_all'

  const cachedData = userCache.get<User[]>(cacheKey)
  if (cachedData) {
    return Promise.resolve(cachedData)
  }

  return api.get('/users/all').then((response) => {
    const res = response as unknown as ApiResponse<{ users: User[] }>
    const users: User[] = res?.data?.users || []
    userCache.set(cacheKey, users)
    return users
  })
}

export const getTimelogAnalysis = (params: TimelogParams = {}): Promise<ApiResponse<TimelogAnalysis>> => {
  const apiParams: Record<string, unknown> = {}
  if (params.productId) apiParams.productId = params.productId
  if (params.projectId) apiParams.projectId = params.projectId
  if (params.executionId) apiParams.executionId = params.executionId
  if (params.assignedTo) apiParams.assignedTo = params.assignedTo
  if (params.dateFrom) apiParams.dateFrom = params.dateFrom
  if (params.dateTo) apiParams.dateTo = params.dateTo
  return api.get('/timelog/analysis', { params: apiParams })
}

export const getTimelogDashboard = (params: TimelogParams = {}): Promise<ApiResponse<TimelogDashboard>> => {
  const apiParams: Record<string, unknown> = {}
  if (params.productId) apiParams.productId = params.productId
  if (params.projectId) apiParams.projectId = params.projectId
  if (params.executionId) apiParams.executionId = params.executionId
  if (params.assignedTo) apiParams.assignedTo = params.assignedTo
  if (params.dateFrom) apiParams.dateFrom = params.dateFrom
  if (params.dateTo) apiParams.dateTo = params.dateTo
  return api.get('/timelog/dashboard', { params: apiParams })
}

export const getTimelogEfforts = (params: TimelogParams = {}): Promise<ApiResponse<TimelogEffort[]>> => {
  const apiParams: Record<string, unknown> = {}
  if (params.productId) apiParams.productId = params.productId
  if (params.projectId) apiParams.projectId = params.projectId
  if (params.executionId) apiParams.executionId = params.executionId
  if (params.assignedTo) apiParams.assignedTo = params.assignedTo
  if (params.dateFrom) apiParams.dateFrom = params.dateFrom
  if (params.dateTo) apiParams.dateTo = params.dateTo
  return api.get('/timelog/efforts', { params: apiParams })
}

export const getTimelogExecutions = (params: Record<string, unknown> = {}): Promise<ApiResponse<Execution[]>> => {
  return api.get('/executions', { params })
}

export const uploadInitConfig = (formData: FormData): Promise<ApiResponse<unknown>> => {
  return api.post('/init/upload', formData, {
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
}

export const testZentaoConnection = (): Promise<ApiResponse<unknown>> => {
  return api.get('/users/current')
}

export const getDashboard = (productId: number): Promise<ApiResponse<DashboardData>> => {
  return api.get('/dashboard', { params: { productId } })
}

export interface InitStatusData {
  isFirstStart: boolean
  hasConfig: boolean
  message: string
}

export const getInitStatus = (): Promise<ApiResponse<InitStatusData>> => {
  return api.get('/init/status')
}

export interface AccountInfo {
  domain: string
  account: string
  connected: boolean
}

export const getAccountInfo = (): Promise<ApiResponse<AccountInfo>> => {
  return api.get('/init/account')
}

export const search = (params: {
  keyword: string
  productId?: number
  page?: number
  pageSize?: number
}): Promise<ApiResponse<SearchResult>> => {
  const apiParams: Record<string, unknown> = { keyword: params.keyword }
  if (params.productId) apiParams.productId = params.productId
  if (params.page) apiParams.page = params.page
  if (params.pageSize) apiParams.pageSize = params.pageSize
  return api.get('/search', { params: apiParams })
}
