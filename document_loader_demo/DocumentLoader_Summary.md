# 📂 Eino DocumentLoader 组件完全指南

## 🚀 快速开始

### 🛠️ 配置文件
项目使用 `config.yaml` 配置文件，也可以通过环境变量设置：
```yaml
ARK_API_KEY: "${ARK_API_KEY}"
EMBEDDER_MODEL: "doubao-embedding-text-240715"
MILVUS_ADDRESS: "localhost:19530"
MILVUS_COLLECTION: "eino_demo_collection"
```

---

## 📖 基本介绍

`DocumentLoader` 组件是一个专门用于**从多种数据源加载文档**的智能组件。它的主要作用是将不同来源的文档（本地文件、网络资源、云存储）统一转换为标准的 `Document` 格式，为后续的文档处理流程奠定基础。这个组件在 AI 应用开发中扮演着**"文档获取引擎"**的角色。

### 🎯 核心价值

在传统的文档处理中，我们需要为不同数据源编写不同的处理逻辑。而 DocumentLoader 组件让我们能够：

```
传统方式：多套处理逻辑 + 格式不一致 + 重复代码  ❌
DocumentLoader：统一接口 + 标准格式 + 插件化扩展  ✅
```

### 🚀 主要应用场景

- **📁 本地文件加载**: 批量处理本地文档文件，支持多种格式
- **🌐 网络资源获取**: 从 URL 动态加载网页内容和在线文档
- **☁️ 云存储集成**: 连接 Amazon S3 等云存储服务获取文档
- **📄 格式自动识别**: 根据文件扩展名自动选择合适的解析器
- **🔄 批量文档处理**: 支持大规模文档集合的高效加载和转换
- **🧩 组件协作**: 与 Parser、Transformer、Indexer 等组件无缝集成

---

## 🔧 核心接口

`DocumentLoader` 组件提供了简洁而强大的接口设计：

### 基础接口

```go
type Loader interface {
    Load(ctx context.Context, src Source, opts ...LoaderOption) ([]*schema.Document, error)
}
```

### 接口详解

#### 📥 Load 方法
- **功能**: 从指定数据源加载文档并转换为标准格式
- **输入**:
    - `ctx`: 上下文对象，用于控制超时、取消等
    - `src`: 数据源对象 (`Source` 接口)
    - `opts`: 可选配置参数 (`LoaderOption`)
- **输出**:
    - `[]*schema.Document`: 标准化的文档列表
    - `error`: 加载过程中的错误信息

---

## 📨 Document 和 Source 结构

### Document 结构体

`Document` 是文档的标准数据结构，与 Indexer 组件保持一致：

```go
type Document struct {
    // ID 是文档的唯一标识符
    ID string
    // Content 是文档的主要文本内容
    Content string
    // MetaData 存储文档的元数据信息
    MetaData map[string]interface{}
}
```

### Source 接口

`Source` 自定义了数据源的抽象接口：

```go
type Source interface {
    // 获取数据源的标识符
    GetID() string
    // 获取数据源的读取器
    GetReader(ctx context.Context) (io.Reader, error)
    // 获取数据源的元数据
    GetMetadata() map[string]interface{}
}
```

### 🎭 内置数据源类型

#### 1. 📁 FileSource - 本地文件源

```go
type FileSource struct {
    Path     string                 // 文件路径
    MetaData map[string]interface{} // 文件元数据
}

func NewFileSource(path string) *FileSource {
    return &FileSource{
        Path: path,
        MetaData: map[string]interface{}{
            "source_type": "file",
            "file_path":   path,
        },
    }
}
```

#### 2. 🌐 URLSource - 网络资源源

```go
type URLSource struct {
    URL      string                 // 资源URL
    Headers  map[string]string      // HTTP请求头
    MetaData map[string]interface{} // 资源元数据
}

func NewURLSource(url string) *URLSource {
    return &URLSource{
        URL: url,
        MetaData: map[string]interface{}{
            "source_type": "url",
            "url":         url,
        },
    }
}
```

#### 3. ☁️ S3Source - Amazon S3 云存储源

