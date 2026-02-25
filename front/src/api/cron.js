import gm from '@/utils/gmssh'

/** 获取调度器状态 */
export function cronStatus() {
    return gm.request('cronStatus')
}

/** 列出所有定时任务 */
export function listCronJobs() {
    return gm.request('listCronJobs')
}

/** 新增定时任务 */
export function addCronJob(params) {
    return gm.request('addCronJob', params)
}

/** 编辑定时任务 */
export function editCronJob(params) {
    return gm.request('editCronJob', params)
}

/** 删除定时任务 */
export function removeCronJob(params) {
    return gm.request('removeCronJob', params)
}

/** 启用定时任务 */
export function enableCronJob(params) {
    return gm.request('enableCronJob', params)
}

/** 禁用定时任务 */
export function disableCronJob(params) {
    return gm.request('disableCronJob', params)
}

/** 手动执行定时任务 */
export function runCronJob(params) {
    return gm.request('runCronJob', params)
}

/** 获取运行历史 */
export function getCronRuns(params) {
    return gm.request('getCronRuns', params)
}
