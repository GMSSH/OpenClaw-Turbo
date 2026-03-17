<template>
  <div class="models-page">
    <div class="models-container fade-in-up">
      <!-- 顶部 -->
      <div class="models-header">
        <div class="header-left">
          <h2 class="page-title">
            <svg viewBox="0 0 24 24" width="20" height="20" fill="none">
              <path d="M12 2a7 7 0 017 7c0 2.38-1.19 4.47-3 5.74V17a2 2 0 01-2 2h-4a2 2 0 01-2-2v-2.26C6.19 13.47 5 11.38 5 9a7 7 0 017-7z" stroke="currentColor" stroke-width="1.5"/>
              <line x1="9" y1="21" x2="15" y2="21" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
            </svg>
            {{ t('models.title') }}
          </h2>
          <span class="header-hint">{{ t('models.subtitle') }}</span>
        </div>
        <button class="refresh-btn" @click="loadConfig" :disabled="loading">
          <svg :class="{ spinning: loading }" viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 12a9 9 0 1 1-9-9c2.52 0 4.93 1 6.74 2.74L21 8"/><path d="M21 3v5h-5"/></svg>
        </button>
      </div>

      <!-- 默认模型 -->
      <div class="default-model-section">
        <span class="dm-label">{{ t('models.globalDefault') }}</span>
        <n-select
          v-model:value="defaultModel"
          :options="allModelOptions"
          :placeholder="t('models.defaultPlaceholder')"
          size="small"
          style="flex:1"
          filterable
        />
      </div>

      <!-- 供应商列表 -->
      <div v-if="loading" class="loading-state"><div class="loading-spinner"></div></div>
      <template v-else>
        <div v-for="(provider, pid) in providers" :key="pid" class="provider-card">
          <div class="provider-header" @click="toggleExpand(pid)">
            <div class="ph-left">
              <span class="provider-name">{{ pid }}</span>
              <span class="provider-api">{{ provider.api || 'openai' }}</span>
              <span class="model-count">{{ t('models.modelCount', { count: (provider.models || []).length }) }}</span>
            </div>
            <div class="ph-right">
              <button class="icon-btn danger" @click.stop="confirmRemoveProvider(pid)">
                <svg viewBox="0 0 24 24" width="14" height="14" fill="none"><line x1="18" y1="6" x2="6" y2="18" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/><line x1="6" y1="6" x2="18" y2="18" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/></svg>
              </button>
              <svg viewBox="0 0 24 24" width="14" height="14" fill="none" class="expand-chevron" :class="{ open: expanded[pid] }"><polyline points="6,9 12,15 18,9" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>
            </div>
          </div>
          <div v-show="expanded[pid]" class="provider-body">
            <!-- 供应商信息 -->
            <div class="provider-fields">
              <div class="pf-row">
                <label>{{ t('models.providerApiProtocol') }}</label>
                <n-select v-model:value="provider.api" :options="apiModeOptions" size="small" style="width:200px" />
              </div>
              <div class="pf-row">
                <label>Base URL</label>
                <n-input v-model:value="provider.baseUrl" size="small" placeholder="https://api.example.com/v1" />
              </div>
              <div class="pf-row">
                <label>API Key</label>
                <n-input v-model:value="provider.apiKey" size="small" type="password" show-password-on="click" placeholder="sk-..." />
              </div>
            </div>
            <!-- 模型列表 -->
            <div class="model-list-title">{{ t('models.modelList') }}</div>
            <div class="model-list">
              <div v-for="(model, i) in (provider.models || [])" :key="i" class="model-row">
                <n-input v-model:value="model.id" size="small" :placeholder="t('models.modelIdPlaceholder')" class="model-id-input" />
                <button class="icon-btn danger sm" @click="provider.models.splice(i, 1)">
                  <svg viewBox="0 0 24 24" width="12" height="12" fill="none"><line x1="18" y1="6" x2="6" y2="18" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/><line x1="6" y1="6" x2="18" y2="18" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/></svg>
                </button>
              </div>
              <button class="add-model-btn" @click="addModel(pid)">
                <svg viewBox="0 0 24 24" width="14" height="14" fill="none"><line x1="12" y1="5" x2="12" y2="19" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/><line x1="5" y1="12" x2="19" y2="12" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/></svg>
                {{ t('models.addModel') }}
              </button>
            </div>
          </div>
        </div>

        <!-- 添加供应商 -->
        <button class="add-provider-btn" @click="showAddProvider = true">
          <svg viewBox="0 0 24 24" width="16" height="16" fill="none"><line x1="12" y1="5" x2="12" y2="19" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/><line x1="5" y1="12" x2="19" y2="12" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/></svg>
          {{ t('models.addProvider') }}
        </button>

        <!-- 从预设添加 -->
        <div class="preset-section">
          <span class="preset-label">{{ t('models.quickAddPreset') }}</span>
          <div class="preset-chips">
            <button
              v-for="p in availablePresets"
              :key="p.provider"
              class="preset-chip"
              @click="addPresetProvider(p)"
            >{{ p.displayName }}</button>
          </div>
        </div>
      </template>

      <!-- 保存 -->
      <div class="save-bar">
        <n-button type="primary" @click="doSave" :loading="saving">
          <svg viewBox="0 0 24 24" width="14" height="14" fill="none" style="margin-right:4px"><path d="M19 21H5a2 2 0 01-2-2V5a2 2 0 012-2h11l5 5v11a2 2 0 01-2 2z" stroke="currentColor" stroke-width="1.5"/><path d="M17 21v-8H7v8M7 3v5h8" stroke="currentColor" stroke-width="1.5"/></svg>
          {{ t('common.saveConfig') }}
        </n-button>
      </div>

      <!-- 添加供应商弹框 -->
      <n-modal v-model:show="showAddProvider" preset="card" :bordered="false" size="small" style="max-width:400px" :title="t('models.addProviderTitle')">
        <div style="display:flex;flex-direction:column;gap:12px">
          <n-input v-model:value="newProvider.id" :placeholder="t('models.providerIdPlaceholder')" size="small" />
          <n-select v-model:value="newProvider.api" :options="apiModeOptions" size="small" :placeholder="t('models.providerApiProtocol')" />
          <n-input v-model:value="newProvider.baseUrl" placeholder="Base URL" size="small" />
          <n-input v-model:value="newProvider.apiKey" :placeholder="t('models.apiKeyOptional')" size="small" type="password" show-password-on="click" />
        </div>
        <template #footer>
          <div style="display:flex;justify-content:flex-end;gap:8px">
            <n-button size="small" @click="showAddProvider = false">{{ t('common.cancel') }}</n-button>
            <n-button type="primary" size="small" :disabled="!newProvider.id.trim()" @click="doConfirmAddProvider">{{ t('common.add') }}</n-button>
          </div>
        </template>
      </n-modal>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { NButton, NInput, NSelect, NModal } from 'naive-ui'
