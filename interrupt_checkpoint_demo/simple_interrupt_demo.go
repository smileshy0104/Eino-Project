// Package main 演示简单中断机制的实现和使用
// 该文件展示了如何在工作流中实现基本的中断处理，
// 包括单点中断、多点中断和条件中断等不同场景，
// 适合初学者理解中断机制的基本概念和实现方式
package main

import (
	"context" // 上下文管理，用于控制工作流执行
	"fmt"     // 格式化输出，用于显示执行状态
	"log"     // 日志记录，用于记录配置加载信息
	"time"    // 时间处理，用于模拟处理延迟

	"github.com/spf13/viper" // 配置文件管理库
)

// SimpleTask 简单任务结构
// 表示工作流中的基本任务单元，包含任务标识、执行步骤、内容和完成状态
// 用于在工作流的各个步骤之间传递和更新任务状态信息
type SimpleTask struct {
	TaskID   string `json:"task_id"`  // 任务唯一标识符
	Step     int    `json:"step"`     // 当前执行步骤编号
	Content  string `json:"content"`  // 任务内容描述，记录执行过程
	Complete bool   `json:"complete"` // 任务是否已完成
}

// SimpleInterruptError 自定义简单中断错误类型
// 当工作流需要在特定点中断时抛出的错误
// 包含中断消息、中断位置和是否需要人工操作等信息
type SimpleInterruptError struct {
	Message     string // 中断消息描述
	InterruptAt string // 中断发生的位置或阶段
	NeedsAction bool   // 是否需要人工操作才能继续
}

// Error 实现error接口
// 返回格式化的中断错误信息，包含中断位置和消息描述
func (e *SimpleInterruptError) Error() string {
	return fmt.Sprintf("中断于 %s: %s", e.InterruptAt, e.Message)
}

// preparationStep 步骤1：准备阶段
// 执行工作流的准备工作，通常不会触发中断
// 更新任务步骤为1，并在内容中记录准备完成状态
// 参数 ctx: 上下文对象，用于控制执行流程
// 参数 task: 当前任务对象
// 返回更新后的任务对象和可能的错误
func preparationStep(ctx context.Context, task SimpleTask) (SimpleTask, error) {
	fmt.Printf("🔄 执行准备阶段，任务ID: %s\n", task.TaskID)

	task.Step = 1              // 设置当前步骤为1
	task.Content += " -> 准备完成" // 在内容中记录准备完成

	time.Sleep(500 * time.Millisecond) // 模拟准备阶段的处理时间

	fmt.Printf("✅ 准备阶段完成\n")
	return task, nil // 返回更新后的任务，无错误
}

// processingStep 步骤2：处理阶段（可能中断）
// 执行工作流的核心处理逻辑，在处理完成后触发中断
// 这是一个演示中断机制的关键步骤，会主动抛出中断错误
// 参数 ctx: 上下文对象，用于控制执行流程
// 参数 task: 当前任务对象
// 返回更新后的任务对象和中断错误
func processingStep(ctx context.Context, task SimpleTask) (SimpleTask, error) {
	fmt.Printf("🔄 执行处理阶段，任务ID: %s\n", task.TaskID)

	task.Step = 2              // 设置当前步骤为2
	task.Content += " -> 开始处理" // 记录开始处理状态

	time.Sleep(500 * time.Millisecond) // 模拟处理阶段的处理时间

	// 模拟在处理后需要中断检查
	fmt.Printf("⏸️  处理阶段完成，需要质量检查\n")
	task.Content += " -> 处理完成，等待检查" // 记录等待检查状态

	// 返回中断错误，触发工作流中断
	return task, &SimpleInterruptError{
		Message:     "需要质量检查确认", // 中断原因说明
		InterruptAt: "处理阶段后",    // 中断发生位置
		NeedsAction: true,       // 标记需要人工操作
	}
}

// completionStep 步骤3：完成阶段
// 执行工作流的最终完成逻辑，标记任务为已完成状态
// 这是工作流的最后一个步骤，通常不会触发中断
// 参数 ctx: 上下文对象，用于控制执行流程
// 参数 task: 当前任务对象
// 返回最终完成的任务对象和可能的错误
func completionStep(ctx context.Context, task SimpleTask) (SimpleTask, error) {
	fmt.Printf("🔄 执行完成阶段，任务ID: %s\n", task.TaskID)

	task.Step = 3              // 设置当前步骤为3
	task.Content += " -> 最终完成" // 记录最终完成状态
	task.Complete = true       // 标记任务为已完成

	time.Sleep(500 * time.Millisecond) // 模拟完成阶段的处理时间

	fmt.Printf("✅ 完成阶段完成\n")
	return task, nil // 返回最终完成的任务，无错误
}

