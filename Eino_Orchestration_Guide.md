# Eino 编排系统详解 - 从工厂流水线到智能编排

## 🎯 核心设计理念：类型对齐与确定性

Eino 的编排系统建立在一个核心理念之上：**类型对齐与确定性**。这意味着：

- 🔒 **编译时类型检查**：在代码编写阶段就能发现类型不匹配问题
- 🎯 **确定性执行**：相同输入总是产生相同输出，便于调试和测试
- 🔗 **强类型连接**：组件间的连接必须类型匹配，避免运行时错误

### 💡 为什么需要类型对齐？

想象你在组装一台精密设备：

```
❌ 错误示例：
圆形螺丝 → 方形螺孔 = 💥 装配失败

✅ 正确示例：
圆形螺丝 → 圆形螺孔 = ✨ 完美契合
```

在软件系统中也是如此：
```go
// ❌ 类型不匹配会导致编译错误
func processText(text string) int { return len(text) }
func analyzeData(data string) string { return "分析结果" }

// processText 输出 int，但 analyzeData 需要 string

// ✅ 类型对齐的正确示例
func processText(text string) string { return strings.ToUpper(text) }
func analyzeData(data string) string { return "分析: " + data }
```

## 🤔 什么是编排？

想象一下你最喜欢的汽车制造厂，从原材料到成品汽车，需要经过无数个工序：

```
🔩 原材料 → 🔧 零件加工 → 🔩 组装 → 🎨 喷漆 → 🔍 质检 → 🚗 成品车
```

**编排**就是这条流水线的设计师和指挥官，它负责：
- 📋 规划每个工序的顺序
- 🔄 协调各部门之间的配合
- 📦 确保上一道工序的输出能被下一道工序接收
- ⚡ 优化整个生产流程的效率

在 Eino 框架中，编排就是将各种 AI 组件（如文档处理、向量检索、模型调用等）按照业务逻辑串联起来，形成完整的智能应用工作流。

---

## 🎭 编排的两大明星：Chain 和 Graph

### 🔗 Chain - 单线程流水线工人

**Chain** 就像一条严格按顺序执行的流水线：

```
   输入数据
      ↓
 ┌─────────────┐
 │  步骤 1     │ ← Transformer（文档分割）
 │ 处理原料    │
 └─────────────┘
      ↓
 ┌─────────────┐
 │  步骤 2     │ ← Embedder（向量化）
 │ 初步加工    │
 └─────────────┘
      ↓
 ┌─────────────┐
 │  步骤 3     │ ← Indexer（存储）
 │ 精细处理    │
 └─────────────┘
      ↓
   最终结果
```

**特点：**
- ✅ 简单直观，容易理解
- ✅ 步骤按顺序执行，不会乱
- ✅ 上一步的输出自动成为下一步的输入
- ❌ 无法并行处理，效率可能不高

### 🕸️ Graph - 多线程协作团队

**Graph** 就像一个多部门协作的智能工厂：

```
                      输入数据
                         ↓
                   ┌─────────────┐
                   │  数据预处理  │
                   └─────────────┘
                      ↓     ↓
              ┌───────┘     └───────┐
              ↓                     ↓
      ┌─────────────┐         ┌─────────────┐
      │  路径A处理   │         │  路径B处理   │  ← 并行执行
      │ (文档分析)   │         │ (向量检索)   │
      └─────────────┘         └─────────────┘
              ↓                     ↓
              └───────┐     ┌───────┘
                      ↓     ↓
                   ┌─────────────┐
                   │  结果合并    │ ← 等待所有分支完成
                   └─────────────┘
                         ↓
                      最终结果
```

**特点：**
- ✅ 支持并行处理，效率高
- ✅ 可以处理复杂的分支逻辑
- ✅ 支持条件判断和循环
- ⚠️ 相对复杂，需要仔细设计

---

## 🧱 类型对齐 - 编排的基石

### 问题：积木不匹配怎么办？

想象你在搭乐高积木：

```
🔴 圆形积木 → 🔷 方形接口 ❌ 无法连接！
```

在编程中也是如此：

```go
// 错误示例：类型不匹配
func processText(text string) int { return len(text) }
func analyzeNumber(num string) string { return "分析：" + num }

// ❌ 这样连接会出错：
// processText 输出 int，但 analyzeNumber 需要 string
```

### 解决方案：Eino 的类型对齐机制

Eino 就像一个智能的积木连接器，通过**编**连接节点，确保每个组件都能完美对接：

```
┌─────────────┐     边(edge)     ┌─────────────┐
│  上游节点   │ ────────────→   │  下游节点   │
│ 输出: string │   ✅ 类型匹配     │ 输入: string │
└─────────────┘                 └─────────────┘
```

**Eino 支持三种类型匹配方式：**

### 1. 📐 完全相同类型
```go
// ✅ 类型完全匹配
func textProcessor() string { 
    return "处理后的文本" 
}

func textAnalyzer(input string) AnalysisResult { 
    return AnalysisResult{Content: input, WordCount: len(strings.Split(input, " "))} 
}

// 在 Chain 中连接
chain := compose.NewChain[string, AnalysisResult]()
chain.AppendLambda(compose.InvokableLambda(textProcessor))
chain.AppendLambda(compose.InvokableLambda(textAnalyzer))
```

### 2. 🔌 接口实现匹配
```go
// 定义接口
type Readable interface {
    Read() string
}

// 上游输出实现了接口
type Document struct {
    content string
}
func (d Document) Read() string { return d.content }

func documentCreator() Document { 
    return Document{content: "文档内容"} 
}

// 下游接收接口类型
func interfaceProcessor(input Readable) string { 
    return "处理：" + input.Read()
}

// ✅ Document 实现了 Readable 接口，可以匹配
```

### 3. 🎭 接口到具体类型（any 匹配 任何类型）
```go
// 上游输出接口类型
func flexibleOutput() interface{} { 
    return "可以是任何类型" 
}

// 下游接收具体类型
func stringProcessor(input string) ProcessResult {
    // Eino 在运行时进行类型断言和检查
    return ProcessResult{Data: input}
}

// ⚠️ 运行时检查，需要确保实际类型匹配
```

### 🔍 编译时检查 vs 运行时检查

```go
// 编译时检查示例
type TypeSafeChain struct{}

func (t *TypeSafeChain) Build() {
    // ✅ 编译时就能发现类型不匹配
    chain := compose.NewChain[string, int]()
    chain.AppendLambda(compose.InvokableLambda(func(s string) int {
        return len(s)
    }))
    chain.AppendLambda(compose.InvokableLambda(func(i int) string {
        return fmt.Sprintf("长度：%d", i)
    }))
    // ❌ 编译错误：Chain[string, int] 不能接收输出 string 的 Lambda
}

// 运行时检查示例
func runtimeExample() {
    flexibleChain := compose.NewChain[interface{}, interface{}]()
    flexibleChain.AppendLambda(compose.InvokableLambda(func(input interface{}) interface{} {
        // 运行时类型断言
        if str, ok := input.(string); ok {
            return len(str)
        }
        return nil
    }))
}
```

---

## 🏭 实际案例：智能问答系统的编排

让我们看一个真实的例子，构建一个像 ChatGPT 一样的问答系统：

### 📋 需求分析
用户问："北京明天天气怎么样？"，系统需要：
1. 理解用户意图
2. 决定是否需要调用天气工具
3. 获取天气信息
4. 生成自然语言回答

### 🔗 Chain 版本 - 线性处理

```
用户问题："北京明天天气怎么样？"
    ↓
┌─────────────────────┐
│  步骤1: 意图识别     │ → 识别为"天气查询"
│  Lambda处理         │
└─────────────────────┘
    ↓
┌─────────────────────┐
│  步骤2: 工具调用     │ → 调用天气API
│  WeatherTool        │
└─────────────────────┘
    ↓
┌─────────────────────┐
│  步骤3: 回答生成     │ → "明天北京晴天，25°C"
│  ChatModel          │
└─────────────────────┘
```

