// basic_interrupt.go - Eino 框架基础中断功能演示
// 本文件演示了如何在工作流执行过程中实现基础的中断机制
// 主要功能包括：
// 1. 节点前中断 - 在特定节点执行前暂停工作流
// 2. 节点后中断 - 在特定节点执行后暂停工作流
// 3. 多点中断 - 在多个节点设置中断点
// 4. 中断恢复 - 从中断点继续执行工作流
// 这些功能在需要人工干预、状态检查或分阶段执行的场景中非常有用
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/spf13/viper" // 配置管理库
)

// StepInput 工作流步骤的输入数据结构
// 定义了每个处理节点接收的标准输入格式
type StepInput struct {
	Content string `json:"content"` // 处理内容，包含需要处理的数据
	Step    int    `json:"step"`    // 步骤编号，用于跟踪执行进度
}

// StepOutput 工作流步骤的输出数据结构
// 定义了每个处理节点产生的标准输出格式
type StepOutput struct {
	Result string `json:"result"` // 处理结果，包含节点处理后的数据
	Step   int    `json:"step"`   // 步骤编号，标识当前完成的步骤
}

// BasicInterruptError 基础中断错误类型
// 用于表示工作流执行过程中的中断状态，包含中断的详细信息
// 实现了 error 接口，可以作为标准错误类型使用
type BasicInterruptError struct {
	Message       string                 // 中断消息，描述中断的原因或类型
	InterruptNode string                 // 中断节点名称，标识在哪个节点发生中断
	Metadata      map[string]interface{} // 元数据，存储中断时的额外信息
}

// Error 实现 error 接口的 Error 方法
// 返回格式化的错误信息，包含中断节点和消息
// 返回值: 格式化的错误描述字符串
func (e *BasicInterruptError) Error() string {
	return fmt.Sprintf("中断于节点 %s: %s", e.InterruptNode, e.Message)
}

// NewBasicInterruptError 创建新的基础中断错误实例
// 这是 BasicInterruptError 的构造函数，用于创建标准化的中断错误
// 参数:
//   - message: 中断消息，描述中断的具体原因
//   - node: 中断节点名称，标识发生中断的节点
//
// 返回值: 初始化完成的 BasicInterruptError 指针
func NewBasicInterruptError(message, node string) *BasicInterruptError {
	return &BasicInterruptError{
		Message:       message,                      // 设置中断消息
		InterruptNode: node,                         // 设置中断节点
		Metadata:      make(map[string]interface{}), // 初始化空的元数据映射
	}
}

// ProcessingNode 处理节点函数类型定义
// 定义了工作流中每个处理节点的标准函数签名
// 所有的处理节点都必须符合这个函数类型，确保工作流的一致性
// 参数:
//   - ctx: 上下文对象，用于控制节点执行的生命周期
//   - input: 节点输入数据，包含需要处理的内容和步骤信息
//
// 返回值:
//   - StepOutput: 节点输出数据，包含处理结果和步骤信息
//   - error: 执行过程中的错误信息
type ProcessingNode func(ctx context.Context, input StepInput) (StepOutput, error)

// createProcessingNode 创建标准化的处理节点函数
// 这是一个工厂函数，用于生成具有统一行为的处理节点
// 生成的节点会模拟实际的处理过程，包括执行时间和结果生成
// 参数:
//   - stepName: 步骤名称，用于标识和日志输出
//   - stepNum: 步骤编号，用于跟踪执行顺序
//
// 返回值: 符合 ProcessingNode 类型的处理函数
func createProcessingNode(stepName string, stepNum int) ProcessingNode {
	return func(ctx context.Context, input StepInput) (StepOutput, error) {
		// 输出节点执行开始的日志信息
		fmt.Printf("🔄 执行 %s (步骤 %d): %s\n", stepName, stepNum, input.Content)

		// 模拟实际的处理时间 - 在真实场景中这里会是具体的业务逻辑
		time.Sleep(300 * time.Millisecond)

		// 生成处理结果 - 模拟节点处理后的数据
		result := fmt.Sprintf("已完成%s处理: %s", stepName, input.Content)

		// 返回标准化的输出结果
		return StepOutput{
			Result: result,  // 处理后的结果数据
			Step:   stepNum, // 当前步骤编号
		}, nil
	}
}

