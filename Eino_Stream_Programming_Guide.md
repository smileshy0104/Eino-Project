# Eino 流式编程核心要点 - 从水流到数据流的艺术

## 🌊 什么是流式编程？

想象一下河流中的水，它不是一次性全部流淌，而是连续不断地一股股流过：

```
普通数据处理：
📦 [完整数据包] → 🔧 [处理器] → 📦 [完整结果]
   等待全部      一次性处理     一次性输出

流式数据处理：
🌊 data1 → 🔧 → result1
🌊 data2 → 🔧 → result2  ← 边接收边处理边输出
🌊 data3 → 🔧 → result3
🌊 data4 → 🔧 → result4
```

**流式编程**就是让数据像水流一样，可以分成一帧一帧（chunk）进行处理，而不需要等待所有数据都到齐再开始处理。

### 🎯 流式编程的核心价值

- ⚡ **低延迟**：不用等待完整数据，收到一部分就开始处理
- 💾 **内存友好**：不需要把所有数据都加载到内存
- 🔄 **实时响应**：用户可以立即看到部分结果
- 📈 **可扩展性**：能处理任意大小的数据

---

## 🎭 四大流式范式详解

Eino 提供了四种流式交互模式，就像四种不同的对话方式：

### 1. 🎯 Invoke - 传统问答（Ping-Pong）

**模式特点**：一问一答，简单直接

```
👤 用户: "你好" → 🤖 AI: "您好！有什么可以帮您的吗？"
```

**实际应用示例：**
```go
// 创建传统的问答节点
func createSimpleQANode() compose.Lambda {
    return compose.InvokableLambda(func(ctx context.Context, question string) (string, error) {
        // 简单问答处理
        switch strings.ToLower(question) {
        case "你是谁":
            return "我是 AI 助手，很高兴为您服务！", nil
        case "今天天气如何":
            return "今天天气晴朗，适合外出活动。", nil
        case "推荐一本书":
            return "推荐《三体》，一部优秀的科幻小说。", nil
        default:
            return "感谢您的问题，让我想想... 这是一个很有趣的话题。", nil
        }
    })
}

// 在链中使用（需要保证“类型对齐”）
func buildSimpleQAChain() compose.Chain[string, string] {
    chain := compose.NewChain[string, string]()
    
    // 预处理问题
    preprocessor := compose.InvokableLambda(func(rawQuestion string) string {
        // 清理问题格式
        cleaned := strings.TrimSpace(rawQuestion)
        cleaned = strings.ReplaceAll(cleaned, "？", "")
        cleaned = strings.ReplaceAll(cleaned, "?", "")
        return cleaned
    })
    
    // 主要问答处理
    qaNode := createSimpleQANode()
    
    // 后处理答案
    postprocessor := compose.InvokableLambda(func(answer string) string {
        // 添加礼貌用语
        return answer + " 还有什么其他问题吗？"
    })
    
    chain.AppendLambda(preprocessor) // 输出string
    chain.AppendLambda(qaNode) // 输入string → 输出string
    chain.AppendLambda(postprocessor) // 输入string
    
    return chain
}
```

### 2. 📡 Stream - 服务器推送（Server-Streaming）

**模式特点**：一次请求，连续响应（流式输出）

```
👤 用户: "给我讲个故事"

🤖 AI: "从前有一座山，" ← 第1帧
🤖 AI: "山上有座庙，"     ← 第2帧  
🤖 AI: "庙里有个老和尚..." ← 第3帧
```

