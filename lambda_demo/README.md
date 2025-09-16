# 🔧 Eino Lambda 组件演示

这个演示程序展示了 Eino 框架中 Lambda 组件的各种用法和交互模式，是一个完整的 Lambda 组件学习和实践项目。

## 🚀 快速开始

### 🛠️ 环境要求
- Go 1.19 或更高版本
- Eino 框架 v0.4.4 或更高版本

### ⚡ 快速运行
```bash
# 进入演示目录
cd lambda_demo

# 安装依赖
go mod tidy

# 运行所有演示
go run main.go

# 运行特定演示
go run main.go invokable     # InvokableLambda演示
go run main.go streamable    # StreamableLambda演示
go run main.go collectable   # CollectableLambda演示
go run main.go chain         # Chain集成演示
```

---

## 📖 什么是 Lambda？

`Lambda` 是 Eino 框架中**最基础的组件类型**，用于在工作流中嵌入自定义函数逻辑。它是一个函数包装器，让你可以将普通的 Go 函数集成到 Eino 的 Chain 和 Graph 工作流中。

### 🎯 核心特点
- **🔧 函数包装**: 将普通函数包装成 Eino 组件
- **🌊 多种交互模式**: 支持同步调用、流处理、数据收集和转换
- **⚡ 类型安全**: 支持泛型，确保输入输出类型安全
- **🧩 灵活集成**: 可与其他 Eino 组件无缝组合
- **🎭 零开销抽象**: 保持原生函数的执行效率

### 🔄 四种交互模式

1. **📝 Invoke**: 标准函数调用（输入→输出）
   ```go
   lambda := compose.InvokableLambda(func(ctx context.Context, input string) (string, error) {
       return strings.ToUpper(input), nil
   })
   ```

2. **🌊 Stream**: 返回流式输出（输入→多个输出）
   ```go
   lambda := compose.StreamableLambda(func(ctx context.Context, input int, output func(int)) error {
       for i := 1; i <= input; i++ {
           output(i)
       }
       return nil
   })
   ```

3. **📊 Collect**: 处理流式输入（多个输入→输出）
   ```go
   lambda := compose.CollectableLambda(func(ctx context.Context, inputs <-chan int) (int, error) {
       sum := 0
       for num := range inputs {
           sum += num
       }
       return sum, nil
   })
   ```

4. **⚡ Transform**: 流到流转换（多个输入→多个输出）
   ```go
   lambda := compose.TransformableLambda(func(ctx context.Context, inputs <-chan int, output func(int)) error {
       for num := range inputs {
           output(num * num)
       }
       return nil
   })
   ```

---

## 🎭 演示内容详解

### 1. 📝 InvokableLambda - 标准函数调用
**功能**: 最常用的 Lambda 类型，实现一对一的输入输出转换

**演示场景**:
- **用户数据处理**: 包含ID生成、分类逻辑、元数据管理
- **文本格式化**: 清理、大小写转换、格式标准化

**核心价值**:
- 封装业务逻辑，提高代码复用性
- 支持复杂数据结构的处理和转换
- 与 Chain 工作流无缝集成

```go
// 示例：用户数据处理器
userProcessor := compose.InvokableLambda(func(ctx context.Context, input UserInput) (*ProcessedData, error) {
    // 生成唯一ID
    id := fmt.Sprintf("user_%d_%s", time.Now().Unix(),
        strings.ReplaceAll(strings.ToLower(input.Name), " ", "_"))

    // 用户分类逻辑
    category := "成年人"
    if input.Age < 18 {
        category = "未成年人"
    } else if input.Age >= 60 {
        category = "老年人"
    }

    return &ProcessedData{
        ID: id,
        UserInfo: fmt.Sprintf("%s (%d岁) 来自 %s", input.Name, input.Age, input.City),
        Category: category,
        // ...更多处理逻辑
    }, nil
})
```

**适用场景**:
- 数据转换和格式化
- 业务规则计算
- 输入验证和清理

### 2. 🌊 StreamableLambda - 流式输出
**功能**: 从单个输入生成多个输出的流式处理

**演示场景**:
- **数字序列生成**: 展示如何生成连续数据流
- **任务分解**: 将复杂任务分解为多个子任务

