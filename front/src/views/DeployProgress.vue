<template>
  <div class="deploy-progress-page">
    <div class="bg-grid"></div>

    <div class="progress-container fade-in-up">
      <!-- 能量环 -->
      <div class="energy-ring-section">
        <div class="energy-ring" :class="{ done: deployDone, failed: deployFailed }">
          <svg class="ring-svg" viewBox="0 0 120 120">
            <!-- 底环 -->
            <circle cx="60" cy="60" r="52" fill="none" stroke-width="5"
              stroke="rgba(var(--jm-accent-2-rgb), 0.2)" stroke-linecap="round"/>
            <!-- 进度环 -->
            <circle cx="60" cy="60" r="52" fill="none" stroke-width="5"
              :stroke="deployFailed ? 'var(--jm-error-color)' : deployDone ? '#22c55e' : 'var(--jm-primary-1)'"
              stroke-linecap="round"
              :stroke-dasharray="ringCirc"
              :stroke-dashoffset="ringOffset"
              class="ring-progress"
              transform="rotate(-90 60 60)"/>
            <!-- 流光 -->
            <circle v-if="!deployDone && !deployFailed" cx="60" cy="60" r="52" fill="none"
              stroke-width="6" stroke="url(#glow)" stroke-linecap="round"
              stroke-dasharray="20 307" class="ring-glow"
              transform="rotate(-90 60 60)"/>
            <defs>
              <linearGradient id="glow">
                <stop offset="0%" stop-color="var(--jm-primary-1)" stop-opacity="0"/>
                <stop offset="50%" stop-color="var(--jm-primary-1)" stop-opacity="0.8"/>
                <stop offset="100%" stop-color="var(--jm-primary-1)" stop-opacity="0"/>
              </linearGradient>
            </defs>
          </svg>
          <!-- 中心内容 -->
          <div class="ring-center">
            <template v-if="deployDone">
              <svg viewBox="0 0 24 24" width="32" height="32" class="done-check">
                <circle cx="12" cy="12" r="11" fill="rgba(34,197,94,0.1)"/>
                <path d="M8 12l3 3 5-5" stroke="#22c55e" stroke-width="2" fill="none" stroke-linecap="round" stroke-linejoin="round"/>
              </svg>
            </template>
            <template v-else-if="deployFailed">
              <svg viewBox="0 0 24 24" width="32" height="32">
                <circle cx="12" cy="12" r="11" fill="rgba(239,68,68,0.1)"/>
                <path d="M15 9l-6 6M9 9l6 6" stroke="#ef4444" stroke-width="2" fill="none" stroke-linecap="round"/>
              </svg>
            </template>
            <template v-else>
              <span class="ring-percent font-mono">{{ progressPercent }}%</span>
              <span class="ring-label">{{ currentPhase }}</span>
            </template>
          </div>
          <!-- 成功波纹 -->
          <div v-if="deployDone" class="ripple"></div>
        </div>

        <!-- 标题 -->
        <h2 class="progress-title">{{ deployDone ? '部署完成' : deployFailed ? '部署失败' : '正在部署' }}</h2>
        <p class="progress-desc">
          <template v-if="deployDone">OpenClaw 已成功部署并启动</template>
          <template v-else-if="deployFailed">部署过程中出现错误</template>
          <template v-else>{{ isLocal ? '正在编译安装 OpenClaw...' : '正在拉取镜像并启动容器...' }}</template>
        </p>
      </div>

      <!-- 终端日志 -->
      <div class="log-panel">
        <div class="log-header">
          <div class="log-title-row">
            <svg viewBox="0 0 24 24" width="12" height="12" fill="none">
              <rect x="3" y="3" width="18" height="18" rx="2" stroke="currentColor" stroke-width="1.5"/>
              <path d="M8 8l4 4-4 4M14 16h4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
            <span>部署日志</span>
          </div>
          <div v-if="!deployDone && !deployFailed" class="log-live">
            <span class="live-dot"></span>
            LIVE
          </div>
        </div>
        <div class="log-content" ref="logContainer">
          <div v-for="(line, i) in logLines" :key="i" class="log-line">
            <span class="log-num font-mono">{{ String(i + 1).padStart(3, ' ') }}</span>
            <span class="log-time font-mono">{{ line.time }}</span>
            <span class="log-text" :class="line.type">{{ line.text }}</span>
          </div>
          <div v-if="!deployDone && !deployFailed" class="log-cursor">▌</div>
        </div>
      </div>

      <!-- 完成操作 -->
      <div v-if="deployDone" class="done-section spring-in">
        <button class="cta-btn" @click="goToDashboard">
          <svg viewBox="0 0 24 24" width="16" height="16" fill="none">
            <rect x="3" y="3" width="7" height="7" rx="1.5" stroke="currentColor" stroke-width="1.5"/>
            <rect x="14" y="3" width="7" height="7" rx="1.5" stroke="currentColor" stroke-width="1.5"/>
            <rect x="3" y="14" width="7" height="7" rx="1.5" stroke="currentColor" stroke-width="1.5"/>
            <rect x="14" y="14" width="7" height="7" rx="1.5" stroke="currentColor" stroke-width="1.5"/>
          </svg>
          <span class="cta-text">进入仪表板</span>
        </button>
      </div>

      <!-- 失败操作 -->
      <div v-if="deployFailed" class="fail-section spring-in">
        <div class="error-card">
          <div class="error-title">
            <svg viewBox="0 0 24 24" width="14" height="14" fill="none">
              <circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="1.5" opacity="0.5"/>
              <path d="M12 8v4M12 16h.01" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
            </svg>
            排查步骤
          </div>
          <div class="error-steps" v-if="isLocal">
            <p>1. 确保 Node.js 和 pnpm 已正确安装</p>
            <p>2. 确保网络连接正常（需要访问 Gitee 仓库）</p>
            <p>3. 查看上方日志获取详细错误信息</p>
            <p class="error-contact">如仍无法解决，请联系 <b>GMSSH 客服</b>，秒回复帮您免费配置 ✨</p>
          </div>
          <div class="error-steps" v-else>
            <p>1. 确保 Docker 正常运行</p>
            <p>2. 确保网络连接正常</p>
            <p>3. 确认 Docker 镜像加速已正确配置</p>
            <p class="error-contact">如仍无法解决，请联系 <b>GMSSH 客服</b>，秒回复帮您免费配置 ✨</p>
          </div>
        </div>
        <button class="back-capsule" @click="$router.push('/console?step=setup')">
          <svg class="back-arrow" viewBox="0 0 24 24" width="14" height="14" fill="none">
            <path d="M11 19l-7-7 7-7M4 12h16" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
          返回配置
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, inject, onMounted, onUnmounted, nextTick } from 'vue'
import { useRouter, useRoute } from 'vue-router'
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