**实际应用示例：**
```go
// 故事生成器 - 流式输出
func createStoryStreamer() compose.Lambda {
    return compose.StreamableLambda(func(ctx context.Context, prompt string) (<-chan string, error) {
        output := make(chan string, 10)
        
        go func() {
            defer close(output)
            
            // 根据提示生成故事片段
            storyParts := generateStoryParts(prompt)
            
            for i, part := range storyParts {
                select {
                case <-ctx.Done():
                    return
                case output <- part:
                    // 模拟生成延迟，让流式效果更明显
                    time.Sleep(200 * time.Millisecond)
                }
            }
        }()
        
        return output, nil
    })
}

// 生成故事片段的辅助函数
func generateStoryParts(prompt string) []string {
    baseStory := []string{
        "在一个遥远的王国里，",
        "住着一位勇敢的骑士。",
        "他有一匹神奇的马，",
        "能够飞越高山和河流。",
        "一天，王国遇到了危机，",
        "邪恶的龙威胁着村庄。",
        "骑士决定挺身而出，",
        "踏上了冒险的旅程...",
    }
    
    // 根据提示词个性化故事
    if strings.Contains(strings.ToLower(prompt), "科幻") {
        return []string{
            "在2024年的地球上，",
            "科技已经高度发达。",
            "人工智能与人类和谐共处，",
            "一起探索宇宙的奥秘。",
            "突然，从深空传来神秘信号，",
            "预示着外星文明的到来...",
        }
    }
    
    return baseStory
}

// 实时聊天应用示例
func buildStreamingChatApp() {
    streamer := createStoryStreamer()
    
    // 模拟用户请求
    ctx := context.Background()
    storyStream, err := streamer.Invoke(ctx, "给我讲一个科幻故事")
    if err != nil {
        log.Fatal("创建故事流失败:", err)
    }
    
    // 实时接收和显示故事片段
    fmt.Println("🎭 开始讲故事：")
    for storyPart := range storyStream.(<-chan string) {
        fmt.Printf("📖 %s\n", storyPart)
        // 在真实应用中，这里会发送到前端界面
        // websocket.Send(storyPart)
    }
    fmt.Println("✨ 故事讲完了！")
}
```

### 3. 📥 Collect - 客户端推送（Client-Streaming）

**模式特点**：连续输入，最终汇总（流式输入）

```
👤 用户: "今天天气不错" → 
👤 用户: "心情很好"     →  🤖 AI: "综合分析：您今天状态很棒！"
👤 用户: "工作顺利"     →
```

**实际应用示例：**
```go
// 情感分析聚合器 - 收集多条输入并分析整体情感
func createSentimentCollector() compose.Lambda {
    return compose.CollectableLambda(func(ctx context.Context, messages <-chan string) (SentimentReport, error) {
        var allMessages []string
        var totalScore float64
        var messageCount int
        
        // 收集所有消息
        for message := range messages {
            allMessages = append(allMessages, message)
            
            // 分析单条消息情感
            score := analyzeSingleMessageSentiment(message)
            totalScore += score
            messageCount++
        }
        
        if messageCount == 0 {
            return SentimentReport{}, fmt.Errorf("没有收到任何消息")
        }
        
        // 计算整体情感
        averageScore := totalScore / float64(messageCount)
        overallSentiment := categorizeOverallSentiment(averageScore)
        
        return SentimentReport{
            Messages:         allMessages,
            MessageCount:     messageCount,
            AverageScore:     averageScore,
            OverallSentiment: overallSentiment,
            AnalyzedAt:       time.Now(),
        }, nil
    })
}

// 情感报告结构
type SentimentReport struct {
    Messages         []string  `json:"messages"`
    MessageCount     int       `json:"message_count"`
    AverageScore     float64   `json:"average_score"`
    OverallSentiment string    `json:"overall_sentiment"`
    AnalyzedAt       time.Time `json:"analyzed_at"`
}

// 分析单条消息情感（简化版）
func analyzeSingleMessageSentiment(message string) float64 {
    positiveWords := []string{"好", "棒", "开心", "高兴", "顺利", "成功", "赞"}
    negativeWords := []string{"坏", "糟", "难过", "失望", "困难", "失败", "烦"}
    
    score := 0.5 // 中性基准
    
    for _, word := range positiveWords {
        if strings.Contains(message, word) {
            score += 0.2
        }
    }
    
    for _, word := range negativeWords {
        if strings.Contains(message, word) {
            score -= 0.2
        }
    }
    
    // 确保分数在0-1之间
    if score > 1.0 {
        score = 1.0
    } else if score < 0.0 {
        score = 0.0
    }
    
    return score
}

// 分类整体情感
func categorizeOverallSentiment(score float64) string {
    switch {
    case score >= 0.7:
        return "非常积极 😊"
    case score >= 0.6:
        return "积极 🙂"
    case score >= 0.4:
        return "中性 😐"
    case score >= 0.3:
        return "消极 🙁"
    default:
        return "非常消极 😢"
    }
}

// 使用示例：聊天记录情感分析
func buildSentimentAnalysisApp() {
    collector := createSentimentCollector()
    
    // 模拟客户端发送多条消息
    messageStream := make(chan string, 10)
    
    go func() {
        defer close(messageStream)
        
        messages := []string{
            "今天天气真好",
            "工作进展很顺利",
            "同事们都很友善",
            "项目终于完成了",
            "感觉很有成就感",
        }
        
        for _, msg := range messages {
            messageStream <- msg
            time.Sleep(100 * time.Millisecond) // 模拟实时输入
        }
    }()
    
    // 收集并分析
    ctx := context.Background()
    report, err := collector.Invoke(ctx, messageStream)
    if err != nil {
        log.Fatal("情感分析失败:", err)
    }
    
    result := report.(SentimentReport)
    fmt.Printf("📊 情感分析报告：\n")
    fmt.Printf("   消息数量: %d\n", result.MessageCount)
    fmt.Printf("   平均情感分: %.2f\n", result.AverageScore)
    fmt.Printf("   整体情感: %s\n", result.OverallSentiment)
    fmt.Printf("   分析时间: %s\n", result.AnalyzedAt.Format("2006-01-02 15:04:05"))
}
```

