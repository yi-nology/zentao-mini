<template>
  <div class="product-selector">
    <el-select
      :model-value="selectedProduct"
      placeholder="请选择产品"
      clearable
      filterable
      value-key="id"
      style="width: 200px"
      @update:model-value="handleProductChange"
    >
      <el-option
        v-for="item in productOptions"
        :key="item.id"
        :label="item.name"
        :value="item"
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
import { ref, computed, onMounted, watch } from 'vue'
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
const selectedProduct = ref<Product | null>(null)
const selectedProject = ref<string>('')

const selectedProductId = computed(() => selectedProduct.value?.id?.toString() ?? '')

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

const handleProductChange = async (product: Product | null): Promise<void> => {
  selectedProduct.value = product
  selectedProject.value = ''
  projectOptions.value = []

  if (product) {
    await fetchProjects(product.id)
  }

  emitSelection()
}

const handleProjectChange = (projectId: string | number): void => {
  selectedProject.value = String(projectId || '')
  emitSelection()
}

const emitSelection = (): void => {
  const selection: SelectionValue = {
    product: selectedProductId.value,
    project: selectedProject.value
  }
  emit('update:modelValue', selection)
  emit('change', selection)
}

watch(() => props.modelValue, (newVal) => {
  if (newVal && newVal.product !== selectedProductId.value) {
    selectedProject.value = newVal.project || ''
    if (newVal.product) {
      const productId = Number(newVal.product)
      const found = productOptions.value.find(p => p.id === productId)
      selectedProduct.value = found || null
      fetchProjects(newVal.product)
    } else {
      selectedProduct.value = null
    }
  }
}, { immediate: true, deep: true })

onMounted(() => {
  fetchProducts()
  if (props.modelValue && props.modelValue.product) {
    const productId = Number(props.modelValue.product)
    const found = productOptions.value.find(p => p.id === productId)
    selectedProduct.value = found || null
    selectedProject.value = props.modelValue.project || ''
    fetchProjects(props.modelValue.product)
  }
})
</script>

<style scoped>
.product-selector {
  display: flex;
  gap: 10px;
}
</style>
