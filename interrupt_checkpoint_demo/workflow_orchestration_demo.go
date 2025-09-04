package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// ==================== 基础数据结构 ====================

// 工作流输入数据
type WorkflowInput struct {
	Text    string                 `json:"text"`
	UserID  string                 `json:"user_id"`
	Options map[string]interface{} `json:"options"`
}

// 工作流输出数据
type WorkflowOutput struct {
	Result   string                 `json:"result"`
	Metadata map[string]interface{} `json:"metadata"`
	Status   string                 `json:"status"`
}

// 节点处理结果
type NodeResult struct {
	Data      interface{}            `json:"data"`
	Metadata  map[string]interface{} `json:"metadata"`
	NodeName  string                 `json:"node_name"`
	Timestamp string                 `json:"timestamp"`
}

// ==================== Lambda 节点函数类型 ====================

type LambdaFunc func(ctx context.Context, input interface{}) (NodeResult, error)

// ==================== 工作流框架核心 ====================

// 工作流节点
type WorkflowNode struct {
	Name     string
	Function LambdaFunc
	Inputs   map[string]string // 字段映射配置
	Static   map[string]interface{}
}

// 工作流引擎
type WorkflowEngine struct {
	nodes       map[string]*WorkflowNode
	edges       map[string][]string // 依赖关系
	staticVals  map[string]interface{}
	nodeResults map[string]NodeResult
}

func NewWorkflowEngine() *WorkflowEngine {
	return &WorkflowEngine{
		nodes:       make(map[string]*WorkflowNode),
		edges:       make(map[string][]string),
		staticVals:  make(map[string]interface{}),
		nodeResults: make(map[string]NodeResult),
	}
}

// 添加Lambda节点
func (we *WorkflowEngine) AddLambdaNode(name string, fn LambdaFunc) *WorkflowEngine {
	we.nodes[name] = &WorkflowNode{
		Name:     name,
		Function: fn,
		Inputs:   make(map[string]string),
		Static:   make(map[string]interface{}),
	}
	return we
}

// 配置节点输入映射
func (we *WorkflowEngine) AddInput(nodeName string, inputMappings map[string]string) *WorkflowEngine {
	if node, exists := we.nodes[nodeName]; exists {
		for key, value := range inputMappings {
			node.Inputs[key] = value
		}

		// 构建依赖关系
		for _, mapping := range inputMappings {
			if strings.Contains(mapping, ".") {
				sourceNode := strings.Split(mapping, ".")[0]
				if sourceNode != "static" && sourceNode != "input" {
					we.addEdge(sourceNode, nodeName)
				}
			}
		}
	}
	return we
}

// 设置静态值
func (we *WorkflowEngine) SetStaticValue(key string, value interface{}) *WorkflowEngine {
	we.staticVals[key] = value
	return we
}

// 添加边（依赖关系）
func (we *WorkflowEngine) addEdge(sourceNode, nodeName string) {
	if we.edges[sourceNode] == nil {
		we.edges[sourceNode] = []string{}
	}

	// 避免重复添加
	for _, existing := range we.edges[sourceNode] {
		if existing == nodeName {
			return
		}
	}

	we.edges[sourceNode] = append(we.edges[sourceNode], nodeName)
}

// 执行工作流
func (we *WorkflowEngine) Invoke(ctx context.Context, input WorkflowInput) (WorkflowOutput, error) {
	fmt.Println("🚀 启动工作流执行...")
	we.nodeResults = make(map[string]NodeResult) // 重置结果

	// 拓扑排序获取执行顺序
	executionOrder, err := we.topologicalSort()
	if err != nil {
		return WorkflowOutput{}, fmt.Errorf("工作流拓扑排序失败: %v", err)
	}

	fmt.Printf("📋 执行顺序: %v\n", executionOrder)

	// 按顺序执行节点
	for _, nodeName := range executionOrder {
		node := we.nodes[nodeName]
		if node == nil {
			continue
		}

		fmt.Printf("🔄 执行节点: %s\n", nodeName)

		// 准备节点输入
		nodeInput, err := we.prepareNodeInput(nodeName, input)
		if err != nil {
			return WorkflowOutput{}, fmt.Errorf("准备节点 %s 输入失败: %v", nodeName, err)
		}

		// 执行节点
		result, err := node.Function(ctx, nodeInput)
		if err != nil {
			return WorkflowOutput{}, fmt.Errorf("节点 %s 执行失败: %v", nodeName, err)
		}

		result.NodeName = nodeName
		result.Timestamp = time.Now().Format(time.RFC3339)
		we.nodeResults[nodeName] = result

		fmt.Printf("✅ 节点 %s 执行完成\n", nodeName)
	}

	// 构建最终输出
	finalResult := we.buildFinalOutput()
	fmt.Println("🎉 工作流执行完成！")

	return finalResult, nil
}

