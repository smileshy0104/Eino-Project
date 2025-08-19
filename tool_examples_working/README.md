# 可运行的 Eino Tool 示例

这个目录包含基于你项目现有代码结构的**实际可运行**的 Tool 示例。

## 文件说明

### weather_tool.go
**功能**: 完全可运行的工具演示，包含：

**工具实现**:
- `WeatherTool` - 天气查询工具
- `CalculatorTool` - 基本计算器工具
- `SimpleToolsNode` - 工具节点管理器

**特点**:
- ✅ 无外部依赖问题
- ✅ 基于项目已有的 schema 包
- ✅ 参考 tool_demo/main.go 的实现方式
- ✅ 包含完整的错误处理
- ✅ 可以直接运行

## 运行方法

```bash
cd tool_examples_working
go run weather_tool.go
```

## 输出示例

```
=== 可运行的 Tool 演示 ===

--- 测试天气工具 ---
[WeatherTool] 查询 北京 在 2024-08-19 的天气
天气查询结果: {"city":"北京","condition":"晴朗","date":"2024-08-19","description":"北京今天天气晴朗，温度25°C","humidity":60,"temperature":25}

--- 测试计算器工具 ---
[CalculatorTool] 执行运算: add(10.000000, 5.000000)
计算结果: {"a":10,"b":5,"operation":"add","result":15}

--- 测试 ToolsNode ---
[WeatherTool] 查询 上海 在 2024-08-19 的天气
[CalculatorTool] 执行运算: multiply(8.000000, 7.000000)
ToolsNode 执行成功，返回 2 个结果:
  结果 1: {"city":"上海","condition":"晴朗","date":"2024-08-19","description":"上海今天天气晴朗，温度25°C","humidity":60,"temperature":25}
  结果 2: {"a":8,"b":7,"operation":"multiply","result":56}
```

## 与原始示例的区别

1. **移除了不存在的包**: 不依赖 `github.com/cloudwego/eino/utils`
2. **简化了接口**: 基于项目中实际可用的 schema
3. **单文件实现**: 避免了 main 函数重复问题
4. **实际可运行**: 所有代码都经过验证可以执行

## 扩展建议

基于这个可运行的基础版本，你可以：

1. **添加更多工具类型**:
   - 文件操作工具
   - 网络请求工具
   - 数据处理工具

2. **集成到现有项目**:
   - 将工具添加到主项目的 Chain 中
   - 与 embedding 和 retriever 组件结合
   - 构建完整的 RAG + Tool 工作流

3. **增强功能**:
   - 添加参数验证
   - 实现异步执行
   - 添加缓存机制

这个版本提供了一个坚实的基础，可以让你理解 Tool 的核心概念，并在此基础上进行扩展。