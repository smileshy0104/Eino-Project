package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/spf13/viper"
)

// 天气查询工具
type WeatherTool struct{}

// Info 返回工具的元数据信息
func (w *WeatherTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "get_weather",
		Desc: "查询指定城市的天气情况",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"city": {
				Type:     "string",
				Desc:     "要查询天气的城市名称，例如：北京、上海、广州等",
				Required: true,
			},
		}),
	}, nil
}

// InvokableRun 执行天气查询逻辑
func (w *WeatherTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var args struct {
		City string `json:"city"`
	}

	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("参数解析失败: %v", err)
	}

	// 模拟天气数据
	weatherData := map[string]string{
		"北京": "晴，25°C，微风",
		"上海": "多云，28°C，东南风",
		"广州": "雷阵雨，30°C，南风",
		"深圳": "晴转多云，29°C，东风",
		"杭州": "小雨，22°C，北风",
		"成都": "阴，24°C，无风",
		"西安": "晴，26°C，西北风",
		"南京": "多云，23°C，东风",
	}

	weather, exists := weatherData[args.City]
	if !exists {
		weather = "晴，25°C，微风（默认天气）"
	}

	result := map[string]interface{}{
		"city":    args.City,
		"weather": weather,
		"message": fmt.Sprintf("🌤️ %s今天的天气：%s", args.City, weather),
	}

	resultJSON, _ := json.Marshal(result)
	return string(resultJSON), nil
}

// 数学计算工具
type CalculatorTool struct{}

// Info 返回工具的元数据信息
func (c *CalculatorTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "calculate",
		Desc: "进行数学计算，支持基本的四则运算",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"expression": {
				Type:     "string",
				Desc:     "数学表达式，例如：10+5、20-8、6*7、15/3",
				Required: true,
			},
		}),
	}, nil
}

// InvokableRun 执行数学计算逻辑
func (c *CalculatorTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var args struct {
		Expression string `json:"expression"`
	}

	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("参数解析失败: %v", err)
	}

	// 简单的数学表达式计算（仅支持基本四则运算）
	resultValue, err := simpleCalculate(args.Expression)
	if err != nil {
		return "", fmt.Errorf("计算错误: %v", err)
	}

	result := map[string]interface{}{
		"expression": args.Expression,
		"result":     resultValue,
		"message":    fmt.Sprintf("🔢 计算结果：%s = %.2f", args.Expression, resultValue),
	}

	resultJSON, _ := json.Marshal(result)
	return string(resultJSON), nil
}

// 简单计算器实现
func simpleCalculate(expression string) (float64, error) {
	expression = strings.ReplaceAll(expression, " ", "")

	// 处理加法
	if strings.Contains(expression, "+") {
		parts := strings.Split(expression, "+")
		if len(parts) == 2 {
			a, err1 := strconv.ParseFloat(parts[0], 64)
			b, err2 := strconv.ParseFloat(parts[1], 64)
			if err1 == nil && err2 == nil {
				return a + b, nil
			}
		}
	}

	// 处理减法
	if strings.Contains(expression, "-") && !strings.HasPrefix(expression, "-") {
		parts := strings.Split(expression, "-")
		if len(parts) == 2 {
			a, err1 := strconv.ParseFloat(parts[0], 64)
			b, err2 := strconv.ParseFloat(parts[1], 64)
			if err1 == nil && err2 == nil {
				return a - b, nil
			}
		}
	}

	// 处理乘法
	if strings.Contains(expression, "*") {
		parts := strings.Split(expression, "*")
		if len(parts) == 2 {
			a, err1 := strconv.ParseFloat(parts[0], 64)
			b, err2 := strconv.ParseFloat(parts[1], 64)
			if err1 == nil && err2 == nil {
				return a * b, nil
			}
		}
	}

	// 处理除法
	if strings.Contains(expression, "/") {
		parts := strings.Split(expression, "/")
		if len(parts) == 2 {
			a, err1 := strconv.ParseFloat(parts[0], 64)
			b, err2 := strconv.ParseFloat(parts[1], 64)
			if err1 == nil && err2 == nil {
				if b == 0 {
					return 0, fmt.Errorf("除数不能为零")
				}
				return a / b, nil
			}
		}
	}

	// 如果是单个数字
	if num, err := strconv.ParseFloat(expression, 64); err == nil {
		return num, nil
	}

	return 0, fmt.Errorf("不支持的表达式格式")
}

// 时间查询工具
type TimeTool struct{}

// Info 返回工具的元数据信息
func (t *TimeTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "get_time",
		Desc: "获取当前的日期和时间信息",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"format": {
				Type:     "string",
				Desc:     "时间格式：date（仅日期）、time（仅时间）、datetime（日期时间）、timestamp（时间戳）",
				Required: false,
				Enum:     []string{"date", "time", "datetime", "timestamp"},
			},
		}),
	}, nil
}

// InvokableRun 执行时间查询逻辑
func (t *TimeTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var args struct {
		Format string `json:"format"`
	}

	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("参数解析失败: %v", err)
	}

	if args.Format == "" {
		args.Format = "datetime" // 默认格式
	}

	now := time.Now()
	var message string
	var value interface{}

	switch args.Format {
	case "date":
		value = now.Format("2006年01月02日")
		message = fmt.Sprintf("📅 今天是：%s", value)
	case "time":
		value = now.Format("15:04:05")
		message = fmt.Sprintf("⏰ 当前时间：%s", value)
	case "datetime":
		value = now.Format("2006年01月02日 15:04:05")
		message = fmt.Sprintf("📅⏰ 当前日期时间：%s", value)
	case "timestamp":
		value = now.Unix()
		message = fmt.Sprintf("🕐 时间戳：%d", value)
	default:
		value = now.Format("2006年01月02日 15:04:05")
		message = fmt.Sprintf("📅⏰ 当前日期时间：%s", value)
	}

	result := map[string]interface{}{
		"format":  args.Format,
		"value":   value,
		"message": message,
	}

	resultJSON, _ := json.Marshal(result)
	return string(resultJSON), nil
}

