# 🔧 Eino Lambda 组件完全指南

## 🚀 快速开始

### 🛠️ 配置文件
Lambda 组件作为纯函数封装器，不需要额外的配置文件，可直接使用：
```go
import "github.com/cloudwego/eino/compose"

// 直接创建 Lambda 组件
lambda := compose.InvokableLambda(func(ctx context.Context, input string) (string, error) {
    return strings.ToUpper(input), nil
})
```

---

## 📖 基本介绍

`Lambda` 组件是 Eino 框架中**最基础的组件类型**，它的主要作用是将普通的 Go 函数包装成 Eino 组件，使其能够在 Chain 和 Graph 工作流中无缝集成。Lambda 在 AI 应用开发中扮演着**"自定义逻辑容器"**的角色。

### 🎯 核心价值

在传统的应用开发中，业务逻辑往往与框架紧密耦合。而 Lambda 组件让我们能够：

```
传统开发：业务逻辑 + 框架耦合 + 难以复用  ❌
Lambda：纯函数 + 组件化封装 + 灵活编排  ✅
```

### 🚀 主要应用场景

- **🔄 数据转换**: 在工作流中进行格式化、清洗和转换操作
- **✅ 业务逻辑**: 封装自定义算法、验证规则和业务计算
- **🌊 流式处理**: 支持流数据的生成、收集和转换操作
- **🔗 工作流集成**: 与其他 Eino 组件无缝组合构建复杂流程
- **🧩 函数式编程**: 支持纯函数风格的组件化开发
- **⚡ 灵活扩展**: 快速添加自定义处理逻辑到现有工作流

---

## 🔧 核心接口

`Lambda` 组件提供了四种不同的交互模式来适应各种场景：

### 四种交互模式

#### 1. 🔄 Invoke（同步调用）
- **功能**: 标准的函数调用模式，一对一输入输出
- **使用场景**: 数据转换、业务逻辑、同步处理
```go
lambda := compose.InvokableLambda(func(ctx context.Context, input InputType) (OutputType, error) {
    // 处理逻辑
    return result, nil
})
```

#### 2. 🌊 Stream（流式输出）
- **功能**: 从单个输入生成多个输出的流式处理
- **使用场景**: 数据分解、批量生成、分页处理
```go
lambda := compose.StreamableLambda(func(ctx context.Context, input InputType, output func(OutputType)) error {
    // 生成多个输出
    output(result1)
    output(result2)
    return nil
})
```

#### 3. 📊 Collect（流式收集）
- **功能**: 收集多个流式输入并生成单个输出
- **使用场景**: 数据聚合、统计计算、批处理
```go
lambda := compose.CollectableLambda(func(ctx context.Context, inputs <-chan InputType) (OutputType, error) {
    // 收集并处理输入流
    return aggregatedResult, nil
})
```

#### 4. ⚡ Transform（流到流转换）
- **功能**: 转换流式数据，支持流输入到流输出
- **使用场景**: 实时处理、数据映射、流水线处理
```go
lambda := compose.TransformableLambda(func(ctx context.Context, inputs <-chan InputType, output func(OutputType)) error {
    // 转换流数据
    for input := range inputs {
        result := transform(input)
        output(result)
    }
    return nil
})
```

---

## 🎭 AnyLambda 多模式支持

`AnyLambda` 允许单个组件同时支持多种交互模式：

```go
type MultiModeProcessor struct{}

func (p *MultiModeProcessor) Invoke(ctx context.Context, input string) (string, error) {
    return "同步处理: " + strings.ToUpper(input), nil
}

func (p *MultiModeProcessor) Stream(ctx context.Context, input string, output func(string)) error {
    words := strings.Fields(input)
    for _, word := range words {
        output("流处理: " + strings.ToUpper(word))
    }
    return nil
}

// 创建支持多模式的 Lambda
anyLambda := compose.AnyLambda(&MultiModeProcessor{})
```

---

## 🏗️ 创建和使用 Lambda

### 基础使用流程