// 拓扑排序
func (we *WorkflowEngine) topologicalSort() ([]string, error) {
	inDegree := make(map[string]int)
	allNodes := make(map[string]bool)

	// 计算所有节点的入度
	for nodeName := range we.nodes {
		allNodes[nodeName] = true
		inDegree[nodeName] = 0
	}

	for _, tos := range we.edges {
		for _, to := range tos {
			inDegree[to]++
		}
	}

	// 找到所有入度为0的节点
	queue := []string{}
	for nodeName, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, nodeName)
		}
	}

	result := []string{}

	// 拓扑排序
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)

		// 更新后继节点的入度
		for _, next := range we.edges[current] {
			inDegree[next]--
			if inDegree[next] == 0 {
				queue = append(queue, next)
			}
		}
	}

	// 检查是否有环
	if len(result) != len(we.nodes) {
		return nil, fmt.Errorf("工作流存在循环依赖")
	}

	return result, nil
}

// 准备节点输入
func (we *WorkflowEngine) prepareNodeInput(nodeName string, workflowInput WorkflowInput) (interface{}, error) {
	node := we.nodes[nodeName]
	if node == nil {
		return nil, fmt.Errorf("节点 %s 不存在", nodeName)
	}

	// 构建节点输入数据
	nodeInput := make(map[string]interface{})

	// 处理字段映射
	for fieldName, mapping := range node.Inputs {
		value, err := we.resolveFieldMapping(mapping, workflowInput)
		if err != nil {
			return nil, fmt.Errorf("解析字段映射 %s -> %s 失败: %v", fieldName, mapping, err)
		}
		nodeInput[fieldName] = value
	}

	// 处理静态值
	for fieldName, value := range node.Static {
		nodeInput[fieldName] = value
	}

	return nodeInput, nil
}

// 解析字段映射
func (we *WorkflowEngine) resolveFieldMapping(mapping string, workflowInput WorkflowInput) (interface{}, error) {
	// 处理静态值
	if strings.HasPrefix(mapping, "static.") {
		staticKey := strings.TrimPrefix(mapping, "static.")
		if value, exists := we.staticVals[staticKey]; exists {
			return value, nil
		}
		return nil, fmt.Errorf("静态值 %s 不存在", staticKey)
	}

	// 处理输入字段
	if strings.HasPrefix(mapping, "input.") {
		fieldPath := strings.TrimPrefix(mapping, "input.")
		return we.getValueFromPath(workflowInput, fieldPath)
	}

	// 处理节点输出字段
	if strings.Contains(mapping, ".") {
		parts := strings.SplitN(mapping, ".", 2)
		nodeName := parts[0]
		fieldPath := parts[1]

		if result, exists := we.nodeResults[nodeName]; exists {
			return we.getValueFromPath(result, fieldPath)
		}
		return nil, fmt.Errorf("节点 %s 的结果不存在", nodeName)
	}

	// 直接返回映射值（可能是字符串字面量）
	return mapping, nil
}

