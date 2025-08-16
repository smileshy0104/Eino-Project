package main

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/schema"
)

// --- 自定义 Option ---

// SentenceSplitterOptions 定义了我们自定义 Transformer 的特定选项。
type SentenceSplitterOptions struct {
	MinSentenceLength int // 句子的最小长度，低于此长度的句子将被忽略
}

// WithMinSentenceLength 是一个 Option 函数，用于设置最小句子长度。
// 它遵循文档中描述的自定义 Option 实现模式。
func WithMinSentenceLength(length int) document.TransformerOption {
	// WrapTransformerImplSpecificOptFn 用于将特定于实现的 Option 函数包装成通用的 TransformerOption 类型。
	return document.WrapTransformerImplSpecificOptFn(func(o *SentenceSplitterOptions) {
		o.MinSentenceLength = length
	})
}

// --- 自定义 Transformer ---

// SentenceSplitter 是一个自定义的 Transformer，它按句子分割文档。
type SentenceSplitter struct {
	defaultMinSentenceLength int
}

// NewSentenceSplitter 创建一个新的 SentenceSplitter 实例。
func NewSentenceSplitter() *SentenceSplitter {
	return &SentenceSplitter{
		defaultMinSentenceLength: 5, // 默认句子的最小长度为5个字符
	}
}

// Transform 实现了 Transformer 接口的核心方法。
func (s *SentenceSplitter) Transform(ctx context.Context, src []*schema.Document, opts ...document.TransformerOption) ([]*schema.Document, error) {
	// 1. 处理自定义 Option
	options := &SentenceSplitterOptions{
		MinSentenceLength: s.defaultMinSentenceLength,
	}
	// GetTransformerImplSpecificOptions 是一个辅助函数，用于解析用户传入的 opts 并更新 options 结构体。
	document.GetTransformerImplSpecificOptions(options, opts...)

	log.Printf("开始转换文档，最小句子长度: %d\n", options.MinSentenceLength)

	var transformedDocs []*schema.Document
	// 使用正则表达式按句末标点符号分割句子
	re := regexp.MustCompile(`[.!?]`)

	for _, doc := range src {
		sentences := re.Split(doc.Content, -1)
		for i, sentence := range sentences {
			// 去除首尾空格
			trimmedSentence := strings.TrimSpace(sentence)
			// 应用最小长度过滤
			if len(trimmedSentence) >= options.MinSentenceLength {
				newDoc := &schema.Document{
					ID:      fmt.Sprintf("%s-part%d", doc.ID, i),
					Content: trimmedSentence,
					MetaData: map[string]interface{}{
						"original_doc_id": doc.ID,
						"part_num":        i,
					},
				}
				transformedDocs = append(transformedDocs, newDoc)
			}
		}
	}

	return transformedDocs, nil
}

func main() {
	ctx := context.Background()

	// 1. 初始化我们自定义的 Transformer
	splitter := NewSentenceSplitter()
	fmt.Println("自定义 Transformer (SentenceSplitter) 初始化成功。")

	// 2. 准备一个待转换的文档
	originalDoc := &schema.Document{
		ID:      "news-article-01",
		Content: "Eino is a framework. It simplifies building LLM apps. Try it. It's great!",
	}

	// 3. 执行第一次转换 (使用默认的最小句子长度)
	fmt.Println("\n--- 第一次转换 (使用默认最小句子长度) ---")
	docs, err := splitter.Transform(ctx, []*schema.Document{originalDoc})
	if err != nil {
		log.Fatalf("转换失败: %v", err)
	}

	fmt.Printf("转换后得到 %d 个文档:\n", len(docs))
	for _, doc := range docs {
		fmt.Printf("  - ID: %s, 内容: '%s'\n", doc.ID, doc.Content)
	}

	// 4. 再次执行转换，但这次使用 Option 来设置一个更大的最小句子长度
	fmt.Println("\n--- 第二次转换 (使用 WithMinSentenceLength(10) 选项) ---")
	docs, err = splitter.Transform(ctx, []*schema.Document{originalDoc}, WithMinSentenceLength(10))
	if err != nil {
		log.Fatalf("转换失败: %v", err)
	}

	fmt.Printf("转换后得到 %d 个文档:\n", len(docs))
	for _, doc := range docs {
		fmt.Printf("  - ID: %s, 内容: '%s'\n", doc.ID, doc.Content)
	}
}