### 4. 🔄 Transform - 双向流式（Bidirectional-Streaming）

**模式特点**：连续对话，实时互动

```
👤 用户: "帮我翻译" →  🤖 AI: "好的，请发送要翻译的内容"
👤 用户: "Hello"   →  🤖 AI: "你好"
👤 用户: "World"   →  🤖 AI: "世界"
👤 用户: "Thanks"  →  🤖 AI: "谢谢"
```

**实际应用示例：**
```go
// 实时翻译器 - 双向流式处理
func createRealTimeTranslator() compose.Lambda {
    return compose.TransformableLambda(func(ctx context.Context, input <-chan string) (<-chan string, error) {
        output := make(chan string, 10)
        
        go func() {
            defer close(output)
            
            for text := range input {
                select {
                case <-ctx.Done():
                    return
                default:
                    // 实时翻译处理
                    translated := translateText(text)
                    
                    select {
                    case output <- translated:
                        // 翻译成功发送
                    case <-ctx.Done():
                        return
                    }
                }
            }
        }()
        
        return output, nil
    })
}

// 翻译函数（简化版）
func translateText(text string) string {
    // 简单的中英文互译字典
    translations := map[string]string{
        "hello":     "你好",
        "world":     "世界",
        "thanks":    "谢谢",
        "goodbye":   "再见",
        "computer":  "计算机",
        "程序":      "program",
        "你好":      "hello",
        "世界":      "world",
        "谢谢":      "thanks",
        "再见":      "goodbye",
        "计算机":    "computer",
    }
    
    // 尝试直接翻译
    if translated, exists := translations[strings.ToLower(text)]; exists {
        return translated
    }
    
    // 检测语言并处理
    if isChineseText(text) {
        return fmt.Sprintf("[EN] %s", text) // 简化处理
    } else {
        return fmt.Sprintf("[中文] %s", text)
    }
}

// 简单的中文检测
func isChineseText(text string) bool {
    for _, r := range text {
        if r >= 0x4e00 && r <= 0x9fff {
            return true
        }
    }
    return false
}

// 智能代码助手 - 双向流式编程辅助
func createCodeAssistant() compose.Lambda {
    return compose.TransformableLambda(func(ctx context.Context, codeInput <-chan CodeRequest) (<-chan CodeSuggestion, error) {
        suggestions := make(chan CodeSuggestion, 10)
        
        go func() {
            defer close(suggestions)
            
            for request := range codeInput {
                select {
                case <-ctx.Done():
                    return
                default:
                    // 分析代码并提供建议
                    suggestion := analyzeAndSuggest(request)
                    
                    select {
                    case suggestions <- suggestion:
                        // 建议发送成功
                    case <-ctx.Done():
                        return
                    }
                }
            }
        }()
        
        return suggestions, nil
    })
}

// 代码请求结构
type CodeRequest struct {
    Language    string `json:"language"`
    Code        string `json:"code"`
    Context     string `json:"context"`
    RequestType string `json:"request_type"` // "review", "optimize", "explain"
}

// 代码建议结构
type CodeSuggestion struct {
    Type        string    `json:"type"`
    Message     string    `json:"message"`
    Suggestion  string    `json:"suggestion,omitempty"`
    Line        int       `json:"line,omitempty"`
    Severity    string    `json:"severity"`
    GeneratedAt time.Time `json:"generated_at"`
}

// 代码分析和建议生成
func analyzeAndSuggest(request CodeRequest) CodeSuggestion {
    switch request.RequestType {
    case "review":
        return reviewCode(request)
    case "optimize":
        return optimizeCode(request)
    case "explain":
        return explainCode(request)
    default:
        return CodeSuggestion{
            Type:        "error",
            Message:     "未知的请求类型",
            Severity:    "error",
            GeneratedAt: time.Now(),
        }
    }
}

func reviewCode(request CodeRequest) CodeSuggestion {
    // 简化的代码审查逻辑
    issues := []string{}
    
    if strings.Contains(request.Code, "var ") && request.Language == "go" {
        issues = append(issues, "建议使用短变量声明 := 而不是 var")
    }
    
    if strings.Contains(request.Code, "fmt.Print") && !strings.Contains(request.Code, "fmt.Printf") {
        issues = append(issues, "考虑使用 fmt.Printf 进行格式化输出")
    }
    
    if len(issues) == 0 {
        return CodeSuggestion{
            Type:        "success",
            Message:     "代码看起来不错！没有发现明显问题。",
            Severity:    "info",
            GeneratedAt: time.Now(),
        }
    }
    
    return CodeSuggestion{
        Type:        "warning",
        Message:     "发现一些可以改进的地方",
        Suggestion:  strings.Join(issues, "; "),
        Severity:    "warning",
        GeneratedAt: time.Now(),
    }
}

func optimizeCode(request CodeRequest) CodeSuggestion {
    return CodeSuggestion{
        Type:        "optimization",
        Message:     "性能优化建议",
        Suggestion:  "考虑使用更高效的数据结构，例如使用 map 替代 slice 进行查找操作",
        Severity:    "info",
        GeneratedAt: time.Now(),
    }
}

func explainCode(request CodeRequest) CodeSuggestion {
    return CodeSuggestion{
        Type:        "explanation",
        Message:     "代码解释",
        Suggestion:  "这段代码实现了基本的数据处理逻辑，首先验证输入，然后进行转换，最后返回结果",
        Severity:    "info",
        GeneratedAt: time.Now(),
    }
}

// 实时编程助手应用
func buildRealTimeCodingApp() {
    assistant := createCodeAssistant()
    
    // 创建代码输入流
    codeInput := make(chan CodeRequest, 5)
    
    // 启动助手
    ctx := context.Background()
    suggestions, err := assistant.Invoke(ctx, codeInput)
    if err != nil {
        log.Fatal("启动代码助手失败:", err)
    }
    
    // 启动建议接收goroutine
    go func() {
        for suggestion := range suggestions.(chan CodeSuggestion) {
            fmt.Printf("💡 %s: %s\n", suggestion.Type, suggestion.Message)
            if suggestion.Suggestion != "" {
                fmt.Printf("   建议: %s\n", suggestion.Suggestion)
            }
        }
    }()
    
    // 模拟发送代码请求
    requests := []CodeRequest{
        {
            Language:    "go",
            Code:        "var name string = \"Alice\"",
            Context:     "变量声明",
            RequestType: "review",
        },
        {
            Language:    "go",
            Code:        "fmt.Print(\"Hello, World!\")",
            Context:     "输出语句",
            RequestType: "review",
        },
        {
            Language:    "go",
            Code:        "for i := 0; i < len(arr); i++ { ... }",
            Context:     "循环优化",
            RequestType: "optimize",
        },
    }
    
    for _, req := range requests {
        codeInput <- req
        time.Sleep(500 * time.Millisecond) // 模拟实时输入
    }
    
    close(codeInput)
    time.Sleep(1 * time.Second) // 等待处理完成
}
```

