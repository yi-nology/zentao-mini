<template>
  <el-container class="layout-container">
    <el-aside width="200px" class="aside">
      <div class="logo">
        <span>禅道 Mini</span>
      </div>
      <el-menu
        :default-active="$route.path"
        router
        class="menu"
        background-color="#304156"
        text-color="#bfcbd9"
        active-text-color="#409EFF"
      >
        <el-menu-item index="/bugs">
          <el-icon><svg viewBox="0 0 1024 1024" width="1em" height="1em"><path fill="currentColor" d="M512 64C264.6 64 64 264.6 64 512s200.6 448 448 448 448-200.6 448-448S759.4 64 512 64zm0 820c-205.4 0-372-166.6-372-372s166.6-372 372-372 372 166.6 372 372-166.6 372-372 372z"/><path fill="currentColor" d="M464 336a48 48 0 1096 0 48 48 0 10-96 0zm72 112h-48c-4.4 0-8 3.6-8 8v272c0 4.4 3.6 8 8 8h48c4.4 0 8-3.6 8-8V456c0-4.4-3.6-8-8-8z"/></svg></el-icon>
          <span>Bug 查询</span>
        </el-menu-item>
        <el-menu-item index="/stories">
          <el-icon><svg viewBox="0 0 1024 1024" width="1em" height="1em"><path fill="currentColor" d="M832 64H192c-17.7 0-32 14.3-32 32v832c0 17.7 14.3 32 32 32h640c17.7 0 32-14.3 32-32V96c0-17.7-14.3-32-32-32zm-600 72h560v80H232v-80zm560 640H232V320h560v456z"/></svg></el-icon>
          <span>需求查询</span>
        </el-menu-item>
        <el-menu-item index="/tasks">
          <el-icon><svg viewBox="0 0 1024 1024" width="1em" height="1em"><path fill="currentColor" d="M880 112H144c-17.7 0-32 14.3-32 32v736c0 17.7 14.3 32 32 32h736c17.7 0 32-14.3 32-32V144c0-17.7-14.3-32-32-32zM368 744H232V608h136v136zm0-192H232V416h136v136zm0-192H232V224h136v136zm192 384H416V608h136v136zm0-192H416V416h136v136zm0-192H416V224h136v136zm192 384H608V608h136v136zm0-192H608V416h136v136zm0-192H608V224h136v136z"/></svg></el-icon>
          <span>任务查询</span>
        </el-menu-item>
        <el-menu-item index="/timelog">
          <el-icon><svg viewBox="0 0 1024 1024" width="1em" height="1em"><path fill="currentColor" d="M512 64C264.6 64 64 264.6 64 512s200.6 448 448 448 448-200.6 448-448S759.4 64 512 64zm0 820c-205.4 0-372-166.6-372-372s166.6-372 372-372 372 166.6 372 372-166.6 372-372 372z"/><path fill="currentColor" d="M512 192a320 320 0 100 640 320 320 0 000-640zm0 680a360 360 0 110-720 360 360 0 010 720z"/><path fill="currentColor" d="M480 464h64v192h-64z"/></svg></el-icon>
          <span>工时统计</span>
        </el-menu-item>
        <el-menu-item index="/mcp-guide">
          <el-icon><svg viewBox="0 0 1024 1024" width="1em" height="1em"><path fill="currentColor" d="M512 64C264.6 64 64 264.6 64 512s200.6 448 448 448 448-200.6 448-448S759.4 64 512 64zm0 820c-205.4 0-372-166.6-372-372s166.6-372 372-372 372 166.6 372 372-166.6 372-372 372z"/><path fill="currentColor" d="M480 464h64v192h-64z"/><path fill="currentColor" d="M416 384h192v64H416z"/></svg></el-icon>
          <span>MCP对接指南</span>
        </el-menu-item>
      </el-menu>
    </el-aside>
    <el-container>
      <el-header class="header">
        <div class="header-title">{{ pageTitle }}</div>
        <div class="header-selector">
          <ProductSelector
            :model-value="globalSelection"
            @update:model-value="handleSelectionChange"
          />
        </div>
      </el-header>
      <el-main class="main">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { computed, provide, reactive } from 'vue'
import { useRoute } from 'vue-router'
import ProductSelector from '@/components/ProductSelector.vue'

const route = useRoute()
const globalSelection = reactive({ product: '', project: '' })

const pageTitle = computed(() => {
  const titles = {
    '/bugs': 'Bug 查询',
    '/stories': '需求查询',
    '/tasks': '任务查询',
    '/timelog': '工时统计',
    '/mcp-guide': 'MCP对接指南'
  }
  return titles[route.path] || '禅道 Mini'
})

const handleSelectionChange = (selection) => {
  globalSelection.product = selection.product
  globalSelection.project = selection.project
}

provide('globalSelection', globalSelection)
</script>

<style scoped>
.layout-container {
  height: 100vh;
  display: flex;
  flex-direction: row;
  overflow: hidden;
}

.aside {
  background-color: #2c3e50;
  width: 200px;
  flex-shrink: 0;
  transition: all 0.3s ease;
}

.aside:hover {
  background-color: #243342;
}

.logo {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 18px;
  font-weight: bold;
  border-bottom: 1px solid #1f2d3d;
  background-color: #243342;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.menu {
  border-right: none;
  height: calc(100vh - 60px);
}

.menu :deep(.el-menu-item) {
  height: 50px;
  line-height: 50px;
  margin: 0 10px;
  border-radius: 4px;
  margin-bottom: 5px;
  transition: all 0.3s ease;
}

.menu :deep(.el-menu-item:hover) {
  background-color: rgba(255, 255, 255, 0.1) !important;
}

.menu :deep(.el-menu-item.is-active) {
  background-color: #409EFF !important;
  color: #fff !important;
}

.header {
  background-color: #fff;
  box-shadow: 0 1px 4px rgba(0, 21, 41, 0.1);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  height: 60px;
  flex-shrink: 0;
}

.header-title {
  font-size: 18px;
  font-weight: 600;
  color: #2c3e50;
}

.header-selector {
  display: flex;
  align-items: center;
  gap: 16px;
}

.main {
  background-color: #f5f7fa;
  padding: 24px;
  flex: 1;
  overflow-y: auto;
  transition: all 0.3s ease;
}

/* 响应式调整 */
@media screen and (max-width: 768px) {
  .layout-container {
    flex-direction: column;
  }
  
  .aside {
    width: 100% !important;
    height: auto;
  }
  
  .logo {
    height: 50px;
  }
  
  .menu {
    display: flex;
    overflow-x: auto;
    height: auto;
  }
  
  .menu :deep(.el-menu-item) {
    flex: 1;
    min-width: 100px;
    margin: 5px;
  }
  
  .main {
    padding: 16px;
  }
}
</style>