import { getModelsConfig, saveModelsConfig } from '@/api/model'
import { MODEL_PROVIDERS } from '@/api/deploy'
import gm from '@/utils/gmssh'

const { t } = useI18n()

const loading = ref(false)
const saving = ref(false)
const providers = reactive({})
const expanded = reactive({})
const defaultModel = ref('')

const showAddProvider = ref(false)
const newProvider = reactive({ id: '', api: 'openai-completions', baseUrl: '', apiKey: '' })

const apiModeOptions = computed(() => [
  { label: t('models.apiModes.openai'), value: 'openai' },
  { label: t('models.apiModes.openaiCompat'), value: 'openai-completions' },
  { label: t('models.apiModes.anthropic'), value: 'anthropic' },
  { label: t('models.apiModes.anthropicMessages'), value: 'anthropic-messages' },
  { label: t('models.apiModes.gemini'), value: 'gemini' },
  { label: t('models.apiModes.ollama'), value: 'ollama' },
])

// 所有配置的模型合成选项
const allModelOptions = computed(() => {
  const opts = []
  for (const [pid, p] of Object.entries(providers)) {
    for (const m of (p.models || [])) {
      if (!m.id) continue
      opts.push({ label: `${pid} / ${m.id}`, value: `${pid}/${m.id}` })
    }
  }
  return opts
})

