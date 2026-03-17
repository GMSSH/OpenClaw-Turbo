import gm from '@/utils/gmssh'

/**
 * 获取所有 Agent 的实时状态（working/online/idle/offline）
 * 通过扫描 ~/.openclaw/agents/{id}/sessions/*.jsonl 实现
 */
export function getAgentStatuses() {
    return gm.request('getAgentStatuses')
}

/**
 * 获取 Agent Token 消耗统计
 * @param {string} agentId - 指定 Agent ID，空字符串 = 所有 Agent 聚合
 * @param {number} days - 统计天数（默认 7）
 */
export function getAgentTokenStats(params = {}) {
    return gm.request('getAgentTokenStats', params)
}

/**
 * 探测 Gateway 健康状态
 * 通过 HTTP GET ~/.openclaw/gateway -> /api/health 实现
 */
export function getGatewayHealth() {
    return gm.request('getGatewayHealth')
}

/**
 * 测试平台连通性（Feishu / Discord / Telegram）
 * @param {string} agentId - 指定 Agent，空 = 所有
 * @param {string} platform - 'feishu' | 'discord' | 'telegram' | '' = 全部
 */
export function testPlatformConn(params = {}) {
    return gm.request('testPlatformConn', params)
}

/**
 * 获取指定 Agent 的会话列表
 * @param {string} agentId - Agent ID（必填）
 */
export function getAgentSessions(params) {
    return gm.request('getAgentSessions', params)
}
