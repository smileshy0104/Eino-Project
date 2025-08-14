# Eino Indexer 组件使用示例说明

本文档旨在解释 `indexer_demo/main.go` 文件中的代码，帮助您理解如何使用 Eino 的 `Indexer` 组件与火山引擎 VikingDB 进行交互。

---

## 1. 核心目标

本示例的核心目标是演示 `Indexer` 组件最核心的功能：**将结构化的文档数据存储到向量数据库中**。

具体来说，示例将展示：
1.  如何配置和初始化一个连接到火山引擎 VikingDB 的 `Indexer`。
2.  如何构建 `schema.Document` 对象，并为其附加自定义的元数据。
3.  如何调用 `Store` 方法将这些文档持久化。

---

## 2. 代码结构解析

`main.go` 文件主要由两部分组成：

### 2.1. `runIndexerExample` 函数

这是示例的核心逻辑，它按照以下步骤执行：

1.  **加载配置**:
    - 代码通过 `viper` 从配置文件 (`config.yaml`) 或环境变量中安全地加载连接 VikingDB 所需的敏感信息，包括：
        - `VIKINGDB_AK` / `VIKINGDB_SK`: 访问密钥。
        - `VIKINGDB_HOST`: VikingDB 实例的地址。
        - `VIKINGDB_REGION`: 实例所在的区域。
        - `VIKINGDB_COLLECTION`: 要操作的数据集（Collection）名称。
    - 这种方式避免了将敏感信息硬编码在代码中。

2.  **初始化 `Indexer`**:
    - 创建一个 `volc_vikingdb.IndexerConfig` 结构体，填入上一步获取的配置。
    - **关键配置**：`EmbeddingConfig.UseBuiltin: true`。这告诉 `Indexer` 组件，我们不需要在客户端（代码中）进行文本到向量的转换，而是希望 VikingDB 服务端利用其**内置的 Embedding 模型**来完成这个过程。这可以简化客户端的逻辑并可能提高效率。
    - 调用 `volc_vikingdb.NewIndexer` 创建 `Indexer` 实例。

3.  **准备文档 (`schema.Document`)**:
    - 创建了两个 `schema.Document` 对象。每个对象都包含：
        - `ID`: 文档的唯一标识符。
        - `Content`: 文档的主要文本内容，这部分内容将被 VikingDB 内置的模型转换为向量。
    - **附加元数据**: 使用 `volc_vikingdb.SetExtraDataFields` 辅助函数为每个文档添加了自定义的元数据字段（如 `source`, `author`）。这些字段也必须在 VikingDB 的 Collection 结构中预先定义好，它们可以用于后续的过滤查询。

4.  **调用 `Store` 方法**:
    - 将包含两个文档的列表传递给 `indexer.Store` 方法。
    - 组件内部会处理与 VikingDB 的所有网络通信、认证和数据格式化，最终将文档及其生成的向量存入指定的 Collection。

5.  **打印结果**:
    - `Store` 方法成功后会返回一个包含已存储文档 ID 的列表。
    - 程序打印这个列表，以确认操作成功。

### 2.2. `main` 函数

`main` 函数是程序的入口，它的职责很简单：
1.  初始化 `viper`，使其能够从 `config.yaml` 和环境变量中读取配置。
2.  调用 `runIndexerExample()` 来执行示例的核心逻辑。

---

## 3. 如何运行

1.  **准备 VikingDB**:
    - 确保您已经创建了一个火山引擎 VikingDB 实例和一个 Collection。
    - 在 Collection 中，除了 `ID` 和 `vector` 字段外，还需根据示例代码创建 `source` (string) 和 `author` (string) 这两个元数据字段。

2.  **设置配置**:
    - **方式一 (推荐)**: 在项目根目录下创建一个 `config.yaml` 文件，并填入以下内容：
      ```yaml
      VIKINGDB_AK: "YOUR_ACCESS_KEY"
      VIKINGDB_SK: "YOUR_SECRET_KEY"
      VIKINGDB_HOST: "your-vikingdb-instance.volces.com"
      VIKINGDB_REGION: "cn-beijing"
      VIKINGDB_COLLECTION: "your_collection_name"
      ```
    - **方式二**: 将上述 `KEY` 设置为环境变量。

3.  **运行代码**:
    - 在终端中，进入 `indexer_demo` 目录。
    - 执行 `go run .` 或 `go run main.go`。

您将在终端看到文档被成功存储并返回 ID 的确认信息。