**代码示例：**
```go
// 创建 Chain
chain := compose.NewChain[string, string]()

// 步骤1：意图识别
intentRecognition := compose.InvokableLambda(func(ctx context.Context, question string) (Intent, error) {
    // 分析用户问题，识别意图
    if strings.Contains(question, "天气") {
        return Intent{Type: "weather", Location: "北京"}, nil
    }
    return Intent{Type: "general"}, nil
})

// 步骤2：工具调用
toolExecution := compose.InvokableLambda(func(ctx context.Context, intent Intent) (ToolResult, error) {
    if intent.Type == "weather" {
        // 调用天气工具
        return weatherTool.Call(intent.Location)
    }
    return ToolResult{}, nil
})

// 步骤3：回答生成
responseGeneration := compose.InvokableLambda(func(ctx context.Context, result ToolResult) (string, error) {
    // 基于工具结果生成自然语言回答
    return fmt.Sprintf("根据查询，%s", result.Data), nil
})

// 组装链条
chain.AppendLambda(intentRecognition)
chain.AppendLambda(toolExecution)
chain.AppendLambda(responseGeneration)
```

### 🕸️ Graph 版本 - 并行智能处理

```
                    用户问题
                        ↓
                 ┌─────────────┐
                 │ 问题分析节点 │
                 └─────────────┘
                    ↓     ↓
            ┌───────┘     └───────┐
            ↓                     ↓
     ┌─────────────┐       ┌─────────────┐
     │ 知识库检索   │       │ 意图识别     │ ← 并行执行
     │ (可能有相关   │       │ (是否需要工具) │
     │  天气知识)   │       └─────────────┘
     └─────────────┘              ↓
            ↓                ┌─────────────┐
            │                │ 工具决策     │
            │                └─────────────┘
            │                     ↓
            │                ┌─────────────┐
            │                │ 天气API调用  │
            │                └─────────────┘
            ↓                     ↓
            └───────┐     ┌───────┘
                    ↓     ↓
                 ┌─────────────┐
                 │ 智能合并节点 │ ← 合并所有信息
                 │(知识+工具结果)│
                 └─────────────┘
                        ↓
                 ┌─────────────┐
                 │ 最终回答生成 │
                 └─────────────┘
```

**优势对比：**

| 特性 | Chain 版本 | Graph 版本 |
|------|------------|-------------|
| **执行效率** | 🐌 串行，较慢 | ⚡ 并行，更快 |
| **资源利用** | 📱 单线程 | 💻 多线程 |
| **复杂度** | 😊 简单易懂 | 🤔 稍微复杂 |
| **灵活性** | 📏 固定流程 | 🎯 动态分支 |
| **适用场景** | 简单流水线 | 复杂业务逻辑 |

---

## 🎨 核心设计原则详解

### 1. 🔒 外部变量只读原则（引用类型——〉影响全局）

**原则说明**：编排系统内部不能修改外部定义的变量，只能读取。

```
        编排系统边界
    ┌─────────────────────┐
    │                     │
    │  ┌─────┐    ┌─────┐  │
外部  │  │节点A│ →  │节点B│  │  只能读取
变量  │  └─────┘    └─────┘  │  不能修改
 ↓   │                     │    ↑
🔒   │     内部数据流       │   🚫
只读  │  ←─────────────→    │  禁止写入
    │                     │
    └─────────────────────┘
```

**实际示例：**
```go
// 外部配置变量
var globalConfig = Config{
    MaxRetries: 3,
    Timeout: 30 * time.Second,
}

// ✅ 正确：只读取外部变量
func createRetryNode() compose.Lambda {
    return compose.InvokableLambda(func(ctx context.Context, input Request) (Response, error) {
        // 只读取配置，不修改
        maxRetries := globalConfig.MaxRetries
        timeout := globalConfig.Timeout
        
        return processWithConfig(input, maxRetries, timeout)
    })
}

// ❌ 错误：修改外部变量
func badNode() compose.Lambda {
    return compose.InvokableLambda(func(ctx context.Context, input Request) (Response, error) {
        // 这样做违反了只读原则
        globalConfig.MaxRetries = 5 // ❌ 不允许修改
        return processRequest(input)
    })
}
```

**为什么要只读？**
- 🛡️ **避免副作用**：防止编排过程意外修改全局状态
- 🔄 **确保可重复性**：相同输入总是产生相同输出
- 🐛 **减少调试难度**：避免隐藏的状态变化导致的 bug
- 🧪 **便于测试**：测试结果更加可预测

### 2. 🌊 流式处理自动补全（在编排场景中，Eino 自动帮助所有的节点补全缺失的流式范式。）

**原则说明**：Eino 自动处理流式和非流式组件的转换，让开发者专注业务逻辑。

```
非流式组件 → 流式环境 → 自动流式化
┌─────────┐    🌊      ┌─────────┐
│单个处理  │  ────→     │流式处理  │
│ f(x)   │  自动转换    │ f(x1),  │
└─────────┘            │ f(x2),  │
                       │ f(x3)...│
                       └─────────┘
```

**实际示例：**
```go
// 普通的单个文档处理函数
func processSingleDoc(doc Document) ProcessedDoc {
    return ProcessedDoc{
        Title:   extractTitle(doc.Content),
        Summary: generateSummary(doc.Content),
        Tags:    extractTags(doc.Content),
    }
}

// Eino 自动将其流式化
func createDocumentProcessor() compose.Lambda {
    return compose.InvokableLambda(processSingleDoc)
}

// 在流式环境中使用
func buildStreamingPipeline() compose.Chain[<-chan Document, <-chan ProcessedDoc] {
    chain := compose.NewChain[<-chan Document, <-chan ProcessedDoc]()
    
    // ✅ Eino 自动处理流式转换
    // 单个文档处理函数自动变成流式处理
    chain.AppendLambda(createDocumentProcessor())
    
    return chain
}

// 实际使用
func main() {
    pipeline := buildStreamingPipeline()
    
    // 输入文档流
    docStream := make(chan Document, 10)
    go func() {
        for i := 0; i < 100; i++ {
            docStream <- Document{Content: fmt.Sprintf("文档%d内容", i)}
        }
        close(docStream)
    }()
    
    // 自动流式处理
    processedStream, err := pipeline.Invoke(context.Background(), docStream)
    if err != nil {
        log.Fatal(err)
    }
    
    // 处理结果
    for processed := range processedStream {
        fmt.Printf("处理完成: %s\n", processed.Title)
    }
}
```

**流式处理的优势：**
- ⚡ **内存效率**：不需要将所有数据加载到内存
- 🔄 **实时处理**：数据到达即处理，降低延迟
- 📈 **可扩展性**：能处理任意大小的数据集
- 🔧 **自动优化**：Eino 自动优化流的合并和分发

### 3. 🎯 智能类型转换

**原则说明**：支持自定义类型合并、拼接和转换函数，处理复杂的数据流转换。

```
   多输入合并               类型转换               输出适配
┌─────────┐ ┌─────────┐   ┌─────────┐         ┌─────────┐
│ Result1 │ │ Result2 │ → │自定义合并│ ──→     │ Target  │
│ {A: 1}  │ │ {B: 2}  │   │函数     │  类型转换  │ Format  │
└─────────┘ └─────────┘   └─────────┘         └─────────┘
                              ↓
                         ┌─────────┐
                         │Combined │
                         │{A:1,B:2}│
                         └─────────┘
```

