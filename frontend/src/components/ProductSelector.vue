<template>
  <div class="product-selector">
    <el-select
      :model-value="selectedProduct"
      placeholder="请选择产品"
      clearable
      filterable
      style="width: 200px"
      @update:model-value="handleProductChange"
    >
      <el-option
        v-for="item in productOptions"
        :key="item.id"
        :label="item.name"
        :value="item.id"
      />
    </el-select>
    <el-select
      :model-value="selectedProject"
      placeholder="请选择项目"
      clearable
      filterable
      :disabled="!selectedProduct"
      style="width: 200px"
      @update:model-value="handleProjectChange"
    >
      <el-option
        v-for="item in projectOptions"
        :key="item.id"
        :label="item.name"
        :value="item.id"
      />
    </el-select>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { getProducts, getProjects } from '@/api/zentao'
import type { Product, Project } from '@/types/api'

interface SelectionValue {
  product: string
  project: string
}

const props = defineProps<{
  modelValue?: SelectionValue
}>()

const emit = defineEmits<{
  'update:modelValue': [value: SelectionValue]
  'change': [value: SelectionValue]
}>()

const productOptions = ref<Product[]>([])
const projectOptions = ref<Project[]>([])
const selectedProduct = ref<number | string>('')
const selectedProject = ref<number | string>('')

const fetchProducts = async (): Promise<void> => {
  try {
    const res = await getProducts()
    productOptions.value = res.data || []
  } catch (error) {
    console.error('获取产品列表失败:', error)
  }
}

const fetchProjects = async (productId: string | number): Promise<void> => {
  try {
    const params = productId ? { productId: productId } : {}
    const res = await getProjects(params)
    projectOptions.value = res.data || []
  } catch (error) {
    console.error('获取项目列表失败:', error)
  }
}

const handleProductChange = async (productId: string | number): Promise<void> => {
  selectedProduct.value = productId ?? ''
  selectedProject.value = ''
  projectOptions.value = []

  if (productId) {
    await fetchProjects(productId)
  }

  emitSelection()
}

const handleProjectChange = (projectId: string | number): void => {
  selectedProject.value = projectId ?? ''
  emitSelection()
}

const emitSelection = (): void => {
  const selection: SelectionValue = {
    product: String(selectedProduct.value),
    project: String(selectedProject.value)
  }
  emit('update:modelValue', selection)
  emit('change', selection)
}

watch(() => props.modelValue, (newVal) => {
  if (newVal) {
    const newProduct = newVal.product ? Number(newVal.product) || newVal.product : ''
    const newProject = newVal.project ? Number(newVal.project) || newVal.project : ''

    if (newProduct !== selectedProduct.value) {
      selectedProduct.value = newProduct
      selectedProject.value = newProject
      if (newProduct) {
        fetchProjects(newProduct)
      }
    }
  }
}, { immediate: true, deep: true })

onMounted(() => {
  fetchProducts()
  if (props.modelValue && props.modelValue.product) {
    const productId = Number(props.modelValue.product) || props.modelValue.product
    const projectId = Number(props.modelValue.project) || props.modelValue.project
    selectedProduct.value = productId
    selectedProject.value = projectId
    if (productId) {
      fetchProjects(productId)
    }
  }
})
</script>

<style scoped>
.product-selector {
  display: flex;
  gap: 10px;
}
</style>
