# Eino Tool 演示示例

这个目录包含了 Eino 框架中 Tool 组件的各种使用示例，帮助理解如何创建和使用工具。

## 文件说明

### 1. basic_tool.go
**功能**: 演示如何手动实现 Eino Tool 接口的基本示例

**包含工具**:
- `CalculatorTool` - 基本数学运算（加减乘除）
- `TextProcessorTool` - 文本处理（大小写转换、长度计算、字符串反转）
- `MathTool` - 高级数学函数（sin、cos、sqrt、log等）

**特点**:
- 直接实现 `BaseTool` 和 `InvokableTool` 接口
- 手动处理 JSON 参数解析和结果序列化
- 完全控制工具的行为和错误处理

### 2. newtool_example.go
**功能**: 演示如何使用 NewTool 包装普通函数为 Eino Tool

**包含工具**:
- 加法运算工具 - 包装 `addNumbers` 函数
- 字符串格式化工具 - 包装 `formatString` 函数
- 数据验证工具 - 包装 `validateUserData` 函数
- 单位转换工具 - 包装 `convertUnits` 函数

**特点**:
- 使用 `utils.NewTool` 将普通函数转换为工具
- 需要手动定义 `ToolInfo` 结构
- 更简洁的函数实现，自动处理 JSON 序列化

### 3. infertool_example.go
**功能**: 演示如何使用 InferTool 自动从结构体标签推断工具信息

**包含工具**:
- 用户管理工具 - 创建用户账户
- 订单计算工具 - 计算订单总价和税费
- 数据分析工具 - 分析数值数据集
- 文本处理工具 - 文本统计和内容提取

**特点**:
- 使用 `utils.InferTool` 自动生成工具描述
- 通过 `jsonschema` 标签定义参数约束
- 最简洁的工具创建方式

### 4. streamable_tool.go
**功能**: 演示如何实现支持流式输出的 StreamableTool

**包含工具**:
- 流式文本生成工具 - 逐段生成文本内容
- 流式数据处理工具 - 批量处理数据并实时返回进度
- 流式日志分析工具 - 实时分析日志内容

**特点**:
- 实现 `StreamableTool` 接口
- 支持流式输出，适用于长时间处理任务
- 实时返回处理进度和中间结果

### 5. toolsnode_example.go
**功能**: 演示如何创建和使用 ToolsNode 来管理多个工具

**包含工具**:
- 天气查询工具 - 查询城市天气信息
- 计算器工具 - 数学表达式计算
- 翻译工具 - 文本翻译
- 文件管理工具 - 文件和目录操作

**特点**:
- 使用 `compose.NewToolsNode` 管理多个工具
- 支持单个和并行工具调用
- 演示在 Chain 中集成 ToolsNode
- 包含错误处理示例

## 使用方法

### 运行单个示例
```bash
cd tool_examples
go run basic_tool.go
go run newtool_example.go
go run infertool_example.go
go run streamable_tool.go
go run toolsnode_example.go
```

### 注意事项
1. 某些示例需要 `github.com/cloudwego/eino/utils` 包，如果包不存在，可能需要调整导入路径
2. 流式工具示例中的 `StreamReader` 实现是模拟的，实际使用时需要参考 Eino 框架的具体实现
3. 所有示例都包含详细的日志输出和错误处理

## 工具创建最佳实践

### 1. 选择合适的创建方式
- **手动实现**: 需要完全控制工具行为时使用
- **NewTool**: 已有函数逻辑，需要包装为工具时使用
- **InferTool**: 新开发工具，希望最简洁实现时使用

### 2. 参数设计
- 使用有意义的参数名称和描述
- 合理设置必需参数和可选参数
- 使用枚举值限制参数选项
- 提供默认值以提高易用性

### 3. 错误处理
- 提供清晰的错误消息
- 验证输入参数的有效性
- 优雅处理异常情况
- 返回结构化的错误信息

### 4. 结果格式
- 使用一致的 JSON 格式返回结果
- 包含足够的上下文信息
- 提供执行状态和时间戳
- 考虑结果的可读性和可解析性

## 扩展建议

1. **添加更多工具类型**: 网络请求、数据库操作、图像处理等
2. **实现工具链**: 将多个工具组合成复杂的工作流
3. **添加缓存机制**: 避免重复计算和网络请求
4. **集成外部服务**: 连接真实的 API 和服务
5. **性能优化**: 并行处理、连接池、资源管理等