**核心价值**:
- 支持数据的分解和批量生成
- 适合分页处理和数据扩展场景
- 可控制输出速率，避免系统过载

```go
// 示例：任务分解器
taskDecomposer := compose.StreamableLambda(func(ctx context.Context, user UserInput, output func(string)) error {
    tasks := []string{
        fmt.Sprintf("validate_email:%s", user.Email),
        fmt.Sprintf("check_age:%d", user.Age),
        fmt.Sprintf("normalize_name:%s", user.Name),
        fmt.Sprintf("geocode_city:%s", user.City),
    }

    for _, task := range tasks {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            output(task)
        }
    }
    return nil
})
```

**适用场景**:
- 数据分解和任务分配
- 批量数据生成
- 分页处理和流式输出

### 3. 📊 CollectableLambda - 流式收集
**功能**: 收集多个流式输入并生成单个聚合输出

**演示场景**:
- **数字统计**: 计算数字流的统计信息（计数、总和、平均值、最值）
- **文本聚合**: 收集多个文本片段并进行合并处理

**核心价值**:
- 实现数据的聚合和统计计算
- 支持实时数据分析
- 处理无限流数据，支持上下文取消

```go
// 示例：数字统计器
numberStats := compose.CollectableLambda(func(ctx context.Context, numbers <-chan int) (*StatResult, error) {
    var sum, count, min, max int
    min = math.MaxInt32

    for {
        select {
        case num, ok := <-numbers:
            if !ok {
                return &StatResult{
                    Count: count,
                    Average: float64(sum) / float64(count),
                    Sum: sum,
                    Min: min,
                    Max: max,
                }, nil
            }
            // 更新统计信息...
        case <-ctx.Done():
            return &StatResult{}, ctx.Err()
        }
    }
})
```

**适用场景**:
- 实时数据统计
- 流数据聚合
- 批处理和数据汇总

### 4. ⚡ TransformableLambda - 流到流转换
**功能**: 转换流式数据，支持流输入到流输出的实时处理

**演示场景**:
- **数字平方转换**: 实时计算数字流的平方值
- **文本处理流水线**: 多步骤文本处理（去空格→大写→添加前缀）

**核心价值**:
- 支持流数据的实时转换
- 可构建复杂的数据处理流水线
- 保持低内存占用，适合大数据量处理

```go
// 示例：文本处理流水线
textTransformer := compose.TransformableLambda(func(ctx context.Context, inputs <-chan string, output func(string)) error {
    for {
        select {
        case text, ok := <-inputs:
            if !ok {
                return nil
            }
            // 多步处理: 去空格 -> 大写 -> 添加前缀
            processed := strings.TrimSpace(text)
            processed = strings.ToUpper(processed)
            processed = fmt.Sprintf("[PROCESSED] %s", processed)
            output(processed)
        case <-ctx.Done():
            return ctx.Err()
        }
    }
})
```

**适用场景**:
- 实时数据转换
- 流水线处理
- 数据映射和变换

### 5. 🎭 AnyLambda - 多交互模式支持
**功能**: 单个组件同时支持多种交互模式的灵活处理器

**演示场景**:
- **多模式处理器**: 同一个处理器可以作为 Invokable 或 Streamable 使用
- **灵活性展示**: 展示 Lambda 的多样性和适应能力

**核心价值**:
- 提供最大的灵活性
- 减少代码重复
- 适应不同的使用场景

```go
// 示例：多模式处理器
type MultiModeProcessor struct {
    prefix string
}

func (p *MultiModeProcessor) Invoke(ctx context.Context, input string) (string, error) {
    return fmt.Sprintf("%s[INVOKE] %s", p.prefix, strings.ToUpper(input)), nil
}

func (p *MultiModeProcessor) Stream(ctx context.Context, input string, output func(string)) error {
    words := strings.Fields(input)
    for i, word := range words {
        result := fmt.Sprintf("%s[STREAM-%d] %s", p.prefix, i+1, strings.ToUpper(word))
        output(result)
    }
    return nil
}

// 创建支持多模式的 Lambda
anyLambda := compose.AnyLambda(&MultiModeProcessor{prefix: "🔧"})
```

**适用场景**:
- 需要多种调用方式的通用组件
- 适配不同工作流需求
- 灵活的处理逻辑封装

