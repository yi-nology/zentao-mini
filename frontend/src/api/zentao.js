import api from './api'

// 获取产品列表
export const getProducts = () => {
  return api.get('/products')
}

// 获取项目列表
export const getProjects = (params = {}) => {
  return api.get('/projects', { params })
}

// 获取执行/迭代列表
export const getExecutions = (params = {}) => {
  return api.get('/executions', { params })
}

// 获取 Bug 列表
export const getBugs = (params = {}) => {
  // 转换参数名以匹配后端 API
  const apiParams = {}
  if (params.product || params.productID) apiParams.productID = params.product || params.productID
  if (params.project || params.projectID) apiParams.projectID = params.project || params.projectID
  if (params.assignedTo) apiParams.assignedTo = params.assignedTo
  if (params.status) apiParams.status = params.status
  if (params.page) apiParams.page = params.page
  if (params.pageSize) apiParams.pageSize = params.pageSize
  return api.get('/bugs', { params: apiParams })
}

// 获取需求列表
export const getStories = (params = {}) => {
  // 转换参数名以匹配后端 API
  const apiParams = {}
  if (params.product) apiParams.productID = params.product
  if (params.project) apiParams.projectID = params.project
  if (params.execution) apiParams.executionID = params.execution
  if (params.assignedTo) apiParams.assignedTo = params.assignedTo
  if (params.status) apiParams.status = params.status
  if (params.page) apiParams.page = params.page
  if (params.pageSize) apiParams.pageSize = params.pageSize
  return api.get('/stories', { params: apiParams })
}

// 获取任务列表
export const getTasks = (params = {}) => {
  // 转换参数名以匹配后端 API
  const apiParams = {}
  if (params.execution || params.executionID) apiParams.executionID = params.execution || params.executionID
  if (params.assignedTo) apiParams.assignedTo = params.assignedTo
  if (params.status) apiParams.status = params.status
  if (params.startDate) apiParams.startDate = params.startDate
  if (params.endDate) apiParams.endDate = params.endDate
  if (params.page) apiParams.page = params.page
  if (params.pageSize) apiParams.pageSize = params.pageSize
  return api.get('/tasks', { params: apiParams })
}

// 获取 Bug 状态选项
export const getBugStatusOptions = () => {
  return [
    { label: '激活', value: 'active' },
    { label: '已解决', value: 'resolved' },
    { label: '已关闭', value: 'closed' }
  ]
}

// 获取 Bug 严重程度选项
export const getBugSeverityOptions = () => {
  return [
    { label: '1', value: 1 },
    { label: '2', value: 2 },
    { label: '3', value: 3 },
    { label: '4', value: 4 }
  ]
}

// 获取需求状态选项
export const getStoryStatusOptions = () => {
  return [
    { label: '草稿', value: 'draft' },
    { label: '激活', value: 'active' },
    { label: '已变更', value: 'changed' },
    { label: '已关闭', value: 'closed' }
  ]
}

// 获取需求阶段选项
export const getStoryStageOptions = () => {
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

// 获取任务状态选项
export const getTaskStatusOptions = () => {
  return [
    { label: '未开始', value: 'wait' },
    { label: '进行中', value: 'doing' },
    { label: '已完成', value: 'done' },
    { label: '已暂停', value: 'pause' },
    { label: '已取消', value: 'cancel' },
    { label: '已关闭', value: 'closed' }
  ]
}

// 用户列表内存缓存
const userCache = {
  data: {},
  // 缓存用户列表
  set(key, data) {
    this.data[key] = {
      data,
      timestamp: Date.now()
    }
  },
  // 获取缓存的用户列表
  get(key) {
    const cached = this.data[key]
    if (cached) {
      // 简单的缓存过期检查（5分钟）
      if (Date.now() - cached.timestamp < 5 * 60 * 1000) {
        return cached.data
      } else {
        // 缓存过期，删除
        delete this.data[key]
      }
    }
    return null
  }
}

// 获取用户列表
export const getUsers = (params = {}) => {
  // 构建缓存键
  const cacheKey = 'users_all'
  
  // 尝试从缓存获取
  const cachedData = userCache.get(cacheKey)
  if (cachedData) {
    return Promise.resolve(cachedData)
  }
  
  // 发起请求
  return api.get('/users/all').then(data => {
    // 缓存结果
    userCache.set(cacheKey, data)
    return data
  })
}

// 获取工时统计分析
export const getTimelogAnalysis = (params = {}) => {
  // 转换参数名以匹配后端 API
  const apiParams = {}
  if (params.productId || params.productID) apiParams.productId = params.productId || params.productID
  if (params.projectId || params.projectID) apiParams.projectId = params.projectId || params.projectID
  if (params.executionId || params.executionID) apiParams.executionId = params.executionId || params.executionID
  if (params.assignedTo) apiParams.assignedTo = params.assignedTo
  if (params.dateFrom) apiParams.dateFrom = params.dateFrom
  if (params.dateTo) apiParams.dateTo = params.dateTo
  return api.get('/timelog/analysis', { params: apiParams })
}

// 获取工时看板数据
export const getTimelogDashboard = (params = {}) => {
  // 转换参数名以匹配后端 API
  const apiParams = {}
  if (params.productId || params.productID) apiParams.productId = params.productId || params.productID
  if (params.projectId || params.projectID) apiParams.projectId = params.projectId || params.projectID
  if (params.executionId || params.executionID) apiParams.executionId = params.executionId || params.executionID
  if (params.assignedTo) apiParams.assignedTo = params.assignedTo
  if (params.dateFrom) apiParams.dateFrom = params.dateFrom
  if (params.dateTo) apiParams.dateTo = params.dateTo
  return api.get('/timelog/dashboard', { params: apiParams })
}

// 获取工时明细数据
export const getTimelogEfforts = (params = {}) => {
  // 转换参数名以匹配后端 API
  const apiParams = {}
  if (params.productId || params.productID) apiParams.productId = params.productId || params.productID
  if (params.projectId || params.projectID) apiParams.projectId = params.projectId || params.projectID
  if (params.executionId || params.executionID) apiParams.executionId = params.executionId || params.executionID
  if (params.assignedTo) apiParams.assignedTo = params.assignedTo
  if (params.dateFrom) apiParams.dateFrom = params.dateFrom
  if (params.dateTo) apiParams.dateTo = params.dateTo
  return api.get('/timelog/efforts', { params: apiParams })
}

// 获取执行列表（用于工时页面）
export const getTimelogExecutions = (params = {}) => {
  return api.get('/executions', { params })
}

// 上传初始化配置文件
export const uploadInitConfig = (formData) => {
  return api.post('/init/upload', formData, {
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
}

// 测试禅道连接
export const testZentaoConnection = () => {
  return api.get('/users/current')
}

// 获取初始化状态
export const getInitStatus = () => {
  return api.get('/init/status')
}
