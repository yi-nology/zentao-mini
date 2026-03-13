import type { AxiosPromise } from 'axios'
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
  SelectOption
} from '@/types/api'
import api from './api'

interface BugParams {
  product?: number
  productID?: number
  project?: number
  projectID?: number
  assignedTo?: string
  status?: string
  page?: number
  pageSize?: number
}

interface StoryParams {
  product?: number
  project?: number
  execution?: number
  assignedTo?: string
  status?: string
  page?: number
  pageSize?: number
}

interface TaskParams {
  execution?: number
  executionID?: number
  assignedTo?: string
  status?: string
  startDate?: string
  endDate?: string
  page?: number
  pageSize?: number
}

interface TimelogParams {
  productId?: number
  productID?: number
  projectId?: number
  projectID?: number
  executionId?: number
  executionID?: number
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

export const getProducts = (): AxiosPromise<Product[]> => {
  return api.get('/products')
}

export const getProjects = (params: Record<string, unknown> = {}): AxiosPromise<Project[]> => {
  return api.get('/projects', { params })
}

export const getExecutions = (params: Record<string, unknown> = {}): AxiosPromise<Execution[]> => {
  return api.get('/executions', { params })
}

export const getBugs = (params: BugParams = {}): AxiosPromise<Bug[]> => {
  const apiParams: Record<string, unknown> = {}
  if (params.product || params.productID) apiParams.productID = params.product || params.productID
  if (params.project || params.projectID) apiParams.projectID = params.project || params.projectID
  if (params.assignedTo) apiParams.assignedTo = params.assignedTo
  if (params.status) apiParams.status = params.status
  if (params.page) apiParams.page = params.page
  if (params.pageSize) apiParams.pageSize = params.pageSize
  return api.get('/bugs', { params: apiParams })
}

export const getStories = (params: StoryParams = {}): AxiosPromise<Story[]> => {
  const apiParams: Record<string, unknown> = {}
  if (params.product) apiParams.productID = params.product
  if (params.project) apiParams.projectID = params.project
  if (params.execution) apiParams.executionID = params.execution
  if (params.assignedTo) apiParams.assignedTo = params.assignedTo
  if (params.status) apiParams.status = params.status
  if (params.page) apiParams.page = params.page
  if (params.pageSize) apiParams.pageSize = params.pageSize
  return api.get('/stories', { params: apiParams })
}

export const getTasks = (params: TaskParams = {}): AxiosPromise<Task[]> => {
  const apiParams: Record<string, unknown> = {}
  if (params.execution || params.executionID) apiParams.executionID = params.execution || params.executionID
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
  
  return api.get('/users/all').then((data: unknown) => {
    userCache.set(cacheKey, data)
    return data as User[]
  })
}

export const getTimelogAnalysis = (params: TimelogParams = {}): AxiosPromise<TimelogAnalysis> => {
  const apiParams: Record<string, unknown> = {}
  if (params.productId || params.productID) apiParams.productId = params.productId || params.productID
  if (params.projectId || params.projectID) apiParams.projectId = params.projectId || params.projectID
  if (params.executionId || params.executionID) apiParams.executionId = params.executionId || params.executionID
  if (params.assignedTo) apiParams.assignedTo = params.assignedTo
  if (params.dateFrom) apiParams.dateFrom = params.dateFrom
  if (params.dateTo) apiParams.dateTo = params.dateTo
  return api.get('/timelog/analysis', { params: apiParams })
}

export const getTimelogDashboard = (params: TimelogParams = {}): AxiosPromise<TimelogDashboard> => {
  const apiParams: Record<string, unknown> = {}
  if (params.productId || params.productID) apiParams.productId = params.productId || params.productID
  if (params.projectId || params.projectID) apiParams.projectId = params.projectId || params.projectID
  if (params.executionId || params.executionID) apiParams.executionId = params.executionId || params.executionID
  if (params.assignedTo) apiParams.assignedTo = params.assignedTo
  if (params.dateFrom) apiParams.dateFrom = params.dateFrom
  if (params.dateTo) apiParams.dateTo = params.dateTo
  return api.get('/timelog/dashboard', { params: apiParams })
}

export const getTimelogEfforts = (params: TimelogParams = {}): AxiosPromise<TimelogEffort[]> => {
  const apiParams: Record<string, unknown> = {}
  if (params.productId || params.productID) apiParams.productId = params.productId || params.productID
  if (params.projectId || params.projectID) apiParams.projectId = params.projectId || params.projectID
  if (params.executionId || params.executionID) apiParams.executionId = params.executionId || params.executionID
  if (params.assignedTo) apiParams.assignedTo = params.assignedTo
  if (params.dateFrom) apiParams.dateFrom = params.dateFrom
  if (params.dateTo) apiParams.dateTo = params.dateTo
  return api.get('/timelog/efforts', { params: apiParams })
}

export const getTimelogExecutions = (params: Record<string, unknown> = {}): AxiosPromise<Execution[]> => {
  return api.get('/executions', { params })
}

export const uploadInitConfig = (formData: FormData): AxiosPromise<unknown> => {
  return api.post('/init/upload', formData, {
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
}

export const testZentaoConnection = (): AxiosPromise<unknown> => {
  return api.get('/users/current')
}

export const getInitStatus = (): AxiosPromise<unknown> => {
  return api.get('/init/status')
}