**实际示例：**
```go
// 定义复杂的数据类型
type SearchResult struct {
    Documents []Document
    Score     float64
}

type AnalysisResult struct {
    Keywords []string
    Sentiment float64
}

type CombinedResult struct {
    Documents []Document
    Keywords  []string
    Score     float64
    Sentiment float64
}

// 自定义合并函数
func combineResults(search SearchResult, analysis AnalysisResult) CombinedResult {
    return CombinedResult{
        Documents: search.Documents,
        Keywords:  analysis.Keywords,
        Score:     search.Score,
        Sentiment: analysis.Sentiment,
    }
}

// 在 Graph 中使用扇入模式
func buildAnalysisGraph() *compose.Graph[string, CombinedResult] {
    graph := compose.NewGraph[string, CombinedResult]()
    
    // 搜索节点
    searchNode := compose.InvokableLambda(func(query string) SearchResult {
        // 模拟搜索逻辑
        return SearchResult{
            Documents: searchDocuments(query),
            Score:     calculateScore(query),
        }
    })
    
    // 分析节点
    analysisNode := compose.InvokableLambda(func(query string) AnalysisResult {
        // 模拟分析逻辑
        return AnalysisResult{
            Keywords:  extractKeywords(query),
            Sentiment: analyzeSentiment(query),
        }
    })
    
    // 合并节点
    mergeNode := compose.InvokableLambda(combineResults)
    
    // 构建图结构
    graph.AddNode("search", searchNode)
    graph.AddNode("analysis", analysisNode)
    graph.AddNode("merge", mergeNode)
    
    // 设置边
    graph.AddEdge("search", "merge")
    graph.AddEdge("analysis", "merge")
    
    return graph
}
```

**类型转换的应用场景：**
- 🔄 **多源数据合并**：将不同来源的数据整合
- 📊 **格式标准化**：统一不同格式的数据
- 🎯 **结果聚合**：将并行处理的结果合并
- 🔧 **接口适配**：适配不同组件间的数据格式

### 4. 🚀 双引擎支持

**原则说明**：Eino 提供两种运行引擎，适应不同的编排需求。

```
        编排类型选择
           ↓
    ┌─────────────────┐
    │   DAG 引擎      │ ← 有向无环图
    │ (简单、高效)     │
    └─────────────────┘
           ↓
    ┌─────────────────┐
    │  Pregel 引擎    │ ← 有向有环图
    │ (复杂、灵活)     │
    └─────────────────┘
```

**DAG 引擎示例**：
```go
// 简单的文档处理流水线
func buildDAGPipeline() *compose.Graph[Document, ProcessedDoc] {
    graph := compose.NewGraph[Document, ProcessedDoc]()
    
    // 线性处理流程：解析 → 分析 → 索引
    parseNode := compose.InvokableLambda(parseDocument)
    analyzeNode := compose.InvokableLambda(analyzeContent)
    indexNode := compose.InvokableLambda(indexDocument)
    
    graph.AddNode("parse", parseNode)
    graph.AddNode("analyze", analyzeNode)
    graph.AddNode("index", indexNode)
    
    // DAG：无环图结构
    graph.AddEdge("parse", "analyze")
    graph.AddEdge("analyze", "index")
    
    return graph
}
```

**Pregel 引擎示例**：
```go
// 迭代优化的推荐系统
func buildPregelRecommendation() *compose.Graph[UserProfile, Recommendations] {
    graph := compose.NewGraph[UserProfile, Recommendations]()
    
    // 可能包含循环的复杂逻辑
    initNode := compose.InvokableLambda(initializeRecommendations)
    refineNode := compose.InvokableLambda(refineRecommendations)
    evaluateNode := compose.InvokableLambda(evaluateQuality)
    
    graph.AddNode("init", initNode)
    graph.AddNode("refine", refineNode)
    graph.AddNode("evaluate", evaluateNode)
    
    // 可能包含循环：评估 → 细化 → 评估
    graph.AddEdge("init", "refine")
    graph.AddEdge("refine", "evaluate")
    graph.AddEdge("evaluate", "refine") // 循环边
    
    return graph
}
```

---

## 🚀 进阶编排模式与实践

### 1. 🍴 扇出模式 (Fan-out) - 并行分发处理

**模式说明**：将单个输入分发给多个处理器并行处理，提高效率。

```
              输入文档
                 ↓
            ┌─────────┐
            │ 分发器   │ ← 自动复制输入
            └─────────┘
           ↙     ↓     ↘
    ┌─────────┐ ┌─────────┐ ┌─────────┐
    │文本提取  │ │图片提取  │ │表格提取  │ ← 并行处理
    └─────────┘ └─────────┘ └─────────┘
         ↓         ↓         ↓
    📄 文本内容  🖼️ 图片列表  📊 表格数据
```

**完整实现示例：**
```go
type Document struct {
    Content   string
    Images    [][]byte
    Tables    []string
}

type ExtractedContent struct {
    TextContent  string
    ImageList    []string
    TableData    []TableInfo
}

// 文本提取处理器
func extractText(doc Document) string {
    // 提取纯文本内容
    return strings.TrimSpace(doc.Content)
}

// 图片提取处理器
func extractImages(doc Document) []string {
    var imageNames []string
    for i, imgData := range doc.Images {
        // 保存图片并返回路径
        filename := fmt.Sprintf("image_%d.jpg", i)
        saveImage(filename, imgData)
        imageNames = append(imageNames, filename)
    }
    return imageNames
}

// 表格提取处理器
func extractTables(doc Document) []TableInfo {
    var tables []TableInfo
    for _, tableStr := range doc.Tables {
        // 解析表格数据
        table := parseTable(tableStr)
        tables = append(tables, table)
    }
    return tables
}

// 构建扇出图
func buildFanOutGraph() *compose.Graph[Document, ExtractedContent] {
    graph := compose.NewGraph[Document, ExtractedContent]()
    
    // 添加并行处理节点
    textNode := compose.InvokableLambda(extractText)
    imageNode := compose.InvokableLambda(extractImages)
    tableNode := compose.InvokableLambda(extractTables)
    
    // 合并节点
    mergeNode := compose.InvokableLambda(func(
        textContent string,
        imageList []string,
        tableData []TableInfo,
    ) ExtractedContent {
        return ExtractedContent{
            TextContent: textContent,
            ImageList:   imageList,
            TableData:   tableData,
        }
    })
    
    // 构建图结构
    graph.AddNode("text_extractor", textNode)
    graph.AddNode("image_extractor", imageNode)
    graph.AddNode("table_extractor", tableNode)
    graph.AddNode("merger", mergeNode)
    
    // 扇出：一个输入到多个处理器
    graph.AddEdge("text_extractor", "merger")
    graph.AddEdge("image_extractor", "merger")
    graph.AddEdge("table_extractor", "merger")
    
    return graph
}
```

**使用场景：**
- 📄 **多格式文档处理**：PDF、Word、Excel 同时处理
- 🔍 **多维度数据分析**：情感、关键词、主题并行分析
- 📊 **数据验证管道**：格式、完整性、业务规则并行验证
- 🎯 **多目标优化**：性能、质量、成本同时优化

### 2. 🎯 扇入模式 (Fan-in) - 智能结果合并

**模式说明**：将多个并行处理的结果智能合并为单一输出。

```
    ┌─────────┐ ┌─────────┐ ┌─────────┐
    │ 搜索结果 │ │ 推荐结果 │ │ 分析结果 │ ← 多个输入源
    │Score:0.9│ │Score:0.8│ │Score:0.7│
    └─────────┘ └─────────┘ └─────────┘
           ↘       ↓       ↙
            ┌─────────────────┐
            │   智能合并器     │ ← 权重计算+排序
            │ • 去重         │
            │ • 评分融合     │
            │ • 排序         │
            └─────────────────┘
                     ↓
            ┌─────────────────┐
            │   最终推荐列表   │
            │ [item1, item2,  │
            │  item3...]      │
            └─────────────────┘
```

