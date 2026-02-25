<template>
  <div class="deploy-setup-page">
    <div class="setup-container fade-in-up">
      <n-form :model="formData" label-placement="top" class="setup-form" size="medium">
        <!-- 模型提供商 -->
        <n-form-item label="模型提供商" path="provider">
          <n-select
            v-model:value="formData.provider"
            :options="providerOptions"
            placeholder="选择模型提供商"
            @update:value="onProviderChange"
          />
        </n-form-item>

        <!-- 提供商提示 -->
        <div v-if="selectedProviderInfo && !isCustom && !isOllama" class="provider-tip">
          <span class="tip-url">{{ selectedProviderInfo.baseUrl }}</span>
          <a
            v-if="selectedProviderInfo.tokenUrl"
            :href="selectedProviderInfo.tokenUrl"
            target="_blank"
            rel="noopener"
            class="tip-link"
          >获取 Key ↗</a>
        </div>

        <!-- Ollama 端口配置 -->
        <div v-if="isOllama" class="ollama-port-section">
          <n-form-item label="Ollama 端口" path="ollamaPort">
            <n-input-number v-model:value="formData.ollamaPort" :min="1024" :max="65535" placeholder="11434" style="width: 160px" />
            <span class="port-hint">默认 11434，对应 http://localhost:{{ formData.ollamaPort }}/v1</span>
          </n-form-item>
        </div>

        <!-- 自定义：协议模式 + Base URL -->
        <n-form-item v-if="isCustom" label="API 协议模式">
          <n-select
            v-model:value="formData.customApiMode"
            :options="[
              { label: 'OpenAI Chat Completions (大多数平台)', value: 'openai' },
              { label: 'Anthropic Messages (Claude 兼容)', value: 'anthropic' },
            ]"
          />
        </n-form-item>
        <n-form-item v-if="isCustom" label="API Base URL" path="customBaseUrl">
          <n-input
            v-model:value="formData.customBaseUrl"
            :placeholder="formData.customApiMode === 'anthropic' ? '例如: https://api.example.com' : '例如: https://api.example.com/v1'"
          />
        </n-form-item>

        <!-- 模型：预设选择 / 自定义输入 -->
        <n-form-item label="模型" path="model">
          <n-input
            v-if="isCustom"
            v-model:value="formData.model"
            placeholder="输入模型名称，如 gpt-4o、deepseek-chat"
          />
          <n-select
            v-else
            v-model:value="formData.model"
            :options="modelOptions"
            :disabled="!formData.provider"
            placeholder="选择模型"
          />
        </n-form-item>

        <!-- API Key -->
        <n-form-item label="API Key" path="apiKey">
          <n-input
            v-model:value="formData.apiKey"
            :type="showApiKey ? 'text' : 'password'"
            placeholder="输入 API Key"
          >
            <template #suffix>
              <n-button quaternary size="tiny" @click="showApiKey = !showApiKey" style="padding: 0 4px;">
                <svg viewBox="0 0 24 24" width="14" height="14" fill="none">
                  <template v-if="showApiKey">
                    <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8S1 12 1 12z" stroke="var(--jm-accent-5)" stroke-width="1.5"/>
                    <circle cx="12" cy="12" r="3" stroke="var(--jm-accent-5)" stroke-width="1.5"/>
                  </template>
                  <template v-else>
                    <path d="M17.94 17.94A10.07 10.07 0 0112 20c-7 0-11-8-11-8a18.45 18.45 0 015.06-5.94M9.9 4.24A9.12 9.12 0 0112 4c7 0 11 8 11 8a18.5 18.5 0 01-2.16 3.19" stroke="var(--jm-accent-5)" stroke-width="1.5" stroke-linecap="round"/>
                    <line x1="1" y1="1" x2="23" y2="23" stroke="var(--jm-accent-5)" stroke-width="1.5" stroke-linecap="round"/>
                  </template>
                </svg>
              </n-button>
            </template>
          </n-input>
        </n-form-item>

        <!-- API 测试连接 -->
        <div class="test-api-row">
          <button
            class="test-btn"
            :class="{ testing: testLoading, success: testResult === 'ok', fail: testResult === 'fail' }"
            :disabled="!formData.provider || !formData.apiKey || testLoading"
            @click="testConnection"
          >
            <svg v-if="testLoading" class="spin" viewBox="0 0 24 24" width="14" height="14" fill="none">
              <path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
            </svg>
            <svg v-else-if="testResult === 'ok'" viewBox="0 0 24 24" width="14" height="14" fill="none">
              <path d="M20 6L9 17l-5-5" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
            <svg v-else-if="testResult === 'fail'" viewBox="0 0 24 24" width="14" height="14" fill="none">
              <path d="M18 6L6 18M6 6l12 12" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
            </svg>
            <svg v-else viewBox="0 0 24 24" width="14" height="14" fill="none">
              <path d="M22 11.08V12a10 10 0 11-5.93-9.14" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
              <polyline points="22,4 12,14.01 9,11.01" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
            {{ testLoading ? '测试中...' : '测试连接' }}
          </button>
          <span v-if="testMessage" class="test-msg" :class="{ 'msg-ok': testResult === 'ok', 'msg-fail': testResult === 'fail' }">{{ testMessage }}</span>
        </div>

        <!-- 高级配置折叠 -->
        <div class="adv-toggle" @click="showAdvanced = !showAdvanced">
          <div class="adv-toggle-left">
            <svg viewBox="0 0 24 24" width="14" height="14" fill="none">
              <path d="M12.22 2h-.44a2 2 0 00-2 2v.18a2 2 0 01-1 1.73l-.43.25a2 2 0 01-2 0l-.15-.08a2 2 0 00-2.73.73l-.22.38a2 2 0 00.73 2.73l.15.1a2 2 0 011 1.72v.51a2 2 0 01-1 1.74l-.15.09a2 2 0 00-.73 2.73l.22.38a2 2 0 002.73.73l.15-.08a2 2 0 012 0l.43.25a2 2 0 011 1.73V20a2 2 0 002 2h.44a2 2 0 002-2v-.18a2 2 0 011-1.73l.43-.25a2 2 0 012 0l.15.08a2 2 0 002.73-.73l.22-.39a2 2 0 00-.73-2.73l-.15-.08a2 2 0 01-1-1.74v-.5a2 2 0 011-1.74l.15-.09a2 2 0 00.73-2.73l-.22-.38a2 2 0 00-2.73-.73l-.15.08a2 2 0 01-2 0l-.43-.25a2 2 0 01-1-1.73V4a2 2 0 00-2-2z" stroke="currentColor" stroke-width="1.5"/>
              <circle cx="12" cy="12" r="3" stroke="currentColor" stroke-width="1.5"/>
            </svg>
            <span class="adv-label">高级配置</span>
            <span class="adv-hint">Token · 端口</span>
          </div>
          <svg viewBox="0 0 24 24" width="14" height="14" fill="none" class="adv-chevron" :class="{ open: showAdvanced }">
            <polyline points="6,9 12,15 18,9" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
        </div>

        <template v-if="showAdvanced">
          <n-form-item label="访问 Token" path="token">
            <div class="token-row">
              <n-input
                v-model:value="formData.token"
                :type="showToken ? 'text' : 'password'"
                placeholder="32位随机令牌"
                class="flex-1"
              />
              <n-button quaternary size="small" @click="showToken = !showToken" class="sm-btn">
                <svg viewBox="0 0 24 24" width="14" height="14" fill="none">
                  <template v-if="showToken">
                    <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8S1 12 1 12z" stroke="var(--jm-accent-5)" stroke-width="1.5"/>
                    <circle cx="12" cy="12" r="3" stroke="var(--jm-accent-5)" stroke-width="1.5"/>
                  </template>
                  <template v-else>
                    <path d="M17.94 17.94A10.07 10.07 0 0112 20c-7 0-11-8-11-8a18.45 18.45 0 015.06-5.94" stroke="var(--jm-accent-5)" stroke-width="1.5" stroke-linecap="round"/>
                    <line x1="1" y1="1" x2="23" y2="23" stroke="var(--jm-accent-5)" stroke-width="1.5" stroke-linecap="round"/>
                  </template>
                </svg>
              </n-button>
              <n-button quaternary size="small" @click="copyToken" class="sm-btn">
                <svg viewBox="0 0 24 24" width="14" height="14" fill="none">
                  <rect x="9" y="9" width="13" height="13" rx="2" stroke="var(--jm-accent-5)" stroke-width="1.5"/>
                  <path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1" stroke="var(--jm-accent-5)" stroke-width="1.5"/>
                </svg>
              </n-button>
              <n-button quaternary size="small" @click="regenerateToken" class="sm-btn">
                <svg viewBox="0 0 24 24" width="14" height="14" fill="none">
                  <path d="M1 4v6h6M23 20v-6h-6" stroke="var(--jm-accent-5)" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
                  <path d="M20.49 9A9 9 0 005.64 5.64L1 10m22 4l-4.64 4.36A9 9 0 013.51 15" stroke="var(--jm-accent-5)" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
                </svg>
              </n-button>
            </div>
          </n-form-item>

          <div class="port-row">
            <n-form-item label="Web 端口" path="webPort">
              <n-input-number v-model:value="formData.webPort" :min="1024" :max="65535" style="width: 100%" />
            </n-form-item>
            <n-form-item label="通讯端口" path="bridgePort">
              <n-input-number v-model:value="formData.bridgePort" :min="1024" :max="65535" style="width: 100%" />
            </n-form-item>
          </div>
        </template>
      </n-form>

      <!-- 底部操作 -->
      <div class="actions">
        <n-button @click="$router.back()" quaternary size="small">返回</n-button>
        <n-button
          type="primary"
          @click="startDeploy"
          :loading="deploying"
          :disabled="!formData.model || !formData.apiKey || (isCustom && !formData.customBaseUrl)"
        >
          开始部署
        </n-button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { NButton, NForm, NFormItem, NInput, NInputNumber, NSelect } from 'naive-ui'
