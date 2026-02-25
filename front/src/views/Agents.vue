<template>
  <div class="agents-page">
    <div class="agents-container fade-in-up">
      <!-- 顶部操作栏 -->
      <div class="agents-header">
        <div class="header-left">
          <h2 class="page-title">
            <svg viewBox="0 0 24 24" width="20" height="20" fill="none">
              <circle cx="12" cy="8" r="4" stroke="currentColor" stroke-width="1.5"/>
              <path d="M4 21v-1a6 6 0 0112 0v1" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
              <path d="M16 11h6M19 8v6" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
            </svg>
            Agent 人格管理
          </h2>
          <span class="header-hint">编辑 AI 的身份、记忆和灵魂</span>
        </div>
        <div class="header-actions">
          <button class="refresh-btn" @click="fetchData()" :disabled="loading" title="刷新">
            <svg :class="{ spinning: loading }" viewBox="0 0 24 24" width="16" height="16" fill="none"><path d="M1 4v6h6M23 20v-6h-6" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/><path d="M20.49 9A9 9 0 005.64 5.64L1 10m22 4l-4.64 4.36A9 9 0 013.51 15" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>
          </button>
          <n-dropdown
            :options="templateDropdownOptions"
            trigger="click"
            @select="onTemplateSelect"
            :to="false"
          >
            <button class="tpl-btn">
              <svg viewBox="0 0 24 24" width="14" height="14" fill="none">
                <path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z" stroke="currentColor" stroke-width="1.5" stroke-linejoin="round"/>
              </svg>
              应用模板
              <svg viewBox="0 0 24 24" width="12" height="12" fill="none">
                <polyline points="6,9 12,15 18,9" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
              </svg>
            </button>
          </n-dropdown>
        </div>
      </div>

      <!-- 当前人格提示 -->
      <div class="current-persona" v-if="activeTemplate">
        <svg viewBox="0 0 24 24" width="14" height="14" fill="none">
          <path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z" stroke="currentColor" stroke-width="1.5" stroke-linejoin="round"/>
        </svg>
        <span>当前人格：<strong>{{ activeTemplate }}</strong></span>
      </div>

      <!-- Tab 切换 -->
      <div class="tab-bar">
        <button
          v-for="tab in tabs"
          :key="tab.key"
          class="tab-item"
          :class="{ active: activeTab === tab.key }"
          @click="switchTab(tab.key)"
        >
          <span class="tab-icon" v-html="tab.icon"></span>
          <span class="tab-label">{{ tab.label }}</span>
          <span class="tab-desc">{{ tab.desc }}</span>
          <span v-if="isModified(tab.key)" class="tab-dot"></span>
        </button>
      </div>

      <!-- 编辑区 -->
      <div class="editor-panel">
        <div class="editor-toolbar">
          <div class="toolbar-left">
            <span class="file-badge">{{ activeFileName }}</span>
            <span v-if="isModified(activeTab)" class="modified-hint">● 已修改</span>
          </div>
          <div class="toolbar-right">
            <button class="tool-btn" @click="resetFile" :disabled="saving">
              <svg viewBox="0 0 24 24" width="14" height="14" fill="none">
                <path d="M1 4v6h6M23 20v-6h-6" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
                <path d="M20.49 9A9 9 0 005.64 5.64L1 10m22 4l-4.64 4.36A9 9 0 013.51 15" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
              </svg>
              恢复默认
            </button>
            <button class="tool-btn save-btn" @click="saveFile" :disabled="saving || !isModified(activeTab)">
              <svg v-if="saving" class="spin" viewBox="0 0 24 24" width="14" height="14" fill="none">
                <path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
              </svg>
              <svg v-else viewBox="0 0 24 24" width="14" height="14" fill="none">
                <path d="M19 21H5a2 2 0 01-2-2V5a2 2 0 012-2h11l5 5v11a2 2 0 01-2 2z" stroke="currentColor" stroke-width="1.5"/>
                <polyline points="17,21 17,13 7,13 7,21" stroke="currentColor" stroke-width="1.5"/>
                <polyline points="7,3 7,8 15,8" stroke="currentColor" stroke-width="1.5"/>
              </svg>
              保存
            </button>
          </div>
        </div>

        <div class="editor-wrapper" v-if="!loading">
          <textarea
            ref="editorRef"
            class="editor-textarea"
            :value="currentContent"
            @input="onInput"
            placeholder="在此编辑 Markdown 内容..."
            spellcheck="false"
          ></textarea>
        </div>
        <div v-else class="editor-loading">
          <div class="loading-spinner"></div>
          <span>加载中...</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { NDropdown } from 'naive-ui'