// 可添加的预设供应商（排除已添加的）
const availablePresets = computed(() =>
  MODEL_PROVIDERS.filter(p => p.provider !== 'custom' && !providers[p.provider])
)

async function loadConfig() {
  loading.value = true
  try {
    const res = await getModelsConfig()
    const models = res?.models || {}
    const rawProviders = models.providers || {}
    // 清空再赋值
    Object.keys(providers).forEach(k => delete providers[k])
    for (const [pid, p] of Object.entries(rawProviders)) {
      providers[pid] = { ...p }
      expanded[pid] = false
    }
    // 读取默认模型
    if (res?.defaultModel) {
      defaultModel.value = res.defaultModel
    }
  } catch (e) {
    gm.error(t('models.loadFailed') + ': ' + (e.message || ''))
  } finally {
    loading.value = false
  }
}

function addModel(pid) {
  if (!providers[pid].models) providers[pid].models = []
  providers[pid].models.push({ id: '' })
}

function confirmRemoveProvider(pid) {
  // 不允许删除包含默认模型的供应商
  if (defaultModel.value && defaultModel.value.startsWith(pid + '/')) {
    gm.warning(t('models.providerHasDefault', { id: pid }))
    return
  }
  delete providers[pid]
  delete expanded[pid]
}

function toggleExpand(pid) {
  expanded[pid] = !expanded[pid]
}

function doConfirmAddProvider() {
  const id = newProvider.id.trim()
  if (!id) return
  if (providers[id]) {
    gm.warning(t('models.providerExists', { id }))
    return
  }
  providers[id] = {
    api: newProvider.api,
    baseUrl: newProvider.baseUrl,
    apiKey: newProvider.apiKey,
    models: [],
  }
  expanded[id] = true
  showAddProvider.value = false
  newProvider.id = ''
  newProvider.api = 'openai-completions'
  newProvider.baseUrl = ''
  newProvider.apiKey = ''
}

function addPresetProvider(preset) {
  const pid = preset.provider
  if (providers[pid]) return
  providers[pid] = {
    api: preset.apiMode || 'openai',
    baseUrl: preset.baseUrl || '',
    apiKey: '',
    models: preset.models.map(m => ({
      id: m.id.includes('/') ? m.id.split('/')[1] : m.id,
    })),
  }
  expanded[pid] = true
}

async function doSave() {
  saving.value = true
  try {
    const modelsConfig = {
      mode: 'merge',
      providers: { ...providers },
    }
    await saveModelsConfig({
      models: modelsConfig,
      defaultModel: defaultModel.value,
    })
    gm.success(t('models.saveSuccess'))
  } catch (e) {
    gm.error(t('models.saveFailed') + ': ' + (e.message || ''))
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  loadConfig()
})
</script>

<style scoped>
.models-page { width: 100%; height: 100%; overflow-y: auto; padding: 20px 24px; }
.models-container { max-width: 100%; margin: 0 auto; display: flex; flex-direction: column; gap: 16px; }

.models-header { display: flex; align-items: flex-start; justify-content: space-between; }
.header-left { display: flex; flex-direction: column; gap: 4px; }
.page-title { display: flex; align-items: center; gap: 8px; font-size: 18px; font-weight: 600; color: var(--jm-accent-7); margin: 0; }
.header-hint { font-size: 12px; color: var(--jm-accent-4); padding-left: 28px; }

