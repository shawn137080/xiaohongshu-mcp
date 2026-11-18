package xiaohongshu

import (
	"net/url"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/pkg/errors"
)

const (
	urlOfCreatorEntry     = "https://creator.rednote.com/publish/publish"
	creatorUserSelector   = ".user-info .name-box"
	creatorSessionTimeout = 15 * time.Second
)

// ensureCreatorSession 以创作者页面确认登录，不能仅凭残留 cookie 判断。
func ensureCreatorSession(page *rod.Page) error {
	// 让浏览器日期解析与界面的北京时间标注一致。
	if err := (proto.EmulationSetTimezoneOverride{TimezoneID: "Asia/Shanghai"}).Call(page); err != nil {
		return errors.Wrap(err, "设置发布页面时区失败")
	}
	if err := page.Navigate(urlOfCreatorEntry); err != nil {
		return errors.Wrap(err, "打开 RedNote 创作者中心失败")
	}
	if err := page.Timeout(creatorSessionTimeout).WaitLoad(); err != nil {
		return err
	}
	if _, err := page.Timeout(creatorSessionTimeout).Element(creatorUserSelector); err != nil {
		return errors.Wrap(err, "RedNote 创作者登录失效，请重新登录")
	}
	ok, err := creatorLoggedIn(page)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("未确认 RedNote 创作者登录状态")
	}
	return nil
}

func isCreatorURL(raw string) bool {
	u, err := url.Parse(raw)
	return err == nil && u.Scheme == "https" && u.Host == "creator.rednote.com" && !strings.HasPrefix(u.Path, "/login")
}

func creatorLoggedIn(page *rod.Page) (bool, error) {
	info, err := page.Info()
	if err != nil {
		return false, err
	}
	if !isCreatorURL(info.URL) {
		return false, nil
	}
	exists, el, err := page.Has(creatorUserSelector)
	if err != nil || !exists {
		return false, err
	}
	visible, err := el.Visible()
	if err != nil || !visible {
		return false, err
	}
	name, err := el.Text()
	return strings.TrimSpace(name) != "", err
}
