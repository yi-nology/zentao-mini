package zentao

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

// 纯评论通道：禅道 REST API v1 没有"加备注"端点（仅 confirm/resolve/close/activate/assign
// 能附带 comment），但 Web 端存在 action-comment 路由。这里用配置的禅道账号建立
// Web 会话（zentaosid Cookie），复刻 Web 端评论提交：
//   1. GET  {server}/user-login.html                  拿 zentaosid
//   2. POST {server}/user-login.html                  表单登录（account/password）
//   3. POST {server}/action-comment-bug-{id}.html     comment=<内容>
// 会话缓存在 Client 上，失败时重登一次。

const webSessionTTL = 8 * time.Hour

type webSession struct {
	jar       *cookiejar.Jar
	loggedIn  bool
	expiresAt time.Time
}

// Client.webSession.session 即 webSession：与 REST Token 相互独立的 Web 登录态

func (c *Client) getWebSession() (*webSession, error) {
	c.webSession.mu.Lock()
	defer c.webSession.mu.Unlock()

	s := &c.webSession.session
	if s.jar != nil && s.loggedIn && time.Now().Before(s.expiresAt) {
		return s, nil
	}
	if s.jar == nil {
		jar, err := cookiejar.New(nil)
		if err != nil {
			return nil, fmt.Errorf("创建 Cookie Jar 失败: %w", err)
		}
		s.jar = jar
	}
	if err := c.webLogin(s); err != nil {
		return nil, err
	}
	s.loggedIn = true
	s.expiresAt = time.Now().Add(webSessionTTL)
	return s, nil
}

// ResetWebSession 配置变更或会话失效时清空 Web 会话
func (c *Client) ResetWebSession() {
	c.webSession.mu.Lock()
	defer c.webSession.mu.Unlock()
	c.webSession.session = webSession{}
}

