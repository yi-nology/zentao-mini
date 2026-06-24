<template>
  <div class="init-guide">
    <div class="init-card">
      <div class="init-icon">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="32" height="32">
          <path d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
        </svg>
      </div>
      <h1 class="init-title">系统初始化</h1>
      <p class="init-subtitle">请上传加密配置文件以完成系统初始化</p>

      <form @submit.prevent="submitForm" class="init-form">
        <div class="upload-area" @click="triggerFileInput" @dragover.prevent @drop.prevent="handleDrop">
          <input type="file" ref="fileInput" @change="handleFileChange" accept=".json" style="display: none" />
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="40" height="40" class="upload-icon">
            <path d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
          </svg>
          <p class="upload-text">点击或拖拽文件到此处</p>
          <p class="upload-hint">支持 .json 格式的配置文件</p>
        </div>

        <div v-if="selectedFile" class="file-info">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16">
            <path d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
          </svg>
          <span>{{ selectedFile.name }} ({{ formatFileSize(selectedFile.size) }})</span>
        </div>

        <div class="hint-box">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16">
            <path d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <span>请上传使用 generate-encryption.sh 脚本生成的 auth-config.json 文件</span>
        </div>

        <button type="submit" class="init-btn" :disabled="loading || !selectedFile">
          {{ loading ? '初始化中...' : '开始初始化' }}
        </button>
      </form>

      <div v-if="error" class="message error">{{ error }}</div>
      <div v-if="success" class="message success">
        {{ success }}
        <div class="success-actions">
          <router-link to="/" class="action-btn primary">进入系统</router-link>
          <button @click="testZentao" class="action-btn secondary" :disabled="testing">
            {{ testing ? '测试中...' : '测试禅道连接' }}
          </button>
        </div>
      </div>
      <div v-if="testResult" class="result-box">
        <h3>测试结果</h3>
        <pre>{{ testResult }}</pre>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { uploadInitConfig, testZentaoConnection } from '@/api/zentao'

const router = useRouter()
const fileInput = ref<HTMLInputElement | null>(null)
const selectedFile = ref<File | null>(null)
const loading = ref<boolean>(false)
const error = ref<string>('')
const success = ref<string>('')
const testing = ref<boolean>(false)
const testResult = ref<string>('')

const triggerFileInput = (): void => { fileInput.value?.click() }

const handleDrop = (event: DragEvent): void => {
  const files = event.dataTransfer?.files
  if (files && files.length > 0) {
    const file = files[0]
    if (!file.name.endsWith('.json')) { error.value = '请上传 .json 格式的配置文件'; return }
    selectedFile.value = file; error.value = ''
  }
}

const handleFileChange = (event: Event): void => {
  const target = event.target as HTMLInputElement
  if (target.files && target.files.length > 0) {
    const file = target.files[0]
    if (!file.name.endsWith('.json')) { error.value = '请上传 .json 格式的配置文件'; selectedFile.value = null; target.value = ''; return }
    selectedFile.value = file; error.value = ''
  }
}

const formatFileSize = (bytes: number): string => {
  if (bytes === 0) return '0 B'
  const k = 1024; const sizes = ['B', 'KB', 'MB', 'GB']; const i = Math.floor(Math.log(bytes) / Math.log(k))
  return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i]
}

const submitForm = async (): Promise<void> => {
  if (!selectedFile.value) { error.value = '请选择加密配置文件'; return }
  loading.value = true; error.value = ''; success.value = ''
  try {
    const formData = new FormData(); formData.append('configFile', selectedFile.value)
    const response = await uploadInitConfig(formData)
    if (response.code !== 200) throw new Error(response.message || '初始化失败')
    success.value = '初始化成功！系统已准备就绪，即将跳转到主页...'
    setTimeout(() => { router.push('/') }, 2000)
  } catch (err) {
    const msg = err instanceof Error ? err.message : String(err)
    error.value = msg.includes('Network') ? '网络错误，请检查网络连接后重试。' : msg.includes('timeout') ? '请求超时，请稍后重试。' : '初始化失败：' + msg
  } finally { loading.value = false }
}

