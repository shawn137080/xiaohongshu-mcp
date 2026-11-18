//go:build integration

package xiaohongshu

import (
	"context"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// 使用海外编辑页已核实的 DOM 结构，不登录或访问外部网站。
func TestRednoteAIContentLocal(t *testing.T) {
	b := newMentionFixtureBrowser(t)
	for _, change := range []bool{true, false} {
		t.Run(map[bool]string{true: "设置并重复调用", false: "点击无效时中止"}[change], func(t *testing.T) {
			page := b.MustPage("about:blank")
			defer page.Close()
			page.MustSetDocumentContent(`<div class="d-select" onclick="document.querySelector('.d-option-name').style.display='block'">添加内容类型声明<span class="d-select-description"></span></div>
    <div class="d-option-name" style="display:none">笔记含AI合成内容</div>`)
			if change {
				page.MustEval(`() => document.querySelector('.d-option-name').onclick = () => {
     document.querySelector('.d-select-description').textContent = '笔记含AI合成内容';
     document.querySelector('.d-option-name').remove();
    }`)
			}
			ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
			defer cancel()
			err := setAIContent(page.Context(ctx))
			if !change {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NoError(t, setAIContent(page.Context(ctx)))
			require.Equal(t, aiDeclaration, page.MustElement(".d-select-description").MustText())
		})
	}
}

func TestRednoteCoverLocal(t *testing.T) {
	b := newMentionFixtureBrowser(t)
	cover := filepath.Join(t.TempDir(), "cover.png")
	f, err := os.Create(cover)
	require.NoError(t, err)
	require.NoError(t, png.Encode(f, image.NewRGBA(image.Rect(0, 0, 30, 40))))
	require.NoError(t, f.Close())
	for _, uploadOK := range []bool{true, false} {
		t.Run(map[bool]string{true: "等待上传完成", false: "上传无更新时中止"}[uploadOK], func(t *testing.T) {
			page := b.MustPage("about:blank")
			defer page.Close()
			page.MustSetDocumentContent(`<style>.cover-edit-entry{display:none}.default:hover .cover-edit-entry{display:block}</style>
    <div class="cover-plugin-preview" aria-busy="false"><div class="cover"><div class="default" style="background-image:url('old.png');width:200px;height:200px"><button class="cover-edit-entry">编辑封面</button></div></div></div>
    <div id="modal" style="display:none"><h6 class="d-tabs-header-label">上传封面</h6>
     <div class="upload-wrapper" style="display:none"><input type="file" accept="image/png"></div>
     <div class="d-modal-footer"><button class="btn-confirm" disabled>确定</button></div></div>`)
			page.MustEval(`ok => {
    const modal=document.querySelector('#modal');
    document.querySelector('.cover-edit-entry').onclick=()=>modal.style.display='block';
    document.querySelector('h6').onclick=()=>document.querySelector('.upload-wrapper').style.display='block';
    document.querySelector('input').onchange=()=>document.querySelector('.btn-confirm').disabled=false;
    document.querySelector('.btn-confirm').onclick=()=>{
     modal.style.display='none';
     const root=document.querySelector('.cover-plugin-preview');
     root.setAttribute('aria-busy','true');
     if(ok) setTimeout(()=>{
      document.querySelector('.default').style.backgroundImage="url('uploaded.png')";
      root.setAttribute('aria-busy','false');
     },100);
    };
   }`, uploadOK)
			ctx, cancel := context.WithTimeout(t.Context(), 3*time.Second)
			defer cancel()
			err := setVideoCover(page.Context(ctx), cover)
			if !uploadOK {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, "cover.png", page.MustEval(`()=>document.querySelector('input').files[0].name`).String())
			require.Contains(t, *page.MustElement(".default").MustAttribute("style"), "uploaded.png")
		})
	}
}

func TestRednoteScheduleLocal(t *testing.T) {
	b := newMentionFixtureBrowser(t)
	page := b.MustPage("about:blank").Timeout(10 * time.Second)
	defer page.Close()
	page.MustSetDocumentContent(`<div class="date-picker-container"><input value="2026-11-18 10:00"><div class="d-datepicker-suffix">北京时间</div></div><button>blur</button>`)
	value, err := time.Parse(time.RFC3339, "2026-11-19T10:30:00-05:00")
	require.NoError(t, err)
	require.NoError(t, setDateTime(page.GetContext(), page, value))
	require.Equal(t, "2026-11-19 23:30", page.MustElement("input").MustProperty("value").String())
}
