# Eino ADK 智能体开发演示

这是一个完整的 Eino ADK（Agent Development Kit）使用演示，展示了如何构建和组合多个智能体来处理不同类型的任务。

## 功能特性

### 🤖 多智能体系统
- **数学助手**：处理数学计算问题，集成计算器工具
- **生活助手**：提供生活服务，如天气查询
- **智能助理**：主控制器，自动路由到合适的子智能体

### 🔧 工具集成
- **计算器工具**：执行基础数学运算
- **天气查询工具**：获取城市天气信息

### ⚡ ADK 核心特性演示
- **统一接口**：所有 Agent 实现相同的接口规范
- **智能路由**：根据用户请求自动选择合适的处理器
- **事件监控**：完整的执行过程追踪
- **工作流组合**：展示顺序和并行执行模式

## 项目结构

```
adk_demo/
├── main.go                                    # 基础工具演示（简化版）
├── shared_types.go                            # 共享类型定义
├── authentic_adk_demo.go                      # 🌟 官方Agent接口真实演示
├── multi_agent_composition_demo.go            # 🌟 多Agent组合模式演示
├── resumable_agent_demo.go                    # 🌟 ResumableAgent中断恢复演示
├── examples/                                  # 完整演示程序集合
│   ├── corrected_official_demo.go             # 官方标准工具演示
│   ├── stable_extension_demo.go               # Agent 扩展机制演示（稳定版）
│   ├── agent_extension_demo.go                # Agent 扩展机制演示（完整版）
│   └── working_demo.go                        # 概念演示版本
├── go.mod                                     # Go 模块依赖
├── go.sum                                     # 依赖校验文件  
├── README.md                                  # 本文档
├── Eino_ADK_Guide.md                          # ADK架构完整指南
└── demo_data/                                 # 演示数据和检查点目录
```

## 快速开始

### 1. 安装依赖

```bash
cd adk_demo
go mod tidy
```

### 2. 运行演示

#### 🌟 官方ADK架构演示（推荐）

```bash
# 1. 基于官方Agent接口的真实演示 - 核心架构
go run authentic_adk_demo.go

# 2. 多Agent组合模式演示 - 任务转移、工具化、层次结构
go run multi_agent_composition_demo.go

# 3. ResumableAgent中断恢复机制 - 企业级可靠性
go run resumable_agent_demo.go
```

#### 🔧 基础工具演示

```bash
# 基础概念演示（简化版，快速入门）
go run main.go

# 官方标准工具演示（InvokableTool接口）
go run examples/corrected_official_demo.go

# Agent 扩展机制演示（稳定版）
go run examples/stable_extension_demo.go
```

## 演示内容

### 🌟 核心ADK架构特性

#### 1. Agent抽象接口演示 (authentic_adk_demo.go)
```
🤖 Agent标准接口:
- Name() 和 Description() 元信息方法
- Run() 异步事件流生成
- AgentInput 统一输入格式
- AsyncIterator[*AgentEvent] 流式输出

📡 事件流示例:
[15:04:05] MathAgent: 🧮 数学Agent开始处理任务...
[15:04:05] MathAgent: 💬 我计算了 25 + 17，结果是 42
[15:04:05] MathAgent: 🎬 动作: exit (数学计算任务完成)
```

#### 2. 多Agent组合模式 (multi_agent_composition_demo.go)
```
🔧 模式1: Agent作为工具使用
- 直接调用 Invoke() 方法同步返回结果

🔄 模式2: Agent间任务转移
- AgentAction.Type='transfer'
- 保持执行上下文连续性

🏗️ 模式3: 层次化Agent结构
- RunPath 显示调用层次: RouterAgent → MathAgent
- 主Agent协调多个子Agent
```

#### 3. 中断恢复机制 (resumable_agent_demo.go)
```
⏸️  中断流程:
[15:04:05] DocumentProcessingAgent: ⚠️ 检测到需要人工干预
[15:04:05] DocumentProcessingAgent: 🎬 动作: interrupt
[15:04:05] Runner: ⏸️ Runner处理了Agent中断，状态已保存

🔄 恢复流程:
[15:04:07] Runner: 🔄 Runner开始恢复会话
[15:04:07] DocumentProcessingAgent: 📥 从步骤 3 恢复文档处理
[15:04:07] DocumentProcessingAgent: ✅ 用户确认继续处理
```

### 🔧 工具系统特性

#### 基础工具演示 (main.go)
- **简化接口**：展示最基本的工具接口实现
- **JSON 数据交换**：参数解析和结果返回的标准流程
- **错误处理**：基本的参数验证和错误响应

#### 官方工具标准 (examples/corrected_official_demo.go)
- **真实 InvokableTool 接口**：完全符合官方 GitHub 仓库接口定义
- **生产就绪**：可直接用于 ReAct Agent 的完整实现
- **完整工具演示**：计算器和天气查询工具的标准实现

