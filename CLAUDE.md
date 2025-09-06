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
- `Eino_Open_Source_Overview.md` - 开源框架全景解析（搭积木比喻）
- `ByteDance_Eino_Practice.md` - 字节跳动实践案例（足球队比喻）
- `AI_Programming_Assistant_Guide.md` - AI 辅助编程新手完全指南
- 各演示目录包含对应的 README.md 和说明文档

### 主要代码文件
- `main.go` - 项目主入口，演示 RAG 编排流程
- `examples/` - 示例代码
- `use_methods/` - 使用方法示例

## 最近活动记录
### 2025-08-22 基础文档创建阶段
- 用户询问 Claude 如何获取历史对话记录
- 创建了此 CLAUDE.md 文件用于记录项目信息和会话历史
- 完成项目整体结构分析，了解了 Eino 框架的完整生态
- 创建了形象生动的组件关系说明文档 `Eino_Components_Relationship.md`
- 创建了编排概念详解文档 `Eino_Orchestration_Guide.md`
- 创建了 Eino 开源框架全景解析文档 `Eino_Open_Source_Overview.md`
- 创建了字节跳动 Eino 实践案例解析文档 `ByteDance_Eino_Practice.md`
- 创建了 AI 辅助编程完全新手指南 `AI_Programming_Assistant_Guide.md`

### 2025-09-06 ADK 演示开发阶段
- 阅读并分析了 Eino ADK 概览文档（https://www.cloudwego.io/zh/docs/eino/core_modules/eino_adk/outline/）
- 创建了通俗易懂的 ADK 指南文档 `Eino_ADK_Guide.md`
- 开发了基于概念的多智能体演示代码
- 基于真实 GitHub 仓库（https://github.com/cloudwego/eino）修复了所有演示代码
- 创建了符合官方标准的工具演示 `corrected_official_demo.go`

### 2025-09-06 Agent 扩展机制开发阶段  
- 深入研究了 Agent 扩展文档（https://www.cloudwego.io/zh/docs/eino/core_modules/eino_adk/agent_extension/）
- 创建了 Agent 扩展机制完整指南 `Eino_ADK_Agent_Extension_Guide.md`
- 开发了完整的中断与恢复功能演示 `agent_extension_demo.go`
- 创建了稳定可靠的扩展机制演示 `stable_extension_demo.go`
- 完善了项目文档，包含两个主要演示方向：工具接口 + 扩展机制

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
- 我希望每次回答Eino相关问题时，都要根据https://github.com/cloudwego/eino该github进行