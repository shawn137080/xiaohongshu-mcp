# RedNote 海外发布版

上游：<https://github.com/xpzouying/xiaohongshu-mcp>
Fork：<https://github.com/shawn137080/xiaohongshu-mcp>
开发分支：`codex/rednote-overseas`。

## 适配范围

- 登录与发布仅访问 `creator.rednote.com`。登录检查使用创作者账号标识，不再依赖大陆社区首页或残留 cookie。
- 默认会话文件为 `rednote-cookies.json`，不读取旧 `cookies.json` 或 `/tmp/cookies.json`；可用 `COOKIES_PATH` 指定路径。新文件权限为 `0600`。
- 视频请求新增 `cover`：本地 PNG/JPEG 绝对路径。自动进入封面编辑器、上传、确认，等待封面地址更新且上传状态完成。失败则中止发布。
- 图文和视频请求新增 `ai_content`：为 `true` 时选择「笔记含AI合成内容」，回读已选值；设置失败则中止发布。默认 `false` 保持原行为，不代替原创声明。
- 保留话题、定时发布和可见范围。当前适配**简体中文创作者界面**。
- 本地视频最大 1 GiB，上传与处理共用 30 分钟上限，不压缩或经 MCP 传输视频二进制。封面最大 20 MiB。
- 只认可海外发布成功页或笔记管理页的跳转；跳到登录页不会被误报为发布成功。

海外页面显示最大 20 GB，但此版未验证大于 1 GiB 文件，因此暂不开放。
上游搜索、信息流、个人主页、评论、点赞、收藏及通知工具尚未适配，**不对外开放**；实现源码保留以便后续合并上游。

## 登录与启动

需要 `go.mod` 指定的 Go 工具链。沿用上游内置浏览器下载机制。

```sh
go run ./cmd/login
go run . -headless=false -port 127.0.0.1:18060
```

登录命令打开独立浏览器，由用户完成短信或扫码登录及条款确认，最长等待 10 分钟。不会借用 Chrome 的现有会话。
扫码入口的散列 class 来自当前海外登录页，网站更新后可能需要调整；若扫码接口失败，可使用上述可见浏览器登录。

上述命令仅监听 `127.0.0.1:18060`（省略 `-port` 时沿用上游的所有网卡监听）；MCP 地址为 `http://127.0.0.1:18060/mcp`。非本机访问须显式设置 `-port` 并配置 `AUTH_TOKEN`。
不要使用上游预编译镜像或二进制来运行本 fork 的新功能，应从当前分支构建。

开放的 MCP 工具：

- `check_login_status`
- `get_login_qrcode`
- `delete_cookies`
- `publish_content`
- `publish_with_video`

REST 接口：`GET /api/v1/login/status`、`GET /api/v1/login/qrcode`、`DELETE /api/v1/login/cookies`、`POST /api/v1/publish`、`POST /api/v1/publish_video`。

视频请求示例（**调用会实际发布，不是草稿或预览接口**）：

```json
{
  "title": "示例标题",
  "content": "示例正文",
  "video": "/absolute/path/video.mp4",
  "cover": "/absolute/path/cover.png",
  "ai_content": true,
  "tags": ["AI实操"],
  "visibility": "仅自己可见"
}
```

`schedule_at` 使用带时区的 ISO8601 格式，范围为 1 小时至 14 天，例如 `2026-11-19T10:30:00-05:00`。
当前海外编辑页标注「北京时间」；发布页浏览器时区固定为 `Asia/Shanghai`，输入也转换为 UTC+08:00 并回读，避免本机时区与页面标注不一致。不认识的时区标注会中止。
**平台实际执行时间未经发布验证**，首次生产定时发布前需人工核对。

## 验收记录（2026-11-18）

通过浏览器在当前海外账号完成了不发布的三秒合成视频测试：

- 视频上传并进入编辑页。
- 自定义纯色 PNG 封面上传成功，预览和资源地址更新，上传状态结束。
- AI 声明已选中并回读为「笔记含AI合成内容」。
- 标题、正文、仅自己可见及定时开关可操作。话题联想框能打开，但测试词返回无结果，未验证真实话题关联。
- **没有点击发布，也没有点击暂存离开**。测试素材仍留在浏览器编辑页中。

这次线上验证由浏览器工具操作，**没有让 Go 服务直接连接真实账号跑完整流程**。Go 的新增封面、AI 声明和时间输入实现则通过使用已核实 DOM 结构的本地浏览器回归测试验证。

检查命令：

```sh
go test ./...
go test -tags integration ./xiaohongshu -run TestRednote -count=1 -v
go build ./...
git diff --check
```

回归覆盖参数到 MCP/REST 的传递、无效文件拒绝、声明幂等及点击失败、封面异步完成及失败、时区转换和输入、登录/成功跳转域名边界，以及不开放大陆工具。

未验证：真实发布、发布成功页的实际跳转、定时送达、接近 1 GiB 的上传、完整扫码登录、图文发布、原创/@用户/商品等继承功能。未确认发布成功时不要盲目重试，应先到笔记管理检查，避免重复发布。
