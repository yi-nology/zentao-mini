<template>
  <div class="about">
    <h1>关于 禅道 Mini</h1>
    <p>这是一个基于 Vue3 + Vite + Element Plus 构建的轻量级项目管理工具</p>
    
    <div class="official-site">
      <h2>官方网站</h2>
      <p>了解更多信息，请访问我们的官网</p>
      <el-button type="primary" @click="openOfficialSite">
        <el-icon><Link /></el-icon>
        访问官网
      </el-button>
    </div>
    
    <el-button @click="goBack">
      <el-icon><Arrow-Left /></el-icon>
      返回首页
    </el-button>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { Link } from '@element-plus/icons-vue'

const router = useRouter()

const goBack = (): void => {
  router.push('/')
}

const openOfficialSite = (): void => {
  window.open('https://murphyyi.com', '_blank')
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

.official-site {
  margin: 40px 0;
  padding: 30px;
  background-color: #f5f7fa;
  border-radius: 8px;
}

.official-site h2 {
  color: #409eff;
  font-size: 20px;
  margin-bottom: 15px;
}

.official-site p {
  color: #666;
  margin-bottom: 20px;
}
</style>