```go
type S3Source struct {
    Bucket   string                 // S3存储桶
    Key      string                 // 对象键
    Region   string                 // AWS区域
    MetaData map[string]interface{} // 对象元数据
}

func NewS3Source(bucket, key, region string) *S3Source {
    return &S3Source{
        Bucket: bucket,
        Key:    key,
        Region: region,
        MetaData: map[string]interface{}{
            "source_type": "s3",
            "bucket":      bucket,
            "key":         key,
            "region":      region,
        },
    }
}
```

---

## 🧩 DocumentParser 集成

DocumentLoader 与 DocumentParser 紧密集成，实现格式自动识别和智能解析：

### Parser 接口

```go
type Parser interface {
    Parse(ctx context.Context, reader io.Reader, opts ...ParseOption) ([]*schema.Document, error)
}
```

### 🛠️ 内置解析器类型

#### 1. 📝 TextParser - 纯文本解析器

```go
func NewTextParser() *TextParser {
    return &TextParser{}
}

func (p *TextParser) Parse(ctx context.Context, reader io.Reader, opts ...ParseOption) ([]*schema.Document, error) {
    content, err := io.ReadAll(reader)
    if err != nil {
        return nil, err
    }

    doc := &schema.Document{
        Content:  string(content),
        MetaData: map[string]interface{}{
            "parser_type": "text",
            "size":        len(content),
        },
    }

    return []*schema.Document{doc}, nil
}
```

#### 2. 🔧 ExtParser - 扩展名自动识别解析器

```go
func NewExtParser(ctx context.Context, config *ExtParserConfig) (*ExtParser, error) {
    return &ExtParser{
        parsers:         config.Parsers,
        fallbackParser:  config.FallbackParser,
        defaultMetadata: config.DefaultMetadata,
    }, nil
}

type ExtParserConfig struct {
    // 扩展名到解析器的映射
    Parsers map[string]Parser
    // 默认解析器（当扩展名不匹配时使用）
    FallbackParser Parser
    // 默认元数据
    DefaultMetadata map[string]interface{}
}
```

#### 3. 📄 专用格式解析器

```go
// HTML解析器示例
type HTMLParser struct {
    ExtractText bool // 是否只提取文本内容
    KeepTags    []string // 保留的HTML标签
}

// PDF解析器示例
type PDFParser struct {
    ExtractImages bool // 是否提取图片
    PageRange     []int // 处理的页面范围
}
```

---

## 🏗️ 创建和使用 DocumentLoader

### 基础使用流程

```go
import (
    "context"
    "fmt"
    "log"

    "github.com/cloudwego/eino/schema"
    "github.com/cloudwego/eino-ext/components/document/loader"
    "github.com/cloudwego/eino-ext/components/document/parser"
)

func basicLoaderExample() {
    ctx := context.Background()

    // 1️⃣ 创建解析器
    textParser := parser.NewTextParser()
    htmlParser := parser.NewHTMLParser(&parser.HTMLConfig{
        ExtractText: true,
    })

    // 2️⃣ 创建扩展解析器
    extParser, err := parser.NewExtParser(ctx, &parser.ExtParserConfig{
        Parsers: map[string]parser.Parser{
            ".txt":  textParser,
            ".md":   textParser,
            ".html": htmlParser,
            ".htm":  htmlParser,
        },
        FallbackParser: textParser,
    })
    if err != nil {
        log.Fatal("解析器创建失败:", err)
    }

    // 3️⃣ 创建文档加载器
    documentLoader, err := loader.NewDocumentLoader(ctx, &loader.Config{
        Parser: extParser,
    })
    if err != nil {
        log.Fatal("DocumentLoader 创建失败:", err)
    }

    // 4️⃣ 创建数据源
    fileSource := loader.NewFileSource("./documents/sample.txt")

    // 5️⃣ 加载文档
    documents, err := documentLoader.Load(ctx, fileSource)
    if err != nil {
        log.Fatal("文档加载失败:", err)
    }

    // 6️⃣ 处理加载的文档
    for _, doc := range documents {
        fmt.Printf("文档ID: %s\n", doc.ID)
        fmt.Printf("内容长度: %d 字符\n", len(doc.Content))
        fmt.Printf("元数据: %v\n", doc.MetaData)
    }
}
```

### 🎯 实用配置示例

#### 多数据源批量加载

