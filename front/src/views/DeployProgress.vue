<template>
  <div class="deploy-progress-page">
    <div class="progress-container fade-in-up">
      <!-- 标题 -->
      <div class="progress-header">
        <div class="header-icon" :class="{ 'icon-done': deployDone }">
          <svg v-if="!deployDone" viewBox="0 0 24 24" width="28" height="28" class="spin-icon">
            <circle cx="12" cy="12" r="10" stroke="var(--jm-primary-1)" stroke-width="2" fill="none" opacity="0.2"/>
            <path d="M12 2a10 10 0 019.95 9" stroke="var(--jm-primary-1)" stroke-width="2" fill="none" stroke-linecap="round"/>
          </svg>
          <svg v-else viewBox="0 0 24 24" width="28" height="28">
            <circle cx="12" cy="12" r="10" fill="var(--jm-success-color)" opacity="0.15"/>
            <path d="M8 12l3 3 5-5" stroke="var(--jm-success-color)" stroke-width="2" fill="none" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
        </div>
        <div class="header-text">
          <h2 class="progress-title">{{ deployDone ? '部署完成' : '正在部署' }}</h2>
          <p class="progress-desc" v-if="!deployDone">{{ isLocal ? '正在编译安装 OpenClaw，请稍候...' : '正在拉取镜像并启动容器，请稍候...' }}</p>
          <p class="progress-desc" v-else>OpenClaw 已成功部署并启动</p>
        </div>
      </div>

      <!-- 进度条 -->
      <div class="progress-bar-section">
        <n-progress
          type="line"
          :percentage="progressPercent"
          :status="deployDone ? 'success' : 'default'"
          :show-indicator="true"
          :height="6"
          :border-radius="3"
        />
      </div>

      <!-- 日志面板 -->
      <div class="log-panel">
        <div class="log-header">
          <div class="log-title-row">
            <svg viewBox="0 0 24 24" width="14" height="14" fill="none">
              <rect x="3" y="3" width="18" height="18" rx="2" stroke="var(--jm-accent-4)" stroke-width="1.5"/>
              <path d="M8 8l4 4-4 4M14 16h4" stroke="var(--jm-accent-4)" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
            <span>部署日志</span>
          </div>
          <div v-if="!deployDone" class="log-live-indicator">
            <span class="live-dot"></span>
            LIVE
          </div>
        </div>
        <div class="log-content" ref="logContainer">
          <div v-for="(line, i) in logLines" :key="i" class="log-line">
            <span class="log-time">{{ line.time }}</span>
            <span class="log-text" :class="line.type">{{ line.text }}</span>
          </div>
          <div v-if="!deployDone" class="log-cursor">▌</div>
        </div>
      </div>

      <!-- 完成操作 -->
      <div v-if="deployDone" class="done-actions fade-in-up">
        <n-button type="primary" size="large" @click="goToDashboard" class="dashboard-btn">
          <template #icon>
            <svg viewBox="0 0 24 24" width="16" height="16" fill="none">
              <rect x="3" y="3" width="7" height="7" rx="1.5" stroke="currentColor" stroke-width="1.5"/>
              <rect x="14" y="3" width="7" height="7" rx="1.5" stroke="currentColor" stroke-width="1.5"/>
              <rect x="3" y="14" width="7" height="7" rx="1.5" stroke="currentColor" stroke-width="1.5"/>
              <rect x="14" y="14" width="7" height="7" rx="1.5" stroke="currentColor" stroke-width="1.5"/>
            </svg>
          </template>
          进入仪表板
        </n-button>
      </div>

      <!-- 失败操作 -->
      <div v-if="deployFailed" class="done-actions fade-in-up">
        <div class="error-hint">
          <svg viewBox="0 0 24 24" width="16" height="16" fill="none">
            <circle cx="12" cy="12" r="10" stroke="var(--jm-error-color)" stroke-width="1.5" opacity="0.5"/>
            <path d="M12 8v4M12 16h.01" stroke="var(--jm-error-color)" stroke-width="1.5" stroke-linecap="round"/>
          </svg>
          <span>部署过程中出现错误，请按以下步骤排查：</span>
        </div>
        <div class="error-checklist" v-if="isLocal">
          <p>1. 确保 Node.js 和 pnpm 已正确安装</p>
          <p>2. 确保网络连接正常（需要访问 Gitee 仓库）</p>
          <p>3. 查看上方日志获取详细错误信息</p>
          <p class="error-contact">如仍无法解决，请联系 <b>GMSSH 客服</b>，秒回复帮您免费配置 ✨</p>
        </div>
        <div class="error-checklist" v-else>
          <p>1. 确保 Docker 正常运行</p>
          <p>2. 确保网络连接正常</p>
          <p>3. 进入 <b>Docker 管理器 → 设置 → 镜像加速</b>，确认已正确配置</p>
          <p class="error-contact">如仍无法解决，请联系 <b>GMSSH 客服</b>，秒回复帮您免费配置 ✨</p>
        </div>
        <n-button @click="$router.push('/console?step=setup')" quaternary type="primary">返回配置</n-button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, inject, onMounted, onUnmounted, nextTick } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { NButton, NProgress } from 'naive-ui'
import { getDeployLogs } from '@/api/deploy'

