package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/spf13/viper"
)

// 配置结构体
type Config struct {
	APIKey  string `mapstructure:"api_key"`
	Model   string `mapstructure:"model"`
	BaseURL string `mapstructure:"base_url"`
}

// 初始化配置
func initConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./")
	_ = viper.ReadInConfig() // 忽略错误，因为我们也会检查环境变量

	return &Config{
		APIKey:  viper.GetString("ARK_API_KEY"),
		Model:   viper.GetString("ARK_MODEL"),
		BaseURL: viper.GetString("BASE_URL"),
	}, nil
}

// 初始化ChatModel
func initChatModel(ctx context.Context, config *Config) (model.BaseChatModel, error) {
	timeout := 30 * time.Second
	chatModelConfig := &ark.ChatModelConfig{
		APIKey:  config.APIKey,
		Model:   config.Model,
		Timeout: &timeout,
	}

	if config.BaseURL != "" {
		chatModelConfig.BaseURL = config.BaseURL
	}

	cm, err := ark.NewChatModel(ctx, chatModelConfig)
	if err != nil {
		return nil, fmt.Errorf("初始化ChatModel失败: %w", err)
	}

	return cm, nil
}

// 基础模板格式化示例
func basicTemplateExample(ctx context.Context) {
	fmt.Println("\n=== 基础模板格式化示例 ===")

	// 创建简单模板
	template := prompt.FromMessages(schema.FString,
		schema.SystemMessage("你是一个{role}。"),
		schema.MessagesPlaceholder("history_key", false),
		schema.UserMessage("我的任务是：{task}。"),
	)

	// 准备变量
	variables := map[string]any{
		"role": "专业的翻译家",
		"task": "将下面的句子翻译成英文：'海内存知己，天涯若比邻。'",
		"history_key": []*schema.Message{
			{Role: schema.User, Content: "你好"},
			{Role: schema.Assistant, Content: "你好！有什么可以帮助你的吗？"},
		},
	}

	// 格式化模板
	formattedMessages, err := template.Format(ctx, variables)
	if err != nil {
		log.Printf("格式化模板失败: %v", err)
		return
	}

	// 显示结果
	fmt.Println("📝 格式化后的消息:")
	for i, msg := range formattedMessages {
		fmt.Printf("  消息%d [%s]: %s\n", i+1, msg.Role, msg.Content)
	}
}

// Chain编排示例
func chainTemplateExample(ctx context.Context, cm model.BaseChatModel) {
	fmt.Println("\n=== Chain编排模板示例 ===")

	// 定义聊天模板
	template := prompt.FromMessages(schema.FString,
		schema.SystemMessage("你是一个{role}"),
		schema.MessagesPlaceholder("history_key", false),
		&schema.Message{
			Role:    schema.User,
			Content: "请帮帮我，{task}",
		},
	)

	// 创建Chain
	chain := compose.NewChain[map[string]any, *schema.Message]()
	chain.AppendChatTemplate(template)
	chain.AppendChatModel(cm)

	// 编译Chain
	runnable, err := chain.Compile(ctx)
	if err != nil {
		log.Printf("编译Chain失败: %v", err)
		return
	}

	// 准备输入变量
	variables := map[string]any{
		"role": "AI助手",
		"task": "解释一下什么是聊天模板",
		"history_key": []*schema.Message{
			{Role: schema.User, Content: "你好"},
			{Role: schema.Assistant, Content: "你好！有什么可以帮助你的吗？"},
		},
	}

	// 执行Chain
	fmt.Println("🔄 执行Chain...")
	result, err := runnable.Invoke(ctx, variables)
	if err != nil {
		log.Printf("执行Chain失败: %v", err)
		return
	}

	fmt.Printf("✅ Chain执行结果: %s\n", result.Content)
}

// 复杂模板示例
func complexTemplateExample(ctx context.Context) {
	fmt.Println("\n=== 复杂模板示例 ===")

	// 创建包含多种占位符的复杂模板
	template := prompt.FromMessages(schema.FString,
		schema.SystemMessage("你是一个{role}，专门处理{domain}领域的问题。你的风格是{style}。"),
		schema.MessagesPlaceholder("conversation_history", true), // optional=true
		schema.UserMessage("当前任务：{task}"),
		schema.UserMessage("附加信息：{additional_info}"),
		schema.UserMessage("请按照以下格式回答：{format_requirements}"),
	)

	// 准备复杂变量
	variables := map[string]any{
		"role":                "高级软件架构师",
		"domain":              "微服务架构设计",
		"style":               "技术严谨且通俗易懂",
		"task":                "设计一个电商系统的微服务架构",
		"additional_info":     "系统需要支持高并发，预计日活用户100万",
		"format_requirements": "1. 架构概述 2. 核心服务列表 3. 技术选型建议",
		"conversation_history": []*schema.Message{
			{Role: schema.User, Content: "我需要设计一个新的电商系统"},
			{Role: schema.Assistant, Content: "我可以帮你设计电商系统架构，请告诉我具体需求"},
			{Role: schema.User, Content: "主要是B2C模式，需要支持大量用户"},
		},
	}

	// 格式化模板
	formattedMessages, err := template.Format(ctx, variables)
	if err != nil {
		log.Printf("格式化复杂模板失败: %v", err)
		return
	}

	// 显示结果
	fmt.Println("📝 复杂模板格式化结果:")
	for i, msg := range formattedMessages {
		content := msg.Content
		if len(content) > 100 {
			content = content[:100] + "..."
		}
		fmt.Printf("  消息%d [%s]: %s\n", i+1, msg.Role, content)
	}
}

