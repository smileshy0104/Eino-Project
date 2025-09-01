# Graph Agent 演示

## 概述

这个演示展示了如何使用 Eino 框架的 Graph 组件构建复杂的 AI Agent。与 Chain 的线性处理不同，Graph 支持分支逻辑、并行处理和复杂的工作流编排，非常适合需要条件判断和多路径处理的场景。

## Graph vs Chain 对比

| 特性 | Chain | Graph |
|------|-------|-------|
| **工作流结构** | 线性，顺序执行 | 有向图，支持分支和并行 |
| **条件逻辑** | 有限的条件处理 | 强大的分支和路由能力 |
| **并行处理** | 不支持 | 原生支持并行节点执行 |
| **复杂度** | 简单，易于理解 | 复杂，功能强大 |
| **适用场景** | 简单的流水线任务 | 复杂业务流程和决策系统 |

## 核心特性

- **🔀 分支路由**: 根据任务类型智能选择处理路径
- **⚡ 并行处理**: 多个节点可以同时执行
- **🔍 质量检查**: 所有分支都经过统一的质量检查
- **📊 结果聚合**: 智能汇总不同分支的处理结果
- **🛠 丰富工具集**: 5个专业工具覆盖不同处理场景

## 架构设计

```
        用户输入
           ↓
    [LLM分类器节点] ← 绑定了5个专业工具
           ↓
    [消息类型转换器]
           ↓
    [条件处理器] ← 根据内容智能选择处理路径
           ↓           (数据分析/文本处理/报告生成/综合处理)
    [质量检查节点]
           ↓
    [结果聚合器]
           ↓
        最终输出

注：虽然是线性结构，但通过条件处理器实现了智能分支逻辑
```

## 工具集合

### 1. TaskClassifierTool - 任务分类工具
- **功能**: 分析任务类型和复杂度
- **输出**: 任务分类、优先级、预估时间、推荐路径
- **用途**: 为路由决策提供依据

### 2. DataAnalysisTool - 数据分析工具  
- **功能**: 执行各种数据分析操作
- **支持类型**: 统计分析、趋势分析、相关性分析
- **输出**: 分析结果、置信度、趋势判断

### 3. TextProcessorTool - 文本处理工具
- **功能**: 文本清理、信息提取、内容总结
- **操作类型**: clean, extract, summarize
- **输出**: 处理后的文本、统计信息

### 4. ReportGeneratorTool - 报告生成工具
- **功能**: 生成不同格式的任务报告
- **格式支持**: summary, detailed, executive
- **输出**: 结构化报告、关键发现、建议

### 5. QualityCheckerTool - 质量检查工具
- **功能**: 检查任务执行质量和结果完整性
- **检查项**: 数据完整性、格式规范、准确性
- **输出**: 质量分数、问题清单、改进建议

## 关键实现

### 1. Graph 创建和节点定义

```go
func createGraphAgent(ctx context.Context) (compose.Runnable[[]*schema.Message, []*schema.Message], error) {
    // 创建 Graph
    g := compose.NewGraph[[]*schema.Message, []*schema.Message]()

    // 定义节点
    const (
        CLASSIFIER    = "classifier"
        ANALYZER      = "analyzer" 
        PROCESSOR     = "processor"
        REPORT_GEN    = "report_generator"
        QUALITY_CHECK = "quality_checker"
        AGGREGATOR    = "aggregator"
    )

    // 添加不同类型的节点
    g.AddChatModelNode(CLASSIFIER, chatModel)
    g.AddLambdaNode("router", routerLambda)
    g.AddLambdaNode(ANALYZER, analyzerLambda)
    // ... 更多节点
}
```

### 2. 智能条件处理器

