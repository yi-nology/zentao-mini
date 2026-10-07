// 补盲区用例：Logs 系统日志页（渲染/筛选/清空确认取消） + InitGuide 配置上传交互（route 拦截，不污染真实后端）
const h = require('./helpers')

async function navTo(page, label) {
  await page.locator('.nav-menu .nav-item', { hasText: label }).first().click()
  await page.waitForTimeout(500)
}

async function testLogs(page) {
  const s = h.suite('Logs 系统日志')
  await navTo(page, '系统日志')

  // LG1 状态栏渲染
  try {
    const statusBar = page.locator('.status-bar')
    await statusBar.waitFor({ state: 'visible', timeout: 8000 })
    const text = (await statusBar.textContent()) || ''
    h.record(s, 'LG1 缓冲状态栏渲染', text.includes('缓冲') ? 'pass' : 'warn', text.trim().replace(/\s+/g, ' ').slice(0, 50))
  } catch (e) {
    h.record(s, 'LG1 缓冲状态栏渲染', 'fail', e.message)
  }

  // LG2 日志表格有数据行（后端 ring buffer 记录了本次会话的请求日志）
  try {
    const rows = await h.waitTableRows(page)
    h.record(s, 'LG2 日志表格渲染数据行', rows > 0 ? 'pass' : 'warn', `行数 ${rows}`)
  } catch (e) {
    h.record(s, 'LG2 日志表格渲染数据行', 'fail', e.message)
  }

  // LG3 级别筛选（选 Error 后刷新，结果可能为空但不应报错）
  try {
    const levelSel = page.locator('.filter-card .el-select').first()
    await levelSel.locator('.el-select__wrapper').first().click()
    const item = page.locator('.el-select__popper:visible .el-select-dropdown__item', { hasText: 'Error' }).first()
    await item.waitFor({ state: 'visible', timeout: 5000 })
    await item.click()
    await page.waitForTimeout(800)
    const rows = await page.locator('.el-table__row').count()
    h.record(s, 'LG3 级别筛选(Error)', 'pass', `筛选后行数 ${rows}`)
  } catch (e) {
    h.record(s, 'LG3 级别筛选(Error)', 'fail', e.message)
  }

  // LG4 清空缓冲：只验证确认框弹出并取消，不真正清空（避免破坏其他用例的日志上下文）
  try {
    await page.locator('.filter-card .el-button', { hasText: '清空缓冲' }).first().click()
    const box = page.locator('.el-message-box:visible')
    await box.waitFor({ state: 'visible', timeout: 5000 })
    const boxText = (await box.textContent()) || ''
    // 取消按钮是按钮组第一个（locale 可能是中文"取消"或英文"Cancel"），按位置定位更稳
    await box.locator('.el-message-box__btns .el-button').first().click()
    await page.waitForTimeout(400)
    const boxGone = await box.count().catch(() => 0)
    h.record(s, 'LG4 清空缓冲确认框(取消)', boxText.includes('确定要清空') && boxGone === 0 ? 'pass' : 'warn', '已取消，未清空')
  } catch (e) {
    h.record(s, 'LG4 清空缓冲确认框(取消)', 'fail', e.message)
  }

  await h.shot(page, 'logs-page')
}

async function testInitGuideUpload(page) {
  const s = h.suite('InitGuide 配置上传')

  // 直接访问初始化页（路由守卫豁免 /init-guide）
  await page.goto(h.BASE + '/init-guide', { waitUntil: 'domcontentloaded', timeout: 30000 })
  await page.waitForTimeout(800)

  const fileInput = page.locator('input[type="file"]')
  const submitBtn = page.locator('.init-btn')

  // I1 非 .json 文件被前端直接拒绝（无网络请求）
  try {
    await fileInput.setInputFiles({ name: 'config.txt', mimeType: 'text/plain', buffer: Buffer.from('not json') })
    await page.waitForTimeout(400)
    const errText = await page.locator('.message.error').textContent().catch(() => '')
    h.record(s, 'I1 非json文件前端拒绝', (errText || '').includes('.json') ? 'pass' : 'fail', `提示="${(errText || '').trim()}"`)
  } catch (e) {
    h.record(s, 'I1 非json文件前端拒绝', 'fail', e.message)
  }

  // I2 上传成功流（route 拦截返回成功响应，不动真实后端）
  try {
    await page.route('**/api/init/upload', (route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ code: 200, message: '配置已保存，正在后台连接禅道...', data: null }),
      })
    )
    await fileInput.setInputFiles({ name: 'auth-config.json', mimeType: 'application/json', buffer: Buffer.from('{"salt":"x","iv":"y","encrypted_data":"z"}') })
    await submitBtn.click()
    const ok = page.locator('.message.success')
    await ok.waitFor({ state: 'visible', timeout: 6000 })
    const okText = (await ok.textContent()) || ''
    h.record(s, 'I2 上传成功提示', okText.includes('初始化成功') ? 'pass' : 'warn', okText.trim().slice(0, 40))
    await page.unroute('**/api/init/upload')
    // 阻止 2 秒后自动跳转主页影响后续用例
    await page.goto(h.BASE + '/init-guide', { waitUntil: 'domcontentloaded' })
    await page.waitForTimeout(500)
  } catch (e) {
    await page.unroute('**/api/init/upload').catch(() => {})
    h.record(s, 'I2 上传成功提示', 'fail', e.message)
  }

  // I3 上传失败流（route 拦截返回服务端错误）
  try {
    await page.route('**/api/init/upload', (route) =>
      route.fulfill({
        status: 500,
        contentType: 'application/json',
        body: JSON.stringify({ code: 50000, message: '加载认证配置失败', data: null }),
      })
    )
    await fileInput.setInputFiles({ name: 'bad.json', mimeType: 'application/json', buffer: Buffer.from('{"broken":true}') })
    await submitBtn.click()
    const err = page.locator('.message.error')
    await err.waitFor({ state: 'visible', timeout: 6000 })
    const errText = (await err.textContent()) || ''
    h.record(s, 'I3 上传失败提示', errText.includes('初始化失败') || errText.includes('加载认证配置') ? 'pass' : 'warn', errText.trim().slice(0, 40))
    await page.unroute('**/api/init/upload')
  } catch (e) {
    await page.unroute('**/api/init/upload').catch(() => {})
    h.record(s, 'I3 上传失败提示', 'fail', e.message)
  }

  await h.shot(page, 'init-guide-upload')
}

async function run(page) {
  // 上一套件（test-misc）结束时停在 init-guide 页（Layout 外没有侧边栏），先回主页
  const product = process.env.TEST_PRODUCT || '1'
  await page.goto(h.BASE + '/?product=' + product, { waitUntil: 'domcontentloaded', timeout: 30000 })
  await page.waitForSelector('.nav-menu', { timeout: 20000 })
  await page.waitForTimeout(800)
  await testLogs(page)
  await testInitGuideUpload(page)
}

module.exports = run