import { generateToken, deploy, checkPorts, testApiConnection, MODEL_PROVIDERS } from '@/api/deploy'
import gm from '@/utils/gmssh'

const props = defineProps({
  deployMode: { type: String, default: 'docker' }
})

const router = useRouter()
const showToken = ref(false)
const showApiKey = ref(false)
const showAdvanced = ref(false)
const deploying = ref(false)
const testLoading = ref(false)
const testResult = ref('')   // '' | 'ok' | 'fail'
const testMessage = ref('')

const formData = ref({
  token: '',
  webPort: 18789,
  bridgePort: 18790,
  provider: null,
  model: null,
  apiKey: '',
  customBaseUrl: '',
  customApiMode: 'openai',
  ollamaPort: 11434,
})

const isCustom = computed(() => formData.value.provider === 'custom')
const isOllama = computed(() => formData.value.provider === 'ollama')

const providerOptions = computed(() =>
  MODEL_PROVIDERS.map(p => ({ label: p.displayName, value: p.provider }))
)

const modelOptions = computed(() => {
  if (!formData.value.provider) return []
  const p = MODEL_PROVIDERS.find(p => p.provider === formData.value.provider)
  return p ? p.models.map(m => ({ label: m.name, value: m.id })) : []
})

const selectedProviderInfo = computed(() =>
  formData.value.provider
    ? MODEL_PROVIDERS.find(p => p.provider === formData.value.provider)
    : null
)