```go
func batchLoadDocuments() ([]*schema.Document, error) {
    ctx := context.Background()

    // 创建加载器
    loader, err := createDocumentLoader(ctx)
    if err != nil {
        return nil, err
    }

    // 定义多种数据源
    sources := []loader.Source{
        loader.NewFileSource("./docs/guide.md"),
        loader.NewFileSource("./docs/api.pdf"),
        loader.NewURLSource("https://example.com/article.html"),
        loader.NewS3Source("my-bucket", "documents/report.docx", "us-west-2"),
    }

    var allDocuments []*schema.Document

    // 批量处理所有数据源
    for i, src := range sources {
        fmt.Printf("正在加载第 %d 个数据源...\n", i+1)

        docs, err := loader.Load(ctx, src)
        if err != nil {
            fmt.Printf("数据源 %s 加载失败: %v\n", src.GetID(), err)
            continue
        }

        allDocuments = append(allDocuments, docs...)
        fmt.Printf("成功加载 %d 个文档\n", len(docs))
    }

    return allDocuments, nil
}
```

#### 高级解析配置

```go
func createAdvancedLoader(ctx context.Context) (*loader.DocumentLoader, error) {
    // 1. 创建 HTML 解析器
    htmlParser := parser.NewHTMLParser(&parser.HTMLConfig{
        ExtractText: true,
        KeepTags:    []string{"h1", "h2", "p", "li"},
        RemoveAds:   true,
    })

    // 2. 创建 PDF 解析器
    pdfParser := parser.NewPDFParser(&parser.PDFConfig{
        ExtractImages: false,
        PageRange:     []int{1, -1}, // 处理所有页面
        OCR:          true,         // 启用 OCR
    })

    // 3. 创建 Markdown 解析器
    mdParser := parser.NewMarkdownParser(&parser.MarkdownConfig{
        ExtractMetadata: true,
        PreserveCode:    true,
    })

    // 4. 组装扩展解析器
    extParser, err := parser.NewExtParser(ctx, &parser.ExtParserConfig{
        Parsers: map[string]parser.Parser{
            ".html": htmlParser,
            ".htm":  htmlParser,
            ".pdf":  pdfParser,
            ".md":   mdParser,
            ".txt":  parser.NewTextParser(),
        },
        FallbackParser: parser.NewTextParser(),
        DefaultMetadata: map[string]interface{}{
            "processed_at": time.Now().Unix(),
            "loader_version": "1.0",
        },
    })
    if err != nil {
        return nil, err
    }

    // 5. 创建加载器
    return loader.NewDocumentLoader(ctx, &loader.Config{
        Parser: extParser,
        Options: &loader.Options{
            MaxFileSize:    50 * 1024 * 1024, // 50MB
            Timeout:        30 * time.Second,
            EnableCache:    true,
            CacheDir:       "./cache",
            RetryAttempts:  3,
        },
    })
}
```

---

## 🚀 编排集成最佳实践

虽然可以直接使用 DocumentLoader，但**官方强烈推荐**将其集成到编排工作流中，与其他组件协同工作。

### 🔗 Chain 编排模式

```go
import "github.com/cloudwego/eino/compose"

// 1️⃣ 文档加载 -> 转换 -> 索引 完整流水线
func createDocumentProcessingPipeline() (*compose.Runnable, error) {
    ctx := context.Background()

    // 初始化组件
    loader, err := createDocumentLoader(ctx)
    if err != nil {
        return nil, err
    }

    transformer, err := createDocumentTransformer(ctx)
    if err != nil {
        return nil, err
    }

    indexer, err := createDocumentIndexer(ctx)
    if err != nil {
        return nil, err
    }

    // 构建处理链
    // Source -> Documents -> Chunks -> Index IDs
    chain := compose.NewChain[loader.Source, []string]()
    chain.AppendLoader(loader)      // Source -> Documents
    chain.AppendTransformer(transformer) // Documents -> Chunks
    chain.AppendIndexer(indexer)    // Chunks -> IDs

    // 编译成可运行实例
    return chain.Compile(ctx)
}
```

### 🔄 Graph 编排模式

```go
func createDocumentProcessingGraph() (*compose.Runnable, error) {
    ctx := context.Background()

    // 创建 Graph
    graph := compose.NewGraph[loader.Source, map[string]interface{}]()

    // 添加节点
    graph.AddLoaderNode("load", documentLoader)
    graph.AddTransformerNode("transform", documentTransformer)
    graph.AddIndexerNode("index", documentIndexer)
    graph.AddRetrieverNode("search", documentRetriever)

    // 连接节点
    graph.AddEdge(compose.START, "load")
    graph.AddEdge("load", "transform")
    graph.AddEdge("transform", "index")
    graph.AddEdge("index", "search")
    graph.AddEdge("search", compose.END)

    // 编译
    return graph.Compile(ctx)
}
```