---

## 🔄 自动类型转换的魔法

Eino 最强大的特性之一就是**自动流式范式转换**。就像一个万能翻译器，它能让不同"语言"的组件无缝交流：

### 🎯 转换矩阵

```
输入类型 → 输出类型 → 自动转换策略

T        → T        → 直接传递 ✅
T        → Stream[T] → 装箱为单帧流 📦→🌊
Stream[T] → T        → 连接所有帧为完整数据 🌊→📦
Stream[T] → Stream[T] → 直接流式传递 🌊→🌊
```

### 💡 实际转换示例

```go
// 混合范式的智能处理链
func buildMixedModeChain() compose.Chain[string, ProcessResult] {
    chain := compose.NewChain[string, ProcessResult]()
    
    // 节点1：普通处理（T → T）
    normalProcessor := compose.InvokableLambda(func(input string) string {
        return "预处理：" + input
    })
    
    // 节点2：流式生成（T → Stream[T]）
    streamGenerator := compose.StreamableLambda(func(ctx context.Context, input string) (<-chan string, error) {
        output := make(chan string, 3)
        
        go func() {
            defer close(output)
            
            // 生成多个片段
            parts := []string{
                "第一部分：" + input,
                "第二部分：分析中...",
                "第三部分：处理完成",
            }
            
            for _, part := range parts {
                select {
                case output <- part:
                    time.Sleep(100 * time.Millisecond)
                case <-ctx.Done():
                    return
                }
            }
        }()
        
        return output, nil
    })
    
    // 节点3：流式收集（Stream[T] → T）
    streamCollector := compose.CollectableLambda(func(ctx context.Context, stream <-chan string) (string, error) {
        var collected []string
        
        for part := range stream {
            collected = append(collected, part)
        }
        
        return strings.Join(collected, " | "), nil
    })
    
    // 节点4：最终处理（T → ProcessResult）
    finalProcessor := compose.InvokableLambda(func(input string) ProcessResult {
        return ProcessResult{
            Data:      input,
            Success:   true,
            Timestamp: time.Now(),
        }
    })
    
    // 连接所有节点 - Eino 自动处理类型转换
    chain.AppendLambda(normalProcessor)     // string → string
    chain.AppendLambda(streamGenerator)     // string → <-chan string (自动转换)
    chain.AppendLambda(streamCollector)     // <-chan string → string (自动转换)  
    chain.AppendLambda(finalProcessor)      // string → ProcessResult
    
    return chain
}

type ProcessResult struct {
    Data      string    `json:"data"`
    Success   bool      `json:"success"`
    Timestamp time.Time `json:"timestamp"`
}

// 测试混合模式链
func testMixedModeProcessing() {
    chain := buildMixedModeChain()
    
    ctx := context.Background()
    result, err := chain.Invoke(ctx, "测试数据")
    if err != nil {
        log.Fatal("处理失败:", err)
    }
    
    processResult := result.(ProcessResult)
    fmt.Printf("🎯 处理结果：\n")
    fmt.Printf("   数据: %s\n", processResult.Data)
    fmt.Printf("   状态: %v\n", processResult.Success)
    fmt.Printf("   时间: %s\n", processResult.Timestamp.Format("15:04:05"))
}
```