// executeSimpleWorkflow 简单工作流执行器
// 按顺序执行准备、处理、完成三个阶段，并处理可能的中断
// 当遇到中断错误时，会进行相应的处理并继续执行后续步骤
// 该函数演示了基本的工作流执行逻辑和中断处理机制
// 参数 ctx: 上下文对象，用于控制执行流程
// 参数 initialTask: 要执行的初始任务对象
func executeSimpleWorkflow(ctx context.Context, initialTask SimpleTask) {
	fmt.Printf("📝 开始执行工作流，任务: %s\n", initialTask.TaskID)

	currentTask := initialTask // 保存当前任务状态
	var err error

	// 步骤1：准备阶段
	// 执行工作流的准备工作，通常不会触发中断
	fmt.Println("\n--- 步骤1: 准备阶段 ---")
	currentTask, err = preparationStep(ctx, currentTask)
	if err != nil {
		fmt.Printf("❌ 准备阶段失败: %v\n", err)
		return // 准备阶段出错，终止执行
	}

	// 步骤2：处理阶段（可能中断）
	// 执行核心处理逻辑，这里会触发中断进行质量检查
	fmt.Println("\n--- 步骤2: 处理阶段 ---")
	currentTask, err = processingStep(ctx, currentTask)
	if err != nil {
		// 检查是否是中断错误
		if interruptErr, ok := err.(*SimpleInterruptError); ok {
			fmt.Printf("⏸️  工作流中断: %s\n", interruptErr.Error())

			// 检查是否需要人工干预
			if interruptErr.NeedsAction {
				fmt.Println("👤 进行质量检查...")  // 模拟人工质量检查
				time.Sleep(1 * time.Second) // 模拟检查时间
				fmt.Println("✅ 质量检查通过，继续执行")
			}
			// 中断处理完成，继续执行后续步骤
		} else {
			fmt.Printf("❌ 处理阶段失败: %v\n", err)
			return // 非中断错误，终止执行
		}
	}

	// 步骤3：完成阶段
	// 执行最终完成逻辑，标记任务为已完成状态
	fmt.Println("\n--- 步骤3: 完成阶段 ---")
	currentTask, err = completionStep(ctx, currentTask)
	if err != nil {
		fmt.Printf("❌ 完成阶段失败: %v\n", err)
		return // 完成阶段出错，终止执行
	}

	fmt.Printf("🎉 工作流执行完成！\n")
	fmt.Printf("📊 最终状态: %+v\n", currentTask) // 显示最终任务状态
}

// multiInterruptDemo 多中断点演示
// 演示在单个工作流中设置多个中断点的场景
// 该函数模拟一个包含验证、处理、审核、发布四个步骤的工作流
// 在处理和审核两个关键步骤设置中断点，需要人工确认后才能继续
// 这种场景常见于需要多重检查的业务流程，如文档发布、代码部署等
// 参数 ctx: 上下文对象，用于控制执行流程
func multiInterruptDemo(ctx context.Context) {
	fmt.Println("=== 多中断点演示 ===")

	// 定义工作流的各个步骤
	steps := []string{"验证", "处理", "审核", "发布"}
	task := SimpleTask{
		TaskID:  "MULTI-001", // 多中断点演示任务ID
		Content: "初始任务",      // 初始任务内容
	}

	// 逐步执行工作流，在关键步骤设置中断点
	for i, stepName := range steps {
		fmt.Printf("\n--- 步骤%d: %s ---\n", i+1, stepName)

		// 模拟每个步骤的执行
		fmt.Printf("🔄 执行 %s...\n", stepName)
		time.Sleep(300 * time.Millisecond) // 模拟步骤执行时间

		task.Step = i + 1                                 // 更新当前步骤
		task.Content += fmt.Sprintf(" -> %s完成", stepName) // 记录步骤完成状态

		// 在特定步骤设置中断点，需要人工确认
		if stepName == "处理" || stepName == "审核" {
			fmt.Printf("⏸️  %s阶段完成，需要确认\n", stepName)

			// 模拟人工确认过程
			fmt.Println("👤 等待确认...")
			time.Sleep(800 * time.Millisecond) // 模拟人工确认时间
			fmt.Println("✅ 确认通过，继续")
		} else {
			// 非中断步骤，直接完成
			fmt.Printf("✅ %s阶段完成\n", stepName)
		}
	}

	task.Complete = true // 标记任务完成
	fmt.Printf("🎉 多步骤工作流完成！\n")
	fmt.Printf("📊 最终状态: %+v\n", task) // 显示最终任务状态
}

