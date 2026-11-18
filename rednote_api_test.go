package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestRednoteVideoCoverThroughMCPAndHTTP(t *testing.T) {
	video := filepath.Join(t.TempDir(), "video.mp4")
	require.NoError(t, os.WriteFile(video, []byte("fixture"), 0600))
	args := map[string]any{"title": "test", "content": "test", "video": video, "cover": "relative.png", "ai_content": true}
	app := NewAppServer(NewXiaohongshuService(), "")
	ctx := context.Background()
	ct, st := mcp.NewInMemoryTransports()
	ss, err := app.mcpServer.Connect(ctx, st, nil)
	require.NoError(t, err)
	defer ss.Close()
	client := mcp.NewClient(&mcp.Implementation{Name: "rednote-test", Version: "1"}, nil)
	session, err := client.Connect(ctx, ct, nil)
	require.NoError(t, err)
	defer session.Close()
	tools, err := session.ListTools(ctx, nil)
	require.NoError(t, err)
	require.Len(t, tools.Tools, 5)
	for _, tool := range tools.Tools {
		if tool.Name == "publish_with_video" {
			schema, err := json.Marshal(tool.InputSchema)
			require.NoError(t, err)
			require.Contains(t, string(schema), `"cover"`)
			require.Contains(t, string(schema), `"ai_content"`)
		}
	}
	result, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "publish_with_video", Arguments: args})
	require.NoError(t, err)
	require.True(t, result.IsError)
	require.Contains(t, result.Content[0].(*mcp.TextContent).Text, "封面必须使用本地绝对路径")
	body, err := json.Marshal(args)
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/publish_video", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	setupRoutes(app).ServeHTTP(recorder, req)
	require.Equal(t, http.StatusInternalServerError, recorder.Code)
	require.Contains(t, recorder.Body.String(), "封面必须使用本地绝对路径")
}

func TestHealthIdentifiesRednote(t *testing.T) {
	recorder := httptest.NewRecorder()
	setupRoutes(NewAppServer(NewXiaohongshuService(), "")).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/health", nil))
	require.Equal(t, http.StatusOK, recorder.Code)
	var result struct {
		Data map[string]any `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &result))
	require.Equal(t, "rednote-mcp", result.Data["service"])
	require.Equal(t, "rednote", result.Data["platform"])
	require.Equal(t, "https://creator.rednote.com", result.Data["creator_url"])
}