---

## 🚀 实际应用场景

### 1. 💬 智能客服系统

```go
// 构建流式智能客服系统
func buildStreamingCustomerService() *compose.Graph[CustomerQuery, ServiceResponse] {
    graph := compose.NewGraph[CustomerQuery, ServiceResponse]()
    
    // 意图识别（普通处理）
    intentRecognition := compose.InvokableLambda(func(query CustomerQuery) Intent {
        // 分析客户查询意图
        return Intent{
            Type:       classifyIntent(query.Message),
            Confidence: calculateConfidence(query.Message),
            Context:    query.Context,
        }
    })
    
    // 知识库检索（流式输出相关内容）
    knowledgeRetrieval := compose.StreamableLambda(func(ctx context.Context, intent Intent) (<-chan KnowledgeItem, error) {
        items := make(chan KnowledgeItem, 5)
        
        go func() {
            defer close(items)
            
            // 根据意图检索相关知识
            relevantItems := searchKnowledgeBase(intent.Type)
            
            for _, item := range relevantItems {
                select {
                case items <- item:
                    // 模拟检索延迟
                    time.Sleep(50 * time.Millisecond)
                case <-ctx.Done():
                    return
                }
            }
        }()
        
        return items, nil
    })
    
    // 回答生成（流式收集知识并生成回答）
    answerGeneration := compose.CollectableLambda(func(ctx context.Context, knowledge <-chan KnowledgeItem) (string, error) {
        var allKnowledge []KnowledgeItem
        
        for item := range knowledge {
            allKnowledge = append(allKnowledge, item)
        }
        
        // 基于收集的知识生成回答
        return generateAnswer(allKnowledge), nil
    })
    
    // 响应包装（最终处理）
    responseWrapper := compose.InvokableLambda(func(answer string) ServiceResponse {
        return ServiceResponse{
            Answer:     answer,
            Timestamp:  time.Now(),
            Confidence: 0.85,
            Source:     "AI Assistant",
        }
    })
    
    // 构建图
    graph.AddNode("intent", intentRecognition)
    graph.AddNode("knowledge", knowledgeRetrieval)  
    graph.AddNode("generate", answerGeneration)
    graph.AddNode("response", responseWrapper)
    
    graph.AddEdge("intent", "knowledge")
    graph.AddEdge("knowledge", "generate")
    graph.AddEdge("generate", "response")
    
    return graph
}

// 相关数据结构
type CustomerQuery struct {
    Message   string            `json:"message"`
    UserID    string            `json:"user_id"`
    Context   map[string]string `json:"context"`
    Timestamp time.Time         `json:"timestamp"`
}

type Intent struct {
    Type       string            `json:"type"`
    Confidence float64           `json:"confidence"`
    Context    map[string]string `json:"context"`
}

type KnowledgeItem struct {
    ID          string  `json:"id"`
    Title       string  `json:"title"`
    Content     string  `json:"content"`
    Relevance   float64 `json:"relevance"`
    Category    string  `json:"category"`
}

type ServiceResponse struct {
    Answer     string    `json:"answer"`
    Timestamp  time.Time `json:"timestamp"`
    Confidence float64   `json:"confidence"`
    Source     string    `json:"source"`
}
```