**完整实现示例：**
```go
type RecommendationItem struct {
    ID     string
    Title  string
    Score  float64
    Source string
}

type SearchResults struct {
    Items []RecommendationItem
    Query string
}

type PersonalizedResults struct {
    Items     []RecommendationItem
    UserID    string
    Algorithm string
}

type TrendingResults struct {
    Items      []RecommendationItem
    TimeWindow string
}

type FinalRecommendations struct {
    Items          []RecommendationItem
    TotalScore     float64
    MixedSources   []string
    GeneratedTime  time.Time
}

// 智能合并函数
func mergeRecommendations(
    search SearchResults,
    personalized PersonalizedResults,
    trending TrendingResults,
) FinalRecommendations {
    // 创建项目映射表进行去重和评分融合
    itemMap := make(map[string]*RecommendationItem)
    sources := make(map[string]bool)
    
    // 处理搜索结果（权重：0.5）
    for _, item := range search.Items {
        key := item.ID
        if existingItem, exists := itemMap[key]; exists {
            // 融合评分
            existingItem.Score += item.Score * 0.5
        } else {
            newItem := item
            newItem.Score *= 0.5
            itemMap[key] = &newItem
        }
        sources[item.Source] = true
    }
    
    // 处理个性化结果（权重：0.3）
    for _, item := range personalized.Items {
        key := item.ID
        if existingItem, exists := itemMap[key]; exists {
            existingItem.Score += item.Score * 0.3
        } else {
            newItem := item
            newItem.Score *= 0.3
            itemMap[key] = &newItem
        }
        sources[item.Source] = true
    }
    
    // 处理趋势结果（权重：0.2）
    for _, item := range trending.Items {
        key := item.ID
        if existingItem, exists := itemMap[key]; exists {
            existingItem.Score += item.Score * 0.2
        } else {
            newItem := item
            newItem.Score *= 0.2
            itemMap[key] = &newItem
        }
        sources[item.Source] = true
    }
    
    // 转换为切片并排序
    var finalItems []RecommendationItem
    var totalScore float64
    
    for _, item := range itemMap {
        finalItems = append(finalItems, *item)
        totalScore += item.Score
    }
    
    // 按评分降序排序
    sort.Slice(finalItems, func(i, j int) bool {
        return finalItems[i].Score > finalItems[j].Score
    })
    
    // 收集源列表
    var sourceList []string
    for source := range sources {
        sourceList = append(sourceList, source)
    }
    
    return FinalRecommendations{
        Items:         finalItems[:min(len(finalItems), 10)], // 取前10个
        TotalScore:    totalScore,
        MixedSources:  sourceList,
        GeneratedTime: time.Now(),
    }
}

// 构建扇入图
func buildFanInRecommendationGraph() *compose.Graph[string, FinalRecommendations] {
    graph := compose.NewGraph[string, FinalRecommendations]()
    
    // 三个并行的推荐源
    searchNode := compose.InvokableLambda(func(query string) SearchResults {
        // 模拟搜索推荐
        return SearchResults{
            Items: generateSearchRecommendations(query),
            Query: query,
        }
    })
    
    personalizedNode := compose.InvokableLambda(func(query string) PersonalizedResults {
        // 模拟个性化推荐
        return PersonalizedResults{
            Items:     generatePersonalizedRecommendations(query),
            UserID:    getCurrentUserID(),
            Algorithm: "collaborative_filtering",
        }
    })
    
    trendingNode := compose.InvokableLambda(func(query string) TrendingResults {
        // 模拟趋势推荐
        return TrendingResults{
            Items:      generateTrendingRecommendations(query),
            TimeWindow: "last_24h",
        }
    })
    
    // 合并节点
    mergeNode := compose.InvokableLambda(mergeRecommendations)
    
    // 构建图结构
    graph.AddNode("search", searchNode)
    graph.AddNode("personalized", personalizedNode)
    graph.AddNode("trending", trendingNode)
    graph.AddNode("merge", mergeNode)
    
    // 扇入：多个输入到一个合并器
    graph.AddEdge("search", "merge")
    graph.AddEdge("personalized", "merge")
    graph.AddEdge("trending", "merge")
    
    return graph
}
```

**使用场景：**
- 🔍 **多源信息整合**：搜索、推荐、趋势结果合并
- 📊 **统计结果汇总**：多个指标的综合评分
- 🎯 **决策支持系统**：多个专家系统的决策合并
- 🔄 **A/B 测试结果**：多个版本结果的智能选择

### 3. 🔄 条件分支模式 - 动态路由处理

**模式说明**：根据输入数据的特征，动态选择不同的处理路径。

```
              输入数据
                 ↓
            ┌─────────────┐
            │ 智能路由器   │ ← 分析数据特征
            │ 条件判断逻辑 │
            └─────────────┘
           ↙       ↓       ↘
    文本类型      图片类型     视频类型
         ↓         ↓         ↓
    ┌─────────┐ ┌─────────┐ ┌─────────┐
    │文本处理  │ │图像处理  │ │视频处理  │
    │管道     │ │管道     │ │管道     │
    └─────────┘ └─────────┘ └─────────┘
         ↓         ↓         ↓
         └─────────┼─────────┘
                   ↓
            ┌─────────────┐
            │   结果标准化 │ ← 统一输出格式
            └─────────────┘
```

**完整实现示例：**
```go
type MediaFile struct {
    Path     string
    MimeType string
    Size     int64
    Metadata map[string]interface{}
}

type ProcessedMedia struct {
    OriginalFile MediaFile
    ProcessedAt  time.Time
    OutputPath   string
    Thumbnails   []string
    Metadata     map[string]interface{}
    ProcessType  string
}

// 条件判断函数
func determineMediaType(file MediaFile) string {
    mimeType := strings.ToLower(file.MimeType)
    
    switch {
    case strings.HasPrefix(mimeType, "text/"):
        return "text"
    case strings.HasPrefix(mimeType, "image/"):
        return "image"
    case strings.HasPrefix(mimeType, "video/"):
        return "video"
    case strings.HasPrefix(mimeType, "audio/"):
        return "audio"
    default:
        return "binary"
    }
}

// 不同类型的处理器
func processText(file MediaFile) ProcessedMedia {
    // 文本处理逻辑
    outputPath := processTextFile(file.Path)
    
    return ProcessedMedia{
        OriginalFile: file,
        ProcessedAt:  time.Now(),
        OutputPath:   outputPath,
        ProcessType:  "text_processing",
        Metadata: map[string]interface{}{
            "word_count":      countWords(file.Path),
            "language":        detectLanguage(file.Path),
            "encoding":        detectEncoding(file.Path),
        },
    }
}

func processImage(file MediaFile) ProcessedMedia {
    // 图像处理逻辑
    outputPath := processImageFile(file.Path)
    thumbnails := generateThumbnails(file.Path)
    
    return ProcessedMedia{
        OriginalFile: file,
        ProcessedAt:  time.Now(),
        OutputPath:   outputPath,
        Thumbnails:   thumbnails,
        ProcessType:  "image_processing",
        Metadata: map[string]interface{}{
            "dimensions":    getImageDimensions(file.Path),
            "color_profile": getColorProfile(file.Path),
            "format":        getImageFormat(file.Path),
        },
    }
}

func processVideo(file MediaFile) ProcessedMedia {
    // 视频处理逻辑
    outputPath := processVideoFile(file.Path)
    thumbnails := extractVideoThumbnails(file.Path)
    
    return ProcessedMedia{
        OriginalFile: file,
        ProcessedAt:  time.Now(),
        OutputPath:   outputPath,
        Thumbnails:   thumbnails,
        ProcessType:  "video_processing",
        Metadata: map[string]interface{}{
            "duration":    getVideoDuration(file.Path),
            "resolution":  getVideoResolution(file.Path),
            "codec":       getVideoCodec(file.Path),
            "bitrate":     getVideoBitrate(file.Path),
        },
    }
}

// 构建条件分支图
func buildConditionalMediaProcessor() *compose.Graph[MediaFile, ProcessedMedia] {
    graph := compose.NewGraph[MediaFile, ProcessedMedia]()
    
    // 路由节点
    routerNode := compose.InvokableLambda(determineMediaType)
    
    // 不同类型的处理节点
    textProcessor := compose.InvokableLambda(processText)
    imageProcessor := compose.InvokableLambda(processImage)
    videoProcessor := compose.InvokableLambda(processVideo)
    
    // 默认处理器（用于未知类型）
    defaultProcessor := compose.InvokableLambda(func(file MediaFile) ProcessedMedia {
        return ProcessedMedia{
            OriginalFile: file,
            ProcessedAt:  time.Now(),
            OutputPath:   file.Path, // 不做处理
            ProcessType:  "passthrough",
        }
    })
    
    // 构建图结构
    graph.AddNode("router", routerNode)
    graph.AddNode("text_processor", textProcessor)
    graph.AddNode("image_processor", imageProcessor)
    graph.AddNode("video_processor", videoProcessor)
    graph.AddNode("default_processor", defaultProcessor)
    
    // 条件边：根据路由结果选择不同的处理器
    graph.AddConditionalEdges("router", map[string]string{
        "text":  "text_processor",
        "image": "image_processor",
        "video": "video_processor",
    }, "default_processor") // 默认路径
    
    return graph
}

// 使用示例
func processMediaFiles(files []MediaFile) []ProcessedMedia {
    processor := buildConditionalMediaProcessor()
    var results []ProcessedMedia
    
    for _, file := range files {
        result, err := processor.Invoke(context.Background(), file)
        if err != nil {
            log.Printf("处理文件 %s 失败: %v", file.Path, err)
            continue
        }
        results = append(results, result)
    }
    
    return results
}
```

