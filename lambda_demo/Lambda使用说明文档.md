# Eino Lambda 组件使用说明文档

## 目录
- [概述](#概述)
- [核心概念](#核心概念)
- [API 接口](#api-接口)
- [使用方法](#使用方法)
- [最佳实践](#最佳实践)
- [常见场景](#常见场景)
- [错误处理](#错误处理)
- [性能优化](#性能优化)
- [示例代码](#示例代码)

## 概述

Lambda 是 Eino 框架中的核心组件，用于在工作流（Chain 和 Graph）中嵌入自定义函数逻辑。它提供了一种将普通 Go 函数包装成 Eino 组件的简洁方式，使开发者能够轻松构建复杂的数据处理管道。

### 主要特性
- **函数包装**：将普通函数转换为 Eino 组件
- **类型安全**：支持泛型，确保编译时类型检查
- **链式调用**：可与其他 Eino 组件无缝集成
- **灵活性**：支持任意输入输出类型
- **错误处理**：内置错误传播机制

## 核心概念

### Lambda 类型

#### InvokableLambda
最基础的 Lambda 类型，执行标准的输入输出转换。

**特点：**
- 一个输入 → 一个输出
- 同步执行
- 适用于大多数数据处理场景

**使用场景：**
- 数据格式转换
- 业务逻辑处理
- 数据验证
- 计算操作

### 函数签名
```go
func(ctx context.Context, input T) (output R, err error)
```

- `ctx`: 上下文，用于超时控制和取消操作
- `input`: 输入参数，类型为泛型 T
- `output`: 返回值，类型为泛型 R  
- `err`: 错误信息

## API 接口

### compose.InvokableLambda

```go
func InvokableLambda[T, R any](fn func(context.Context, T) (R, error)) *Lambda
```

**参数说明：**
- `T`: 输入类型（泛型）
- `R`: 输出类型（泛型）
- `fn`: 要包装的函数

**返回值：**
- `*Lambda`: Lambda 组件实例

## 使用方法

### 基本使用流程

#### 1. 创建 Lambda 函数
```go
// 定义处理函数
processData := compose.InvokableLambda(func(ctx context.Context, input string) (string, error) {
    // 业务逻辑
    result := strings.ToUpper(input)
    return result, nil
})
```

#### 2. 添加到 Chain
```go
// 创建处理链
chain := compose.NewChain[string, string]()
chain.AppendLambda(processData)
```

#### 3. 编译和执行
```go
// 编译链
runnable, err := chain.Compile(ctx)
if err != nil {
    return err
}

// 执行
result, err := runnable.Invoke(ctx, "hello")
```

### 完整示例

```go
package main

import (
    "context"
    "fmt"
    "log"
    "strings"
    
    "github.com/cloudwego/eino/compose"
)

func main() {
    ctx := context.Background()
    
    // 创建处理链
    chain := compose.NewChain[string, string]()
    
    // 添加文本清理 Lambda
    cleanText := compose.InvokableLambda(func(ctx context.Context, input string) (string, error) {
        return strings.TrimSpace(input), nil
    })
    
    // 添加大写转换 Lambda
    toUpper := compose.InvokableLambda(func(ctx context.Context, input string) (string, error) {
        return strings.ToUpper(input), nil
    })
    
    // 构建链
    chain.AppendLambda(cleanText)
    chain.AppendLambda(toUpper)
    
    // 编译并执行
    runnable, err := chain.Compile(ctx)
    if err != nil {
        log.Fatal(err)
    }
    
    result, err := runnable.Invoke(ctx, "  hello world  ")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println(result) // 输出: HELLO WORLD
}
```

## 最佳实践

### 1. 类型安全
```go
// 推荐：使用具体类型
type UserInput struct {
    Name string
    Age  int
}

type UserOutput struct {
    Info string
}

processUser := compose.InvokableLambda(func(ctx context.Context, input UserInput) (UserOutput, error) {
    return UserOutput{
        Info: fmt.Sprintf("%s is %d years old", input.Name, input.Age),
    }, nil
})
```

### 2. 错误处理
```go
validateInput := compose.InvokableLambda(func(ctx context.Context, input string) (string, error) {
    if input == "" {
        return "", fmt.Errorf("输入不能为空")
    }
    if len(input) > 100 {
        return "", fmt.Errorf("输入长度不能超过100字符")
    }
    return input, nil
})
```

### 3. 上下文使用
```go
processWithTimeout := compose.InvokableLambda(func(ctx context.Context, input string) (string, error) {
    select {
    case <-ctx.Done():
        return "", ctx.Err() // 处理取消或超时
    default:
        // 执行业务逻辑
        time.Sleep(100 * time.Millisecond) // 模拟处理时间
        return "processed: " + input, nil
    }
})
```

### 4. 资源管理
```go
processFile := compose.InvokableLambda(func(ctx context.Context, filename string) (string, error) {
    file, err := os.Open(filename)
    if err != nil {
        return "", err
    }
    defer file.Close() // 确保资源释放
    
    // 处理文件内容
    content, err := ioutil.ReadAll(file)
    if err != nil {
        return "", err
    }
    
    return string(content), nil
})
```

## 常见场景

### 1. 数据转换
```go
// JSON 解析
parseJSON := compose.InvokableLambda(func(ctx context.Context, jsonStr string) (map[string]interface{}, error) {
    var data map[string]interface{}
    err := json.Unmarshal([]byte(jsonStr), &data)
    return data, err
})

// 数据格式化
formatData := compose.InvokableLambda(func(ctx context.Context, data map[string]interface{}) (string, error) {
    return fmt.Sprintf("Name: %v, Age: %v", data["name"], data["age"]), nil
})
```

### 2. 数据验证
```go
validateUser := compose.InvokableLambda(func(ctx context.Context, user UserInput) (UserInput, error) {
    if user.Name == "" {
        return user, errors.New("姓名不能为空")
    }
    if user.Age < 0 || user.Age > 150 {
        return user, errors.New("年龄必须在0-150之间")
    }
    return user, nil
})
```

### 3. 业务逻辑处理
```go
calculateDiscount := compose.InvokableLambda(func(ctx context.Context, order Order) (Order, error) {
    if order.Amount > 1000 {
        order.Discount = order.Amount * 0.1 // 10% 折扣
    } else if order.Amount > 500 {
        order.Discount = order.Amount * 0.05 // 5% 折扣
    }
    order.FinalAmount = order.Amount - order.Discount
    return order, nil
})
```

### 4. 外部服务调用
```go
callExternalAPI := compose.InvokableLambda(func(ctx context.Context, request APIRequest) (APIResponse, error) {
    client := &http.Client{Timeout: 10 * time.Second}
    
    // 构建请求
    reqBody, _ := json.Marshal(request)
    req, err := http.NewRequestWithContext(ctx, "POST", "https://api.example.com", bytes.NewBuffer(reqBody))
    if err != nil {
        return APIResponse{}, err
    }
    
    // 发送请求
    resp, err := client.Do(req)
    if err != nil {
        return APIResponse{}, err
    }
    defer resp.Body.Close()
    
    // 解析响应
    var response APIResponse
    err = json.NewDecoder(resp.Body).Decode(&response)
    return response, err
})
```

## 错误处理

### 错误传播
```go
chain := compose.NewChain[string, string]()

// 如果任何一个 Lambda 返回错误，整个链会停止执行
step1 := compose.InvokableLambda(func(ctx context.Context, input string) (string, error) {
    if input == "" {
        return "", errors.New("步骤1: 输入为空")
    }
    return input + "-step1", nil
})

step2 := compose.InvokableLambda(func(ctx context.Context, input string) (string, error) {
    if len(input) > 50 {
        return "", errors.New("步骤2: 输入过长")
    }
    return input + "-step2", nil
})

chain.AppendLambda(step1)
chain.AppendLambda(step2)
```

### 错误恢复
```go
recoverableProcess := compose.InvokableLambda(func(ctx context.Context, input string) (string, error) {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("Lambda panic 恢复: %v", r)
        }
    }()
    
    // 可能发生 panic 的代码
    result := riskyOperation(input)
    return result, nil
})
```

### 自定义错误类型
```go
type ValidationError struct {
    Field   string
    Message string
}

func (e ValidationError) Error() string {
    return fmt.Sprintf("验证失败 [%s]: %s", e.Field, e.Message)
}

validateUser := compose.InvokableLambda(func(ctx context.Context, user User) (User, error) {
    if user.Email == "" {
        return user, ValidationError{Field: "email", Message: "邮箱不能为空"}
    }
    if !strings.Contains(user.Email, "@") {
        return user, ValidationError{Field: "email", Message: "邮箱格式无效"}
    }
    return user, nil
})
```

## 性能优化

### 1. 避免重复计算
```go
// 使用缓存
var cache = make(map[string]string)
var mu sync.RWMutex

cachedProcess := compose.InvokableLambda(func(ctx context.Context, input string) (string, error) {
    // 读取缓存
    mu.RLock()
    if result, exists := cache[input]; exists {
        mu.RUnlock()
        return result, nil
    }
    mu.RUnlock()
    
    // 计算结果
    result := expensiveOperation(input)
    
    // 写入缓存
    mu.Lock()
    cache[input] = result
    mu.Unlock()
    
    return result, nil
})
```

### 2. 并发控制
```go
// 使用信号量限制并发
semaphore := make(chan struct{}, 10) // 最多10个并发

limitedProcess := compose.InvokableLambda(func(ctx context.Context, input string) (string, error) {
    select {
    case semaphore <- struct{}: // 获取信号量
        defer func() { <-semaphore }() // 释放信号量
    case <-ctx.Done():
        return "", ctx.Err()
    }
    
    // 执行处理逻辑
    return processInput(input), nil
})
```

### 3. 内存优化
```go
// 处理大数据时使用流式处理
processLargeData := compose.InvokableLambda(func(ctx context.Context, input io.Reader) (Summary, error) {
    scanner := bufio.NewScanner(input)
    var summary Summary
    
    for scanner.Scan() {
        line := scanner.Text()
        summary.LineCount++
        summary.TotalLength += len(line)
        
        // 定期检查取消信号
        select {
        case <-ctx.Done():
            return summary, ctx.Err()
        default:
        }
    }
    
    return summary, scanner.Err()
})
```

## 调试和监控

### 1. 添加日志
```go
loggedProcess := compose.InvokableLambda(func(ctx context.Context, input string) (string, error) {
    start := time.Now()
    log.Printf("Lambda 开始处理: %s", input)
    
    result, err := businessLogic(input)
    
    duration := time.Since(start)
    if err != nil {
        log.Printf("Lambda 处理失败: %s, 耗时: %v, 错误: %v", input, duration, err)
    } else {
        log.Printf("Lambda 处理成功: %s, 耗时: %v", input, duration)
    }
    
    return result, err
})
```

### 2. 性能指标
```go
// 使用 Prometheus 监控
var (
    lambdaCounter = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "lambda_invocations_total",
            Help: "Total number of lambda invocations",
        },
        []string{"status"},
    )
    
    lambdaDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "lambda_duration_seconds",
            Help: "Lambda execution duration",
        },
        []string{"lambda_name"},
    )
)

monitoredProcess := compose.InvokableLambda(func(ctx context.Context, input string) (string, error) {
    timer := prometheus.NewTimer(lambdaDuration.WithLabelValues("process"))
    defer timer.ObserveDuration()
    
    result, err := businessLogic(input)
    
    if err != nil {
        lambdaCounter.WithLabelValues("error").Inc()
    } else {
        lambdaCounter.WithLabelValues("success").Inc()
    }
    
    return result, err
})
```

## 测试

### 1. 单元测试
```go
func TestProcessUser(t *testing.T) {
    processUser := compose.InvokableLambda(func(ctx context.Context, user User) (string, error) {
        return fmt.Sprintf("Hello, %s!", user.Name), nil
    })
    
    // 创建测试链
    chain := compose.NewChain[User, string]()
    chain.AppendLambda(processUser)
    
    runnable, err := chain.Compile(context.Background())
    assert.NoError(t, err)
    
    // 测试正常情况
    result, err := runnable.Invoke(context.Background(), User{Name: "Alice"})
    assert.NoError(t, err)
    assert.Equal(t, "Hello, Alice!", result)
}
```

### 2. 集成测试
```go
func TestDataProcessingChain(t *testing.T) {
    ctx := context.Background()
    
    // 构建完整的处理链
    chain := compose.NewChain[string, string]()
    
    validate := compose.InvokableLambda(validateInput)
    transform := compose.InvokableLambda(transformData)
    format := compose.InvokableLambda(formatOutput)
    
    chain.AppendLambda(validate)
    chain.AppendLambda(transform)
    chain.AppendLambda(format)
    
    runnable, err := chain.Compile(ctx)
    require.NoError(t, err)
    
    // 测试端到端流程
    result, err := runnable.Invoke(ctx, "test input")
    assert.NoError(t, err)
    assert.Contains(t, result, "processed")
}
```

### 3. 错误场景测试
```go
func TestValidationErrors(t *testing.T) {
    validate := compose.InvokableLambda(func(ctx context.Context, input string) (string, error) {
        if input == "" {
            return "", errors.New("输入不能为空")
        }
        return input, nil
    })
    
    chain := compose.NewChain[string, string]()
    chain.AppendLambda(validate)
    
    runnable, err := chain.Compile(context.Background())
    require.NoError(t, err)
    
    // 测试错误情况
    _, err = runnable.Invoke(context.Background(), "")
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "输入不能为空")
}
```

## 进阶用法

### 1. 条件分支
```go
conditionalProcess := compose.InvokableLambda(func(ctx context.Context, input ProcessRequest) (ProcessResult, error) {
    switch input.Type {
    case "text":
        return processText(input.Data)
    case "image":
        return processImage(input.Data)
    case "video":
        return processVideo(input.Data)
    default:
        return ProcessResult{}, fmt.Errorf("不支持的处理类型: %s", input.Type)
    }
})
```

### 2. 动态配置
```go
func createConfigurableLambda(config ProcessConfig) *compose.Lambda {
    return compose.InvokableLambda(func(ctx context.Context, input string) (string, error) {
        if config.EnableValidation && len(input) == 0 {
            return "", errors.New("验证失败: 输入为空")
        }
        
        result := input
        if config.ToUpperCase {
            result = strings.ToUpper(result)
        }
        
        if config.AddPrefix != "" {
            result = config.AddPrefix + result
        }
        
        return result, nil
    })
}
```

### 3. 中间件模式
```go
func withLogging(name string, lambda *compose.Lambda) *compose.Lambda {
    return compose.InvokableLambda(func(ctx context.Context, input interface{}) (interface{}, error) {
        log.Printf("[%s] 开始处理", name)
        start := time.Now()
        
        // 这里需要根据实际情况调整，因为无法直接调用其他 Lambda
        // 实际实现中，可以通过 Chain 组合来实现中间件效果
        
        log.Printf("[%s] 处理完成，耗时: %v", name, time.Since(start))
        return input, nil
    })
}
```

## 常见问题

### Q: Lambda 中可以调用其他 Lambda 吗？
A: 不能直接调用。Lambda 应该通过 Chain 或 Graph 进行组合。如果需要复杂的调用关系，建议使用 Graph 组件。

### Q: Lambda 是否支持并发执行？
A: Lambda 本身是线程安全的，但内部的业务逻辑需要开发者确保线程安全。可以在 Chain 中并行执行多个分支。

### Q: 如何处理 Lambda 中的长时间运行任务？
A: 使用 context.Context 进行超时控制和取消操作，定期检查 `ctx.Done()` 信号。

### Q: Lambda 的性能如何？
A: Lambda 本身开销很小，主要性能取决于内部的业务逻辑。建议进行性能测试和监控。

### Q: 是否支持热更新？
A: 当前版本不支持运行时热更新。需要重新编译和部署应用程序。

## 总结

Lambda 是 Eino 框架中强大而灵活的组件，通过合理使用可以构建出高效、可维护的数据处理管道。关键要点：

1. **正确使用**：通过 Chain 或 Graph 使用 Lambda，不要直接调用
2. **类型安全**：充分利用泛型确保类型安全
3. **错误处理**：完善的错误处理机制
4. **性能优化**：合理使用缓存、并发控制等技术
5. **测试覆盖**：编写完整的单元测试和集成测试

通过遵循这些最佳实践，可以充分发挥 Lambda 组件的优势，构建出高质量的 AI 应用。