### 6. 🔗 Chain 集成演示
**功能**: 演示 Lambda 在完整工作流中的实际应用

**演示场景**:
- **完整用户数据处理工作流**: JSON解析 → 数据验证 → 业务处理 → 格式化输出
- **错误处理机制**: 展示如何在工作流中处理各种错误情况

**核心价值**:
- 展示 Lambda 的实际应用价值
- 构建端到端的数据处理流程
- 实现复杂业务逻辑的组件化

### 7. 🎓 高级用法演示
**功能**: 展示 Lambda 的高级特性和最佳实践

**演示场景**:
- **重试机制**: 带有重试逻辑的容错处理
- **性能监控**: 集成性能指标收集的 Lambda
- **动态配置**: 基于运行时参数调整行为

**核心价值**:
- 提供生产环境可用的解决方案
- 展示企业级应用的最佳实践
- 实现可观测性和可靠性

### 8. 🔬 性能测试演示
**功能**: 对比不同类型 Lambda 的性能表现

**演示场景**:
- **简单 vs 复杂处理**: 对比不同复杂度操作的性能
- **内存使用测试**: 测试内存密集型操作的表现
- **吞吐量分析**: 计算每种 Lambda 的处理能力

**核心价值**:
- 为性能优化提供数据支持
- 帮助选择合适的 Lambda 类型
- 验证系统的扩展能力

---

## 📊 输出示例

运行程序后会看到详细的演示输出：

```
🚀 Eino Lambda 组件演示程序
==================================================
📁 工作目录: /path/to/lambda_demo
🕗 启动时间: 2024-09-15T22:30:45+08:00

正在运行: InvokableLambda示例

=== InvokableLambda示例 ===
📝 示例1: 用户数据处理
输入: {Name:张三 Age:25 City:北京 Email:zhangsan@example.com}
输出: ID=user_1726408245_张三, 信息=张三 (25岁) 来自 北京, 分类=成年人

📄 示例2: 文本处理
输入: '  hello   WORLD  from  eino  '
输出: 'Hello World From Eino'
✅ InvokableLambda示例完成！

正在运行: StreamableLambda示例

=== StreamableLambda示例 ===
🔄 示例1: 数字序列生成
生成从1到5的数字流:
生成数字: 1
生成数字: 2
生成数字: 3
生成数字: 4
生成数字: 5
接收到的数字: 1 2 3 4 5

⚡ 示例2: 任务分解
分解用户处理任务，共5个子任务:
  任务1: validate_email:lisi@example.com
  任务2: check_age:30
  任务3: normalize_name:李四
  任务4: geocode_city:上海
  任务5: generate_avatar:李四
✅ StreamableLambda示例完成！

... [其他示例输出]

🎉 所有示例运行完成！

使用方法:
  go run main.go                  # 运行所有示例
  go run main.go invokable        # 运行InvokableLambda示例
  go run main.go streamable       # 运行StreamableLambda示例
  ...
```

---

## 🌟 应用场景

### 📊 数据处理管道
Lambda 组件特别适合构建数据处理管道：

- **JSON/XML 解析和转换**: 使用 InvokableLambda 进行结构化数据处理
- **数据验证和清洗**: 实现复杂的验证规则和数据清理逻辑
- **格式化和输出**: 生成标准化的输出格式

### 🌊 流式数据处理
支持各种流式数据处理场景：

- **实时数据转换**: 使用 TransformableLambda 处理流数据
- **批量数据聚合**: 使用 CollectableLambda 进行数据统计
- **事件流处理**: 处理实时事件和消息流

### 🏢 业务逻辑集成
灵活集成各种业务逻辑：

- **自定义业务规则**: 封装复杂的业务计算逻辑
- **算法集成**: 集成机器学习模型和算法
- **第三方服务调用**: 封装外部API调用

### 🔄 工作流编排
构建复杂的工作流系统：

- **复杂业务流程**: 使用 Chain 编排多步骤处理
- **条件分支处理**: 实现基于条件的处理分支
- **错误处理和重试**: 构建健壮的错误恢复机制

---

## 💡 最佳实践

### 1. 🎯 选择合适的 Lambda 类型
- **同步处理**: 使用 `InvokableLambda`
- **数据分解**: 使用 `StreamableLambda`
- **数据聚合**: 使用 `CollectableLambda`
- **流转换**: 使用 `TransformableLambda`
- **多模式**: 使用 `AnyLambda`

