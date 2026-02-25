<template>
  <div class="channels-page">
    <div class="channels-container fade-in-up">
      <!-- 顶部 -->
      <div class="channels-header">
        <div class="header-left">
          <h2 class="page-title">
            <svg viewBox="0 0 24 24" width="20" height="20" fill="none">
              <path d="M21 15a2 2 0 01-2 2H7l-4 4V5a2 2 0 012-2h14a2 2 0 012 2z" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
            三方平台接入
          </h2>
          <span class="header-hint">管理 OpenClaw 与外部通讯平台的连接</span>
        </div>
        <div class="header-actions">
          <button class="refresh-btn" @click="fetchChannels()" :disabled="loading" title="刷新">
            <svg :class="{ spinning: loading }" viewBox="0 0 24 24" width="16" height="16" fill="none"><path d="M1 4v6h6M23 20v-6h-6" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/><path d="M20.49 9A9 9 0 005.64 5.64L1 10m22 4l-4.64 4.36A9 9 0 013.51 15" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>
          </button>
          <button class="add-btn" @click="showAddPanel = true" v-if="!showAddPanel && !editingChannel">
          <svg viewBox="0 0 24 24" width="14" height="14" fill="none">
            <line x1="12" y1="5" x2="12" y2="19" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
            <line x1="5" y1="12" x2="19" y2="12" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
          </svg>
          新增通道
        </button>
        </div>
      </div>

      <!-- 已接入的通道卡片 -->
      <div v-if="channels.length > 0 && !showAddPanel && !editingChannel" class="channel-cards">
        <div v-for="ch in channels" :key="ch.key" class="channel-card">
          <div class="card-header">
            <div class="card-icon" v-html="getChannelIcon(ch.key)"></div>
            <div class="card-info">
              <span class="card-name">{{ getChannelDisplayName(ch.key) }}</span>
              <span class="card-key">{{ ch.key }}</span>
            </div>
            <div class="card-status" :class="ch.enabled ? 'enabled' : 'disabled'">
              {{ ch.enabled ? '已启用' : '已禁用' }}
            </div>
          </div>
          <div class="card-details">
            <div v-for="field in getChannelFields(ch)" :key="field.label" class="detail-item">
              <span class="detail-label">{{ field.label }}</span>
              <div class="detail-value-wrap">
                <n-tooltip :disabled="displayValue(field, ch.key).length <= 16" placement="top">
                  <template #trigger>
                    <span class="detail-value">{{ displayValue(field, ch.key) }}</span>
                  </template>
                  {{ displayValue(field, ch.key) }}
                </n-tooltip>
                <button v-if="field.masked" class="eye-btn" @click="toggleReveal(ch.key + field.label)">
                  <svg v-if="revealedFields[ch.key + field.label]" viewBox="0 0 24 24" width="13" height="13" fill="none"><path d="M17.94 17.94A10.07 10.07 0 0112 20c-7 0-11-8-11-8a18.45 18.45 0 015.06-5.94M9.9 4.24A9.12 9.12 0 0112 4c7 0 11 8 11 8a18.5 18.5 0 01-2.16 3.19M1 1l22 22" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/></svg>
                  <svg v-else viewBox="0 0 24 24" width="13" height="13" fill="none"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" stroke="currentColor" stroke-width="1.5"/><circle cx="12" cy="12" r="3" stroke="currentColor" stroke-width="1.5"/></svg>
                </button>
              </div>
            </div>
          </div>
          <div class="card-actions">
            <button class="card-act-btn" @click="editChannel(ch)">
              <svg viewBox="0 0 24 24" width="13" height="13" fill="none"><path d="M11 4H4a2 2 0 00-2 2v14a2 2 0 002 2h14a2 2 0 002-2v-7" stroke="currentColor" stroke-width="1.5"/><path d="M18.5 2.5a2.121 2.121 0 113 3L12 15l-4 1 1-4 9.5-9.5z" stroke="currentColor" stroke-width="1.5"/></svg>
              编辑
            </button>
            <button class="card-act-btn" @click="toggleCh(ch)">
              {{ ch.enabled ? '禁用' : '启用' }}
            </button>
            <button class="card-act-btn danger" @click="deleteCh(ch)">删除</button>
          </div>
        </div>
      </div>

      <!-- 空状态 -->
      <div v-if="channels.length === 0 && !showAddPanel && !editingChannel && !loading" class="empty-state">
        <svg viewBox="0 0 24 24" width="40" height="40" fill="none" opacity="0.3">
          <path d="M21 15a2 2 0 01-2 2H7l-4 4V5a2 2 0 012-2h14a2 2 0 012 2z" stroke="currentColor" stroke-width="1.5"/>
        </svg>
        <p>暂未接入任何通道</p>
        <button class="add-btn" @click="showAddPanel = true">
          <svg viewBox="0 0 24 24" width="14" height="14" fill="none">
            <line x1="12" y1="5" x2="12" y2="19" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
            <line x1="5" y1="12" x2="19" y2="12" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
          </svg>
          新增通道
        </button>
      </div>

      <!-- 选择通道类型 -->
      <div v-if="showAddPanel && !selectedType" class="type-select-panel">
        <div class="panel-header">
          <h3>选择通道类型</h3>
          <button class="close-btn" @click="showAddPanel = false">&times;</button>
        </div>
        <div class="type-grid">
          <button
            v-for="t in channelTypes"
            :key="t.key"
            class="type-card"
            :class="{ soon: !t.available }"
            @click="t.available && selectType(t.key)"
          >
            <span class="type-icon" v-html="t.icon"></span>
            <span class="type-name">{{ t.name }}</span>
            <span v-if="!t.available" class="type-soon">即将支持</span>
          </button>
        </div>
      </div>

      <!-- 企业微信配置表单 -->
      <div v-if="(showAddPanel && selectedType === 'wecom-app') || (editingChannel && editingChannel.key === 'wecom-app')" class="config-form-panel">
        <div class="panel-header">
          <h3>{{ editingChannel ? '编辑' : '新增' }}企业微信应用</h3>
          <button class="close-btn" @click="cancelForm">&times;</button>
        </div>

        <n-form :model="wecomForm" label-placement="top" class="channel-form" size="medium">
          <n-form-item label="Corp ID (企业 ID)" required>
            <n-input v-model:value="wecomForm.corpId" placeholder="输入企业 ID（如 wwxxxxxxxxxx）" />
          </n-form-item>
          <n-form-item label="Corp Secret (应用密钥)" required>
            <n-input v-model:value="wecomForm.corpSecret" type="password" show-password-on="click" placeholder="应用的 Secret" />
          </n-form-item>
          <n-form-item label="Agent ID (应用 ID)" required>
            <n-input-number v-model:value="wecomForm.agentId" :min="1" style="width: 100%" placeholder="1000002" />
          </n-form-item>
          <n-form-item label="Token (接收消息Token)" required>
            <n-input v-model:value="wecomForm.token" placeholder="应用回调 Token" />
          </n-form-item>
          <n-form-item label="Encoding AES Key" required>
            <n-input v-model:value="wecomForm.encodingAESKey" placeholder="消息加密密钥" />
          </n-form-item>
        </n-form>

        <div class="form-footer">
          <a class="help-link" href="https://mp.weixin.qq.com/s/vG8BRAzvjVUfgJYaKgfCiw" target="_blank" rel="noopener">不会配置，点这里 →</a>
          <div class="form-actions">
            <n-button quaternary size="small" @click="cancelForm">取消</n-button>
            <n-button type="primary" @click="saveWecom" :loading="saving"
              :disabled="!wecomForm.corpId || !wecomForm.corpSecret || !wecomForm.agentId || !wecomForm.token || !wecomForm.encodingAESKey"
            >保存</n-button>
          </div>
        </div>
      </div>

      <!-- QQ 机器人配置表单 -->
      <div v-if="(showAddPanel && selectedType === 'qqbot') || (editingChannel && editingChannel.key === 'qqbot')" class="config-form-panel">
        <div class="panel-header">
          <h3>{{ editingChannel ? '编辑' : '新增' }} QQ 机器人</h3>
          <button class="close-btn" @click="cancelForm">&times;</button>
        </div>

        <n-form :model="qqForm" label-placement="top" class="channel-form" size="medium">
          <n-form-item label="App ID (机器人 AppID)" required>
            <n-input v-model:value="qqForm.appId" placeholder="你的 AppID" />
          </n-form-item>
          <n-form-item label="Client Secret (AppSecret)" required>
            <n-input v-model:value="qqForm.clientSecret" type="password" show-password-on="click" placeholder="你的 AppSecret" />
          </n-form-item>
        </n-form>

        <div class="form-footer">
          <a class="help-link" href="https://mp.weixin.qq.com/s/vG8BRAzvjVUfgJYaKgfCiw" target="_blank" rel="noopener">不会配置，点这里 →</a>
          <div class="form-actions">
            <n-button quaternary size="small" @click="cancelForm">取消</n-button>
            <n-button type="primary" @click="saveQQBot" :loading="saving"
              :disabled="!qqForm.appId || !qqForm.clientSecret"
            >保存</n-button>
          </div>
        </div>
      </div>

      <!-- 钉钉配置表单 -->
      <div v-if="(showAddPanel && selectedType === 'dingtalk') || (editingChannel && editingChannel.key === 'dingtalk')" class="config-form-panel">
        <div class="panel-header">
          <h3>{{ editingChannel ? '编辑' : '新增' }}钉钉应用</h3>
          <button class="close-btn" @click="cancelForm">&times;</button>
        </div>

        <n-form :model="dingtalkForm" label-placement="top" class="channel-form" size="medium">
          <n-form-item label="AgentID" required>
            <n-input v-model:value="dingtalkForm.agentId" placeholder="123456789" />
          </n-form-item>
          <n-form-item label="Client ID (AppKey)" required>
            <n-input v-model:value="dingtalkForm.clientId" placeholder="dingxxxxxx" />
          </n-form-item>
          <n-form-item label="Client Secret (AppSecret)" required>
            <n-input v-model:value="dingtalkForm.clientSecret" type="password" show-password-on="click" placeholder="应用 AppSecret" />
          </n-form-item>
          <n-form-item label="Robot Code (与 Client ID 相同)" required>
            <n-input v-model:value="dingtalkForm.robotCode" placeholder="dingxxxxxx" />
          </n-form-item>
          <n-form-item label="Corp ID (企业 ID)" required>
            <n-input v-model:value="dingtalkForm.corpId" placeholder="dingxxxxxx" />
          </n-form-item>
        </n-form>

        <div class="form-footer">
          <a class="help-link" href="https://mp.weixin.qq.com/s/vG8BRAzvjVUfgJYaKgfCiw" target="_blank" rel="noopener">不会配置，点这里 →</a>
          <div class="form-actions">
            <n-button quaternary size="small" @click="cancelForm">取消</n-button>
            <n-button type="primary" @click="saveDingTalk" :loading="saving"
              :disabled="!dingtalkForm.clientId || !dingtalkForm.clientSecret || !dingtalkForm.robotCode || !dingtalkForm.corpId || !dingtalkForm.agentId"
            >保存</n-button>
          </div>
        </div>
      </div>

      <!-- 飞书配置表单 -->
      <div v-if="(showAddPanel && selectedType === 'feishu') || (editingChannel && editingChannel.key === 'feishu')" class="config-form-panel">
        <div class="panel-header">
          <h3>{{ editingChannel ? '编辑' : '新增' }}飞书应用</h3>
          <button class="close-btn" @click="cancelForm">&times;</button>
        </div>

        <n-form :model="feishuForm" label-placement="top" class="channel-form" size="medium">
          <n-form-item label="App ID (应用 ID)" required>
            <n-input v-model:value="feishuForm.appId" placeholder="cli_xxx" />
          </n-form-item>
          <n-form-item label="App Secret (应用密钥)" required>
            <n-input v-model:value="feishuForm.appSecret" type="password" show-password-on="click" placeholder="应用 Secret" />
          </n-form-item>
          <n-form-item label="Bot Name (机器人名称)" required>
            <n-input v-model:value="feishuForm.botName" placeholder="我的AI助手" />
          </n-form-item>
        </n-form>

        <div class="form-footer">
          <a class="help-link" href="https://mp.weixin.qq.com/s/vG8BRAzvjVUfgJYaKgfCiw" target="_blank" rel="noopener">不会配置，点这里 →</a>
          <div class="form-actions">
            <n-button quaternary size="small" @click="cancelForm">取消</n-button>
            <n-button type="primary" @click="saveFeishu" :loading="saving"
              :disabled="!feishuForm.appId || !feishuForm.appSecret || !feishuForm.botName"
            >保存</n-button>
          </div>
        </div>
      </div>

      <!-- 加载 -->
      <div v-if="loading" class="loading-state">
        <div class="loading-spinner"></div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { NForm, NFormItem, NInput, NInputNumber, NButton, NTooltip } from 'naive-ui'
