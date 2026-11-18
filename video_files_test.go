package main

import (
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateVideoFiles(t *testing.T) {
	dir := t.TempDir()
	video := filepath.Join(dir, "video.mp4")
	require.NoError(t, os.WriteFile(video, []byte("fixture"), 0600))
	cover := filepath.Join(dir, "cover.png")
	f, err := os.Create(cover)
	require.NoError(t, err)
	require.NoError(t, png.Encode(f, image.NewRGBA(image.Rect(0, 0, 10, 10))))
	require.NoError(t, f.Close())
	require.NoError(t, validateVideoFiles(video, cover))
	require.NoError(t, validateVideoFiles(video, ""))
	require.Error(t, validateVideoFiles("relative.mp4", cover))
	require.Error(t, validateVideoFiles(dir, cover))
	require.Error(t, validateVideoFiles(video, video))
	require.Error(t, validateVideoFiles(video, "relative.png"))
	require.Error(t, validateVideoFiles(video, filepath.Join(dir, "missing.png")))
	require.NoError(t, os.Truncate(video, (1<<30)+1))
	require.Error(t, validateVideoFiles(video, cover))
	require.NoError(t, os.Truncate(video, 0))
	require.Error(t, validateVideoFiles(video, cover))
}