// demonstrateBasicWorkflow 演示基础工作流的完整执行过程
// 创建一个包含三个步骤的简单工作流，展示正常的顺序执行流程
// 这个演示展示了工作流执行器在没有中断的情况下的标准行为
func demonstrateBasicWorkflow() {
	fmt.Println("\n=== 基础工作流演示 ===")

	// 创建工作流执行器实例
	executor := NewBasicWorkflowExecutor()

	// 添加三个处理节点，模拟典型的数据处理流程
	executor.AddNode("step1", createProcessingNode("数据预处理", 1)) // 第一步：数据预处理
	executor.AddNode("step2", createProcessingNode("数据分析", 2))  // 第二步：数据分析
	executor.AddNode("step3", createProcessingNode("结果生成", 3))  // 第三步：结果生成

	// 设置节点的执行顺序
	executor.SetExecutionSequence([]string{"step1", "step2", "step3"})

	// 准备初始输入数据
	input := StepInput{
		Content: "原始数据", // 输入内容
		Step:    1,      // 起始步骤编号
	}

	fmt.Println("📝 开始执行工作流...")
	// 执行完整的工作流
	result, err := executor.Execute(context.Background(), input)
	if err != nil {
		// 处理执行错误
		fmt.Printf("执行失败: %v\n", err)
		return
	}

	// 输出最终执行结果
	fmt.Printf("✅ 工作流完成，最终结果: %s\n", result.Result)
}

// BasicWorkflowExecutor 基础工作流执行器
// 支持中断配置的工作流执行引擎，是中断功能演示的核心组件
// 主要功能:
// 1. 管理工作流中的所有处理节点
// 2. 控制节点的执行顺序
// 3. 支持在指定节点前后设置中断点
// 4. 处理中断逻辑并保存执行状态
type BasicWorkflowExecutor struct {
	nodes             map[string]ProcessingNode // 节点映射表，存储所有注册的处理节点
	interruptAfter    []string                  // 节点后中断列表，指定在哪些节点执行后中断
	interruptBefore   []string                  // 节点前中断列表，指定在哪些节点执行前中断
	executionSequence []string                  // 执行序列，定义节点的执行顺序
}

// NewBasicWorkflowExecutor 创建新的基础工作流执行器实例
// 这是 BasicWorkflowExecutor 的构造函数，初始化所有必要的字段
// 返回值: 初始化完成的工作流执行器指针
func NewBasicWorkflowExecutor() *BasicWorkflowExecutor {
	return &BasicWorkflowExecutor{
		nodes:           make(map[string]ProcessingNode), // 初始化空的节点映射表
		interruptAfter:  []string{},                      // 初始化空的节点后中断列表
		interruptBefore: []string{},                      // 初始化空的节点前中断列表
	}
}

// AddNode 向工作流执行器添加处理节点
// 将指定的处理节点注册到执行器中，供后续执行时调用
// 参数:
//   - name: 节点名称，用于标识和引用节点
//   - node: 处理节点函数，实现具体的处理逻辑
func (we *BasicWorkflowExecutor) AddNode(name string, node ProcessingNode) {
	we.nodes[name] = node // 将节点添加到节点映射表中
}

// SetExecutionSequence 设置工作流的执行顺序
// 定义节点的执行序列，决定工作流的处理流程
// 参数:
//   - sequence: 节点名称序列，按执行顺序排列
func (we *BasicWorkflowExecutor) SetExecutionSequence(sequence []string) {
	we.executionSequence = sequence // 设置节点执行顺序
}

// WithInterruptAfterNodes 配置节点后中断点
// 设置在指定节点执行完成后触发中断的节点列表
// 这种中断方式适用于需要在节点完成后进行状态检查或人工确认的场景
// 参数:
//   - nodes: 需要在执行后中断的节点名称列表
//
// 返回值: 返回执行器自身，支持链式调用
func (we *BasicWorkflowExecutor) WithInterruptAfterNodes(nodes []string) *BasicWorkflowExecutor {
	we.interruptAfter = nodes // 设置节点后中断列表
	return we                 // 返回自身支持链式调用
}