```go
import (
    "context"
    "strings"
    "github.com/cloudwego/eino/compose"
)

// 1️⃣ 创建简单的 InvokableLambda
simpleLambda := compose.InvokableLambda(func(ctx context.Context, input string) (string, error) {
    return strings.ToUpper(input), nil
})

// 2️⃣ 在 Chain 中使用
chain := compose.NewChain[string, string]()
chain.AppendLambda(simpleLambda)

// 3️⃣ 编译和运行
runnable, err := chain.Compile(ctx)
if err != nil {
    log.Fatal("Chain 编译失败:", err)
}

result, err := runnable.Invoke(ctx, "hello world")
if err != nil {
    log.Fatal("运行失败:", err)
}

fmt.Println(result) // 输出: "HELLO WORLD"
```

### 🎯 实用示例集合

#### 复杂数据处理 Lambda
```go
type UserData struct {
    Name  string `json:"name"`
    Email string `json:"email"`
    Age   int    `json:"age"`
}

type ProcessedUser struct {
    ID          string `json:"id"`
    DisplayName string `json:"display_name"`
    Category    string `json:"category"`
    ProcessedAt string `json:"processed_at"`
}

// 用户数据处理 Lambda
userProcessor := compose.InvokableLambda(func(ctx context.Context, user UserData) (*ProcessedUser, error) {
    // 生成唯一ID
    id := fmt.Sprintf("user_%d_%s", time.Now().Unix(),
        strings.ReplaceAll(strings.ToLower(user.Name), " ", "_"))

    // 用户分类
    category := "adult"
    if user.Age < 18 {
        category = "minor"
    } else if user.Age >= 60 {
        category = "senior"
    }

    // 格式化显示名称
    displayName := fmt.Sprintf("%s <%s>", user.Name, user.Email)

    return &ProcessedUser{
        ID:          id,
        DisplayName: displayName,
        Category:    category,
        ProcessedAt: time.Now().Format(time.RFC3339),
    }, nil
})
```

#### JSON 解析和验证 Lambda
```go
// JSON 解析 Lambda
jsonParser := compose.InvokableLambda(func(ctx context.Context, jsonStr string) (map[string]interface{}, error) {
    var data map[string]interface{}
    if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
        return nil, fmt.Errorf("JSON 解析失败: %w", err)
    }

    // 验证必需字段
    requiredFields := []string{"name", "email"}
    for _, field := range requiredFields {
        if _, exists := data[field]; !exists {
            return nil, fmt.Errorf("缺少必需字段: %s", field)
        }
    }

    return data, nil
})
```

#### 数据聚合 StreamableLambda
```go
// 数据分解 Lambda - 将单个用户生成多个处理任务
taskGenerator := compose.StreamableLambda(func(ctx context.Context, user UserData, output func(string)) error {
    // 生成多个处理任务
    tasks := []string{
        fmt.Sprintf("validate_email:%s", user.Email),
        fmt.Sprintf("check_age:%d", user.Age),
        fmt.Sprintf("normalize_name:%s", user.Name),
        fmt.Sprintf("generate_avatar:%s", user.Name),
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

#### 数据统计 CollectableLambda
```go
type StatResult struct {
    Count   int     `json:"count"`
    Average float64 `json:"average"`
    Sum     int     `json:"sum"`
    Min     int     `json:"min"`
    Max     int     `json:"max"`
}

