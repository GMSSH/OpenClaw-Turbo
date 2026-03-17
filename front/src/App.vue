<template>
  <n-config-provider :theme="naiveTheme" :theme-overrides="themeOverrides">
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
            :title="item.label"
          >
            <span class="nav-icon" v-html="item.icon"></span>
            <span class="nav-label">{{ item.label }}</span>
            <svg v-if="!deployed && item.key !== 'console'" class="lock-icon" viewBox="0 0 24 24" width="10" height="10" fill="none">
              <rect x="3" y="11" width="18" height="11" rx="2" stroke="currentColor" stroke-width="1.5"/>
              <path d="M7 11V7a5 5 0 0110 0v4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
            </svg>
          </button>
        </nav>

        <div class="sidebar-bottom">
          <button class="theme-toggle" @click="toggleTheme" :title="isDark ? t('nav.switchToLight') : t('nav.switchToDark')">
            <div class="toggle-track" :class="{ light: !isDark }">
              <div class="toggle-thumb">
                <svg v-if="isDark" viewBox="0 0 24 24" width="12" height="12" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M21 12.79A9 9 0 1111.21 3 7 7 0 0021 12.79z"/>
                </svg>
                <svg v-else viewBox="0 0 24 24" width="12" height="12" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <circle cx="12" cy="12" r="5"/>
                  <line x1="12" y1="1" x2="12" y2="3"/><line x1="12" y1="21" x2="12" y2="23"/>
                  <line x1="4.22" y1="4.22" x2="5.64" y2="5.64"/><line x1="18.36" y1="18.36" x2="19.78" y2="19.78"/>
                  <line x1="1" y1="12" x2="3" y2="12"/><line x1="21" y1="12" x2="23" y2="12"/>
                  <line x1="4.22" y1="19.78" x2="5.64" y2="18.36"/><line x1="18.36" y1="5.64" x2="19.78" y2="4.22"/>
                </svg>
              </div>
            </div>
          </button>
        </div>
      </aside>

      <!-- 右侧内容区 -->
      <main class="main-content">
        <router-view v-slot="{ Component, route }">
            <component :is="Component" :key="route.path" />
        </router-view>
      </main>
    </div>
  </n-config-provider>
</template>

