<template>
  <div class="dashboard-panel">
    <!-- 顶部状态概览 -->
    <div class="section-row">
      <div class="status-card" :class="!statusLoaded ? 'card-loading' : (running ? 'card-ok' : 'card-off')">
        <div class="status-header">
          <div class="status-indicator" :class="!statusLoaded ? 'ind-loading' : (running ? 'ind-ok' : 'ind-off')">
            <div class="pulse-ring" v-if="running"></div>
          </div>
          <div class="status-text">
            <span class="status-label">{{ !statusLoaded ? '获取中...' : (running ? '运行中' : '已停止') }}</span>
            <span class="status-name">{{ statusData.containerName }}</span>
          </div>
          <div class="status-actions">
            <n-tooltip trigger="hover">
              <template #trigger>
                <button class="ctrl-btn" @click="handleRestart" :disabled="actionLoading">
                  <svg viewBox="0 0 24 24" width="14" height="14" fill="none">
                    <path d="M1 4v6h6M23 20v-6h-6" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
                    <path d="M20.49 9A9 9 0 005.64 5.64L1 10m22 4l-4.64 4.36A9 9 0 013.51 15" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
                  </svg>
                </button>
              </template>
              重启
            </n-tooltip>
            <n-tooltip trigger="hover">
              <template #trigger>
                <button class="ctrl-btn" @click="handleStop" :disabled="actionLoading">
                  <svg viewBox="0 0 24 24" width="14" height="14" fill="none">
                    <rect x="6" y="6" width="12" height="12" rx="2" stroke="currentColor" stroke-width="1.5"/>
                  </svg>
                </button>
              </template>
              停止
            </n-tooltip>
            <n-tooltip trigger="hover">
              <template #trigger>
                <button class="ctrl-btn ctrl-danger" @click="handleUninstall" :disabled="actionLoading">
                  <svg viewBox="0 0 24 24" width="14" height="14" fill="none">
                    <polyline points="3,6 5,6 21,6" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
                    <path d="M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a2 2 0 012-2h4a2 2 0 012 2v2" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
                  </svg>
                </button>
              </template>
              卸载 (移除部署和配置数据)
            </n-tooltip>
          </div>
        </div>
        <div class="status-meta">
          <div class="meta-info">
            <span>{{ statusData.image }}</span>
            <span class="divider">|</span>
            <span>Uptime: {{ statusData.uptime || '-' }}</span>
          </div>
          <!-- 快捷操作按钮 -->
          <div class="status-tools">
            <n-dropdown trigger="click" :options="webUIOptions" @select="onWebUISelect" :theme-overrides="{ color: 'var(--jm-bg-2)', optionColorHover: 'rgba(var(--jm-primary-1-rgb), 0.1)', borderRadius: '8px', boxShadow: '0 4px 16px rgba(0,0,0,0.3)', optionTextColor: 'var(--jm-accent-6)', padding: '4px', dividerColor: 'var(--jm-accent-2)' }">
              <button class="act-btn primary">
                <svg viewBox="0 0 24 24" width="13" height="13" fill="none">
                  <path d="M18 13v6a2 2 0 01-2 2H5a2 2 0 01-2-2V8a2 2 0 012-2h6" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
                  <polyline points="15,3 21,3 21,9" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
                  <line x1="10" y1="14" x2="21" y2="3" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
                </svg>
                打开 Web UI
                <svg viewBox="0 0 24 24" width="10" height="10" fill="none" style="margin-left: 2px;">
                  <polyline points="6,9 12,15 18,9" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                </svg>
              </button>
            </n-dropdown>
            <button class="act-btn" @click="viewLogs">
              <svg viewBox="0 0 24 24" width="13" height="13" fill="none">
                <rect x="3" y="3" width="18" height="18" rx="2" stroke="currentColor" stroke-width="1.5"/>
                <path d="M8 8l4 4-4 4M14 16h4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
              </svg>
              日志
            </button>
            <button class="act-btn" @click="openConfig">
              <svg viewBox="0 0 24 24" width="13" height="13" fill="none">
                <path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
                <polyline points="14,2 14,8 20,8" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
              </svg>
              配置
            </button>
            <button class="act-btn" @click="refreshStatus">
              <svg viewBox="0 0 24 24" width="13" height="13" fill="none">
                <path d="M1 4v6h6M23 20v-6h-6" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
                <path d="M20.49 9A9 9 0 005.64 5.64L1 10m22 4l-4.64 4.36A9 9 0 013.51 15" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
              </svg>
              刷新
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 数据指标行 -->
    <div class="metrics-row">
      <div class="metric-card">
        <div class="metric-icon skill">
          <svg viewBox="0 0 24 24" width="16" height="16" fill="none">
            <path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
        </div>
        <div class="metric-data">
          <span class="metric-value">{{ !statusLoaded ? '-' : !skillLoaded ? '...' : skillCount }}</span>
          <span class="metric-label">能力数</span>
        </div>
      </div>
      <div class="metric-card">
        <div class="metric-icon cron">
          <svg viewBox="0 0 24 24" width="16" height="16" fill="none">
            <circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="1.5"/>
            <path d="M12 6v6l4 2" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
        </div>
        <div class="metric-data">
          <span class="metric-value">{{ !statusLoaded ? '-' : !jobLoaded ? '...' : jobCount }}</span>
          <span class="metric-label">定时任务</span>
        </div>
      </div>
      <div class="metric-card">
        <div class="metric-icon port">
          <svg viewBox="0 0 24 24" width="16" height="16" fill="none">
            <circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="1.5"/>
            <path d="M2 12h20" stroke="currentColor" stroke-width="1.5"/>
            <path d="M12 2a15.3 15.3 0 014 10 15.3 15.3 0 01-4 10 15.3 15.3 0 01-4-10 15.3 15.3 0 014-10z" stroke="currentColor" stroke-width="1.5"/>
          </svg>
        </div>
        <div class="metric-data">
          <span class="metric-value mono">{{ statusData.webPort || 18789 }}</span>
          <span class="metric-label">Web 端口</span>
        </div>
      </div>
      <div class="metric-card">
        <div class="metric-icon bridge">
          <svg viewBox="0 0 24 24" width="16" height="16" fill="none">
            <path d="M4.5 16.5c-1.5 1.26-2 5-2 5s3.74-.5 5-2c.71-.84.7-2.13-.09-2.91a2.18 2.18 0 00-2.91-.09z" stroke="currentColor" stroke-width="1.5"/>
            <path d="M12 15l-3-3a22 22 0 012-3.95A12.88 12.88 0 0122 2c0 2.72-.78 7.5-6 11a22.35 22.35 0 01-4 2z" stroke="currentColor" stroke-width="1.5"/>
          </svg>
        </div>
        <div class="metric-data">
          <span class="metric-value mono">{{ statusData.bridgePort || 18790 }}</span>
          <span class="metric-label">通讯端口</span>
        </div>
      </div>
    </div>

    <!-- 双列内容区：AI 模型 + 网关/地址 -->
    <div class="content-grid">
      <!-- 左列：AI 模型配置 -->
      <div class="model-section">
        <div class="section-title">
          <svg viewBox="0 0 24 24" width="14" height="14" fill="none">
            <path d="M12 2L2 7l10 5 10-5-10-5z" stroke="var(--jm-primary-1)" stroke-width="1.5" stroke-linejoin="round"/>
            <path d="M2 17l10 5 10-5" stroke="var(--jm-primary-2)" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
            <path d="M2 12l10 5 10-5" stroke="var(--jm-primary-1)" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" opacity="0.5"/>
          </svg>
          <span>AI 模型配置</span>
          <button class="switch-model-btn" @click="showModelSwitch = true" :disabled="!running">
            <svg viewBox="0 0 24 24" width="12" height="12" fill="none">
              <path d="M16 3h5v5M4 20L21 3M8 21H3v-5M20 4L3 21" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
            切换
          </button>
        </div>
        <div class="model-card">
          <div class="model-main">
            <div class="model-provider-badge">{{ configData.provider || '-' }}</div>
            <div class="model-name">{{ configData.modelName || configData.primaryModel || '-' }}</div>
          </div>
          <div class="model-details">
            <div class="detail-item">
              <span class="detail-label">API 地址</span>
              <span class="detail-value mono">{{ configData.baseUrl || '-' }}</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">API Key</span>
              <span class="detail-value mono">{{ configData.apiKeyMasked || '-' }}</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">默认模型</span>
              <span class="detail-value">{{ configData.primaryModel || '-' }}</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">上下文窗口</span>
              <span class="detail-value">{{ configData.contextWindow ? formatNum(configData.contextWindow) + ' tokens' : '-' }}</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">最大输出</span>
              <span class="detail-value">{{ configData.maxTokens ? formatNum(configData.maxTokens) + ' tokens' : '-' }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- 右列：网关信息 + 访问地址 -->
      <div class="right-col">
        <div class="gateway-section">
          <div class="section-title">
            <svg viewBox="0 0 24 24" width="14" height="14" fill="none">
              <rect x="2" y="2" width="20" height="8" rx="2" stroke="var(--jm-primary-1)" stroke-width="1.5"/>
              <rect x="2" y="14" width="20" height="8" rx="2" stroke="var(--jm-primary-2)" stroke-width="1.5"/>
              <circle cx="6" cy="6" r="1.5" fill="var(--jm-primary-1)"/>
              <circle cx="6" cy="18" r="1.5" fill="var(--jm-primary-2)"/>
            </svg>
            <span>网关信息</span>
          </div>
          <div class="gateway-grid">
            <div class="gw-item">
              <span class="gw-label">认证方式</span>
              <span class="gw-value">{{ configData.authMode || '-' }}</span>
            </div>
            <div class="gw-item">
              <span class="gw-label">绑定模式</span>
              <span class="gw-value">{{ configData.gatewayBind || '-' }}</span>
            </div>
            <div class="gw-item">
              <span class="gw-label">网关模式</span>
              <span class="gw-value">{{ configData.gatewayMode || '-' }}</span>
            </div>
            <div class="gw-item">
              <span class="gw-label">服务状态</span>
              <span class="gw-value">{{ statusData.status || 'unknown' }}</span>
            </div>
          </div>
        </div>


      </div>
    </div>



    <!-- 切换模型弹窗 -->
    <n-modal v-model:show="showModelSwitch" preset="card" title="切换 AI 模型" :style="{ width: '480px' }" :mask-closable="false">
      <n-form label-placement="top" :show-feedback="false">
        <n-form-item label="模型提供商">
          <n-select v-model:value="switchForm.provider" :options="providerOptions" @update:value="onProviderChange" placeholder="选择模型提供商" />
        </n-form-item>
        <n-form-item v-if="switchForm.provider === 'custom'" label="API 协议模式">
          <n-select v-model:value="switchForm.apiMode" :options="[
            { label: 'OpenAI Chat Completions', value: 'openai' },
            { label: 'Anthropic Messages', value: 'anthropic' },
          ]" />
        </n-form-item>
        <n-form-item label="模型">
          <n-input v-if="switchForm.provider === 'custom'" v-model:value="switchForm.model" placeholder="输入模型名称" />
          <n-select v-else v-model:value="switchForm.model" :options="switchModelOptions" placeholder="选择模型" />
        </n-form-item>
        <n-form-item label="API Key">
          <n-input v-model:value="switchForm.apiKey" type="password" show-password-on="click" placeholder="输入 API Key" />
        </n-form-item>
        <n-form-item label="API Base URL">
          <n-input v-model:value="switchForm.baseUrl" :placeholder="switchForm.provider === 'custom' ? '例如: https://api.example.com/v1' : '留空使用默认地址'" />
        </n-form-item>
      </n-form>
      <template #footer>
        <div style="display:flex;gap:8px;justify-content:flex-end">
          <n-button @click="showModelSwitch = false" :disabled="switchLoading">取消</n-button>
          <n-button type="primary" @click="doSwitchModel" :loading="switchLoading" :disabled="!switchForm.provider || !switchForm.model || !switchForm.apiKey">确认切换</n-button>
        </div>
      </template>
    </n-modal>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, inject, h } from 'vue'