/* 默认模型 */
.default-model-section {
  display: flex; align-items: center; gap: 12px;
  padding: 14px 18px; border-radius: 14px;
  background: rgba(var(--jm-accent-1-rgb), 0.2);
  backdrop-filter: blur(12px); -webkit-backdrop-filter: blur(12px);
  border: 1px solid var(--jm-glass-border);
  box-shadow: 
    var(--jm-glass-inner-glow),
    0 1px 3px rgba(0, 0, 0, 0.04), 0 4px 12px rgba(0, 0, 0, 0.03);
}
.dm-label { font-size: 12px; font-weight: 600; color: var(--jm-accent-5); white-space: nowrap; }

/* 供应商卡片 */
.provider-card {
  border: 1px solid var(--jm-glass-border);
  border-radius: 14px;
  background: var(--jm-glass-bg);
  backdrop-filter: blur(12px); -webkit-backdrop-filter: blur(12px);
  overflow: hidden;
  transition: all 0.35s cubic-bezier(0.34, 1.56, 0.64, 1);
  box-shadow:
    
    var(--jm-glass-inner-glow),
    0 1px 3px rgba(0, 0, 0, 0.04),
    0 4px 12px rgba(0, 0, 0, 0.03);
}
.provider-card:hover {
  border-color: var(--jm-glass-border-hover);
  box-shadow:
    0 2px 4px rgba(0, 0, 0, 0.06),
    0 12px 28px rgba(0, 0, 0, 0.06);
}
.provider-header {
  display: flex; align-items: center; justify-content: space-between;
  padding: 14px 18px; cursor: pointer;
  transition: background 0.2s;
}
.provider-header:hover { background: rgba(var(--jm-accent-1-rgb), 0.3); }
.ph-left { display: flex; align-items: center; gap: 10px; }
.ph-right { display: flex; align-items: center; gap: 8px; }
.provider-name { font-size: 14px; font-weight: 600; color: var(--jm-accent-7); letter-spacing: -0.01em; }
.provider-api {
  font-size: 10px; padding: 2px 10px; border-radius: 6px;
  background: rgba(var(--jm-primary-1-rgb), 0.08); color: var(--jm-primary-2);
  font-weight: 500; font-family: var(--jm-font-mono, monospace);
}
.model-count { font-size: 11px; color: var(--jm-accent-4); }

.expand-chevron { color: var(--jm-accent-4); transition: transform 0.2s; }
.expand-chevron.open { transform: rotate(180deg); }

.provider-body {
  padding: 0 18px 18px;
  margin: 0 6px 6px;
  border-radius: 0 0 10px 10px;
  background: rgba(var(--jm-accent-1-rgb), 0.08);
  box-shadow: inset 0 2px 6px rgba(0, 0, 0, 0.04);
  padding-top: 14px;
}

/* 供应商字段 */
.provider-fields { display: flex; flex-direction: column; gap: 10px; margin-bottom: 14px; }
.pf-row { display: flex; align-items: center; gap: 10px; }
.pf-row label { font-size: 12px; color: var(--jm-accent-5); min-width: 70px; flex-shrink: 0; }
.pf-row .n-input, .pf-row .n-select { flex: 1; }

/* 模型列表 */
.model-list-title { font-size: 12px; font-weight: 600; color: var(--jm-accent-5); margin-bottom: 8px; }
.model-list { display: flex; flex-direction: column; gap: 6px; }
.model-row {
  display: flex; align-items: center; gap: 6px;
  padding: 4px 8px; border-radius: 8px;
  background: var(--jm-glass-bg);
  transition: background 0.15s;
}
.model-row:hover { background: rgba(var(--jm-accent-1-rgb), 0.3); }
.model-id-input { flex: 1; }
.add-model-btn {
  display: flex; align-items: center; gap: 6px;
  padding: 7px 14px; border: none; border-radius: 8px;
  background: rgba(var(--jm-accent-1-rgb), 0.3); color: var(--jm-accent-4); font-size: 11px;
  cursor: pointer; transition: all 0.25s;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.04);
}
.add-model-btn:hover { background: rgba(var(--jm-primary-1-rgb), 0.08); color: var(--jm-primary-2); box-shadow: 0 2px 8px rgba(var(--jm-primary-1-rgb), 0.1); }

