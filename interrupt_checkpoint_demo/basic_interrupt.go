package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/spf13/viper"
)

// 定义节点的输入输出结构
type StepInput struct {
	Content string `json:"content"`
	Step    int    `json:"step"`
}

type StepOutput struct {
	Result string `json:"result"`
	Step   int    `json:"step"`
}

// 基础中断错误
type BasicInterruptError struct {
	Message       string
	InterruptNode string
	Metadata      map[string]interface{}
}

func (e *BasicInterruptError) Error() string {
	return fmt.Sprintf("中断于节点 %s: %s", e.InterruptNode, e.Message)
}

func NewBasicInterruptError(message, node string) *BasicInterruptError {
	return &BasicInterruptError{
		Message:       message,
		InterruptNode: node,
		Metadata:      make(map[string]interface{}),
	}
}

// 创建处理节点函数
type ProcessingNode func(ctx context.Context, input StepInput) (StepOutput, error)

func createProcessingNode(stepName string, stepNum int) ProcessingNode {
	return func(ctx context.Context, input StepInput) (StepOutput, error) {
		fmt.Printf("🔄 执行 %s (步骤 %d): %s\n", stepName, stepNum, input.Content)

		// 模拟处理时间
		time.Sleep(300 * time.Millisecond)
		result := fmt.Sprintf("已完成%s处理: %s", stepName, input.Content)

		return StepOutput{
			Result: result,
			Step:   stepNum,
		}, nil
	}
}

// 工作流执行器，支持中断配置
type BasicWorkflowExecutor struct {
	nodes             map[string]ProcessingNode
	interruptAfter    []string
	interruptBefore   []string
	executionSequence []string
}

func NewBasicWorkflowExecutor() *BasicWorkflowExecutor {
	return &BasicWorkflowExecutor{
		nodes:           make(map[string]ProcessingNode),
		interruptAfter:  []string{},
		interruptBefore: []string{},
	}
}

func (we *BasicWorkflowExecutor) AddNode(name string, node ProcessingNode) {
	we.nodes[name] = node
}

func (we *BasicWorkflowExecutor) SetExecutionSequence(sequence []string) {
	we.executionSequence = sequence
}

func (we *BasicWorkflowExecutor) WithInterruptAfterNodes(nodes []string) *BasicWorkflowExecutor {
	we.interruptAfter = nodes
	return we
}

func (we *BasicWorkflowExecutor) WithInterruptBeforeNodes(nodes []string) *BasicWorkflowExecutor {
	we.interruptBefore = nodes
	return we
}

func (we *BasicWorkflowExecutor) contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func (we *BasicWorkflowExecutor) Execute(ctx context.Context, input StepInput) (StepOutput, error) {
	currentInput := input
	var currentOutput StepOutput

	for _, nodeName := range we.executionSequence {
		node, exists := we.nodes[nodeName]
		if !exists {
			return StepOutput{}, fmt.Errorf("节点 %s 不存在", nodeName)
		}

		// 检查节点前中断
		if we.contains(we.interruptBefore, nodeName) {
			fmt.Printf("⏸️  在节点 '%s' 前中断\n", nodeName)
			return currentOutput, NewBasicInterruptError("节点前中断", nodeName)
		}

		// 执行节点
		output, err := node(ctx, currentInput)
		if err != nil {
			return output, err
		}

		currentOutput = output
		currentInput = StepInput{
			Content: output.Result,
			Step:    output.Step,
		}

		// 检查节点后中断
		if we.contains(we.interruptAfter, nodeName) {
			fmt.Printf("⏸️  在节点 '%s' 后中断\n", nodeName)
			return currentOutput, NewBasicInterruptError("节点后中断", nodeName)
		}
	}

	return currentOutput, nil
}

// 基础中断演示
func basicInterruptDemo(ctx context.Context) {
	fmt.Println("=== 基础中断演示 ===")

	// 创建工作流执行器
	executor := NewBasicWorkflowExecutor()

	// 添加节点
	executor.AddNode("step1", createProcessingNode("数据预处理", 1))
	executor.AddNode("step2", createProcessingNode("数据分析", 2))
	executor.AddNode("step3", createProcessingNode("结果生成", 3))

	// 设置执行顺序
	executor.SetExecutionSequence([]string{"step1", "step2", "step3"})

	// 配置在 step2 之后中断
	executor.WithInterruptAfterNodes([]string{"step2"})

	// 执行工作流
	input := StepInput{
		Content: "测试数据",
		Step:    1,
	}

	fmt.Println("📝 开始执行工作流...")
	result, err := executor.Execute(ctx, input)

	// 检查是否发生中断
	if err != nil {
		if interruptErr, ok := err.(*BasicInterruptError); ok {
			fmt.Printf("⏸️  工作流在节点 '%s' 中断\n", interruptErr.InterruptNode)
			fmt.Printf("📊 中断时的状态: %+v\n", result)

			// 演示恢复执行（创建新的执行器，不设置中断）
			fmt.Println("🔄 恢复执行...")

			resumeExecutor := NewBasicWorkflowExecutor()
			resumeExecutor.AddNode("step1", createProcessingNode("数据预处理", 1))
			resumeExecutor.AddNode("step2", createProcessingNode("数据分析", 2))
			resumeExecutor.AddNode("step3", createProcessingNode("结果生成", 3))
			resumeExecutor.SetExecutionSequence([]string{"step1", "step2", "step3"})
			// 不设置中断点

			finalResult, err := resumeExecutor.Execute(ctx, input)
			if err != nil {
				log.Printf("恢复执行失败: %v", err)
			} else {
				fmt.Printf("✅ 最终结果: %+v\n", finalResult)
			}
		} else {
			log.Printf("执行工作流失败: %v", err)
		}
	} else {
		fmt.Printf("✅ 直接完成，结果: %+v\n", result)
	}
}