import { getClawStatus, getClawConfig, stopClaw, restartClaw, uninstallClaw, updateModelConfig, MODEL_PROVIDERS } from '@/api/deploy'
import { getActiveSkillCount } from '@/api/skill'
import { listCronJobs } from '@/api/cron'
import { useRouter } from 'vue-router'
import { NTooltip, NModal, NForm, NFormItem, NSelect, NInput, NButton, NSwitch, NDropdown } from 'naive-ui'
import gm from '@/utils/gmssh'

const router = useRouter()
const setDeployed = inject('setDeployed', () => {})

const statusData = ref({
  running: false, containerName: 'gmssh-openclaw', status: 'unknown',
  webPort: 0, bridgePort: 0, uptime: '-', image: 'gmssh/openclaw:2026.02.17',
})


const configData = ref({
  provider: '', modelName: '', primaryModel: '', baseUrl: '', apiKeyMasked: '',
  contextWindow: 0, maxTokens: 0, gatewayPort: 0, authMode: '', gatewayBind: '',
  gatewayMode: '', containerCPU: '', containerMem: '',
})

const running = computed(() => statusData.value.running)
const statusLoaded = ref(false)

const skillCount = ref(0)
const skillLoaded = ref(false)
const jobCount = ref(0)
const jobLoaded = ref(false)