const router = useRouter()
const route = useRoute()
const logContainer = ref(null)
const logLines = ref([])
const deployDone = ref(false)
const deployFailed = ref(false)
const setDeployed = inject('setDeployed', () => {})
const progressPercent = ref(0)
const isLocal = ref(route.query.mode === 'local')

let pollTimer = null

function getTime() {
  const d = new Date()
  return `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}:${String(d.getSeconds()).padStart(2, '0')}`
}

function addLog(text, type = 'info') {
  logLines.value.push({ time: getTime(), text, type })
  nextTick(() => {
    if (logContainer.value) {
      logContainer.value.scrollTop = logContainer.value.scrollHeight
    }
  })
}

async function pollLogs() {
  try {
    const result = await getDeployLogs()

    if (result.logs && result.logs.length > 0) {
      result.logs.forEach(line => {
        addLog(line, line.toLowerCase().includes('error') ? 'error' : 'info')
      })
    }

    // 模拟进度递增
    if (!deployDone.value && progressPercent.value < 90) {
      progressPercent.value = Math.min(90, Math.round(progressPercent.value + Math.random() * 15 + 5))
    }

    if (result.finished) {
      clearInterval(pollTimer)
      pollTimer = null

      if (result.success) {
        progressPercent.value = 100
        addLog('✅ 部署完成，所有服务已启动', 'success')
        deployDone.value = true
      } else {
        addLog('❌ 部署失败', 'error')
        deployFailed.value = true
      }
    }
  } catch (e) {
    addLog('⚠ 获取日志失败: ' + (e.message || ''), 'error')
  }
}

function goToDashboard() {
  setDeployed(true)
  router.push({ path: '/console' })
}

onMounted(() => {
  addLog('🚀 开始部署 OpenClaw...', 'info')
  addLog(isLocal.value ? '📦 检查 Node.js 环境...' : '📦 检查 Docker 环境...', 'info')
  progressPercent.value = 5

  pollTimer = setInterval(pollLogs, 2000)
  // 首次立即拉取
  setTimeout(pollLogs, 800)
})

onUnmounted(() => {
  if (pollTimer) {
    clearInterval(pollTimer)
  }
})
</script>

<style scoped>
.deploy-progress-page {
  width: 100%;
  height: 100%;
  overflow-y: auto;
  padding: 24px;
}

.progress-container {
  max-width: 640px;
  margin: 0 auto;
}

.progress-header {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 24px;
}

.header-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 52px;
  height: 52px;
  border-radius: 14px;
  background: rgba(var(--jm-primary-1-rgb), 0.08);
  flex-shrink: 0;
}

.header-icon.icon-done {
  background: rgba(59, 173, 91, 0.08);
}

.spin-icon {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.progress-title {
  font-size: 18px;
  font-weight: 600;
  color: var(--jm-accent-7);
  margin-bottom: 4px;
}

.progress-desc {
  font-size: 13px;
  color: var(--jm-accent-5);
}

.progress-bar-section {
  margin-bottom: 20px;
}

/* 日志面板 */
.log-panel {
  background: rgba(var(--jm-accent-1-rgb), 0.6);
  border: 1px solid var(--jm-accent-2);
  border-radius: 10px;
  overflow: hidden;
}

.log-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 14px;
  border-bottom: 1px solid var(--jm-accent-2);
}

.log-title-row {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: var(--jm-accent-5);
  font-weight: 500;
}

.log-live-indicator {
  display: flex;
  align-items: center;
  gap: 5px;
  font-size: 10px;
  font-weight: 600;
  color: var(--jm-success-color);
  letter-spacing: 1px;
}

.live-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--jm-success-color);
  animation: pulse 1s ease-in-out infinite;
}

.log-content {
  height: 320px;
  overflow-y: auto;
  padding: 12px 14px;
  font-family: 'SF Mono', 'Fira Code', 'Consolas', monospace;
  font-size: 12px;
  line-height: 1.7;
}

.log-line {
  display: flex;
  gap: 10px;
}

.log-time {
  color: var(--jm-accent-4);
  flex-shrink: 0;
  font-size: 11px;
}

.log-text {
  color: var(--jm-accent-6);
  word-break: break-all;
}

.log-text.error {
  color: var(--jm-error-color);
}

.log-text.success {
  color: var(--jm-success-color);
}

.log-cursor {
  color: var(--jm-primary-1);
  animation: pulse 0.8s ease-in-out infinite;
}

/* 完成操作 */
.done-actions {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  margin-top: 24px;
}

.dashboard-btn {
  min-width: 180px;
  height: 42px;
  font-size: 14px;
  font-weight: 500;
  border-radius: 10px;
}

.dashboard-btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 16px rgba(var(--jm-primary-1-rgb), 0.3);
}

.error-hint {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: var(--jm-error-color);
}

.error-checklist {
  background: rgba(var(--jm-accent-1-rgb), 0.5);
  border: 1px solid var(--jm-accent-2);
  border-radius: 8px;
  padding: 12px 16px;
  text-align: left;
  width: 100%;
  max-width: 400px;
}
.error-checklist p {
  margin: 6px 0;
  font-size: 12px;
  color: var(--jm-accent-5);
  line-height: 1.6;
}
.error-checklist b { color: var(--jm-accent-7); }
.error-contact {
  margin-top: 10px !important;
  padding-top: 8px;
  border-top: 1px solid var(--jm-accent-2);
  color: var(--jm-primary-2) !important;
}
</style>
