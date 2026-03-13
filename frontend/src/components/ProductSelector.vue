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
const selectedProduct = ref<string>('')
const selectedProject = ref<string>('')

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
    const params = productId ? { productID: productId } : {}
    const res = await getProjects(params)
    projectOptions.value = res.data || []
  } catch (error) {
    console.error('获取项目列表失败:', error)
  }
}

const handleProductChange = async (productId: string | number): Promise<void> => {
  selectedProduct.value = String(productId || '')
  selectedProject.value = ''
  projectOptions.value = []
  
  if (productId) {
    await fetchProjects(productId)
  }
  
  emitSelection()
}

const handleProjectChange = (projectId: string | number): void => {
  selectedProject.value = String(projectId || '')
  emitSelection()
}

const emitSelection = (): void => {
  const selection: SelectionValue = {
    product: selectedProduct.value,
    project: selectedProject.value
  }
  emit('update:modelValue', selection)
  emit('change', selection)
}

watch(() => props.modelValue, (newVal) => {
  if (newVal && newVal.product !== selectedProduct.value) {
    selectedProduct.value = newVal.product || ''
    selectedProject.value = newVal.project || ''
    if (newVal.product) {
      fetchProjects(newVal.product)
    }
  }
}, { immediate: true, deep: true })

onMounted(() => {
  fetchProducts()
  if (props.modelValue && props.modelValue.product) {
    selectedProduct.value = props.modelValue.product
    selectedProject.value = props.modelValue.project || ''
    if (props.modelValue.product) {
      fetchProjects(props.modelValue.product)
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
