//go:build integration

package xiaohongshu

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/go-rod/rod"
	"github.com/stretchr/testify/require"
)

func TestRednoteQrcodeReplacedImageLocal(t *testing.T) {
	b := newMentionFixtureBrowser(t)
	page := b.MustPage()
	defer page.Close()
	router := page.HijackRequests()
	router.MustAdd("*", func(h *rod.Hijack) {
		h.Response.SetHeader("Content-Type", "text/html; charset=utf-8")
		if !strings.HasPrefix(h.Request.URL().Path, "/login") {
			h.Response.SetBody(`<script>location.replace('/login')</script>`)
			return
		}
		h.Response.SetBody(`<img class="css-wemwzq" style="width:60px;height:60px" src="data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='60' height='60'%3E%3Crect width='60' height='60'/%3E%3C/svg%3E">
   <script>document.querySelector('img').onclick=()=>{
    const qr=document.createElement('img');qr.className='css-1lhmg90';document.body.append(qr);
    setTimeout(()=>{const loaded=qr.cloneNode();loaded.src="data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='244' height='244'%3E%3Crect width='244' height='244'/%3E%3C/svg%3E";qr.replaceWith(loaded);},200);
   }</script>`)
	})
	go router.Run()
	defer router.Stop()
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	img, loggedIn, err := NewLogin(page).FetchQrcodeImage(ctx)
	require.NoError(t, err)
	require.False(t, loggedIn)
	require.True(t, strings.HasPrefix(img, "data:image/"))
}
