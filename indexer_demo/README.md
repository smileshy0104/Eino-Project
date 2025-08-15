# Eino Indexer (Milvus) 组件使用示例说明

本文档旨在解释 `indexer_demo/main.go` 文件中的代码，帮助您理解如何使用 Eino 的 `Indexer` 组件与 **Milvus** 向量数据库进行交互。

---

## 1. 核心目标

本示例的核心目标是演示 `Indexer` 组件如何作为一个**桥梁**，将 Eino 框架中的标准数据结构 (`schema.Document`) 无缝地存储到 Milvus 向量数据库中。

具体来说，示例将展示：
1.  如何定义一个 Milvus Collection 的**模式 (Schema)**。
2.  如何配置和初始化一个连接到 Milvus 的 `Indexer`。
3.  如何将**客户端的 Embedding 组件**与 `Indexer` 结合，实现**客户端向量化**。
4.  如何构建 `schema.Document` 对象，并为其附加将在 Milvus 中存储的元数据。
5.  如何调用 `Store` 方法将这些文档及其向量持久化到 Milvus。

---

## 2. 代码结构解析

`main.go` 文件主要由三部分组成：

### 2.1. Milvus Schema 定义 (`fields` 变量)

代码首先定义了一个全局变量 `fields`，它描述了我们希望在 Milvus 中创建的 Collection 的结构。这包括：
- **`id`**: 主键，用于唯一标识每个文档。
- **`vector`**: 向量字段，用于存储文本内容的向量表示，是进行相似度搜索的基础。
- **`content`**: 原始的文本内容，方便直接查看。
- **`metadata`**: JSON 字段，用于存储任意结构化的元数据，如来源、作者等，可用于后续的元数据过滤。

### 2.2. `runIndexerExample` 函数

这是示例的核心逻辑，它按照以下步骤执行：

1.  **初始化 `Embedder`**:
    - 在与 Milvus 交互之前，我们需要先将文本内容转换为向量。
    - 代码首先创建了一个 `ark.NewEmbedder` 实例。这意味着**向量化是在客户端完成的**，`Indexer` 将接收已经包含向量的文档。

2.  **加载配置与初始化 `Indexer`**:
    - 代码通过 `viper` 从配置文件 (`config.yaml`) 或环境变量中加载 Milvus 的连接地址 (`MILVUS_ADDRESS`) 和要操作的集合名称 (`MILVUS_COLLECTION`)。
    - 创建一个 Milvus Go SDK 的原生客户端 (`cli.NewClient`)。
    - 创建一个 `milvus.IndexerConfig`，这是关键的配置步骤，它将：
        - Milvus 客户端
        - Collection 名称
        - **上一步创建的 `Embedder` 实例**
        - Collection 的 `fields` 定义
      ...全部关联起来。
    - 调用 `milvus.NewIndexer` 创建 `Indexer` 实例。如果 Collection 不存在，`NewIndexer` 会根据提供的 `fields` 定义**自动创建**它。

3.  **准备文档 (`schema.Document`)**:
    - 创建了两个 `schema.Document` 对象。每个对象都包含：
        - `ID`: 文档的唯一标识符，将映射到 Milvus 中的 `id` 字段。
        - `Content`: 文档的主要文本内容。
        - `MetaData`: 一个 `map[string]interface{}`，将映射到 Milvus 中的 `metadata` (JSON) 字段。

4.  **调用 `Store` 方法**:
    - 将包含两个文档的列表传递给 `indexer.Store` 方法。
    - **内部工作流程**：
        1.  `Indexer` 遍历文档列表。
        2.  调用配置中指定的 `Embedder`，将每个文档的 `Content` 转换为向量。
        3.  将文档的 `ID`, `Content`, `MetaData` 以及新生成的**向量**打包成符合 Milvus 格式的数据。
        4.  通过 Milvus 客户端将这些数据批量插入到指定的 Collection 中。

5.  **打印结果**:
    - `Store` 方法成功后会返回一个包含已存储文档 ID 的列表。
    - 程序打印这个列表，以确认操作成功。

### 2.3. `main` 函数

`main` 函数是程序的入口，它的职责是：
1.  初始化 `viper`，使其能够从 `config.yaml` 和环境变量中读取所有配置（包括 Ark Embedding 和 Milvus 的）。
2.  调用 `runIndexerExample()` 来执行示例的核心逻辑。

---

## 3. 如何运行

1.  **准备 Milvus**:
    - 确保您有一个正在运行的 Milvus 实例。
    - 您**不需要**手动创建 Collection，代码会自动创建。

2.  **设置配置**:
    - **方式一 (推荐)**: 在项目根目录下创建一个 `config.yaml` 文件，并填入以下内容：
      ```yaml
      # Ark Embedding Service
      ARK_API_KEY: "YOUR_ARK_API_KEY"
      EMBEDDER_MODEL: "bge-large-zh" # 或其他模型

      # Milvus Service
      MILVUS_ADDRESS: "localhost:19530" # 您的 Milvus 实例地址
      MILVUS_COLLECTION: "eino_milvus_demo"
      ```
    - **方式二**: 将上述 `KEY` 设置为环境变量。

3.  **运行代码**:
    - 在终端中，进入 `indexer_demo` 目录。
    - 执行 `go run .` 或 `go run main.go`。

您将在终端看到文档被成功存储并返回 ID 的确认信息。