import { getChannels, saveChannel, deleteChannel, toggleChannel } from '@/api/channel'
import gm from '@/utils/gmssh'
import cache from '@/stores/cache'

const loading = ref(true)
const saving = ref(false)
const channels = ref([])
const showAddPanel = ref(false)
const selectedType = ref(null)
const editingChannel = ref(null)

const wecomForm = ref({
  corpId: '',
  corpSecret: '',
  agentId: null,
  token: '',
  encodingAESKey: '',
})

const qqForm = ref({
  appId: '',
  clientSecret: '',
})

const dingtalkForm = ref({
  clientId: '',
  clientSecret: '',
  robotCode: '',
  corpId: '',
  agentId: '',
})

const feishuForm = ref({
  appId: '',
  appSecret: '',
  botName: '',
})

const channelTypes = [
  {
    key: 'wecom-app',
    name: '企业微信',
    available: true,
    icon: '<svg viewBox="0 0 24 24" width="24" height="24" fill="none"><rect x="3" y="3" width="18" height="18" rx="3" stroke="currentColor" stroke-width="1.3"/><path d="M8 10.5a1.5 1.5 0 100-3 1.5 1.5 0 000 3zM14 10.5a1.5 1.5 0 100-3 1.5 1.5 0 000 3zM7 14.5s1.5 2 5 2 5-2 5-2" stroke="currentColor" stroke-width="1.3" stroke-linecap="round"/></svg>',
  },
  {
    key: 'qqbot',
    name: 'QQ 机器人',
    available: true,
    icon: '<svg viewBox="0 0 24 24" width="24" height="24" fill="none"><circle cx="12" cy="12" r="9" stroke="currentColor" stroke-width="1.3"/><circle cx="9" cy="10" r="1.5" fill="currentColor"/><circle cx="15" cy="10" r="1.5" fill="currentColor"/><path d="M8 15s2 2 4 2 4-2 4-2" stroke="currentColor" stroke-width="1.3" stroke-linecap="round"/></svg>',
  },
  {
    key: 'dingtalk',
    name: '钉钉',
    available: true,
    icon: '<svg viewBox="0 0 24 24" width="24" height="24" fill="none"><path d="M12 2L2 7l10 5 10-5-10-5z" stroke="currentColor" stroke-width="1.3" stroke-linejoin="round"/><path d="M2 17l10 5 10-5M2 12l10 5 10-5" stroke="currentColor" stroke-width="1.3" stroke-linejoin="round"/></svg>',
  },
  {
    key: 'feishu',
    name: '飞书',
    available: true,
    icon: '<svg viewBox="0 0 24 24" width="24" height="24" fill="none"><path d="M4 4l16 8-16 8V4z" stroke="currentColor" stroke-width="1.3" stroke-linejoin="round"/></svg>',
  },
  {
    key: 'whatsapp',
    name: 'WhatsApp',
    available: false,
    icon: '<svg viewBox="0 0 24 24" width="24" height="24" fill="none"><path d="M21 11.5a8.38 8.38 0 01-.9 3.8 8.5 8.5 0 01-7.6 4.7 8.38 8.38 0 01-3.8-.9L3 21l1.9-5.7a8.38 8.38 0 01-.9-3.8 8.5 8.5 0 014.7-7.6 8.38 8.38 0 013.8-.9h.5a8.48 8.48 0 018 8v.5z" stroke="currentColor" stroke-width="1.3"/></svg>',
  },
  {
    key: 'telegram',
    name: 'Telegram',
    available: false,
    icon: '<svg viewBox="0 0 24 24" width="24" height="24" fill="none"><path d="M22 2L11 13M22 2l-7 20-4-9-9-4 20-7z" stroke="currentColor" stroke-width="1.3" stroke-linejoin="round"/></svg>',
  },
  {
    key: 'imessage',
    name: 'iMessage',
    available: false,
    icon: '<svg viewBox="0 0 24 24" width="24" height="24" fill="none"><path d="M21 15a2 2 0 01-2 2H7l-4 4V5a2 2 0 012-2h14a2 2 0 012 2z" stroke="currentColor" stroke-width="1.3"/></svg>',
  },
]