// 从路径获取值
func (we *WorkflowEngine) getValueFromPath(data interface{}, path string) (interface{}, error) {
	// 简化实现：支持基本的字段访问
	pathParts := strings.Split(path, ".")
	current := data

	for _, part := range pathParts {
		switch v := current.(type) {
		case map[string]interface{}:
			if val, exists := v[part]; exists {
				current = val
			} else {
				return nil, fmt.Errorf("字段 %s 不存在", part)
			}
		case NodeResult:
			// 访问NodeResult的字段
			switch part {
			case "data":
				current = v.Data
			case "metadata":
				current = v.Metadata
			case "node_name":
				current = v.NodeName
			case "timestamp":
				current = v.Timestamp
			default:
				// 尝试访问Data中的字段
				if dataMap, ok := v.Data.(map[string]interface{}); ok {
					if val, exists := dataMap[part]; exists {
						current = val
					} else {
						return nil, fmt.Errorf("NodeResult.Data中字段 %s 不存在", part)
					}
				} else {
					return nil, fmt.Errorf("无法访问NodeResult中的字段 %s", part)
				}
			}
		case WorkflowInput:
			// 访问WorkflowInput的字段
			switch part {
			case "text":
				current = v.Text
			case "user_id":
				current = v.UserID
			case "options":
				current = v.Options
			default:
				return nil, fmt.Errorf("WorkflowInput中字段 %s 不存在", part)
			}
		default:
			return nil, fmt.Errorf("无法访问字段 %s，类型不支持", part)
		}
	}

	return current, nil
}

// 构建最终输出
func (we *WorkflowEngine) buildFinalOutput() WorkflowOutput {
	// 简化实现：返回最后一个节点的结果
	var lastResult NodeResult
	var lastTimestamp time.Time

	for _, result := range we.nodeResults {
		if ts, err := time.Parse(time.RFC3339, result.Timestamp); err == nil {
			if ts.After(lastTimestamp) {
				lastTimestamp = ts
				lastResult = result
			}
		}
	}

	// 构建元数据
	metadata := make(map[string]interface{})
	metadata["total_nodes"] = len(we.nodeResults)
	metadata["execution_time"] = lastTimestamp.Format(time.RFC3339)

	// 合并所有节点的元数据
	allMetadata := make(map[string]interface{})
	for nodeName, result := range we.nodeResults {
		allMetadata[nodeName] = result.Metadata
	}
	metadata["node_metadata"] = allMetadata

	return WorkflowOutput{
		Result:   fmt.Sprintf("%v", lastResult.Data),
		Metadata: metadata,
		Status:   "success",
	}
}

// ==================== 演示用的Lambda函数 ====================

// 1. 文本预处理节点
func createTextPreprocessor() LambdaFunc {
	return func(ctx context.Context, input interface{}) (NodeResult, error) {
		fmt.Println("  📝 执行文本预处理...")

		inputMap := input.(map[string]interface{})
		text, _ := inputMap["text"].(string)

		// 模拟文本预处理
		processed := strings.TrimSpace(strings.ToLower(text))
		wordCount := len(strings.Fields(processed))

		time.Sleep(200 * time.Millisecond)

		return NodeResult{
			Data: map[string]interface{}{
				"processed_text": processed,
				"word_count":     wordCount,
				"char_count":     len(processed),
			},
			Metadata: map[string]interface{}{
				"processing_time": "200ms",
				"operations":      []string{"trim", "lowercase", "count"},
			},
		}, nil
	}
}

// 2. 情感分析节点
func createSentimentAnalyzer() LambdaFunc {
	return func(ctx context.Context, input interface{}) (NodeResult, error) {
		fmt.Println("  🎭 执行情感分析...")

		inputMap := input.(map[string]interface{})
		text, _ := inputMap["text"].(string)

		// 模拟情感分析
		sentimentScore := 0.0
		sentiment := "neutral"

		if strings.Contains(text, "好") || strings.Contains(text, "棒") {
			sentimentScore = 0.8
			sentiment = "positive"
		} else if strings.Contains(text, "坏") || strings.Contains(text, "差") {
			sentimentScore = -0.6
			sentiment = "negative"
		}

		time.Sleep(300 * time.Millisecond)

		return NodeResult{
			Data: map[string]interface{}{
				"sentiment":       sentiment,
				"sentiment_score": sentimentScore,
				"confidence":      0.85,
			},
			Metadata: map[string]interface{}{
				"model":           "simple_rule_based",
				"processing_time": "300ms",
			},
		}, nil
	}
}