import { getAgentFiles, saveAgentFile, resetAgentFile, getAgentTemplates, applyAgentTemplate } from '@/api/agent'
import gm from '@/utils/gmssh'
import cache from '@/stores/cache'

const loading = ref(true)
const saving = ref(false)
const activeTab = ref('IDENTITY')
const activeTemplate = ref(sessionStorage.getItem('agent_active_template') || '')
const editorRef = ref(null)

// 通过外部 cache 缓存
const files = ref({ IDENTITY: '', USER: '', SOUL: '' })
const originals = ref({ IDENTITY: '', USER: '', SOUL: '' })
const templates = ref([])

const tabs = [
  {
    key: 'IDENTITY',
    label: 'Identity',
    desc: '我是谁',
    icon: '<svg viewBox="0 0 24 24" width="16" height="16" fill="none"><circle cx="12" cy="8" r="4" stroke="currentColor" stroke-width="1.5"/><path d="M6 21v-1a6 6 0 0112 0v1" stroke="currentColor" stroke-width="1.5"/></svg>',
  },
  {
    key: 'USER',
    label: 'User',
    desc: '你是谁',
    icon: '<svg viewBox="0 0 24 24" width="16" height="16" fill="none"><path d="M17 21v-2a4 4 0 00-4-4H5a4 4 0 00-4 4v2" stroke="currentColor" stroke-width="1.5"/><circle cx="9" cy="7" r="4" stroke="currentColor" stroke-width="1.5"/><path d="M23 21v-2a4 4 0 00-3-3.87M16 3.13a4 4 0 010 7.75" stroke="currentColor" stroke-width="1.5"/></svg>',
  },
  {
    key: 'SOUL',
    label: 'Soul',
    desc: '怎么聊',
    icon: '<svg viewBox="0 0 24 24" width="16" height="16" fill="none"><path d="M20.84 4.61a5.5 5.5 0 00-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 00-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 000-7.78z" stroke="currentColor" stroke-width="1.5"/></svg>',
  },
]

const activeFileName = computed(() => {
  const map = { IDENTITY: 'IDENTITY.md', USER: 'USER.md', SOUL: 'SOUL.md' }
  return map[activeTab.value] || ''
})

const currentContent = computed(() => files.value[activeTab.value] || '')

function isModified(key) {
  return files.value[key] !== originals.value[key]
}

function switchTab(key) {
  activeTab.value = key
}

function onInput(e) {
  files.value[activeTab.value] = e.target.value
}

const templateDropdownOptions = computed(() =>
  templates.value.map(t => ({
    label: `${t.name}`,
    key: t.key,
  }))
)

// 模板特征识别（IDENTITY.md 首行关键词 → 模板名）
const templateSignatures = {
  '首席系统架构师': '顶级架构师',
  'Chief SRE': '顶级架构师',
  '自动化运维专家': '高效运维助手',
  'DevOps Ninja': '高效运维助手',
  '安全运维专家': '云原生安全官',
  'SecOps Lead': '云原生安全官',
}

function detectActiveTemplate() {
  const identity = files.value.IDENTITY || ''
  const firstLine = identity.split('\n').find(l => l.trim().startsWith('#')) || ''
  for (const [keyword, name] of Object.entries(templateSignatures)) {
    if (firstLine.includes(keyword)) {
      activeTemplate.value = name
      return
    }
  }
  // 没匹配到已知模板，检查 sessionStorage
  const saved = sessionStorage.getItem('agent_active_template')
  if (saved) activeTemplate.value = saved
}

async function fetchData() {
  loading.value = true
  try {
    const [filesRes, tplRes] = await Promise.all([getAgentFiles(), getAgentTemplates()])
    if (filesRes?.files) {
      filesRes.files.forEach(f => {
        files.value[f.name] = f.content
        originals.value[f.name] = f.content
      })
    }
    if (tplRes?.templates) {
      templates.value = tplRes.templates
    }
    cache.agentFiles = { ...files.value }
    cache.agentOriginals = { ...originals.value }
    cache.agentTemplates = [...templates.value]
    detectActiveTemplate()
  } catch (e) {
    gm.error('加载失败: ' + (e.message || ''))
  } finally {
    loading.value = false
  }
}