function onProviderChange() {
  formData.value.model = null
  formData.value.customBaseUrl = ''
  testResult.value = ''
  testMessage.value = ''
}

async function testConnection() {
  if (!formData.value.provider || !formData.value.apiKey) {
    gm.warning('请先选择提供商并填写 API Key')
    return
  }
  testLoading.value = true
  testResult.value = ''
  testMessage.value = ''
  try {
    const provider = MODEL_PROVIDERS.find(p => p.provider === formData.value.provider)
    let baseUrl = isCustom.value ? formData.value.customBaseUrl : (provider?.baseUrl || '')
    if (isOllama.value) {
      baseUrl = `http://localhost:${formData.value.ollamaPort}/v1`
    }
    const apiMode = isCustom.value ? formData.value.customApiMode : (provider?.apiMode || 'openai')
    const result = await testApiConnection({
      baseUrl,
      apiKey: formData.value.apiKey,
      provider: formData.value.provider,
      apiMode,
    })
    if (result.reachable) {
      testResult.value = 'ok'
      testMessage.value = result.message
      gm.success(result.message)
    } else {
      testResult.value = 'fail'
      testMessage.value = result.message
      gm.error(result.message)
    }
  } catch (e) {
    testResult.value = 'fail'
    testMessage.value = '测试失败: ' + (e.message || '未知错误')
    gm.error('连接测试失败')
  } finally {
    testLoading.value = false
  }
}