### 🏭 完整文档处理工作流示例

```go
type DocumentProcessingWorkflow struct {
    loader      *loader.DocumentLoader
    transformer *transformer.DocumentTransformer
    indexer     *indexer.DocumentIndexer
    retriever   *retriever.DocumentRetriever
}

func NewDocumentProcessingWorkflow(ctx context.Context) (*DocumentProcessingWorkflow, error) {
    // 初始化所有组件
    l, err := createDocumentLoader(ctx)
    if err != nil {
        return nil, fmt.Errorf("加载器创建失败: %w", err)
    }

    t, err := createDocumentTransformer(ctx)
    if err != nil {
        return nil, fmt.Errorf("转换器创建失败: %w", err)
    }

    i, err := createDocumentIndexer(ctx)
    if err != nil {
        return nil, fmt.Errorf("索引器创建失败: %w", err)
    }

    r, err := createDocumentRetriever(ctx)
    if err != nil {
        return nil, fmt.Errorf("检索器创建失败: %w", err)
    }

    return &DocumentProcessingWorkflow{
        loader:      l,
        transformer: t,
        indexer:     i,
        retriever:   r,
    }, nil
}

func (w *DocumentProcessingWorkflow) ProcessDocuments(ctx context.Context, sources []loader.Source) error {
    for _, src := range sources {
        // 1. 加载文档
        docs, err := w.loader.Load(ctx, src)
        if err != nil {
            return fmt.Errorf("文档加载失败 %s: %w", src.GetID(), err)
        }

        // 2. 转换文档
        chunks, err := w.transformer.Transform(ctx, docs)
        if err != nil {
            return fmt.Errorf("文档转换失败 %s: %w", src.GetID(), err)
        }

        // 3. 索引文档
        ids, err := w.indexer.Store(ctx, chunks)
        if err != nil {
            return fmt.Errorf("文档索引失败 %s: %w", src.GetID(), err)
        }

        fmt.Printf("成功处理文档 %s，生成 %d 个索引\n", src.GetID(), len(ids))
    }

    return nil
}
```

---

## ⚙️ 高级配置和选项

### Option 配置

DocumentLoader 支持通过 Option 在运行时传入额外配置：

```go
// WithTimeout - 设置加载超时时间
docs, err := loader.Load(ctx, source,
    loader.WithTimeout(30*time.Second),
)

// WithMetadata - 添加自定义元数据
docs, err := loader.Load(ctx, source,
    loader.WithMetadata(map[string]interface{}{
        "batch_id": "batch_001",
        "priority": "high",
    }),
)

// WithRetry - 设置重试策略
docs, err := loader.Load(ctx, source,
    loader.WithRetry(3, 2*time.Second),
)
```

### Callback 机制

回调机制允许在关键生命周期节点注入自定义逻辑：

```go
// 创建回调处理器
callbackHandler := callbacks.NewHandlerBuilder().
    OnStartFn(func(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
        src := input.(loader.Source)
        fmt.Printf("📥 开始加载文档: %s\n", src.GetID())
        return ctx
    }).
    OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) {
        docs := output.([]*schema.Document)
        fmt.Printf("✅ 成功加载 %d 个文档\n", len(docs))
    }).
    OnErrorFn(func(ctx context.Context, info *callbacks.RunInfo, err error) {
        fmt.Printf("❌ 文档加载失败: %v\n", err)
    }).
    Build()

// 在编排中使用回调
chain := compose.NewChain[loader.Source, []*schema.Document]()
chain.AppendLoader(documentLoader, compose.WithCallbacks(callbackHandler))
```

---

## 🎓 高级用法和技巧

### 1. 📊 自定义数据源实现

