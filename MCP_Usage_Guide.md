# MCP (模型、组件、平台) 使用指南

本文档旨在说明如何在软件开发中导入和使用不同类型的 MCP (Models, Components, Platforms)。

我们将 MCP 分为三大类：**平台/软件 (Platforms/Software)**、**组件/库 (Components/Libraries)** 和 **框架 (Frameworks)**。

---

## 1. 平台 / 软件 (Platforms / Software)

这类工具通常是独立的应用程序，你需要下载、安装，然后直接使用。

*   **涵盖的工具示例:**
    *   **测试/产品:** `Postman`, `JMeter`, `Jira`, `Confluence`, `Figma`, `Miro`
    *   **开发环境 (IDE):** `GoLand`, `PhpStorm`

*   **如何使用:**
    1.  **访问官网:** 在浏览器中搜索工具名称，找到官方网站。
    2.  **下载与安装:** 根据你的操作系统（Windows, macOS, Linux）下载对应的安装包并进行安装。
    3.  **注册/登录:** 大部分平台需要你注册一个账户才能使用，特别是协作类工具。
    4.  **学习使用:**
        *   **官方文档:** 这是最权威、最准确的学习资源。通常会有“快速入门”（Getting Started）指南。
        *   **图形用户界面 (GUI):** 你主要通过点击按钮、填写表单等方式与这些软件交互。例如，在 Postman 中，你可以在界面上输入 API 地址、设置参数并发起请求。

---

## 2. 组件 / 库 (Components / Libraries)

这类工具是代码级别的，你需要通过特定语言的包管理器将它们集成到你的项目中。

*   **涵盖的工具示例:**
    *   **Go:** `GORM`, `Cobra`, `Testify`
    *   **PHP:** `PHPUnit` (通过 Composer)
    *   **前端:** `React`, `Vue.js`, `Tailwind CSS`
    *   **测试:** `Selenium`, `Playwright`

*   **如何使用 (以 Go 和 前端 为例):**

    *   **Go 语言 (使用 Go Modules):**
        1.  **安装 (导入):** 在你的项目根目录下，打开终端，使用 `go get` 命令。
            ```bash
            # 示例：安装 GORM 和 SQLite 驱动
            go get -u gorm.io/gorm
            go get -u gorm.io/driver/sqlite
            ```
        2.  **在代码中使用:** 在你的 `.go` 文件中，使用 `import` 关键字引入包，然后就可以调用它提供的函数和结构体。
            ```go
            import (
              "gorm.io/gorm"
              "gorm.io/driver/sqlite"
            )

            func main() {
              // ... 使用 gorm.Open() 等函数
            }
            ```

    *   **前端 (使用 npm / yarn):**
        1.  **安装 (导入):** 在你的项目根目录下，打开终端，使用 `npm install` 或 `yarn add`。
            ```bash
            # 示例：安装 React
            npm install react react-dom

            # 示例：安装 Playwright 用于测试
            npm install --save-dev @playwright/test
            ```
        2.  **在代码中使用:** 在你的 JavaScript/TypeScript 文件中，使用 `import` 语句引入。
            ```javascript
            import React from 'react';
            import ReactDOM from 'react-dom/client';

            function App() {
              return <h1>Hello, React!</h1>;
            }
            // ...
            ```

    *   **PHP (使用 Composer):**
        1.  **安装 (导入):** 在项目根目录下，运行 `composer require`。
            ```bash
            # 示例：安装 PHPUnit
            composer require --dev phpunit/phpunit
            ```
        2.  **使用:** Composer 会自动处理类的加载（autoloading）。你只需要在代码中直接 `use` 相应的命名空间即可。

---

## 3. 框架 (Frameworks)

框架通常提供了一整套项目的骨架和规范，你不是“导入”它，而是**基于它来创建新项目**。

*   **涵盖的工具示例:**
    *   **Go:** `Gin` (既可作库，也可作框架)
    *   **PHP:** `Laravel`, `Symfony`
    *   **前端:** `React` (通过 `Create React App`), `Vue.js` (通过 `create-vue`)

*   **如何使用 (以 Laravel 和 Vite 为例):**

    *   **PHP - Laravel:**
        1.  **创建项目:** 你需要通过 Composer 来创建一个全新的 Laravel 项目。
            ```bash
            # 使用 Composer 创建一个名为 my-app 的新 Laravel 项目
            composer create-project laravel/laravel my-app
            ```
        2.  **进入目录并开发:**
            ```bash
            cd my-app
            php artisan serve # 启动开发服务器
            ```
        3.  **遵循框架规范:** 你将在框架预设好的目录结构（如 `app/Http/Controllers`, `routes/web.php`）中添加自己的业务逻辑。

    *   **前端 - Vite (用于启动项目):**
        1.  **创建项目:** 使用 npm/yarn/pnpm 的 `create` 命令。
            ```bash
            # 使用 npm 创建一个基于 Vue 的新 Vite 项目
            npm create vite@latest my-vue-app -- --template vue
            ```
        2.  **安装依赖并启动:**
            ```bash
            cd my-vue-app
            npm install
            npm run dev # 启动开发服务器
            ```
        3.  **开发:** 在 `src` 目录下进行组件开发，Vite 会提供极速的热更新体验。

---

## 总结

| 类型          | 如何“导入”                                            | 如何使用                             |
| :------------ | :---------------------------------------------------- | :----------------------------------- |
| **平台/软件** | 官网下载安装                                          | 打开软件，通过图形界面操作           |
| **组件/库**   | 使用包管理器 (`go get`, `npm install`, `composer require`) | 在代码中 `import` 或 `use`           |
| **框架**      | 使用官方脚手架命令创建新项目                          | 在框架预设的结构和规范下进行开发     |
