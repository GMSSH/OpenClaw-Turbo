<template>
  <div class="console-page">
    <!-- 未部署：显示部署流程 -->
    <template v-if="!deployed">
      <!-- 步骤1：选择部署方式 -->
      <div v-if="step === 'mode'" class="mode-select-wrap">
        <div class="bg-grid"></div>
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
                <svg viewBox="0 0 24 24" width="26" height="26" fill="none">
                  <rect x="2" y="8" width="4" height="3" rx="0.5" stroke="currentColor" stroke-width="1.2"/>
                  <rect x="7" y="8" width="4" height="3" rx="0.5" stroke="currentColor" stroke-width="1.2"/>
                  <rect x="12" y="8" width="4" height="3" rx="0.5" stroke="currentColor" stroke-width="1.2"/>
                  <rect x="7" y="4" width="4" height="3" rx="0.5" stroke="currentColor" stroke-width="1.2"/>
                  <rect x="12" y="4" width="4" height="3" rx="0.5" stroke="currentColor" stroke-width="1.2"/>
                  <path d="M1 13c1.5 3 5 5 11 5s8-1 10-3" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
                </svg>
              </div>
              <div class="mode-card-info">
                <span class="mode-card-title">Docker 容器部署 <span class="mode-tag recommend">推荐</span></span>
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
                <svg viewBox="0 0 24 24" width="26" height="26" fill="none">
                  <rect x="3" y="3" width="18" height="18" rx="2" stroke="currentColor" stroke-width="1.5"/>
                  <path d="M8 8l4 4-4 4M14 16h4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
                </svg>
              </div>
              <div class="mode-card-info">
                <span class="mode-card-title">部署到本机</span>
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

          <button class="cta-btn" :disabled="!selectedMode" @click="step = 'check'">
            <span class="cta-text">继续</span>
            <svg viewBox="0 0 24 24" width="16" height="16" fill="none">
              <path d="M13 5l7 7-7 7M5 12h14" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
          </button>

          <!-- 引流卡片 -->
          <div class="promo-card">
            <div class="promo-left">
              <div class="promo-brand">
                <span class="promo-lobster">🦞</span>
                <div>
                  <div class="promo-title">一键部署 OpenClaw</div>
                  <div class="promo-sub">GMSSH × OpenClaw</div>
                </div>
              </div>
              <p class="promo-desc">一键部署你的7×24小时专属AI助理，预置50+热门Skills，支持一键接入飞书等即时通讯工具。全程可视化，无需命令行。</p>
            </div>
            <div class="promo-right">
              <img src="/qrcode.png" alt="扫码加入交流群" class="promo-qr" />
              <span class="promo-qr-label">扫码加入交流群</span>
            </div>
          </div>
        </div>
      </div>

      <!-- 步骤2：环境检测 -->
      <EnvironmentCheck v-else-if="step === 'check'" :mode="selectedMode" @passed="step = 'setup'" @back="step = 'mode'" />

      <!-- 步骤3：配置部署 -->
      <DeploySetupInline v-else-if="step === 'setup'" :deploy-mode="selectedMode" @back="step = 'check'" />
    </template>

    <!-- 已部署：显示仪表盘 -->
    <template v-else>
      <DashboardPanel />
    </template>
  </div>
</template>

<script setup>
import { ref, inject, onMounted, provide, watch } from 'vue'
import { useRoute } from 'vue-router'
import EnvironmentCheck from '@/components/EnvironmentCheck.vue'
import DeploySetupInline from '@/components/DeploySetupInline.vue'
import DashboardPanel from '@/components/DashboardPanel.vue'

const deployed = inject('deployed')
const route = useRoute()
const DEPLOY_MODE_KEY = 'deploy_selected_mode'
const savedMode = sessionStorage.getItem(DEPLOY_MODE_KEY)
const selectedMode = ref(
  // 优先从 sessionStorage 恢复，避免跳过 mode 选择步骤时默认回 docker
  (savedMode === 'local' || savedMode === 'docker') ? savedMode : 'docker'
)
const step = ref(route.query.step === 'setup' ? 'setup' : 'mode')

