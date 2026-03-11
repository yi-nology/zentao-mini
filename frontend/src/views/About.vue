<template>
  <div class="about">
    <h1>关于 禅道 Mini</h1>
    <p>这是一个基于 Vue3 + Vite + Element Plus 构建的轻量级项目管理工具</p>
    <el-button @click="goBack">
      <el-icon><Arrow-Left /></el-icon>
      返回首页
    </el-button>
  </div>
</template>

<script setup>
import { useRouter } from 'vue-router'

const router = useRouter()

const goBack = () => {
  router.push('/')
}
</script>

<style scoped>
.about {
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
