<template>
  <div class="console-page">
    <!-- 未部署：显示部署流程 -->
    <template v-if="!deployed">
      <!-- 步骤1：选择部署方式 -->
      <div v-if="step === 'mode'" class="mode-select-wrap">
        <div class="mode-select-container fade-in-up">
          <div class="mode-header">
            <div class="mode-header-icon">
              <svg viewBox="0 0 24 24" width="22" height="22" fill="none">
                <rect x="2" y="4" width="20" height="16" rx="2.5" stroke="var(--jm-primary-1)" stroke-width="1.5" fill="rgba(var(--jm-primary-1-rgb), 0.08)"/>
                <rect x="5" y="13" width="3" height="4" rx="0.5" fill="var(--jm-primary-2)" opacity="0.7"/>
                <rect x="9" y="11" width="3" height="6" rx="0.5" fill="var(--jm-primary-1)"/>
                <rect x="13" y="9" width="3" height="8" rx="0.5" fill="var(--jm-primary-2)" opacity="0.7"/>
                <rect x="17" y="7" width="3" height="10" rx="0.5" fill="var(--jm-primary-1)"/>
              </svg>
            </div>
            <div>
              <h2>选择部署方式</h2>
              <p class="mode-desc">根据服务器环境选择最适合的部署方式</p>
            </div>
          </div>

          <div class="mode-cards">
            <div class="mode-card" :class="{ active: selectedMode === 'docker' }" @click="selectedMode = 'docker'">
              <div class="mode-card-icon docker">
                <svg viewBox="0 0 24 24" width="28" height="28" fill="none">
                  <rect x="2" y="8" width="4" height="3" rx="0.5" stroke="currentColor" stroke-width="1.2"/>
                  <rect x="7" y="8" width="4" height="3" rx="0.5" stroke="currentColor" stroke-width="1.2"/>
                  <rect x="12" y="8" width="4" height="3" rx="0.5" stroke="currentColor" stroke-width="1.2"/>
                  <rect x="7" y="4" width="4" height="3" rx="0.5" stroke="currentColor" stroke-width="1.2"/>
                  <rect x="12" y="4" width="4" height="3" rx="0.5" stroke="currentColor" stroke-width="1.2"/>
                  <path d="M1 13c1.5 3 5 5 11 5s8-1 10-3" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
                </svg>
              </div>
              <div class="mode-card-info">
                <span class="mode-card-title">Docker 部署 <span class="mode-tag recommend">推荐</span></span>
                <span class="mode-card-desc">一键拉取镜像，容器化运行，开箱即用</span>
                <div class="mode-features">
                  <span class="feat-item good">✓ 简单快捷</span>
                  <span class="feat-item good">✓ 无环境依赖</span>
                  <span class="feat-item good">✓ 支持所有 Linux</span>
                </div>
              </div>
              <div class="mode-check" v-if="selectedMode === 'docker'">
                <svg viewBox="0 0 24 24" width="18" height="18"><circle cx="12" cy="12" r="11" fill="var(--jm-primary-1)"/><path d="M8 12.5l2.5 2.5 5.5-5.5" stroke="#fff" stroke-width="2" fill="none" stroke-linecap="round" stroke-linejoin="round"/></svg>
              </div>
            </div>

            <div class="mode-card" :class="{ active: selectedMode === 'local' }" @click="selectedMode = 'local'">
              <div class="mode-card-icon local">
                <svg viewBox="0 0 24 24" width="28" height="28" fill="none">
                  <rect x="3" y="3" width="18" height="18" rx="2" stroke="currentColor" stroke-width="1.5"/>
                  <path d="M8 8l4 4-4 4M14 16h4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
                </svg>
              </div>
              <div class="mode-card-info">
                <span class="mode-card-title">本地 Shell 部署</span>
                <span class="mode-card-desc">源码编译安装，直接运行于宿主机</span>
                <div class="mode-features">
                  <span class="feat-item good">✓ 可管理服务器进程</span>
                  <span class="feat-item">核心建议 2C4G 以上</span>
                  <span class="feat-item">自动装 Node+pnpm</span>
                  <span class="feat-item">CentOS7+ / Debian9+</span>
                </div>
              </div>
              <div class="mode-check" v-if="selectedMode === 'local'">
                <svg viewBox="0 0 24 24" width="18" height="18"><circle cx="12" cy="12" r="11" fill="var(--jm-primary-1)"/><path d="M8 12.5l2.5 2.5 5.5-5.5" stroke="#fff" stroke-width="2" fill="none" stroke-linecap="round" stroke-linejoin="round"/></svg>
              </div>
            </div>
          </div>

          <n-button type="primary" class="mode-go-btn" size="large" :disabled="!selectedMode" @click="step = 'check'">
            <template #icon>
              <svg viewBox="0 0 24 24" width="16" height="16" fill="none">
                <path d="M13 5l7 7-7 7M5 12h14" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
              </svg>
            </template>
            继续
          </n-button>
        </div>
      </div>

      <!-- 步骤2：环境检测 -->
      <EnvironmentCheck v-else-if="step === 'check'" :mode="selectedMode" @passed="step = 'setup'" />

      <!-- 步骤3：配置部署 -->
      <DeploySetupInline v-else-if="step === 'setup'" :deploy-mode="selectedMode" />
    </template>

    <!-- 已部署：显示仪表盘 -->
    <template v-else>
      <DashboardPanel />
    </template>
  </div>
