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
//  文件: transformer_demo/main.go
//  功能: 演示如何独立使用 Document Transformer 组件，特别是 Markdown Header Splitter，
//        来将一个长文档按标题分割成多个小文档。
//
// =============================================================================

func main() {
	ctx := context.Background()

	// --- 步骤 1: 准备一个包含 Markdown 标题的示例文档 ---
	// 这个文档包含了多个二级标题 (##)，我们将根据这些标题进行分割。
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

	// --- 步骤 2: 初始化 Markdown Header Splitter ---
	// 我们配置 splitter，让它根据二级标题 "##" 来分割文档。
	// Headers map 的 value 可以用来给分割后的文档的 metadata 添加额外信息。
	// 例如，{"##": "h2_title"} 会在分割出的文档块的 metadata 中添加 {"h2_title": "核心组件"} 这样的键值对。
	splitter, err := markdown.NewHeaderSplitter(ctx, &markdown.HeaderConfig{
		Headers: map[string]string{
			"##": "Header 2",
		},
	})
	if err != nil {
		log.Fatalf("创建 HeaderSplitter 失败: %v", err)
	}

	// --- 步骤 3: 调用 Transform 方法进行分割 ---
	fmt.Println("\n正在调用 Transform 方法进行分割...")
	transformedDocs, err := splitter.Transform(ctx, []*schema.Document{doc})
	if err != nil {
		log.Fatalf("转换文档失败: %v", err)
	}

	// --- 步骤 4: 打印分割后的结果 ---
	fmt.Printf("\n--- 分割完成，共得到 %d 个新文档 ---\n", len(transformedDocs))
	for i, d := range transformedDocs {
		fmt.Printf("\n--- 文档块 %d ---\n", i+1)
		fmt.Printf("ID: %s\n", d.ID)
		fmt.Printf("内容:\n%s\n", d.Content)
		fmt.Printf("元数据: %v\n", d.MetaData)
	}
	fmt.Println(strings.Repeat("-", 30))
}
