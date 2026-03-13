<template>
  <div class="home">
    <h1>欢迎使用 禅道 Mini</h1>
    <p>这是一个 Vue3 + Vite + Element Plus 项目</p>
    <el-button type="primary" @click="goToAbout">
      <el-icon><Info-Filled /></el-icon>
      关于我们
    </el-button>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'

const router = useRouter()

const goToAbout = (): void => {
  router.push('/about')
}
</script>

<style scoped>
.home {
  text-align: center;
  padding: 40px;
}

h1 {
  color: #409eff;
  margin-bottom: 20px;
}

p {
  color: #666;
  margin-bottom: 30px;
}
</style>