// 3. 关键词提取节点
func createKeywordExtractor() LambdaFunc {
	return func(ctx context.Context, input interface{}) (NodeResult, error) {
		fmt.Println("  🔍 执行关键词提取...")

		inputMap := input.(map[string]interface{})
		text, _ := inputMap["text"].(string)

		// 模拟关键词提取
		words := strings.Fields(text)
		keywords := []string{}

		for _, word := range words {
			if len(word) > 2 && !strings.Contains("的了是在有和", word) {
				keywords = append(keywords, word)
			}
		}

		// 限制关键词数量
		if len(keywords) > 5 {
			keywords = keywords[:5]
		}

		time.Sleep(250 * time.Millisecond)

		return NodeResult{
			Data: map[string]interface{}{
				"keywords":      keywords,
				"keyword_count": len(keywords),
			},
			Metadata: map[string]interface{}{
				"algorithm":       "simple_filter",
				"processing_time": "250ms",
			},
		}, nil
	}
}

// 4. 质量评估节点
func createQualityAssessor() LambdaFunc {
	return func(ctx context.Context, input interface{}) (NodeResult, error) {
		fmt.Println("  ⚖️  执行质量评估...")

		inputMap := input.(map[string]interface{})

		// 获取各种输入数据
		wordCount, _ := inputMap["word_count"].(int)
		sentimentScore, _ := inputMap["sentiment_score"].(float64)
		keywordCount, _ := inputMap["keyword_count"].(int)

		// 计算质量分数
		qualityScore := 0.0

		// 基于词数量的评分
		if wordCount > 10 {
			qualityScore += 0.3
		} else if wordCount > 5 {
			qualityScore += 0.2
		}

		// 基于情感强度的评分
		qualityScore += math.Abs(sentimentScore) * 0.3

		// 基于关键词数量的评分
		if keywordCount > 3 {
			qualityScore += 0.4
		} else if keywordCount > 1 {
			qualityScore += 0.2
		}

		// 确保分数在0-1之间
		if qualityScore > 1.0 {
			qualityScore = 1.0
		}

		quality := "low"
		if qualityScore > 0.7 {
			quality = "high"
		} else if qualityScore > 0.4 {
			quality = "medium"
		}

		time.Sleep(150 * time.Millisecond)

		return NodeResult{
			Data: map[string]interface{}{
				"quality_score": qualityScore,
				"quality_level": quality,
				"factors": map[string]interface{}{
					"word_count":    wordCount,
					"sentiment_abs": math.Abs(sentimentScore),
					"keyword_count": keywordCount,
				},
			},
			Metadata: map[string]interface{}{
				"algorithm":       "weighted_scoring",
				"processing_time": "150ms",
			},
		}, nil
	}
}

// 5. 结果汇总节点
func createResultAggregator() LambdaFunc {
	return func(ctx context.Context, input interface{}) (NodeResult, error) {
		fmt.Println("  📊 执行结果汇总...")

		inputMap := input.(map[string]interface{})

		// 汇总所有分析结果
		summary := map[string]interface{}{
			"text_analysis": map[string]interface{}{
				"processed_text": inputMap["processed_text"],
				"word_count":     inputMap["word_count"],
				"char_count":     inputMap["char_count"],
			},
			"sentiment_analysis": map[string]interface{}{
				"sentiment":       inputMap["sentiment"],
				"sentiment_score": inputMap["sentiment_score"],
				"confidence":      inputMap["confidence"],
			},
			"keyword_extraction": map[string]interface{}{
				"keywords":      inputMap["keywords"],
				"keyword_count": inputMap["keyword_count"],
			},
			"quality_assessment": map[string]interface{}{
				"quality_score": inputMap["quality_score"],
				"quality_level": inputMap["quality_level"],
				"factors":       inputMap["factors"],
			},
		}

		time.Sleep(100 * time.Millisecond)

		return NodeResult{
			Data: summary,
			Metadata: map[string]interface{}{
				"aggregation_type": "comprehensive_summary",
				"processing_time":  "100ms",
			},
		}, nil
	}
}

