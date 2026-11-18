package xiaohongshu

import (
	"context"
	"time"

	"github.com/go-rod/rod"
	"github.com/pkg/errors"
)

type LoginAction struct {
	page *rod.Page
}

func NewLogin(page *rod.Page) *LoginAction {
	return &LoginAction{page: page}
}

func (a *LoginAction) CheckLoginStatus(ctx context.Context) (bool, error) {
	pp := a.page.Context(ctx).Timeout(30 * time.Second)
	if err := pp.Navigate(urlOfCreatorEntry); err != nil {
		return false, err
	}
	if err := pp.WaitLoad(); err != nil {
		return false, err
	}
	// 等待前端完成登录重定向或用户信息渲染。
	deadline := time.Now().Add(10 * time.Second)
	for {
		ok, err := creatorLoggedIn(pp)
		if err != nil || ok {
			return ok, err
		}
		info, err := pp.Info()
		if err != nil {
			return false, err
		}
		if !isCreatorURL(info.URL) || time.Now().After(deadline) {
			return false, nil
		}
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		case <-time.After(200 * time.Millisecond):
		}
	}
}

type CurrentUser struct {
	Nickname string `json:"nickname"`
	UserID   string `json:"userId"`
}

// CurrentUser 不把创作者账号数字误当成个人主页 ID。
func (a *LoginAction) CurrentUser(ctx context.Context) (*CurrentUser, error) {
	el, err := a.page.Context(ctx).Timeout(10 * time.Second).Element(creatorUserSelector)
	if err != nil {
		return nil, err
	}
	name, err := el.Text()
	if err != nil {
		return nil, err
	}
	return &CurrentUser{Nickname: name}, nil
}

func (a *LoginAction) Login(ctx context.Context) error {
	ok, err := a.CheckLoginStatus(ctx)
	if err != nil || ok {
		return err
	}
	// 保留短信和扫码入口，由用户完成登录及条款确认。
	if !a.WaitForLogin(ctx) {
		return errors.New("等待 RedNote 登录超时或取消")
	}
	return nil
}

func (a *LoginAction) FetchQrcodeImage(ctx context.Context) (string, bool, error) {
	ok, err := a.CheckLoginStatus(ctx)
	if err != nil || ok {
		return "", ok, err
	}
	pp := a.page.Context(ctx).Timeout(15 * time.Second)
	// shortcut: 海外登录页使用散列 class，页面升级时重新核实扫码切换和二维码选择器。
	toggle, err := pp.Element("img.css-wemwzq")
	if err != nil {
		return "", false, errors.Wrap(err, "未找到扫码入口，请运行 go run ./cmd/login 手动登录")
	}
	if err := toggle.Click("left", 1); err != nil {
		return "", false, err
	}
	qr, err := pp.Element("img.css-1lhmg90")
	if err != nil {
		return "", false, errors.Wrap(err, "未找到 RedNote 二维码")
	}
	src, err := qr.Attribute("src")
	if err != nil {
		return "", false, err
	}
	if src == nil || *src == "" {
		return "", false, errors.New("RedNote 二维码为空")
	}
	return *src, false, nil
}

func (a *LoginAction) WaitForLogin(ctx context.Context) bool {
	pp := a.page.Context(ctx)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return false
		case <-ticker.C:
			ok, err := creatorLoggedIn(pp)
			// 登录跳转会短暂销毁页面上下文，继续等待而非误判失败。
			if err == nil && ok {
				return true
			}
		}
	}
}
