# Claude Code Windows 原生版安装指南 🚀

这份指南将帮助完全没有程序设计经验的朋友，在 Windows 11 上直接安装 Claude Code 原生版本。

## 📋 开始前的准备

确认您有以下条件：
- Windows 11 电脑（或 Windows 10）
- Claude Pro 订阅（您已经有了！）
- 稳定的网络连接
- 管理员权限

## 🖥️ 步骤一：安装 VS Code

### 下载并安装

1. 前往官网：https://code.visualstudio.com/download
2. 下载 Windows 版本：点击蓝色的「Download for Windows」
3. 执行安装：
   - 双击下载的文件
   - 接受条款
   - **重要**：在安装选项中勾选：
     - ✅ 添加至 PATH
     - ✅ 在桌面建立图标
     - ✅ 添加「使用 Code 打开」动作

### 启动 VS Code

✅ **成功标志**：VS Code 成功打开，显示欢迎画面

## 📦 步骤二：安装 Node.js

### 1. 下载并安装 Node.js

1. 前往官网：https://nodejs.org
2. 下载 LTS 版本：点击绿色的「LTS」按钮（推荐版本）
3. 执行安装程序：
   - 双击下载的 .msi 文件
   - 点击「Next」接受所有默认设置
   - **重要**：确保勾选「Add to PATH」选项
   - 完成安装

### 2. 验证安装

1. 打开 VS Code
2. 按 `Ctrl + `` 打开终端（反引号键，在数字1的左边）
3. 确认环境：
   - 终端提示符应该显示类似 `C:\Users\您的用户名>`
   - 这表示您在 Windows CMD 环境中

4. 验证 Node.js 安装：
```cmd
node --version
npm --version
```

✅ **成功标志**：两个命令都会显示版本号（例如：v20.x.x 和 10.x.x）

### 3. 设置 VS Code 使用 CMD 终端（重要！）

由于 PowerShell 容易遇到权限问题，建议将 VS Code 默认终端改为 CMD：

1. 在 VS Code 中按 `Ctrl + Shift + P` 打开命令面板
2. 搜索并选择：`Terminal: Select Default Profile`
3. 选择 `Command Prompt` 而不是 `PowerShell`
4. 关闭现有终端：点击终端右上角的垃圾桶图标
5. 重新打开终端：按 `Ctrl + `` 打开新的 CMD 终端

💡 **为什么要这样做？** PowerShell 的执行策略可能会阻止某些 npm 命令运行，使用 CMD 可以避免大部分权限问题。

✅ **确认成功**：新终端的提示符应该显示 `C:\Users\您的用户名>`

## 🤖 步骤三：安装和设置 Claude Code

### 1. 安装 Claude Code

在 VS Code 的终端中输入：

💡 **重要说明**：确认终端提示符显示 `C:\Users\您的用户名>`（CMD 环境），而不是 `PS C:\Users\...>`（PowerShell 环境）。如果还是 PowerShell，请回到步骤二重新设置终端。

```cmd
npm install -g @anthropic-ai/claude-code
```

### 2. 首次设置和认证

1. **启动 Claude Code**：
```cmd
claude
```

2. **自动认证流程**：
   - Claude Code 会自动检测您需要认证
   - 选择「浏览器认证」（推荐选项）
   - 系统会自动打开浏览器或提供认证链接
   - 在浏览器中登录您的 Claude Pro 账户
   - 授权 Claude Code 访问您的账户
   - 浏览器会显示认证成功，您可以关闭浏览器回到终端

💡 **说明**：整个过程中您不需要手动创建或输入 API 密钥，Claude Code 会自动处理所有认证细节。

3. **测试功能**：
认证成功后，可以测试：
```
> 你好！我想开始学习程序设计
```

✅ **成功标志**：Claude Code 正常回应您的消息

## 🎯 步骤四：建立项目并开始使用

### 1. 建立项目文件夹

**在 Windows 中建立文件夹**：
- 在桌面或任何位置建立新文件夹，命名为「my-coding-journey」
- 或使用终端命令：

```cmd
mkdir C:\Users\%USERNAME%\Desktop\my-coding-journey
cd C:\Users\%USERNAME%\Desktop\my-coding-journey
```

**在 VS Code 中打开文件夹**：
1. 按 `Ctrl + K, Ctrl + O`（先按 Ctrl+K，放开，再按 Ctrl+O）
2. 或点击「File」→「Open Folder」
3. 选择您刚建立的 `my-coding-journey` 文件夹
4. 点击「选择文件夹」

### 2. 开始与 Claude Code 对话