```go
type DatabaseSource struct {
    Query    string
    DB       *sql.DB
    MetaData map[string]interface{}
}

func (ds *DatabaseSource) GetID() string {
    return fmt.Sprintf("db_query_%x", md5.Sum([]byte(ds.Query)))
}

func (ds *DatabaseSource) GetReader(ctx context.Context) (io.Reader, error) {
    rows, err := ds.DB.QueryContext(ctx, ds.Query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var results []map[string]interface{}

    for rows.Next() {
        // 扫描行数据
        var title, content string
        if err := rows.Scan(&title, &content); err != nil {
            return nil, err
        }

        results = append(results, map[string]interface{}{
            "title":   title,
            "content": content,
        })
    }

    // 转换为JSON格式
    jsonData, err := json.Marshal(results)
    if err != nil {
        return nil, err
    }

    return strings.NewReader(string(jsonData)), nil
}

func (ds *DatabaseSource) GetMetadata() map[string]interface{} {
    return ds.MetaData
}
```

### 2. 🔄 并发加载优化

```go
type ConcurrentLoader struct {
    loader    *loader.DocumentLoader
    maxWorkers int
    semaphore chan struct{}
}

func NewConcurrentLoader(loader *loader.DocumentLoader, maxWorkers int) *ConcurrentLoader {
    return &ConcurrentLoader{
        loader:    loader,
        maxWorkers: maxWorkers,
        semaphore: make(chan struct{}, maxWorkers),
    }
}

func (cl *ConcurrentLoader) LoadConcurrently(ctx context.Context, sources []loader.Source) ([]*schema.Document, error) {
    var wg sync.WaitGroup
    var mu sync.Mutex
    var allDocs []*schema.Document
    var errors []error

    for _, src := range sources {
        wg.Add(1)

        go func(source loader.Source) {
            defer wg.Done()

            // 获取信号量
            cl.semaphore <- struct{}{}
            defer func() { <-cl.semaphore }()

            docs, err := cl.loader.Load(ctx, source)

            mu.Lock()
            if err != nil {
                errors = append(errors, fmt.Errorf("源 %s: %w", source.GetID(), err))
            } else {
                allDocs = append(allDocs, docs...)
            }
            mu.Unlock()
        }(src)
    }

    wg.Wait()

    if len(errors) > 0 {
        return allDocs, fmt.Errorf("部分加载失败: %v", errors)
    }

    return allDocs, nil
}
```

### 3. 📈 加载性能监控

```go
type LoaderMetrics struct {
    TotalSources     int64
    SuccessfulLoads  int64
    FailedLoads      int64
    TotalDocuments   int64
    AverageLoadTime  time.Duration
    LastLoadTime     time.Time
    mu               sync.RWMutex
}

func (m *LoaderMetrics) RecordLoad(docCount int, duration time.Duration, success bool) {
    m.mu.Lock()
    defer m.mu.Unlock()

    m.TotalSources++
    m.LastLoadTime = time.Now()

    if success {
        m.SuccessfulLoads++
        m.TotalDocuments += int64(docCount)
    } else {
        m.FailedLoads++
    }

    // 更新平均加载时间
    m.AverageLoadTime = (m.AverageLoadTime + duration) / 2
}

func (m *LoaderMetrics) GetStats() map[string]interface{} {
    m.mu.RLock()
    defer m.mu.RUnlock()

    return map[string]interface{}{
        "total_sources":     m.TotalSources,
        "successful_loads":  m.SuccessfulLoads,
        "failed_loads":      m.FailedLoads,
        "total_documents":   m.TotalDocuments,
        "success_rate":      float64(m.SuccessfulLoads) / float64(m.TotalSources),
        "average_load_time": m.AverageLoadTime.String(),
        "last_load_time":    m.LastLoadTime.Format(time.RFC3339),
    }
}
```

### 4. 🧹 智能缓存机制

```go
type CachedLoader struct {
    loader *loader.DocumentLoader
    cache  map[string]*CacheEntry
    ttl    time.Duration
    mu     sync.RWMutex
}

type CacheEntry struct {
    Documents []*schema.Document
    Timestamp time.Time
    Hash      string
}

func (cl *CachedLoader) Load(ctx context.Context, src loader.Source) ([]*schema.Document, error) {
    cacheKey := cl.generateCacheKey(src)

    // 检查缓存
    if docs := cl.getFromCache(cacheKey); docs != nil {
        fmt.Printf("缓存命中: %s\n", src.GetID())
        return docs, nil
    }

    // 缓存未命中，执行真实加载
    docs, err := cl.loader.Load(ctx, src)
    if err != nil {
        return nil, err
    }

    // 存入缓存
    cl.putToCache(cacheKey, docs)

    return docs, nil
}

func (cl *CachedLoader) generateCacheKey(src loader.Source) string {
    metadata := src.GetMetadata()
    keyData := fmt.Sprintf("%s_%v", src.GetID(), metadata)
    hash := md5.Sum([]byte(keyData))
    return fmt.Sprintf("%x", hash)
}
```