function getChannelDisplayName(key) {
  const t = channelTypes.find(c => c.key === key)
  return t ? t.name : key
}

function getChannelIcon(key) {
  const t = channelTypes.find(c => c.key === key)
  return t ? t.icon : ''
}

// 卡片字段显示
const revealedFields = ref({})

function getChannelFields(ch) {
  if (ch.key === 'wecom-app') {
    return [
      { label: 'Corp ID', value: String(ch.corpId || ''), masked: false },
      { label: 'Agent ID', value: String(ch.agentId || ''), masked: false },
      { label: 'Corp Secret', value: String(ch.corpSecret || ''), masked: true },
      { label: 'Token', value: String(ch.token || ''), masked: true },
      { label: 'AES Key', value: String(ch.encodingAESKey || ''), masked: true },
    ].filter(f => f.value)
  }
  if (ch.key === 'qqbot') {
    return [
      { label: 'App ID', value: String(ch.appId || ''), masked: false },
      { label: 'Client Secret', value: String(ch.clientSecret || ''), masked: true },
    ].filter(f => f.value)
  }
  if (ch.key === 'dingtalk') {
    return [
      { label: 'Client ID', value: String(ch.clientId || ''), masked: false },
      { label: 'Robot Code', value: String(ch.robotCode || ''), masked: false },
      { label: 'Corp ID', value: String(ch.corpId || ''), masked: false },
      { label: 'Agent ID', value: String(ch.agentId || ''), masked: false },
      { label: 'Client Secret', value: String(ch.clientSecret || ''), masked: true },
    ].filter(f => f.value)
  }
  if (ch.key === 'feishu') {
    const acc = ch.accounts?.main || {}
    return [
      { label: 'App ID', value: String(acc.appId || ''), masked: false },
      { label: 'Bot Name', value: String(acc.botName || ''), masked: false },
      { label: 'App Secret', value: String(acc.appSecret || ''), masked: true },
    ].filter(f => f.value)
  }
  return []
}