### 2. 📊 实时数据分析管道

```go
// 构建实时数据分析管道
func buildRealTimeAnalyticsPipeline() compose.Chain[DataSource, AnalyticsReport] {
    chain := compose.NewChain[DataSource, AnalyticsReport]()
    
    // 数据获取（流式读取数据）
    dataFetcher := compose.StreamableLambda(func(ctx context.Context, source DataSource) (<-chan DataPoint, error) {
        dataStream := make(chan DataPoint, 100)
        
        go func() {
            defer close(dataStream)
            
            // 模拟实时数据源
            ticker := time.NewTicker(time.Duration(source.IntervalMs) * time.Millisecond)
            defer ticker.Stop()
            
            counter := 0
            for {
                select {
                case <-ticker.C:
                    if counter >= source.MaxPoints {
                        return
                    }
                    
                    // 生成模拟数据点
                    dataPoint := DataPoint{
                        Timestamp: time.Now(),
                        Value:     generateRandomValue(source.ValueRange),
                        Metric:    source.MetricName,
                        Tags:      source.Tags,
                    }
                    
                    select {
                    case dataStream <- dataPoint:
                        counter++
                    case <-ctx.Done():
                        return
                    }
                    
                case <-ctx.Done():
                    return
                }
            }
        }()
        
        return dataStream, nil
    })
    
    // 实时数据处理（流式转换）
    dataProcessor := compose.TransformableLambda(func(ctx context.Context, input <-chan DataPoint) (<-chan ProcessedData, error) {
        output := make(chan ProcessedData, 50)
        
        go func() {
            defer close(output)
            
            window := NewSlidingWindow(10) // 10个点的滑动窗口
            
            for dataPoint := range input {
                // 添加到滑动窗口
                window.Add(dataPoint)
                
                // 计算窗口统计
                if window.IsFull() {
                    processed := ProcessedData{
                        TimeRange: TimeRange{
                            Start: window.OldestTime(),
                            End:   window.NewestTime(),
                        },
                        Statistics: WindowStatistics{
                            Count:   window.Count(),
                            Average: window.Average(),
                            Min:     window.Min(),
                            Max:     window.Max(),
                            StdDev:  window.StandardDeviation(),
                        },
                        Trend: window.CalculateTrend(),
                    }
                    
                    select {
                    case output <- processed:
                    case <-ctx.Done():
                        return
                    }
                }
            }
        }()
        
        return output, nil
    })
    
    // 报告生成（收集处理后的数据生成最终报告）
    reportGenerator := compose.CollectableLambda(func(ctx context.Context, processedData <-chan ProcessedData) (AnalyticsReport, error) {
        var allData []ProcessedData
        var totalPoints int
        var avgTrend float64
        
        for data := range processedData {
            allData = append(allData, data)
            totalPoints += data.Statistics.Count
            avgTrend += data.Trend
        }
        
        if len(allData) == 0 {
            return AnalyticsReport{}, fmt.Errorf("没有收到任何处理后的数据")
        }
        
        avgTrend /= float64(len(allData))
        
        // 生成综合报告
        report := AnalyticsReport{
            GeneratedAt:    time.Now(),
            DataWindows:    len(allData),
            TotalDataPoints: totalPoints,
            OverallTrend:   avgTrend,
            TrendAnalysis:  analyzeTrend(avgTrend),
            Summary:        generateSummary(allData),
            Recommendations: generateRecommendations(allData, avgTrend),
        }
        
        return report, nil
    })
    
    // 构建处理链
    chain.AppendLambda(dataFetcher)    // DataSource → <-chan DataPoint
    chain.AppendLambda(dataProcessor)  // <-chan DataPoint → <-chan ProcessedData  
    chain.AppendLambda(reportGenerator) // <-chan ProcessedData → AnalyticsReport
    
    return chain
}

// 相关数据结构和工具类
type DataSource struct {
    MetricName   string            `json:"metric_name"`
    IntervalMs   int               `json:"interval_ms"`
    MaxPoints    int               `json:"max_points"`
    ValueRange   [2]float64        `json:"value_range"`
    Tags         map[string]string `json:"tags"`
}

type DataPoint struct {
    Timestamp time.Time         `json:"timestamp"`
    Value     float64           `json:"value"`
    Metric    string            `json:"metric"`
    Tags      map[string]string `json:"tags"`
}

type ProcessedData struct {
    TimeRange  TimeRange        `json:"time_range"`
    Statistics WindowStatistics `json:"statistics"`
    Trend      float64          `json:"trend"`
}

type AnalyticsReport struct {
    GeneratedAt     time.Time `json:"generated_at"`
    DataWindows     int       `json:"data_windows"`
    TotalDataPoints int       `json:"total_data_points"`
    OverallTrend    float64   `json:"overall_trend"`
    TrendAnalysis   string    `json:"trend_analysis"`
    Summary         string    `json:"summary"`
    Recommendations []string  `json:"recommendations"`
}

// 滑动窗口工具类
type SlidingWindow struct {
    data     []DataPoint
    maxSize  int
    current  int
    isFull   bool
}

func NewSlidingWindow(size int) *SlidingWindow {
    return &SlidingWindow{
        data:    make([]DataPoint, size),
        maxSize: size,
        current: 0,
        isFull:  false,
    }
}

func (sw *SlidingWindow) Add(point DataPoint) {
    sw.data[sw.current] = point
    sw.current = (sw.current + 1) % sw.maxSize
    
    if !sw.isFull && sw.current == 0 {
        sw.isFull = true
    }
}

func (sw *SlidingWindow) IsFull() bool {
    return sw.isFull
}

func (sw *SlidingWindow) Count() int {
    if sw.isFull {
        return sw.maxSize
    }
    return sw.current
}

func (sw *SlidingWindow) Average() float64 {
    count := sw.Count()
    if count == 0 {
        return 0
    }
    
    sum := 0.0
    for i := 0; i < count; i++ {
        sum += sw.data[i].Value
    }
    
    return sum / float64(count)
}
```

