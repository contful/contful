# Contful

简体中文 | [English](./README_EN.md)

Contful是一个开源的Headless CMS项目，以其卓越的功能和特性为企业级内容管理提供了全新的解决方案。作为Contentful的缩写，Contful融合了“Content”（内容）和“ful”（充满的）两个词汇，象征着它致力于打造一个充满丰富、有深度和有广度内容的管理平台。

Contful以API为中心，通过强大的API接口，实现了前后端分离的内容管理模式，使得内容的创建、编辑、管理和发布变得更加高效和灵活。

通过以API优先、无限扩展性、可视化内容编辑、发布预览、业务可定制、私有云部署和强有力的数据安全保障等特点，具有无限扩展性，无论企业的内容需求多么庞大，它都能够轻松应对，并提供稳定可靠的服务；

业务定制功能，根据企业的业务需求和特点，可以定制开发特定的功能和模块，以满足企业的个性化需求。同时，支持在私有云上进行部署，保证了数据的安全性和隐私性。在数据安全保障方面，采用了先进的数据加密技术和安全防护措施，确保用户数据的安全性和完整性。此外，还提供了强大的数据统计和分析功能，帮助企业更好地了解和掌握内容使用情况和趋势。

## Contful Admin API

基于Springboot 3而开发的管理API，与控制台项目配套使用。

## 更新代码

```shell
git clone git@github.com:contful/contful.git
cd contful
git submodule update --init --recursive
```

## 仓库

Contful 是一个 `multi-repo`， Contful 有如下代码仓库：

```
// 项目结构
contful (https://github.com/contful/contful)
├── api  (https://github.com/contful/api) // 一级子模块
└── dashboard (https://github.com/contful/admin-dashboard) // 一级子模块
    └── api (https://github.com/contful/admin-api) // 二级子模块
```

| 仓库                                                            | 描述                 | 状态      |
| --------------------------------------------------------------- | -------------------- | --------- |
| [contful](https://github.com/contful/contful)           | 包含服务API、管理API、控制台的仓库集合       | `开发中` |
| [api](https://github.com/contful/api)           | 提供给客户端使用的服务API       | `开发中` |
| [dashboard](https://github.com/contful/admin-dashboard) | 控制台前端页面       | `开发中` |
| [admin-api](https://github.com/contful/admin-api)       | 与控制台配套管理API    | `开发中` |

## 为什么开源

通过开源，Contful 期望持续打磨出更加完善易用的内容管理系统，收获丰富的生态。

借助社区，Contful 期望与更多产品设计和开发者持续交流，成为更加有价值的产品。

对 Contful 而言，开源是一个新的起点。

## 贡献者

❤️ 感谢 Contful 所有的贡献者：

## 参与贡献

## 开源协议

Contful 使用 [Apache2.0 协议](./LICENSE)。