// WithInterruptBeforeNodes 配置节点前中断点
// 设置在指定节点执行前触发中断的节点列表
// 这种中断方式适用于需要在关键节点执行前进行准备工作或权限检查的场景
// 参数:
//   - nodes: 需要在执行前中断的节点名称列表
//
// 返回值: 返回执行器自身，支持链式调用
func (we *BasicWorkflowExecutor) WithInterruptBeforeNodes(nodes []string) *BasicWorkflowExecutor {
	we.interruptBefore = nodes // 设置节点前中断列表
	return we                  // 返回自身支持链式调用
}

// contains 检查字符串切片是否包含指定项
// 这是一个辅助方法，用于判断某个节点是否在中断列表中
// 参数:
//   - slice: 要检查的字符串切片
//   - item: 要查找的字符串项
//
// 返回值: 如果找到则返回 true，否则返回 false
func (we *BasicWorkflowExecutor) contains(slice []string, item string) bool {
	// 遍历切片中的每个元素
	for _, s := range slice {
		if s == item {
			return true // 找到匹配项，返回 true
		}
	}
	return false // 未找到匹配项，返回 false
}

// TODO Execute 执行工作流的核心方法
// 按照预定义的执行序列依次执行各个节点，并在配置的中断点处暂停执行
// 这是工作流执行器的核心功能，实现了完整的中断控制逻辑
// 参数:
//   - ctx: 上下文对象，用于控制执行过程的生命周期
//   - input: 工作流的初始输入数据
//
// 返回值:
//   - StepOutput: 执行结果或中断时的状态数据
//   - error: 执行错误或中断错误
func (we *BasicWorkflowExecutor) Execute(ctx context.Context, input StepInput) (StepOutput, error) {
	// 初始化执行状态
	currentInput := input        // 当前节点的输入数据
	var currentOutput StepOutput // 当前节点的输出数据

	// 按照执行序列依次处理每个节点
	for _, nodeName := range we.executionSequence {
		// 检查节点是否存在
		node, exists := we.nodes[nodeName]
		if !exists {
			// 节点不存在时返回错误
			return StepOutput{}, fmt.Errorf("节点 %s 不存在", nodeName)
		}

		// 第一阶段：检查节点前中断
		// 在节点执行前检查是否需要中断，适用于需要预先验证或准备的场景
		if we.contains(we.interruptBefore, nodeName) {
			fmt.Printf("⏸️  在节点 '%s' 前中断\n", nodeName)
			// 返回中断错误，保留当前执行状态
			return currentOutput, NewBasicInterruptError("节点前中断", nodeName)
		}

		// TODO 第二阶段：执行当前节点函数
		// 调用节点的处理函数，执行具体的业务逻辑
		output, err := node(ctx, currentInput)
		if err != nil {
			// 节点执行失败时直接返回错误
			return output, err
		}

		// 更新执行状态
		currentOutput = output // 保存当前节点的输出结果
		// 将当前节点的输出作为下一个节点的输入
		currentInput = StepInput{
			Content: output.Result, // 使用处理结果作为下一步的输入内容
			Step:    output.Step,   // 更新步骤编号
		}

		// 第三阶段：检查节点后中断
		// 在节点执行后检查是否需要中断，适用于需要状态检查或人工确认的场景
		if we.contains(we.interruptAfter, nodeName) {
			fmt.Printf("⏸️  在节点 '%s' 后中断\n", nodeName)
			// 返回中断错误，保留当前执行状态（包含已完成节点的结果）
			return currentOutput, NewBasicInterruptError("节点后中断", nodeName)
		}
	}

	// 所有节点执行完成，返回最终结果
	return currentOutput, nil
}

