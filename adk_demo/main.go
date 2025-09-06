package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// ============= 基础 Eino ADK 演示 =============
// 简化版本，展示最基本的工具接口实现

// 简化的计算器工具（基础演示版本）
type BasicCalculatorTool struct{}

func (c *BasicCalculatorTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	params := map[string]*schema.ParameterInfo{
		"expression": {
			Type:     schema.String,
			Desc:     "简单的数学表达式，如 '25+17'",
			Required: true,
		},
	}

	return &schema.ToolInfo{
		Name:        "simple_calculator",
		Desc:        "执行简单的数学计算",
		ParamsOneOf: schema.NewParamsOneOfByParams(params),
	}, nil
}

func (c *BasicCalculatorTool) InvokableRun(ctx context.Context, paramsInJSON string, opts ...tool.Option) (string, error) {
	var params map[string]interface{}
	if err := json.Unmarshal([]byte(paramsInJSON), &params); err != nil {
		return "", fmt.Errorf("解析参数失败: %w", err)
	}

	expression, ok := params["expression"].(string)
	if !ok {
		return "", fmt.Errorf("expression 参数必须是字符串类型")
	}

	// 极简的计算逻辑
	expression = strings.ReplaceAll(expression, " ", "")

	var result float64
	var operation string

	if strings.Contains(expression, "+") {
		parts := strings.Split(expression, "+")
		operation = "加法"
		if len(parts) == 2 {
			var a, b float64
			fmt.Sscanf(parts[0], "%f", &a)
			fmt.Sscanf(parts[1], "%f", &b)
			result = a + b
		}
	} else {
		return "", fmt.Errorf("目前仅支持加法运算")
	}

	resultJSON, _ := json.Marshal(map[string]interface{}{
		"operation":  operation,
		"expression": expression,
		"result":     result,
		"message":    fmt.Sprintf("计算结果: %s = %.2f", expression, result),
	})

	return string(resultJSON), nil
}

// 演示函数
func demonstrateBasicTool() {
	fmt.Println("🎯 基础工具演示")
	fmt.Println(strings.Repeat("=", 50))

	ctx := context.Background()
	calculator := &BasicCalculatorTool{}

	// 获取工具信息
	info, err := calculator.Info(ctx)
	if err != nil {
		fmt.Printf("❌ 获取工具信息失败: %v\n", err)
		return
	}

	fmt.Printf("🔧 工具名称: %s\n", info.Name)
	fmt.Printf("📝 工具描述: %s\n", info.Desc)
	fmt.Println()

	// 测试用例
	testCases := []struct {
		name       string
		paramsJSON string
		desc       string
	}{
		{
			name:       "简单加法",
			paramsJSON: `{"expression": "10 + 5"}`,
			desc:       "测试基本加法运算",
		},
		{
			name:       "大数加法",
			paramsJSON: `{"expression": "123 + 456"}`,
			desc:       "测试较大数值的加法运算",
		},
		{
			name:       "小数加法",
			paramsJSON: `{"expression": "3.14 + 2.86"}`,
			desc:       "测试小数加法运算",
		},
	}

	for i, testCase := range testCases {
		fmt.Printf("📋 测试用例 %d: %s\n", i+1, testCase.name)
		fmt.Printf("📄 说明: %s\n", testCase.desc)
		fmt.Printf("📥 输入参数: %s\n", testCase.paramsJSON)

		result, err := calculator.InvokableRun(ctx, testCase.paramsJSON)
		if err != nil {
			fmt.Printf("❌ 执行失败: %v\n", err)
		} else {
			fmt.Printf("📤 执行结果: %s\n", result)
		}

		fmt.Println(strings.Repeat("-", 40))
		time.Sleep(500 * time.Millisecond)
	}
}

func demonstrateAgentConcepts() {
	fmt.Println("🤖 Agent 概念演示")
	fmt.Println(strings.Repeat("=", 50))

	fmt.Println("📚 ADK (Agent Development Kit) 核心概念:")
	fmt.Println("  • 统一接口: 所有工具都实现 tool.InvokableTool 接口")
	fmt.Println("  • 标准化: Info() 方法描述工具能力")
	fmt.Println("  • JSON 交互: InvokableRun() 处理结构化数据")
	fmt.Println("  • 可组合性: 多个工具可以组合成复杂 Agent")

	fmt.Println("\n🎯 工具到 Agent 的演进:")
	fmt.Println("  1. 工具 (Tool): 单一功能，如计算器")
	fmt.Println("  2. 智能体 (Agent): 多工具组合 + 推理能力")
	fmt.Println("  3. 工作流 (Workflow): 多智能体协作")

	fmt.Println("\n💡 Eino ADK 的价值:")
	fmt.Println("  ✅ 快速开发: 标准化接口减少重复工作")
	fmt.Println("  ✅ 易于维护: 组件化设计，职责清晰")
	fmt.Println("  ✅ 高度复用: 工具可在多个 Agent 间复用")
	fmt.Println("  ✅ 扩展性强: 新工具可无缝集成")
}

func main() {
	fmt.Println("🎊 Eino ADK 基础演示")
	fmt.Println("展示工具接口的核心概念和基本用法")
	fmt.Println(strings.Repeat("=", 60))

	// 基础工具演示
	demonstrateBasicTool()

	time.Sleep(time.Second)

	// Agent 概念演示
	demonstrateAgentConcepts()

	// 总结
	fmt.Println("\n🎯 演示总结")
	fmt.Println(strings.Repeat("=", 60))

	fmt.Println("✨ 本演示展示了:")
	fmt.Println("  🔧 tool.InvokableTool 接口的基本实现")
	fmt.Println("  📊 JSON 参数和结果的标准化处理")
	fmt.Println("  🎯 从简单工具到复杂 Agent 的发展路径")

	fmt.Println("\n📚 学习建议:")
	fmt.Println("  1. 🌟 推荐运行: go run corrected_official_demo.go")
	fmt.Println("     (完整的工具实现，包含计算器和天气查询)")
	fmt.Println("  2. 🌟 推荐运行: go run stable_extension_demo.go")
	fmt.Println("     (中断与恢复机制，展示企业级可靠性)")
	fmt.Println("  3. 📖 阅读文档: Eino_ADK_Guide.md")
	fmt.Println("     (深入理解 ADK 架构和最佳实践)")

	fmt.Println("\n🎉 开始你的 AI Agent 开发之旅！")
}
