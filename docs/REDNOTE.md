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
已核实的海外布局分别标注「北京时间」或 `Asia/Shanghai`；发布页浏览器时区固定为 `Asia/Shanghai`，输入也转换为 UTC+08:00 并回读，避免本机时区与页面标注不一致。不认识的时区标注会中止。
**已用私密视频验证多伦多时间与编辑页时间的换算及到点发布**，首次生产定时发布前需人工核对。

## 验收记录（2026-11-18）

已启动本地 Go 服务及项目自带 Chromium，以真实海外账号完成发布前实测：

- `creator.rednote.com` 登录成功并保存独立会话；重新启动浏览器后识别账号。
- 三秒合成视频实际上传成功。
- 自定义 PNG 封面上传成功，预览地址更新。
- 标题、正文、AI 声明、仅自己可见均设置成功并回读。
- `AI` 话题实际关联为编辑器话题节点，而非普通文字。
- 定时开关及时间输入成功：示例多伦多 `2026-11-18 16:15` 转为页面 `2026-11-19 05:15`，回读一致。
- 无登录会话的二维码接口在真实页面成功返回图片。
- 运行中的 HTTP `/health` 返回 `platform=rednote`、`creator_url=https://creator.rednote.com`；MCP `tools/list` 仅开放上述五个工具。

**本轮发布前测试停在编辑页。** 测试调用正式的 `prepareVideo` 准备阶段；生产 `PublishVideo` 调用同一阶段后才执行提交。测试浏览器已关闭。

实测发现并修复的兼容问题：

1. 二维码加载时会替换图片节点，需要重新查询节点，不能等待旧节点。
2. 一种海外封面布局没有 `aria-busy` 属性，完成检测需兼容该布局并仍验证封面地址变化。
3. 定时控件存在两种布局及两种时区标注，现均兼容并限制等待时长。

本地 `debug/`（不提交 Git）保留真实测试结果 JSON 和截图。临时线上探测测试已删除。

检查命令：

```sh
go test ./...
go test -tags integration ./xiaohongshu -run TestRednote -count=1 -v
go build ./...
git diff --check
```

持久回归覆盖 MCP/REST 参数传递、无效文件拒绝、二维码节点替换、声明幂等与失败、封面异步完成与失败、两种定时布局、只填表不提交、域名边界及不开放大陆工具。

Docker 无头实测通过：Linux amd64 容器以只读挂载加载 1,019,178,009 字节的 120 秒 MP4 和 PNG 封面。海外编辑页在约 105 秒后允许发布，自定义封面预览地址已更新。测试停在编辑页，全程约 139 秒。

Mac 原生无头实测：提交前回读仅自己可见和定时时间。多伦多时间 15:36 管理页仍显示定时发布，15:37:08 复查时定时标记已消失，记录显示对应的发布时间、互动统计及仅自己可见。

未验证：图文发布、原创/@用户/商品等继承功能。未确认发布成功时不要盲目重试，应先到笔记管理检查，避免重复发布。