**使用场景：**
- 📁 **多媒体内容处理**：根据文件类型选择不同处理管道
- 🔐 **安全策略路由**：根据威胁级别选择不同的安全措施
- 🌍 **地域化处理**：根据用户地域选择不同的业务逻辑
- 📈 **负载均衡**：根据系统负载选择不同的处理路径
- 🎯 **个性化推荐**：根据用户画像选择不同的推荐算法

---

## 🎯 最佳实践与生产经验

### 1. 📏 编排方式选择指南

#### Chain vs Graph 决策树

```
              开始编排设计
                   ↓
            是否需要并行处理？
                ↙      ↘
             是         否
             ↓          ↓
    是否有复杂分支逻辑？    步骤是否线性依赖？
         ↙    ↘           ↙        ↘
        是     否         是         否
        ↓      ↓          ↓          ↓
    使用Graph  评估性能   使用Chain   重新分析需求
              ↓
         性能要求高？
         ↙        ↘
        是         否
        ↓          ↓
    使用Graph   使用Chain
```

**Chain 适用场景详解：**
```go
// ✅ 典型的 Chain 场景：数据预处理管道
func buildDataPreprocessingChain() compose.Chain[RawData, CleanData] {
    chain := compose.NewChain[RawData, CleanData]()
    
    // 步骤1：数据验证
    validationStep := compose.InvokableLambda(func(raw RawData) (ValidData, error) {
        if err := validateDataFormat(raw); err != nil {
            return ValidData{}, fmt.Errorf("数据格式验证失败: %w", err)
        }
        return ValidData{Data: raw.Content, Timestamp: time.Now()}, nil
    })
    
    // 步骤2：数据清洗
    cleaningStep := compose.InvokableLambda(func(valid ValidData) (CleanedData, error) {
        cleaned := removeNoise(valid.Data)
        normalized := normalizeFormat(cleaned)
        return CleanedData{Content: normalized, ProcessedAt: time.Now()}, nil
    })
    
    // 步骤3：数据标准化
    standardizationStep := compose.InvokableLambda(func(cleaned CleanedData) (CleanData, error) {
        standardized := applyStandardization(cleaned.Content)
        return CleanData{
            FinalContent: standardized,
            Quality:      calculateQualityScore(standardized),
            ProcessTime:  time.Since(cleaned.ProcessedAt),
        }, nil
    })
    
    chain.AppendLambda(validationStep)
    chain.AppendLambda(cleaningStep)
    chain.AppendLambda(standardizationStep)
    
    return chain
}
```

**Graph 适用场景详解：**
```go
// ✅ 典型的 Graph 场景：多模态内容分析
func buildMultiModalAnalysisGraph() *compose.Graph[Content, AnalysisReport] {
    graph := compose.NewGraph[Content, AnalysisReport]()
    
    // 并行分析不同模态
    textAnalyzer := compose.InvokableLambda(func(content Content) TextAnalysis {
        return TextAnalysis{
            Keywords:  extractKeywords(content.Text),
            Sentiment: analyzeSentiment(content.Text),
            Language:  detectLanguage(content.Text),
        }
    })
    
    imageAnalyzer := compose.InvokableLambda(func(content Content) ImageAnalysis {
        var results []ImageResult
        for _, img := range content.Images {
            result := ImageResult{
                Objects:    detectObjects(img),
                Scene:      classifyScene(img),
                Emotions:   detectEmotions(img),
            }
            results = append(results, result)
        }
        return ImageAnalysis{Results: results}
    })
    
    audioAnalyzer := compose.InvokableLambda(func(content Content) AudioAnalysis {
        return AudioAnalysis{
            Transcript: speechToText(content.Audio),
            Mood:       analyzeMood(content.Audio),
            Speaker:    identifySpeaker(content.Audio),
        }
    })
    
    // 综合分析节点
    synthesizer := compose.InvokableLambda(func(
        text TextAnalysis,
        image ImageAnalysis,
        audio AudioAnalysis,
    ) AnalysisReport {
        return AnalysisReport{
            OverallSentiment: combineMultiModalSentiment(text.Sentiment, image, audio.Mood),
            MainTopics:       synthesizeTopics(text.Keywords, image.Results),
            Confidence:       calculateConfidence(text, image, audio),
            GeneratedAt:      time.Now(),
        }
    })
    
    // 构建并行处理图
    graph.AddNode("text_analyzer", textAnalyzer)
    graph.AddNode("image_analyzer", imageAnalyzer)
    graph.AddNode("audio_analyzer", audioAnalyzer)
    graph.AddNode("synthesizer", synthesizer)
    
    // 扇入模式合并结果
    graph.AddEdge("text_analyzer", "synthesizer")
    graph.AddEdge("image_analyzer", "synthesizer")
    graph.AddEdge("audio_analyzer", "synthesizer")
    
    return graph
}
```

### 2. 🔧 类型设计最佳实践

#### 类型设计原则

```go
// ✅ 遵循单一职责原则
type DocumentMetadata struct {
    Title       string    `json:"title"`
    Author      string    `json:"author"`
    CreatedAt   time.Time `json:"created_at"`
    Tags        []string  `json:"tags"`
}

type DocumentContent struct {
    Text     string   `json:"text"`
    Images   []string `json:"images"`
    Tables   []Table  `json:"tables"`
}

type ProcessedDocument struct {
    Metadata    DocumentMetadata `json:"metadata"`
    Content     DocumentContent  `json:"content"`
    ProcessInfo ProcessingInfo   `json:"process_info"`
}

// ❌ 避免：把所有字段混在一起
type BadDocument struct {
    Title     string
    Author    string
    Content   string
    Images    []string
    ProcessAt time.Time
    Quality   float64
    // ... 更多混乱的字段
}
```

#### 接口设计模式