const ringCirc = 2 * Math.PI * 52 // ≈ 326.73
const ringOffset = computed(() => ringCirc - (ringCirc * progressPercent.value / 100))

const currentPhase = computed(() => {
  const p = progressPercent.value
  if (p < 15) return '初始化...'
  if (p < 40) return isLocal.value ? '编译中...' : '拉取镜像...'
  if (p < 70) return '安装依赖...'
  if (p < 90) return '启动服务...'
  return '即将完成...'
})

let pollTimer = null
let finishedLogged = false  // 守卫：防止并发请求重复输出完成/失败日志

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

    if (!deployDone.value && progressPercent.value < 90) {
      progressPercent.value = Math.min(90, Math.round(progressPercent.value + Math.random() * 15 + 5))
    }

    if (result.finished && !finishedLogged) {
      finishedLogged = true
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
  setTimeout(pollLogs, 800)
})

onUnmounted(() => {
  if (pollTimer) clearInterval(pollTimer)
})
</script>

<style scoped>
/* ===== 页面 ===== */
.deploy-progress-page {
  position: relative;
  width: 100%; height: 100%;
  overflow-y: auto; padding: 24px;
}

.bg-grid {
  position: fixed; inset: 0; z-index: 0; pointer-events: none;
  background-image:
    linear-gradient(rgba(var(--jm-accent-2-rgb), 0.15) 1px, transparent 1px),
    linear-gradient(90deg, rgba(var(--jm-accent-2-rgb), 0.15) 1px, transparent 1px);
  background-size: 40px 40px;
  mask-image: radial-gradient(ellipse 60% 50% at 50% 50%, black 40%, transparent 100%);
  -webkit-mask-image: radial-gradient(ellipse 60% 50% at 50% 50%, black 40%, transparent 100%);
}

.progress-container {
  position: relative; z-index: 1;
  max-width: 540px; margin: 0 auto;
  background: var(--jm-glass-bg);
  backdrop-filter: blur(24px); -webkit-backdrop-filter: blur(24px);
  border: 1px solid var(--jm-glass-border);
  border-radius: 20px;
  padding: 32px 28px 24px;
  box-shadow:
    var(--jm-glass-inner-glow),
    0 8px 40px rgba(0, 0, 0, 0.12),
    0 0 0 1px rgba(255, 255, 255, 0.04);
}

