package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/cloudwego/eino-ext/components/document/transformer/splitter/markdown"
	"github.com/cloudwego/eino/schema"
)

// =============================================================================
//
//  文件: transformer_base/main.go
//  功能: 本示例展示了如何独立使用 Document Transformer 组件中的 Markdown Header Splitter。
//        该功能对于文档预处理至关重要，尤其是在 RAG (Retrieval-Augmented Generation) 场景中，
//        通过将大型 Markdown 文档按标题结构（如 H1, H2）分割成逻辑相关的块，
//        可以显著提高后续检索和生成步骤的准确性和效率。
//
// =============================================================================

func main() {
	ctx := context.Background()

	// --- 步骤 1: 定义源文档 ---
	// 创建一个包含多级 Markdown 标题的示例文档。
	// splitter 将依据这些标题（此处为 '##'）对文档进行切分。
	markdownContent := `
# Eino 框架介绍

Eino 是一个先进的大模型应用开发框架。

## 核心组件
Eino 提供了多种核心组件，包括 Model, Retriever, Indexer, 和 Transformer。
这些组件可以帮助开发者快速构建强大的 RAG 应用。

## Transformer 详解
Transformer 组件负责文档的预处理。
它可以将长文档分割成小块，过滤无关信息，或进行格式转换。
这是确保检索质量的关键一步。

## 快速开始
要开始使用 Eino，请参考我们的官方文档和示例代码。
`
	doc := &schema.Document{
		ID:       "eino-intro-doc",
		Content:  markdownContent,
		MetaData: map[string]interface{}{"source": "official-docs"},
	}
	fmt.Println("--- 原始文档 ---")
	fmt.Printf("ID: %s\n内容长度: %d\n", doc.ID, len(doc.Content))
	fmt.Println(strings.Repeat("-", 20))

	// --- 步骤 2: 配置并初始化 Markdown Header Splitter ---
	// NewHeaderSplitter 用于创建一个基于标题的分割器实例。
	// HeaderConfig.Headers 是一个 map，用于定义分割规则：
	// - key:   指定用于分割的 Markdown 标题标记，例如 "##" 代表二级标题。
	// - value: 指定一个元数据键名。分割后，每个新文档块的 MetaData 中会增加这个键，
	//          其对应的值是该块的标题内容。
	// 例如，规则 {"##": "section_header"} 会在遇到 "## 核心组件" 时，
	// 将 "核心组件" 作为值，"section_header" 作为键，存入新文档块的元数据中。
	splitter, err := markdown.NewHeaderSplitter(ctx, &markdown.HeaderConfig{
		Headers: map[string]string{
			"##": "section_header", // 按二级标题分割，并将标题内容存入 "section_header" 元数据字段
		},
	})
	if err != nil {
		log.Fatalf("创建 HeaderSplitter 失败: %v", err)
	}

	// --- 步骤 3: 执行文档转换 ---
	// 调用 splitter 的 Transform 方法，传入原始文档列表。
	// 该方法会根据初始化时定义的规则，返回一个被分割后的新文档列表。
	fmt.Println("\n正在调用 Transform 方法进行分割...")
	transformedDocs, err := splitter.Transform(ctx, []*schema.Document{doc})
	if err != nil {
		log.Fatalf("转换文档失败: %v", err)
	}

	// --- 步骤 4: 验证并输出结果 ---
	// 遍历转换后的文档列表，打印每个文档块的内容和元数据。
	// 检查元数据可以验证分割器是否按预期工作，将标题信息正确地附加到了每个块上。
	fmt.Printf("\n--- 分割完成，共得到 %d 个新文档 ---\n", len(transformedDocs))
	for i, d := range transformedDocs {
		fmt.Printf("\n--- 文档块 %d ---\n", i+1)
		fmt.Printf("ID: %s\n", d.ID)
		fmt.Printf("内容:\n%s\n", d.Content)
		fmt.Printf("元数据: %v\n", d.MetaData)
	}
	fmt.Println(strings.Repeat("-", 30))
}
