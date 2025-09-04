# Eino 回调机制演示

## 📋 项目简介

这个演示项目展示了 Eino 框架中回调机制的实际使用方式，基于官方 callbacks 包的正确用法，包括直接回调触发、自定义组件回调、流式处理回调以及 Chain 编排中的回调集成。

## 🎯 演示内容

### 1. 直接使用 callbacks 包
- **功能**: 直接在业务逻辑中触发回调事件
- **实现**: `callbacks.OnStart()`, `callbacks.OnEnd()`, `callbacks.OnError()`
- **特点**: 手动控制回调时机和数据
- **应用场景**: 自定义组件开发、精细化监控控制

### 2. 自定义组件内置回调
- **功能**: 在自定义聊天模型中集成回调机制
- **实现**: `DemoChatModel` 实现 `model.ChatModel` 接口
- **特点**: 遵循 Eino 组件标准，自动触发回调
- **应用场景**: 自定义模型组件、第三方服务集成

### 3. 流式处理回调支持
- **功能**: 流式数据处理中的回调机制
- **实现**: `callbacks.OnEndWithStreamOutput()` 
- **特点**: 支持实时流式数据监控
- **应用场景**: 实时数据处理、流式AI模型调用

### 4. Chain 编排回调集成
- **功能**: Chain 编排中的自动回调触发
- **实现**: Chain 中嵌入带回调的组件
- **特点**: 编排级别的自动回调管理
- **应用场景**: 复杂业务流程监控、端到端追踪

## 🚀 运行演示

### 环境要求
- Go 1.24 或更高版本
- Eino 框架 v0.4.4

### 运行步骤

1. **进入演示目录**
   ```bash
   cd callback_demo
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

演示程序会按顺序展示四种回调注入方式：

```
🎯 Eino 回调机制完整演示
============================

🚀 === Eino 回调机制演示 ===

1️⃣  演示全局回调注入
   设置全局监控处理器...
🌐 [全局监控] 组件 [global_processor] 开始执行 (类型: Lambda)
   📝 请求详情: 用户=user_global, 优先级=3, 文本长度=20
✅ [全局监控] 组件 [global_processor] 执行成功
   📊 结果统计: Token数量=3, 处理时间=20ms, 质量分数=0.20
🎉 全局回调测试完成，结果: HELLO GLOBAL CALLBACK

2️⃣  演示Graph内回调注入
   创建带回调的Graph...
   🧪 测试请求 1 (优先级: 9)
⏱️  [性能分析] 开始监控组件 [urgent_processor]
📈 [性能分析] 组件 [urgent_processor] 执行耗时: 11ms
   📊 平均执行时间: 11ms (总计 1 次调用)
✅ 请求 1 处理完成: 【高优先级】URGENT TASK

📋 === 性能报告 ===
🔧 组件 [urgent_processor]: 调用 1 次, 平均 11ms, 最短 11ms, 最长 11ms
🔧 组件 [normal_processor]: 调用 1 次, 平均 13ms, 最短 13ms, 最长 13ms
🔧 组件 [low_processor]: 调用 1 次, 平均 17ms, 最短 17ms, 最长 17ms
================

3️⃣  演示Graph外回调注入(单组件)
   创建独立组件回调...
   🧪 执行独立组件测试...
📝 [业务日志] 记录组件 [standalone_text_processor] 开始执行
📝 [业务日志] 记录组件 [standalone_text_processor] 执行结果
✅ 独立组件执行成功: 【高优先级】STANDALONE COMPONENT TEST WITH DETAILED LOGGING
📄 详细日志已写入 callback_demo.log 文件

4️⃣  演示组合回调处理器
   创建组合回调处理器...
   🧪 执行组合回调测试...
   🔄 预处理步骤
   ⚙️  主处理步骤  
   ✨ 后处理步骤
