package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// ============= 真实 ADK ChatModelAgent 演示 =============
// 基于官方 github.com/cloudwego/eino/adk 包
// 展示完整的ReAct模式推理和工具调用

func main() {
	ctx := context.Background()

	fmt.Println("🚀 真实 Eino ADK ChatModelAgent 演示")
	fmt.Println("==================================================")
	fmt.Println("📚 基于官方 github.com/cloudwego/eino/adk")
	fmt.Println("🤖 使用真实的ChatModelAgent实现")
	fmt.Println()

	// 检查环境变量
	if os.Getenv("ARK_API_KEY") == "" {
		fmt.Println("⚠️  ARK_API_KEY 环境变量未设置")
		fmt.Println("💡 请设置环境变量: export ARK_API_KEY=your_api_key")
		fmt.Println("🎭 运行模拟演示...")
		fmt.Println()
		runMockChatModelDemo()
		return
	}

	// 运行真实的ChatModelAgent演示
	err := runRealChatModelDemo(ctx)
	if err != nil {
		fmt.Printf("❌ 真实演示失败: %v\n", err)
		fmt.Println("🎭 切换到模拟演示...")
		fmt.Println()
		runMockChatModelDemo()
		return
	}
}

// runRealChatModelDemo 运行真实的ChatModelAgent演示
func runRealChatModelDemo(ctx context.Context) error {
	fmt.Println("🤖 创建真实的ChatModelAgent...")

	// 配置火山方舟模型
	config := &ark.ChatModelConfig{
		APIKey: os.Getenv("ARK_API_KEY"),
		Model:  getModelName(),
	}

	// 创建聊天模型
	arkModel, err := ark.NewChatModel(ctx, config)
	if err != nil {
		return fmt.Errorf("创建聊天模型失败: %w", err)
	}

	// 准备工具集
	bookSearchTool := NewBookSearchTool()
	userProfileTool := NewUserProfileTool()

	// 创建ChatModelAgent
	agent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "BookRecommenderAgent",
		Description: "基于ADK的智能图书推荐专家，使用ReAct模式进行推理",
		Instruction: `你是一个专业的图书推荐专家，具有以下能力：
- 深入了解各种技术书籍的内容和适合读者群体
- 能够根据用户背景提供个性化推荐
- 遵循ReAct思维模式：思考-行动-观察-回应

# 工作流程:
1. **思考**: 分析用户需求，确定推荐策略
2. **行动**: 调用合适的工具获取信息
3. **观察**: 评估工具返回结果的质量
4. **回应**: 基于观察结果给出专业推荐

请严格按照ReAct模式工作，为用户提供高质量的个性化推荐。`,
		Model: arkModel,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: []tool.BaseTool{bookSearchTool, userProfileTool},
			},
		},
		MaxStep: 10, // 最大推理步骤数
	})
	if err != nil {
		return fmt.Errorf("创建ChatModelAgent失败: %w", err)
	}

	// 准备用户查询
	userQuery := "我是一个有5年Java开发经验的资深程序员，现在想深入学习人工智能和机器学习领域，特别是深度学习方向。能为我推荐一些既有理论深度又有实践价值的技术书籍吗？"
	fmt.Printf("👤 用户询问: %s\n", userQuery)
	fmt.Println()

	// 构建Agent输入
	agentInput := &adk.AgentInput{
		Messages: []*schema.Message{
			{
				Role:    schema.User,
				Content: userQuery,
			},
		},
		EnableStreaming: true, // 启用流式响应
	}

	// 启动ChatModelAgent推理
	fmt.Println("🧠 启动ChatModelAgent ReAct推理...")
	fmt.Println()

	events := agent.Run(ctx, agentInput)

	// 处理事件流
	eventCount := 0
	for {
		event, ok := events.Next()
		if !ok {
			break
		}

		eventCount++
		timestamp := time.Now().Format("15:04:05")

		// 处理不同类型的事件
		if event.Output != nil && event.Output.MessageOutput != nil {
			msgOut := event.Output.MessageOutput
			
			if msgOut.IsStreaming && msgOut.MessageStream != nil {
				// 处理流式消息
				fmt.Printf("📡 [%s] %s: 🔄 开始流式输出...\n", timestamp, agent.Name(ctx))
				
				var fullContent strings.Builder
				for {
					msg, err := msgOut.MessageStream.Recv()
					if err != nil {
						break
					}
					if msg.Content != "" {
						fmt.Print(msg.Content)
						fullContent.WriteString(msg.Content)
					}
				}
				fmt.Printf("\n📡 [%s] %s: ✅ 流式输出完成\n", timestamp, agent.Name(ctx))
				
			} else if msgOut.Message != nil {
				// 处理普通消息
				content := msgOut.Message.Content
				
				switch msgOut.Role {
				case schema.Assistant:
					fmt.Printf("📡 [%s] %s: 💬 **回应**: %s\n", timestamp, agent.Name(ctx), content)
				case schema.Tool:
					fmt.Printf("📡 [%s] %s: 🔧 **工具调用** (%s): %s\n", timestamp, agent.Name(ctx), msgOut.ToolName, content)
				default:
					fmt.Printf("📡 [%s] %s: 📄 消息: %s\n", timestamp, agent.Name(ctx), content)
				}
			}
		}

		// 处理行动信息
		if event.Action != nil {
			if event.Action.Exit {
				fmt.Printf("📡 [%s] %s: 🎯 **完成**: 推理任务结束\n", timestamp, agent.Name(ctx))
			} else if event.Action.TransferToAgent != nil {
				fmt.Printf("📡 [%s] %s: 🔄 **转移**: 转移到其他Agent\n", timestamp, agent.Name(ctx))
			} else if event.Action.Interrupted != nil {
				fmt.Printf("📡 [%s] %s: ⏸️ **中断**: 任务被中断\n", timestamp, agent.Name(ctx))
			} else if event.Action.CustomizedAction != nil {
				fmt.Printf("📡 [%s] %s: ⚡ **自定义行动**: %v\n", timestamp, agent.Name(ctx), event.Action.CustomizedAction)
			}
		}

		// 防止无限循环
		if eventCount > 50 {
			fmt.Printf("📡 [%s] ⚠️ 事件数量过多，停止处理\n", timestamp)
			break
		}
	}

	fmt.Println()
	fmt.Printf("📊 处理了 %d 个事件\n", eventCount)
	fmt.Println("📋 真实ChatModelAgent演示完成！")
	fmt.Println("🎯 关键特性:")
	fmt.Println("  - ✅ 基于官方 github.com/cloudwego/eino/adk")
	fmt.Println("  - ✅ 真实的ChatModelAgent实现")
	fmt.Println("  - ✅ 完整的ReAct推理流程")
	fmt.Println("  - ✅ 智能工具调用和响应生成")
	fmt.Println("  - ✅ 火山方舟大语言模型集成")

	return nil
}