// 数字统计 Lambda
numberStats := compose.CollectableLambda(func(ctx context.Context, numbers <-chan int) (*StatResult, error) {
    var sum, count, min, max int
    min = math.MaxInt32
    max = math.MinInt32

    for num := range numbers {
        count++
        sum += num
        if num < min {
            min = num
        }
        if num > max {
            max = num
        }
    }

    if count == 0 {
        return &StatResult{}, nil
    }

    return &StatResult{
        Count:   count,
        Average: float64(sum) / float64(count),
        Sum:     sum,
        Min:     min,
        Max:     max,
    }, nil
})
```

---

## 🚀 编排集成最佳实践

Lambda 组件的真正威力在于与其他组件的协同编排。

### 🔗 Chain 编排模式

Chain 编排适合线性处理流程：

```go
// 完整的用户数据处理工作流
func createUserProcessingWorkflow() (*compose.Runnable, error) {
    ctx := context.Background()

    // 🔧 创建处理链
    chain := compose.NewChain[string, string]()

    // Step 1: JSON 解析
    parseJSON := compose.InvokableLambda(func(ctx context.Context, jsonStr string) (UserData, error) {
        var user UserData
        if err := json.Unmarshal([]byte(jsonStr), &user); err != nil {
            return UserData{}, fmt.Errorf("JSON 解析失败: %w", err)
        }
        return user, nil
    })

    // Step 2: 数据验证
    validateUser := compose.InvokableLambda(func(ctx context.Context, user UserData) (UserData, error) {
        if user.Name == "" {
            return UserData{}, fmt.Errorf("用户名不能为空")
        }
        if user.Age < 0 || user.Age > 150 {
            return UserData{}, fmt.Errorf("年龄无效: %d", user.Age)
        }
        if !strings.Contains(user.Email, "@") {
            return UserData{}, fmt.Errorf("邮箱格式无效: %s", user.Email)
        }
        return user, nil
    })

    // Step 3: 业务处理
    processUser := compose.InvokableLambda(func(ctx context.Context, user UserData) (*ProcessedUser, error) {
        // 实现用户处理逻辑（参考前面的示例）
        return userProcessor.Invoke(ctx, user)
    })

    // Step 4: 格式化输出
    formatOutput := compose.InvokableLambda(func(ctx context.Context, processed *ProcessedUser) (string, error) {
        output := fmt.Sprintf(`用户处理完成:
ID: %s
显示名: %s
分类: %s
处理时间: %s`,
            processed.ID, processed.DisplayName,
            processed.Category, processed.ProcessedAt)
        return output, nil
    })

    // 🔗 构建处理链
    chain.AppendLambda(parseJSON)
    chain.AppendLambda(validateUser)
    chain.AppendLambda(processUser)
    chain.AppendLambda(formatOutput)

    // ⚙️ 编译成可运行实例
    return chain.Compile(ctx)
}

// 使用示例
func processUserJSON(jsonStr string) (string, error) {
    workflow, err := createUserProcessingWorkflow()
    if err != nil {
        return "", fmt.Errorf("工作流创建失败: %w", err)
    }

    result, err := workflow.Invoke(context.Background(), jsonStr)
    if err != nil {
        return "", fmt.Errorf("用户处理失败: %w", err)
    }

    return result, nil
}
```

---

## ⚙️ 高级配置和选项

### Callback 机制

Lambda 组件支持在编排中使用回调机制：

```go
// 创建回调处理器
callbackHandler := callbacks.NewHandlerBuilder().
    OnStartFn(func(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
        fmt.Printf("🔧 开始 Lambda 处理: %v\n", input)
        return ctx
    }).
    OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) {
        fmt.Printf("✅ Lambda 处理完成: %v\n", output)
    }).
    OnErrorFn(func(ctx context.Context, info *callbacks.RunInfo, err error) {
        fmt.Printf("❌ Lambda 处理失败: %v\n", err)
    }).
    Build()

// 在 Chain 中使用回调
chain := compose.NewChain[string, string]()
chain.AppendLambda(myLambda, compose.WithCallbacks(callbackHandler))
```

### 上下文传递和状态管理

```go
type ProcessingContext struct {
    UserID    string
    RequestID string
    StartTime time.Time
}

// 带上下文的 Lambda
contextAwareLambda := compose.InvokableLambda(func(ctx context.Context, input string) (string, error) {
    // 从上下文中获取处理信息
    if procCtx, ok := ctx.Value("processing").(ProcessingContext); ok {
        duration := time.Since(procCtx.StartTime)
        return fmt.Sprintf("用户 %s 处理完成，耗时: %v, 请求ID: %s",
            procCtx.UserID, duration, procCtx.RequestID), nil
    }

    return "处理完成", nil
})

// 使用时设置上下文
ctx := context.WithValue(context.Background(), "processing", ProcessingContext{
    UserID:    "user_123",
    RequestID: "req_456",
    StartTime: time.Now(),
})
```

---

## 🎓 高级用法和技巧

### 1. 📊 动态 Lambda 工厂

根据不同场景动态创建 Lambda：

```go
type LambdaFactory struct {
    processors map[string]func() compose.Lambda
}

func NewLambdaFactory() *LambdaFactory {
    return &LambdaFactory{
        processors: make(map[string]func() compose.Lambda),
    }
}