🎉 组合回调测试完成: 【高优先级】【高优先级】【高优先级】COMPLEX MULTI-STEP PROCESSING WITH COMPOSITE CALLBACKS
📄 详细日志已写入 composite_demo.log 文件

🎉 === 演示完成 ===
📚 这个演示展示了:
   • 全局回调注入 - 系统级监控
   • Graph内回调注入 - 编排级监控  
   • Graph外回调注入 - 组件级监控
   • 组合回调处理器 - 多功能集成
   • 性能分析和业务日志记录
   • 横切面功能的无侵入集成
```

### 生成的文件

运行后会生成以下日志文件：

- **`callback_demo.log`** - 业务日志记录文件
- **`composite_demo.log`** - 组合回调日志文件

## 📁 项目结构

```
callback_demo/
├── main.go           # 主演示程序
├── go.mod           # Go模块定义
├── README.md        # 项目说明文档
├── callback_demo.log     # 业务日志输出(运行时生成)
└── composite_demo.log    # 组合回调日志(运行时生成)
```

## 🧩 核心组件说明

### 数据结构

- **`ProcessRequest`** - 处理请求结构
- **`ProcessResponse`** - 处理响应结构

### 回调处理器

1. **`GlobalMonitorHandler`** - 全局监控处理器
   - 监控所有组件的执行状态
   - 记录请求详情和执行结果
   - 实现系统级的统一监控

2. **`PerformanceAnalyzer`** - 性能分析处理器
   - 测量组件执行时间
   - 计算平均性能指标
   - 生成详细的性能报告

3. **`BusinessLogHandler`** - 业务日志处理器
   - 记录详细的业务执行日志
   - 支持文件输出
   - 适用于审计和调试

4. **`CompositeHandler`** - 组合处理器
   - 整合多个回调处理器
   - 支持动态添加处理器
   - 实现复杂的监控需求

### 业务组件

- **`TextProcessor`** - 文本处理组件
- **`SmartRouter`** - 智能路由组件

## 💡 关键特性展示

### 1. 无侵入性设计
回调机制不会影响业务逻辑的核心功能，所有监控和日志功能都通过回调实现。

### 2. 多层级监控
支持全局、编排、组件三个层级的回调注入，满足不同粒度的监控需求。

### 3. 可组合性
通过组合模式，可以将多个回调处理器组合使用，实现复杂的监控需求。

### 4. 性能友好
回调机制设计考虑了性能影响，支持异步处理和批量优化。

## 🔧 扩展开发

### 自定义回调处理器

创建自定义回调处理器只需要实现相应的回调接口：

```go
type MyCustomHandler struct {
    // 自定义字段
}

func (h *MyCustomHandler) OnStart(ctx context.Context, runInfo *callbacks.RunInfo, input interface{}) {
    // 开始执行时的处理逻辑
}

func (h *MyCustomHandler) OnEnd(ctx context.Context, runInfo *callbacks.RunInfo, output interface{}) {
    // 成功结束时的处理逻辑
}

func (h *MyCustomHandler) OnError(ctx context.Context, runInfo *callbacks.RunInfo, err error) {
    // 错误处理逻辑
}
```

### 集成外部监控系统

可以很容易地集成 Prometheus、Jaeger、ELK 等监控系统：

```go
type PrometheusHandler struct {
    metricsClient *prometheus.Client
}

func (h *PrometheusHandler) OnEnd(ctx context.Context, runInfo *callbacks.RunInfo, output interface{}) {
    h.metricsClient.Counter("component_success_total").
        With("component", runInfo.Name).
        Inc()
}
```

## 📖 参考文档

- [Eino 官方文档](https://www.cloudwego.io/zh/docs/eino/)
- [回调机制详细指南](../Eino_Callback_Manual_Guide.md)
- [编排系统指南](../Eino_Orchestration_Guide.md)

## 🤝 贡献指南

欢迎提交 Issue 和 Pull Request 来改进这个演示项目！

## 📄 许可证

本项目遵循项目主许可证。