# Claude Code 完整使用指南（基于完成对应安装）

## 目录
1. [简介](#简介)
2. [安装与配置](#安装与配置)
3. [基础使用](#基础使用)
4. [核心功能详解](#核心功能详解)
5. [高级特性](#高级特性)
6. [最佳实践](#最佳实践)
7. [常见问题解决](#常见问题解决)
8. [实际使用案例](#实际使用案例)

---

## 简介

### 什么是 Claude Code？
Claude Code 是 Anthropic 官方推出的终端AI编程助手，采用代理式开发模式，能够：
- 🤖 **智能编程**：通过自然语言描述需求，自动生成代码
- 🔍 **代码分析**：深度理解项目结构，识别问题并提供解决方案
- ⚡ **自动化操作**：修复 lint 错误、解决合并冲突、生成发布说明
- 🛠️ **直接操作**：编辑文件、运行命令、创建提交

### 核心优势
- **原生 Claude 4 模型**：使用最新的 Sonnet/Opus 模型
- **200K+ 上下文窗口**：支持大型项目的完整理解
- **无限工具调用**：可执行复杂的多步骤任务
- **终端原生**：遵循 Unix 哲学，可组合和脚本化
- **企业级安全**：内置安全和隐私保护

---

## 安装与配置

### 系统要求
- **Node.js**: 18+ 版本
- **操作系统**: macOS、Linux、Windows
- **Claude 账户**: Claude.ai 或 Claude Console 账户

### 1. 快速安装验证（30秒启动）

```bash
# 全局安装 Claude Code（NPM方式安装）
npm install -g @anthropic-ai/claude-code

# 验证安装
claude --version

# 导航到需要开发项目目录（或者直接在项目终端打开）
cd your-awesome-project

# 首次启动（会提示登录）
claude

# 按照提示使用您的账户登录
/login

# 
```

### 2. 详细配置系统

#### 配置系统概述
Claude Code 使用分层配置系统，优先级从高到低：
1. **企业管理策略** (企业级配置)
2. **命令行参数** (`--option value`)
3. **本地项目设置** (`.claude/settings.local.json`)
4. **共享项目设置** (`.claude/settings.json`)
5. **用户设置** (`~/.claude/settings.json`)
6. **默认值**

### 配置文件位置
```bash
# 用户设置
~/.claude/settings.json

# 项目设置
.claude/settings.json        # 共享项目设置（可提交到版本控制）
.claude/settings.local.json  # 个人项目设置（不应提交）
```

#### API 密钥和认证配置
```bash
# 方法一：环境变量（推荐）
export ANTHROPIC_API_KEY="sk-ant-api03-..."
echo 'export ANTHROPIC_API_KEY="sk-ant-api03-..."' >> ~/.zshrc
source ~/.zshrc

# 方法二：Claude Code 登录
claude
# 在界面中使用 /login 命令进行登录

# 方法三：配置文件设置（根据官方文档）
# 在用户设置文件中配置
~/.claude/settings.json
```

#### 权限配置
```bash
# Claude Code 支持细粒度权限控制
# 权限级别：allow（允许）、deny（拒绝）、ask（询问）

# 通过配置文件设置权限
# 在 .claude/settings.json 中配置工具权限
```

#### 配置文件结构
```json
// ~/.claude/settings.json 示例
{
  "permissions": {
    "tools": {
      "Read": "allow",
      "Write": "ask",
      "Bash": "deny"
    }
  },
  "model": {
    "name": "claude-3-sonnet-20240229"
  },
  "logging": {
    "enabled": true,
    "level": "info"
  },
  "hooks": {
    "user-prompt-submit": "echo 'Processing prompt...'"
  }
}
```

#### 命令行配置
```bash
# 查看所有可用配置
claude config list

# 基于官方文档的可用配置
claude config set -g theme "dark"              # 设置主题
claude config set -g verbose true              # 启用详细输出
claude config set -g todoFeatureEnabled true   # 启用待办功能

# 项目特定配置
claude config set permissions.tools.Read "allow"
claude config set permissions.tools.Write "ask"
```

#### 高级配置功能
```bash
# 子代理配置（官方文档提到的功能）
# 支持自定义提示的子代理

# 敏感文件排除
# 可配置排除敏感文件不被AI访问

# Hook 配置
# 支持Hook扩展工具功能

# 遥测设置
# 可配置日志和遥测选项
```


#### 常用配置命令
```bash
# 查看配置
claude config list                    # 所有配置
claude config ls                     # 简写形式
claude config get <key>              # 特定配置值
claude config list --verbose         # 详细信息（包括来源）
claude config ls -v                  # 简写详细信息

# 设置配置
claude config set <key> <value>      # 项目配置
claude config set -g <key> <value>   # 全局配置

# 删除配置
claude config remove <key>           # 删除配置
claude config rm <key>               # 简写形式
claude config remove -g <key>        # 删除全局配置

# 数组配置管理
claude config add <key> <values...>  # 添加到配置数组
claude config add -g allowed_tools "WebFetch" "Glob"  # 示例
claude config remove allowed_tools "Bash"             # 从数组删除

# 获取帮助
claude config --help                 # 主帮助
claude config help <command>         # 特定命令帮助
```


### 3. 基于官方文档的配置方法

#### 用户级别配置
```bash
# 创建或编辑用户设置文件
~/.claude/settings.json

# 示例用户配置
{
  "permissions": {
    "tools": {
      "Read": "allow",
      "Write": "ask",
      "Edit": "allow",
      "Bash": "ask"
    }
  },
  "model": {
    "name": "claude-3-sonnet-20240229"
  },
  "theme": "dark",
  "verbose": true
}
```

#### 项目级别配置
```bash
# 共享项目设置（可提交到版本控制）
.claude/settings.json

# 个人项目设置（不应提交）
.claude/settings.local.json

# 示例项目配置
{
  "permissions": {
    "tools": {
      "WebFetch": "allow",
      "Glob": "allow"
    }
  },
  "excludeFiles": [
    "*.env",
    "secrets.json",
    ".ssh/*"
  ],
  "hooks": {
    "user-prompt-submit": "echo 'Starting task...'"
  }
}
```

#### 权限控制配置
```bash
# 三种权限级别
# "allow" - 始终允许
# "deny" - 始终拒绝
# "ask" - 每次询问

# 工具权限示例
{
  "permissions": {
    "tools": {
      "Read": "allow",        # 始终允许读取文件
      "Write": "ask",         # 写入文件时询问
      "Bash": "deny",         # 禁止执行命令
      "WebFetch": "allow"     # 允许网络请求
    }
  }
}
```

#### Hook 扩展配置
```bash
# Hook 配置示例
{
  "hooks": {
    "user-prompt-submit": "echo 'Processing: $PROMPT'",
    "tool-call-pre": "echo 'About to call: $TOOL_NAME'",
    "tool-call-post": "echo 'Completed: $TOOL_NAME'"
  }
}
```

#### 配置方法对比

| 配置方式 | 用途 | 优先级 | 示例 |
|----------|------|--------|------|
| 环境变量 | API密钥等敏感信息 | 高 | `export ANTHROPIC_API_KEY="sk-..."` |
| 命令行参数 | 临时配置覆盖 | 高 | `claude --verbose` |
| 配置文件 | 持久化设置 | 中 | `~/.claude/settings.json` |
| Claude Config | 用户界面设置 | 低 | `claude config set theme "dark"` |

#### 推荐配置策略
```bash
# 1. 敏感信息用环境变量
export ANTHROPIC_API_KEY="sk-ant-api03-..."

# 2. 功能权限用配置文件
# 编辑 ~/.claude/settings.json
{
  "permissions": {
    "tools": {
      "Read": "allow",
      "Write": "ask",
      "Bash": "ask"
    }
  }
}

# 3. 界面设置用命令行
claude config set -g theme "dark"
claude config set -g verbose true
```

---

## 基础使用

### 1. 交互模式

```bash
# 启动交互式会话
claude

# 示例对话
> 帮我分析这个项目的结构
> 创建一个 Express 服务器，支持用户认证
> 修复所有的 TypeScript 错误
> 运行测试并生成报告
```

### 2. 单次查询模式

```bash
# 执行单个任务
claude -p "分析项目中的性能瓶颈"

# 复杂任务示例
claude -p "创建一个 React 组件，包含表单验证和提交功能"

# 自动化任务
claude -p "修复所有 ESLint 错误并格式化代码"
```

### 3. 会话管理

```bash
# 继续上一个会话
claude -c

# 恢复特定会话
claude -r <session-id>

# 查看会话历史
claude sessions list

# 删除会话
claude sessions delete <session-id>
```

---

## 核心功能详解

### 1. 项目记忆系统（CLAUDE.md）

Claude Code 会自动创建和维护 `CLAUDE.md` 文件来记住项目信息：

```markdown
# 项目记忆示例
## 项目概述
这是一个 React + TypeScript 项目

## 技术栈
- React 18
- TypeScript 4.9
- Vite 构建工具
- Tailwind CSS

## 最近活动
- 2024-01-15: 添加了用户认证功能
- 2024-01-14: 重构了组件结构
```

**优势**：
- 🧠 跨会话记忆项目状态
- 📝 自动记录重要决策
- 🔄 团队成员间共享上下文

### 2. 智能代码生成

#### 功能开发示例
```bash
# 用户输入
> 创建一个用户管理系统，包含增删改查功能

# Claude 的响应流程
1. 📋 制定计划
2. 🏗️ 创建数据模型
3. 🔧 实现 API 接口
4. 🎨 创建前端组件
5. ✅ 编写测试
6. 🚀 验证功能
```

#### 实际生成的代码结构
```
src/
├── models/
│   └── User.ts
├── api/
│   └── userApi.ts
├── components/
│   ├── UserList.tsx
│   ├── UserForm.tsx
│   └── UserDetail.tsx
└── tests/
    └── user.test.ts
```

### 3. 智能调试

#### 错误分析示例
```bash
# 当项目有错误时
> 我的应用启动失败了，帮我找出问题

# Claude 的调试流程
1. 🔍 分析错误日志
2. 📁 检查相关文件
3. 🔧 识别问题根源
4. ✨ 提供修复方案
5. 🧪 验证修复结果
```

#### 实际调试输出
```
❌ 发现问题: 缺少依赖 'express-session'
📝 解决方案:
1. 安装缺失依赖: npm install express-session
2. 更新类型定义: npm install -D @types/express-session
3. 修复导入语句: import session from 'express-session'

✅ 问题已解决，应用可以正常启动
```

### 4. 自动化任务

#### Git 集成
```bash
# 自动提交功能
> 帮我创建一个包含所有修改的 git 提交

# Claude 执行步骤
1. git add .
2. 分析修改内容
3. 生成语义化提交信息
4. git commit -m "feat: add user authentication system"
```

#### 代码质量
```bash
# 代码格式化和修复
> 修复所有代码质量问题

# 执行任务
1. npm run lint -- --fix
2. npm run format
3. 修复 TypeScript 错误
4. 更新文档
```

---

## 高级特性

### 1. MCP 服务器集成

#### 配置 MCP 服务器
```json
// claude_config.json
{
  "mcpServers": {
    "filesystem": {
      "command": "npx",
      "args": ["@modelcontextprotocol/server-filesystem", "/path/to/project"]
    },
    "git": {
      "command": "npx",
      "args": ["@modelcontextprotocol/server-git", "--repository", "."]
    }
  }
}
```

#### 使用 MCP 功能
```bash
# 启用 MCP 服务器
claude --mcp

# 示例任务
> 使用 git MCP 查看最近的提交历史并分析代码变更趋势
```

### 2. 自定义工具和环境

#### 创建自定义工具
```javascript
// tools/custom-tool.js
module.exports = {
  name: "deploy",
  description: "部署应用到服务器",
  execute: async (args) => {
    // 自定义部署逻辑
    return await deployApplication(args);
  }
};
```

#### 环境配置
```bash
# 设置开发环境
claude config set -g environment "development"

# 配置部署目标
claude config set -g deploy_target "staging"
```

### 3. 并发和批量操作

#### 多任务并行处理
```bash
# 同时执行多个任务
claude -p "并行执行以下任务：1)运行测试 2)构建生产版本 3)更新文档"

# 批量文件处理
claude -p "批量重构所有组件文件，添加 TypeScript 严格模式"
```

### 4. CI/CD 集成

#### GitHub Actions 集成
```yaml
# .github/workflows/claude-code.yml
name: Claude Code Automation
on: [push]
jobs:
  claude-tasks:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Run Claude Code
        run: |
          npm install -g @anthropic-ai/claude-code
          claude -p "分析代码质量并生成报告"
        env:
          ANTHROPIC_API_KEY: ${{ secrets.ANTHROPIC_API_KEY }}
```

---

## 最佳实践

### 1. 提示词优化

#### ✅ 好的提示词
```bash
# 具体明确
> 创建一个 React 用户注册组件，包含邮箱验证、密码强度检查和表单提交处理

# 提供上下文
> 基于现有的 Auth 系统，添加两步验证功能，需要兼容当前的用户数据结构

# 明确要求
> 重构 UserService 类，使用 TypeScript 泛型，确保类型安全，并添加完整的 JSDoc 注释
```

#### ❌ 避免的提示词
```bash
# 过于模糊
> 帮我写个网站

# 缺乏上下文
> 修复这个bug

# 要求不明确
> 让代码更好一些
```

### 2. 权限管理

#### 分层权限策略
```bash
# 开发环境（宽松权限）
claude config set -g dev_permissions "Read,Write,Edit,Bash,Git"

# 生产环境（严格权限）
claude config set -g prod_permissions "Read,Edit"

# 审查环境（只读权限）
claude config set -g review_permissions "Read,Glob,Grep"
```

#### 安全检查清单
- ✅ 定期审查 API 密钥权限
- ✅ 使用最小必要权限原则
- ✅ 监控工具使用情况
- ✅ 设置敏感目录访问限制

### 3. 成本优化

#### 有效使用策略
```bash
# 使用简洁明确的提示
claude -p "fix linting errors in src/components/"

# 避免重复会话
claude -c  # 继续现有会话而非创建新会话

# 批量处理任务
claude -p "run tests, fix errors, update docs"  # 一次性处理多个任务
```

### 4. 团队协作

#### 共享配置
```bash
# 创建团队配置文件
cat > .claude/team-config.json << EOF
{
  "coding_standards": "eslint-config-company",
  "test_framework": "jest",
  "deployment_target": "staging",
  "review_required": true
}
EOF

# 团队成员同步配置
claude config sync --team
```

#### 项目模板
```bash
# 创建项目模板
claude template create --name "react-ts-app" --path "."

# 使用模板
claude template apply "react-ts-app" --to "/new/project"
```

---

## 常见问题解决

### 1. 安装问题

#### Node.js 版本问题
```bash
# 检查 Node.js 版本
node --version

# 升级 Node.js (使用 nvm)
nvm install node
nvm use node

# 重新安装 Claude Code
npm uninstall -g @anthropic-ai/claude-code
npm install -g @anthropic-ai/claude-code
```

#### 权限问题
```bash
# macOS/Linux 权限修复
sudo chown -R $(whoami) $(npm config get prefix)/{lib/node_modules,bin,share}

# Windows 权限修复 (以管理员身份运行)
npm install -g @anthropic-ai/claude-code
```

### 2. 配置问题

#### API 密钥验证失败
```bash
# 验证 API 密钥
claude config test

# 重新设置 API 密钥
claude config set -g anthropic_api_key "new-api-key"

# 清除缓存
claude config clear-cache
```

#### 会话恢复失败
```bash
# 清理损坏的会话
claude sessions clean

# 重置配置
claude config reset

# 重新初始化
claude init
```

### 3. 性能问题

#### 响应速度慢
```bash
# 使用更快的模型
claude config set -g model "claude-3-haiku"

# 减少上下文长度
claude config set -g max_context_length 50000

# 启用缓存
claude config set -g enable_cache true
```

#### 内存使用过高
```bash
# 限制并发任务
claude config set -g max_concurrent_tasks 2

# 清理历史会话
claude sessions cleanup --older-than 7d
```

---

## 实际使用案例

### 案例1：全栈 Web 应用开发

#### 需求描述
创建一个博客管理系统，包含用户认证、文章管理、评论系统。

#### Claude Code 执行过程
```bash
# 第一步：项目初始化
> 创建一个 Next.js + TypeScript + Prisma 的博客项目

# Claude 执行：
1. 创建项目结构
2. 配置 TypeScript
3. 设置 Prisma 数据库
4. 配置身份验证
5. 创建基础组件

# 第二步：核心功能开发
> 实现文章的 CRUD 功能，支持 Markdown 编辑

# Claude 执行：
1. 创建 Prisma 模型
2. 实现 API 路由
3. 创建前端表单
4. 添加 Markdown 支持
5. 编写测试用例

# 第三步：部署优化
> 优化性能并准备生产部署

# Claude 执行：
1. 代码分割优化
2. 图片压缩处理
3. SEO 元数据配置
4. Docker 容器化
5. CI/CD 管道设置
```

#### 生成的项目结构
```
blog-management/
├── prisma/
│   ├── schema.prisma
│   └── migrations/
├── src/
│   ├── pages/
│   │   ├── api/
│   │   ├── admin/
│   │   └── blog/
│   ├── components/
│   │   ├── Auth/
│   │   ├── Blog/
│   │   └── UI/
│   ├── lib/
│   └── types/
├── tests/
├── Dockerfile
└── docker-compose.yml
```

### 案例2：遗留代码重构

#### 场景描述
重构一个老旧的 jQuery 项目到现代 React 架构。

#### 重构过程
```bash
# 分析现有代码
> 分析这个 jQuery 项目的结构，制定迁移到 React 的计划

# Claude 分析报告：
📊 代码分析结果：
- 总计 50 个 HTML 文件
- 35 个 JavaScript 文件
- 主要功能：用户管理、数据展示、表单处理
- 重构难点：DOM 操作、事件处理、状态管理

📋 迁移计划：
1. 创建 React 项目架构
2. 组件化页面结构
3. 重构数据流
4. 渐进式迁移
5. 测试验证

# 执行迁移
> 开始执行迁移计划，首先处理用户管理模块

# 迁移过程：
1. ✅ 创建 React 项目
2. ✅ 转换 HTML 模板为 JSX
3. ✅ 重构 jQuery 逻辑为 React Hooks
4. ✅ 迁移样式为 CSS Modules
5. ✅ 添加 TypeScript 类型
6. ✅ 编写单元测试
```

### 案例3：API 集成与测试

#### 需求描述
集成第三方支付API，确保安全性和可靠性。

#### 实现过程
```bash
# API 集成
> 集成 Stripe 支付 API，实现安全的支付流程

# Claude 实现：
1. 🔧 配置 Stripe SDK
2. 🔐 实现服务端支付逻辑
3. 🎨 创建前端支付组件
4. 🛡️ 添加安全验证
5. 🧪 编写测试用例

# 生成的核心文件
src/
├── services/
│   └── stripe.service.ts
├── components/
│   └── PaymentForm.tsx
├── api/
│   ├── payment/
│   │   ├── create-intent.ts
│   │   └── confirm-payment.ts
└── tests/
    └── payment.test.ts

# 安全性检查
> 检查支付系统的安全性，确保符合 PCI DSS 标准

# Claude 安全审计：
✅ API 密钥安全存储
✅ 服务端验证实现
✅ HTTPS 强制使用
✅ 输入数据验证
✅ 错误处理机制
✅ 审计日志记录
```

### 案例4：性能优化

#### 场景描述
优化一个 React 应用的性能问题。

#### 优化过程
```bash
# 性能分析
> 分析应用性能瓶颈，提供优化建议

# Claude 性能分析报告：
🔍 发现的问题：
1. 大型组件重复渲染
2. 未优化的图片加载
3. 过大的 JavaScript 包
4. 缺少代码分割
5. 内存泄漏问题

# 执行优化
> 根据分析结果执行性能优化

# 优化实施：
1. ⚡ 实现 React.memo 和 useMemo
2. 🖼️ 添加图片懒加载和压缩
3. 📦 配置 Webpack 代码分割
4. 🚀 实现路由级别的懒加载
5. 🧹 修复内存泄漏问题

# 性能提升结果：
📈 优化成果：
- 首次加载时间：3.2s → 1.1s (65% 提升)
- 包大小：2.1MB → 800KB (62% 减少)
- 内存使用：45MB → 28MB (38% 减少)
- 页面切换速度：800ms → 200ms (75% 提升)
```

---

## 总结

Claude Code 是一个功能强大的 AI 编程助手，能够显著提升开发效率和代码质量。通过本指南，你应该能够：

✅ **快速上手**：完成安装配置，开始基础使用
✅ **深度应用**：利用高级特性处理复杂项目
✅ **最佳实践**：遵循安全、高效的使用模式
✅ **问题解决**：独立解决常见问题

### 关键要点回顾
1. **安全第一**：始终使用最小权限原则
2. **明确提示**：提供具体、清晰的任务描述
3. **渐进使用**：从简单任务开始，逐步尝试复杂功能
4. **持续学习**：关注 Claude Code 的更新和最佳实践

### 进一步学习资源
- 📚 [Claude Code 官方文档](https://docs.claude.com/zh-CN/docs/claude-code)
- 💬 [社区讨论区](https://github.com/anthropics/claude-code/discussions)
- 🎯 [最佳实践示例](https://github.com/anthropics/claude-code-examples)
- 🔧 [工具扩展开发](https://docs.claude.com/zh-CN/docs/claude-code/tools)

---

*本文档版本：v1.0*
*最后更新：2024年1月*
*如有问题或建议，请联系：[issues](https://github.com/anthropics/claude-code/issues)*