async function saveFile() {
  saving.value = true
  try {
    await saveAgentFile({
      name: activeTab.value,
      content: files.value[activeTab.value],
    })
    originals.value[activeTab.value] = files.value[activeTab.value]
    gm.success(`${activeFileName.value} 已保存`)
  } catch (e) {
    gm.error('保存失败: ' + (e.message || ''))
  } finally {
    saving.value = false
  }
}

async function resetFile() {
  const gmApi = gm.getGmApi()
  if (gmApi?.dialog) {
    gmApi.dialog.warning({
      title: '恢复默认',
      content: `确定要将 ${activeFileName.value} 恢复为 OpenClaw 默认内容吗？当前内容将被覆盖。`,
      positiveText: '确定',
      negativeText: '取消',
      onPositiveClick: doReset,
    })
  } else {
    if (confirm(`确定恢复 ${activeFileName.value} 为默认内容？`)) doReset()
  }
}

async function doReset() {
  saving.value = true
  try {
    await resetAgentFile({ name: activeTab.value })
    // 重新拉取内容
    const res = await getAgentFiles()
    if (res?.files) {
      res.files.forEach(f => {
        files.value[f.name] = f.content
        originals.value[f.name] = f.content
      })
    }
    gm.success(`${activeFileName.value} 已恢复默认`)
  } catch (e) {
    gm.error('重置失败: ' + (e.message || ''))
  } finally {
    saving.value = false
  }
}

async function onTemplateSelect(key) {
  const tpl = templates.value.find(t => t.key === key)
  if (!tpl) return

  const doApply = async () => {
    saving.value = true
    try {
      await applyAgentTemplate({ key })
      // 重新拉取
      const res = await getAgentFiles()
      if (res?.files) {
        res.files.forEach(f => {
          files.value[f.name] = f.content
          originals.value[f.name] = f.content
        })
      }
      activeTemplate.value = tpl.name
      sessionStorage.setItem('agent_active_template', tpl.name)
      gm.success(`已应用「${tpl.name}」模板`)
    } catch (e) {
      gm.error('应用模板失败: ' + (e.message || ''))
    } finally {
      saving.value = false
    }
  }

  const gmApi = gm.getGmApi()
  if (gmApi?.dialog) {
    gmApi.dialog.warning({
      title: '应用模板',
      content: `确定应用「${tpl.name}」模板吗？这将覆盖 IDENTITY.md、USER.md、SOUL.md 三个文件的内容。`,
      positiveText: '确定',
      negativeText: '取消',
      onPositiveClick: doApply,
    })
  } else {
    if (confirm(`确定应用「${tpl.name}」模板？将覆盖三个文件。`)) doApply()
  }
}

onMounted(() => {
  if (cache.agentFiles !== null) {
    files.value = { ...cache.agentFiles }
    originals.value = { ...cache.agentOriginals }
    templates.value = [...(cache.agentTemplates || [])]
    loading.value = false
    detectActiveTemplate()
    return
  }
  fetchData()
})
</script>

<style scoped>
.agents-page {
  width: 100%;
  height: 100%;
  overflow-y: auto;
  padding: 20px;
}

.agents-container {
  max-width: 800px;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

/* 顶部 */
.agents-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}
.header-left {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.page-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 18px;
  font-weight: 600;
  color: var(--jm-accent-7);
  margin: 0;
}
.header-hint {
  font-size: 12px;
  color: var(--jm-accent-4);
  padding-left: 28px;
}

/* 模板按钮 */
.tpl-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 7px 14px;
  border-radius: 8px;
  border: 1px solid var(--jm-accent-2);
  background: rgba(var(--jm-accent-1-rgb), 0.3);
  color: var(--jm-accent-6);
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;
}
.tpl-btn:hover {
  border-color: var(--jm-primary-2);
  color: var(--jm-primary-2);
  background: rgba(var(--jm-primary-1-rgb), 0.08);
}

/* 下拉菜单样式修正 */
:deep(.n-dropdown-menu) {
  background: #1e1e2e !important;
  border: 1px solid var(--jm-accent-2) !important;
  border-radius: 8px !important;
  box-shadow: 0 8px 24px rgba(0,0,0,0.4) !important;
}
:deep(.n-dropdown-option-body) {
  color: var(--jm-accent-6) !important;
  font-size: 13px !important;
}
:deep(.n-dropdown-option-body:hover),
:deep(.n-dropdown-option-body--pending) {
  background: rgba(var(--jm-primary-1-rgb), 0.12) !important;
  color: var(--jm-primary-1) !important;
}

