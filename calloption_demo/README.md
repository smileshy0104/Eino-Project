# Eino CallOption 能力演示

## 📋 项目简介

这个演示项目展示了 Eino 框架中 CallOption (调用选项) 能力的实际使用方式，包括通用选项、特定实现选项、预设配置、动态选项选择以及节点特定配置等核心功能。

## 🎯 演示内容

### 1. 基础选项配置
- **功能**: 展示基础调用选项的使用方法
- **特点**: 直接使用 `WithTemperature`, `WithMaxTokens`, `WithTopP` 等通用选项
- **应用场景**: 简单的参数调整需求

### 2. 预设配置管理
- **功能**: 预定义的选项配置组合
- **实现**: `CreativeWritingPreset`, `TechnicalAnalysisPreset`, `ConversationalPreset`
- **特点**: 模块化的配置管理，易于复用
- **应用场景**: 标准化的任务处理配置

### 3. 动态选项选择
- **功能**: 根据请求内容动态选择最适合的配置
- **实现**: `SelectOptionsForTask()` 智能选择函数
- **特点**: 自适应的配置策略
- **应用场景**: 个性化的AI服务，智能参数调整

### 4. 节点特定配置
- **功能**: 为不同节点配置不同的处理参数
- **实现**: `DesignateNode()` 方法指定目标节点
- **特点**: 精确的粒度控制
- **应用场景**: 多角色AI协作，专业化分工

### 5. 选项组合
- **功能**: 灵活组合通用选项和特定实现选项
- **特点**: 通用选项 + Ark特有选项的混合使用
- **应用场景**: 充分利用组件的所有能力

## 🚀 运行演示

### 环境要求
- Go 1.24 或更高版本
- Eino 框架 v0.4.4

### 运行步骤

1. **进入演示目录**
   ```bash
   cd calloption_demo
   ```

2. **安装依赖**
   ```bash
   go mod tidy
   ```

3. **运行演示**
   ```bash
   go run main.go
   ```

### 预期输出

```
🎯 Eino CallOption 能力演示
===============================
✅ 配置加载成功
🎯 === Eino CallOption 能力演示 ===

1️⃣  基础选项配置演示
   展示基础调用选项的使用...
   🧪 处理任务: 请分析人工智能的发展趋势
🔧 [basic_processor] 使用配置: Temperature=0.5, MaxTokens=800, TopP=0.9, Cache=false
✅ 处理完成: ⚖️平衡版本: 请分析人工智能的发展趋势 [逻辑分析]
   📊 使用Token: 16, 处理时间: 234µs

2️⃣  预设配置演示
   展示预设配置的使用...
   🧪 测试用例 1: 创意写作
🔧 [preset_processor] 使用配置: Temperature=0.9, MaxTokens=2000, TopP=1.0, Cache=false
      ✅ 结果: ✨创意版本: 写一首关于春天的诗 [创意增强]
      📊 配置: T=0.9, Tokens=18, 用时=180µs
   🧪 测试用例 2: 技术分析
🔧 [preset_processor] 使用配置: Temperature=0.2, MaxTokens=1500, TopP=0.8, Cache=true
      ✅ 结果: 📊分析版本: 解释区块链技术原理 [逻辑分析] [缓存加速]
      📊 配置: T=0.2, Tokens=8, 用时=154µs

[更多输出内容...]

🎉 === 演示完成 ===
📚 这个演示展示了:
   • CallOption 基础概念和使用方法
   • 通用选项 vs 特定实现选项
   • 预设配置的模块化设计
   • 动态选项选择策略
   • 节点特定配置的精确控制
   • 选项的灵活组合和复用
```

## 📁 项目结构

```
calloption_demo/
├── main.go          # 完整演示程序
├── go.mod          # Go 模块依赖
└── README.md       # 项目说明文档
```

## 🧩 核心组件说明

### CallOption 接口定义

```go
type CallOption interface {
    Type() string                     // 选项类型
    Value() interface{}               // 选项值
    Apply(ctx context.Context) context.Context  // 应用到上下文
}
```

### 通用选项

- **`CommonChatModelOption`** - 通用的聊天模型选项
  - `WithTemperature(float64)` - 生成温度控制
  - `WithMaxTokens(int)` - 最大Token数限制
  - `WithTopP(float64)` - 核心采样参数
  - `WithStopWords([]string)` - 停止词设置

### 特定实现选项

- **`ArkSpecificOption`** - 火山方舟特有选项
  - `WithArkCache(bool)` - 缓存策略
  - `WithArkRetryCount(int)` - 重试次数
  - `WithArkTimeout(duration)` - 超时设置

### 业务处理器

- **`IntelligentTextProcessor`** - 智能文本处理器
  - 支持从上下文提取配置
  - 根据配置模拟不同的处理逻辑
  - 提供详细的处理统计信息

## 💡 关键特性展示

### 1. 动态配置能力
CallOption 允许在运行时根据不同的业务需求动态调整模型参数，无需重新编译或重启系统。

### 2. 精确的作用域控制
支持全局、组件类型、特定节点三个层级的配置控制，满足不同粒度的需求。

### 3. 类型安全设计
通过强类型系统确保配置的正确性，在编译时就能发现配置错误。

### 4. 可组合性
支持多种选项的自由组合，可以同时使用通用选项和特定实现选项。

### 5. 预设配置复用
通过预设配置模式，可以轻松管理和复用常用的配置组合。

## 🔧 扩展开发

### 自定义选项类型

```go
// 定义新的选项类型
type CustomOption struct {
    customField *string
    targetNode  string
}

func (o *CustomOption) Type() string {
    return "Custom"
}

func (o *CustomOption) Apply(ctx context.Context) context.Context {
    if o.customField != nil {
        ctx = context.WithValue(ctx, "custom_field", *o.customField)
    }
    return ctx
}

// 选项构造函数
func WithCustomField(value string) CallOption {
    return &CustomOption{customField: &value}
}
```

### 选项验证器

```go
func ValidateOptions(options []CallOption) error {
    for _, opt := range options {
        if err := validateSingleOption(opt); err != nil {
            return fmt.Errorf("invalid option %s: %w", opt.Type(), err)
        }
    }
    return nil
}
```

### 配置持久化

```go
func SaveConfiguration(name string, options []CallOption) error {
    config := map[string]interface{}{}
    for _, opt := range options {
        config[opt.Type()] = opt.Value()
    }
    return saveToFile(name, config)
}
```

## 📖 参考文档

- [Eino 官方文档](https://www.cloudwego.io/zh/docs/eino/)
- [CallOption 能力详细指南](../Eino_CallOption_Capabilities_Guide.md)
- [编排系统指南](../Eino_Orchestration_Guide.md)

## 🤝 贡献指南

欢迎提交 Issue 和 Pull Request 来改进这个演示项目！

## 📄 许可证

本项目遵循项目主许可证。