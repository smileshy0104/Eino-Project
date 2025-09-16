# Markdown 示例文档

这是一个 **Markdown 格式**的示例文档。

## 功能特性

DocumentLoader 支持以下功能：

### 1. 多格式支持
- Markdown (.md)
- 纯文本 (.txt)
- HTML (.html)
- PDF (.pdf)

### 2. 智能解析
- 自动格式识别
- 元数据提取
- 内容结构分析

## 代码示例

```go
// 创建 DocumentLoader
loader, err := loader.NewDocumentLoader(ctx, &loader.Config{
    Parser: extParser,
})

// 加载文档
docs, err := loader.Load(ctx, fileSource)
```

## 总结

DocumentLoader 是一个强大而灵活的文档加载组件，为 AI 应用的文档处理提供了坚实的基础。