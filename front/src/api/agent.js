import gm from '@/utils/gmssh'

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
