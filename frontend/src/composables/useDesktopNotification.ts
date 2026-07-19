/**
 * 桌面通知 composable
 * 1. 监听 Wails runtime 事件 'notification'（后端 EventBus 推送的定时任务结果）
 * 2. 通过 Notification API 显示桌面通知（需要用户授权）
 * 3. 同时通过 Element Plus 的 ElNotification 在应用内显示（不依赖授权）
 *
 * 仅在 Wails 桌面环境下生效（其他模式后端不会发 EventsEmit）
 */
import { onMounted, onBeforeUnmount } from 'vue'
import { ElNotification } from 'element-plus'
import * as runtime from '@wailsjs/runtime/runtime'

const STORAGE_KEY = 'zentao-mini-notification'

export interface NotificationPayload {
  type?: string
  level?: 'info' | 'warning' | 'error'
  title: string
  body: string
  taskId?: string
  taskName?: string
  status?: string
  bugTotal?: number
}

export function isNotificationEnabled(): boolean {
  return localStorage.getItem(STORAGE_KEY) !== '0'
}

export function setNotificationEnabled(enabled: boolean): void {
  localStorage.setItem(STORAGE_KEY, enabled ? '1' : '0')
}

/** 请求 Notification API 权限（如果尚未授权） */
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

function showOSNotification(payload: NotificationPayload): void {
  if (typeof Notification === 'undefined' || Notification.permission !== 'granted') {
    return
  }
  try {
    const n = new Notification(payload.title, {
      body: payload.body,
      // icon 可选
    })
    // 点击通知聚焦窗口
    n.onclick = () => {
      try {
        runtime.WindowShow && runtime.WindowShow()
      } catch {
        /* ignore */
      }
      n.close()
    }
    // 5 秒后自动关闭（部分浏览器需要）
    setTimeout(() => n.close(), 8000)
  } catch {
    /* Safari 等 iframe 内可能不可用 */
  }
}

function showAppNotification(payload: NotificationPayload): void {
  const type = payload.level === 'error' ? 'error'
    : payload.level === 'warning' ? 'warning' : 'success'
  ElNotification({
    title: payload.title,
    message: payload.body,
    type,
    duration: 6000,
    position: 'bottom-right'
  })
}

/**
 * 注册桌面通知监听，返回取消监听函数
 */
export function useDesktopNotification(): { cancel: () => void } {
  let cancelFn: (() => void) | null = null

  const handler = (payload: NotificationPayload): void => {
    if (!isNotificationEnabled()) return
    showAppNotification(payload)
    showOSNotification(payload)
  }

  onMounted(() => {
    // Wails runtime events
    if (runtime && typeof runtime.EventsOn === 'function') {
      runtime.EventsOn('notification', handler)
      cancelFn = () => runtime.EventsOff('notification')
    }
  })

  onBeforeUnmount(() => {
    if (cancelFn) cancelFn()
  })

  return {
    cancel: () => cancelFn?.()
  }
}