// 条件模板示例
func conditionalTemplateExample(ctx context.Context) {
	fmt.Println("\n=== 条件模板示例 ===")

	// 演示不同条件下的模板使用
	scenarios := []struct {
		name           string
		role           string
		includeHistory bool
		task           string
	}{
		{"新用户场景", "新手引导助手", false, "帮助用户了解基本功能"},
		{"老用户场景", "高级顾问", true, "协助解决复杂技术问题"},
		{"专家场景", "技术专家", true, "提供深度技术分析"},
	}

	for _, scenario := range scenarios {
		fmt.Printf("\n🎯 场景: %s\n", scenario.name)

		// 创建模板
		template := prompt.FromMessages(schema.FString,
			schema.SystemMessage("你是一个{role}。"),
			schema.MessagesPlaceholder("history_key", !scenario.includeHistory), // 根据场景决定是否可选
			schema.UserMessage("任务：{task}"),
		)

		// 准备变量
		variables := map[string]any{
			"role": scenario.role,
			"task": scenario.task,
		}

		// 根据场景添加历史记录
		if scenario.includeHistory {
			variables["history_key"] = []*schema.Message{
				{Role: schema.User, Content: "之前我们讨论过相关技术"},
				{Role: schema.Assistant, Content: "是的，我记得我们的讨论"},
			}
		} else {
			variables["history_key"] = []*schema.Message{} // 空历史
		}

		// 格式化并显示
		messages, err := template.Format(ctx, variables)
		if err != nil {
			log.Printf("格式化场景 %s 失败: %v", scenario.name, err)
			continue
		}

		fmt.Printf("  消息数量: %d\n", len(messages))
		for i, msg := range messages {
			fmt.Printf("    %d. [%s]: %s\n", i+1, msg.Role, msg.Content)
		}
	}
}

// 模板性能测试示例
func templatePerformanceExample(ctx context.Context) {
	fmt.Println("\n=== 模板性能测试示例 ===")

	// 创建模板
	template := prompt.FromMessages(schema.FString,
		schema.SystemMessage("你是{role}"),
		schema.UserMessage("处理任务：{task}"),
	)

	// 准备测试数据
	testCount := 1000
	variables := map[string]any{
		"role": "测试助手",
		"task": "执行性能测试",
	}

	// 性能测试
	fmt.Printf("🏃 开始格式化 %d 次模板...\n", testCount)
	startTime := time.Now()

	for i := 0; i < testCount; i++ {
		_, err := template.Format(ctx, variables)
		if err != nil {
			log.Printf("第 %d 次格式化失败: %v", i+1, err)
			return
		}
	}

	duration := time.Since(startTime)
	avgTime := duration / time.Duration(testCount)

	fmt.Printf("✅ 性能测试结果:\n")
	fmt.Printf("  总耗时: %v\n", duration)
	fmt.Printf("  平均耗时: %v\n", avgTime)
	fmt.Printf("  每秒处理: %.0f 次\n", float64(testCount)/duration.Seconds())
}

