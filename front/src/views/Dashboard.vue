<template>
  <div class="dashboard-layout">
    <!-- 左侧边栏 -->
    <aside class="sidebar">
      <div class="sidebar-logo">
        <img src="@/openclaw.png" alt="GMClaw" class="logo-img" />
        <span class="logo-text">GMClaw</span>
      </div>

      <nav class="sidebar-nav">
        <button
          v-for="item in menuItems"
          :key="item.key"
          class="nav-item"
          :class="{ active: activeMenu === item.key }"
          @click="activeMenu = item.key"
        >
          <span class="nav-icon" v-html="item.icon"></span>
          <span class="nav-label">{{ item.label }}</span>
        </button>
      </nav>

      <div class="sidebar-footer">
        <div class="status-badge" :class="running ? 'online' : 'offline'">
          <div class="status-dot"></div>
          <span>{{ running ? '运行中' : '已停止' }}</span>
        </div>
      </div>
    </aside>

    <!-- 右侧内容区 -->
    <main class="content">
      <!-- 控制台页面 -->
      <div v-if="activeMenu === 'console'" class="console-page">
        <!-- 状态栏 -->
        <div class="status-bar" :class="running ? 'bar-ok' : 'bar-off'">
          <div class="bar-left">
            <svg v-if="running" viewBox="0 0 24 24" width="18" height="18" fill="none">
              <circle cx="12" cy="12" r="10" stroke="var(--jm-success-color)" stroke-width="1.5" opacity="0.4"/>
              <path d="M9 12l2 2 4-4" stroke="var(--jm-success-color)" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
            <svg v-else viewBox="0 0 24 24" width="18" height="18" fill="none">
              <circle cx="12" cy="12" r="10" stroke="var(--jm-error-color)" stroke-width="1.5" opacity="0.4"/>
              <path d="M15 9l-6 6M9 9l6 6" stroke="var(--jm-error-color)" stroke-width="2" stroke-linecap="round"/>
            </svg>
            <div>
              <span class="bar-status">{{ statusData.containerName }}</span>
              <span class="bar-sub">{{ statusData.image }}</span>
            </div>
          </div>
          <n-button quaternary size="tiny" @click="refreshStatus">
            <svg viewBox="0 0 24 24" width="14" height="14" fill="none">
              <path d="M1 4v6h6M23 20v-6h-6" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
              <path d="M20.49 9A9 9 0 005.64 5.64L1 10m22 4l-4.64 4.36A9 9 0 013.51 15" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
          </n-button>
        </div>

        <!-- 访问地址 -->
        <div v-if="running && accessUrl" class="url-card">
          <div class="url-header">
            <svg viewBox="0 0 24 24" width="14" height="14" fill="none">
              <path d="M10 13a5 5 0 007.54.54l3-3a5 5 0 00-7.07-7.07l-1.72 1.71" stroke="var(--jm-primary-1)" stroke-width="1.5" stroke-linecap="round"/>
              <path d="M14 11a5 5 0 00-7.54-.54l-3 3a5 5 0 007.07 7.07l1.71-1.71" stroke="var(--jm-primary-2)" stroke-width="1.5" stroke-linecap="round"/>
            </svg>
            <span>控制台地址</span>
          </div>
          <div class="url-body">
            <code>{{ accessUrl }}</code>
            <button class="copy-btn" @click="copyUrl" title="复制">
              <svg viewBox="0 0 24 24" width="13" height="13" fill="none">
                <rect x="9" y="9" width="13" height="13" rx="2" stroke="currentColor" stroke-width="1.5"/>
                <path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1" stroke="currentColor" stroke-width="1.5"/>
              </svg>
            </button>
          </div>
        </div>

        <!-- 信息网格 -->
        <div class="info-grid">
          <div class="info-item">
            <span class="info-label">运行时间</span>
            <span class="info-value">{{ statusData.uptime || '-' }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">Web 端口</span>
            <span class="info-value mono">:{{ statusData.webPort || 18789 }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">通讯端口</span>
            <span class="info-value mono">:{{ statusData.bridgePort || 18790 }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">容器状态</span>
            <span class="info-value">{{ statusData.status || 'unknown' }}</span>
          </div>
        </div>

        <!-- 快捷操作 -->
        <div class="actions-row">
          <button class="action-btn primary" @click="openWebUI">
            <svg viewBox="0 0 24 24" width="16" height="16" fill="none">
              <path d="M18 13v6a2 2 0 01-2 2H5a2 2 0 01-2-2V8a2 2 0 012-2h6" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
              <polyline points="15,3 21,3 21,9" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
              <line x1="10" y1="14" x2="21" y2="3" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
            </svg>
            打开 Web UI
          </button>
          <button class="action-btn" @click="viewLogs">
            <svg viewBox="0 0 24 24" width="16" height="16" fill="none">
              <rect x="3" y="3" width="18" height="18" rx="2" stroke="currentColor" stroke-width="1.5"/>
              <path d="M8 8l4 4-4 4M14 16h4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
            查看日志
          </button>
          <button class="action-btn" @click="openConfig">
            <svg viewBox="0 0 24 24" width="16" height="16" fill="none">
              <circle cx="12" cy="12" r="3" stroke="currentColor" stroke-width="1.5"/>
              <path d="M12 1v2M12 21v2M4.22 4.22l1.42 1.42M18.36 18.36l1.42 1.42M1 12h2M21 12h2M4.22 19.78l1.42-1.42M18.36 5.64l1.42-1.42" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
            </svg>
            查看配置
          </button>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { NButton } from 'naive-ui'
import { getClawStatus } from '@/api/deploy'
import gm from '@/utils/gmssh'

const route = useRoute()
const activeMenu = ref('console')

const menuItems = [
  {
    key: 'console',
    label: '控制台',
    icon: '<svg viewBox="0 0 24 24" width="18" height="18" fill="none"><rect x="3" y="3" width="7" height="7" rx="1.5" stroke="currentColor" stroke-width="1.5"/><rect x="14" y="3" width="7" height="7" rx="1.5" stroke="currentColor" stroke-width="1.5"/><rect x="3" y="14" width="7" height="7" rx="1.5" stroke="currentColor" stroke-width="1.5"/><rect x="14" y="14" width="7" height="7" rx="1.5" stroke="currentColor" stroke-width="1.5"/></svg>',
  },
]

const statusData = ref({
  running: false,
  containerName: 'gmssh-openclaw',
  status: 'unknown',
  webPort: 18789,
  bridgePort: 18790,
  uptime: '-',
  image: 'gmssh/openclaw:2026.02.17',
})

const running = computed(() => statusData.value.running)

const accessUrl = computed(() => {
  const ip = gm.getPublicIp()
  const port = statusData.value.webPort || 18789
  const token = route.query.token || sessionStorage.getItem('deploy_token') || ''
  if (!token) return ''
  return `http://${ip}:${port}/?token=${token}`
})

async function fetchStatus() {
  try {
    const result = await getClawStatus()
    statusData.value = { ...statusData.value, ...result }
  } catch (e) {
    console.error('获取状态失败:', e)
  }
}

function openWebUI() {
  if (accessUrl.value) {
    window.open(accessUrl.value, '_blank')
  } else {
    const port = statusData.value.webPort || 18789
    const ip = gm.getPublicIp()
    window.open(`http://${ip}:${port}`, '_blank')
  }
}

function copyUrl() {
  if (!accessUrl.value) return
  navigator.clipboard.writeText(accessUrl.value)
    .then(() => gm.success('已复制'))
    .catch(() => gm.warning('复制失败'))
}

function viewLogs() {
  const gmApi = gm.getGmApi()
  if (gmApi?.openShell) {
    gmApi.openShell({ arr: ['docker logs -f gmssh-openclaw\n'] })
  } else {
    gm.info('请在终端执行: docker logs -f gmssh-openclaw')
  }
}

function refreshStatus() {
  fetchStatus()
  gm.success('状态已刷新')
}

function openConfig() {
  const gmApi = gm.getGmApi()
  if (gmApi?.openCodeEditor) {
    gmApi.openCodeEditor('/opt/gmclaw/conf/openclaw.json')
  } else {
    gm.info('配置文件: /opt/gmclaw/conf/openclaw.json')
  }
}

onMounted(() => {
  // 保存 token 以便后续访问
  if (route.query.token) {
    sessionStorage.setItem('deploy_token', route.query.token)
  }
  fetchStatus()
})
</script>

<style scoped>
.dashboard-layout {
  display: flex;
  width: 100%;
  height: 100%;
  overflow: hidden;
}

/* ========== 侧边栏 ========== */
.sidebar {
  width: 180px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  background: rgba(var(--jm-accent-1-rgb), 0.5);
  border-right: 1px solid var(--jm-accent-2);
}

.sidebar-logo {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 14px 16px;
  border-bottom: 1px solid var(--jm-accent-2);
}

.logo-img {
  width: 26px;
  height: 26px;
  border-radius: 6px;
}

.logo-text {
  font-size: 14px;
  font-weight: 600;
  background: linear-gradient(135deg, var(--jm-primary-2), var(--jm-primary-1));
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}

.sidebar-nav {
  flex: 1;
  padding: 8px;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 9px 10px;
  border-radius: 8px;
  border: none;
  background: transparent;
  color: var(--jm-accent-5);
  cursor: pointer;
  font-size: 13px;
  transition: all 0.15s;
  text-align: left;
}

.nav-item:hover {
  background: rgba(var(--jm-accent-1-rgb), 0.6);
  color: var(--jm-accent-7);
}

.nav-item.active {
  background: rgba(var(--jm-primary-1-rgb), 0.1);
  color: var(--jm-primary-2);
}

.nav-icon {
  display: flex;
  align-items: center;
  width: 18px;
  height: 18px;
  flex-shrink: 0;
}

.sidebar-footer {
  padding: 12px 16px;
  border-top: 1px solid var(--jm-accent-2);
}

.status-badge {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 11px;
}

.status-badge.online { color: var(--jm-success-color); }
.status-badge.offline { color: var(--jm-error-color); }

.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: currentColor;
  box-shadow: 0 0 6px currentColor;
}

/* ========== 内容区 ========== */
.content {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
  min-width: 0;
}

.console-page {
  max-width: 600px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

/* 状态栏 */
.status-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 14px;
  border-radius: 10px;
  border: 1px solid var(--jm-accent-2);
  background: rgba(var(--jm-accent-1-rgb), 0.4);
}

.bar-left {
  display: flex;
  align-items: center;
  gap: 10px;
}

.bar-left > div {
  display: flex;
  flex-direction: column;
}

.bar-status {
  font-size: 13px;
  font-weight: 500;
  color: var(--jm-accent-7);
}

.bar-sub {
  font-size: 11px;
  color: var(--jm-accent-4);
}

/* URL 卡片 */
.url-card {
  border: 1px solid var(--jm-accent-2);
  border-radius: 10px;
  padding: 10px 14px;
  background: rgba(var(--jm-accent-1-rgb), 0.3);
}

.url-header {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 11px;
  color: var(--jm-accent-5);
  margin-bottom: 6px;
}

.url-body {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.url-body code {
  font-size: 12px;
  color: var(--jm-primary-2);
  font-family: 'SF Mono', Consolas, monospace;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.copy-btn {
  flex-shrink: 0;
  padding: 4px;
  border: none;
  background: transparent;
  color: var(--jm-accent-5);
  cursor: pointer;
  border-radius: 4px;
  display: flex;
  transition: color 0.15s;
}
.copy-btn:hover {
  color: var(--jm-primary-2);
}

/* 信息网格 */
.info-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px;
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding: 10px 12px;
  border-radius: 8px;
  background: rgba(var(--jm-accent-1-rgb), 0.4);
  border: 1px solid rgba(var(--jm-accent-2-rgb, 255,255,255), 0.05);
}

.info-label {
  font-size: 11px;
  color: var(--jm-accent-4);
}

.info-value {
  font-size: 13px;
  font-weight: 500;
  color: var(--jm-accent-7);
}

.mono {
  font-family: 'SF Mono', Consolas, monospace;
}

/* 快捷操作 */
.actions-row {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.action-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 14px;
  border-radius: 8px;
  border: 1px solid var(--jm-accent-2);
  background: rgba(var(--jm-accent-1-rgb), 0.3);
  color: var(--jm-accent-6);
  cursor: pointer;
  font-size: 12px;
  transition: all 0.15s;
}

.action-btn:hover {
  border-color: var(--jm-accent-3);
  background: rgba(var(--jm-accent-1-rgb), 0.6);
  color: var(--jm-accent-7);
}

.action-btn.primary {
  background: rgba(var(--jm-primary-1-rgb), 0.1);
  border-color: rgba(var(--jm-primary-1-rgb), 0.2);
  color: var(--jm-primary-2);
}

.action-btn.primary:hover {
  background: rgba(var(--jm-primary-1-rgb), 0.18);
  border-color: rgba(var(--jm-primary-1-rgb), 0.3);
}
</style>
