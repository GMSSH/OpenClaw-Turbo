<template>
  <n-config-provider :theme="darkTheme" :theme-overrides="theme">
    <div class="app-layout">
      <!-- 左侧边栏 -->
      <aside class="sidebar">

        <nav class="sidebar-nav">
          <button
            v-for="item in menuItems"
            :key="item.key"
            class="nav-item"
            :class="{
              active: activeMenu === item.key,
              disabled: !deployed && item.key !== 'console',
            }"
            @click="onMenuClick(item)"
          >
            <span class="nav-icon" v-html="item.icon"></span>
            <span class="nav-label">{{ item.label }}</span>
            <svg v-if="!deployed && item.key !== 'console'" class="lock-icon" viewBox="0 0 24 24" width="12" height="12" fill="none">
              <rect x="3" y="11" width="18" height="11" rx="2" stroke="currentColor" stroke-width="1.5"/>
              <path d="M7 11V7a5 5 0 0110 0v4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
            </svg>
          </button>
        </nav>

      </aside>

      <!-- 右侧内容区 -->
      <main class="main-content">
        <router-view v-slot="{ Component }">
          <transition name="page-fade" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </main>
    </div>
  </n-config-provider>
</template>

<script setup>
import { ref, provide, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { NConfigProvider, darkTheme } from 'naive-ui'
import { theme } from '@/theme/dark'
import { getClawStatus } from '@/api/deploy'

const router = useRouter()
const route = useRoute()
const deployed = ref(false)
const clawRunning = ref(false)
const statusLoading = ref(true)
const activeMenu = ref('console')

const menuItems = [
  {
    key: 'console',
    label: '控制台',
    route: '/console',
    icon: '<svg viewBox="0 0 24 24" width="18" height="18" fill="none"><rect x="3" y="3" width="7" height="7" rx="1.5" stroke="currentColor" stroke-width="1.5"/><rect x="14" y="3" width="7" height="7" rx="1.5" stroke="currentColor" stroke-width="1.5"/><rect x="3" y="14" width="7" height="7" rx="1.5" stroke="currentColor" stroke-width="1.5"/><rect x="14" y="14" width="7" height="7" rx="1.5" stroke="currentColor" stroke-width="1.5"/></svg>',
  },
  {
    key: 'chat',
    label: '三方平台接入',
    route: '/chat',
    icon: '<svg viewBox="0 0 24 24" width="18" height="18" fill="none"><path d="M21 15a2 2 0 01-2 2H7l-4 4V5a2 2 0 012-2h14a2 2 0 012 2z" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>',
  },
  {
    key: 'agents',
    label: 'Agent 管理',
    route: '/agents',
    icon: '<svg viewBox="0 0 24 24" width="18" height="18" fill="none"><circle cx="12" cy="8" r="4" stroke="currentColor" stroke-width="1.5"/><path d="M6 21v-1a6 6 0 0112 0v1" stroke="currentColor" stroke-width="1.5"/><path d="M16 3l2-2M8 3L6 1" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/></svg>',
  },
  {
    key: 'abilities',
    label: '能力',
    route: '/abilities',
    icon: '<svg viewBox="0 0 24 24" width="18" height="18" fill="none"><polygon points="13,2 3,14 12,14 11,22 21,10 12,10" stroke="currentColor" stroke-width="1.5" stroke-linejoin="round"/></svg>',
  },
  {
    key: 'cron',
    label: '定时任务',
    route: '/cron',
    icon: '<svg viewBox="0 0 24 24" width="18" height="18" fill="none"><circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="1.5"/><polyline points="12,6 12,12 16,14" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>',
  },
  {
    key: 'cyber-worker',
    label: '赛博员工',
    route: '/cyber-worker',
    icon: '<svg viewBox="0 0 24 24" width="18" height="18" fill="none"><rect x="4" y="4" width="16" height="12" rx="2" stroke="currentColor" stroke-width="1.5"/><path d="M9 9h.01M15 9h.01" stroke="currentColor" stroke-width="2" stroke-linecap="round"/><path d="M9 13h6" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/><path d="M8 20h8M12 16v4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/></svg>',
  },
]

function onMenuClick(item) {
  if (!deployed.value && item.key !== 'console') return
  activeMenu.value = item.key
  router.push(item.route)
}

function setDeployed(val) {
  deployed.value = val
}

// 提供给子组件
provide('deployed', deployed)
provide('setDeployed', setDeployed)

// 同步 activeMenu 与路由
function syncMenu() {
  const path = route.path
  const found = menuItems.find(m => path.startsWith(m.route))
  if (found) activeMenu.value = found.key
  else activeMenu.value = 'console'
}

onMounted(async () => {
  syncMenu()
  try {
    const status = await getClawStatus()
    if (status && (status.running || status.status !== 'stopped' || status.webPort > 0)) {
      deployed.value = true
      clawRunning.value = !!status.running
      if (['/', '/env-check', '/setup', '/progress'].includes(route.path)) {
        router.replace('/console')
      }
    } else {
      if (!['/', '/env-check', '/setup', '/progress'].includes(route.path)) {
        router.replace('/console')
      }
    }
  } catch {
    // 检测失败，默认未部署
  } finally {
    statusLoading.value = false
  }
})
</script>

<style scoped>
.app-layout {
  display: flex;
  width: 100%;
  height: 100%;
  overflow: hidden;
  background: var(--jm-bg-color);
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
  position: relative;
}

.nav-item:hover:not(.disabled) {
  background: rgba(var(--jm-accent-1-rgb), 0.6);
  color: var(--jm-accent-7);
}

.nav-item.active {
  background: rgba(var(--jm-primary-1-rgb), 0.1);
  color: var(--jm-primary-2);
}

.nav-item.disabled {
  opacity: 0.35;
  cursor: not-allowed;
}

.nav-icon {
  display: flex;
  align-items: center;
  width: 18px;
  height: 18px;
  flex-shrink: 0;
}

.nav-label {
  flex: 1;
}

.lock-icon {
  flex-shrink: 0;
  opacity: 0.6;
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
.status-badge.warning { color: var(--jm-warning-color); }
.status-badge.offline { color: var(--jm-accent-4); }
.status-badge.loading { color: var(--jm-accent-4); }

.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: currentColor;
}
.status-dot.dot-spin {
  animation: pulse 1.2s ease-in-out infinite;
}
@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.3; }
}

.status-badge.online .status-dot {
  box-shadow: 0 0 6px currentColor;
}

/* ========== 内容区 ========== */
.main-content {
  flex: 1;
  overflow: hidden;
  position: relative;
  min-width: 0;
}

.page-fade-enter-active,
.page-fade-leave-active {
  transition: opacity 0.2s ease;
}
.page-fade-enter-from,
.page-fade-leave-to {
  opacity: 0;
}
</style>

<!-- 全局样式：NConfigProvider 内部 div 需要撑满高度 -->
<style>
#app {
  position: absolute;
  inset: 0;
  overflow: hidden;
}
#app > .n-config-provider {
  width: 100%;
  height: 100%;
}
</style>
