# XhhRobot

小黑盒类Grok机器人

# 能做什么？

自动检查指定用户的@消息并使用Ai回复

- 自定义提示词
- OpenAI Chat Completions 兼容 AI 接口
- 自动获取小黑盒 Web 请求版本参数

# 开始使用

[手把手教程](https://blog.sakurasen.cn/post/1778819699353/)

## AI 默认配置

AI 默认使用 DeepSeek 的 OpenAI 兼容接口和 `deepseek-flash` 模型。新生成的 `config.json` 已预填模型、接口地址和提示词，AI 部分通常只需要填写：

```json
"apiKey": "你的 DeepSeek API Key"
```

旧配置中的 `ai.token` 继续兼容，不需要立即迁移。MCP 默认关闭；需要工具调用时再手动开启并添加服务器配置。

## 小黑盒 Web 版本参数

默认启用 `xhh.autoVersion=true`。启动时程序会尝试访问小黑盒官网，解析首页及其引用的前端脚本，从网页 fetch/XHR 的公共参数中识别 `version` 和 `web_version`，再用于后续 API 请求。

自动探测失败时会使用内置回退值：

- `version=999.0.4`
- `web_version=3.0`

如需完全手工指定，可将 `autoVersion` 改为 `false`，然后填写 `version` / `webver`。

## 下载

前往[Release下载](https://github.com/ZzzHe2333/xhhRobot/releases)您对应的系统版本

## 云端更新

程序支持从 `ZzzHe2333/xhhRobot` 的 GitHub Releases 拉取最新更新包并更新当前程序：

```bash
xhhRobot -mode update
```

Windows 用户也可以把 `update.bat` 与 `xhhRobot.exe` 放在同一目录，双击 `update.bat` 一键更新。

更新器会按当前操作系统和 CPU 架构选择 ZIP 更新包；Release 提供 SHA-256 digest 时会自动校验。更新只替换主程序，不覆盖本地的 `config.json`、`cookie.json`、数据库和日志。

如需测试其他更新仓库，可临时设置环境变量 `XHHROBOT_UPDATE_REPO=owner/repo`；不设置时固定使用 `ZzzHe2333/xhhRobot`。

# PR&Issues

欢迎各位提出Pr以及Issues。

在提pr前还是建议先去Issues请求一下，避免与其他人冲突。

如果您拥有基于本项目的二次开源项目，欢迎提交Pr来修改下方内容：

# 基于本项目的其他项目

>排名不分先后

## [Openxhh](https://github.com/Www8881313/Openxhh)

Openxhh 是一个面向小黑盒的 AI 自动回复机器人。它不是只会看见一句 @机器人 的关键词脚本，而是尽量把帖子、楼层、图片、上下文和被点名的人一起读进去，再用 OpenAI 兼容接口生成更像真人接话的回复。
