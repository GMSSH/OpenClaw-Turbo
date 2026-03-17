import gm from '@/utils/gmssh'

// ========== 原有 Agent 人格文件接口 ==========

/**
 * 获取所有 Agent 人格文件
 */
export function getAgentFiles() {
    return gm.request('getAgentFiles')
}

/**
 * 保存单个 Agent 文件
 */
export function saveAgentFile(params) {
    return gm.request('saveAgentFile', params)
}

/**
 * 重置 Agent 文件为默认内容
 */
export function resetAgentFile(params) {
    return gm.request('resetAgentFile', params)
}

/**
 * 获取预设模板列表
 */
export function getAgentTemplates() {
    return gm.request('getAgentTemplates')
}

/**
 * 应用预设模板
 */
export function applyAgentTemplate(params) {
    return gm.request('applyAgentTemplate', params)
}

// ========== 多 Agent 管理接口 ==========

/**
 * 获取 Agent 列表
 */
export function listAgents() {
    return gm.request('listAgents')
}

/**
 * 创建 Agent
 */
export function createAgent(params) {
    return gm.request('createAgent', params)
}

/**
 * 更新 Agent
 */
export function updateAgent(params) {
    return gm.request('updateAgent', params)
}

/**
 * 删除 Agent
 */
export function deleteAgent(params) {
    return gm.request('deleteAgent', params)
}

/**
 * 获取 Agent 详情（含人格文件）
 */
export function getAgentDetail(params) {
    return gm.request('getAgentDetail', params)
}

/**
 * 获取已配置的对话模型列表
 */
export function getConfiguredModels() {
    return gm.request('getConfiguredModels')
}

/**
 * 获取 Agent 实时活跃状态 + 子 Agent（临时工）
 */
export function getAgentActivity() {
    return gm.request('getAgentActivity')
}