</template>

<script setup>
import { ref, inject, onMounted, provide } from 'vue'
import { useRoute } from 'vue-router'
import { NButton } from 'naive-ui'
import EnvironmentCheck from '@/components/EnvironmentCheck.vue'
import DeploySetupInline from '@/components/DeploySetupInline.vue'
import DashboardPanel from '@/components/DashboardPanel.vue'

const deployed = inject('deployed')
const route = useRoute()
const selectedMode = ref('docker')
const step = ref(route.query.step === 'setup' ? 'setup' : 'mode')

provide('deployMode', selectedMode)
</script>

<style scoped>
.console-page {
  width: 100%;
  height: 100%;
  overflow-y: auto;
}

/* 部署方式选择 */
.mode-select-wrap {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%; height: 100%;
  padding: 24px;
}
.mode-select-container { width: 100%; max-width: 480px; }
.mode-header {
  display: flex; align-items: center; gap: 12px; margin-bottom: 24px;
}
.mode-header-icon {
  display: flex; align-items: center; justify-content: center;
  width: 42px; height: 42px; border-radius: 10px;
  background: rgba(var(--jm-primary-1-rgb), 0.06);
  border: 1px solid rgba(var(--jm-primary-1-rgb), 0.1);
  flex-shrink: 0;
}
.mode-header h2 { font-size: 16px; font-weight: 600; color: var(--jm-accent-7); margin: 0 0 2px; }
.mode-desc { font-size: 12px; color: var(--jm-accent-4); margin: 0; }

.mode-cards { display: flex; flex-direction: column; gap: 8px; margin-bottom: 20px; }

.mode-card {
  display: flex; align-items: center; gap: 14px;
  padding: 16px 18px; border-radius: 12px;
  background: rgba(var(--jm-accent-1-rgb), 0.4);
  border: 2px solid var(--jm-accent-2);
  cursor: pointer; transition: all 0.2s;
}
.mode-card:hover { border-color: var(--jm-accent-3); }
.mode-card.active {
  border-color: var(--jm-primary-1);
  background: rgba(var(--jm-primary-1-rgb), 0.06);
}

.mode-card-icon {
  width: 48px; height: 48px; border-radius: 10px;
  display: flex; align-items: center; justify-content: center;
  flex-shrink: 0;
}
.mode-card-icon.docker {
  background: rgba(33, 150, 243, 0.08); color: #2196f3;
}
.mode-card-icon.local {
  background: rgba(76, 175, 80, 0.08); color: #4caf50;
}
.mode-card-info { flex: 1; display: flex; flex-direction: column; gap: 4px; }
.mode-card-title { font-size: 14px; font-weight: 600; color: var(--jm-accent-7); display: flex; align-items: center; gap: 6px; }
.mode-card-desc { font-size: 11px; color: var(--jm-accent-4); line-height: 1.6; }

.mode-tag {
  font-size: 10px; font-weight: 500; padding: 1px 6px;
  border-radius: 3px;
}
.mode-tag.recommend {
  background: rgba(var(--jm-primary-1-rgb), 0.12);
  color: var(--jm-primary-1);
}

.mode-features {
  display: flex; flex-wrap: wrap; gap: 4px 8px; margin-top: 4px;
}
.feat-item {
  font-size: 10px; color: var(--jm-accent-4);
  padding: 1px 6px; border-radius: 3px;
  background: rgba(var(--jm-accent-1-rgb), 0.6);
}
.feat-item.good { color: var(--jm-success-color); }
.mode-check { flex-shrink: 0; }

.mode-go-btn {
  width: 100%; height: 40px; font-size: 14px; font-weight: 500;
  border-radius: 10px; transition: transform 0.15s, box-shadow 0.15s;
}
.mode-go-btn:hover { transform: translateY(-1px); box-shadow: 0 4px 16px rgba(var(--jm-primary-1-rgb), 0.25); }

.fade-in-up { animation: fadeInUp 0.3s ease-out both; }
@keyframes fadeInUp { from { opacity: 0; transform: translateY(8px); } to { opacity: 1; transform: translateY(0); } }
</style>
