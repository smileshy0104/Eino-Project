package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/spf13/viper"
)

// 简单任务结构
type SimpleTask struct {
	TaskID   string `json:"task_id"`
	Step     int    `json:"step"`
	Content  string `json:"content"`
	Complete bool   `json:"complete"`
}

// 自定义中断错误
type SimpleInterruptError struct {
	Message     string
	InterruptAt string
	NeedsAction bool
}

func (e *SimpleInterruptError) Error() string {
	return fmt.Sprintf("中断于 %s: %s", e.InterruptAt, e.Message)
}

// 步骤1：准备阶段
func preparationStep(ctx context.Context, task SimpleTask) (SimpleTask, error) {
	fmt.Printf("🔄 执行准备阶段，任务ID: %s\n", task.TaskID)

	task.Step = 1
	task.Content += " -> 准备完成"

	time.Sleep(500 * time.Millisecond)

	fmt.Printf("✅ 准备阶段完成\n")
	return task, nil
}

// 步骤2：处理阶段（可能中断）
func processingStep(ctx context.Context, task SimpleTask) (SimpleTask, error) {
	fmt.Printf("🔄 执行处理阶段，任务ID: %s\n", task.TaskID)

	task.Step = 2
	task.Content += " -> 开始处理"

	time.Sleep(500 * time.Millisecond)

	// 模拟在处理后需要中断检查
	fmt.Printf("⏸️  处理阶段完成，需要质量检查\n")
	task.Content += " -> 处理完成，等待检查"

	return task, &SimpleInterruptError{
		Message:     "需要质量检查确认",
		InterruptAt: "处理阶段后",
		NeedsAction: true,
	}
}

// 步骤3：完成阶段
func completionStep(ctx context.Context, task SimpleTask) (SimpleTask, error) {
	fmt.Printf("🔄 执行完成阶段，任务ID: %s\n", task.TaskID)

	task.Step = 3
	task.Content += " -> 最终完成"
	task.Complete = true

	time.Sleep(500 * time.Millisecond)

	fmt.Printf("✅ 完成阶段完成\n")
	return task, nil
}

// 简单工作流执行器
func executeSimpleWorkflow(ctx context.Context, initialTask SimpleTask) {
	fmt.Printf("📝 开始执行工作流，任务: %s\n", initialTask.TaskID)

	currentTask := initialTask
	var err error

	// 步骤1：准备阶段
	fmt.Println("\n--- 步骤1: 准备阶段 ---")
	currentTask, err = preparationStep(ctx, currentTask)
	if err != nil {
		fmt.Printf("❌ 准备阶段失败: %v\n", err)
		return
	}

	// 步骤2：处理阶段
	fmt.Println("\n--- 步骤2: 处理阶段 ---")
	currentTask, err = processingStep(ctx, currentTask)
	if err != nil {
		if interruptErr, ok := err.(*SimpleInterruptError); ok {
			fmt.Printf("⏸️  工作流中断: %s\n", interruptErr.Error())

			if interruptErr.NeedsAction {
				fmt.Println("👤 进行质量检查...")
				time.Sleep(1 * time.Second)
				fmt.Println("✅ 质量检查通过，继续执行")
			}
		} else {
			fmt.Printf("❌ 处理阶段失败: %v\n", err)
			return
		}
	}

	// 步骤3：完成阶段
	fmt.Println("\n--- 步骤3: 完成阶段 ---")
	currentTask, err = completionStep(ctx, currentTask)
	if err != nil {
		fmt.Printf("❌ 完成阶段失败: %v\n", err)
		return
	}

	fmt.Printf("🎉 工作流执行完成！\n")
	fmt.Printf("📊 最终状态: %+v\n", currentTask)
}

// 多中断点演示
func multiInterruptDemo(ctx context.Context) {
	fmt.Println("=== 多中断点演示 ===")

	steps := []string{"验证", "处理", "审核", "发布"}
	task := SimpleTask{
		TaskID:  "MULTI-001",
		Content: "初始任务",
	}

	for i, stepName := range steps {
		fmt.Printf("\n--- 步骤%d: %s ---\n", i+1, stepName)

		// 模拟每个步骤
		fmt.Printf("🔄 执行 %s...\n", stepName)
		time.Sleep(300 * time.Millisecond)

		task.Step = i + 1
		task.Content += fmt.Sprintf(" -> %s完成", stepName)

		// 在特定步骤设置中断点
		if stepName == "处理" || stepName == "审核" {
			fmt.Printf("⏸️  %s阶段完成，需要确认\n", stepName)

			// 模拟人工确认
			fmt.Println("👤 等待确认...")
			time.Sleep(800 * time.Millisecond)
			fmt.Println("✅ 确认通过，继续")
		} else {
			fmt.Printf("✅ %s阶段完成\n", stepName)
		}
	}

	task.Complete = true
	fmt.Printf("🎉 多步骤工作流完成！\n")
	fmt.Printf("📊 最终状态: %+v\n", task)
}

// 条件中断演示
func conditionalInterruptDemo(ctx context.Context) {
	fmt.Println("\n=== 条件中断演示 ===")

	tasks := []SimpleTask{
		{TaskID: "COND-001", Content: "低风险任务"},
		{TaskID: "COND-002", Content: "中风险任务"},
		{TaskID: "COND-003", Content: "高风险任务"},
	}

	for _, task := range tasks {
		fmt.Printf("\n📝 处理任务: %s\n", task.TaskID)

		// 模拟风险评估
		var riskLevel string
		switch task.TaskID {
		case "COND-001":
			riskLevel = "低"
		case "COND-002":
			riskLevel = "中"
		case "COND-003":
			riskLevel = "高"
		}

		fmt.Printf("🔍 风险评估: %s风险\n", riskLevel)
		time.Sleep(300 * time.Millisecond)

		// 根据风险级别决定是否中断
		switch riskLevel {
		case "低":
			fmt.Println("✅ 低风险，自动通过")
			task.Complete = true

		case "中":
			fmt.Println("⏸️  中风险，需要主管确认")
			fmt.Println("👤 等待主管确认...")
			time.Sleep(1 * time.Second)
			fmt.Println("✅ 主管确认通过")
			task.Complete = true

		case "高":
			fmt.Println("⏸️  高风险，需要委员会审议")
			fmt.Println("👥 提交委员会审议...")
			time.Sleep(1500 * time.Millisecond)
			fmt.Println("✅ 委员会审议通过")
			task.Complete = true
		}

		fmt.Printf("📊 任务 %s 处理完成，状态: 完成=%v\n", task.TaskID, task.Complete)
	}
}

func initSimpleConfig() {
	viper.SetConfigFile("../config.yaml")
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("读取配置文件失败: %v (使用默认配置)", err)
	}
}

func main() {
	initSimpleConfig()
	ctx := context.Background()

	// 基础中断演示
	fmt.Println("=== 基础中断演示 ===")
	task := SimpleTask{
		TaskID:  "BASIC-001",
		Content: "基础任务",
	}
	executeSimpleWorkflow(ctx, task)

	// 多中断点演示
	multiInterruptDemo(ctx)

	// 条件中断演示
	conditionalInterruptDemo(ctx)

	fmt.Println("\n🎉 所有中断演示完成！")
}
