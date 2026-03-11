import { createRouter, createWebHistory } from 'vue-router'
import Layout from '../views/Layout.vue'

const routes = [
  {
    path: '/',
    component: Layout,
    redirect: '/bugs',
    children: [
      {
        path: 'bugs',
        name: 'Bugs',
        component: () => import('../views/Bugs.vue'),
        meta: { title: 'Bug 查询' }
      },
      {
        path: 'stories',
        name: 'Stories',
        component: () => import('../views/Stories.vue'),
        meta: { title: '需求查询' }
      },
      {
        path: 'tasks',
        name: 'Tasks',
        component: () => import('../views/Tasks.vue'),
        meta: { title: '任务查询' }
      },
      {
        path: 'timelog',
        name: 'Timelog',
        component: () => import('../views/Timelog.vue'),
        meta: { title: '工时统计' }
      },
      {
        path: 'mcp-guide',
        name: 'MCPGuide',
        component: () => import('../views/MCPGuide.vue'),
        meta: { title: 'MCP对接指南' }
      }
    ]
  },
  {
    path: '/init-guide',
    name: 'InitGuide',
    component: () => import('../views/InitGuide.vue'),
    meta: { title: '初始化引导' }
  },
  {
    path: '/init-status',
    name: 'InitStatus',
    component: () => import('../views/InitStatus.vue'),
    meta: { title: '初始化状态' }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 路由守卫，检查是否需要初始化
router.beforeEach((to, from, next) => {
  // 检查是否已经初始化
  const isInitialized = localStorage.getItem('initialized')
  
  // 如果未初始化且不是访问初始化相关页面，则跳转到初始化引导页面
  if (!isInitialized && to.path !== '/init-guide' && to.path !== '/init-status') {
    next('/init-guide')
  } else {
    next()
  }
})

export default router
