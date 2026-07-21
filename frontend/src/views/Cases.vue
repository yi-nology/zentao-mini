<template>
  <div class="page-container">
    <div class="filter-card">
      <el-form :inline="true" :model="filterForm" class="filter-form">
        <el-form-item label="产品">
          <el-select v-model="filterForm.productId" placeholder="请选择产品" filterable style="width: 220px" @change="handleSearch">
            <el-option v-for="item in productOptions" :key="item.id" :label="item.name" :value="item.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="关键字">
          <el-input v-model="filterForm.keyword" placeholder="标题/ID" clearable style="width: 180px" @keyup.enter="handleSearch" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch"><el-icon><Search /></el-icon>查询</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </div>

    <div class="table-card">
      <div class="table-header">
        <span>共 {{ filteredList.length }} 条用例</span>
      </div>
      <el-table v-loading="loading" :data="pagedList" border stripe style="width: 100%">
        <el-table-column prop="id" label="ID" width="90" align="center" />
        <el-table-column prop="title" label="标题" min-width="240" show-overflow-tooltip>
          <template #default="{ row }">
            <a href="javascript:void(0)" @click="openExternalLink(buildZentaoUrl(`testcase-view-${row.id}.html`))" class="entity-title">{{ row.title }}</a>
          </template>
        </el-table-column>
        <el-table-column prop="type" label="类型" width="100" align="center" />
        <el-table-column prop="status" label="状态" width="90" align="center">
          <template #default="{ row }">
            <el-tag size="small">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="pri" label="优先级" width="80" align="center" />
        <el-table-column prop="openedBy.account" label="创建人" width="110" align="center">
          <template #default="{ row }">{{ row.openedBy?.realname || row.openedBy?.account }}</template>
        </el-table-column>
        <el-table-column prop="openedDate" label="创建时间" width="160" align="center" />
      </el-table>
      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :total="filteredList.length"
          :page-sizes="[20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          background
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, inject } from 'vue'
import { Search } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { getCases, getProducts } from '@/api/zentao'
import { useZentaoConfig } from '@/composables/useZentaoConfig'
import { openExternalLink } from '@/composables/useExternalLink'

interface GlobalSelection { product: number | null }
const globalSelection = inject<GlobalSelection>('globalSelection')!

interface CaseItem {
  id: number
  title: string
  type: string
  status: string
  pri: number
  openedBy: { account: string; realname: string }
  openedDate: string
}

const { buildUrl: buildZentaoUrl } = useZentaoConfig()

const loading = ref(false)
const list = ref<CaseItem[]>([])
const productOptions = ref<{ id: number; name: string }[]>([])

const filterForm = reactive<{ productId: number | null; keyword: string }>({
  productId: null,
  keyword: ''
})

const pagination = reactive({ page: 1, pageSize: 20 })

const filteredList = computed<CaseItem[]>(() => {
  if (!filterForm.keyword) return list.value
  const kw = filterForm.keyword.toLowerCase()
  return list.value.filter(c =>
    c.title.toLowerCase().includes(kw) || String(c.id).includes(kw)
  )
})

const pagedList = computed<CaseItem[]>(() => {
  const start = (pagination.page - 1) * pagination.pageSize
  return filteredList.value.slice(start, start + pagination.pageSize)
})

const fetchProducts = async (): Promise<void> => {
  try {
    const res = await getProducts()
    productOptions.value = (res.data || []).map((p: { id: number; name: string }) => ({ id: p.id, name: p.name }))
  } catch { /* ignore */ }
}

const fetchData = async (): Promise<void> => {
  if (!filterForm.productId) {
    ElMessage.warning('请先选择产品')
    return
  }
  loading.value = true
  try {
    const res = await getCases({ productId: filterForm.productId, page: 1, pageSize: 1000 })
    list.value = ((res.data as { list?: CaseItem[] })?.list) || []
    pagination.page = 1
  } catch (err) {
    list.value = []
  } finally {
    loading.value = false
  }
}

const handleSearch = (): void => { fetchData() }
const handleReset = (): void => {
  filterForm.keyword = ''
  fetchData()
}

onMounted(async (): Promise<void> => {
  await fetchProducts()
  if (globalSelection.product) {
    filterForm.productId = globalSelection.product
    await fetchData()
  }
})
</script>

<style scoped>
.filter-card { background: var(--color-bg-card); border-radius: var(--radius-md); padding: 16px 20px; margin-bottom: 16px; box-shadow: var(--shadow-sm); }
.filter-form { display: flex; flex-wrap: wrap; gap: 0; }
.table-card { background: var(--color-bg-card); border-radius: var(--radius-md); padding: 16px 20px; box-shadow: var(--shadow-sm); }
.table-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; font-size: 13px; color: var(--color-text-secondary); }
.pagination-wrapper { display: flex; justify-content: flex-end; margin-top: 16px; }
.entity-title { color: var(--color-primary); text-decoration: none; }
.entity-title:hover { text-decoration: underline; }
</style>