// ========== Web UI 下拉 ==========
const webUIOptions = [
  {
    label: () => h('div', { style: 'display:flex;align-items:center;gap:8px' }, [
      h('svg', { viewBox: '0 0 16 16', width: 14, height: 14, innerHTML: '<circle cx="8" cy="8" r="7" stroke="currentColor" stroke-width="1.2" fill="none"/><path d="M1 8h14M8 1c2 2 3 4.5 3 7s-1 5-3 7M8 1c-2 2-3 4.5-3 7s1 5 3 7" stroke="currentColor" stroke-width="1.2" fill="none"/>' }),
      '公网访问'
    ]),
    key: 'public'
  },
  {
    label: () => h('div', { style: 'display:flex;align-items:center;gap:8px' }, [
      h('svg', { viewBox: '0 0 16 16', width: 14, height: 14, innerHTML: '<rect x="2" y="4" width="12" height="8" rx="1.5" stroke="currentColor" stroke-width="1.2" fill="none"/><path d="M5 4V3a3 3 0 016 0v1" stroke="currentColor" stroke-width="1.2" fill="none" stroke-linecap="round"/><circle cx="8" cy="8.5" r="1.2" fill="currentColor"/>' }),
      '内网访问'
    ]),
    key: 'private'
  },
]

function getPrivateIp() {
  return window.$gm?.privateIp || 'localhost'
}