function maskValue(val) {
  if (!val || val.length <= 4) return '****'
  return val.slice(0, 3) + '****' + val.slice(-3)
}

function displayValue(field, channelKey) {
  if (field.masked && !revealedFields.value[channelKey + field.label]) {
    return maskValue(field.value)
  }
  return field.value
}

function toggleReveal(key) {
  revealedFields.value[key] = !revealedFields.value[key]
}

function selectType(key) {
  selectedType.value = key
  resetForms()
}

function resetForms() {
  wecomForm.value = { corpId: '', corpSecret: '', agentId: null, token: '', encodingAESKey: '' }
  qqForm.value = { appId: '', clientSecret: '' }
  dingtalkForm.value = { clientId: '', clientSecret: '', robotCode: '', corpId: '', agentId: '' }
  feishuForm.value = { appId: '', appSecret: '', botName: '' }
}

function cancelForm() {
  showAddPanel.value = false
  selectedType.value = null
  editingChannel.value = null
  resetForms()
}

function editChannel(ch) {
  editingChannel.value = ch
  if (ch.key === 'wecom-app') {
    wecomForm.value = {
      corpId: ch.corpId || '',
      corpSecret: ch.corpSecret || '',
      agentId: ch.agentId || null,
      token: ch.token || '',
      encodingAESKey: ch.encodingAESKey || '',
    }
  } else if (ch.key === 'dingtalk') {
    dingtalkForm.value = {
      clientId: ch.clientId || '',
      clientSecret: ch.clientSecret || '',
      robotCode: ch.robotCode || '',
      corpId: ch.corpId || '',
      agentId: String(ch.agentId || ''),
    }
  } else if (ch.key === 'qqbot') {
    qqForm.value = {
      appId: ch.appId || '',
      clientSecret: ch.clientSecret || '',
    }
  } else if (ch.key === 'feishu') {
    const acc = ch.accounts?.main || {}
    feishuForm.value = {
      appId: acc.appId || '',
      appSecret: acc.appSecret || '',
      botName: acc.botName || '',
    }
  }
}