// basicInterruptDemo 基础中断功能演示
// 演示如何在工作流执行过程中设置中断点，以及如何处理中断和恢复执行
// 主要展示节点后中断的基本用法和中断恢复机制
// 参数:
//   - ctx: 上下文对象，用于控制演示过程的生命周期
func basicInterruptDemo(ctx context.Context) {
	fmt.Println("=== 基础中断演示 ===")

	// 创建工作流执行器实例
	executor := NewBasicWorkflowExecutor()

	// 添加三个处理节点，构建完整的数据处理流程
	executor.AddNode("step1", createProcessingNode("数据预处理", 1)) // 第一步：数据预处理
	executor.AddNode("step2", createProcessingNode("数据分析", 2))  // 第二步：数据分析
	executor.AddNode("step3", createProcessingNode("结果生成", 3))  // 第三步：结果生成

	// 设置节点的执行顺序，定义工作流的处理流程
	executor.SetExecutionSequence([]string{"step1", "step2", "step3"})

	// 配置中断点：在 step2 执行完成后触发中断
	// 这种配置适用于需要在数据分析完成后进行人工检查的场景
	executor.WithInterruptAfterNodes([]string{"step2"})

	// 准备工作流的初始输入数据
	input := StepInput{
		Content: "测试数据", // 输入的测试数据
		Step:    1,      // 起始步骤编号
	}

	// 开始执行工作流
	fmt.Println("📝 开始执行工作流...")
	result, err := executor.Execute(ctx, input)

	// 检查执行结果，判断是否发生了预期的中断
	if err != nil {
		// 尝试将错误转换为中断错误类型
		if interruptErr, ok := err.(*BasicInterruptError); ok {
			// 成功捕获中断错误，输出中断信息
			fmt.Printf("⏸️  工作流在节点 '%s' 中断\n", interruptErr.InterruptNode)
			fmt.Printf("📊 中断时的状态: %+v\n", result)

			// 演示中断恢复机制：创建新的执行器继续执行剩余步骤
			fmt.Println("🔄 恢复执行...")

			// 创建用于恢复执行的新执行器
			resumeExecutor := NewBasicWorkflowExecutor()
			// 重新添加所有节点（在实际应用中可能只需要添加剩余节点）
			resumeExecutor.AddNode("step1", createProcessingNode("数据预处理", 1))
			resumeExecutor.AddNode("step2", createProcessingNode("数据分析", 2))
			resumeExecutor.AddNode("step3", createProcessingNode("结果生成", 3))
			resumeExecutor.SetExecutionSequence([]string{"step1", "step2", "step3"})
			// 注意：恢复执行时不设置中断点，确保能够完整执行

			// 执行完整的工作流（从头开始，实际应用中可能从中断点继续）
			finalResult, err := resumeExecutor.Execute(ctx, input)
			if err != nil {
				// 恢复执行失败
				log.Printf("恢复执行失败: %v", err)
			} else {
				// 恢复执行成功，输出最终结果
				fmt.Printf("✅ 最终结果: %+v\n", finalResult)
			}
		} else {
			// 非中断错误，输出错误信息
			log.Printf("执行工作流失败: %v", err)
		}
	} else {
		// 工作流直接完成（没有触发中断），输出结果
		fmt.Printf("✅ 直接完成，结果: %+v\n", result)
	}
}