```go
// ✅ 定义清晰的接口
type Processor interface {
    Process(ctx context.Context, input interface{}) (interface{}, error)
}

type TextProcessor interface {
    ProcessText(ctx context.Context, text string) (ProcessedText, error)
}

type ImageProcessor interface {
    ProcessImage(ctx context.Context, image []byte) (ProcessedImage, error)
}

// ✅ 使用组合模式
type MultiModalProcessor struct {
    textProcessor  TextProcessor
    imageProcessor ImageProcessor
}

func (m *MultiModalProcessor) Process(ctx context.Context, content MultiModalContent) (ProcessedContent, error) {
    var wg sync.WaitGroup
    var textResult ProcessedText
    var imageResult ProcessedImage
    var textErr, imageErr error
    
    // 并行处理不同模态
    wg.Add(2)
    
    go func() {
        defer wg.Done()
        textResult, textErr = m.textProcessor.ProcessText(ctx, content.Text)
    }()
    
    go func() {
        defer wg.Done()
        imageResult, imageErr = m.imageProcessor.ProcessImage(ctx, content.Image)
    }()
    
    wg.Wait()
    
    if textErr != nil {
        return ProcessedContent{}, fmt.Errorf("文本处理失败: %w", textErr)
    }
    if imageErr != nil {
        return ProcessedContent{}, fmt.Errorf("图像处理失败: %w", imageErr)
    }
    
    return ProcessedContent{
        Text:  textResult,
        Image: imageResult,
    }, nil
}
```

### 3. 🚦 企业级错误处理策略

#### 分层错误处理

```go
// 定义错误类型
type ProcessingError struct {
    Stage   string
    Code    string
    Message string
    Cause   error
}

func (e *ProcessingError) Error() string {
    return fmt.Sprintf("[%s] %s: %s", e.Stage, e.Code, e.Message)
}

// 包装业务错误
func wrapBusinessError(stage, code, message string, err error) error {
    return &ProcessingError{
        Stage:   stage,
        Code:    code,
        Message: message,
        Cause:   err,
    }
}

// ✅ 具有重试机制的节点
func createResilientNode(operation func(context.Context, interface{}) (interface{}, error)) compose.Lambda {
    return compose.InvokableLambda(func(ctx context.Context, input interface{}) (interface{}, error) {
        maxRetries := 3
        backoff := time.Second
        
        for attempt := 1; attempt <= maxRetries; attempt++ {
            result, err := operation(ctx, input)
            if err == nil {
                return result, nil
            }
            
            // 检查是否是可重试的错误
            if !isRetryableError(err) {
                return nil, wrapBusinessError("processing", "FATAL_ERROR", "不可重试的错误", err)
            }
            
            if attempt == maxRetries {
                return nil, wrapBusinessError("processing", "MAX_RETRY_EXCEEDED", 
                    fmt.Sprintf("重试%d次后仍然失败", maxRetries), err)
            }
            
            // 指数退避
            select {
            case <-ctx.Done():
                return nil, ctx.Err()
            case <-time.After(backoff):
                backoff *= 2
            }
            
            log.Printf("第%d次重试，等待%v后继续...", attempt, backoff)
        }
        
        return nil, wrapBusinessError("processing", "UNEXPECTED", "不应该到达这里", nil)
    })
}

// 判断错误是否可重试
func isRetryableError(err error) bool {
    switch err.(type) {
    case *net.OpError, *url.Error:
        return true
    case *ProcessingError:
        procErr := err.(*ProcessingError)
        return procErr.Code == "TEMPORARY_FAILURE" || procErr.Code == "RATE_LIMIT"
    default:
        return false
    }
}
```

#### 全链路错误监控

```go
type ErrorCollector struct {
    errors []ProcessingError
    mu     sync.RWMutex
}

func (ec *ErrorCollector) RecordError(err error) {
    ec.mu.Lock()
    defer ec.mu.Unlock()
    
    if procErr, ok := err.(*ProcessingError); ok {
        ec.errors = append(ec.errors, *procErr)
    } else {
        ec.errors = append(ec.errors, ProcessingError{
            Stage:   "unknown",
            Code:    "GENERIC_ERROR",
            Message: err.Error(),
            Cause:   err,
        })
    }
}

func (ec *ErrorCollector) GetErrorSummary() ErrorSummary {
    ec.mu.RLock()
    defer ec.mu.RUnlock()
    
    summary := ErrorSummary{
        TotalErrors: len(ec.errors),
        ByStage:     make(map[string]int),
        ByCode:      make(map[string]int),
    }
    
    for _, err := range ec.errors {
        summary.ByStage[err.Stage]++
        summary.ByCode[err.Code]++
    }
    
    return summary
}
```

### 4. 📊 生产级性能监控

#### 全方位性能指标收集

```go
type PerformanceCollector struct {
    nodeMetrics map[string]*NodeMetrics
    mu          sync.RWMutex
}

type NodeMetrics struct {
    ExecutionCount int64
    TotalDuration  time.Duration
    MinDuration    time.Duration
    MaxDuration    time.Duration
    ErrorCount     int64
    LastError      time.Time
}

func (pc *PerformanceCollector) WrapNode(nodeName string, node compose.Lambda) compose.Lambda {
    return compose.InvokableLambda(func(ctx context.Context, input interface{}) (interface{}, error) {
        start := time.Now()
        
        result, err := node.Invoke(ctx, input)
        
        duration := time.Since(start)
        pc.recordMetrics(nodeName, duration, err != nil)
        
        return result, err
    })
}

func (pc *PerformanceCollector) recordMetrics(nodeName string, duration time.Duration, hasError bool) {
    pc.mu.Lock()
    defer pc.mu.Unlock()
    
    if pc.nodeMetrics == nil {
        pc.nodeMetrics = make(map[string]*NodeMetrics)
    }
    
    metrics, exists := pc.nodeMetrics[nodeName]
    if !exists {
        metrics = &NodeMetrics{
            MinDuration: duration,
            MaxDuration: duration,
        }
        pc.nodeMetrics[nodeName] = metrics
    }
    
    metrics.ExecutionCount++
    metrics.TotalDuration += duration
    
    if duration < metrics.MinDuration {
        metrics.MinDuration = duration
    }
    if duration > metrics.MaxDuration {
        metrics.MaxDuration = duration
    }
    
    if hasError {
        metrics.ErrorCount++
        metrics.LastError = time.Now()
    }
}

// 性能报告生成
func (pc *PerformanceCollector) GenerateReport() PerformanceReport {
    pc.mu.RLock()
    defer pc.mu.RUnlock()
    
    report := PerformanceReport{
        GeneratedAt: time.Now(),
        Nodes:       make(map[string]NodeReport),
    }
    
    for nodeName, metrics := range pc.nodeMetrics {
        avgDuration := time.Duration(0)
        if metrics.ExecutionCount > 0 {
            avgDuration = metrics.TotalDuration / time.Duration(metrics.ExecutionCount)
        }
        
        errorRate := float64(0)
        if metrics.ExecutionCount > 0 {
            errorRate = float64(metrics.ErrorCount) / float64(metrics.ExecutionCount)
        }
        
        report.Nodes[nodeName] = NodeReport{
            ExecutionCount: metrics.ExecutionCount,
            AverageDuration: avgDuration,
            MinDuration:    metrics.MinDuration,
            MaxDuration:    metrics.MaxDuration,
            ErrorRate:      errorRate,
            LastError:      metrics.LastError,
        }
    }
    
    return report
}
```

#### 实时性能告警

```go
type PerformanceAlerts struct {
    thresholds map[string]AlertThreshold
    alertChan  chan AlertEvent
}

type AlertThreshold struct {
    MaxDuration time.Duration
    MaxErrorRate float64
}

type AlertEvent struct {
    NodeName    string
    AlertType   string
    Value       interface{}
    Threshold   interface{}
    Timestamp   time.Time
}

func (pa *PerformanceAlerts) CheckPerformance(report PerformanceReport) {
    for nodeName, nodeReport := range report.Nodes {
        threshold, exists := pa.thresholds[nodeName]
        if !exists {
            continue
        }
        
        // 检查执行时间告警
        if nodeReport.AverageDuration > threshold.MaxDuration {
            pa.alertChan <- AlertEvent{
                NodeName:  nodeName,
                AlertType: "HIGH_LATENCY",
                Value:     nodeReport.AverageDuration,
                Threshold: threshold.MaxDuration,
                Timestamp: time.Now(),
            }
        }
        
        // 检查错误率告警
        if nodeReport.ErrorRate > threshold.MaxErrorRate {
            pa.alertChan <- AlertEvent{
                NodeName:  nodeName,
                AlertType: "HIGH_ERROR_RATE",
                Value:     nodeReport.ErrorRate,
                Threshold: threshold.MaxErrorRate,
                Timestamp: time.Now(),
            }
        }
    }
}

---

## 🌟 高级特性与生产实践

### 1. 📈 Workflow - 字段级映射编排

**特性说明**：Workflow 提供更精细的字段级映射控制，适合复杂的数据转换场景。

```go
// 定义复杂的业务数据结构
type CustomerOrder struct {
    OrderID     string
    CustomerID  string
    Products    []Product
    TotalAmount float64
    OrderDate   time.Time
}