async function fetchChannels() {
  loading.value = true
  try {
    const res = await getChannels()
    channels.value = res?.channels || []
    cache.channels = [...channels.value]
  } catch (e) {
    gm.error('获取通道失败: ' + (e.message || ''))
  } finally {
    loading.value = false
  }
}

async function saveWecom() {
  saving.value = true
  try {
    await saveChannel({
      channelKey: 'wecom-app',
      enabled: editingChannel.value ? (editingChannel.value.enabled !== false) : true,
      corpId: wecomForm.value.corpId,
      corpSecret: wecomForm.value.corpSecret,
      agentId: wecomForm.value.agentId,
      token: wecomForm.value.token,
      encodingAESKey: wecomForm.value.encodingAESKey,
      dmPolicy: 'pairing',
      groupPolicy: 'open',
    })
    gm.success('通道配置已保存')
    cancelForm()
    await fetchChannels()
  } catch (e) {
    gm.error('保存失败: ' + (e.message || ''))
  } finally {
    saving.value = false
  }
}

async function saveQQBot() {
  saving.value = true
  try {
    await saveChannel({
      channelKey: 'qqbot',
      enabled: editingChannel.value ? (editingChannel.value.enabled !== false) : true,
      appId: qqForm.value.appId,
      clientSecret: qqForm.value.clientSecret,
    })
    gm.success('QQ 机器人通道已保存')
    cancelForm()
    await fetchChannels()
  } catch (e) {
    gm.error('保存失败: ' + (e.message || ''))
  } finally {
    saving.value = false
  }
}