function buildWebUrl(ip) {
  const port = statusData.value.webPort
  const token = sessionStorage.getItem('deploy_token') || ''
  if (!port) return ''
  return token ? `http://${ip}:${port}/?token=${token}` : `http://${ip}:${port}`
}

function onWebUISelect(key) {
  const ip = key === 'private' ? getPrivateIp() : gm.getPublicIp()
  const url = buildWebUrl(ip)
  if (url) window.open(url, '_blank')
  else gm.warning('端口信息不可用')
}

function formatNum(n) {
  if (n >= 1000) return (n / 1000).toFixed(0) + 'K'
  return String(n)
}

async function fetchAll() {
  skillLoaded.value = false
  jobLoaded.value = false
  try {
    const [status, config] = await Promise.all([
      getClawStatus(), 
      getClawConfig()
    ])
    statusData.value = { ...statusData.value, ...status }
    configData.value = { ...configData.value, ...config }
    statusLoaded.value = true

    // 异步获取数据指标不阻塞渲染
    getActiveSkillCount().then(res => { skillCount.value = res?.count || 0 }).catch(console.error).finally(() => { skillLoaded.value = true })
    listCronJobs().then(res => { jobCount.value = res?.jobs?.length || 0 }).catch(console.error).finally(() => { jobLoaded.value = true })
  } catch (e) {
    console.error('获取数据失败:', e)
  }
}


