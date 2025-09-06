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
├── main.go                                    # 基础演示程序（简化版）⭐
├── shared_types.go                            # 共享类型定义
├── examples/                                  # 完整演示程序集合
│   ├── corrected_official_demo.go             # 官方标准工具演示 ⭐
│   ├── stable_extension_demo.go               # Agent 扩展机制演示（稳定版）⭐
│   ├── agent_extension_demo.go                # Agent 扩展机制演示（完整版）
│   └── working_demo.go                        # 概念演示版本
├── go.mod                                     # Go 模块依赖
├── go.sum                                     # 依赖校验文件  
├── README.md                                  # 本文档
└── demo_data/                                 # 演示数据目录
```

## 快速开始

### 1. 安装依赖

```bash
cd adk_demo
go mod tidy
```

### 2. 运行演示

```bash
# 🌟 推荐：基础概念演示（简化版，快速入门）
go run main.go

# 🌟 推荐：官方标准工具演示（基于真实 GitHub 仓库）
go run examples/corrected_official_demo.go

# 🌟 推荐：Agent 扩展机制演示（中断与恢复功能）
go run examples/stable_extension_demo.go

# Agent 扩展完整演示（包含更多细节）
go run examples/agent_extension_demo.go

# 概念演示版本
go run examples/working_demo.go
```

## 演示内容

### 基础功能测试
程序会自动运行以下测试用例：

1. **数学计算测试**
   ```
   用户：帮我计算 25 + 17
   助理：我帮你计算了表达式 '25 + 17'，计算结果: 42.00
   ```

2. **天气查询测试**
   ```
   用户：北京今天天气怎么样？
   助理：北京的天气：晴天，温度 25°C，微风
   ```

3. **通用对话测试**
   ```
   用户：你好，你能做什么？
   助理：👋 您好！我是智能助理，可以帮您：
         1. 🧮 数学计算 - 说"计算 2+3"或"帮我算算 10*5"
         2. 🌤️ 查询天气 - 说"北京天气如何"或"上海的天气"
   ```

### 高级特性演示

#### 基础版 (main.go)
- **工作流组合**：展示顺序执行（数学分析→结果验证→报告生成）
- **并行处理**：同时处理多个不同类型的任务
- **事件追踪**：实时显示 Agent 和工具的执行状态

#### 官方标准版 (corrected_official_demo.go) ⭐ **工具演示推荐**
- **真实 InvokableTool 接口**：完全符合官方 GitHub 仓库的接口定义
- **JSON 数据交换**：标准的参数传递和结果返回格式
- **完整工具演示**：计算器和天气查询工具的完整实现
- **错误处理**：包含参数验证、错误响应等完整处理
- **生产就绪**：所有代码都可以直接用于真实的 ReAct Agent

#### Agent 扩展版 (stable_extension_demo.go) ⭐ **扩展机制推荐**
- **中断与恢复**：完整的任务中断和断点续传功能
- **检查点管理**：基于 Gob 编码的高效状态序列化
- **生命周期管理**：从任务启动到完成的全程状态管理
- **实际场景模拟**：文档处理任务的真实中断恢复流程
- **企业级可靠性**：适用于长时间运行的复杂 AI 任务

## 代码架构

### 核心组件（基于真实 Eino GitHub 仓库）

#### 1. 工具定义（官方 InvokableTool 接口）
```go
type CalculatorTool struct{}

// 工具信息定义
func (c *CalculatorTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
    params := map[string]*schema.ParameterInfo{
        "expression": {
            Type:     schema.String,
            Desc:     "数学表达式，如 '25+17'",
            Required: true,
        },
    }
    return &schema.ToolInfo{
        Name:        "calculator",
        Desc:        "执行基础数学计算",
        ParamsOneOf: schema.NewParamsOneOfByParams(params),
    }, nil
}

// 工具执行逻辑
func (c *CalculatorTool) InvokableRun(ctx context.Context, paramsInJSON string, opts ...tool.Option) (string, error) {
    // 解析 JSON 参数并执行计算
    // 返回 JSON 格式的结果
}
```

#### 2. ReAct Agent 集成（真实使用方式）
```go
// 创建工具列表
tools := []tool.InvokableTool{
    &CalculatorTool{},
    &WeatherTool{},
}

// 创建 ReAct Agent（需要真实的 ChatModel）
config := &react.AgentConfig{
    Model: chatModel,  // 来自 eino-ext 的实际模型
    Tools: tools,
}

agent, err := react.NewAgent(ctx, config)
if err != nil {
    return err
}

// 运行对话
messages := []*schema.Message{
    schema.UserMessage("帮我计算 25 + 17"),
}
result, err := agent.Generate(ctx, messages)
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

## 技术特点

### ADK 核心优势
1. **标准化接口**：统一的 Agent 开发规范
2. **组合能力**：支持复杂的 Agent 组织结构
3. **事件系统**：完整的执行过程可观测性
4. **工具生态**：丰富的工具集成能力
5. **性能优化**：支持并发和流式处理

### 与传统开发方式对比
- ✅ **快速开发**：标准化组件，减少重复工作
- ✅ **易于维护**：模块化设计，降低复杂度
- ✅ **高度复用**：工具和 Agent 可在多个项目间复用
- ✅ **扩展性强**：插件化架构，轻松添加新功能

## 进阶学习

1. **深入了解 ADK 架构**：阅读 `Eino_ADK_Guide.md`
2. **探索更多组件**：查看其他 Eino 演示项目
3. **集成真实模型**：连接火山方舟等大语言模型服务
4. **生产环境部署**：添加日志、监控、错误恢复等功能

## 常见问题

**Q: 如何集成真实的大语言模型？**
A: 参考 `comprehensive_demo` 中的模型配置，使用火山方舟 API。

**Q: 可以添加更多工具吗？**
A: 可以！只需实现 `tool.InvokableTool` 接口即可。

**Q: 如何处理复杂的工作流？**
A: 使用 ADK 的 Workflow 功能，支持顺序、并行、条件执行等模式。

**Q: 性能如何优化？**
A: 利用 ADK 的并发能力，合理设计 Agent 职责分工，避免不必要的重复计算。

---

🎉 **开始你的 AI 智能体开发之旅吧！**