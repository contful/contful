# Contful

[简体中文](./README.md) | English

The main repository of Contful consists of Dashboard, Admin API, and Service API.

Contful is an open source headless CMS project that provides a new solution for enterprise-level content management with
its superior features and characteristics. As an abbreviation of Contentful, Contful combines the words "Content" and "
ful", symbolizing its commitment to creating a management platform full of rich, deep, and wide-ranging content.

Contful is API-centric, and through a powerful API interface, it implements a front-end and back-end separation content
management model, making content creation, editing, management, and publishing more efficient and flexible.

With features such as API prioritization, unlimited scalability, visual content editing, publishing preview, business
customization, private cloud deployment, and strong data security guarantees, it has unlimited scalability and can
easily respond to the content needs of enterprises, regardless of their size, and provide stable and reliable services.

Business customization function enables the development of specific functions and modules based on the business needs
and characteristics of enterprises to meet their personalized requirements. Meanwhile, it supports deployment on private
clouds, ensuring data security and privacy. In terms of data security, advanced data encryption technology and security
protection measures are adopted to ensure the security and integrity of user data. In addition, it provides powerful
data statistics and analysis functions to help enterprises better understand and grasp the usage and trends of content.

## Contful Admin API

Developed based on Springboot 3, the management API serves as a supporting service for the Dashboard.

## Clone Project

```shell
git clone git@github.com:contful/contful.git
cd contful
git submodule update --init --recursive
```

## Repositories

The following are the repositories for Contful:

```
// Project Struct
contful (https://github.com/contful/contful)
├── api (https://github.com/contful/api) // Primary Submodule
└── dashboard (https://github.com/contful/admin-dashboard) // Primary Submodule
    └── api (https://github.com/contful/admin-api) // Secondary Submodule
```

| Repositorie                                             | Desc                                                                              | Status       |
|---------------------------------------------------------|-----------------------------------------------------------------------------------|--------------|
| [contful](https://github.com/contful/contful)           | A collection of warehouses containing service APIs, management APIs, and consoles | `developing` |
| [api](https://github.com/contful/api)                   | Service APIs provided to clients for use                                          | `developing` |
| [dashboard](https://github.com/contful/admin-dashboard) | Console front-end page                                                            | `developing` |
| [admin-api](https://github.com/contful/admin-api)       | Matched with console management API                                               | `developing` |

## Why Open-source

Through open sourcing, Contful aims to refine a complete and user-friendly component library and cultivate a rich design
ecosystem.

With the help of the community, Contful aspires to maintain ongoing communication with more product designers and
developers, evolving into a more valuable product.

Open-source marks a new beginning for Contful.

## Contributors

❤️ Thanks for all the Contful contributors：

## Contributing Guide

## License

The Apache 2.0 License. Please see [the license file](./LICENSE) for more information.