type Product struct {
    ProductID string
    Name      string
    Price     float64
    Quantity  int
}

type OrderAnalysis struct {
    OrderInfo    CustomerOrder
    RiskScore    float64
    RecommendProducts []Product
    ProcessingTime   time.Duration
}

// 使用 Workflow 进行字段级映射
func buildOrderProcessingWorkflow() *compose.Workflow[CustomerOrder, OrderAnalysis] {
    workflow := compose.NewWorkflow[CustomerOrder, OrderAnalysis]()
    
    // 风险评估节点
    riskAssessment := compose.InvokableLambda(func(order CustomerOrder) float64 {
        // 基于订单金额、客户历史等计算风险分数
        baseRisk := math.Min(order.TotalAmount/10000.0, 1.0) // 金额风险
        
        // 可以添加更多风险因子
        customerRisk := assessCustomerRisk(order.CustomerID)
        productRisk := assessProductRisk(order.Products)
        
        return (baseRisk + customerRisk + productRisk) / 3
    })
    
    // 推荐系统节点
    recommendationEngine := compose.InvokableLambda(func(order CustomerOrder) []Product {
        // 基于购买历史和当前订单推荐产品
        return generateRecommendations(order.CustomerID, order.Products)
    })
    
    // 字段映射配置
    workflow.MapField("OrderInfo", compose.Identity[CustomerOrder]()) // 直接映射原始订单
    workflow.MapField("RiskScore", riskAssessment)                    // 映射到风险评估结果
    workflow.MapField("RecommendProducts", recommendationEngine)      // 映射到推荐结果
    workflow.MapField("ProcessingTime", compose.InvokableLambda(func(_ CustomerOrder) time.Duration {
        return time.Since(time.Now()) // 处理时间会在实际执行时计算
    }))
    
    return workflow
}
```

### 2. 🔄 节点嵌套与图的嵌套

**特性说明**：支持将整个 Chain 或 Graph 作为节点嵌入到另一个编排中，实现模块化设计。

```go
// 构建文档预处理子图
func buildDocumentPreprocessingSubgraph() *compose.Graph[RawDocument, ProcessedDocument] {
    subgraph := compose.NewGraph[RawDocument, ProcessedDocument]()
    
    // 文档解析
    parseNode := compose.InvokableLambda(func(raw RawDocument) ParsedDocument {
        return ParsedDocument{
            Title:    extractTitle(raw.Content),
            Content:  extractContent(raw.Content),
            Metadata: extractMetadata(raw.Content),
        }
    })
    
    // 内容清洗
    cleanNode := compose.InvokableLambda(func(parsed ParsedDocument) CleanDocument {
        return CleanDocument{
            Title:        sanitizeText(parsed.Title),
            Content:      sanitizeText(parsed.Content),
            CleanedAt:    time.Now(),
        }
    })
    
    // 质量检查
    qualityNode := compose.InvokableLambda(func(clean CleanDocument) ProcessedDocument {
        quality := assessContentQuality(clean.Content)
        return ProcessedDocument{
            Document: clean,
            Quality:  quality,
            IsValid:  quality > 0.7,
        }
    })
    
    subgraph.AddNode("parse", parseNode)
    subgraph.AddNode("clean", cleanNode)
    subgraph.AddNode("quality", qualityNode)
    
    subgraph.AddEdge("parse", "clean")
    subgraph.AddEdge("clean", "quality")
    
    return subgraph
}

// 主处理图，嵌入预处理子图
func buildMainProcessingGraph() *compose.Graph[RawDocument, FinalResult] {
    mainGraph := compose.NewGraph[RawDocument, FinalResult]()
    
    // 嵌入文档预处理子图
    preprocessingSubgraph := buildDocumentPreprocessingSubgraph()
    
    // 向量化处理
    vectorizationNode := compose.InvokableLambda(func(doc ProcessedDocument) VectorizedDocument {
        if !doc.IsValid {
            return VectorizedDocument{} // 无效文档跳过向量化
        }
        
        vectors := generateEmbeddings(doc.Document.Content)
        return VectorizedDocument{
            Document: doc,
            Vectors:  vectors,
        }
    })
    
    // 索引存储
    indexingNode := compose.InvokableLambda(func(vec VectorizedDocument) FinalResult {
        if len(vec.Vectors) == 0 {
            return FinalResult{Success: false, Reason: "无法生成向量"}
        }
        
        indexID := storeInVectorDB(vec.Vectors, vec.Document.Document.Content)
        return FinalResult{
            Success:   true,
            IndexID:   indexID,
            Timestamp: time.Now(),
        }
    })
    
    // 将子图作为节点添加到主图
    mainGraph.AddSubgraph("preprocessing", preprocessingSubgraph)
    mainGraph.AddNode("vectorization", vectorizationNode)
    mainGraph.AddNode("indexing", indexingNode)
    
    // 连接子图和其他节点
    mainGraph.AddEdge("preprocessing", "vectorization")
    mainGraph.AddEdge("vectorization", "indexing")
    
    return mainGraph
}
```

### 3. 🎛️ 回调注入机制

**特性说明**：通过回调函数在编排执行过程中注入自定义逻辑，实现高度可定制化。

```go
type ExecutionCallback struct {
    OnNodeStart    func(nodeName string, input interface{})
    OnNodeComplete func(nodeName string, output interface{}, duration time.Duration)
    OnNodeError    func(nodeName string, err error)
    OnGraphStart   func()
    OnGraphComplete func(finalResult interface{})
}

// 创建带回调的执行器
func createCallbackEnabledGraph() *compose.Graph[string, ProcessResult] {
    graph := compose.NewGraph[string, ProcessResult]()
    
    // 创建回调实例
    callbacks := &ExecutionCallback{
        OnNodeStart: func(nodeName string, input interface{}) {
            log.Printf("🚀 节点 [%s] 开始执行，输入: %+v", nodeName, input)
            
            // 记录到监控系统
            metrics.IncrementCounter("node_execution_start", map[string]string{
                "node_name": nodeName,
            })
        },
        
        OnNodeComplete: func(nodeName string, output interface{}, duration time.Duration) {
            log.Printf("✅ 节点 [%s] 执行完成，耗时: %v，输出: %+v", nodeName, duration, output)
            
            // 记录性能指标
            metrics.RecordHistogram("node_execution_duration", duration.Seconds(), map[string]string{
                "node_name": nodeName,
            })
        },
        
        OnNodeError: func(nodeName string, err error) {
            log.Printf("❌ 节点 [%s] 执行失败: %v", nodeName, err)
            
            // 记录错误指标
            metrics.IncrementCounter("node_execution_error", map[string]string{
                "node_name": nodeName,
                "error_type": fmt.Sprintf("%T", err),
            })
            
            // 发送告警
            alerts.SendAlert(AlertLevel.ERROR, fmt.Sprintf("节点 %s 执行失败", nodeName), err)
        },
        
        OnGraphStart: func() {
            log.Printf("📊 图执行开始")
        },
        
        OnGraphComplete: func(finalResult interface{}) {
            log.Printf("🎉 图执行完成，最终结果: %+v", finalResult)
        },
    }
    
    // 包装节点以支持回调
    wrappedNode := func(nodeName string, originalNode compose.Lambda) compose.Lambda {
        return compose.InvokableLambda(func(ctx context.Context, input interface{}) (interface{}, error) {
            callbacks.OnNodeStart(nodeName, input)
            start := time.Now()
            
            result, err := originalNode.Invoke(ctx, input)
            duration := time.Since(start)
            
            if err != nil {
                callbacks.OnNodeError(nodeName, err)
                return nil, err
            }
            
            callbacks.OnNodeComplete(nodeName, result, duration)
            return result, nil
        })
    }
    
    // 添加包装后的节点
    node1 := wrappedNode("step1", compose.InvokableLambda(func(input string) string {
        time.Sleep(100 * time.Millisecond) // 模拟处理时间
        return "processed: " + input
    }))
    
    node2 := wrappedNode("step2", compose.InvokableLambda(func(input string) ProcessResult {
        return ProcessResult{
            Data:      input + " -> final",
            Success:   true,
            Timestamp: time.Now(),
        }
    }))
    
    graph.AddNode("step1", node1)
    graph.AddNode("step2", node2)
    graph.AddEdge("step1", "step2")
    
    return graph
}
```

### 4. 🎚️ 灵活的 Option 分配

**特性说明**：通过 Option 模式为编排提供灵活的配置选项。

```go
// 定义编排选项
type OrchestrationOptions struct {
    MaxConcurrency    int
    Timeout          time.Duration
    RetryPolicy      RetryPolicy
    CircuitBreaker   CircuitBreakerConfig
    RateLimiter      RateLimiterConfig
    Metrics          MetricsConfig
}