// beforeNodeInterruptDemo 节点前中断功能演示
// 演示如何在关键节点执行前设置中断点，实现执行前的验证和确认机制
// 这种中断方式特别适用于需要在重要操作前进行人工审核或权限检查的场景
// 参数:
//   - ctx: 上下文对象，用于控制演示过程的生命周期
func beforeNodeInterruptDemo(ctx context.Context) {
	fmt.Println("\n=== 节点前中断演示 ===")

	// 创建工作流执行器实例
	executor := NewBasicWorkflowExecutor()

	// 添加三个处理节点，模拟包含关键操作的工作流
	executor.AddNode("preparation", createProcessingNode("准备阶段", 1)) // 准备阶段：进行前期准备工作
	executor.AddNode("critical", createProcessingNode("关键处理", 2))    // 关键处理：执行重要的业务逻辑
	executor.AddNode("cleanup", createProcessingNode("清理阶段", 3))     // 清理阶段：进行后续清理工作

	// 设置节点执行顺序，确保按照预期的流程执行
	executor.SetExecutionSequence([]string{"preparation", "critical", "cleanup"})

	// 配置节点前中断：在关键节点执行前触发中断
	// 这样可以在执行重要操作前进行最后的确认或验证
	executor.WithInterruptBeforeNodes([]string{"critical"})

	// 准备输入数据，模拟需要谨慎处理的重要数据
	input := StepInput{
		Content: "重要数据处理", // 重要的业务数据
		Step:    1,        // 起始步骤编号
	}

	// 开始执行工作流
	fmt.Println("📝 开始执行工作流...")
	result, err := executor.Execute(ctx, input)

	// 处理执行结果和中断逻辑
	if err != nil {
		// 检查是否为预期的中断错误
		if interruptErr, ok := err.(*BasicInterruptError); ok {
			// 成功捕获节点前中断
			fmt.Printf("⏸️  工作流在节点 '%s' 前中断\n", interruptErr.InterruptNode)
			fmt.Printf("📊 中断时的状态: %+v\n", result)

			// 模拟人工确认过程
			// 在实际应用中，这里可能包括：
			// - 权限验证
			// - 数据完整性检查
			// - 业务规则验证
			// - 人工审批流程
			fmt.Println("🔍 进行人工确认...")
			time.Sleep(1 * time.Second) // 模拟确认过程的耗时
			fmt.Println("✅ 确认通过，继续执行")

			// 创建新的执行器继续执行剩余流程
			resumeExecutor := NewBasicWorkflowExecutor()
			// 重新添加所有节点
			resumeExecutor.AddNode("preparation", createProcessingNode("准备阶段", 1))
			resumeExecutor.AddNode("critical", createProcessingNode("关键处理", 2))
			resumeExecutor.AddNode("cleanup", createProcessingNode("清理阶段", 3))
			resumeExecutor.SetExecutionSequence([]string{"preparation", "critical", "cleanup"})
			// 注意：恢复执行时不设置中断点，确保流程能够完整执行

			// 执行完整的工作流
			finalResult, err := resumeExecutor.Execute(ctx, input)
			if err != nil {
				// 恢复执行失败
				log.Printf("恢复执行失败: %v", err)
			} else {
				// 恢复执行成功，输出最终结果
				fmt.Printf("✅ 最终结果: %+v\n", finalResult)
			}
		} else {
			// 非中断错误，输出错误信息
			log.Printf("执行工作流失败: %v", err)
		}
	}
}