---

## ❓ 常见问题和解决方案

### Q1: 大文件加载内存溢出

**问题**: 加载大文件时出现内存不足错误
```
fatal error: runtime: out of memory
```

**解决方案**:
```go
// ✅ 流式处理大文件
type StreamingFileLoader struct {
    chunkSize int
    maxSize   int64
}

func (sfl *StreamingFileLoader) LoadLargeFile(ctx context.Context, filePath string) ([]*schema.Document, error) {
    file, err := os.Open(filePath)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    // 检查文件大小
    stat, err := file.Stat()
    if err != nil {
        return nil, err
    }

    if stat.Size() > sfl.maxSize {
        return nil, fmt.Errorf("文件太大: %d 字节，超过限制 %d", stat.Size(), sfl.maxSize)
    }

    var documents []*schema.Document
    buffer := make([]byte, sfl.chunkSize)
    chunkIndex := 0

    for {
        n, err := file.Read(buffer)
        if err != nil && err != io.EOF {
            return nil, err
        }

        if n > 0 {
            doc := &schema.Document{
                ID:      fmt.Sprintf("%s_chunk_%d", filePath, chunkIndex),
                Content: string(buffer[:n]),
                MetaData: map[string]interface{}{
                    "source_file": filePath,
                    "chunk_index": chunkIndex,
                    "file_size":   stat.Size(),
                },
            }
            documents = append(documents, doc)
            chunkIndex++
        }

        if err == io.EOF {
            break
        }
    }

    return documents, nil
}
```

### Q2: 网络资源加载超时

**问题**: 从网络加载文档时频繁超时
```go
// ❌ 没有合适的超时和重试机制
docs, err := loader.Load(ctx, urlSource)  // 可能长时间阻塞
```

**解决方案**:
```go
// ✅ 带有重试和超时的网络加载
type RobustURLLoader struct {
    client       *http.Client
    maxRetries   int
    retryDelay   time.Duration
}

func NewRobustURLLoader() *RobustURLLoader {
    return &RobustURLLoader{
        client: &http.Client{
            Timeout: 30 * time.Second,
            Transport: &http.Transport{
                MaxIdleConns:        100,
                MaxIdleConnsPerHost: 10,
                IdleConnTimeout:     90 * time.Second,
            },
        },
        maxRetries: 3,
        retryDelay: 2 * time.Second,
    }
}

func (rul *RobustURLLoader) LoadWithRetry(ctx context.Context, url string) ([]*schema.Document, error) {
    var lastErr error

    for attempt := 0; attempt < rul.maxRetries; attempt++ {
        if attempt > 0 {
            time.Sleep(rul.retryDelay * time.Duration(attempt))
        }

        docs, err := rul.loadURL(ctx, url)
        if err == nil {
            return docs, nil
        }

        lastErr = err
        fmt.Printf("尝试 %d 失败: %v\n", attempt+1, err)
    }

    return nil, fmt.Errorf("所有重试都失败了，最后错误: %w", lastErr)
}
```

### Q3: 解析器格式识别错误

**问题**: 文件格式识别不准确，导致解析失败
```go
// ❌ 仅依赖文件扩展名
parser, exists := parsers[filepath.Ext(filename)]
if !exists {
    return nil, fmt.Errorf("不支持的文件格式")
}
```

**解决方案**:
```go
// ✅ 结合文件头和扩展名的智能识别
type SmartParser struct {
    parsers map[string]parser.Parser
    detector *filetype.Detector
}

func (sp *SmartParser) detectFileType(reader io.Reader) (string, error) {
    // 读取文件头用于格式检测
    buffer := make([]byte, 512)
    n, err := reader.Read(buffer)
    if err != nil && err != io.EOF {
        return "", err
    }

    // 使用 magic number 检测
    kind, _ := filetype.Match(buffer[:n])
    if kind != filetype.Unknown {
        return kind.Extension, nil
    }

    // 回退到内容分析
    content := string(buffer[:n])
    if strings.Contains(content, "<!DOCTYPE html") || strings.Contains(content, "<html") {
        return ".html", nil
    }
    if strings.HasPrefix(content, "# ") || strings.Contains(content, "## ") {
        return ".md", nil
    }

    return ".txt", nil // 默认文本
}
```

