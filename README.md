# XhhRobot

小黑盒类Grok机器人

# 能做什么？

自动检查指定用户的@消息并使用Ai回复

- 自定义提示词

- 自定义Ai接口

# 开始使用

[手把手教程](https://blog.sakurasen.cn/post/1778819699353/)

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