// ==================== 演示函数 ====================

// 基础工作流演示
func basicWorkflowDemo(ctx context.Context) {
	fmt.Println("=== 基础工作流演示 ===")

	// 创建工作流引擎
	workflow := NewWorkflowEngine()

	// 添加节点
	workflow.AddLambdaNode("preprocessor", createTextPreprocessor()).
		AddInput("preprocessor", map[string]string{
			"text": "input.text",
		})

	workflow.AddLambdaNode("sentiment", createSentimentAnalyzer()).
		AddInput("sentiment", map[string]string{
			"text": "preprocessor.data.processed_text",
		})

	// 执行工作流
	input := WorkflowInput{
		Text:   "这个产品真的很好，我很喜欢！",
		UserID: "user123",
		Options: map[string]interface{}{
			"language": "zh-CN",
		},
	}

	fmt.Printf("📥 输入数据: %+v\n", input)
	result, err := workflow.Invoke(ctx, input)

	if err != nil {
		log.Printf("工作流执行失败: %v", err)
		return
	}

	fmt.Printf("📤 输出结果: %+v\n", result)
}

// 复杂工作流演示
func complexWorkflowDemo(ctx context.Context) {
	fmt.Println("\n=== 复杂工作流演示 ===")

	// 创建复杂的工作流
	workflow := NewWorkflowEngine()

	// 设置静态配置
	workflow.SetStaticValue("min_quality_threshold", 0.5).
		SetStaticValue("max_keywords", 10)

	// 1. 文本预处理节点
	workflow.AddLambdaNode("preprocessor", createTextPreprocessor()).
		AddInput("preprocessor", map[string]string{
			"text":    "input.text",
			"user_id": "input.user_id",
		})

	// 2. 并行分析节点
	workflow.AddLambdaNode("sentiment", createSentimentAnalyzer()).
		AddInput("sentiment", map[string]string{
			"text": "preprocessor.data.processed_text",
		})

	workflow.AddLambdaNode("keywords", createKeywordExtractor()).
		AddInput("keywords", map[string]string{
			"text":      "preprocessor.data.processed_text",
			"max_count": "static.max_keywords",
		})

	// 3. 质量评估节点（依赖多个前驱节点）
	workflow.AddLambdaNode("quality", createQualityAssessor()).
		AddInput("quality", map[string]string{
			"word_count":      "preprocessor.data.word_count",
			"sentiment_score": "sentiment.data.sentiment_score",
			"keyword_count":   "keywords.data.keyword_count",
			"threshold":       "static.min_quality_threshold",
		})

	// 4. 结果汇总节点
	workflow.AddLambdaNode("aggregator", createResultAggregator()).
		AddInput("aggregator", map[string]string{
			// 从预处理节点
			"processed_text": "preprocessor.data.processed_text",
			"word_count":     "preprocessor.data.word_count",
			"char_count":     "preprocessor.data.char_count",
			// 从情感分析节点
			"sentiment":       "sentiment.data.sentiment",
			"sentiment_score": "sentiment.data.sentiment_score",
			"confidence":      "sentiment.data.confidence",
			// 从关键词提取节点
			"keywords":      "keywords.data.keywords",
			"keyword_count": "keywords.data.keyword_count",
			// 从质量评估节点
			"quality_score": "quality.data.quality_score",
			"quality_level": "quality.data.quality_level",
			"factors":       "quality.data.factors",
		})

	// 测试数据
	testInputs := []WorkflowInput{
		{
			Text:   "这个新功能设计得非常棒，用户体验很流畅，界面也很美观，我强烈推荐大家使用！",
			UserID: "user001",
			Options: map[string]interface{}{
				"language": "zh-CN",
				"detailed": true,
			},
		},
		{
			Text:   "不好用",
			UserID: "user002",
			Options: map[string]interface{}{
				"language": "zh-CN",
				"detailed": false,
			},
		},
		{
			Text:   "今天天气不错，适合出门散步，心情很愉快。",
			UserID: "user003",
			Options: map[string]interface{}{
				"language": "zh-CN",
				"detailed": true,
			},
		},
	}

	// 执行多个测试用例
	for i, input := range testInputs {
		fmt.Printf("\n--- 测试用例 %d ---\n", i+1)
		fmt.Printf("📥 输入: %s\n", input.Text)

		result, err := workflow.Invoke(ctx, input)
		if err != nil {
			log.Printf("工作流执行失败: %v", err)
			continue
		}

		fmt.Printf("📤 处理状态: %s\n", result.Status)

		// 解析并显示详细结果
		if result.Metadata != nil {
			if totalNodes, ok := result.Metadata["total_nodes"].(int); ok {
				fmt.Printf("📊 执行节点数: %d\n", totalNodes)
			}
			if execTime, ok := result.Metadata["execution_time"].(string); ok {
				fmt.Printf("⏱️  执行时间: %s\n", execTime)
			}
		}

		// 尝试解析结果数据
		fmt.Printf("📋 分析摘要:\n")
		if summaryData := parseResultData(result.Result); summaryData != nil {
			displayAnalysisSummary(summaryData)
		}
	}
}