function copyToken() {
  if (!formData.value.token) return
  navigator.clipboard.writeText(formData.value.token)
    .then(() => gm.success('Token 已复制'))
    .catch(() => gm.warning('复制失败'))
}

async function regenerateToken() {
  try {
    const r = await generateToken()
    formData.value.token = r.token
    gm.success('已生成新 Token')
  } catch { gm.error('生成失败') }
}

async function startDeploy() {
  if (!formData.value.model || !formData.value.apiKey) {
    gm.warning('请选择模型并填写 API Key')
    return
  }
  deploying.value = true
  try {
    const portResult = await checkPorts([formData.value.webPort, formData.value.bridgePort])
    const occupied = portResult.results.filter(r => !r.available)
    if (occupied.length > 0) {
      const msgs = occupied.map(r => `${r.port}${r.process ? ` (${r.process})` : ''}`)
      gm.error(`端口被占用：${msgs.join('、')}`)
      showAdvanced.value = true
      return
    }
    await deploy({
      token: formData.value.token,
      webPort: formData.value.webPort,
      bridgePort: formData.value.bridgePort,
      provider: formData.value.provider,
      model: formData.value.model,
      apiKey: formData.value.apiKey,
      customBaseUrl: isOllama.value
        ? `http://localhost:${formData.value.ollamaPort}/v1`
        : formData.value.customBaseUrl,
      deployMode: props.deployMode,
    })
    sessionStorage.setItem('deploy_token', formData.value.token)
    // 先替换当前历史条目为 step=setup，这样浏览器返回时跳过环境检查
    await router.replace({ path: '/console', query: { step: 'setup' } })
    router.push({ path: '/progress', query: { token: formData.value.token, mode: props.deployMode } })
  } catch (e) {
    gm.error('部署失败：' + (e.message || '未知错误'))
  } finally {
    deploying.value = false
  }
}

const STORAGE_KEY = 'deploy_setup_form'

function saveForm() {
  try {
    sessionStorage.setItem(STORAGE_KEY, JSON.stringify(formData.value))
  } catch {}
}

function loadForm() {
  try {
    const saved = sessionStorage.getItem(STORAGE_KEY)
    if (saved) {
      const data = JSON.parse(saved)
      Object.assign(formData.value, data)
    }
  } catch {}
}

watch(formData, saveForm, { deep: true })

onMounted(async () => {
  loadForm()
  if (!formData.value.token) {
    try {
      const r = await generateToken()
      formData.value.token = r.token
    } catch {
      const c = 'abcdefghijklmnopqrstuvwxyz0123456789'
      formData.value.token = Array.from({ length: 32 }, () => c[Math.floor(Math.random() * c.length)]).join('')
    }
  }
})
</script>

