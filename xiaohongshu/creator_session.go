package xiaohongshu

import (
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/pkg/errors"
	"github.com/xpzouying/xiaohongshu-mcp/humanize"
)

const (
	urlOfCreatorHome = `https://creator.xiaohongshu.com/?source=official`

	creatorSessionCookie  = "galaxy_creator_session_id"
	creatorSessionTimeout = 15 * time.Second
)

// ensureCreatorSession 确认创作者中心已登录。
// 2026-11 实测小红书不再下发 galaxy_creator_session_id，只认这个 cookie 会把已登录的号误判成失效。
// 现在两样有一样就算登上：还有这个 cookie；或者创作者中心页面没被跳去登录页、且看得到「发布笔记」。
func ensureCreatorSession(page *rod.Page) error {
	if hasCreatorSession(page) {
		return nil
	}

	if err := page.Navigate(urlOfCreatorHome); err != nil {
		return errors.Wrap(err, "打开创作者中心失败")
	}

	deadline := time.Now().Add(creatorSessionTimeout)
	for time.Now().Before(deadline) {
		time.Sleep(500 * time.Millisecond)
		if hasCreatorSession(page) || creatorPageLoggedIn(page) {
			humanize.Delay(page.GetContext(), humanize.AfterNavigate)
			return nil
		}
	}
	return errors.New("创作者中心登录失效，请重新扫码登录")
}

func hasCreatorSession(page *rod.Page) bool {
	cks, err := page.Browser().GetCookies()
	if err != nil {
		return false
	}
	for _, c := range cks {
		if c.Name == creatorSessionCookie {
			return true
		}
	}
	return false
}

// creatorPageLoggedIn 当前页在创作者中心、不是登录页，并且页面上有「发布笔记」入口。
func creatorPageLoggedIn(page *rod.Page) bool {
	info, err := page.Info()
	if err != nil {
		return false
	}
	if !strings.Contains(info.URL, "creator.xiaohongshu.com") || strings.Contains(info.URL, "login") {
		return false
	}
	res, err := page.Eval(`() => !!document.body && document.body.innerText.includes("发布笔记")`)
	if err != nil {
		return false
	}
	return res.Value.Bool()
}