<script setup>
import { ref, computed, provide, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { NConfigProvider } from 'naive-ui'
import { useTheme } from '@/composables/useTheme'
import { getClawStatus, isClawInstalled } from '@/api/deploy'

const { t } = useI18n()

const router = useRouter()
const route = useRoute()
const { isDark, naiveTheme, themeOverrides, toggleTheme } = useTheme()
const deployed = ref(false)
const clawRunning = ref(false)
const statusLoading = ref(true)
const activeMenu = ref('console')

provide('isDark', isDark)

const menuItems = computed(() => [
  {
    key: 'console',
    label: t('nav.console'),
    route: '/console',
    icon: '<svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5"><path d="M5 4h4a1 1 0 0 1 1 1v6a1 1 0 0 1-1 1H5a1 1 0 0 1-1-1V5a1 1 0 0 1 1-1m0 12h4a1 1 0 0 1 1 1v2a1 1 0 0 1-1 1H5a1 1 0 0 1-1-1v-2a1 1 0 0 1 1-1m10-4h4a1 1 0 0 1 1 1v6a1 1 0 0 1-1 1h-4a1 1 0 0 1-1-1v-6a1 1 0 0 1 1-1m0-8h4a1 1 0 0 1 1 1v2a1 1 0 0 1-1 1h-4a1 1 0 0 1-1-1V5a1 1 0 0 1 1-1"/></svg>',
  },
  {
    key: 'conversation',
    label: t('nav.conversation'),
    route: '/conversation',
    icon: '<svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5"><path d="M21 15a2 2 0 01-2 2H7l-4 4V5a2 2 0 012-2h14a2 2 0 012 2z"/></svg>',
  },
  {
    key: 'chat',
    label: t('nav.chat'),
    route: '/chat',
    icon: '<svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5"><path d="M4.9 19.1C1 15.2 1 8.8 4.9 4.9M7.8 16.2a5.5 5.5 0 010-8.4"/><circle cx="12" cy="12" r="2"/><path d="M16.2 7.8a5.5 5.5 0 010 8.4M19.1 4.9C23 8.8 23 15.2 19.1 19.1"/></svg>',
  },
  {
    key: 'agents',
    label: t('nav.agents'),
    route: '/agents',
    icon: '<svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5"><path d="M6 6a2 2 0 0 1 2-2h8a2 2 0 0 1 2 2v4a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2zm6-4v2m-3 8v9m6-9v9M5 16l4-2m6 0l4 2M9 18h6M10 8v.01M14 8v.01"/></svg>',
  },
  {
    key: 'abilities',
    label: t('nav.abilities'),
    route: '/abilities',
    icon: '<svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5"><path d="M13 3v7h6l-8 11v-7H5z"/></svg>',
  },
  {
    key: 'models',
    label: t('nav.models'),
    route: '/models',
    icon: '<svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5"><path d="M12 2a7 7 0 017 7c0 2.38-1.19 4.47-3 5.74V17a2 2 0 01-2 2h-4a2 2 0 01-2-2v-2.26C6.19 13.47 5 11.38 5 9a7 7 0 017-7z"/><line x1="9" y1="21" x2="15" y2="21"/></svg>',
  },
  {
    key: 'cron',
    label: t('nav.cron'),
    route: '/cron',
    icon: '<svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5"><path d="M3 12a9 9 0 1 0 18 0a9 9 0 1 0-18 0m9 0h3.5M12 7v5"/></svg>',
  },
])

function onMenuClick(item) {
  if (!deployed.value && item.key !== 'console') return
  activeMenu.value = item.key
  router.push(item.route)
}

function setDeployed(val) {
  deployed.value = val
}

provide('deployed', deployed)
provide('setDeployed', setDeployed)

function syncMenu() {
  const path = route.path
  const found = menuItems.value.find(m => path.startsWith(m.route))
  if (found) activeMenu.value = found.key
  else activeMenu.value = 'console'
}

onMounted(async () => {
  syncMenu()
  try {
    const installResult = await isClawInstalled()
    if (installResult && installResult.installed) {
      deployed.value = true
      try {
        const status = await getClawStatus()
        clawRunning.value = !!(status && status.running)
      } catch {
        clawRunning.value = false
      }
      if (['/', '/env-check', '/setup', '/progress'].includes(route.path)) {
        router.replace('/console')
      }
    } else {
      if (!['/', '/env-check', '/setup', '/progress'].includes(route.path)) {
        router.replace('/console')
      }
    }
  } catch {
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
  background: var(--jm-bg-gradient, var(--jm-bg-color));
}

/* ========== 侧边栏 — 玻璃拟态 ========== */
.sidebar {
  width: 64px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  background: var(--jm-glass-bg);
  backdrop-filter: blur(24px);
  -webkit-backdrop-filter: blur(24px);
  border-right: 1px solid var(--jm-glass-border);
  z-index: 10;
  transition: width 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
  overflow: hidden;
}

.sidebar:hover {
  width: 160px;
}

.sidebar-nav {
  flex: 1;
  padding: 10px 8px;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 14px;
  border-radius: 10px;
  border: none;
  background: transparent;
  color: var(--jm-accent-4);
  cursor: pointer;
  font-size: 12px;
  font-weight: 500;
  transition: all 0.2s;
  text-align: left;
  position: relative;
  white-space: nowrap;
  overflow: hidden;
}

.nav-item:hover:not(.disabled) {
  background: var(--jm-glass-bg-hover);
  color: var(--jm-accent-6);
}

.nav-item.active {
  background: rgba(var(--jm-primary-1-rgb), 0.1);
  color: var(--jm-primary-2);
}

/* 活动态左侧光条 */
.nav-item.active::before {
  content: '';
  position: absolute;
  left: 0;
  top: 25%;
  height: 50%;
  width: 3px;
  border-radius: 0 3px 3px 0;
  background: var(--jm-primary-1);
  box-shadow: 0 0 8px var(--jm-glow-blue);
}

.nav-item.disabled {
  opacity: 0.3;
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
  opacity: 0;
  transition: opacity 0.2s 0.05s;
}

.sidebar:hover .nav-label {
  opacity: 1;
}

.lock-icon {
  flex-shrink: 0;
  opacity: 0.5;
}

/* ========== 3D Theme Toggle ========== */
.sidebar-bottom {
  padding: 8px;
}

.theme-toggle {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  padding: 8px;
  border: none;
  background: transparent;
  cursor: pointer;
}

.toggle-track {
  width: 40px;
  height: 22px;
  border-radius: 11px;
  background: linear-gradient(135deg, rgba(var(--jm-primary-1-rgb), 0.25), rgba(var(--jm-primary-1-rgb), 0.15));
  border: 1px solid rgba(var(--jm-primary-1-rgb), 0.15);
  position: relative;
  transition: all 0.4s cubic-bezier(0.34, 1.56, 0.64, 1);
  box-shadow: inset 0 1px 3px rgba(0, 0, 0, 0.3), 0 1px 2px rgba(0, 0, 0, 0.2);
}

.toggle-track.light {
  background: linear-gradient(135deg, rgba(var(--jm-primary-1-rgb), 0.2), rgba(var(--jm-primary-1-rgb), 0.1));
  border-color: rgba(var(--jm-primary-1-rgb), 0.15);
  box-shadow: inset 0 1px 3px rgba(0, 0, 0, 0.06), 0 1px 2px rgba(0, 0, 0, 0.06);
}

.toggle-thumb {
  position: absolute;
  top: 2px;
  left: 2px;
  width: 16px;
  height: 16px;
  border-radius: 50%;
  background: linear-gradient(135deg, #e8e8f0, #fff);
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.4s cubic-bezier(0.34, 1.56, 0.64, 1);
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.3);
}

.toggle-track.light .toggle-thumb {
  left: calc(100% - 18px);
  background: linear-gradient(135deg, #fff, #f0f0ff);
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.15);
}

.toggle-thumb svg {
  color: #444;
}

.toggle-track.light .toggle-thumb svg {
  color: var(--jm-primary-1);
}

/* ========== 内容区 ========== */
.main-content {
  flex: 1;
  overflow: hidden;
  position: relative;
  min-width: 0;
}

/* Page transition */
.page-slide-enter-active {
  transition: all 0.3s cubic-bezier(0.22, 1, 0.36, 1);
}
.page-slide-leave-active {
  transition: all 0.15s ease-in;
}
.page-slide-enter-from {
  opacity: 0;
  transform: translateY(8px);
}
.page-slide-leave-to {
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