type RetryPolicy struct {
    MaxRetries  int
    BackoffFunc func(attempt int) time.Duration
}

type CircuitBreakerConfig struct {
    Threshold     int
    ResetTimeout  time.Duration
    HalfOpenMax   int
}

type RateLimiterConfig struct {
    RequestsPerSecond float64
    BurstSize         int
}

// Option 函数类型
type GraphOption func(*OrchestrationOptions)

// Option 构建函数
func WithMaxConcurrency(max int) GraphOption {
    return func(opts *OrchestrationOptions) {
        opts.MaxConcurrency = max
    }
}

func WithTimeout(timeout time.Duration) GraphOption {
    return func(opts *OrchestrationOptions) {
        opts.Timeout = timeout
    }
}

func WithRetryPolicy(maxRetries int, backoffFunc func(int) time.Duration) GraphOption {
    return func(opts *OrchestrationOptions) {
        opts.RetryPolicy = RetryPolicy{
            MaxRetries:  maxRetries,
            BackoffFunc: backoffFunc,
        }
    }
}

func WithCircuitBreaker(threshold int, resetTimeout time.Duration) GraphOption {
    return func(opts *OrchestrationOptions) {
        opts.CircuitBreaker = CircuitBreakerConfig{
            Threshold:    threshold,
            ResetTimeout: resetTimeout,
            HalfOpenMax:  3,
        }
    }
}

func WithRateLimit(rps float64, burstSize int) GraphOption {
    return func(opts *OrchestrationOptions) {
        opts.RateLimiter = RateLimiterConfig{
            RequestsPerSecond: rps,
            BurstSize:         burstSize,
        }
    }
}

// 创建可配置的图
func NewConfigurableGraph(options ...GraphOption) *compose.Graph[RequestData, ResponseData] {
    // 默认选项
    opts := &OrchestrationOptions{
        MaxConcurrency: 10,
        Timeout:        30 * time.Second,
        RetryPolicy: RetryPolicy{
            MaxRetries: 3,
            BackoffFunc: func(attempt int) time.Duration {
                return time.Duration(attempt) * time.Second
            },
        },
    }
    
    // 应用用户选项
    for _, option := range options {
        option(opts)
    }
    
    graph := compose.NewGraph[RequestData, ResponseData]()
    
    // 应用并发控制
    if opts.MaxConcurrency > 0 {
        graph = graph.WithConcurrencyLimit(opts.MaxConcurrency)
    }
    
    // 应用超时控制
    if opts.Timeout > 0 {
        graph = graph.WithTimeout(opts.Timeout)
    }
    
    // 创建带重试的节点
    resilientNode := compose.InvokableLambda(func(ctx context.Context, input RequestData) (ResponseData, error) {
        for attempt := 1; attempt <= opts.RetryPolicy.MaxRetries; attempt++ {
            result, err := processRequest(ctx, input)
            if err == nil {
                return result, nil
            }
            
            if attempt == opts.RetryPolicy.MaxRetries {
                return ResponseData{}, err
            }
            
            backoff := opts.RetryPolicy.BackoffFunc(attempt)
            select {
            case <-ctx.Done():
                return ResponseData{}, ctx.Err()
            case <-time.After(backoff):
                // 继续重试
            }
        }
        
        return ResponseData{}, fmt.Errorf("不应该到达这里")
    })
    
    graph.AddNode("main_processor", resilientNode)
    
    return graph
}

// 使用示例
func createProductionGraph() *compose.Graph[RequestData, ResponseData] {
    return NewConfigurableGraph(
        WithMaxConcurrency(50),                    // 最大并发50
        WithTimeout(60*time.Second),               // 60秒超时
        WithRetryPolicy(5, exponentialBackoff),    // 5次重试，指数退避
        WithCircuitBreaker(10, 30*time.Second),    // 断路器：10次失败后断开，30秒后重试
        WithRateLimit(1000.0, 100),               // 限流：每秒1000请求，突发100
    )
}

// 指数退避函数
func exponentialBackoff(attempt int) time.Duration {
    base := 100 * time.Millisecond
    return time.Duration(math.Pow(2, float64(attempt-1))) * base
}
```

---

## 📚 总结与最佳实践清单

编排就像是 AI 应用的"指挥中心"，通过学习本指南，你已经掌握了：

### 🎯 核心能力
- 🔒 **类型对齐与确定性**：编译时类型检查，确保系统稳定
- 🌊 **流式处理自动化**：自动处理单个与流式组件的转换  
- 🔧 **智能类型转换**：支持复杂数据结构的合并与转换
- 🚀 **双引擎支持**：DAG 与 Pregel 引擎适应不同场景

### ⚡ 编排模式精通
- 🔗 **Chain**：线性处理，简单可靠
- 🕸️ **Graph**：并行处理，高效灵活  
- 📊 **Workflow**：字段级映射，精细控制
- 🍴 **扇出模式**：并行分发，提高吞吐
- 🎯 **扇入模式**：智能合并，综合决策
- 🔄 **条件分支**：动态路由，业务灵活

### 🛡️ 生产级最佳实践
- 📏 **科学决策**：Chain vs Graph 选择策略
- 🔧 **类型设计**：单一职责，接口清晰
- 🚦 **错误处理**：分层处理，优雅降级
- 📊 **性能监控**：全链路监控，实时告警
- 🎛️ **回调注入**：高度可定制化
- 🎚️ **Option 模式**：灵活配置管理

### 💫 设计原则
- 🔒 **外部变量只读**：确保可重复性和可测试性
- 🌊 **自动流式补全**：开发者专注业务逻辑
- 🎯 **智能类型转换**：处理复杂数据流转换
- 🚀 **引擎选择**：根据场景选择合适的执行引擎

通过合理运用 Eino 的编排系统，你可以构建出**既强大又稳定、既高效又优雅**的 AI 应用！

---

*"优秀的编排不仅让系统运行，更让系统运行得智能、稳定、高效。"* ✨

### 📖 延伸阅读
- [Eino 官方文档](https://www.cloudwego.io/zh/docs/eino/)
- [编排设计模式](https://www.cloudwego.io/zh/docs/eino/core_modules/chain_and_graph_orchestration/)
- [性能优化指南](https://www.cloudwego.io/zh/docs/eino/best_practices/)