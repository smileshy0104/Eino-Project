package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
)

// ============= 简化版 Eino React Agent 演示 =============
// 基于官方 github.com/cloudwego/eino/flow/agent/react
// 展示完整的ReAct推理模式

func main() {
	ctx := context.Background()

	fmt.Println("🌟 Eino React Agent 简化演示")
	fmt.Println("==================================================")
	fmt.Println("📚 基于官方 github.com/cloudwego/eino/flow/agent/react")
	fmt.Println()

	// 检查环境变量
	if os.Getenv("ARK_API_KEY") == "" {
		fmt.Println("⚠️  ARK_API_KEY 环境变量未设置")
		fmt.Println("💡 请设置环境变量: export ARK_API_KEY=your_api_key")
		fmt.Println("🎭 运行模拟演示...")
		fmt.Println()
		runMockReactDemo()
		return
	}

	// 运行真实的React Agent演示
	runRealReactDemo(ctx)
}

// runRealReactDemo 运行真实的React Agent演示
func runRealReactDemo(ctx context.Context) {
	fmt.Println("🤖 创建真实的React Agent...")

	// 配置火山方舟模型
	config := &ark.ChatModelConfig{
		APIKey: os.Getenv("ARK_API_KEY"),
		Model:  getModelName(),
	}

	// 创建聊天模型
	arkModel, err := ark.NewChatModel(ctx, config)
	if err != nil {
		fmt.Printf("❌ 创建聊天模型失败: %v\n", err)
		fmt.Println("🎭 运行模拟演示...")
		fmt.Println()
		runMockReactDemo()
		return
	}

	// 准备工具集
	bookSearchTool := NewBookSearchTool()
	userProfileTool := NewUserProfileTool()

	// 创建React Agent
	reactAgent, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: arkModel,
		ToolsConfig: compose.ToolsNodeConfig{
			Tools: []tool.BaseTool{bookSearchTool, userProfileTool},
		},
	})
	if err != nil {
		fmt.Printf("❌ 创建React Agent失败: %v\n", err)
		fmt.Println("🎭 运行模拟演示...")
		fmt.Println()
		runMockReactDemo()
		return
	}

	// 准备用户查询
	userQuery := "我是一个有3年Java开发经验的程序员，想学习人工智能和机器学习，能推荐一些适合的技术书籍吗？"
	fmt.Printf("👤 用户询问: %s\n", userQuery)
	fmt.Println()

	// 构建输入消息
	input := []*schema.Message{
		{
			Role:    schema.User,
			Content: userQuery,
		},
	}

	// 启动React推理
	fmt.Println("🧠 启动React推理循环...")
	fmt.Println()

	output, err := reactAgent.Generate(ctx, input)
	if err != nil {
		fmt.Printf("❌ React推理失败: %v\n", err)
		return
	}

	fmt.Printf("🎉 最终推荐结果:\n%s\n", output.Content)
	
	fmt.Println()
	fmt.Println("📋 演示完成！")
	fmt.Println("🎯 关键特性:")
	fmt.Println("  - ✅ 真实的 Eino React Agent")
	fmt.Println("  - ✅ 完整的ReAct推理循环")
	fmt.Println("  - ✅ 多工具智能调用")
	fmt.Println("  - ✅ 基于火山方舟大模型")
}