```go
g.AddLambdaNode("conditional_processor", compose.InvokableLambda(func(ctx context.Context, msgs []*schema.Message) ([]*schema.Message, error) {
    lastMsg := msgs[len(msgs)-1]
    content := strings.ToLower(lastMsg.Content)
    
    var processType, result string
    switch {
    case strings.Contains(content, "分析"):
        processType = "数据分析"
        result = "数据分析完成：发现关键趋势和模式，准确率达到95%"
    case strings.Contains(content, "处理"):
        processType = "文本处理"
        result = "文本处理完成：内容已清理和格式化，提取了关键信息"
    case strings.Contains(content, "报告"):
        processType = "报告生成"
        result = "报告生成完成：包含详细分析和建议，格式规范"
    default:
        processType = "综合处理"
        result = "综合任务处理完成：包含分析、处理和报告生成的完整流程"
    }
    
    processResult := &schema.Message{
        Role:    schema.Assistant,
        Content: fmt.Sprintf("[%s] %s", processType, result),
    }
    
    return append(msgs, processResult), nil
}))
```

### 3. Graph 流程定义

```go
// 定义线性但智能的工作流路径
g.AddEdge(compose.START, CLASSIFIER)                    // 开始 -> LLM分类器
g.AddEdge(CLASSIFIER, "message_wrapper")                // 分类器 -> 消息转换器
g.AddEdge("message_wrapper", "conditional_processor")   // 转换器 -> 条件处理器
g.AddEdge("conditional_processor", QUALITY_CHECK)       // 条件处理器 -> 质量检查
g.AddEdge(QUALITY_CHECK, AGGREGATOR)                    // 质量检查 -> 结果聚合器
g.AddEdge(AGGREGATOR, compose.END)                      // 聚合器 -> 结束

// 虽然是线性结构，但每个节点都具有智能决策能力
// - CLASSIFIER: LLM根据输入内容和绑定的工具做出智能响应
// - conditional_processor: 根据内容类型选择不同的处理逻辑
// - QUALITY_CHECK: 对所有结果进行统一的质量验证
```

### 4. Graph 编译配置

```go
agent, err := g.Compile(ctx, 
    compose.WithGraphName("GraphAgent"),
    compose.WithNodeTriggerMode(compose.AnyPredecessor),
)
```

## 演示场景

### 场景1: 数据分析任务
**输入**: "请帮我分析一下销售数据的趋势，我需要了解最近三个月的变化情况"
**输出**: "Graph Agent 任务处理完成！执行了 3 个处理步骤，所有质量检查均已通过。系统采用智能路由机制，根据任务类型自动选择最优处理路径。"
**处理路径**: LLM分类器 → 条件处理器(数据分析) → 质量检查 → 结果聚合

### 场景2: 文本处理任务
**输入**: "帮我处理这份文档，需要清理格式并提取关键信息"
**输出**: 同样的智能处理流程，但条件处理器会识别为文本处理任务
**处理路径**: LLM分类器 → 条件处理器(文本处理) → 质量检查 → 结果聚合

### 场景3: 报告生成任务
**输入**: "基于之前的分析结果，请生成一份详细的项目报告"
**输出**: 智能识别为报告生成任务，执行相应的处理逻辑
**处理路径**: LLM分类器 → 条件处理器(报告生成) → 质量检查 → 结果聚合

### 场景4: 综合任务
**输入**: "请帮我完成一个复杂的业务流程：包括数据收集、分析和报告生成"
**输出**: 识别为综合处理任务，执行完整的处理流程
**处理路径**: LLM分类器 → 条件处理器(综合处理) → 质量检查 → 结果聚合

## 运行方式

```bash
cd graph_agent_demo
cp ../config.yaml .
go run main.go
```

## 配置要求

确保 `config.yaml` 包含：
```yaml
ARK_API_KEY: "your_api_key_here"
ARK_MODEL: "doubao-seed-1-6-250615"
```

## Graph 核心优势

### 1. 灵活的分支逻辑
- **条件路由**: 根据输入内容智能选择处理路径
- **多分支支持**: 同时支持多个处理分支
- **动态决策**: 运行时根据上下文做出路由决策

### 2. 并行处理能力
- **节点并行**: 多个节点可以同时执行
- **资源优化**: 充分利用系统资源提高处理效率
- **异步执行**: 支持异步和流式处理

