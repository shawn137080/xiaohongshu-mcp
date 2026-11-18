package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
)

// validateVideoFiles 在启动浏览器或上传前拒绝无效路径及封面。
func validateVideoFiles(video, cover string) error {
	if !filepath.IsAbs(video) {
		return fmt.Errorf("视频必须使用本地绝对路径")
	}
	info, err := os.Stat(video)
	if err != nil {
		return fmt.Errorf("视频不可访问: %w", err)
	}
	if !info.Mode().IsRegular() || info.Size() == 0 {
		return fmt.Errorf("视频必须是非空普通文件")
	}
	// shortcut: 首版保留 1 GiB 上限，更大文件需验证上传超时后再开放。
	if info.Size() > 1<<30 {
		return fmt.Errorf("视频超过当前适配版的 1 GiB 上限")
	}
	if cover == "" {
		return nil
	}
	if !filepath.IsAbs(cover) {
		return fmt.Errorf("封面必须使用本地绝对路径")
	}
	info, err = os.Stat(cover)
	if err != nil {
		return fmt.Errorf("封面不可访问: %w", err)
	}
	if !info.Mode().IsRegular() || info.Size() == 0 || info.Size() > 20<<20 {
		return fmt.Errorf("封面必须是非空普通图片文件，且不超过 20 MiB")
	}
	f, err := os.Open(cover)
	if err != nil {
		return err
	}
	defer f.Close()
	cfg, _, err := image.DecodeConfig(f)
	if err != nil || cfg.Width <= 0 || cfg.Height <= 0 {
		return fmt.Errorf("封面必须是有效的 PNG 或 JPEG 图片")
	}
	return nil
}
