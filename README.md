![GitHub License](https://img.shields.io/github/license/ZiRunHua/LeapLedger)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/ZiRunHua/LeapLedger)
![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/ZiRunHua/LeapLedger/CI.yml)
![GitHub Release](https://img.shields.io/github/v/release/ZiRunHua/LeapLedger)
![Docker Pulls](https://img.shields.io/docker/pulls/xiaozirun/leap-ledger)
[![Go Report Card](https://goreportcard.com/badge/github.com/ZiRunHua/LeapLedger)](https://goreportcard.com/report/github.com/ZiRunHua/LeapLedger)
![GitHub stars](https://img.shields.io/github/stars/ZiRunHua/LeapLedger?style=social)

<h1 align="center">LeapLedger</h1>
<p align="center">
  <a href="README.en.md">English</a> | <a href="README.md">简体中文</a>
</p>

`LeapLedger`是一个的前后端分离免费开源的记账软件，`flutter`带来丝滑流畅的使用体验，在未来轻松扩展至iOS、Mac和Windows，服务端使用`Gin`框架，基于`Nats`、`Mysql`、`Redis`实现，带来快速的响应和稳定的服务。使用`docker`即可快速部署和构建客户端安装包。

<table>
  <tr>
    <td align="center"><img src="https://github.com/user-attachments/assets/e5151e7a-6b1f-4903-b4f1-8ffdc20c1b46" alt="Description" width="150"></td>
    <td align="center"><img src="https://github.com/user-attachments/assets/03dce625-a340-4aa5-92fd-ac4e59ee18b9" alt="Description" width="150"></td>
    <td align="center"><img src="https://github.com/user-attachments/assets/fd19053c-a469-4fcd-9d1e-9371c094039c" alt="Description" width="150"></td>
    <td align="center"><img src="https://github.com/user-attachments/assets/4d605f41-18fc-41b0-bbdf-d50ae1ecc550" alt="Description" width="150"></td>
    <td align="center"><img src="https://github.com/user-attachments/assets/0579110f-66b5-4739-9cc7-bcaeef4e246f" alt="Description" width="150"></td>
  </tr>
</table>

## 客户端
flutter客户端项目传送：https://github.com/ZiRunHua/LeapLedger-App

最新体验Android安装包下载：https://github.com/ZiRunHua/LeapLedger-App/releases/tag/v1.0.0 (数据不定期删除)
## 目录

- [我们有](#我们有)
  - [功能](#功能)
  - [API服务](#api服务)
- [运行/部署](#运行部署)
  - [构建镜像](#构建镜像)
- [API文档](#api文档)
- [协议](#协议)
- [贡献](#贡献)
- [联系我](#联系我)
- [致谢](#致谢)

## 我们有
* :iphone:基于Flutter的强大架构，流畅的使用体验，未来轻松扩展至iOS、Mac和Windows


* :whale:无论是服务器部署还是客户端打包，一切通过Docker搞定:parasol_on_ground:
### 功能
* :family_man_woman_girl_boy:共享账本，独立记账的同时将记录同步至你的情侣和家庭账本


* :timer_clock:定时记账，每月的房租、通讯费费等固定支出，手动记录太麻烦，没问题我们有


* :credit_card:每月只想处理一次账本，没问题我们可以导入支付宝和微信账单

除了这些我们还有
* :books:多账本管理和:earth_africa:时区账本
* 清晰的视图来了解近期的情况
* 轻松浏览的记录和图表 :bar_chart:

### API服务
我们有一个快速响应和稳定的API服务

* :shield:基于Gin中间件的账本鉴权和JWT身份认证确保安全


* :zap:基于Nats的异步和事件驱动带来更快的响应


* :mailbox:Outbox模式，在异步的同时确保数据的一致性:dart:


* :floppy_disk:死信队列保证消息不丢失，提高系统的稳定性


* :arrows_counterclockwise:基于`Gorm`的数据库更新和初始化，保证API服务的无缝升级

## 运行/部署
克隆项目
```bash
git clone https://github.com/ZiRunHua/LeapLedger.git
```
首次先启动mysql
```bash
docker-compose up -d leap-ledger-mysql
```
查看mysql日志，待显示`ready for connections`再执行`docker-compose up -d`
```bash
docker-compose logs -f leap-ledger-mysql
```
```bash
docker-compose up -d
```
访问http://localhost:8080/public/health 验证服务

自定义配置详见：[./config.yaml](./config.yaml)

客户端打包详见：https://github.com/ZiRunHua/LeapLedger-App


### 构建镜像

`docker/Dockerfile`可以构建携带go编译环境的镜像，但显然go的运行只需要二进制文件即可，所以你可以构建一个最小镜像

最小镜像
```bash
docker build -t xiaozirun/leap-ledger:build -f docker/Dockerfile.build .
```
2345端口的远程调试镜像
```bash
docker build -t xiaozirun/leap-ledger:debug -f docker/Dockerfile.debug .
```
测试镜像
```bash
docker build -t xiaozirun/leap-ledger:test -f docker/Dockerfile.test .
```
请注意镜像标签的不同
## API文档

采用了RESTful API设计风格,可以选择查看以下某种形式的文档
* [ApiFox接口文档](https://apifox.com/apidoc/shared-df940a71-63e8-4af7-9090-1be77ba5c3df)
* [swagger.json](docs/swagger.json)
* [swagger.yaml](docs/swagger.yaml)

## 协议
[协议](LICENSE) 是 [GNU Affero General Public License v3](https://www.gnu.org/licenses/agpl-3.0.html)

可以用来学习或个人使用，不得商业使用
## 贡献
LeapLedger项目仍在初期开发阶段，许多功能和细节还在不断完善中。

我们欢迎任何形式的贡献，包括但不限于：

* 代码贡献: 修复 bug、开发新功能、优化代码、编写测试等。

* 问题反馈: 提交 bug 报告、提出改进建议等。

如果您对LeapLedger感兴趣，请随时加入我们的社区并为LeapLedger的发展做出贡献。

我们会在`develop`分支进行功能开发和调整，而`main`分支则用于发布稳定版本。

## 联系我
邮箱 <a href="mailto:zeng807046079@gmail.com">zeng807046079@gmail.com</a>

## 致谢
感谢我的朋友尤同学帮我测试，这节省了我非常多的精力非常非常感谢。

## Stargazers over time
[![Stargazers over time](https://starchart.cc/ZiRunHua/LeapLedger.svg?variant=adaptive)](https://starchart.cc/ZiRunHua/LeapLedger)