// 每次 selectedMode 变化时持久化
watch(selectedMode, (val) => {
  sessionStorage.setItem(DEPLOY_MODE_KEY, val)
})

provide('deployMode', selectedMode)
</script>

<style scoped>
.console-page { width: 100%; height: 100%; overflow-y: auto; }

/* ===== 部署方式选择 ===== */
.mode-select-wrap {
  position: relative;
  display: flex; align-items: center; justify-content: center;
  width: 100%; height: 100%; padding: 24px;
  overflow: hidden;
}

/* 背景网格 */
.bg-grid {
  position: absolute; inset: 0; z-index: 0;
  background-image:
    linear-gradient(rgba(var(--jm-accent-2-rgb), 0.15) 1px, transparent 1px),
    linear-gradient(90deg, rgba(var(--jm-accent-2-rgb), 0.15) 1px, transparent 1px);
  background-size: 40px 40px;
  mask-image: radial-gradient(ellipse 60% 50% at 50% 50%, black 40%, transparent 100%);
  -webkit-mask-image: radial-gradient(ellipse 60% 50% at 50% 50%, black 40%, transparent 100%);
}

.mode-select-container {
  position: relative; z-index: 1;
  width: 100%; max-width: 480px;
  background: var(--jm-glass-bg);
  backdrop-filter: blur(24px); -webkit-backdrop-filter: blur(24px);
  border: 1px solid var(--jm-glass-border);
  border-radius: 20px;
  padding: 28px 28px 24px;
  box-shadow:
    var(--jm-glass-inner-glow),
    0 8px 40px rgba(0, 0, 0, 0.12),
    0 0 0 1px rgba(255, 255, 255, 0.04);
}

.mode-header { display: flex; align-items: center; gap: 14px; margin-bottom: 22px; }
.mode-header-icon {
  width: 44px; height: 44px; border-radius: 12px;
  background: linear-gradient(135deg, rgba(var(--jm-primary-1-rgb), 0.12), rgba(var(--jm-primary-1-rgb), 0.04));
  border: 1px solid rgba(var(--jm-primary-1-rgb), 0.15);
  display: flex; align-items: center; justify-content: center;
  box-shadow: 0 0 16px rgba(var(--jm-primary-1-rgb), 0.08);
  flex-shrink: 0;
}
.mode-header h2 { font-size: 16px; font-weight: 600; color: var(--jm-accent-7); margin: 0 0 3px; }
.mode-desc { font-size: 12px; color: var(--jm-accent-4); margin: 0; }

.mode-cards { display: flex; flex-direction: column; gap: 8px; margin-bottom: 22px; }

.mode-card {
  display: flex; align-items: center; gap: 14px;
  padding: 16px 18px; border-radius: 14px;
  background: rgba(var(--jm-accent-1-rgb), 0.25);
  border: 2px solid rgba(var(--jm-accent-2-rgb), 0.12);
  border-top: 2px solid rgba(255, 255, 255, 0.05);
  cursor: pointer;
  transition: all 0.25s;
}
.mode-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(0, 0, 0, 0.06);
  border-color: rgba(var(--jm-accent-3-rgb, 200,200,200), 0.3);
}
.mode-card.active {
  border-color: var(--jm-primary-1);
  background: rgba(var(--jm-primary-1-rgb), 0.04);
  box-shadow: 0 0 0 1px rgba(var(--jm-primary-1-rgb), 0.1), 0 4px 20px rgba(var(--jm-primary-1-rgb), 0.08);
}