// runMockReactDemo 运行模拟演示
func runMockReactDemo() {
	userQuery := "我是一个有3年Java开发经验的程序员，想学习人工智能和机器学习，能推荐一些适合的技术书籍吗？"
	fmt.Printf("👤 用户询问: %s\n", userQuery)
	fmt.Println()

	// 模拟完整的ReAct推理过程
	steps := []struct {
		stage   string
		emoji   string
		content string
		delay   time.Duration
	}{
		{
			"思考", "🤔💭",
			"用户是有经验的Java开发者，想转向AI领域。我需要了解用户的具体背景，然后推荐适合的学习路径和书籍。",
			1 * time.Second,
		},
		{
			"行动", "🎯⚡",
			"调用 UserProfileTool 分析用户技术背景和学习偏好",
			800 * time.Millisecond,
		},
		{
			"观察", "👀📊",
			"用户具有扎实的编程基础，偏好实践导向的学习方式，适合从机器学习基础开始，逐步深入AI理论",
			1200 * time.Millisecond,
		},
		{
			"行动", "🎯⚡",
			"调用 BookSearchTool 搜索适合Java开发者的AI/ML书籍",
			800 * time.Millisecond,
		},
		{
			"观察", "👀📊",
			"找到多本优质书籍，包含理论基础、实践项目和进阶内容，能够形成完整的学习体系",
			1000 * time.Millisecond,
		},
		{
			"回应", "💬🎉",
			`基于您的Java开发背景，我推荐以下学习路径：

🎯 **入门阶段** (建立基础概念):
1. 《Python机器学习》- Sebastian Raschka 
   ✅ 为什么推荐：从编程角度介绍ML，对有编程基础的人友好
   📚 学习重点：Python语法 + ML基础算法

🚀 **进阶阶段** (深入理论):
2. 《统计学习方法》- 李航
   ✅ 为什么推荐：数学理论扎实，中文教材理解更容易
   📚 学习重点：算法原理和数学推导

3. 《深度学习》- Ian Goodfellow
   ✅ 为什么推荐：深度学习领域的经典教科书
   📚 学习重点：神经网络理论和前沿技术

💻 **实践阶段** (项目导向):
4. 《机器学习实战》- Peter Harrington
   ✅ 为什么推荐：代码实现导向，适合动手实践
   📚 学习重点：端到端项目开发

📈 **建议学习顺序**: 
   Python基础 → ML理论 → 项目实践 → 深度学习专项`,
			3 * time.Second,
		},
	}

	for i, step := range steps {
		timestamp := time.Now().Format("15:04:05")
		fmt.Printf("📡 [%s] ReactBookAgent: %s **%s**: %s\n",
			timestamp, step.emoji, step.stage, step.content)

		if i < len(steps)-1 {
			time.Sleep(step.delay)
		}
	}

	fmt.Println()
	fmt.Println("📋 模拟演示完成！")
	fmt.Println("🎯 关键特性:")
	fmt.Println("  - ✅ 模拟ReAct推理循环")
	fmt.Println("  - ✅ 完整的思考-行动-观察-回应流程")
	fmt.Println("  - ✅ 个性化推荐策略")
	fmt.Println("  - ✅ 结构化的学习路径")
}

// getModelName 获取模型名称
func getModelName() string {
	model := os.Getenv("ARK_MODEL_NAME")
	if model == "" {
		return "doubao-seed-1-6-250615" // 默认模型
	}
	return model
}

// ============= 工具实现 =============

// BookSearchTool 图书搜索工具
type BookSearchTool struct{}

func NewBookSearchTool() tool.InvokableTool {
	return &BookSearchTool{}
}

func (b *BookSearchTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "book_search_tool",
		Desc: "搜索相关主题的技术书籍，支持AI、机器学习、编程等领域",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"keywords": {
				Type: "array",
				Desc: "搜索关键词列表",
			},
			"category": {
				Type: "string",
				Desc: "书籍分类：理论、实践、入门、进阶",
			},
			"target_audience": {
				Type: "string",
				Desc: "目标读者群体：初学者、有经验开发者等",
			},
		}),
	}, nil
}

func (b *BookSearchTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	// 解析输入参数
	var params struct {
		Keywords       []string `json:"keywords"`
		Category       string   `json:"category"`
		TargetAudience string   `json:"target_audience"`
	}

	if err := json.Unmarshal([]byte(argumentsInJSON), &params); err != nil {
		// 如果解析失败，使用简单的字符串匹配
		params.Keywords = []string{argumentsInJSON}
	}

	// 模拟图书搜索逻辑
	mockResults := map[string][]string{
		"ai":               {"《人工智能：一种现代方法》", "《深度学习》", "《机器学习》"},
		"machine_learning": {"《Python机器学习》", "《统计学习方法》", "《机器学习实战》"},
		"programming":      {"《算法导论》", "《设计模式》", "《重构》"},
		"java":            {"《Effective Java》", "《Java核心技术》", "《Spring实战》"},
		"python":          {"《Python编程：从入门到实践》", "《流畅的Python》", "《Python数据分析》"},
	}

	// 根据关键词搜索书籍
	var books []string
	for _, keyword := range params.Keywords {
		for k, bookList := range mockResults {
			if contains(keyword, k) {
				books = append(books, bookList...)
			}
		}
	}

	// 如果没有匹配的书籍，返回默认推荐
	if len(books) == 0 {
		books = []string{
			"《Python机器学习》- Sebastian Raschka",
			"《统计学习方法》- 李航",
			"《深度学习》- Ian Goodfellow",
			"《机器学习实战》- Peter Harrington",
		}
	}

	result := fmt.Sprintf(`🔍 图书搜索结果:
找到 %d 本相关书籍:
%s

📊 搜索参数:
- 关键词: %v
- 分类: %s  
- 目标读者: %s`,
		len(books),
		formatBookList(books),
		params.Keywords,
		params.Category,
		params.TargetAudience)

	return result, nil
}

