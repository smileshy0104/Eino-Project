# Eino Retriever (Milvus) 使用说明

## 1. 基本介绍

`Retriever` 组件是 Eino 框架中用于从各种数据源（特别是向量数据库）检索文档的核心组件。它的主要职责是根据用户的文本查询（query），从文档库中高效地找出最相关的文档。

本文档将重点介绍如何使用 `Retriever` 组件从 Milvus 向量数据库中检索数据。

**重要前提**: 本示例假设您已经通过其他方式（例如使用 `indexer_demo`）将文档和它们的向量表示存入了 Milvus 集合中。`Retriever` 的职责是查询，而不是写入。

## 2. 核心组件与配置

### 2.1 `Embedder` 组件

`Retriever` 在工作时，需要先将用户的文本查询（`query`）转换成一个向量，然后才能在 Milvus 中进行相似度搜索。这个转换工作由 `Embedder` 组件完成。

```go
// 初始化一个 Embedder，例如使用火山方舟服务
embedder, err := ark.NewEmbedder(ctx, &ark.EmbeddingConfig{
    APIKey:  viper.GetString("ARK_API_KEY"),
    Model:   viper.GetString("EMBEDDER_MODEL"),
})
```
**注意**: 您用于检索的 `Embedder` 模型，必须与当初您用来索引文档的 `Embedder` 模型完全相同，否则向量空间不匹配，将无法获得正确的检索结果。

### 2.2 `RetrieverConfig`

要初始化一个 `milvus.Retriever`，你需要提供一个 `RetrieverConfig` 结构体。

```go
import "github.com/cloudwego/eino-ext/components/retriever/milvus"

// 创建 Milvus 客户端
client, _ := cli.NewClient(ctx, cli.Config{
    Address: viper.GetString("MILVUS_ADDRESS"),
})

// 配置 Retriever
retrieverCfg := &milvus.RetrieverConfig{
    Client:     client,
    Collection: "your_collection_name", // 目标集合
    Embedding:  embedder,               // 上一步创建的 Embedder 实例
}
```

**关键配置项**:

-   `Client`: 一个已经实例化的 Milvus Go SDK 客户端。
-   `Collection`: 指定要在哪个集合中进行检索。
-   `Embedding`: **必须**提供一个 `Embedder` 组件实例。

## 3. 完整使用示例

下面的代码演示了如何配置并运行一个独立的 `Retriever` 来查询 Milvus。完整代码请参考 `retriever_demo/main.go`。

```go
package main

import (
	"context"
	"fmt"
	"log"

	// ... 其他 import
	"github.com/cloudwego/eino-ext/components/embedding/ark"
	"github.com/cloudwego/eino-ext/components/retriever/milvus"
	cli "github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/spf13/viper"
)

func runRetrieverExample() {
	ctx := context.Background()

	// 1. 初始化 Embedder
	embedder, err := ark.NewEmbedder(ctx, &ark.EmbeddingConfig{
		APIKey:  viper.GetString("ARK_API_KEY"),
		Model:   viper.GetString("EMBEDDER_MODEL"),
	})
	if err != nil {
		log.Fatalf("创建 Embedder 失败: %v", err)
	}

	// 2. 创建 Milvus 客户端
	client, err := cli.NewClient(ctx, cli.Config{
		Address: viper.GetString("MILVUS_ADDRESS"),
	})
	if err != nil {
		log.Fatalf("创建 Milvus 客户端失败: %v", err)
	}

	// 3. 配置并创建 Retriever 实例
	cfg := &milvus.RetrieverConfig{
		Client:     client,
		Collection: viper.GetString("MILVUS_COLLECTION"),
		Embedding:  embedder,
	}
	retriever, err := milvus.NewRetriever(ctx, cfg)
	if err != nil {
		log.Fatalf("创建 Retriever 失败: %v", err)
	}
	fmt.Println("Retriever 初始化成功！")

	// 4. 准备查询并执行检索
	query := "Eino 是什么？"
	fmt.Printf("\n正在使用查询 \"%s\" 调用 Retrieve 方法...\n", query)
	retrievedDocs, err := retriever.Retrieve(ctx, query)
	if err != nil {
		log.Fatalf("检索文档失败: %v", err)
	}

	// 5. 打印结果
	fmt.Println("\n--- 检索成功 ---")
	for _, doc := range retrievedDocs {
		fmt.Printf("  - ID: %s, 内容: %s\n", doc.ID, doc.Content)
	}
}

func main() {
    // (省略了 Viper 配置加载代码)
	runRetrieverExample()
}
```

## 4. 如何运行

1.  **确保数据已索引**: 首先，运行 `indexer_demo`，确保 Milvus 中有可供查询的数据。
    ```bash
    go run ../indexer_demo/main.go
    ```
2.  **运行 Retriever**: 然后，在 `retriever_demo` 目录下运行 `main.go`。
    ```bash
    go run main.go
    ```

程序将会输出与查询 "Eino 是什么？" 最相关的文档。