<style scoped>
.deploy-setup-page {
  width: 100%;
  height: 100%;
  overflow-y: auto;
  padding: 20px;
}

.setup-container {
  max-width: 460px;
  margin: 0 auto;
}

.setup-form {
  background: rgba(var(--jm-accent-1-rgb), 0.4);
  border: 1px solid var(--jm-accent-2);
  border-radius: 10px;
  padding: 22px 22px 10px;
}

/* 提供商提示 */
.provider-tip {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 11px;
  color: var(--jm-accent-4);
  padding: 0 2px;
  margin: -8px 0 10px;
}

.ollama-port-section { margin: -4px 0 8px; }
.ollama-port-section .n-form-item { margin-bottom: 0; }
.ollama-port-section .n-form-item-blank { display: flex; align-items: center; gap: 10px; }
.port-hint { font-size: 11px; color: var(--jm-accent-4); white-space: nowrap; }

.tip-link {
  color: var(--jm-primary-2);
  text-decoration: none;
  font-weight: 500;
  white-space: nowrap;
}
.tip-link:hover {
  color: var(--jm-primary-1);
}

/* 高级配置切换 */
.adv-toggle {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 9px 12px;
  margin: 0 0 12px;
  border-radius: 8px;
  border: 1px solid var(--jm-accent-2);
  background: rgba(var(--jm-accent-1-rgb), 0.3);
  cursor: pointer;
  user-select: none;
  transition: border-color 0.2s, background 0.2s;
}
.adv-toggle:hover {
  border-color: var(--jm-accent-3);
  background: rgba(var(--jm-accent-1-rgb), 0.6);
}
.adv-toggle-left {
  display: flex;
  align-items: center;
  gap: 7px;
  color: var(--jm-accent-5);
}
.adv-label {
  font-size: 12px;
  font-weight: 500;
  color: var(--jm-accent-6);
}
.adv-hint {
  font-size: 11px;
  color: var(--jm-accent-4);
}
.adv-chevron {
  color: var(--jm-accent-4);
  transition: transform 0.2s;
}
.adv-chevron.open {
  transform: rotate(180deg);
}

/* Token 行 */
.token-row {
  display: flex;
  align-items: center;
  gap: 2px;
  width: 100%;
}
.flex-1 { flex: 1; }
.sm-btn {
  flex-shrink: 0;
  padding: 0 5px;
  opacity: 0.6;
  transition: opacity 0.15s;
}
.sm-btn:hover { opacity: 1; }

/* 端口 */
.port-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
}

/* 底部操作 */
.actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 0 0;
}

/* 测试连接 */
.test-api-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin: -4px 0 14px;
}
.test-btn {
  display: flex;
  align-items: center;
  gap: 5px;
  padding: 6px 14px;
  border-radius: 6px;
  border: 1px solid var(--jm-accent-2);
  background: rgba(var(--jm-accent-1-rgb), 0.3);
  color: var(--jm-accent-6);
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;
}
.test-btn:hover:not(:disabled) {
  border-color: var(--jm-primary-2);
  color: var(--jm-primary-2);
  background: rgba(var(--jm-primary-1-rgb), 0.08);
}
.test-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}
.test-btn.testing {
  border-color: var(--jm-primary-2);
  color: var(--jm-primary-2);
}
.test-btn.success {
  border-color: var(--jm-success-color);
  color: var(--jm-success-color);
  background: rgba(var(--jm-success-color-rgb, 34,197,94), 0.06);
}
.test-btn.fail {
  border-color: var(--jm-error-color);
  color: var(--jm-error-color);
  background: rgba(var(--jm-error-color-rgb, 239,68,68), 0.06);
}
.test-msg {
  font-size: 11px;
  color: var(--jm-accent-4);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.test-msg.msg-ok { color: var(--jm-success-color); }
.test-msg.msg-fail { color: var(--jm-error-color); }

@keyframes spin {
  to { transform: rotate(360deg); }
}
.spin {
  animation: spin 1s linear infinite;
}
</style>