/* 按钮 */
.icon-btn {
  width: 26px; height: 26px; border-radius: 6px; border: none;
  background: transparent; color: var(--jm-accent-4); cursor: pointer;
  display: flex; align-items: center; justify-content: center;
  transition: all 0.15s;
}
.icon-btn.danger:hover { background: rgba(229,62,62,0.1); color: #fc8181; }
.icon-btn.sm { width: 22px; height: 22px; }

.add-provider-btn {
  display: flex; align-items: center; justify-content: center; gap: 8px;
  padding: 14px; border: 1px solid var(--jm-glass-border); border-radius: 14px;
  background: var(--jm-glass-bg); color: var(--jm-accent-4); font-size: 13px; font-weight: 500;
  cursor: pointer; transition: all 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
}
.add-provider-btn:hover {
  border-color: rgba(var(--jm-primary-1-rgb), 0.25); color: var(--jm-primary-2);
  background: rgba(var(--jm-primary-1-rgb), 0.04);
  transform: translateY(-2px);
  box-shadow: 0 4px 16px rgba(var(--jm-primary-1-rgb), 0.08);
}

/* 预设 */
.preset-section { display: flex; flex-direction: column; gap: 8px; }
.preset-label { font-size: 12px; color: var(--jm-accent-4); }
.preset-chips { display: flex; flex-wrap: wrap; gap: 6px; }
.preset-chip {
  padding: 6px 14px; border-radius: 20px;
  border: 1px solid var(--jm-glass-border);
  background: rgba(var(--jm-accent-1-rgb), 0.2);
  backdrop-filter: blur(8px); -webkit-backdrop-filter: blur(8px);
  color: var(--jm-accent-5); font-size: 11px; font-weight: 500;
  cursor: pointer; transition: all 0.25s cubic-bezier(0.34, 1.56, 0.64, 1); white-space: nowrap;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.04);
}
.preset-chip:hover {
  border-color: var(--jm-glass-border-hover); color: var(--jm-primary-2);
  background: rgba(var(--jm-primary-1-rgb), 0.06);
  transform: translateY(-1px);
  box-shadow: 0 3px 10px rgba(var(--jm-primary-1-rgb), 0.08);
}

/* 保存栏 */
.save-bar { display: flex; justify-content: flex-end; padding-top: 12px; border-top: 1px solid var(--jm-glass-border, rgba(var(--jm-accent-2-rgb), 0.4)); }

/* 刷新按钮 */
.refresh-btn {
  display: flex; align-items: center; justify-content: center;
  width: 32px; height: 32px; border-radius: 8px; border: 1px solid var(--jm-glass-border);
  background: transparent; color: var(--jm-accent-5); cursor: pointer; transition: all 0.2s;
}
.refresh-btn:hover { border-color: var(--jm-accent-3); color: var(--jm-accent-7); }
.refresh-btn:disabled { opacity: 0.35; cursor: not-allowed; }

/* 加载 */
.loading-state { display: flex; justify-content: center; padding: 40px; }
.loading-spinner { width: 24px; height: 24px; border: 2px solid var(--jm-accent-2); border-top-color: var(--jm-primary-1); border-radius: 50%; animation: spin 0.8s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }
.spinning { animation: spin 0.8s linear infinite; }

/* Input overrides */
:deep(.n-input) { border-radius: 10px !important; transition: box-shadow 0.3s !important; }
:deep(.n-input--focus) { box-shadow: 0 0 0 2px rgba(var(--jm-primary-1-rgb), 0.12) !important; }

/* Modal glassmorphism */
:deep(.n-card) {
  background: var(--jm-glass-bg, rgba(var(--jm-accent-1-rgb), 0.6)) !important;
  backdrop-filter: blur(25px) !important; -webkit-backdrop-filter: blur(25px) !important;
  border: 1px solid var(--jm-glass-border, rgba(var(--jm-accent-2-rgb), 0.4)) !important;
  border-radius: 16px !important;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.08) !important;
}
</style>