// Jinja2模板格式示例
func jinja2TemplateExample(ctx context.Context) {
	fmt.Println("\n=== Jinja2模板格式示例 ===")

	// 创建 Jinja2 格式的复杂模板
	template := prompt.FromMessages(schema.Jinja2,
		schema.SystemMessage(`你是{{ role }}{% if expertise %}，专长于{{ expertise }}{% endif %}。

你的服务特点:
{% for feature in service_features %}
- {{ feature }}
{% endfor %}

{% if strict_mode %}请严格按照专业标准回答。{% else %}请以友好的方式回答。{% endif %}`),
		schema.MessagesPlaceholder("conversation_history", true),
		schema.UserMessage(`当前任务类型: {{ task_type }}

具体要求:
{% for requirement in requirements %}
{{ loop.index }}. {{ requirement }}
{% endfor %}

用户问题: {{ user_question }}`),
	)

	// 准备 Jinja2 模板变量
	variables := map[string]any{
		"role":      "AI技术顾问",
		"expertise": "云计算和微服务架构",
		"service_features": []string{
			"提供专业的技术咨询",
			"基于最佳实践的建议",
			"实用的解决方案",
			"详细的实施指导",
		},
		"strict_mode": true,
		"task_type":   "系统架构设计",
		"requirements": []string{
			"分析现有系统瓶颈",
			"设计可扩展的架构方案",
			"提供技术选型建议",
			"制定实施路线图",
		},
		"user_question": "如何设计一个支持千万级用户的电商平台架构？",
		"conversation_history": []*schema.Message{
			{Role: schema.User, Content: "我需要重新设计我们的电商平台"},
			{Role: schema.Assistant, Content: "我是专业的技术顾问，很高兴为您提供架构设计咨询。请告诉我您当前的技术挑战。"},
			{Role: schema.User, Content: "目前系统在高峰期经常出现性能问题"},
		},
	}

	// 格式化 Jinja2 模板
	formattedMessages, err := template.Format(ctx, variables)
	if err != nil {
		log.Printf("格式化Jinja2模板失败: %v", err)
		return
	}

	// 显示结果
	fmt.Println("📝 Jinja2模板格式化结果:")
	for i, msg := range formattedMessages {
		content := msg.Content
		// 对长内容进行截断显示
		lines := strings.Split(content, "\n")
		if len(lines) > 3 {
			content = strings.Join(lines[:3], "\n") + "\n  ... (更多内容已省略)"
		} else if len(content) > 200 {
			content = content[:200] + "..."
		}
		fmt.Printf("  消息%d [%s]: %s\n", i+1, msg.Role, content)
	}

	// 展示 Jinja2 特性
	fmt.Println("\n🌟 Jinja2模板特性演示:")
	fmt.Println("  ✅ 条件判断: {% if expertise %}...{% endif %}")
	fmt.Println("  ✅ 循环处理: {% for item in list %}...{% endfor %}")
	fmt.Println("  ✅ 循环计数: {{ loop.index }}")
	fmt.Println("  ✅ 变量插值: {{ variable_name }}")
	fmt.Println("  ✅ 复杂逻辑: 支持嵌套条件和循环")
}

// 主函数
func main() {
	ctx := context.Background()

	fmt.Println("🤖 Eino ChatTemplate 组件完全示例")
	fmt.Println("=====================================")

	// 1. 初始化配置
	config, err := initConfig()
	if err != nil {
		log.Fatal("配置初始化失败:", err)
	}

	fmt.Printf("使用模型: %s\n", config.Model)

	// 2. 初始化ChatModel (仅在需要时)
	var cm model.BaseChatModel
	if len(os.Args) > 1 && (os.Args[1] == "chain" || os.Args[1] == "all") {
		cm, err = initChatModel(ctx, config)
		if err != nil {
			log.Printf("ChatModel初始化失败: %v (跳过需要模型的示例)", err)
		} else {
			fmt.Println("ChatModel 初始化成功！")
		}
	}

	// 3. 运行示例
	try := func(name string, fn func(context.Context)) {
		fmt.Printf("\n正在运行: %s\n", name)
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("示例 %s 运行出错: %v\n", name, r)
			}
		}()
		fn(ctx)
	}

	tryWithModel := func(name string, fn func(context.Context, model.BaseChatModel)) {
		if cm == nil {
			fmt.Printf("\n跳过 %s (需要ChatModel)\n", name)
			return
		}
		fmt.Printf("\n正在运行: %s\n", name)
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("示例 %s 运行出错: %v\n", name, r)
			}
		}()
		fn(ctx, cm)
	}

	// 检查命令行参数
	if len(os.Args) > 1 {
		exampleName := os.Args[1]
		switch strings.ToLower(exampleName) {
		case "basic":
			try("基础模板格式化示例", basicTemplateExample)
		case "complex":
			try("复杂模板示例", complexTemplateExample)
		case "conditional":
			try("条件模板示例", conditionalTemplateExample)
		case "jinja2":
			try("Jinja2模板格式示例", jinja2TemplateExample)
		case "performance":
			try("模板性能测试示例", templatePerformanceExample)
		case "chain":
			tryWithModel("Chain编排模板示例", chainTemplateExample)
		default:
			fmt.Printf("未知示例: %s\n", exampleName)
			fmt.Println("可用示例: basic, complex, conditional, jinja2, performance, chain")
			return
		}
	} else {
		// 运行所有示例
		try("基础模板格式化示例", basicTemplateExample)
		try("复杂模板示例", complexTemplateExample)
		try("条件模板示例", conditionalTemplateExample)
		try("Jinja2模板格式示例", jinja2TemplateExample)
		try("模板性能测试示例", templatePerformanceExample)
		tryWithModel("Chain编排模板示例", chainTemplateExample)
	}

	fmt.Println("\n🎉 所有示例运行完成！")
	fmt.Println("\n使用方法:")
	fmt.Println("  go run main.go              # 运行所有示例")
	fmt.Println("  go run main.go basic        # 运行基础示例")
	fmt.Println("  go run main.go complex      # 运行复杂模板示例")
	fmt.Println("  go run main.go conditional  # 运行条件模板示例")
	fmt.Println("  go run main.go jinja2       # 运行Jinja2模板示例")
	fmt.Println("  go run main.go performance  # 运行性能测试示例")
	fmt.Println("  go run main.go chain        # 运行Chain编排示例")
}