const testZentao = async (): Promise<void> => {
  testing.value = true; testResult.value = ''
  try {
    const response = await testZentaoConnection()
    if (response.code !== 200) throw new Error(response.message || '测试失败')
    testResult.value = JSON.stringify(response, null, 2)
  } catch (err) { testResult.value = '测试失败: ' + (err instanceof Error ? err.message : String(err)) } finally { testing.value = false }
}
</script>

<style scoped>
.init-guide {
  min-height: 100vh;
  display: flex;
  justify-content: center;
  align-items: center;
  background-color: #F1F5F9;
  padding: 20px;
}

.init-card {
  background: var(--color-bg-card);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-lg);
  padding: 48px;
  max-width: 440px;
  width: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 32px;
}

.init-icon {
  width: 64px;
  height: 64px;
  border-radius: var(--radius-lg);
  background-color: var(--color-primary-light);
  color: var(--color-primary);
  display: flex;
  align-items: center;
  justify-content: center;
}

.init-title {
  font-size: 24px;
  font-weight: 700;
  color: var(--color-text-primary);
  margin: 0;
}

.init-subtitle {
  font-size: 14px;
  color: var(--color-text-secondary);
  margin: 0;
  text-align: center;
}

.init-form {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.upload-area {
  border: 2px dashed var(--color-border);
  border-radius: var(--radius-md);
  padding: 32px;
  text-align: center;
  cursor: pointer;
  transition: all var(--transition-fast);
  background-color: #F8FAFC;
}

.upload-area:hover {
  border-color: var(--color-primary);
  background-color: var(--color-primary-light);
}

.upload-icon {
  color: var(--color-text-tertiary);
  margin-bottom: 12px;
}

.upload-text {
  font-size: 14px;
  color: var(--color-text-secondary);
  margin: 0 0 4px;
}

.upload-hint {
  font-size: 12px;
  color: var(--color-text-tertiary);
  margin: 0;
}

.file-info {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  background-color: var(--color-bg-hover);
  border-radius: var(--radius-sm);
  font-size: 13px;
  color: var(--color-text-primary);
}

.hint-box {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 14px;
  background-color: var(--color-primary-light);
  border-radius: var(--radius-sm);
  font-size: 12px;
  color: var(--color-primary);
  line-height: 1.5;
}

.init-btn {
  width: 100%;
  padding: 14px 32px;
  background-color: var(--color-primary);
  color: var(--color-text-on-primary);
  border: none;
  border-radius: var(--radius-sm);
  font-size: 15px;
  font-weight: 600;
  cursor: pointer;
  transition: background-color var(--transition-fast);
}

.init-btn:hover:not(:disabled) {
  background-color: var(--color-primary-hover);
}

.init-btn:disabled {
  background-color: #CBD5E1;
  cursor: not-allowed;
}

.message {
  padding: 16px;
  border-radius: var(--radius-sm);
  font-size: 14px;
  text-align: center;
}

.message.error {
  background-color: var(--color-danger-light);
  color: #991B1B;
}

.message.success {
  background-color: var(--color-success-light);
  color: #166534;
}

.success-actions {
  display: flex;
  gap: 12px;
  margin-top: 16px;
  justify-content: center;
}

.action-btn {
  padding: 10px 20px;
  border-radius: var(--radius-sm);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  text-decoration: none;
  border: none;
  transition: all var(--transition-fast);
}

.action-btn.primary {
  background-color: var(--color-primary);
  color: var(--color-text-on-primary);
}

.action-btn.secondary {
  background-color: var(--color-bg-card);
  color: var(--color-text-primary);
  border: 1px solid var(--color-border);
}

.result-box {
  width: 100%;
  padding: 16px;
  background-color: var(--color-bg);
  border: 1px solid var(--color-border-light);
  border-radius: var(--radius-sm);
  font-size: 13px;
}

.result-box h3 {
  margin: 0 0 8px;
  font-size: 14px;
  color: var(--color-text-primary);
}

.result-box pre {
  margin: 0;
  padding: 12px;
  background-color: var(--color-bg-hover);
  border-radius: var(--radius-sm);
  font-size: 12px;
  overflow-x: auto;
  white-space: pre-wrap;
  font-family: var(--font-mono);
}
</style>
