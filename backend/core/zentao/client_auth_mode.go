package zentao

// AuthMode 标识禅道客户端使用的认证与数据访问模式。
type AuthMode int

const (
	// AuthModeToken 默认模式：用户名/密码换取 REST Token，所有数据访问走
	// /api.php/v1/* REST 端点（Token HTTP 头）。适用于标准禅道开源版/企业版。
	AuthModeToken AuthMode = iota
	// AuthModeSession 会话模式：通过 /user-login.html 建立 PHP 会话，所有数据
	// 访问走传统 PHP 的 *.json 端点（zentaosid cookie）。
	// 适用于禁用了 REST API 或仅暴露 PHP 入口的禅道实例（如麒麟 pm.kylin.com）。
	AuthModeSession
)

// 麒麟统一认证域（kydc）。
// 禅道 /user-login.html 的 JS 逻辑：
//
//	realm != "kydc" ? md5(md5(password)+rand) : password
//
// 即 kydc 域下密码明文发送（由麒麟 SSO 网关处理）。
const RealmKylinSSO = "kydc"

// RealmLocal 本地认证域（禅道内置账号库），密码需做 md5(md5(pw)+rand) 混淆。
const RealmLocal = "local"