/* ===== 能量环 ===== */
.energy-ring-section {
  display: flex; flex-direction: column; align-items: center;
  margin-bottom: 24px;
}

.energy-ring {
  position: relative;
  width: 140px; height: 140px;
  display: flex; align-items: center; justify-content: center;
  margin-bottom: 16px;
}

.ring-svg { width: 100%; height: 100%; }

.ring-progress {
  transition: stroke-dashoffset 0.8s cubic-bezier(0.4, 0, 0.2, 1), stroke 0.4s;
  filter: drop-shadow(0 0 6px rgba(var(--jm-primary-1-rgb), 0.3));
}

.energy-ring.done .ring-progress {
  filter: drop-shadow(0 0 8px rgba(34, 197, 94, 0.35));
}
.energy-ring.failed .ring-progress {
  filter: drop-shadow(0 0 8px rgba(239, 68, 68, 0.35));
}

.ring-glow {
  animation: glowSpin 2s linear infinite;
  filter: blur(2px);
}
@keyframes glowSpin {
  from { transform: rotate(-90deg); transform-origin: 60px 60px; }
  to { transform: rotate(270deg); transform-origin: 60px 60px; }
}

.ring-center {
  position: absolute; inset: 0;
  display: flex; flex-direction: column;
  align-items: center; justify-content: center;
}

.ring-percent {
  font-size: 28px; font-weight: 700; letter-spacing: -0.02em;
  color: var(--jm-accent-7);
  line-height: 1;
}
.ring-label {
  font-size: 11px; color: var(--jm-accent-4);
  margin-top: 4px;
}

.done-check { animation: checkPop 0.4s cubic-bezier(0.34, 1.56, 0.64, 1) both; }
@keyframes checkPop {
  from { transform: scale(0); opacity: 0; }
  to { transform: scale(1); opacity: 1; }
}

/* 成功波纹 */
.ripple {
  position: absolute; inset: -10px;
  border-radius: 50%;
  border: 2px solid rgba(34, 197, 94, 0.3);
  animation: rippleOut 1s ease-out forwards;
}
@keyframes rippleOut {
  0% { transform: scale(0.8); opacity: 1; }
  100% { transform: scale(1.5); opacity: 0; }
}

.progress-title {
  font-size: 17px; font-weight: 600; color: var(--jm-accent-7);
  margin: 0 0 4px; text-align: center;
}
.progress-desc {
  font-size: 12px; color: var(--jm-accent-4);
  margin: 0; text-align: center;
}

/* ===== 终端日志 ===== */
.log-panel {
  background: rgba(var(--jm-accent-1-rgb), 0.2);
  border: 1px solid rgba(var(--jm-accent-2-rgb), 0.12);
  border-top: 1px solid rgba(255, 255, 255, 0.04);
  border-radius: 14px; overflow: hidden;
}

.log-header {
  display: flex; align-items: center; justify-content: space-between;
  padding: 8px 14px;
  border-bottom: 1px solid rgba(var(--jm-accent-2-rgb), 0.1);
}
.log-title-row {
  display: flex; align-items: center; gap: 6px;
  font-size: 10px; font-weight: 600; letter-spacing: 0.05em;
  text-transform: uppercase; color: var(--jm-accent-4);
}
.log-live {
  display: flex; align-items: center; gap: 5px;
  font-size: 9px; font-weight: 700; letter-spacing: 1px;
  color: #22c55e;
}
.live-dot {
  width: 6px; height: 6px; border-radius: 50%;
  background: #22c55e;
  box-shadow: 0 0 0 2px rgba(34, 197, 94, 0.15), 0 0 6px rgba(34, 197, 94, 0.3);
  animation: breathe 1.5s ease-in-out infinite;
}
@keyframes breathe {
  0%, 100% { opacity: 1; box-shadow: 0 0 0 2px rgba(34,197,94,0.15), 0 0 6px rgba(34,197,94,0.3); }
  50% { opacity: 0.5; box-shadow: 0 0 0 4px rgba(34,197,94,0.08), 0 0 12px rgba(34,197,94,0.15); }
}

.log-content {
  height: 240px; overflow-y: auto;
  padding: 10px 14px;
  font-family: var(--jm-font-mono); font-size: 11px; line-height: 1.7;
}
.log-content::-webkit-scrollbar { width: 4px; }
.log-content::-webkit-scrollbar-thumb { background: var(--jm-accent-3); border-radius: 2px; }