.mode-card-icon {
  width: 48px; height: 48px; border-radius: 12px;
  display: flex; align-items: center; justify-content: center;
  flex-shrink: 0; transition: all 0.3s;
}
.mode-card-icon.docker {
  background: linear-gradient(135deg, rgba(33, 150, 243, 0.1), rgba(33, 150, 243, 0.04));
  border: 1px solid rgba(33, 150, 243, 0.12);
  color: #2196f3;
  box-shadow: 0 0 12px rgba(33, 150, 243, 0.06);
}
.mode-card-icon.local {
  background: linear-gradient(135deg, rgba(76, 175, 80, 0.1), rgba(76, 175, 80, 0.04));
  border: 1px solid rgba(76, 175, 80, 0.12);
  color: #4caf50;
  box-shadow: 0 0 12px rgba(76, 175, 80, 0.06);
}
.mode-card.active .mode-card-icon {
  box-shadow: 0 0 16px rgba(var(--jm-primary-1-rgb), 0.12);
}

.mode-card-info { flex: 1; display: flex; flex-direction: column; gap: 4px; }
.mode-card-title { font-size: 14px; font-weight: 600; color: var(--jm-accent-7); display: flex; align-items: center; gap: 6px; }
.mode-card-desc { font-size: 11px; color: var(--jm-accent-4); line-height: 1.6; }

.mode-tag {
  font-size: 9px; font-weight: 600; letter-spacing: 0.03em;
  padding: 1px 6px; border-radius: 4px;
}
.mode-tag.recommend {
  background: rgba(var(--jm-primary-1-rgb), 0.1);
  color: var(--jm-primary-1);
}

.mode-features { display: flex; flex-wrap: wrap; gap: 4px 6px; margin-top: 4px; }
.feat-item {
  font-size: 10px; color: var(--jm-accent-4);
  padding: 2px 7px; border-radius: 4px;
  background: rgba(var(--jm-accent-1-rgb), 0.4);
}
.feat-item.good { color: #22c55e; background: rgba(34, 197, 94, 0.06); }
.mode-check { flex-shrink: 0; }

/* CTA 按钮 */
.cta-btn {
  width: 100%; height: 42px;
  display: flex; align-items: center; justify-content: center; gap: 8px;
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
.cta-btn:disabled { opacity: 0.4; cursor: not-allowed; transform: none; }

.fade-in-up { animation: fadeInUp 0.35s ease-out both; }
@keyframes fadeInUp { from { opacity: 0; transform: translateY(10px); } to { opacity: 1; transform: translateY(0); } }

/* ===== 引流卡片 ===== */
.promo-card {
  margin-top: 16px;
  display: flex; align-items: center; gap: 14px;
  padding: 14px 16px;
  border-radius: 14px;
  background: rgba(var(--jm-primary-1-rgb), 0.06);
  border: 1px solid rgba(var(--jm-primary-1-rgb), 0.14);
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.06);
}

.promo-left { flex: 1; display: flex; flex-direction: column; gap: 6px; min-width: 0; }

.promo-brand { display: flex; align-items: center; gap: 8px; }
.promo-lobster { font-size: 26px; line-height: 1; flex-shrink: 0; }
.promo-title { font-size: 13px; font-weight: 700; color: var(--jm-accent-7); line-height: 1.3; }
.promo-sub { font-size: 10px; color: var(--jm-accent-4); margin-top: 1px; }

.promo-desc {
  font-size: 10.5px; color: var(--jm-accent-5);
  line-height: 1.65; margin: 0;
  display: -webkit-box; -webkit-line-clamp: 3; -webkit-box-orient: vertical; overflow: hidden;
}

.promo-right {
  display: flex; flex-direction: column; align-items: center; gap: 5px;
  flex-shrink: 0;
}
.promo-qr {
  width: 70px; height: 70px; border-radius: 8px;
  border: 1px solid var(--jm-glass-border);
  object-fit: cover;
}
.promo-qr-label {
  font-size: 9px; color: var(--jm-accent-4);
  white-space: nowrap;
}
</style>