// 解析结果数据
func parseResultData(resultStr string) map[string]interface{} {
	// 简化实现：直接返回mock数据用于演示
	// 在实际应用中，这里应该解析真实的结果字符串
	return map[string]interface{}{
		"quality_level": "high",
		"sentiment":     "positive",
		"keyword_count": 5,
	}
}

// 显示分析摘要
func displayAnalysisSummary(data map[string]interface{}) {
	if quality, ok := data["quality_level"].(string); ok {
		fmt.Printf("  • 内容质量: %s\n", quality)
	}
	if sentiment, ok := data["sentiment"].(string); ok {
		fmt.Printf("  • 情感倾向: %s\n", sentiment)
	}
	if kwCount, ok := data["keyword_count"].(int); ok {
		fmt.Printf("  • 关键词数: %d\n", kwCount)
	}
}

// 流式处理演示
func streamingWorkflowDemo(ctx context.Context) {
	fmt.Println("\n=== 流式处理演示 ===")

	// 创建流式处理工作流
	workflow := NewWorkflowEngine()

	// 流式文本处理节点
	streamProcessor := func(ctx context.Context, input interface{}) (NodeResult, error) {
		fmt.Println("  📡 开始流式处理...")

		inputMap := input.(map[string]interface{})
		text, _ := inputMap["text"].(string)

		// 模拟流式处理 - 分块处理
		chunks := strings.Fields(text)
		results := []string{}

		for i, chunk := range chunks {
			// 模拟流式处理延迟
			time.Sleep(50 * time.Millisecond)
			processed := fmt.Sprintf("chunk_%d:%s", i, strings.ToUpper(chunk))
			results = append(results, processed)
			fmt.Printf("    📦 处理块 %d: %s\n", i+1, processed)
		}

		return NodeResult{
			Data: map[string]interface{}{
				"chunks":      results,
				"chunk_count": len(results),
				"stream_mode": true,
			},
			Metadata: map[string]interface{}{
				"processing_mode": "streaming",
				"chunk_size":      "word_level",
			},
		}, nil
	}

	workflow.AddLambdaNode("stream_processor", streamProcessor).
		AddInput("stream_processor", map[string]string{
			"text": "input.text",
		})

	// 执行流式工作流
	input := WorkflowInput{
		Text:   "云计算 人工智能 大数据 物联网 区块链",
		UserID: "streaming_user",
	}

	fmt.Printf("📥 流式输入: %s\n", input.Text)
	result, err := workflow.Invoke(ctx, input)

	if err != nil {
		log.Printf("流式工作流执行失败: %v", err)
		return
	}

	fmt.Printf("✅ 流式处理完成，状态: %s\n", result.Status)
}