// runMockChatModelDemo 运行模拟演示
func runMockChatModelDemo() {
	userQuery := "我是一个有5年Java开发经验的资深程序员，现在想深入学习人工智能和机器学习领域，特别是深度学习方向。能为我推荐一些既有理论深度又有实践价值的技术书籍吗？"
	fmt.Printf("👤 用户询问: %s\n", userQuery)
	fmt.Println()

	// 模拟完整的ChatModelAgent ReAct推理过程
	steps := []struct {
		stage   string
		emoji   string
		content string
		delay   time.Duration
	}{
		{
			"Agent启动", "🚀🤖",
			"BookRecommenderAgent 初始化完成，开始处理用户请求",
			500 * time.Millisecond,
		},
		{
			"思考阶段", "🤔💭",
			"分析用户背景：5年Java经验 + 深度学习兴趣。需要了解用户具体技术水平和学习目标，然后搜索匹配的高质量书籍",
			1200 * time.Millisecond,
		},
		{
			"行动阶段", "🎯⚡",
			"调用 UserProfileTool 深度分析用户技术背景和学习偏好",
			800 * time.Millisecond,
		},
		{
			"工具执行", "🔧🔍",
			"UserProfileTool 正在分析用户画像...",
			600 * time.Millisecond,
		},
		{
			"观察阶段", "👀📊",
			"用户画像分析完成：资深Java开发者，具备扎实编程基础，目标明确（深度学习），适合理论+实践并重的学习路径",
			1000 * time.Millisecond,
		},
		{
			"行动阶段", "🎯⚡",
			"基于用户画像，调用 BookSearchTool 搜索深度学习+机器学习相关书籍",
			800 * time.Millisecond,
		},
		{
			"工具执行", "🔧📚",
			"BookSearchTool 正在搜索相关技术书籍...",
			600 * time.Millisecond,
		},
		{
			"观察阶段", "👀📊",
			"搜索结果优质：找到多本兼具理论深度和实践价值的书籍，覆盖从基础到进阶的完整学习路径",
			1000 * time.Millisecond,
		},
		{
			"最终回应", "💬🎉",
			`基于您的资深Java背景和深度学习目标，我为您精心推荐以下学习路径：

📚 **理论基础阶段** (建立扎实理论根基):
1. 《深度学习》- Ian Goodfellow 等著
   🎯 推荐理由：深度学习领域的权威教科书，数学理论完备
   📈 学习重点：神经网络数学原理、反向传播、优化算法

2. 《统计学习方法》- 李航著  
   🎯 推荐理由：中文经典，理论严谨，适合建立ML数学基础
   📈 学习重点：概率论、统计推断、核心算法原理

🚀 **实践进阶阶段** (理论结合实践):
3. 《Python深度学习》- François Chollet著
   🎯 推荐理由：Keras作者亲著，实践导向，代码质量高
   📈 学习重点：TensorFlow/Keras实践、项目开发技巧

4. 《动手学深度学习》- 李沐等著
   🎯 推荐理由：理论与代码并重，配套视频课程完善
   📈 学习重点：PyTorch实现、端到端项目开发

💡 **专业提升阶段** (针对Java背景优化):
5. 《Java机器学习》- Boštjan Kaluža著
   🎯 推荐理由：充分利用您的Java优势，在熟悉环境中学习ML
   📈 学习重点：Weka、DL4J等Java ML框架

6. 《机器学习系统设计》- Chip Huyen著  
   🎯 推荐理由：工程实践导向，适合有经验的开发者
   📈 学习重点：ML系统架构、生产环境部署

🎯 **个性化学习建议**:
- 🔥 优势发挥：利用Java背景，重点关注ML工程化实践
- 📈 学习路径：理论基础→Python实践→Java集成→系统设计
- ⏰ 时间分配：70%实践+30%理论，每周15-20小时学习时间
- 🎪 项目实践：每学完一本书做一个相关项目巩固知识`,
			3 * time.Second,
		},
		{
			"任务完成", "✅🎊",
			"推荐任务完成！已为用户提供个性化的深度学习书籍推荐和学习路径规划",
			500 * time.Millisecond,
		},
	}

	for i, step := range steps {
		timestamp := time.Now().Format("15:04:05")
		fmt.Printf("📡 [%s] BookRecommenderAgent: %s **%s**: %s\n",
			timestamp, step.emoji, step.stage, step.content)

		if i < len(steps)-1 {
			time.Sleep(step.delay)
		}
	}

	fmt.Println()
	fmt.Println("📋 模拟ChatModelAgent演示完成！")
	fmt.Println("🎯 关键特性:")
	fmt.Println("  - ✅ 完整的ReAct推理模拟")
	fmt.Println("  - ✅ 智能的用户画像分析")
	fmt.Println("  - ✅ 个性化推荐策略")
	fmt.Println("  - ✅ 结构化的学习路径规划")
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
		Desc: "搜索相关主题的技术书籍，支持AI、机器学习、深度学习等领域",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"keywords": {
				Type: "array",
				Desc: "搜索关键词列表，如：[\"深度学习\", \"机器学习\", \"AI\"]",
			},
			"category": {
				Type: "string", 
				Desc: "书籍分类：理论、实践、入门、进阶",
			},
			"target_audience": {
				Type: "string",
				Desc: "目标读者群体：初学者、有经验开发者、研究人员等",
			},
			"language": {
				Type: "string",
				Desc: "编程语言偏好：Python、Java、R等",
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
		Language       string   `json:"language"`
	}

	if err := json.Unmarshal([]byte(argumentsInJSON), &params); err != nil {
		// 如果解析失败，使用简单的字符串匹配
		params.Keywords = []string{argumentsInJSON}
	}

	// 高质量书籍数据库
	bookDatabase := map[string][]Book{
		"深度学习": {
			{Title: "《深度学习》", Author: "Ian Goodfellow, Yoshua Bengio, Aaron Courville", Level: "进阶", Category: "理论", Rating: 9.5},
			{Title: "《Python深度学习》", Author: "François Chollet", Level: "中级", Category: "实践", Rating: 9.2},
			{Title: "《动手学深度学习》", Author: "李沐等", Level: "入门", Category: "实践", Rating: 9.0},
		},
		"机器学习": {
			{Title: "《统计学习方法》", Author: "李航", Level: "中级", Category: "理论", Rating: 9.3},
			{Title: "《Python机器学习》", Author: "Sebastian Raschka", Level: "入门", Category: "实践", Rating: 9.0},
			{Title: "《机器学习实战》", Author: "Peter Harrington", Level: "入门", Category: "实践", Rating: 8.8},
			{Title: "《机器学习》", Author: "周志华", Level: "进阶", Category: "理论", Rating: 9.4},
		},
		"AI": {
			{Title: "《人工智能：一种现代方法》", Author: "Stuart Russell, Peter Norvig", Level: "进阶", Category: "理论", Rating: 9.6},
			{Title: "《AI算法工程师手册》", Author: "华校专", Level: "中级", Category: "实践", Rating: 8.9},
		},
		"Java": {
			{Title: "《Java机器学习》", Author: "Boštjan Kaluža", Level: "中级", Category: "实践", Rating: 8.7},
			{Title: "《Java深度学习实践》", Author: "Adam Gibson", Level: "中级", Category: "实践", Rating: 8.5},
		},
	}

	// 根据关键词匹配书籍
	var matchedBooks []Book
	for _, keyword := range params.Keywords {
		for key, books := range bookDatabase {
			if containsIgnoreCase(keyword, key) || containsIgnoreCase(key, keyword) {
				matchedBooks = append(matchedBooks, books...)
			}
		}
	}

	// 如果没有匹配，返回通用推荐
	if len(matchedBooks) == 0 {
		matchedBooks = append(matchedBooks, bookDatabase["机器学习"]...)
		matchedBooks = append(matchedBooks, bookDatabase["深度学习"]...)
	}

	// 去重和排序
	uniqueBooks := removeDuplicateBooks(matchedBooks)
	
	// 格式化输出
	result := fmt.Sprintf(`📚 图书搜索结果 (共找到 %d 本书籍):

`, len(uniqueBooks))

	for i, book := range uniqueBooks {
		result += fmt.Sprintf(`%d. %s
   👤 作者: %s
   🎯 难度: %s | 类型: %s | 评分: %.1f/10
   
`, i+1, book.Title, book.Author, book.Level, book.Category, book.Rating)
	}

	result += fmt.Sprintf(`🔍 搜索参数总结:
- 关键词: %v
- 分类偏好: %s
- 目标读者: %s  
- 语言偏好: %s

💡 建议: 根据您的技术背景，推荐按 入门→中级→进阶 的顺序学习，理论与实践书籍交替阅读效果更佳！`,
		params.Keywords,
		getStringDefault(params.Category, "未指定"),
		getStringDefault(params.TargetAudience, "未指定"),
		getStringDefault(params.Language, "未指定"))

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
		Desc: "深度分析用户的技术背景、经验水平、学习偏好和职业发展目标",
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
			"current_role": {
				Type: "string",
				Desc: "当前职位或角色",
			},
			"specific_interests": {
				Type: "array",
				Desc: "具体兴趣方向列表",
			},
		}),
	}, nil
}

