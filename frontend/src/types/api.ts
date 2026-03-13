export interface UserRef {
  id: number
  account: string
  avatar: string
  realname: string
}

export interface Product {
  id: number
  name: string
  code: string
  type: string
  status: string
  desc: string
}

export interface Project {
  id: number
  name: string
  code: string
  model: string
  type: string
  status: string
  pm: unknown
  begin: string
  end: string
  progress: unknown
}

export interface Execution {
  id: number
  project: number
  name: string
  code: string
  status: string
  type: string
  begin: string
  end: string
}

export interface Bug {
  id: number
  project: number
  product: number
  title: string
  keywords: string
  severity: number
  pri: number
  type: string
  os: string
  browser: string
  hardware: string
  steps: string
  status: string
  subStatus: string
  color: string
  confirmed: number
  planTime: string
  openedBy: UserRef
  openedDate: string
  openedBuild: string
  assignedTo: UserRef
  assignedDate: string
  deadline: unknown
  resolvedBy: unknown
  resolution: string
  resolvedBuild: string
  resolvedDate: unknown
  closedBy: unknown
  closedDate: unknown
  statusName: string
  lifeCycle: string
}

export interface Story {
  id: number
  product: number
  module: number
  plan: string
  source: string
  title: string
  spec: string
  verify: string
  type: string
  status: string
  stage: string
  pri: number
  estimate: number
  version: number
  openedBy: unknown
  openedDate: string
  assignedTo: unknown
  assignedDate: string
  closedBy: unknown
  closedDate: string
  closedReason: string
}

export interface Task {
  id: number
  project: number
  execution: number
  name: string
  type: string
  pri: number
  status: string
  assignedTo: UserRef
  estStarted: string
  deadline: string
  estimate: number
  consumed: number
  left: number
  desc: string
  openedBy: unknown
  openedDate: string
  finishedBy: unknown
  finishedDate: unknown
  closedBy: unknown
  closedDate: string
  statusName: string
}

export interface User {
  id: number
  account: string
  realname: string
}

export interface TimelogAnalysis {
  totalHours: number
  effortCount: number
  taskCount: number
  byProject: TimelogByProject[]
  byType: TimelogByType[]
  byDate: TimelogByDate[]
  efforts: TimelogEffort[]
}

export interface TimelogDashboard {
  totalHours: number
  effortCount: number
  taskCount: number
  byProject: TimelogByProject[]
  byType: TimelogByType[]
  byDate: TimelogByDate[]
}

export interface TimelogByProject {
  name: string
  hours: number
  count: number
}

export interface TimelogByType {
  name: string
  hours: number
  count: number
}

export interface TimelogByDate {
  date: string
  hours: number
  count: number
}

export interface TimelogEffort {
  id: number
  taskId: number
  taskName: string
  taskType: string
  project: string
  execution: string
  account: string
  date: string
  consumed: number
  work: string
}

export interface ApiResponse<T = unknown> {
  code: number
  message: string
  data: T
}

export interface PaginatedResponse<T> {
  list: T[]
  total: number
  page: number
  limit: number
}

export interface SelectOption {
  label: string
  value: string | number
}