async function saveDingTalk() {
  saving.value = true
  try {
    await saveChannel({
      channelKey: 'dingtalk',
      enabled: editingChannel.value ? (editingChannel.value.enabled !== false) : true,
      clientId: dingtalkForm.value.clientId,
      clientSecret: dingtalkForm.value.clientSecret,
      robotCode: dingtalkForm.value.robotCode,
      corpId: dingtalkForm.value.corpId,
      agentId: dingtalkForm.value.agentId,
      dmPolicy: 'open',
      groupPolicy: 'open',
      debug: false,
      messageType: 'markdown',
    })
    gm.success('钉钉通道已保存')
    cancelForm()
    await fetchChannels()
  } catch (e) {
    gm.error('保存失败: ' + (e.message || ''))
  } finally {
    saving.value = false
  }
}

async function saveFeishu() {
  saving.value = true
  try {
    await saveChannel({
      channelKey: 'feishu',
      enabled: editingChannel.value ? (editingChannel.value.enabled !== false) : true,
      dmPolicy: 'open',
      accounts: {
        main: {
          appId: feishuForm.value.appId,
          appSecret: feishuForm.value.appSecret,
          botName: feishuForm.value.botName,
        },
      },
    })
    gm.success('飞书通道已保存')
    cancelForm()
    await fetchChannels()
  } catch (e) {
    gm.error('保存失败: ' + (e.message || ''))
  } finally {
    saving.value = false
  }
}