.log-line { display: flex; gap: 0; }
.log-num {
  color: var(--jm-accent-3); flex-shrink: 0;
  width: 30px; text-align: right;
  padding-right: 10px; user-select: none;
  border-right: 1px solid rgba(var(--jm-accent-2-rgb), 0.1);
  margin-right: 10px;
}
.log-time { color: var(--jm-accent-3); flex-shrink: 0; margin-right: 8px; }
.log-text { color: var(--jm-accent-5); word-break: break-all; }
.log-text.error { color: #ef4444; }
.log-text.success { color: #22c55e; }

.log-cursor { color: var(--jm-primary-1); animation: blink 0.8s step-end infinite; margin-left: 40px; }
@keyframes blink { 0%, 100% { opacity: 1; } 50% { opacity: 0; } }

/* ===== 完成区 ===== */
.done-section {
  display: flex; flex-direction: column; align-items: center;
  margin-top: 20px;
}

.cta-btn {
  display: flex; align-items: center; justify-content: center; gap: 8px;
  padding: 10px 32px; height: 42px;
  font-size: 14px; font-weight: 500; color: #fff;
  background: linear-gradient(135deg, var(--jm-primary-1), var(--jm-primary-2));
  border: none; border-radius: 12px; cursor: pointer;
  box-shadow:
    inset 0 1px 0 rgba(255,255,255,0.15),
    0 4px 16px rgba(var(--jm-primary-1-rgb), 0.25),
    0 1px 3px rgba(0,0,0,0.1);
  transition: all 0.2s;
}
.cta-btn:hover {
  transform: translateY(-1px);
  box-shadow:
    inset 0 1px 0 rgba(255,255,255,0.15),
    0 6px 24px rgba(var(--jm-primary-1-rgb), 0.35),
    0 2px 6px rgba(0,0,0,0.12);
}
.cta-btn:active {
  transform: translateY(1px);
  box-shadow:
    inset 0 2px 4px rgba(0,0,0,0.15),
    0 1px 4px rgba(var(--jm-primary-1-rgb), 0.15);
}

/* ===== 失败区 ===== */
.fail-section {
  display: flex; flex-direction: column; align-items: center;
  gap: 14px; margin-top: 20px;
}

.error-card {
  width: 100%;
  background: rgba(239, 68, 68, 0.03);
  border: 1px solid rgba(239, 68, 68, 0.1);
  border-radius: 12px; padding: 14px 16px;
}
.error-title {
  display: flex; align-items: center; gap: 6px;
  font-size: 12px; font-weight: 600; color: #ef4444;
  margin-bottom: 8px;
}
.error-steps p {
  margin: 5px 0; font-size: 11px; color: var(--jm-accent-5);
  line-height: 1.6;
}
.error-steps b { color: var(--jm-accent-7); }
.error-contact {
  margin-top: 10px !important; padding-top: 8px;
  border-top: 1px solid rgba(var(--jm-accent-2-rgb), 0.1);
  color: var(--jm-primary-2) !important;
}

/* 返回胶囊 */
.back-capsule {
  display: flex; align-items: center; gap: 6px;
  padding: 6px 14px 6px 10px; border-radius: 20px;
  border: 1px solid rgba(var(--jm-accent-2-rgb), 0.2);
  background: transparent;
  color: var(--jm-accent-4); font-size: 12px;
  cursor: pointer; transition: all 0.2s;
}
.back-capsule:hover {
  color: var(--jm-primary-2);
  border-color: rgba(var(--jm-primary-1-rgb), 0.2);
  background: rgba(var(--jm-primary-1-rgb), 0.03);
}
.back-arrow { transition: transform 0.2s; }
.back-capsule:hover .back-arrow { animation: arrowBounce 0.4s ease; }
@keyframes arrowBounce {
  0%, 100% { transform: translateX(0); }
  50% { transform: translateX(-4px); }
}

/* ===== 动画 ===== */
.spring-in { animation: springUp 0.5s cubic-bezier(0.34, 1.56, 0.64, 1) both; }
@keyframes springUp {
  from { opacity: 0; transform: translateY(20px) scale(0.95); }
  to { opacity: 1; transform: translateY(0) scale(1); }
}
.fade-in-up { animation: fadeInUp 0.35s ease-out both; }
@keyframes fadeInUp {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}
</style>
