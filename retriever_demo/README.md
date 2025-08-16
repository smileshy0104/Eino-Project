# Eino Retriever (Milvus) 使用文档

## 1. 概述

`Retriever` 是 Eino 框架中负责从向量数据库等数据源检索信息的关键组件。它是构建 RAG (Retrieval-Augmented Generation, 检索增强生成) 应用的核心。

本文档将详细介绍如何使用 `Retriever` 组件与 Milvus 向量数据库进行交互，涵盖独立检索和在 RAG 应用中编排使用两种场景。

**重要前提**: `Retriever` 负责“读取”和“查询”。在运行任何 `Retriever` 示例之前，您**必须**先运行 `indexer_demo` 来完成数据的“写入”和“索引”工作。

## 2. 核心概念与配置

### 2.1 `Retrieve` 方法定义

Retriever 的核心功能由一个简单的接口定义：

```go
type Retriever interface {
    Retrieve(ctx context.Context, query string, opts ...Option) ([]*schema.Document, error)
}
```

**参数说明**:

-   `ctx context.Context`: Go 的标准上下文，用于控制 API 调用超时、传递请求范围的数据等。
-   `query string`: 用户的自然语言查询。
-   `opts ...Option`: 一系列可选参数，用于精细化控制检索行为。例如：
    -   `retriever.WithTopK(5)`: 指定返回最相关的 5 个文档。
    -   `retriever.WithScoreThreshold(0.7)`: 只返回相似度分数高于 0.7 的文档。

**返回值**:

-   `[]*schema.Document`: 检索到的文档列表。
-   `error`: 检索过程中发生的任何错误。

### 2.2 工作流程

`Retriever` 的工作流程如下：

1.  接收一个自然语言查询 (query string)。
2.  使用内部配置的 `Embedder` 组件将该查询字符串转换为向量。
3.  向 Milvus 发起向量相似度搜索请求。
4.  将 Milvus 返回的结果（包含ID、元数据等）解析为 Eino 标准的 `[]*schema.Document` 格式。

### 2.2 关键配置 (`milvus.RetrieverConfig`)

要初始化一个 Milvus Retriever，你需要提供以下核心配置：

```go
retrieverCfg := &milvus.RetrieverConfig{
    // 1. Milvus 客户端实例
    Client:       client,
    // 2. 要查询的集合名称
    Collection:   viper.GetString("MILVUS_COLLECTION"),
    // 3. 用于查询向量化的 Embedder 组件
    Embedding:    embedder,
    // 4. (重要) 指定需要从 Milvus 返回的字段
    OutputFields: []string{"content", "metadata"},
}
```

-   `Client`: 一个已连接的 Milvus Go SDK 客户端。
-   `Collection`: 集合名称，必须与 `indexer_demo` 中创建的集合名称一致。
-   `Embedding`: 一个 `Embedder` 实例。**它使用的模型必须与 `indexer_demo` 中索引数据时使用的模型完全相同**，否则向量空间不匹配，无法获得准确结果。
-   `OutputFields`: Milvus 的向量搜索默认只返回 ID。**必须**在此处明确指定需要一并返回的其它字段（如 `content`），否则检索出的 `Document` 对象中将只有 ID，内容为空。

## 3. 使用场景

`retriever_demo` 目录提供了两种使用 `Retriever` 的示例，您可以通过修改 `retriever_demo/main.go` 中的 `exampleToRun` 变量来选择运行哪一个。

### 场景一：独立使用 (Standalone)

此场景演示了 `Retriever` 的基本用法：接收一个查询，返回相关文档列表。

**核心逻辑**:

```go
// 1. 初始化 Embedder 和 Milvus Client
// ...

// 2. 配置并创建 Retriever 实例
cfg := &milvus.RetrieverConfig{ /* ... */ }
retriever, err := milvus.NewRetriever(ctx, cfg)

// 3. 执行检索
query := "Eino 框架是什么？"
retrievedDocs, err := retriever.Retrieve(ctx, query)

// 4. 处理结果
for _, doc := range retrievedDocs {
    fmt.Printf("  - ID: %s, 内容: %s\n", doc.ID, doc.Content)
}
```

### 场景二：构建 RAG 应用 (Chain)

此场景演示了 `Retriever` 的高级用法：将其作为一个节点编排进一个 `compose.Chain` 中，与大语言模型 (LLM) 协同工作，构建一个完整的 RAG 应用。

**核心逻辑**:

```go
// 1. 初始化所有组件 (Embedder, Milvus Client, Retriever, ChatModel)
// ...

// 2. 创建一个新的 Chain
chain := compose.NewChain[string, *schema.Message]()

// 3. 编排工作流
// 步骤 1: 将 string 类型的 query 转换为 map，以供后续节点使用
chain.AppendLambda(/* ... */)

// 步骤 2: 添加 Retriever 节点，进行文档检索
chain.AppendRetriever(retriever, compose.WithInputKey("query"), compose.WithOutputKey("docs"))

// 步骤 3: 添加 Lambda 节点，根据检索结果构建最终的 Prompt
chain.AppendLambda(compose.InvokableLambda(createPromptFromDocs))

// 步骤 4: 添加 ChatModel 节点，生成最终答案
chain.AppendChatModel(model)

// 4. 编译并运行 Chain
runnable, _ := chain.Compile(ctx)
finalAnswer, _ := runnable.Invoke(ctx, "Eino 框架是什么？")

// 5. 打印结果
fmt.Println(finalAnswer.Content)
```