async function deleteCh(ch) {
  const gmApi = gm.getGmApi()
  const doDelete = async () => {
    try {
      await deleteChannel({ channelKey: ch.key })
      gm.success('通道已删除')
      await fetchChannels()
    } catch (e) {
      gm.error('删除失败: ' + (e.message || ''))
    }
  }
  if (gmApi?.dialog) {
    gmApi.dialog.warning({
      title: '删除通道',
      content: `确定删除「${getChannelDisplayName(ch.key)}」通道吗？配置将从 openclaw.json 中移除。`,
      positiveText: '确定',
      negativeText: '取消',
      onPositiveClick: doDelete,
    })
  } else {
    if (confirm(`确定删除「${getChannelDisplayName(ch.key)}」？`)) doDelete()
  }
}

async function toggleCh(ch) {
  try {
    await toggleChannel({ channelKey: ch.key, enabled: !ch.enabled })
    gm.success(ch.enabled ? '通道已禁用' : '通道已启用')
    await fetchChannels()
  } catch (e) {
    gm.error('操作失败: ' + (e.message || ''))
  }
}

onMounted(() => {
  if (cache.channels !== null) {
    channels.value = [...cache.channels]
    loading.value = false
    return
  }
  fetchChannels()
})
</script>

<style scoped>
.channels-page {
  width: 100%;
  height: 100%;
  overflow-y: auto;
  padding: 20px;
}
.channels-container {
  max-width: 720px;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

/* 顶部 */
.channels-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
}
.header-left { display: flex; flex-direction: column; gap: 4px; }
.page-title {
  display: flex; align-items: center; gap: 8px;
  font-size: 18px; font-weight: 600; color: var(--jm-accent-7); margin: 0;
}
.header-hint { font-size: 12px; color: var(--jm-accent-4); padding-left: 28px; }

.add-btn {
  display: flex; align-items: center; gap: 6px;
  padding: 8px 16px; border-radius: 8px;
  border: 1px solid var(--jm-primary-2); background: rgba(var(--jm-primary-1-rgb), 0.06);
  color: var(--jm-primary-2); font-size: 13px; font-weight: 500; cursor: pointer;
  transition: all 0.2s;
}
.add-btn:hover { background: rgba(var(--jm-primary-1-rgb), 0.15); }

/* 通道卡片 */
.channel-cards { display: flex; flex-direction: column; gap: 12px; }
.channel-card {
  border: 1px solid var(--jm-accent-2); border-radius: 10px;
  background: rgba(var(--jm-accent-1-rgb), 0.4); padding: 16px;
  transition: border-color 0.2s;
}
.channel-card:hover { border-color: var(--jm-accent-3); }

.card-header { display: flex; align-items: center; gap: 12px; margin-bottom: 12px; }
.card-icon { color: var(--jm-primary-1); display: flex; align-items: center; }
.card-info { flex: 1; display: flex; flex-direction: column; }
.card-name { font-size: 14px; font-weight: 600; color: var(--jm-accent-7); }
.card-key { font-size: 11px; color: var(--jm-accent-4); font-family: monospace; }
.card-status {
  font-size: 11px; padding: 2px 8px; border-radius: 4px; font-weight: 500;
}
.card-status.enabled { color: var(--jm-success-color); background: rgba(34,197,94,0.08); }
.card-status.disabled { color: var(--jm-accent-4); background: rgba(var(--jm-accent-1-rgb), 0.5); }

