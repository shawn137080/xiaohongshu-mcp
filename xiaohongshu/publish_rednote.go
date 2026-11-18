package xiaohongshu

import (
	"net/url"
	"time"

	"github.com/go-rod/rod"
	"github.com/pkg/errors"
)

const aiDeclaration = "笔记含AI合成内容"

// setAIContent 回读已选值，不能把下拉列表里的选项误当成声明成功。
func setAIContent(page *rod.Page) error {
	pp := page.Timeout(15 * time.Second)
	exists, selected, err := pp.HasR(".d-select-description", "^"+aiDeclaration+"$")
	if err != nil {
		return err
	}
	if exists {
		return selected.WaitVisible()
	}
	trigger, err := pp.ElementR(".d-select", "添加内容类型声明")
	if err != nil {
		return errors.Wrap(err, "未找到内容类型声明入口")
	}
	if err := trigger.Click("left", 1); err != nil {
		return err
	}
	option, err := pp.ElementR(".d-option-name", "^"+aiDeclaration+"$")
	if err != nil {
		return errors.Wrap(err, "未找到 AI 合成内容选项")
	}
	if err := option.Click("left", 1); err != nil {
		return err
	}
	selected, err = pp.ElementR(".d-select-description", "^"+aiDeclaration+"$")
	if err != nil {
		return errors.Wrap(err, "未确认 AI 声明已选中")
	}
	return selected.WaitVisible()
}

func setVideoCover(page *rod.Page, coverPath string) error {
	pp := page.Timeout(2 * time.Minute)
	preview, err := pp.Element(".cover-plugin-preview .cover > .default")
	if err != nil {
		return errors.Wrap(err, "未找到视频封面预览")
	}
	before, err := preview.Attribute("style")
	if err != nil {
		return err
	}
	// 鼠标进入预览区域后才显示编辑入口。
	if err := preview.Hover(); err != nil {
		return err
	}
	edit, err := pp.Element(".cover-edit-entry")
	if err != nil {
		return err
	}
	if err := edit.Click("left", 1); err != nil {
		return err
	}
	tab, err := pp.ElementR(".d-tabs-header-label", "^上传封面$")
	if err != nil {
		return err
	}
	if err := tab.Click("left", 1); err != nil {
		return err
	}
	input, err := pp.Element(".upload-wrapper input[type=file][accept*='image/']")
	if err != nil {
		return errors.Wrap(err, "未找到封面文件输入框")
	}
	if err := input.SetFiles([]string{coverPath}); err != nil {
		return err
	}
	confirm, err := pp.Element(".d-modal-footer button.btn-confirm")
	if err != nil {
		return err
	}
	if err := confirm.WaitEnabled(); err != nil {
		return err
	}
	if err := confirm.Click("left", 1); err != nil {
		return err
	}
	// 上传异步完成后预览地址更新；任一条件未满足都不能继续发布。
	return pp.Wait(rod.Eval(`before => {
  const root = document.querySelector('.cover-plugin-preview');
  const cover = root && root.querySelector('.cover > .default');
  const style = cover && cover.getAttribute('style');
  return root && root.getAttribute('aria-busy') !== 'true' && style && style !== before &&
   /background-image:\s*url\(/.test(style) && !root.textContent.includes('封面上传中');
 }`, before))
}

func creatorScheduleTime(t time.Time) string {
	return t.In(time.FixedZone("UTC+08:00", 8*60*60)).Format("2006-01-02 15:04")
}

func isPublishSuccessURL(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil || !isCreatorURL(raw) {
		return false
	}
	return u.Path == "/publish/success" || u.Path == "/new/note-manager"
}