### 📚 学习路径建议

1. **ADK入门** (`authentic_adk_demo.go`)
   - 理解 Agent 接口的核心概念
   - 掌握异步事件流处理机制
   - 学习 AgentInput/AgentEvent 数据结构

2. **组合模式** (`multi_agent_composition_demo.go`)
   - 掌握三种Agent协作模式
   - 理解任务转移和层次化结构
   - 学习复杂业务流程的Agent编排

3. **企业应用** (`resumable_agent_demo.go`)
   - 理解中断恢复的业务价值
   - 掌握 ResumableAgent 接口设计
   - 学习 Runner 生命周期管理

4. **工具开发** (`main.go` + `examples/corrected_official_demo.go`)
   - 从简单到复杂的工具实现
   - 理解工具与Agent的关系
   - 掌握生产环境的开发标准

## 代码架构

### 🏗️ 核心ADK架构（基于官方接口）

#### 1. Agent抽象接口（官方 Agent 接口）
```go
// 核心Agent接口 - 严格按照官方定义
type Agent interface {
    Name(ctx context.Context) string
    Description(ctx context.Context) string
    Run(ctx context.Context, input *AgentInput, opts ...AgentRunOption) *AsyncIterator[*AgentEvent]
}

// Agent输入结构
type AgentInput struct {
    Messages        []*Message `json:"messages"`
    EnableStreaming bool       `json:"enable_streaming,omitempty"`
}

// Agent事件结构 - 事件驱动架构核心
type AgentEvent struct {
    AgentName string         `json:"agent_name"`
    RunPath   []string      `json:"run_path,omitempty"`
    Output    interface{}   `json:"output,omitempty"`
    Action    *AgentAction  `json:"action,omitempty"`
    Timestamp time.Time     `json:"timestamp"`
}
```

#### 2. ResumableAgent扩展（官方扩展接口）
```go
// 可恢复Agent接口 - 支持中断恢复
type ResumableAgent interface {
    Agent
    Resume(ctx context.Context, interruptInfo *InterruptInfo, 
           input *AgentInput, opts ...interface{}) *AsyncIterator[*AgentEvent]
}

// Runner生命周期管理 - 官方Runner机制
type Runner struct {
    config          *RunnerConfig
    checkpointStore CheckPointStore
}

func (r *Runner) Execute(ctx context.Context, agent Agent, input *AgentInput) *AsyncIterator[*AgentEvent]
func (r *Runner) Resume(ctx context.Context, sessionID string, newInput *AgentInput) *AsyncIterator[*AgentEvent]
```

#### 3. 真实项目依赖
```bash
# 安装核心框架
go get github.com/cloudwego/eino@latest

# 安装扩展组件（包含各种模型实现）
go get github.com/cloudwego/eino-ext@latest
```

#### 4. 模型配置示例
```go
import (
    "github.com/cloudwego/eino-ext/components/model/ark"
)

// 配置 ARK 模型（字节跳动的大模型服务）
chatModel, err := ark.NewChatModel(ctx, &ark.ChatModelConfig{
    APIKey: "your-ark-api-key",
    Model:  "doubao-seed-1-6-250615",
})
```

## 自定义扩展

### 添加新工具
1. 实现 `tool.InvokableTool` 接口（包含 `Info()` 和 `InvokableRun()` 方法）
2. 使用 `schema.NewParamsOneOfByParams()` 定义参数结构
3. 处理 JSON 格式的参数输入和输出
4. 在 ReAct Agent 配置中注册工具

### 集成真实模型
1. 安装 `github.com/cloudwego/eino-ext` 扩展包
2. 配置 ARK API Key 或其他支持的模型服务
3. 创建对应的 ChatModel 实例
4. 传递给 `react.AgentConfig.Model` 字段

### 工作流定制
- **顺序执行**：Agent A → Agent B → Agent C
- **并行执行**：多个 Agent 同时处理不同子任务
- **条件执行**：根据中间结果决定下一步执行路径

## 最佳实践

### 1. 单一职责原则
每个 Agent 专注处理特定类型的任务：
- 数学助手只处理计算问题
- 生活助手只处理生活服务

### 2. 工具复用
将常用功能封装成工具，可在多个 Agent 间复用：
- 计算器工具可被多个需要计算的 Agent 使用
- 天气工具可集成到各种生活服务 Agent 中

### 3. 错误处理
```go
if err != nil {
    return nil, fmt.Errorf("工具调用失败: %w", err)
}
```