function viewLogs() {
  const gmApi = gm.getGmApi()
  const isLocal = configData.value.deployMode === 'local'
  if (isLocal) {
    if (gmApi?.openShell) gmApi.openShell({ arr: ['journalctl -u openclaw -f --no-pager\n'] })
    else gm.info('请在终端执行: journalctl -u openclaw -f')
  } else {
    if (gmApi?.openShell) gmApi.openShell({ arr: ['docker logs -f gmssh-openclaw\n'] })
    else gm.info('请在终端执行: docker logs -f gmssh-openclaw')
  }
}

function refreshStatus() { fetchAll(); gm.success('已刷新') }

function openConfig() {
  const gmApi = gm.getGmApi()
  const configFile = configData.value.configPath || '/opt/gmclaw/conf/openclaw.json'
  if (gmApi?.openCodeEditor) gmApi.openCodeEditor(configFile)
  else gm.info(`配置文件: ${configFile}`)
}

const actionLoading = ref(false)

async function doStop() {
  actionLoading.value = true
  try {
    await stopClaw()
    gm.success('已停止')
    await fetchAll()
  } catch (e) {
    gm.error('停止失败: ' + e.message)
  } finally {
    actionLoading.value = false
  }
}

function handleStop() {
  window.$gm?.dialog?.warning({
    title: '停止 OpenClaw',
    content: '确定要停止 OpenClaw 服务吗？',
    positiveText: '确定停止',
    negativeText: '取消',
    maskClosable: false,
    onPositiveClick: doStop,
  }) || doStop()
}

async function handleRestart() {
  actionLoading.value = true
  try {
    await restartClaw()
    gm.success('已重启')
    await fetchAll()
  } catch (e) {
    gm.error('重启失败: ' + e.message)
  } finally {
    actionLoading.value = false
  }
}

async function doUninstall() {
  actionLoading.value = true
  try {
    await uninstallClaw()
    gm.success('已完全卸载')
    setDeployed(false)
    sessionStorage.removeItem('deploy_token')
    sessionStorage.removeItem('deploy_web_port')
    sessionStorage.removeItem('deploy_setup_form')
    router.replace('/console')
  } catch (e) {
    gm.error('卸载失败: ' + e.message)
  } finally {
    actionLoading.value = false
  }
}

function handleUninstall() {
  window.$gm?.dialog?.warning({
    title: '卸载 OpenClaw',
    content: '将移除所有部署文件和配置数据，此操作不可撤销！',
    positiveText: '确定卸载',
    negativeText: '取消',
    maskClosable: false,
    onPositiveClick: doUninstall,
  }) || doUninstall()
}

// ===== 模型切换 =====
const showModelSwitch = ref(false)
const switchLoading = ref(false)
const switchForm = ref({ provider: null, model: null, apiKey: '', baseUrl: '', apiMode: 'openai' })

const providerOptions = computed(() =>
  MODEL_PROVIDERS.map(p => ({ label: p.displayName, value: p.provider }))
)

const switchModelOptions = computed(() => {
  const p = MODEL_PROVIDERS.find(x => x.provider === switchForm.value.provider)
  return p?.models?.map(m => ({ label: m.name, value: m.id })) || []
})

function onProviderChange(val) {
  switchForm.value.model = null
  const p = MODEL_PROVIDERS.find(x => x.provider === val)
  if (p) {
    switchForm.value.baseUrl = p.baseUrl || ''
    switchForm.value.apiMode = p.apiMode || 'openai'
  }
}

async function doSwitchModel() {
  switchLoading.value = true
  try {
    const p = MODEL_PROVIDERS.find(x => x.provider === switchForm.value.provider)
    const baseUrl = switchForm.value.baseUrl || p?.baseUrl || ''
    const apiMode = switchForm.value.provider === 'custom' ? switchForm.value.apiMode : (p?.apiMode || 'openai')
    await updateModelConfig({
      provider: switchForm.value.provider,
      model: switchForm.value.model,
      apiKey: switchForm.value.apiKey,
      baseUrl,
      apiMode,
    })
    gm.success('模型已切换，服务正在重启...')
    showModelSwitch.value = false
    setTimeout(fetchAll, 3000)
  } catch (e) {
    gm.error('切换失败: ' + e.message)
  } finally {
    switchLoading.value = false
  }
}

onMounted(fetchAll)
</script>