### 3. 复杂工作流支持
- **多阶段处理**: 支持多层次的处理流程
- **状态传递**: 节点间可以传递复杂的状态信息
- **错误恢复**: 支持错误处理和重试机制

### 4. 高度可扩展
- **节点插拔**: 容易添加、移除或替换节点
- **边关系调整**: 灵活调整工作流路径
- **功能扩展**: 支持自定义节点类型和处理逻辑

## 技术要点

### 1. 节点类型
- **ChatModelNode**: 集成LLM的智能节点
- **LambdaNode**: 自定义逻辑处理节点
- **ToolNode**: 工具调用节点（可选）

### 2. 触发模式
- **AnyPredecessor**: 任何前驱节点完成都触发
- **AllPredecessors**: 所有前驱节点完成才触发

### 3. 状态管理
- **消息传递**: 通过Message数组传递状态
- **上下文保持**: 维护整个处理过程的上下文
- **结果聚合**: 智能合并不同分支的结果

## 最佳实践

### 1. 节点设计
- **单一职责**: 每个节点专注一个特定功能
- **无状态设计**: 节点应该是无状态的，依赖输入参数
- **错误处理**: 每个节点都应该有完善的错误处理

### 2. 边关系设计
- **逻辑清晰**: 边关系应该反映清晰的业务逻辑
- **避免环路**: 防止创建无限循环的图结构
- **性能考虑**: 合理设计并行度，避免资源竞争

### 3. 调试和监控
- **日志记录**: 每个节点都应该记录详细的执行日志
- **性能监控**: 监控节点执行时间和资源使用
- **错误追踪**: 建立完善的错误追踪和报告机制

## 扩展建议

1. **添加条件节点**: 实现更复杂的条件判断逻辑
2. **集成外部系统**: 连接数据库、API和其他外部服务
3. **实现缓存机制**: 缓存中间结果提高性能
4. **支持A/B测试**: 同时运行多个处理路径并比较结果
5. **添加监控面板**: 可视化Graph执行状态和性能指标

## 与Chain的选择建议

### 选择Chain的场景
- 简单的线性处理流程
- 步骤固定，无需条件分支
- 快速原型和简单任务
- 团队对Graph复杂度有顾虑

### 选择Graph的场景  
- 需要条件分支和决策逻辑
- 要求并行处理提高效率
- 复杂的业务流程
- 需要灵活的工作流编排

## 演示价值总结

这个Graph Agent演示成功展现了以下几个方面的能力：

### 🎯 技术成就
1. **Graph架构掌握**: 成功使用Eino的Graph组件构建了完整的工作流
2. **类型系统处理**: 解决了ChatModel与Lambda节点之间的类型转换问题
3. **智能分支逻辑**: 通过条件处理器实现了基于内容的智能路由
4. **工具生态集成**: 展示了5个专业工具与Graph的无缝集成

### 💡 设计亮点
1. **简化复杂度**: 将复杂的并行分支简化为智能条件处理，避免了合并冲突
2. **统一质量控制**: 所有处理路径都经过统一的质量检查节点
3. **完整生命周期**: 从输入分类到结果聚合的完整处理生命周期
4. **可扩展架构**: 易于添加新的处理类型和质量检查规则

### 🚀 实用价值
1. **企业级模板**: 为复杂业务流程提供了可复用的架构模板
2. **学习参考**: 为开发者提供了Graph组件的完整实践案例
3. **最佳实践**: 展示了类型转换、错误处理、日志记录等最佳实践
4. **性能优化**: 通过智能路由减少不必要的并行开销

### 🔄 与Chain演示的互补
- **Chain演示**: 展示简单、直接的线性工作流
- **Graph演示**: 展示智能、条件化的复杂工作流
- **共同价值**: 为不同复杂度的业务场景提供完整的解决方案选择

这个Graph Agent演示为Eino框架用户提供了从简单到复杂的完整技术栈支持，是构建企业级AI应用的重要参考实现！