### 4. 事件监控
通过事件处理器追踪系统运行状态：
```go
type DemoEventHandler struct{}

func (d *DemoEventHandler) OnEvent(ctx context.Context, event *callbacks.Event) error {
    switch event.Type {
    case "agent_start":
        fmt.Printf("🚀 Agent 开始执行: %s\n", event.Data["agent_name"])
    }
    return nil
}
```

## 🎯 技术特点

### 🌟 真实ADK架构优势

1. **官方接口标准化**
   - 严格遵循官方 Agent 抽象接口
   - AgentInput/AgentEvent 标准化数据流
   - AsyncIterator 异步事件流处理

2. **企业级可靠性**
   - ResumableAgent 中断恢复机制
   - Runner 生命周期管理
   - CheckPointStore 状态持久化

3. **灵活组合能力**
   - Agent作为工具使用
   - 任务转移和层次化结构
   - RunPath 完整调用链追踪

4. **事件驱动架构**
   - 实时事件流处理
   - AgentAction 标准化动作控制
   - 松耦合的模块化设计

### 🆚 与传统Agent开发对比

| 特性 | 传统方式 | Eino ADK |
|------|----------|----------|
| **接口标准** | 各自实现 | 官方统一接口 |
| **组合能力** | 硬编码集成 | 灵活的组合模式 |
| **可靠性** | 基础错误处理 | 中断恢复机制 |
| **可观测性** | 日志输出 | 结构化事件流 |
| **扩展性** | 修改核心代码 | 插件化架构 |

## 🚀 进阶学习

### 📖 深度理解ADK
1. **核心架构**：阅读 `Eino_ADK_Guide.md` 了解架构设计原理
2. **官方文档**：
   - [ADK 概述](https://www.cloudwego.io/zh/docs/eino/core_modules/eino_adk/outline/)
   - [Agent 抽象](https://www.cloudwego.io/zh/docs/eino/core_modules/eino_adk/agent_abstract/)
   - [Agent 扩展](https://www.cloudwego.io/zh/docs/eino/core_modules/eino_adk/agent_extension/)

### 🏭 生产环境集成
1. **安装完整框架**：
   ```bash
   go get github.com/cloudwego/eino@latest
   go get github.com/cloudwego/eino-ext@latest
   ```

2. **集成真实模型**：配置火山方舟等大语言模型服务
3. **添加监控**：集成日志、指标、链路追踪
4. **部署优化**：容器化、负载均衡、容错处理

### 🎨 自定义扩展
1. **实现新Agent**：基于 Agent 接口创建专业领域Agent
2. **添加新工具**：实现 InvokableTool 接口扩展能力
3. **定制Runner**：根据业务需求自定义生命周期管理
4. **扩展存储**：实现 CheckPointStore 接口支持不同存储后端

## ❓ 常见问题

**Q: 这些演示与官方文档有什么对应关系？**
A: 我们的演示严格基于官方文档实现：
- `authentic_adk_demo.go` 对应 [Agent 抽象](https://www.cloudwego.io/zh/docs/eino/core_modules/eino_adk/agent_abstract/)
- `multi_agent_composition_demo.go` 对应 [ADK 概述](https://www.cloudwego.io/zh/docs/eino/core_modules/eino_adk/outline/) 中的组合模式
- `resumable_agent_demo.go` 对应 [Agent 扩展](https://www.cloudwego.io/zh/docs/eino/core_modules/eino_adk/agent_extension/)

**Q: 如何集成真实的大语言模型？**
A: 使用 eino-ext 扩展包，配置火山方舟等模型服务：
```go
import "github.com/cloudwego/eino-ext/components/model/ark"
chatModel, err := ark.NewChatModel(ctx, &ark.ChatModelConfig{
    APIKey: "your-api-key",
    Model:  "doubao-seed-1-6-250615",
})
```

**Q: Agent与传统工具有什么区别？**
A: Agent是更高层的抽象：
- **工具**：实现特定功能，通过 InvokableTool 接口
- **Agent**：具有推理能力，可以组合多个工具，通过 Agent 接口

**Q: 中断恢复机制的实际应用场景？**
A: 企业级场景中非常重要：
- 长时间文档处理任务
- 需要人工审核的业务流程  
- 跨系统重启的任务持续性
- 复杂数据分析的断点续传

**Q: 如何扩展更多Agent类型？**
A: 实现 Agent 接口即可：
```go
type MyCustomAgent struct{}
func (m *MyCustomAgent) Name(ctx context.Context) string { ... }
func (m *MyCustomAgent) Description(ctx context.Context) string { ... }
func (m *MyCustomAgent) Run(ctx context.Context, input *AgentInput, opts ...interface{}) *AsyncIterator[*AgentEvent] { ... }
```

---

🎉 **开始你的 Eino ADK 智能体开发之旅！**

🌟 **这就是真正基于官方文档的 Eino ADK 完整演示！**