// 消息修饰器：为消息添加系统提示
func messageModifier(ctx context.Context, msgs []*schema.Message) ([]*schema.Message, error) {
	systemPrompt := &schema.Message{
		Role: schema.System,
		Content: `你是一个智能助手，名字叫小艾诺（Eino）。你拥有以下能力：

1. 🌤️ 天气查询：可以查询中国主要城市的天气情况
2. 🔢 数学计算：可以进行基本的四则运算
3. ⏰ 时间查询：可以获取当前的日期和时间信息

请根据用户的问题，合理使用工具来提供准确的答案。如果用户的问题需要使用多个工具，请按步骤逐一使用。

注意：
- 回答要简洁明了
- 对于天气查询，目前只支持北京、上海、广州、深圳、杭州、成都、西安、南京等城市
- 对于数学计算，目前只支持简单的四则运算（+、-、*、/）
- 始终保持友好和乐于助人的态度`,
	}

	// 如果第一条消息不是系统消息，则添加系统消息
	if len(msgs) == 0 || msgs[0].Role != schema.System {
		return append([]*schema.Message{systemPrompt}, msgs...), nil
	}

	return msgs, nil
}

func main() {
	// 初始化配置
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("读取配置文件失败: %v", err)
	}

	// 设置环境变量
	os.Setenv("ARK_API_KEY", viper.GetString("ARK_API_KEY"))

	ctx := context.Background()

	// 创建 ARK 聊天模型
	config := &ark.ChatModelConfig{
		Model:  viper.GetString("ARK_MODEL"),
		APIKey: viper.GetString("ARK_API_KEY"),
	}

	chatModel, err := ark.NewChatModel(ctx, config)
	if err != nil {
		log.Fatalf("创建聊天模型失败: %v", err)
	}

	// 创建工具列表
	tools := []tool.InvokableTool{
		&WeatherTool{},
		&CalculatorTool{},
		&TimeTool{},
	}

	// 绑定工具到聊天模型
	toolInfos := make([]*schema.ToolInfo, 0, len(tools))
	for _, tool := range tools {
		info, err := tool.Info(ctx)
		if err != nil {
			log.Printf("获取工具信息失败: %v", err)
			continue
		}
		toolInfos = append(toolInfos, info)
		log.Printf("绑定工具: %s - %s", info.Name, info.Desc)
	}

	chatModel.BindTools(toolInfos)

	// 创建简单的Chain Agent - 仅使用聊天模型测试
	chain := compose.NewChain[[]*schema.Message, *schema.Message]()
	chain.AppendChatModel(chatModel, compose.WithNodeName("chat_model"))

	// 编译Chain
	compiledChain, err := chain.Compile(ctx)
	if err != nil {
		log.Fatalf("编译链失败: %v", err)
	}

	fmt.Println("🤖 小艾诺（Eino）Chain Agent Demo 启动成功！")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println("我可以帮助您：")
	fmt.Println("🌤️  查询天气：例如 '北京今天天气怎么样？'")
	fmt.Println("🔢 数学计算：例如 '计算 25 + 17'")
	fmt.Println("⏰ 时间查询：例如 '现在几点了？'")
	fmt.Println("🔄 组合操作：例如 '查询上海天气，然后告诉我现在时间'")
	fmt.Println(strings.Repeat("=", 50))

	// 测试用例
	testCases := []string{
		"北京今天天气怎么样？",
		"帮我计算 125 + 237",
		"现在几点了？",
		"上海的天气如何，另外帮我算一下 15 * 8",
		"查询深圳天气，然后告诉我今天的日期",
	}

	for i, testCase := range testCases {
		fmt.Printf("\n📝 测试用例 %d: %s\n", i+1, testCase)
		fmt.Println(strings.Repeat("-", 40))

		// 添加系统提示
		systemMessage := &schema.Message{
			Role: schema.System,
			Content: `你是一个智能助手，名字叫小艾诺（Eino）。你拥有以下能力：

1. 🌤️ 天气查询：可以查询中国主要城市的天气情况
2. 🔢 数学计算：可以进行基本的四则运算
3. ⏰ 时间查询：可以获取当前的日期和时间信息

请根据用户的问题，合理使用工具来提供准确的答案。回答要简洁明了，保持友好和乐于助人的态度。`,
		}

		// 创建用户消息
		userMessage := &schema.Message{
			Role:    schema.User,
			Content: testCase,
		}

		// 执行链式 Agent
		messages := []*schema.Message{systemMessage, userMessage}
		result, err := compiledChain.Invoke(ctx, messages)
		if err != nil {
			fmt.Printf("❌ 执行失败: %v\n", err)
			continue
		}

		// 输出结果 - result 是 *schema.Message
		if result != nil && result.Content != "" {
			fmt.Printf("🤖 小艾诺: %s\n", result.Content)
		} else {
			fmt.Printf("🤖 小艾诺: [没有收到回复内容]\n")
			fmt.Printf("Debug: result=%+v\n", result)
		}

		// 等待一秒，让输出更清晰
		time.Sleep(1 * time.Second)
	}

	fmt.Println("\n✅ 所有测试用例执行完成！")
	fmt.Println("\n💡 您可以修改 testCases 数组来测试更多场景。")
	fmt.Println("📚 如需了解更多功能，请参考 README.md 文档。")
}
