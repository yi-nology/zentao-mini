// 平台访问控制用例：匿名只读徽章、匿名写被拒跳登录页、登录/登出流程（独立浏览器会话，不污染共享会话的登录态）
const h = require('./helpers')

async function testAuthFlow() {
  const s = h.suite('认证与权限控制')
  // 独立 context：保持匿名状态，避免共享会话已登录的影响
  const { browser, page } = await h.newPage()
  try {
    // A1 匿名打开首页：头部应显示"只读模式"徽章和登录按钮
    try {
      await page.goto(h.BASE + '/', { waitUntil: 'domcontentloaded', timeout: 30000 })
      await page.waitForSelector('.auth-badge.readonly', { timeout: 15000 })
      const badgeText = ((await page.locator('.auth-badge.readonly').textContent()) || '').trim()
      const loginBtn = await page.locator('.auth-area .auth-btn', { hasText: '登录' }).count()
      h.record(s, 'A1 匿名只读徽章+登录按钮', badgeText.includes('只读') && loginBtn === 1 ? 'pass' : 'fail', `徽章="${badgeText}" 登录按钮=${loginBtn}`)
    } catch (e) {
      h.record(s, 'A1 匿名只读徽章+登录按钮', 'fail', e.message)
    }

    // A2 匿名执行写操作（清空日志缓冲）→ 后端 401 → 前端提示并跳转登录页
    try {
      await page.locator('.nav-menu .nav-item', { hasText: '系统日志' }).first().click()
      await page.waitForTimeout(800)
      await page.locator('.filter-card .el-button', { hasText: '清空缓冲' }).first().click()
      // 确认框
      const box = page.locator('.el-message-box:visible')
      await box.waitFor({ state: 'visible', timeout: 5000 })
      await box.locator('.el-message-box__btns .el-button').last().click()
      // 应跳转到 /login（拦截器 1.2s 后跳转）
      await page.waitForURL(/\/login/, { timeout: 8000 })
      h.record(s, 'A2 匿名写操作401跳登录页', page.url().includes('/login') ? 'pass' : 'fail', `url=${page.url()}`)
    } catch (e) {
      h.record(s, 'A2 匿名写操作401跳登录页', 'fail', e.message.split('\n')[0])
    }

    // A3 登录页错误密码 → 报错不跳转
    try {
      await page.locator('#username').fill(h.ADMIN_USER)
      await page.locator('#password').fill('wrong-password')
      await page.locator('.login-btn').click()
      const err = page.locator('.message.error')
      await err.waitFor({ state: 'visible', timeout: 6000 })
      const errText = ((await err.textContent()) || '').trim()
      h.record(s, 'A3 错误密码报错', errText.length > 0 ? 'pass' : 'warn', `提示="${errText.slice(0, 30)}"`)
    } catch (e) {
      h.record(s, 'A3 错误密码报错', 'fail', e.message.split('\n')[0])
    }

    // A4 正确登录 → 回到 redirect 目标页，头部显示管理员徽章 + 退出
    try {
      await page.locator('#password').fill(h.ADMIN_PASS)
      await page.locator('.login-btn').click()
      await page.waitForSelector('.auth-badge.admin', { timeout: 15000 })
      const badgeText = ((await page.locator('.auth-badge.admin').textContent()) || '').trim()
      const logoutBtn = await page.locator('.auth-area .auth-btn', { hasText: '退出' }).count()
      h.record(s, 'A4 登录成功管理员徽章', badgeText.includes('管理员') && logoutBtn === 1 ? 'pass' : 'fail', `徽章="${badgeText}"`)
    } catch (e) {
      h.record(s, 'A4 登录成功管理员徽章', 'fail', e.message.split('\n')[0])
    }

    // A5 管理员执行写操作（清空日志缓冲）不再被 401 拦截
    try {
      await page.locator('.nav-menu .nav-item', { hasText: '系统日志' }).first().click()
      await page.waitForTimeout(600)
      await page.locator('.filter-card .el-button', { hasText: '清空缓冲' }).first().click()
      const box = page.locator('.el-message-box:visible')
      await box.waitFor({ state: 'visible', timeout: 5000 })
      await box.locator('.el-message-box__btns .el-button').last().click()
      await page.waitForTimeout(1000)
      const stillOnLogs = page.url().includes('/logs')
      h.record(s, 'A5 管理员写操作放行', stillOnLogs ? 'pass' : 'fail', `url=${page.url()}`)
    } catch (e) {
      h.record(s, 'A5 管理员写操作放行', 'fail', e.message.split('\n')[0])
    }

    // A6 退出 → 恢复只读徽章
    try {
      await page.locator('.auth-area .auth-btn', { hasText: '退出' }).first().click()
      await page.waitForSelector('.auth-badge.readonly', { timeout: 8000 })
      h.record(s, 'A6 退出恢复只读', 'pass')
    } catch (e) {
      h.record(s, 'A6 退出恢复只读', 'fail', e.message.split('\n')[0])
    }

    await h.shot(page, 'auth-flow')
  } finally {
    await browser.close()
  }
}

async function run(_page) {
  // 使用独立浏览器会话：认证流程会切换登录态，不能复用共享会话
  await testAuthFlow()
}

module.exports = run