// conditionalInterruptDemo 条件中断演示
// 演示基于业务条件动态触发中断的场景
// 该函数模拟风险评估系统，根据不同的风险级别采取不同的处理策略
// 低风险任务自动通过，中风险需要主管确认，高风险需要委员会审议
// 这种场景常见于金融风控、审批流程、质量管控等需要分级处理的业务
// 参数 ctx: 上下文对象，用于控制执行流程
func conditionalInterruptDemo(ctx context.Context) {
	fmt.Println("\n=== 条件中断演示 ===")

	// 创建不同风险级别的测试任务
	tasks := []SimpleTask{
		{TaskID: "COND-001", Content: "低风险任务"}, // 低风险，自动通过
		{TaskID: "COND-002", Content: "中风险任务"}, // 中风险，需要主管确认
		{TaskID: "COND-003", Content: "高风险任务"}, // 高风险，需要委员会审议
	}

	// 逐个处理任务，根据风险级别采取不同策略
	for _, task := range tasks {
		fmt.Printf("\n📝 处理任务: %s\n", task.TaskID)

		// 模拟风险评估过程
		var riskLevel string
		switch task.TaskID {
		case "COND-001":
			riskLevel = "低" // 低风险级别
		case "COND-002":
			riskLevel = "中" // 中风险级别
		case "COND-003":
			riskLevel = "高" // 高风险级别
		}

		fmt.Printf("🔍 风险评估: %s风险\n", riskLevel)
		time.Sleep(300 * time.Millisecond) // 模拟风险评估时间

		// 根据风险级别决定是否中断以及中断处理方式
		switch riskLevel {
		case "低":
			// 低风险：无需中断，自动通过
			fmt.Println("✅ 低风险，自动通过")
			task.Complete = true

		case "中":
			// 中风险：触发中断，需要主管确认
			fmt.Println("⏸️  中风险，需要主管确认")
			fmt.Println("👤 等待主管确认...")
			time.Sleep(1 * time.Second) // 模拟主管确认时间
			fmt.Println("✅ 主管确认通过")
			task.Complete = true

		case "高":
			// 高风险：触发中断，需要委员会审议
			fmt.Println("⏸️  高风险，需要委员会审议")
			fmt.Println("👥 提交委员会审议...")
			time.Sleep(1500 * time.Millisecond) // 模拟委员会审议时间
			fmt.Println("✅ 委员会审议通过")
			task.Complete = true
		}

		fmt.Printf("📊 任务 %s 处理完成，状态: 完成=%v\n", task.TaskID, task.Complete)
	}
}

// initSimpleConfig 初始化简单中断演示配置
// 设置演示程序运行所需的各项配置参数
// 包括中断控制、重试机制、超时设置、恢复策略等关键配置
// 这些配置在实际应用中通常从配置文件或环境变量中读取
func initSimpleConfig() {
	viper.SetConfigFile("../config.yaml")
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("读取配置文件失败: %v (使用默认配置)", err)
	}
}

// RunSimpleInterruptDemo 运行简单中断演示
// 这是简单中断演示的主入口函数，展示了基本的工作流中断机制
// 包含三个演示场景：基础中断演示、多中断点演示、条件中断演示
// 通过这些演示可以了解中断机制的基本工作原理和应用场景
func main() {
	// 初始化配置
	initSimpleConfig() // 初始化演示所需的配置参数
	ctx := context.Background()

	// 演示1：基础中断演示
	// 展示标准的三步骤工作流：准备->处理->完成
	// 在处理阶段会触发中断，演示中断处理机制
	fmt.Println("=== 基础中断演示 ===")
	task := SimpleTask{
		TaskID:  "BASIC-001", // 任务唯一标识
		Content: "基础任务",      // 任务内容描述
	}
	executeSimpleWorkflow(ctx, task)

	// 演示2：多中断点演示
	// 展示在一个工作流中可能发生多次中断的情况
	multiInterruptDemo(ctx)

	// 演示3：条件中断演示
	// 展示基于特定条件触发的中断机制
	conditionalInterruptDemo(ctx)

	fmt.Println("\n🎉 所有中断演示完成！")
}