.card-details {
  display: flex; gap: 20px; padding: 10px 0; margin-bottom: 10px;
  border-top: 1px solid var(--jm-accent-2);
}
.detail-item { display: flex; flex-direction: column; gap: 2px; min-width: 0; flex: 1; max-width: 140px; }
.detail-label { font-size: 10px; color: var(--jm-accent-4); text-transform: uppercase; letter-spacing: 0.5px; }
.detail-value {
  font-size: 12px; color: var(--jm-accent-6); font-family: monospace;
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap; max-width: 120px; display: inline-block;
}
.detail-value-wrap {
  display: flex; align-items: center; gap: 6px;
}
.eye-btn {
  background: none; border: none; padding: 2px; cursor: pointer;
  color: var(--jm-accent-4); display: flex; align-items: center;
  transition: color 0.15s;
}
.eye-btn:hover { color: var(--jm-accent-6); }

.card-actions { display: flex; gap: 8px; }
.card-act-btn {
  padding: 4px 12px; border-radius: 5px; border: 1px solid var(--jm-accent-2);
  background: transparent; color: var(--jm-accent-5); font-size: 11px; cursor: pointer;
  transition: all 0.15s;
}
.card-act-btn:hover { border-color: var(--jm-accent-3); color: var(--jm-accent-6); }
.card-act-btn.danger:hover { border-color: var(--jm-error-color); color: var(--jm-error-color); }

/* 空状态 */
.empty-state {
  display: flex; flex-direction: column; align-items: center; justify-content: center;
  gap: 12px; padding: 60px 0; color: var(--jm-accent-4);
}
.empty-state p { font-size: 13px; margin: 0; }

/* 选择通道类型 */
.type-select-panel, .config-form-panel {
  border: 1px solid var(--jm-accent-2); border-radius: 10px;
  background: rgba(var(--jm-accent-1-rgb), 0.4); padding: 20px;
}
.panel-header {
  display: flex; align-items: center; justify-content: space-between; margin-bottom: 16px;
}
.panel-header h3 { margin: 0; font-size: 15px; font-weight: 600; color: var(--jm-accent-7); }
.close-btn {
  width: 28px; height: 28px; border-radius: 6px; border: none;
  background: transparent; color: var(--jm-accent-4); font-size: 18px;
  cursor: pointer; display: flex; align-items: center; justify-content: center;
  transition: all 0.15s;
}
.close-btn:hover { background: rgba(var(--jm-accent-1-rgb), 0.6); color: var(--jm-accent-6); }

.type-grid {
  display: grid; grid-template-columns: repeat(auto-fill, minmax(120px, 1fr)); gap: 10px;
}
.type-card {
  display: flex; flex-direction: column; align-items: center; gap: 8px;
  padding: 16px 10px; border-radius: 8px; border: 1px solid var(--jm-accent-2);
  background: rgba(var(--jm-accent-1-rgb), 0.3); color: var(--jm-accent-6);
  cursor: pointer; transition: all 0.2s; position: relative;
}
.type-card:hover:not(.soon) {
  border-color: var(--jm-primary-2); color: var(--jm-primary-1);
  background: rgba(var(--jm-primary-1-rgb), 0.06);
}
.type-card.soon { opacity: 0.4; cursor: not-allowed; }
.type-icon { display: flex; align-items: center; }
.type-name { font-size: 12px; font-weight: 500; }
.type-soon { font-size: 10px; color: var(--jm-accent-4); }

/* 配置表单 */
.channel-form {
  background: rgba(var(--jm-accent-1-rgb), 0.3); border: 1px solid var(--jm-accent-2);
  border-radius: 8px; padding: 18px 18px 6px;
}
.form-footer {
  display: flex; align-items: center; justify-content: space-between; margin-top: 12px;
}
.form-actions { display: flex; gap: 8px; }
.help-link {
  font-size: 12px; color: var(--jm-primary-2); text-decoration: none;
}
.help-link:hover { color: var(--jm-primary-1); text-decoration: underline; }

/* 加载 */
.loading-state {
  display: flex; justify-content: center; padding: 40px;
}
.loading-spinner {
  width: 24px; height: 24px;
  border: 2px solid var(--jm-accent-2); border-top-color: var(--jm-primary-1);
  border-radius: 50%; animation: spin 0.8s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }
</style>
