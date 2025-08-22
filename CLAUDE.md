# Eino Project - Claude Memory

## 项目概述
这是一个名为 Eino 的 AI 项目，位于 `/Users/yuyansong/AiProject/Eino`。

**Eino 是什么？**
- 是一个为简化和加速大模型应用构建而设计的云原生开发框架
- 基于 Go 语言开发
- 支持 RAG（检索增强生成）、Tool 系统、向量数据库集成等功能
- 使用火山方舟（Ark）作为大语言模型服务

## 技术架构

### 核心组件
1. **Transformer** - 智能文档分割，支持 Markdown 格式
2. **Indexer** - 文档向量化与存储到 Milvus
3. **Retriever** - 基于语义相似度的知识检索
4. **Tools** - 多种实用工具集成
5. **RAG** - 检索增强生成
6. **Chain** - 端到端工作流编排
7. **Lambda** - 自定义函数逻辑嵌入组件

### 技术栈
- **语言**: Go 1.24.2
- **大模型**: 火山方舟（Ark）API
- **向量数据库**: Milvus
- **主要依赖**: 
  - `github.com/cloudwego/eino v0.4.4`
  - `github.com/milvus-io/milvus-sdk-go/v2 v2.4.2`
  - `github.com/spf13/viper v1.20.1`

## 项目结构

### 配置文件
- `config.yaml` - 包含 API Key 和 Milvus 配置
- `docker-compose.yml` - 容器编排配置
- `go.mod` - Go 模块依赖

### 核心演示目录
- `comprehensive_demo/` - 完整的 RAG + Tool 系统综合演示
- `lambda_demo/` - Lambda 组件各种用法演示
- `tool_demo/` - 工具系统演示
- `retriever_demo/` - 检索器演示
- `embedding_demo/` - 嵌入模型演示
- `indexer_demo/` - 索引器演示
- `transformer_demo/` - 文档转换器演示
- `chattemplate_demo/` - 对话模板演示

### 文档系统
- `AI_Agent_Concepts.md` - AI Agent 核心概念
- `LLM_and_RAG_Concepts.md` - LLM 和 RAG 概念
- `MCP_Concepts.md` - MCP 概念
- `Vector_Database_Concepts.md` - 向量数据库概念
- `Eino_Components_Relationship.md` - 组件关系详解（智能图书馆比喻）
- `Eino_Orchestration_Guide.md` - 编排系统详解（流水线比喻）
- 各演示目录包含对应的 README.md 和说明文档

### 主要代码文件
- `main.go` - 项目主入口，演示 RAG 编排流程
- `examples/` - 示例代码
- `use_methods/` - 使用方法示例

## 最近活动记录
- 2025-08-22: 用户询问 Claude 如何获取历史对话记录
- 2025-08-22: 创建了此 CLAUDE.md 文件用于记录项目信息和会话历史
- 2025-08-22: 完成项目整体结构分析，了解了 Eino 框架的完整生态
- 2025-08-22: 创建了形象生动的组件关系说明文档 `Eino_Components_Relationship.md`
  - 使用"智能图书馆"比喻解释各组件作用
  - 包含完整的组件协作流程图
  - 提供实际应用场景示例
- 2025-08-22: 学习了 Eino 官方编排设计原则
  - 类型对齐原则、设计理念、双引擎支持等
- 2025-08-22: 创建了编排概念详解文档 `Eino_Orchestration_Guide.md`
  - 以汽车流水线比喻编排概念
  - 详细对比 Chain vs Graph 两种编排方式
  - 包含丰富的 ASCII 图解和实际代码示例
  - 涵盖最佳实践和性能优化建议

## 开发偏好
- 平台：macOS (Darwin 24.6.0)
- 工作目录：/Users/yuyansong/AiProject/Eino

## 配置信息
- ARK API Key: 已配置在 config.yaml
- Milvus 地址: localhost:19530
- 模型: doubao-seed-1-6-250615
- 嵌入模型: doubao-embedding-text-240715

## 重要提醒
- 每次 Claude Code 会话都是独立的，不保存历史记录
- 使用此文件记录重要的项目信息和开发进度
- 项目包含完整的 AI 应用开发框架，支持从文档处理到智能问答的全流程