// webHTTPClient 与 SDK 一致：禅道自建服务器常见自签名证书
func (c *Client) webHTTPClient(s *webSession) *http.Client {
	return &http.Client{
		Timeout: 60 * time.Second,
		Jar:     s.jar,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
}

func (c *Client) webLogin(s *webSession) error {
	c.mu.RLock()
	server, account, password := c.server, c.account, c.password.Get()
	c.mu.RUnlock()
	if server == "" || account == "" {
		return fmt.Errorf("禅道服务器或账号未配置，无法建立 Web 会话")
	}

	httpClient := c.webHTTPClient(s)

	// 1. 访问登录页拿 zentaosid
	loginURL := server + "/user-login.html"
	resp, err := httpClient.Get(loginURL)
	if err != nil {
		return fmt.Errorf("访问禅道登录页失败: %w", err)
	}
	_, _ = io.ReadAll(resp.Body)
	resp.Body.Close()

	// 2. 表单登录（密码为账号原始密码，与 Web 端一致；RSA 加密登录的服务器暂不支持）
	form := url.Values{}
	form.Set("account", account)
	form.Set("password", password)
	form.Set("keepLogin", "0")
	req, err := http.NewRequest(http.MethodPost, loginURL, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", loginURL)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	resp, err = httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("禅道 Web 登录请求失败: %w", err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("禅道 Web 登录失败, 状态码: %d, 响应: %.200s", resp.StatusCode, string(body))
	}
	// 登录接口返回 JSON（ajax）时检查失败标记；返回 HTML（locate 跳转）视为成功
	if strings.Contains(string(body), `"result":"fail"`) || strings.Contains(string(body), `"result": "fail"`) {
		return fmt.Errorf("禅道 Web 登录被拒绝（账号或密码错误，或服务器开启了加密登录）")
	}
	return nil
}

// AddBugComment 给 Bug 添加备注（评论）。走 Web 会话通道，成功后禅道 bug 历史出现
// 一条 Commented 记录，与网页端手动评论完全一致。
func (c *Client) AddBugComment(bugID int, comment string) error {
	if err := c.doAddBugComment(bugID, comment); err != nil {
		// 会话可能过期：重登一次再试
		c.ResetWebSession()
		if retryErr := c.doAddBugComment(bugID, comment); retryErr != nil {
			return retryErr
		}
	}
	return nil
}

func (c *Client) doAddBugComment(bugID int, comment string) error {
	c.mu.RLock()
	server := c.server
	c.mu.RUnlock()
	if server == "" {
		return fmt.Errorf("禅道服务器地址为空")
	}

	s, err := c.getWebSession()
	if err != nil {
		return err
	}

	// PATH_INFO 风格为主，GET 风格兜底（取决于禅道 requestType 配置）
	targets := []string{
		fmt.Sprintf("%s/action-comment-bug-%d.html", server, bugID),
		fmt.Sprintf("%s/index.php?m=action&f=comment&objectType=bug&objectID=%d", server, bugID),
	}

	var lastBody string
	var lastStatus int
	for _, target := range targets {
		form := url.Values{}
		form.Set("comment", comment)
		req, err := http.NewRequest(http.MethodPost, target, strings.NewReader(form.Encode()))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Referer", fmt.Sprintf("%s/bug-view-%d.html", server, bugID))
		req.Header.Set("X-Requested-With", "XMLHttpRequest")

		resp, err := c.webHTTPClient(s).Do(req)
		if err != nil {
			return fmt.Errorf("提交评论请求失败: %w", err)
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		lastBody, lastStatus = string(body), resp.StatusCode

		if resp.StatusCode == http.StatusOK && !commentResponseIndicatesFailure(body) {
			return nil
		}
		// 404/405：路由形态不对，试下一个目标；其他状态码直接报错
		if resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusMethodNotAllowed {
			break
		}
	}

	if lastStatus == http.StatusForbidden || strings.Contains(lastBody, "accessDenied") || strings.Contains(lastBody, "已禁止") {
		return fmt.Errorf("禅道拒绝了评论操作（无权限）")
	}
	return fmt.Errorf("提交评论失败, 状态码: %d, 响应: %.200s", lastStatus, lastBody)
}

func commentResponseIndicatesFailure(body []byte) bool {
	s := string(body)
	failMarkers := []string{
		`"result":"fail"`, `"result": "fail"`,
		`"status":"fail"`, `"status": "fail"`,
		"accessDenied", "未登录", "user-login",
	}
	for _, marker := range failMarkers {
		if strings.Contains(s, marker) {
			return true
		}
	}
	return false
}

// WebActivateBug / WebCloseBug：REST v1 缺 activate 端点（老版本禅道）、close 在部分
// 版本上静默失效，均用 Web 会话通道兜底（与评论同一会话），等价网页端操作。
func (c *Client) WebActivateBug(bugID int, assignedTo, comment string) error {
	form := url.Values{}
	form.Set("assignedTo", assignedTo)
	form.Set("comment", comment)
	return c.webBugFormAction(bugID, "activate", form)
}

func (c *Client) WebCloseBug(bugID int, comment string) error {
	form := url.Values{}
	form.Set("comment", comment)
	return c.webBugFormAction(bugID, "close", form)
}

func (c *Client) webBugFormAction(bugID int, action string, form url.Values) error {
	if err := c.doWebBugFormAction(bugID, action, form); err != nil {
		// 会话可能过期：重登一次再试
		c.ResetWebSession()
		return c.doWebBugFormAction(bugID, action, form)
	}
	return nil
}

func (c *Client) doWebBugFormAction(bugID int, action string, form url.Values) error {
	c.mu.RLock()
	server := c.server
	c.mu.RUnlock()
	if server == "" {
		return fmt.Errorf("禅道服务器地址为空")
	}

	s, err := c.getWebSession()
	if err != nil {
		return err
	}

	targets := []string{
		fmt.Sprintf("%s/bug-%s-%d.html", server, action, bugID),
		fmt.Sprintf("%s/index.php?m=bug&f=%s&bugID=%d", server, action, bugID),
	}
	var lastBody string
	var lastStatus int
	for _, target := range targets {
		req, err := http.NewRequest(http.MethodPost, target, strings.NewReader(form.Encode()))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Referer", fmt.Sprintf("%s/bug-view-%d.html", server, bugID))
		req.Header.Set("X-Requested-With", "XMLHttpRequest")

		resp, err := c.webHTTPClient(s).Do(req)
		if err != nil {
			return fmt.Errorf("提交 %s 请求失败: %w", action, err)
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		lastBody, lastStatus = string(body), resp.StatusCode

		if resp.StatusCode == http.StatusOK && !commentResponseIndicatesFailure(body) {
			return nil
		}
		if resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusMethodNotAllowed {
			break
		}
	}
	return fmt.Errorf("禅道 %s 操作失败, 状态码: %d, 响应: %.200s", action, lastStatus, lastBody)
}
