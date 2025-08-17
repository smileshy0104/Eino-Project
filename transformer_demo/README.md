# Eino Document Transformer 使用文档

## 1. 概述

`Document Transformer` 是 Eino 框架中用于对文档进行预处理和转换的核心组件。它的主要作用是在文档被索引或检索之前，对其进行结构化处理，最常见的应用场景是**文档分割 (Splitting)**。

在构建 RAG (Retrieval-Augmented Generation) 应用时，将冗长的原始文档分割成更小的、语义集中的块 (chunks) 是至关重要的一步。这能极大地提升后续向量检索的精确度，确保返回给大语言模型的是最相关、最精炼的上下文信息。

本文档将重点介绍如何使用 `Document Transformer` 的一个具体实现：`Markdown Header Splitter`。

## 2. 核心接口与概念

### 2.1 `Transform` 方法

`Transformer` 的核心功能由 `Transform` 方法定义：

```go
type Transformer interface {
    Transform(ctx context.Context, src []*schema.Document, opts ...TransformerOption) ([]*schema.Document, error)
}
```

-   **功能**: 接收一个 `Document` 列表，对其进行转换，并返回一个新的 `Document` 列表。
-   **参数**:
    -   `src`: 待处理的原始文档列表。
    -   `opts`: 用于配置转换行为的可选参数（具体实现各不相同）。
-   **返回值**: 经过转换处理后的新文档列表。

### 2.2 文档分割 (Splitting)

文档分割是将一个 `*schema.Document` 对象转换为多个 `*schema.Document` 对象的过程。分割后的每个 `Document` 都代表原始文档的一个片段或一个块。

Eino 提供了多种分割策略，例如：
-   **`Markdown Header Splitter`**: 根据 Markdown 的标题层级（如 `##`, `###`）进行分割。
-   **`Text Splitter`**: 根据字符数、句子或特定的分隔符进行分割。

## 3. 使用示例：Markdown Header Splitter

[`transformer_demo/main.go`](transformer_demo/main.go) 文件中的代码演示了如何独立使用 `Markdown Header Splitter`。

### 步骤 1: 准备原始文档

首先，我们创建一个 `*schema.Document` 对象，其 `Content` 字段包含一个完整的 Markdown 格式字符串。

```go
longMarkdownDoc := &schema.Document{
    ID: "eino-intro-doc",
    Content: `
# Eino 框架介绍
Eino 是一个先进的大模型应用开发框架。
## 核心组件
Eino 提供了多种核心组件...
## Transformer 详解
Transformer 组件负责文档的预处理...`,
    MetaData: map[string]interface{}{"source": "official-docs"},
}
```

### 步骤 2: 初始化 Splitter

我们实例化一个 `markdown.NewHeaderSplitter`。在配置中，我们指定将 `##` (二级标题) 作为分割文档的依据。

```go
import "github.com/cloudwego/eino-ext/components/document/transformer/splitter/markdown"

splitter, err := markdown.NewHeaderSplitter(ctx, &markdown.HeaderConfig{
    Headers: map[string]string{
        "##": "Header 2", // 使用 "##" 作为分割符
    },
})
```

### 步骤 3: 执行转换

调用 `splitter` 的 `Transform` 方法，并传入包含原始文档的列表。

```go
// Transform 方法返回一个新的文档列表，其中包含了分割后的所有块
docsToStore, err := splitter.Transform(ctx, []*schema.Document{longMarkdownDoc})
```

### 步骤 4: 查看结果

转换完成后，`docsToStore` 将是一个包含多个 `*schema.Document` 的列表。原始文档被成功地按二级标题分割成了多个独立的块，每个块都保留了原始的元数据，并拥有一个新的、唯一的 ID。

```
--- 分割完成，共得到 3 个新文档 ---

--- 文档块 1 ---
ID: eino-intro-doc_0
内容:
## 核心组件
Eino 提供了多种核心组件...
元数据: map[Header 2:核心组件 source:official-docs]

--- 文档块 2 ---
...
```

## 4. 如何运行

直接在项目根目录下执行以下命令：
```bash
go run transformer_demo/main.go
```
程序将打印出原始文档信息，以及分割后的每一个文档块的详细内容。