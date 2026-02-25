/**
 * GMSSH SDK 封装
 * 通过 $gm.request 与后端通信，$gm 由 gmAppSdk.js 从宿主注入
 */
class GMSSHUtils {
    constructor() {
        // RPC 基础路径: /api/call/{org}/{app}
        this.baseURL = '/api/call/guanxi/eazyClaw'
    }

    /**
     * 获取公网 IP
     */
    getPublicIp() {
        return window.$gm?.publicIp || 'localhost'
    }

    // ====== UI 反馈 ======
    success(message) {
        if (window.$gm?.message) window.$gm.message.success(message)
        else console.log('[SUCCESS]', message)
    }

    error(message) {
        if (window.$gm?.message) window.$gm.message.error(message)
        else console.error('[ERROR]', message)
    }

    warning(message) {
        if (window.$gm?.message) window.$gm.message.warning(message)
        else console.warn('[WARNING]', message)
    }

    info(message) {
        if (window.$gm?.message) window.$gm.message.info(message)
        else console.info('[INFO]', message)
    }

    // ====== 后端 RPC 请求 ======
    /**
     * 发送 JSON-RPC 请求到后端
     * 通过 $gm.request 走 GMSSH 网关 -> Unix Socket -> 后端
     *
     * 响应结构:
     * { code: 200000, data: { code: 200, data: {...}, msg: "..." }, msg: "..." }
     */
    async request(method, params = {}) {
        if (!window.$gm?.request) {
            throw new Error('$gm.request 不可用，请确认应用运行在 GMSSH 环境中')
        }

        const response = await window.$gm.request({
            url: `${this.baseURL}/${method}`,
            method: 'POST',
            timeout: 12000,
            data: { params },
        })

        // 校验网关层
        if (!response || response.code !== 200000) {
            throw new Error(response?.msg || 'Gateway Error')
        }

        // 校验业务层
        const rpcResult = response.data
        if (!rpcResult || rpcResult.code !== 200) {
            throw new Error(rpcResult?.msg || 'RPC Error')
        }

        return rpcResult.data
    }

    // ====== 系统操作 ======
    getGmApi() {
        return window.$gm || null
    }

    closeApp() {
        if (window.$gm?.closeApp) window.$gm.closeApp()
    }
}

export default new GMSSHUtils()