### Q4: 云存储访问权限问题

**问题**: S3 访问权限配置不当导致加载失败
```go
// ❌ 权限不足或配置错误
AccessDenied: Access Denied
```

**解决方案**:
```go
// ✅ 完善的 S3 权限和配置管理
type SecureS3Loader struct {
    client   *s3.Client
    config   *S3Config
}

type S3Config struct {
    Region          string
    AccessKeyID     string
    SecretAccessKey string
    SessionToken    string
    Bucket          string
    Prefix          string
}

func NewSecureS3Loader(config *S3Config) (*SecureS3Loader, error) {
    cfg, err := awsconfig.LoadDefaultConfig(context.TODO(),
        awsconfig.WithRegion(config.Region),
        awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
            config.AccessKeyID,
            config.SecretAccessKey,
            config.SessionToken,
        )),
    )
    if err != nil {
        return nil, err
    }

    client := s3.NewFromConfig(cfg)

    return &SecureS3Loader{
        client: client,
        config: config,
    }, nil
}

func (ssl *SecureS3Loader) LoadObject(ctx context.Context, key string) (*schema.Document, error) {
    // 检查对象是否存在
    _, err := ssl.client.HeadObject(ctx, &s3.HeadObjectInput{
        Bucket: &ssl.config.Bucket,
        Key:    &key,
    })
    if err != nil {
        return nil, fmt.Errorf("对象不存在或无访问权限: %w", err)
    }

    // 获取对象
    result, err := ssl.client.GetObject(ctx, &s3.GetObjectInput{
        Bucket: &ssl.config.Bucket,
        Key:    &key,
    })
    if err != nil {
        return nil, fmt.Errorf("获取对象失败: %w", err)
    }
    defer result.Body.Close()

    content, err := io.ReadAll(result.Body)
    if err != nil {
        return nil, fmt.Errorf("读取对象内容失败: %w", err)
    }

    return &schema.Document{
        ID:      key,
        Content: string(content),
        MetaData: map[string]interface{}{
            "source_type":   "s3",
            "bucket":        ssl.config.Bucket,
            "key":          key,
            "size":         len(content),
            "last_modified": result.LastModified,
            "content_type":  *result.ContentType,
        },
    }, nil
}
```

---

## 🎉 总结

DocumentLoader 是 Eino 框架中的**核心数据获取组件**，掌握它的使用对于构建高质量的文档处理 AI 应用至关重要：

### 🏆 核心优势
- 📂 **统一接口**: 支持多种数据源的统一访问方式
- ⚡ **高性能**: 支持并发加载和智能缓存机制
- 🧩 **可扩展**: 丰富的解析器生态系统和自定义扩展能力
- 🔧 **灵活配置**: 支持多种配置选项和运行时参数调整
- 🛡️ **可靠性**: 完善的错误处理、重试机制和权限管理
- 🌐 **多源支持**: 本地文件、网络资源、云存储的无缝集成

### 💡 最佳实践总结
1. **合理选择数据源**: 根据实际需求选择最适合的数据源类型
2. **智能解析配置**: 使用 ExtParser 实现自动格式识别和解析
3. **性能优化**: 合理配置并发数量、缓存策略和批处理大小
4. **错误处理**: 实施完善的重试机制、超时控制和异常处理
5. **资源管理**: 正确管理文件句柄、网络连接和内存使用
6. **编排集成**: 优先使用 Chain/Graph 编排构建完整的文档处理工作流

### 🔗 相关资源
- 📚 [官方文档](https://www.cloudwego.io/zh/docs/eino/core_modules/components/document_loader_guide/)
- 🌐 [GitHub 仓库](https://github.com/cloudwego/eino)
- 📄 [DocumentParser 接口指南](https://www.cloudwego.io/zh/docs/eino/core_modules/components/document_loader_guide/document_parser_interface_guide/)

通过掌握 DocumentLoader 组件的各种功能和最佳实践，你将能够构建出更加智能、高效和可扩展的文档获取和处理系统！🚀