### 2. 🛡️ 错误处理
```go
lambda := compose.InvokableLambda(func(ctx context.Context, input string) (string, error) {
    // 输入验证
    if input == "" {
        return "", fmt.Errorf("输入不能为空")
    }

    // 业务处理
    result, err := processInput(input)
    if err != nil {
        return "", fmt.Errorf("处理失败: %w", err)
    }

    return result, nil
})
```

### 3. ⚡ 类型安全
```go
// 使用强类型定义
type UserProcessor = compose.Lambda

// 利用泛型确保类型安全
func CreateUserLambda() compose.Lambda {
    return compose.InvokableLambda(func(ctx context.Context, user UserInput) (*ProcessedData, error) {
        // 类型安全的处理逻辑
        return processUser(user)
    })
}
```

### 4. 🔄 资源管理
```go
streamLambda := compose.StreamableLambda(func(ctx context.Context, input int, output func(int)) error {
    for i := 0; i < input; i++ {
        select {
        case <-ctx.Done():
            return ctx.Err()  // 及时响应取消请求
        default:
            output(i)
        }
    }
    return nil
})
```

### 5. 🧪 测试友好
```go
// Lambda 函数应该是纯函数，便于测试
func processUserData(ctx context.Context, user UserInput) (*ProcessedData, error) {
    // 纯函数逻辑，无副作用
    return &ProcessedData{
        ID: generateID(user.Name),
        UserInfo: formatUserInfo(user),
    }, nil
}

// 包装为 Lambda
userLambda := compose.InvokableLambda(processUserData)
```

---

## 🚀 扩展示例

基于这个演示，你可以创建更复杂的 Lambda 组件：

### 📁 文件处理 Lambda
```go
fileProcessor := compose.InvokableLambda(func(ctx context.Context, filePath string) (*FileInfo, error) {
    // 读取、解析和转换文件
    data, err := ioutil.ReadFile(filePath)
    if err != nil {
        return nil, err
    }

    return &FileInfo{
        Path: filePath,
        Size: len(data),
        Hash: calculateHash(data),
    }, nil
})
```

### 🌐 网络请求 Lambda
```go
httpClient := compose.InvokableLambda(func(ctx context.Context, req *http.Request) (*http.Response, error) {
    client := &http.Client{Timeout: 30 * time.Second}
    return client.Do(req.WithContext(ctx))
})
```

### 🗄️ 数据库操作 Lambda
```go
dbQuery := compose.InvokableLambda(func(ctx context.Context, query string) ([]map[string]interface{}, error) {
    rows, err := db.QueryContext(ctx, query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    // 处理查询结果
    return processRows(rows)
})
```

### 🤖 机器学习 Lambda
```go
mlInference := compose.InvokableLambda(func(ctx context.Context, features []float64) (*PredictionResult, error) {
    // 模型推理
    prediction := model.Predict(features)

    return &PredictionResult{
        Prediction: prediction,
        Confidence: calculateConfidence(prediction),
    }, nil
})
```

---

## 🔗 相关资源

- 📚 [Eino Lambda 组件完全指南](./Lambda_Summary.md)
- 🌐 [官方文档](https://www.cloudwego.io/zh/docs/eino/core_modules/components/lambda_guide/)
- 🛠️ [GitHub 仓库](https://github.com/cloudwego/eino)
- 📖 [示例代码](./main.go)

---

## 🎉 总结

Lambda 组件的灵活性和强大功能使其成为构建复杂 AI 应用工作流的理想选择。通过这个完整的演示项目，你可以：

### 🏆 核心收获
- **掌握四种 Lambda 类型**的使用方法和适用场景
- **理解工作流编排**的最佳实践
- **学习错误处理和性能优化**技巧
- **获得生产环境可用**的代码模板

### 💪 实践能力
- 能够选择合适的 Lambda 类型解决具体问题
- 可以构建复杂的数据处理工作流
- 具备调试和优化 Lambda 组件的能力
- 掌握与其他 Eino 组件集成的方法

通过深入学习和实践这些示例，你将能够充分发挥 Eino Lambda 组件的强大能力，构建出高效、可靠的智能应用系统！🚀