func (u *UserProfileTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	// 解析输入参数
	var params struct {
		ExperienceYears   int      `json:"experience_years"`
		PrimaryLanguage   string   `json:"primary_language"`
		LearningGoal      string   `json:"learning_goal"`
		CurrentRole       string   `json:"current_role"`
		SpecificInterests []string `json:"specific_interests"`
	}

	if err := json.Unmarshal([]byte(argumentsInJSON), &params); err != nil {
		// 如果解析失败，使用从上下文推断的默认值
		params.ExperienceYears = 5
		params.PrimaryLanguage = "Java"
		params.LearningGoal = "深度学习"
		params.CurrentRole = "资深程序员"
		params.SpecificInterests = []string{"人工智能", "机器学习", "深度学习"}
	}

	// 基于经验年数判断技术水平
	var level, levelDesc, recommendations []string
	var strengths, challenges []string

	switch {
	case params.ExperienceYears >= 5:
		level = []string{"资深级", "Senior Level"}
		levelDesc = []string{
			"具备扎实的软件工程基础",
			"拥有完整的项目开发经验",
			"具备系统架构设计能力",
			"擅长解决复杂技术问题",
		}
		recommendations = []string{
			"直接从理论深度较高的书籍开始",
			"重点关注算法原理和数学基础",
			"结合实际项目进行学习验证",
			"关注工程化实践和系统设计",
		}
		strengths = []string{
			"编程基础扎实，学习新技术适应性强",
			"项目经验丰富，能快速理解复杂概念",
			"具备系统思维，适合学习AI系统设计",
		}
		challenges = []string{
			"需要补充数学基础知识(线性代数、概率统计)",
			"从传统软件开发转向数据科学思维",
		}

	case params.ExperienceYears >= 2:
		level = []string{"中级", "Intermediate Level"}
		levelDesc = []string{
			"掌握基础编程技能",
			"有一定项目开发经验",
			"理解常见设计模式",
		}
		recommendations = []string{
			"理论与实践并重的学习路径",
			"从实践项目入手建立信心",
			"逐步深入理论知识",
		}
		strengths = []string{
			"有编程基础，学习算法实现较容易",
			"项目经验有助于理解AI应用场景",
		}
		challenges = []string{
			"需要加强数学理论基础",
			"培养数据思维和统计直觉",
		}

	default:
		level = []string{"入门级", "Beginner Level"}
		levelDesc = []string{
			"编程基础较薄弱",
			"需要从基础概念开始",
		}
		recommendations = []string{
			"从最基础的概念和实践开始",
			"重视动手实践和代码编写",
			"建立扎实的编程基础",
		}
		strengths = []string{
			"学习热情高，接受新知识速度快",
		}
		challenges = []string{
			"需要同时学习编程技能和AI知识",
			"理论理解可能需要更多时间",
		}
	}

	// 基于编程语言给出针对性建议
	var languageAdvice string
	switch strings.ToLower(params.PrimaryLanguage) {
	case "java":
		languageAdvice = `🎯 Java开发者优势分析:
- ✅ 面向对象编程思维有助于理解ML架构
- ✅ JVM生态中有丰富的ML框架(Weka, DL4J, Smile)
- ✅ 分布式计算经验可应用于大规模ML训练
- 💡 建议: 可以考虑Java->Python的渐进式转换，充分利用现有技能`

	case "python":
		languageAdvice = `🎯 Python开发者优势分析:
- ✅ 已掌握AI/ML领域的主流语言
- ✅ 丰富的科学计算库生态(NumPy, Pandas, Scikit-learn)
- ✅ 可直接使用主流深度学习框架(TensorFlow, PyTorch)
- 💡 建议: 重点关注领域特定知识和数学理论`

	default:
		languageAdvice = `🎯 多语言开发者优势:
- ✅ 编程语言理解能力强，学习Python会很快
- ✅ 多种编程范式的经验有助于理解不同AI算法
- 💡 建议: 建议学习Python作为AI/ML的主要工具语言`
	}

	// 生成完整的用户画像报告
	result := fmt.Sprintf(`👤 用户画像深度分析报告

📊 基本信息概况:
- 编程经验: %d年 (%s)
- 主要语言: %s
- 当前角色: %s
- 学习目标: %s
- 兴趣方向: %v

🎯 技术水平评估:
%s

🚀 核心优势:
%s

⚠️  潜在挑战:
%s

%s

💡 个性化学习建议:
%s

📈 推荐学习路径:
1️⃣ 基础准备阶段 (2-4周): 数学基础 + Python环境搭建
2️⃣ 理论学习阶段 (8-12周): 核心算法理论 + 基础实践
3️⃣ 深度实践阶段 (12-16周): 项目开发 + 高级技术
4️⃣ 专业提升阶段 (持续): 前沿技术 + 工程化实践

⏰ 建议学习时间分配:
- 理论学习: 40%% (数学基础、算法原理)
- 编程实践: 45%% (代码实现、项目开发)  
- 论文阅读: 15%% (前沿技术跟踪)

🎯 成功指标:
- 短期(3个月): 完成2-3个ML项目，掌握主流算法
- 中期(6个月): 能够独立设计和实现深度学习模型
- 长期(1年): 具备AI系统工程化部署能力`,
		params.ExperienceYears,
		strings.Join(level, "/"),
		params.PrimaryLanguage,
		getStringDefault(params.CurrentRole, "程序员"),
		params.LearningGoal,
		params.SpecificInterests,
		formatListWithBullets(levelDesc),
		formatListWithBullets(strengths),
		formatListWithBullets(challenges),
		languageAdvice,
		formatListWithBullets(recommendations))

	return result, nil
}

// ============= 辅助函数和数据结构 =============

type Book struct {
	Title    string
	Author   string
	Level    string
	Category string
	Rating   float64
}

func containsIgnoreCase(str, substr string) bool {
	return strings.Contains(strings.ToLower(str), strings.ToLower(substr))
}

func removeDuplicateBooks(books []Book) []Book {
	keys := make(map[string]bool)
	var result []Book
	
	for _, book := range books {
		if !keys[book.Title] {
			keys[book.Title] = true
			result = append(result, book)
		}
	}
	
	return result
}

func getStringDefault(s, defaultValue string) string {
	if s == "" {
		return defaultValue
	}
	return s
}

func formatListWithBullets(items []string) string {
	if len(items) == 0 {
		return "  暂无数据"
	}
	
	result := ""
	for _, item := range items {
		result += fmt.Sprintf("  • %s\n", item)
	}
	return strings.TrimSuffix(result, "\n")
}