func (f *LambdaFactory) RegisterProcessor(name string, creator func() compose.Lambda) {
    f.processors[name] = creator
}

func (f *LambdaFactory) CreateProcessor(name string) (compose.Lambda, error) {
    creator, exists := f.processors[name]
    if !exists {
        return nil, fmt.Errorf("未找到处理器: %s", name)
    }
    return creator(), nil
}

// 使用示例
factory := NewLambdaFactory()

// 注册不同类型的处理器
factory.RegisterProcessor("text_upper", func() compose.Lambda {
    return compose.InvokableLambda(func(ctx context.Context, input string) (string, error) {
        return strings.ToUpper(input), nil
    })
})

factory.RegisterProcessor("text_lower", func() compose.Lambda {
    return compose.InvokableLambda(func(ctx context.Context, input string) (string, error) {
        return strings.ToLower(input), nil
    })
})
```

### 2. 🔄 Lambda 组合模式

创建可复用的 Lambda 组合：

```go
// Lambda 组合器
type LambdaComposer struct {
    lambdas []compose.Lambda
}

func NewComposer() *LambdaComposer {
    return &LambdaComposer{
        lambdas: make([]compose.Lambda, 0),
    }
}

func (c *LambdaComposer) Add(lambda compose.Lambda) *LambdaComposer {
    c.lambdas = append(c.lambdas, lambda)
    return c
}

func (c *LambdaComposer) BuildChain() *compose.Chain[any, any] {
    chain := compose.NewChain[any, any]()
    for _, lambda := range c.lambdas {
        chain.AppendLambda(lambda)
    }
    return chain
}

// 使用示例
textProcessor := NewComposer().
    Add(trimSpaceLambda).
    Add(upperCaseLambda).
    Add(addPrefixLambda).
    BuildChain()
```

### 3. 📈 性能监控 Lambda

```go
type PerformanceMetrics struct {
    ProcessCount int64
    TotalTime    time.Duration
    ErrorCount   int64
    LastProcess  time.Time
}

func CreateMonitoredLambda[T, R any](
    name string,
    fn func(context.Context, T) (R, error),
    metrics *PerformanceMetrics,
) compose.Lambda {
    return compose.InvokableLambda(func(ctx context.Context, input T) (R, error) {
        startTime := time.Now()
        defer func() {
            metrics.ProcessCount++
            metrics.TotalTime += time.Since(startTime)
            metrics.LastProcess = time.Now()
        }()

        result, err := fn(ctx, input)
        if err != nil {
            metrics.ErrorCount++
        }

        return result, err
    })
}
```

### 4. 🛡️ 错误恢复 Lambda

```go
// 带重试机制的 Lambda
func CreateRetryableLambda[T, R any](
    fn func(context.Context, T) (R, error),
    maxRetries int,
    delay time.Duration,
) compose.Lambda {
    return compose.InvokableLambda(func(ctx context.Context, input T) (R, error) {
        var lastErr error

        for attempt := 0; attempt <= maxRetries; attempt++ {
            if attempt > 0 {
                select {
                case <-ctx.Done():
                    return *new(R), ctx.Err()
                case <-time.After(delay):
                }
            }

            result, err := fn(ctx, input)
            if err == nil {
                return result, nil
            }

            lastErr = err
            if !isRetryableError(err) {
                break
            }
        }

        return *new(R), fmt.Errorf("重试 %d 次后仍失败: %w", maxRetries, lastErr)
    })
}