<style scoped>
.dashboard-panel {
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 16px;
  overflow-y: auto;
  height: 100%;
}

/* ========== 状态概览 ========== */
.status-card {
  padding: 18px 20px;
  border-radius: 12px;
  border: 1px solid var(--jm-accent-2);
  background: rgba(var(--jm-accent-1-rgb), 0.4);
  transition: border-color 0.2s, box-shadow 0.2s;
}
.status-card:hover {
  border-color: var(--jm-accent-3);
  box-shadow: 0 2px 12px rgba(0,0,0,0.08);
}
.status-header { display: flex; align-items: center; gap: 12px; margin-bottom: 10px; }
.status-indicator {
  width: 10px; height: 10px; border-radius: 50%; position: relative; flex-shrink: 0;
}
.ind-ok { background: var(--jm-success-color); }
.ind-off { background: var(--jm-error-color); }
.ind-loading { background: var(--jm-accent-4); animation: pulse-dot 1.2s ease-in-out infinite; }
@keyframes pulse-dot { 0%, 100% { opacity: 1; } 50% { opacity: 0.3; } }
.pulse-ring {
  position: absolute; inset: -4px; border-radius: 50%; border: 2px solid var(--jm-success-color);
  animation: pulse-ring 2s ease-out infinite; opacity: 0;
}
@keyframes pulse-ring {
  0% { transform: scale(0.8); opacity: 0.6; }
  100% { transform: scale(1.6); opacity: 0; }
}
.status-text { display: flex; flex-direction: column; flex: 1; min-width: 0; }
.status-label { font-size: 15px; font-weight: 600; color: var(--jm-accent-7); }
.status-name { font-size: 11px; color: var(--jm-accent-4); }
.status-meta {
  display: flex; justify-content: space-between; align-items: center; gap: 16px; font-size: 11px; color: var(--jm-accent-4);
  padding-top: 14px; border-top: 1px solid rgba(var(--jm-accent-2-rgb),0.5); margin-top: 4px;
}
.meta-info { display: flex; align-items: center; gap: 8px; }
.divider { color: var(--jm-accent-2); }
.status-tools { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }

/* ========== 指标行 ========== */
.metrics-row { display: grid; grid-template-columns: repeat(4, 1fr); gap: 10px; }
.metric-card {
  display: flex; align-items: center; gap: 12px;
  padding: 14px 16px; border-radius: 10px;
  background: rgba(var(--jm-accent-1-rgb), 0.4);
  border: 1px solid rgba(var(--jm-accent-2-rgb),0.5);
  transition: border-color 0.2s, transform 0.15s, box-shadow 0.2s;
}
.metric-card:hover {
  border-color: var(--jm-accent-3);
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(0,0,0,0.06);
}
.metric-icon {
  width: 36px; height: 36px; border-radius: 10px;
  display: flex; align-items: center; justify-content: center; flex-shrink: 0;
}
.metric-icon.skill { background: rgba(162,120,255,0.1); color: #a278ff; }
.metric-icon.cron { background: rgba(99,180,255,0.1); color: #63b4ff; }
.metric-icon.port { background: rgba(80,200,120,0.1); color: #50c878; }
.metric-icon.bridge { background: rgba(255,165,79,0.1); color: #ffa54f; }
.metric-data { display: flex; flex-direction: column; min-width: 0; }
.metric-value {
  font-size: 14px; font-weight: 600; color: var(--jm-accent-7);
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}
.metric-label { font-size: 11px; color: var(--jm-accent-4); margin-top: 1px; }

/* ========== 双列内容区 ========== */
.content-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
  align-items: start;
}

.right-col {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

/* ========== 模型配置 ========== */
.section-title {
  display: flex; align-items: center; gap: 6px;
  font-size: 12px; font-weight: 500; color: var(--jm-accent-5); margin-bottom: 8px;
}
.switch-model-btn {
  margin-left: auto;
  display: flex; align-items: center; gap: 3px;
  padding: 2px 8px;
  font-size: 11px;
  color: var(--jm-primary-1);
  border: 1px solid var(--jm-primary-1);
  border-radius: 4px;
  background: transparent;
  cursor: pointer;
  transition: all 0.15s;
}
.switch-model-btn:hover:not(:disabled) {
  background: var(--jm-primary-1);
  color: #fff;
}
.switch-model-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}
.model-card {
  border: 1px solid var(--jm-accent-2); border-radius: 10px;
  background: rgba(var(--jm-accent-1-rgb), 0.3); overflow: hidden;
  transition: border-color 0.2s, box-shadow 0.2s;
}
.model-card:hover {
  border-color: var(--jm-accent-3);
  box-shadow: 0 2px 12px rgba(0,0,0,0.06);
}
.model-main {
  display: flex; align-items: center; gap: 10px;
  padding: 14px 16px; border-bottom: 1px solid rgba(var(--jm-accent-2-rgb),0.5);
}
.model-provider-badge {
  padding: 3px 10px; border-radius: 4px; font-size: 11px; font-weight: 600;
  background: rgba(var(--jm-primary-1-rgb), 0.12); color: var(--jm-primary-2);
  text-transform: capitalize;
}
.model-name { font-size: 14px; font-weight: 500; color: var(--jm-accent-7); }
.model-details { padding: 12px 16px; display: flex; flex-direction: column; gap: 8px; }
.detail-item { display: flex; justify-content: space-between; align-items: center; }
.detail-label { font-size: 11px; color: var(--jm-accent-4); flex-shrink: 0; }
.detail-value {
  font-size: 12px; color: var(--jm-accent-6);
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap; text-align: right; margin-left: 12px;
}

/* ========== 网关信息 ========== */
.gateway-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 8px; }
.gw-item {
  display: flex; justify-content: space-between; padding: 10px 14px;
  border-radius: 8px; background: rgba(var(--jm-accent-1-rgb), 0.3);
  border: 1px solid rgba(var(--jm-accent-2-rgb),0.4);
  transition: border-color 0.2s;
}
.gw-item:hover { border-color: var(--jm-accent-3); }
.gw-label { font-size: 11px; color: var(--jm-accent-4); }
.gw-value { font-size: 12px; color: var(--jm-accent-6); font-weight: 500; }



/* ========== 状态栏操作按钮 / 快捷操作 ========== */
.status-actions { display: flex; gap: 4px; margin-left: auto; }
.ctrl-btn {
  display: flex; align-items: center; justify-content: center;
  width: 28px; height: 28px; border-radius: 6px;
  border: 1px solid var(--jm-accent-2); background: transparent;
  color: var(--jm-accent-5); cursor: pointer; transition: all 0.15s;
}
.ctrl-btn:hover { color: var(--jm-accent-7); background: rgba(var(--jm-accent-1-rgb), 0.6); border-color: var(--jm-accent-3); }
.ctrl-btn.ctrl-danger:hover { color: var(--jm-error-color, #dc2626); border-color: rgba(220,38,38,0.3); background: rgba(220,38,38,0.06); }
.ctrl-btn:disabled { opacity: 0.4; cursor: not-allowed; pointer-events: none; }

.act-btn {
  display: flex; align-items: center; gap: 4px; padding: 4px 10px;
  border-radius: 6px; border: 1px solid var(--jm-accent-2);
  background: rgba(var(--jm-bg-1-rgb), 0.3); color: var(--jm-accent-6);
  cursor: pointer; font-size: 11px; transition: all 0.15s;
}
.act-btn:hover {
  border-color: var(--jm-accent-3); background: rgba(var(--jm-bg-2-rgb), 0.5);
  color: var(--jm-accent-7); transform: translateY(-1px);
}
.act-btn.primary {
  background: rgba(var(--jm-primary-1-rgb), 0.1);
  border-color: rgba(var(--jm-primary-1-rgb), 0.2); color: var(--jm-primary-2);
}
.act-btn.primary:hover { background: rgba(var(--jm-primary-1-rgb), 0.18); }

.mono { font-family: 'SF Mono', Consolas, monospace; }

/* ========== 响应式：小屏幕回退单列 ========== */
@media (max-width: 680px) {
  .content-grid { grid-template-columns: 1fr; }
  .metrics-row { grid-template-columns: repeat(2, 1fr); }
}
</style>