// UserProfileTool 用户画像分析工具
type UserProfileTool struct{}

func NewUserProfileTool() tool.InvokableTool {
	return &UserProfileTool{}
}

func (u *UserProfileTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "user_profile_tool",
		Desc: "分析用户的技术背景、经验水平和学习偏好",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"experience_years": {
				Type: "integer",
				Desc: "编程经验年数",
			},
			"primary_language": {
				Type: "string",
				Desc: "主要编程语言",
			},
			"learning_goal": {
				Type: "string",
				Desc: "学习目标领域",
			},
		}),
	}, nil
}

func (u *UserProfileTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	// 解析输入参数
	var params struct {
		ExperienceYears int    `json:"experience_years"`
		PrimaryLanguage string `json:"primary_language"`
		LearningGoal    string `json:"learning_goal"`
	}

	if err := json.Unmarshal([]byte(argumentsInJSON), &params); err != nil {
		// 如果解析失败，使用默认值
		params.ExperienceYears = 3
		params.PrimaryLanguage = "Java"
		params.LearningGoal = "AI/ML"
	}

	// 基于参数生成用户画像
	var level string
	var recommendations []string

	switch {
	case params.ExperienceYears >= 5:
		level = "高级"
		recommendations = []string{
			"直接学习高级AI理论",
			"关注最新研究论文",
			"参与开源AI项目",
		}
	case params.ExperienceYears >= 2:
		level = "中级"
		recommendations = []string{
			"从实践项目入手",
			"理论与实践并重",
			"建立系统性知识体系",
		}
	default:
		level = "初级"
		recommendations = []string{
			"先掌握基础概念",
			"从简单项目开始",
			"注重动手实践",
		}
	}

	result := fmt.Sprintf(`👤 用户画像分析结果:

🎯 基本信息:
- 编程经验: %d年 (%s水平)
- 主要语言: %s
- 学习目标: %s

💡 学习风格分析:
- 技术基础: %s开发者具有扎实的编程基础
- 学习能力: 具备良好的逻辑思维和问题解决能力
- 推荐方式: 实践导向，理论结合项目

🚀 个性化建议:
%s

📈 优势分析:
- ✅ 有%s开发经验，容易理解算法实现
- ✅ 逻辑思维能力强，适合学习ML数学基础
- ✅ 有项目开发经验，能快速上手AI项目实践`,
		params.ExperienceYears,
		level,
		params.PrimaryLanguage,
		params.LearningGoal,
		params.PrimaryLanguage,
		formatRecommendations(recommendations),
		params.PrimaryLanguage)

	return result, nil
}

// ============= 辅助函数 =============

// contains 检查字符串是否包含子字符串
func contains(str, substr string) bool {
	return len(str) >= len(substr) && (str == substr || containsIgnoreCase(str, substr))
}

func containsIgnoreCase(str, substr string) bool {
	str = strings.ToLower(str)
	substr = strings.ToLower(substr)
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func formatBookList(books []string) string {
	result := ""
	for i, book := range books {
		result += fmt.Sprintf("  %d. %s\n", i+1, book)
	}
	return result
}

func formatRecommendations(recs []string) string {
	result := ""
	for i, rec := range recs {
		result += fmt.Sprintf("  %d. %s\n", i+1, rec)
	}
	return result
}