---

## 🎯 最佳实践与性能优化

### 1. 🔧 选择合适的流式范式

```go
// 决策指南
func chooseStreamingPattern(useCase string) string {
    switch useCase {
    case "simple_calculation":
        return "Invoke - 适合简单的一次性计算"
    
    case "real_time_updates":
        return "Stream - 适合推送实时更新（如股价、聊天）"
    
    case "batch_processing":
        return "Collect - 适合批量数据分析"
    
    case "interactive_dialog":
        return "Transform - 适合持续交互的场景"
    
    default:
        return "Invoke - 从简单开始，根据需要升级"
    }
}
```

### 2. 📊 性能优化技巧

```go
// 优化的流式处理器
func createOptimizedProcessor() compose.Lambda {
    return compose.TransformableLambda(func(ctx context.Context, input <-chan DataItem) (<-chan ProcessedItem, error) {
        output := make(chan ProcessedItem, 100) // ✅ 适当的缓冲区大小
        
        // 使用工作池提高并发性能
        const workerCount = 10
        var wg sync.WaitGroup
        
        // 创建工作队列
        workQueue := make(chan DataItem, 200)
        
        // 启动工作协程
        for i := 0; i < workerCount; i++ {
            wg.Add(1)
            go func() {
                defer wg.Done()
                
                for item := range workQueue {
                    // 处理单个数据项
                    processed := expensiveProcessing(item)
                    
                    select {
                    case output <- processed:
                    case <-ctx.Done():
                        return
                    }
                }
            }()
        }
        
        // 分发任务
        go func() {
            defer close(workQueue)
            
            for item := range input {
                select {
                case workQueue <- item:
                case <-ctx.Done():
                    return
                }
            }
        }()
        
        // 等待所有工作完成并关闭输出通道
        go func() {
            wg.Wait()
            close(output)
        }()
        
        return output, nil
    })
}

// 性能监控装饰器
func withPerformanceMonitoring(name string, lambda compose.Lambda) compose.Lambda {
    return compose.InvokableLambda(func(ctx context.Context, input interface{}) (interface{}, error) {
        start := time.Now()
        
        result, err := lambda.Invoke(ctx, input)
        
        duration := time.Since(start)
        
        // 记录性能指标
        fmt.Printf("📊 %s 执行时间: %v\n", name, duration)
        
        // 在生产环境中，这里应该发送到监控系统
        // metrics.RecordDuration(name, duration)
        
        return result, err
    })
}
```

