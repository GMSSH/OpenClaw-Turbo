import gm from '@/utils/gmssh'

/**
 * 检测 Docker 环境
 */
export function checkEnvironment() {
    return gm.request('checkEnvironment')
}

/**
 * 生成随机 Token
 */
export function generateToken() {
    return gm.request('generateToken')
}

/**
 * 执行部署
 */
export function deploy(config) {
    return gm.request('deploy', config)
}

/**
 * 获取部署日志
 */
export function getDeployLogs() {
    return gm.request('getDeployLogs')
}

/**
 * 获取 OpenClaw 运行状态
 */
export function getClawStatus() {
    return gm.request('getClawStatus')
}

/**
 * 获取 OpenClaw 配置信息
 */
export function getClawConfig() {
    return gm.request('getClawConfig')
}

/**
 * 检测端口是否被占用
 */
export function checkPorts(ports) {
    return gm.request('checkPorts', { ports })
}

/**
 * 停止 OpenClaw 容器
 */
export function stopClaw() {
    return gm.request('stopClaw')
}

/**
 * 重启 OpenClaw 容器
 */
export function restartClaw() {
    return gm.request('restartClaw')
}

/**
 * 卸载 OpenClaw (容器 + 镜像 + 数据)
 */
export function uninstallClaw() {
    return gm.request('uninstallClaw')
}

/**
 * 测试 AI API 连通性
 */
export function testApiConnection(params) {
    return gm.request('testApiConnection', params)
}

/**
 * 切换AI模型配置
 */
export function updateModelConfig(params) {
    return gm.request('updateModelConfig', params)
}

/**
 * 更新记忆配置
 */
export function updateMemoryConfig(params) {
    return gm.request('updateMemoryConfig', params)
}

/**
 * 安装 Node.js 环境
 */
export function installNodeEnv() {
    return gm.request('installNodeEnv')
}

/**
 * 大模型提供商配置
 */
export const MODEL_PROVIDERS = [
    {
        provider: 'deepseek',
        displayName: 'DeepSeek (国产之光 - 强力推荐)',
        apiMode: 'openai',  // OpenAI Chat Completions 协议
        baseUrl: 'https://api.deepseek.com/v1',
        tokenUrl: 'https://platform.deepseek.com/api_keys',
        models: [
            { id: 'deepseek/deepseek-chat', name: 'DeepSeek V3 (日常对话 - 极速且聪明)' },
            { id: 'deepseek/deepseek-reasoner', name: 'DeepSeek R1 (深度思考 - 逻辑最强)' },
        ],
    },
    {
        provider: 'openai',
        displayName: 'OpenAI (ChatGPT 官方)',
        apiMode: 'openai',
        baseUrl: 'https://api.openai.com/v1',
        tokenUrl: 'https://platform.openai.com/api-keys',
        models: [
            { id: 'openai/gpt-5', name: 'GPT-5 (2026 年度旗舰 - 最全能)' },
            { id: 'openai/gpt-5-mini', name: 'GPT-5 Mini (智能且性价比极高)' },
            { id: 'openai/o3', name: 'OpenAI o3 (数学与编程专家)' },
            { id: 'openai/gpt-4o', name: 'GPT-4o (经典旗舰模型)' },
        ],
    },
    {
        provider: 'alibaba',
        displayName: '阿里云通义千问 (Qwen)',
        apiMode: 'openai',
        baseUrl: 'https://dashscope.aliyuncs.com/compatible-mode/v1',
        tokenUrl: 'https://dashscope.console.aliyun.com/apiKey',
        models: [
            { id: 'alibaba/qwen-max', name: '通义千问 Max (旗舰版 - 综合实力最强)' },
            { id: 'alibaba/qwen-plus', name: '通义千问 Plus (增强版 - 复杂任务推荐)' },
            { id: 'alibaba/qwen-turbo', name: '通义千问 Turbo (极速版 - 秒回且便宜)' },
            { id: 'alibaba/qwen-long', name: '通义千问 Long (超长文档分析必备)' },
        ],
    },
    {
        provider: 'anthropic',
        displayName: 'Anthropic (Claude 系列)',
        apiMode: 'anthropic',  // Anthropic Messages 协议
        baseUrl: 'https://api.anthropic.com',
        tokenUrl: 'https://console.anthropic.com/settings/keys',
        models: [
            { id: 'anthropic/claude-3-7-sonnet-20250219', name: 'Claude 3.7 Sonnet (写代码/改 Bug 首选)' },
            { id: 'anthropic/claude-opus-4-1', name: 'Claude 4. Opus (文学创作/复杂逻辑专家)' },
            { id: 'anthropic/claude-3-5-haiku-latest', name: 'Claude 3.5 Haiku (快如闪电)' },
        ],
    },
    {
        provider: 'gemini',
        displayName: 'Google Gemini (谷歌系列)',
        apiMode: 'gemini',
        baseUrl: 'https://generativelanguage.googleapis.com',
        tokenUrl: 'https://aistudio.google.com/apikey',
        models: [
            { id: 'google/gemini-2.5-pro', name: 'Gemini 2.5 Pro (超大文件读取专家)' },
            { id: 'google/gemini-2.5-flash', name: 'Gemini 2.5 Flash (谷歌最快模型)' },
        ],
    },
    {
        provider: 'kimi',
        displayName: 'Kimi (月之暗面)',
        apiMode: 'openai',
        baseUrl: 'https://api.moonshot.cn/v1',
        tokenUrl: 'https://platform.moonshot.cn/console/api-keys',
        models: [
            { id: 'kimi/kimi-k2.5', name: 'Kimi K2.5 (长文本阅读/搜索全能)' },
            { id: 'kimi/kimi-k2-thinking', name: 'Kimi K2 Thinking (国产深度推理版)' },
        ],
    },
    {
        provider: 'minimax',
        displayName: 'MiniMax (海螺 AI)',
        apiMode: 'anthropic',  // MiniMax 使用 Anthropic Messages 协议
        baseUrl: 'https://api.minimaxi.com/anthropic',
        tokenUrl: 'https://platform.minimaxi.com/user-center/basic-information/interface-key',
        models: [
            { id: 'minimax/MiniMax-M2.5', name: 'MiniMax M2.5 (2025最新旗舰)' },
            { id: 'minimax/MiniMax-M2.1', name: 'MiniMax M2.1 (擅长长篇对话)' },
        ],
    },
    {
        provider: 'ollama',
        displayName: 'Ollama (本地/私有化部署)',
        apiMode: 'openai',
        baseUrl: 'http://localhost:11434/v1',
        tokenUrl: '',
        models: [
            { id: 'ollama/llama3.3', name: 'Llama 3.3 (本地运行 - 无需联网)' },
            { id: 'ollama/deepseek-r1:7b', name: 'DeepSeek R1 7B (本地思考版 - 安全私密)' },
            { id: 'ollama/qwen2.5:14b', name: 'Qwen 2.5 14B (本地强力中文模型)' },
        ],
    },
    {
        provider: 'custom',
        displayName: '自定义 OpenAI 兼容接口',
        apiMode: 'openai',
        baseUrl: '',
        tokenUrl: '',
        models: [],
    },
]