func isRetryableError(err error) bool {
    // 判断错误是否可重试
    return !strings.Contains(err.Error(), "validation") &&
           !strings.Contains(err.Error(), "parsing")
}
```

---

## ❓ 常见问题和解决方案

### Q1: Lambda 函数中的类型安全问题

**问题**: 编译时类型检查不足，运行时类型转换错误
```go
// ❌ 错误做法：类型不匹配
lambda := compose.InvokableLambda(func(ctx context.Context, input string) (int, error) {
    return input, nil  // 类型错误
})
```

**解决方案**:
```go
// ✅ 正确做法：使用严格的类型定义
lambda := compose.InvokableLambda(func(ctx context.Context, input string) (int, error) {
    result, err := strconv.Atoi(input)
    if err != nil {
        return 0, fmt.Errorf("字符串转整数失败: %w", err)
    }
    return result, nil
})
```

### Q2: 流式 Lambda 的资源泄漏

**问题**: StreamableLambda 中忘记处理上下文取消
```go
// ❌ 可能导致资源泄漏
streamLambda := compose.StreamableLambda(func(ctx context.Context, input int, output func(int)) error {
    for i := 0; i < input; i++ {
        output(i)  // 没有检查上下文取消
        time.Sleep(time.Second)
    }
    return nil
})
```

**解决方案**:
```go
// ✅ 正确处理上下文取消
streamLambda := compose.StreamableLambda(func(ctx context.Context, input int, output func(int)) error {
    for i := 0; i < input; i++ {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            output(i)
            time.Sleep(time.Second)
        }
    }
    return nil
})
```

### Q3: CollectableLambda 的死锁问题

**问题**: 在 CollectableLambda 中不正确处理输入通道
```go
// ❌ 可能导致死锁
collectLambda := compose.CollectableLambda(func(ctx context.Context, inputs <-chan int) (int, error) {
    var sum int
    for {
        input := <-inputs  // 可能永远阻塞
        sum += input
    }
    return sum, nil
})
```

**解决方案**:
```go
// ✅ 正确处理通道关闭
collectLambda := compose.CollectableLambda(func(ctx context.Context, inputs <-chan int) (int, error) {
    var sum int
    for {
        select {
        case input, ok := <-inputs:
            if !ok {
                // 通道已关闭
                return sum, nil
            }
            sum += input
        case <-ctx.Done():
            return sum, ctx.Err()
        }
    }
})
```

### Q4: Lambda 性能优化

**问题**: Lambda 函数执行效率低下
```go
// ❌ 低效的实现
inefficientLambda := compose.InvokableLambda(func(ctx context.Context, input []string) ([]string, error) {
    var result []string
    for _, str := range input {
        // 每次都重新分配内存
        result = append(result, strings.ToUpper(str))
    }
    return result, nil
})
```

**解决方案**:
```go
// ✅ 优化的实现
efficientLambda := compose.InvokableLambda(func(ctx context.Context, input []string) ([]string, error) {
    // 预分配足够的容量
    result := make([]string, 0, len(input))

    // 使用 strings.Builder 进行字符串操作
    var builder strings.Builder

    for _, str := range input {
        builder.Reset()
        builder.Grow(len(str))  // 预分配容量
        builder.WriteString(strings.ToUpper(str))
        result = append(result, builder.String())
    }

    return result, nil
})
```

---

## 🎉 总结

Lambda 是 Eino 框架中的**万能组件**，掌握它的使用对于构建灵活高效的 AI 应用工作流至关重要：

### 🏆 核心优势
- 🔧 **灵活封装**: 将任意 Go 函数包装成 Eino 组件，支持复杂业务逻辑
- 🌊 **多模式支持**: 提供四种交互模式，适应不同的处理场景
- ⚡ **高性能**: 零开销抽象，保持原生函数的执行效率
- 🧩 **组件化**: 与 Eino 生态系统深度集成，构建完整工作流
- 🛡️ **类型安全**: 支持泛型和编译时类型检查，减少运行时错误
- 🔄 **易于测试**: 纯函数设计，便于单元测试和调试

### 💡 最佳实践总结
1. **合理选择**: 根据输入输出特性选择合适的 Lambda 类型
2. **错误处理**: 实施完善的错误检测、分类和恢复机制
3. **资源管理**: 正确处理上下文取消和资源释放
4. **性能优化**: 使用预分配、对象池等技术提升执行效率
5. **类型安全**: 利用泛型和类型断言确保类型安全
6. **编排集成**: 优先使用 Chain/Graph 编排构建自动化工作流

### 🔗 相关资源
- 📚 [官方文档](https://www.cloudwego.io/zh/docs/eino/core_modules/components/lambda_guide/)
- 🌐 [GitHub 仓库](https://github.com/cloudwego/eino)
- 🛠️ [示例代码](./lambda_demo.go)

通过掌握 Lambda 组件的各种功能和最佳实践，你将能够构建出更加灵活、高效和可维护的智能应用工作流！🚀