1. 确认终端还在打开状态（如果没有，按 `Ctrl + `` 打开）
2. 确认在正确的文件夹中：
```cmd
cd C:\Users\%USERNAME%\Desktop\my-coding-journey
```

3. 启动 Claude Code：
```cmd
claude
```

4. **开始对话！**

## 🔐 步骤五：了解 Claude Code 权限管理

### 关于权限提示

当您开始与 Claude Code 协作时，系统会经常询问您是否同意执行某些操作，这是正常且重要的安全机制。

### 常见的权限请求

Claude Code 可能会要求以下权限：

1. **读取文件权限**：
```
> Claude 想要读取 "index.html" 文件，是否同意？ (y/n)
```

2. **写入/建立文件权限**：
```
> Claude 想要建立新文件 "style.css"，是否同意？ (y/n)
```

3. **执行命令权限**：
```
> Claude 想要执行命令 "npm install express"，是否同意？ (y/n)
```

4. **修改现有文件权限**：
```
> Claude 想要修改文件 "package.json"，是否同意？ (y/n)
```

### 如何安全地管理权限

**✅ 建议同意的情况**：
- 读取您项目文件夹中的文件
- 建立新的程序文件（如 .html, .css, .js 文件）
- 安装常见的开发套件（如 express, react 等）
- 修改您明确要求 Claude 帮助的文件

**⚠️ 谨慎考虑的情况**：
- 执行系统级别的命令（如修改系统设置）
- 访问项目文件夹以外的文件
- 安装您不熟悉的套件

**❌ 建议拒绝的情况**：
- 访问个人文件或系统重要文件
- 执行可能影响系统安全的命令
- 任何您觉得不理解的操作

### 实用技巧

1. **初期多确认**：刚开始使用时，建议每次都仔细阅读权限请求
2. **询问 Claude**：如果不确定某个操作是否安全，可以问 Claude：
```
> 请解释一下 "npm install express" 这个命令会做什么？
```
3. **项目文件夹隔离**：建议在专用的学习文件夹中进行实验，避免影响重要文件

### 权限提示范例对话

```
您: 帮我建立一个简单的网页
Claude: 我来帮您建立一个 HTML 文件
> Claude 想要在目前文件夹建立 "index.html"，是否同意？ (y/n)
您: y
> Claude 想要建立 "style.css" 文件，是否同意？ (y/n) 
您: y
Claude: 文件已建立完成！您可以在浏览器中打开 index.html 查看结果。
```

💡 **记住**：这些权限提示是为了保护您的电脑安全，在学习阶段建议多问多确认，随着经验增加，您会更容易判断哪些操作是安全的。

## 🎉 完成！开始您的程序设计之旅

现在您可以：

1. **向 Claude 请教**：
```
> 我想写一个简单的网页，该从什么开始？
```

2. **请 Claude 写程序**：
```
> 帮我写一个显示"Hello World"的 HTML 文件
```

3. **请 Claude 解释程序代码**：
```
> 请解释这段程序代码在做什么
```

4. **学习程序设计概念**：
```
> 什么是变量？可以用简单的例子说明吗？
```

## 🔧 常见问题解决

### 如果 Node.js 安装失败

1. 确认从官方网站 https://nodejs.org 下载
2. 以管理员身份执行安装程序
3. 确认网络连接正常
4. 重新启动电脑后再试

### 如果终端找不到 node 或 npm 命令

1. 确认安装时有勾选「Add to PATH」
2. 重新启动 VS Code
3. 重新启动电脑
4. 如果还是不行，可能需要手动添加到系统 PATH

### 如果 Claude Code 安装失败

1. **确认终端环境是 Windows CMD**（提示符显示 `C:\Users\...>`）
2. **如果是 PowerShell 环境**（提示符显示 `PS C:\Users\...>`）：
   - 按 `Ctrl + Shift + P` → 搜索 `Terminal: Select Default Profile` → 选择 `Command Prompt`
   - 关闭终端并重新打开

3. 确认 Node.js 和 npm 正常运作
4. 确认您的 Claude Pro 订阅状态
5. 检查网络连接是否正常

### 如果遇到 PowerShell 执行策略错误

- **错误消息可能类似**：「因为这个系统上已停用脚本执行...」
- **解决方法**：改用 CMD 终端（参考步骤二的设置）
- CMD 不会有 PowerShell 的执行策略限制

### 如果权限被拒绝

1. 尝试以管理员身份执行 VS Code
2. 确认在您有写入权限的文件夹中工作

### 需要更多帮助？

- Claude Code 官方文档：https://docs.anthropic.com
- 在 Claude Code 中输入 `/help` 查看命令

## 🆚 与 WSL 版本的比较

### Windows 原生版的优点：

- ✅ 安装更简单，不需要 WSL
- ✅ 直接在熟悉的 Windows 环境中工作
- ✅ 文件路径更直观（C:\ 格式）
- ✅ 与 Windows 系统整合更好

### 注意事项：

- 确保使用最新版本的 Claude Code
- 某些进阶功能可能仍在开发中
- 如果遇到问题，可以考虑回到 WSL 版本

## 恭喜！您现在已经准备好开始您的 vibe coding 之旅了！ 🎊

### 重要提醒：

- 每次要使用 Claude Code 时，在 VS Code 终端中输入 `claude` 就可以开始对话
- 所有的程序文件都会保存在您的项目文件夹中
- Windows 原生版本让整个流程更简单直接！

---

*最后更新: 2024年12月*
*版本: 2.0.0 - Windows 原生版专用*