// multiPointInterruptDemo 多点中断功能演示
// 演示如何在一个工作流中设置多个中断点，以及如何处理复杂的中断恢复场景
// 展示了节点前中断和节点后中断的组合使用，模拟真实的分阶段执行需求
// 参数:
//   - ctx: 上下文对象，用于控制演示过程的生命周期
func multiPointInterruptDemo(ctx context.Context) {
	fmt.Println("\n=== 多点中断演示 ===")

	// 创建工作流执行器实例
	executor := NewBasicWorkflowExecutor()

	// 定义四个处理阶段，构建完整的数据处理管道
	nodeNames := []string{"input", "validate", "process", "output"}

	// 批量添加处理节点，每个节点代表数据处理的一个阶段
	for i, name := range nodeNames {
		// 为每个阶段创建对应的处理节点
		executor.AddNode(name, createProcessingNode(fmt.Sprintf("%s阶段", name), i+1))
	}

	// 设置节点的执行顺序，确保数据按照正确的流程处理
	executor.SetExecutionSequence(nodeNames)

	// 配置多个中断点，演示复杂的中断控制策略
	// 1. 在验证阶段完成后中断 - 用于检查验证结果
	// 2. 在输出阶段开始前中断 - 用于最终确认
	executor.WithInterruptAfterNodes([]string{"validate"}).
		WithInterruptBeforeNodes([]string{"output"})

	// 准备初始输入数据
	input := StepInput{
		Content: "多阶段处理数据", // 需要多阶段处理的数据
		Step:    1,         // 起始步骤编号
	}

	fmt.Println("📝 开始执行多点中断工作流...")

	// 实现多次执行和恢复的循环逻辑
	// 模拟真实场景中可能需要多次中断和恢复的情况
	currentInput := input // 当前执行的输入数据
	interruptCount := 0   // 已发生的中断次数计数器
	maxInterrupts := 3    // 最大允许的中断次数

	// 循环处理中断和恢复，直到工作流完成或达到最大中断次数
	for interruptCount < maxInterrupts {
		// 根据当前的中断处理进度创建新的执行器
		testExecutor := NewBasicWorkflowExecutor()

		// 重新添加所有处理节点
		for i, name := range nodeNames {
			testExecutor.AddNode(name, createProcessingNode(fmt.Sprintf("%s阶段", name), i+1))
		}
		testExecutor.SetExecutionSequence(nodeNames)

		// 根据中断次数动态设置不同的中断点
		// 这种策略允许逐步推进工作流的执行
		if interruptCount == 0 {
			// 第一次执行：在验证阶段后中断
			testExecutor.WithInterruptAfterNodes([]string{"validate"})
		} else if interruptCount == 1 {
			// 第二次执行：在输出阶段前中断
			testExecutor.WithInterruptBeforeNodes([]string{"output"})
		}
		// 第三次执行：不设置中断点，完整执行

		// 执行当前配置的工作流
		result, err := testExecutor.Execute(ctx, currentInput)

		// 处理执行结果
		if err != nil {
			// 检查是否为预期的中断错误
			if interruptErr, ok := err.(*BasicInterruptError); ok {
				// 成功捕获中断，更新计数器
				interruptCount++
				fmt.Printf("⏸️  第 %d 次中断在: %s\n", interruptCount, interruptErr.InterruptNode)
				fmt.Printf("📊 中断时的状态: %+v\n", result)

				// 更新输入数据为中断时的状态，为下次恢复执行做准备
				// 在实际应用中，这里可能需要更复杂的状态管理
				if result.Result != "" {
					currentInput = StepInput{
						Content: result.Result, // 使用中断时的结果作为下次的输入
						Step:    result.Step,   // 更新步骤编号
					}
				}

				// 模拟中断处理过程
				// 在实际应用中，这里可能包括：
				// - 状态持久化
				// - 通知相关人员
				// - 执行补偿操作
				// - 更新执行日志
				fmt.Printf("🔧 处理第 %d 次中断...\n", interruptCount)
				time.Sleep(800 * time.Millisecond) // 模拟中断处理耗时
				fmt.Println("✅ 中断处理完成，继续执行")

				// 继续下一轮执行
				continue
			} else {
				// 非中断错误，终止执行
				log.Printf("执行失败: %v", err)
				break
			}
		} else {
			// 工作流成功完成，没有更多中断
			fmt.Printf("✅ 所有中断点处理完成，最终结果: %+v\n", result)
			break
		}
	}

	// 检查是否因为达到最大中断次数而退出
	if interruptCount >= maxInterrupts {
		fmt.Println("⚠️  达到最大中断处理次数")
	}
}

// initBasicConfig 初始化基础配置
// 加载配置文件，为演示程序提供必要的配置参数
// 如果配置文件读取失败，程序将使用默认配置继续运行
func initBasicConfig() {
	// 设置配置文件路径，指向上级目录的 config.yaml 文件
	viper.SetConfigFile("../config.yaml")

	// 尝试读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		// 配置文件读取失败时记录日志，但不中断程序执行
		// 这种容错机制确保演示程序能够在没有配置文件的情况下正常运行
		log.Printf("读取配置文件失败: %v (使用默认配置)", err)
	}
}

// RunBasicInterruptDemo 运行基础中断功能演示
// 这个函数演示了工作流中断和恢复机制的各种使用场景，包括：
// 1. 基础中断演示 - 展示节点后中断的基本用法
// 2. 节点前中断演示 - 展示在关键操作前的中断控制
// 3. 多点中断演示 - 展示复杂的多阶段中断和恢复流程
// 这些演示帮助理解如何在实际应用中实现可控的工作流执行
func main() {
	// 初始化配置，为演示程序做准备
	initBasicConfig()

	// 创建上下文对象，用于控制所有演示的生命周期
	ctx := context.Background()

	// 按顺序运行各种中断演示，展示不同的中断控制策略

	// 0. 基础工作流演示
	demonstrateBasicWorkflow()

	// 1. 基础中断演示：展示最简单的中断和恢复机制
	basicInterruptDemo(ctx)

	// 2. 节点前中断演示：展示在关键节点执行前的中断控制
	beforeNodeInterruptDemo(ctx)

	// 3. 多点中断演示：展示复杂的多阶段中断处理
	multiPointInterruptDemo(ctx)

	// 所有演示完成，输出结束信息
	fmt.Println("\n🎉 基础中断演示完成！")
}