/* 当前人格 */
.current-persona {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 14px;
  border-radius: 8px;
  background: rgba(var(--jm-primary-1-rgb), 0.06);
  border: 1px solid rgba(var(--jm-primary-1-rgb), 0.15);
  color: var(--jm-primary-2);
  font-size: 12px;
}
.current-persona strong {
  color: var(--jm-primary-1);
}

/* Tab 栏 */
.tab-bar {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 8px;
}
.tab-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  padding: 14px 10px;
  border-radius: 10px;
  border: 1px solid var(--jm-accent-2);
  background: rgba(var(--jm-accent-1-rgb), 0.3);
  color: var(--jm-accent-5);
  cursor: pointer;
  transition: all 0.2s;
  position: relative;
}
.tab-item:hover {
  border-color: var(--jm-accent-3);
  background: rgba(var(--jm-accent-1-rgb), 0.6);
}
.tab-item.active {
  border-color: var(--jm-primary-2);
  background: rgba(var(--jm-primary-1-rgb), 0.08);
  color: var(--jm-primary-1);
}
.tab-icon {
  display: flex;
  align-items: center;
}
.tab-label {
  font-size: 13px;
  font-weight: 600;
}
.tab-desc {
  font-size: 11px;
  opacity: 0.7;
}
.tab-dot {
  position: absolute;
  top: 8px;
  right: 8px;
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--jm-primary-1);
}

/* 编辑区 */
.editor-panel {
  border: 1px solid var(--jm-accent-2);
  border-radius: 10px;
  overflow: hidden;
  background: rgba(var(--jm-accent-1-rgb), 0.4);
  display: flex;
  flex-direction: column;
}

.editor-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 14px;
  border-bottom: 1px solid var(--jm-accent-2);
  background: rgba(var(--jm-accent-1-rgb), 0.3);
}
.toolbar-left {
  display: flex;
  align-items: center;
  gap: 8px;
}
.file-badge {
  font-size: 12px;
  font-weight: 600;
  color: var(--jm-accent-5);
  padding: 2px 8px;
  background: rgba(var(--jm-accent-1-rgb), 0.6);
  border-radius: 4px;
  font-family: 'SF Mono', 'Fira Code', monospace;
}
.modified-hint {
  font-size: 11px;
  color: var(--jm-primary-1);
  font-weight: 500;
}
.toolbar-right {
  display: flex;
  align-items: center;
  gap: 6px;
}
.tool-btn {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 5px 10px;
  border-radius: 6px;
  border: 1px solid var(--jm-accent-2);
  background: transparent;
  color: var(--jm-accent-5);
  font-size: 11px;
  cursor: pointer;
  transition: all 0.15s;
}
.tool-btn:hover:not(:disabled) {
  border-color: var(--jm-accent-3);
  color: var(--jm-accent-6);
}
.tool-btn:disabled {
  opacity: 0.35;
  cursor: not-allowed;
}
.save-btn {
  border-color: var(--jm-primary-2);
  color: var(--jm-primary-2);
  background: rgba(var(--jm-primary-1-rgb), 0.06);
}
.save-btn:hover:not(:disabled) {
  background: rgba(var(--jm-primary-1-rgb), 0.15);
}

/* 编辑器 */
.editor-wrapper {
  display: flex;
  flex: 1;
}
.editor-textarea {
  width: 100%;
  min-height: 420px;
  padding: 16px;
  border: none;
  outline: none;
  resize: vertical;
  background: transparent;
  color: var(--jm-accent-7);
  font-family: 'SF Mono', 'Fira Code', 'Consolas', monospace;
  font-size: 13px;
  line-height: 1.7;
  tab-size: 2;
}
.editor-textarea::placeholder {
  color: var(--jm-accent-3);
}

/* 加载 */
.editor-loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 420px;
  gap: 12px;
  color: var(--jm-accent-4);
  font-size: 13px;
}
.loading-spinner {
  width: 24px;
  height: 24px;
  border: 2px solid var(--jm-accent-2);
  border-top-color: var(--jm-primary-1);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}
.spin {
  animation: spin 1s linear infinite;
}
</style>