### 3. 🛡️ 错误处理和恢复

```go
// 带容错的流式处理器
func createResilientProcessor() compose.Lambda {
    return compose.TransformableLambda(func(ctx context.Context, input <-chan DataItem) (<-chan Result, error) {
        output := make(chan Result, 50)
        
        go func() {
            defer close(output)
            
            for item := range input {
                // 带重试的处理
                result := processWithRetry(item, 3)
                
                select {
                case output <- result:
                case <-ctx.Done():
                    return
                }
            }
        }()
        
        return output, nil
    })
}

func processWithRetry(item DataItem, maxRetries int) Result {
    for attempt := 1; attempt <= maxRetries; attempt++ {
        result, err := riskyProcessing(item)
        
        if err == nil {
            return Result{
                Data:    result,
                Success: true,
                Error:   nil,
            }
        }
        
        if attempt == maxRetries {
            return Result{
                Data:    nil,
                Success: false,
                Error:   fmt.Errorf("处理失败，已重试%d次: %w", maxRetries, err),
            }
        }
        
        // 指数退避
        backoff := time.Duration(attempt*attempt) * 100 * time.Millisecond
        time.Sleep(backoff)
    }
    
    return Result{Success: false, Error: fmt.Errorf("不应该到达这里")}
}

type Result struct {
    Data    interface{} `json:"data"`
    Success bool        `json:"success"`
    Error   error       `json:"error,omitempty"`
}
```

---

## 📚 总结

流式编程是构建现代 AI 应用的核心技能，Eino 通过四种流式范式为你提供了完整的工具箱：

### 🎯 核心能力总结
- 🎯 **Invoke**: 传统问答，简单直接
- 📡 **Stream**: 服务器推送，实时响应  
- 📥 **Collect**: 客户端推送，批量分析
- 🔄 **Transform**: 双向流式，持续交互

### ⚡ 自动化优势
- 🔄 **自动类型转换**: T ↔ Stream[T] 无缝切换
- 🧠 **智能适配**: 混合范式自动处理
- 📦 **透明封装**: 开发者专注业务逻辑

### 🚀 应用场景
- 💬 **智能客服**: 实时对话系统
- 📊 **数据分析**: 实时流式计算
- 🤖 **AI Agent**: 连续决策处理
- 🔍 **搜索推荐**: 流式结果生成

通过掌握这些概念和实践，你可以构建出响应迅速、用户体验优秀的现代 AI 应用！

---

*"流式编程不仅是技术，更是一种思维方式 — 让数据像水一样自然流淌。"* 🌊✨

### 📖 延伸阅读
- [Eino 流式编程官方文档](https://www.cloudwego.io/zh/docs/eino/core_modules/chain_and_graph_orchestration/stream_programming_essentials/)
- [编排系统详解](./Eino_Orchestration_Guide.md)
- [性能优化最佳实践](https://www.cloudwego.io/zh/docs/eino/best_practices/)