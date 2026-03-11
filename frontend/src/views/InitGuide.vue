<template>
  <div class="init-guide">
    <div class="init-guide-container">
      <h1>系统初始化</h1>
      <p class="subtitle">请上传加密配置文件以完成系统初始化</p>
      
      <form @submit.prevent="submitForm" class="init-form">
        <div class="form-group">
          <label for="configFile">加密配置文件</label>
          <input 
            type="file" 
            id="configFile" 
            ref="fileInput"
            @change="handleFileChange"
            accept=".json"
            required
          >
          <p class="hint">请上传使用 generate-encryption.sh 脚本生成的 auth-config.json 文件</p>
        </div>
        
        <div v-if="selectedFile" class="file-info">
          <p>已选择文件: {{ selectedFile.name }}</p>
        </div>
        
        <div class="form-actions">
          <button type="submit" class="btn primary" :disabled="loading || !selectedFile">
            {{ loading ? '初始化中...' : '开始初始化' }}
          </button>
        </div>
      </form>
      
      <div v-if="error" class="error-message">
        {{ error }}
      </div>
      
      <div v-if="success" class="success-message">
        {{ success }}
        <router-link to="/" class="btn secondary">进入系统</router-link>
        <button @click="testZentao" class="btn secondary" :disabled="testing">
          {{ testing ? '测试中...' : '测试禅道连接' }}
        </button>
      </div>
      
      <div v-if="testResult" class="test-result">
        <h3>测试结果</h3>
        <pre>{{ testResult }}</pre>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'InitGuide',
  data() {
    return {
      selectedFile: null,
      loading: false,
      error: '',
      success: '',
      testing: false,
      testResult: ''
    }
  },
  methods: {
    handleFileChange(event) {
      this.selectedFile = event.target.files[0]
    },
    async submitForm() {
      if (!this.selectedFile) {
        this.error = '请选择加密配置文件'
        return
      }
      
      this.loading = true
      this.error = ''
      this.success = ''
      
      try {
        // 创建FormData对象
        const formData = new FormData()
        formData.append('configFile', this.selectedFile)
        
        // 上传文件到后端接口
        const response = await fetch('/api/init/upload', {
          method: 'POST',
          body: formData
        })
        
        if (!response.ok) {
          throw new Error('上传失败')
        }
        
        const result = await response.json()
        if (result.code !== 200) {
          throw new Error(result.message || '初始化失败')
        }
        
        // 设置初始化标志
        localStorage.setItem('initialized', 'true')
        
        // 显示成功消息
        this.success = '初始化成功！系统已准备就绪。'
        
        // 3秒后跳转到首页
        setTimeout(() => {
          this.$router.push('/')
        }, 3000)
      } catch (err) {
        this.error = '初始化失败，请检查上传的文件并重试。'
        console.error('初始化错误:', err)
      } finally {
        this.loading = false
      }
    },
    async testZentao() {
      this.testing = true
      this.testResult = ''
      
      try {
        // 调用当前用户接口测试禅道连接
        const response = await fetch('/api/users/current')
        
        if (!response.ok) {
          throw new Error('测试失败')
        }
        
        const result = await response.json()
        if (result.code !== 200) {
          throw new Error(result.message || '测试失败')
        }
        
        // 显示测试结果
        this.testResult = JSON.stringify(result, null, 2)
      } catch (err) {
        this.testResult = '测试失败: ' + err.message
        console.error('测试禅道连接错误:', err)
      } finally {
        this.testing = false
      }
    }
  }
}
</script>

<style scoped>
.init-guide {
  min-height: 100vh;
  display: flex;
  justify-content: center;
  align-items: center;
  background-color: #f5f5f5;
  padding: 20px;
}

.init-guide-container {
  background-color: white;
  border-radius: 8px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
  padding: 40px;
  max-width: 500px;
  width: 100%;
}

h1 {
  color: #333;
  margin-bottom: 10px;
  font-size: 24px;
  text-align: center;
}

.subtitle {
  color: #666;
  margin-bottom: 30px;
  text-align: center;
  font-size: 16px;
}

.init-form {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.form-group label {
  color: #333;
  font-weight: bold;
  font-size: 14px;
}

.form-group input {
  padding: 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 16px;
  transition: border-color 0.3s ease;
}

.form-group input:focus {
  outline: none;
  border-color: #4CAF50;
  box-shadow: 0 0 0 2px rgba(76, 175, 80, 0.1);
}

.form-group .hint {
  color: #666;
  font-size: 12px;
  margin-top: 5px;
  margin-bottom: 0;
}

.file-info {
  margin: 15px 0;
  padding: 10px;
  background-color: #f0f0f0;
  border-radius: 4px;
  font-size: 14px;
}

.form-actions {
  margin-top: 10px;
}

.btn {
  display: inline-block;
  padding: 12px 24px;
  border: none;
  border-radius: 4px;
  font-weight: bold;
  cursor: pointer;
  transition: background-color 0.3s ease;
  text-align: center;
  text-decoration: none;
}

.btn.primary {
  background-color: #4CAF50;
  color: white;
  width: 100%;
}

.btn.primary:hover {
  background-color: #45a049;
}

.btn.primary:disabled {
  background-color: #cccccc;
  cursor: not-allowed;
}

.btn.secondary {
  background-color: #2196F3;
  color: white;
  margin-top: 15px;
}

.btn.secondary:hover {
  background-color: #0b7dda;
}

.error-message {
  margin-top: 20px;
  padding: 15px;
  background-color: #f8d7da;
  color: #721c24;
  border-radius: 4px;
  font-size: 14px;
}

.success-message {
  margin-top: 20px;
  padding: 15px;
  background-color: #d4edda;
  color: #155724;
  border-radius: 4px;
  font-size: 14px;
  text-align: center;
}

.test-result {
  margin-top: 20px;
  padding: 15px;
  background-color: #f8f9fa;
  border: 1px solid #dee2e6;
  border-radius: 4px;
  font-size: 14px;
}

.test-result h3 {
  margin-top: 0;
  color: #333;
  font-size: 16px;
}

.test-result pre {
  margin: 10px 0 0 0;
  padding: 10px;
  background-color: #e9ecef;
  border-radius: 4px;
  font-size: 12px;
  overflow-x: auto;
  white-space: pre-wrap;
  word-wrap: break-word;
}
</style>