// 分支控制演示
func branchControlDemo(ctx context.Context) {
	fmt.Println("\n=== 分支控制演示 ===")

	workflow := NewWorkflowEngine()

	// 分支决策节点
	branchDecision := func(ctx context.Context, input interface{}) (NodeResult, error) {
		fmt.Println("  🔀 执行分支决策...")

		inputMap := input.(map[string]interface{})
		text, _ := inputMap["text"].(string)

		wordCount := len(strings.Fields(text))

		branch := "simple_path"
		if wordCount > 10 {
			branch = "complex_path"
		} else if wordCount > 5 {
			branch = "medium_path"
		}

		fmt.Printf("    📊 词数: %d, 选择分支: %s\n", wordCount, branch)

		return NodeResult{
			Data: map[string]interface{}{
				"branch":     branch,
				"word_count": wordCount,
				"decision":   fmt.Sprintf("基于词数 %d 选择 %s", wordCount, branch),
			},
			Metadata: map[string]interface{}{
				"decision_type": "word_count_based",
			},
		}, nil
	}

	// 简单路径处理
	simplePath := func(ctx context.Context, input interface{}) (NodeResult, error) {
		fmt.Println("  🛤️  执行简单路径处理...")
		time.Sleep(100 * time.Millisecond)

		return NodeResult{
			Data: map[string]interface{}{
				"path":   "simple",
				"result": "简单快速处理完成",
			},
			Metadata: map[string]interface{}{
				"processing_complexity": "low",
			},
		}, nil
	}

	// 复杂路径处理
	complexPath := func(ctx context.Context, input interface{}) (NodeResult, error) {
		fmt.Println("  🧩 执行复杂路径处理...")
		time.Sleep(500 * time.Millisecond)

		inputMap := input.(map[string]interface{})
		wordCount, _ := inputMap["word_count"].(int)

		return NodeResult{
			Data: map[string]interface{}{
				"path":     "complex",
				"result":   "复杂深度分析完成",
				"analysis": fmt.Sprintf("深度分析了 %d 个词", wordCount),
			},
			Metadata: map[string]interface{}{
				"processing_complexity": "high",
			},
		}, nil
	}

	// 构建分支工作流
	workflow.AddLambdaNode("decision", branchDecision).
		AddInput("decision", map[string]string{
			"text": "input.text",
		})

	workflow.AddLambdaNode("simple_process", simplePath).
		AddInput("simple_process", map[string]string{
			"branch": "decision.data.branch",
		})

	workflow.AddLambdaNode("complex_process", complexPath).
		AddInput("complex_process", map[string]string{
			"branch":     "decision.data.branch",
			"word_count": "decision.data.word_count",
		})

	// 测试不同长度的文本
	testCases := []string{
		"短文本",
		"这是一个中等长度的测试文本，包含了一些关键信息",
		"这是一个非常长的测试文本，包含了大量的信息和细节描述，需要进行复杂的分析处理，以便获得更准确和详细的结果，适合测试复杂路径的处理逻辑",
	}

	for i, testText := range testCases {
		fmt.Printf("\n--- 分支测试 %d ---\n", i+1)

		input := WorkflowInput{
			Text:   testText,
			UserID: fmt.Sprintf("branch_user_%d", i+1),
		}

		fmt.Printf("📥 输入文本长度: %d 词\n", len(strings.Fields(testText)))
		fmt.Printf("📄 内容预览: %.50s...\n", testText)

		result, err := workflow.Invoke(ctx, input)
		if err != nil {
			log.Printf("分支工作流执行失败: %v", err)
			continue
		}

		fmt.Printf("✅ 分支处理完成，状态: %s\n", result.Status)
	}
}

// 配置初始化
func initWorkflowConfig() {
	viper.SetConfigFile("../config.yaml")
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("读取配置文件失败: %v (使用默认配置)", err)
	}
}

func main() {
	initWorkflowConfig()
	ctx := context.Background()

	// 运行各种工作流演示
	basicWorkflowDemo(ctx)
	complexWorkflowDemo(ctx)
	streamingWorkflowDemo(ctx)
	branchControlDemo(ctx)

	fmt.Println("\n🎉 工作流编排演示完成！")
}