// 节点前中断演示
func beforeNodeInterruptDemo(ctx context.Context) {
	fmt.Println("\n=== 节点前中断演示 ===")

	// 创建工作流执行器
	executor := NewBasicWorkflowExecutor()
	executor.AddNode("preparation", createProcessingNode("准备阶段", 1))
	executor.AddNode("critical", createProcessingNode("关键处理", 2))
	executor.AddNode("cleanup", createProcessingNode("清理阶段", 3))

	executor.SetExecutionSequence([]string{"preparation", "critical", "cleanup"})

	// 配置在关键节点前中断
	executor.WithInterruptBeforeNodes([]string{"critical"})

	input := StepInput{
		Content: "重要数据处理",
		Step:    1,
	}

	fmt.Println("📝 开始执行工作流...")
	result, err := executor.Execute(ctx, input)

	if err != nil {
		if interruptErr, ok := err.(*BasicInterruptError); ok {
			fmt.Printf("⏸️  工作流在节点 '%s' 前中断\n", interruptErr.InterruptNode)
			fmt.Printf("📊 中断时的状态: %+v\n", result)

			// 在实际应用中，这里可能需要人工确认或额外的验证
			fmt.Println("🔍 进行人工确认...")
			time.Sleep(1 * time.Second)
			fmt.Println("✅ 确认通过，继续执行")

			// 继续执行
			resumeExecutor := NewBasicWorkflowExecutor()
			resumeExecutor.AddNode("preparation", createProcessingNode("准备阶段", 1))
			resumeExecutor.AddNode("critical", createProcessingNode("关键处理", 2))
			resumeExecutor.AddNode("cleanup", createProcessingNode("清理阶段", 3))
			resumeExecutor.SetExecutionSequence([]string{"preparation", "critical", "cleanup"})

			finalResult, err := resumeExecutor.Execute(ctx, input)
			if err != nil {
				log.Printf("恢复执行失败: %v", err)
			} else {
				fmt.Printf("✅ 最终结果: %+v\n", finalResult)
			}
		} else {
			log.Printf("执行工作流失败: %v", err)
		}
	}
}

// 多点中断演示
func multiPointInterruptDemo(ctx context.Context) {
	fmt.Println("\n=== 多点中断演示 ===")

	// 创建工作流执行器
	executor := NewBasicWorkflowExecutor()

	nodeNames := []string{"input", "validate", "process", "output"}
	for i, name := range nodeNames {
		executor.AddNode(name, createProcessingNode(fmt.Sprintf("%s阶段", name), i+1))
	}

	executor.SetExecutionSequence(nodeNames)

	// 配置多个中断点
	executor.WithInterruptAfterNodes([]string{"validate"}).
		WithInterruptBeforeNodes([]string{"output"})

	input := StepInput{
		Content: "多阶段处理数据",
		Step:    1,
	}

	fmt.Println("📝 开始执行多点中断工作流...")

	// 模拟多次执行和恢复
	currentInput := input
	interruptCount := 0
	maxInterrupts := 3

	for interruptCount < maxInterrupts {
		// 根据已经发生的中断次数调整执行器
		testExecutor := NewBasicWorkflowExecutor()
		for i, name := range nodeNames {
			testExecutor.AddNode(name, createProcessingNode(fmt.Sprintf("%s阶段", name), i+1))
		}
		testExecutor.SetExecutionSequence(nodeNames)

		// 根据中断次数设置不同的中断点
		if interruptCount == 0 {
			testExecutor.WithInterruptAfterNodes([]string{"validate"})
		} else if interruptCount == 1 {
			testExecutor.WithInterruptBeforeNodes([]string{"output"})
		}
		// 第三次不设置中断点

		result, err := testExecutor.Execute(ctx, currentInput)

		if err != nil {
			if interruptErr, ok := err.(*BasicInterruptError); ok {
				interruptCount++
				fmt.Printf("⏸️  第 %d 次中断在: %s\n", interruptCount, interruptErr.InterruptNode)
				fmt.Printf("📊 中断时的状态: %+v\n", result)

				// 更新输入为中断时的状态
				if result.Result != "" {
					currentInput = StepInput{
						Content: result.Result,
						Step:    result.Step,
					}
				}

				// 模拟处理中断
				fmt.Printf("🔧 处理第 %d 次中断...\n", interruptCount)
				time.Sleep(800 * time.Millisecond)
				fmt.Println("✅ 中断处理完成，继续执行")

				continue
			} else {
				log.Printf("执行失败: %v", err)
				break
			}
		} else {
			// 成功完成
			fmt.Printf("✅ 所有中断点处理完成，最终结果: %+v\n", result)
			break
		}
	}

	if interruptCount >= maxInterrupts {
		fmt.Println("⚠️  达到最大中断处理次数")
	}
}

func initBasicConfig() {
	viper.SetConfigFile("../config.yaml")
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("读取配置文件失败: %v (使用默认配置)", err)
	}
}

func main() {
	initBasicConfig()
	ctx := context.Background()

	// 运行各种中断演示
	basicInterruptDemo(ctx)
	beforeNodeInterruptDemo(ctx)
	multiPointInterruptDemo(ctx)

	fmt.Println("\n🎉 基础中断演示完成！")
}
