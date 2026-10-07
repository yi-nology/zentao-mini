/**
 * 通知偏好 composable(纯网页版)
 * 1. localStorage 记录用户是否启用应用内通知(ElNotification)
 * 2. 浏览器 Notification API 的权限申请与状态查询
 *
 * 注:原 Wails 桌面事件推送(@wailsio/runtime Events)已随桌面版移除
 */
const STORAGE_KEY = 'zentao-mini-notification'

export function isNotificationEnabled(): boolean {
  return localStorage.getItem(STORAGE_KEY) !== '0'
}

export function setNotificationEnabled(enabled: boolean): void {
  localStorage.setItem(STORAGE_KEY, enabled ? '1' : '0')
}

/** 请求 Notification API 权限(如果尚未授权),返回是否已授权 */
export async function ensurePermission(): Promise<boolean> {
  if (typeof Notification === 'undefined') return false
  if (Notification.permission === 'granted') return true
  if (Notification.permission === 'denied') return false
  try {
    const result = await Notification.requestPermission()
    return result === 'granted'
  } catch {
    return false
  }
}
