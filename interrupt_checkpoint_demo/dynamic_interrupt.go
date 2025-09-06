// Package main 演示动态中断机制的实现和使用
// 该文件展示了如何在业务流程中实现智能的动态中断处理，
// 根据业务条件和风险评估动态决定是否中断工作流执行，
// 支持不同类型的中断处理策略和自动重试机制
package main

import (
	"context"   // 上下文管理
	"fmt"       // 格式化输出
	"log"       // 日志记录
	"math/rand" // 随机数生成，用于模拟业务场景
	"time"      // 时间处理

	"github.com/spf13/viper" // 配置文件管理
)

// BusinessData 业务数据结构
// 表示订单处理过程中的业务数据，包含订单信息、处理状态和元数据
// 用于在工作流的各个步骤之间传递和更新业务状态信息
type BusinessData struct {
	OrderID     string                 `json:"order_id"`     // 订单唯一标识符
	Amount      float64                `json:"amount"`       // 订单金额
	Status      string                 `json:"status"`       // 当前处理状态
	Items       []string               `json:"items"`        // 订单商品列表
	Metadata    map[string]interface{} `json:"metadata"`     // 存储额外的业务元数据
	ProcessStep int                    `json:"process_step"` // 当前处理步骤编号
}

// DynamicInterruptError 自定义动态中断错误类型
// 当业务流程需要根据条件动态中断时抛出的特殊错误
// 包含中断原因、是否需要人工审批等信息，支持智能的中断处理策略
type DynamicInterruptError struct {
	Message          string                 // 中断消息描述
	RequiresApproval bool                   // 是否需要人工审批
	InterruptReason  string                 // 中断的具体原因
	Metadata         map[string]interface{} // 中断相关的元数据
	Timestamp        string                 // 中断发生的时间戳
}

// Error 实现error接口
// 返回格式化的动态中断错误信息，包含中断消息、原因和是否需要审批
func (e *DynamicInterruptError) Error() string {
	return fmt.Sprintf("动态中断: %s", e.Message)
}

func NewDynamicInterruptError(reason string, requiresApproval bool) *DynamicInterruptError {
	return &DynamicInterruptError{
		Message:          fmt.Sprintf("业务中断: %s", reason),
		RequiresApproval: requiresApproval,
		InterruptReason:  reason,
		Metadata:         make(map[string]interface{}),
		Timestamp:        time.Now().Format(time.RFC3339),
	}
}

// BusinessStep 业务步骤处理函数类型
// 定义业务流程中单个步骤的函数签名
// 接收上下文和业务数据，返回处理后的业务数据和可能的错误（包括动态中断错误）
type BusinessStep func(ctx context.Context, data BusinessData) (BusinessData, error)

// createRiskAssessmentNode 创建风险评估节点 - 可能触发动态中断
// 返回一个业务步骤函数，用于评估订单的风险等级
// 根据随机生成的风险分数决定是否需要中断处理：
// - 风险分数 > 80: 高风险订单，需要人工审核（触发需要审批的中断）
// - 风险分数 > 60: 中等风险订单，需要额外验证（触发自动处理的中断）
// - 风险分数 <= 60: 低风险订单，自动通过
// 同时为订单数据添加风险评分等元数据信息
func createRiskAssessmentNode() BusinessStep {
	return func(ctx context.Context, data BusinessData) (BusinessData, error) {
		fmt.Printf("🔍 执行风险评估，订单ID: %s, 金额: %.2f\n", data.OrderID, data.Amount)

		// 更新处理步骤和状态
		data.ProcessStep = 1
		data.Status = "风险评估中"

		// 模拟风险评估逻辑 - 生成0-100的随机风险分数
		riskScore := rand.Float64() * 100

		// 初始化元数据映射（如果尚未初始化）
		if data.Metadata == nil {
			data.Metadata = make(map[string]interface{})
		}
		// 保存风险评分到元数据中
		data.Metadata["risk_score"] = riskScore

		fmt.Printf("📊 风险分数: %.2f\n", riskScore)

		// 高风险订单需要人工审核 - 触发需要审批的动态中断
		if riskScore > 80 {
			fmt.Printf("⚠️  高风险订单，需要人工审核\n")
			return data, NewDynamicInterruptError("高风险订单检测", true)
		}

		// 中等风险订单需要额外验证 - 触发自动处理的动态中断
		if riskScore > 60 {
			fmt.Printf("⚠️  中等风险订单，需要额外验证\n")
			return data, NewDynamicInterruptError("中等风险订单检测", false)
		}

		// 低风险订单自动通过，更新状态
		data.Status = "风险评估通过"
		fmt.Printf("✅ 低风险订单，自动通过\n")
		return data, nil // 返回更新后的数据，无错误
	}
}

// createInventoryCheckNode 创建库存检查节点 - 可能触发动态中断
// 返回一个业务步骤函数，用于检查订单中所有商品的库存状态
// 遍历订单中的每个商品，生成随机库存数量进行模拟检查：
// - 库存数量 < 10: 触发库存不足中断，需要补货（不需要人工审批）
// - 库存数量 >= 10: 库存充足，继续处理
// 将每个商品的库存信息保存到业务数据的元数据中
func createInventoryCheckNode() BusinessStep {
	return func(ctx context.Context, data BusinessData) (BusinessData, error) {
		fmt.Printf("📦 检查库存，订单项目: %v\n", data.Items)

		// 更新处理步骤和状态
		data.ProcessStep = 2
		data.Status = "库存检查中"

		// 模拟库存检查 - 遍历所有订单商品
		for i, item := range data.Items {
			// 生成0-99的随机库存数量
			stock := rand.Intn(100)
			// 将库存信息保存到元数据中
			data.Metadata[fmt.Sprintf("stock_%d", i)] = stock

			fmt.Printf("📊 商品 %s 库存: %d\n", item, stock)

			// 库存不足需要中断 - 阈值为10
			if stock < 10 {
				fmt.Printf("⚠️  商品 %s 库存不足，需要补货\n", item)
				// 触发库存不足的动态中断，不需要人工审批
				return data, NewDynamicInterruptError(fmt.Sprintf("商品 %s 库存不足", item), false)
			}
		}

		// 所有商品库存充足，更新状态
		data.Status = "库存充足"
		fmt.Printf("✅ 所有商品库存充足\n")
		return data, nil // 返回更新后的数据，无错误
	}
}

// createPaymentProcessNode 创建支付处理节点 - 可能触发动态中断
// 返回一个业务步骤函数，用于处理订单的支付流程
// 支持两种中断场景：
// 1. 支付失败（30%概率）- 触发自动重试的中断，不需要人工审批
// 2. 大额支付（金额>10000）- 触发需要人工确认的中断
// 成功时记录支付时间等元数据信息
func createPaymentProcessNode() BusinessStep {
	return func(ctx context.Context, data BusinessData) (BusinessData, error) {
		fmt.Printf("💳 处理支付，订单金额: %.2f\n", data.Amount)

		// 更新处理步骤和状态
		data.ProcessStep = 3
		data.Status = "支付处理中"

		// 模拟支付处理 - 70%成功率
		paymentSuccess := rand.Float64() > 0.3

		// 支付失败触发自动重试中断
		if !paymentSuccess {
			fmt.Printf("❌ 支付失败，需要重试\n")
			return data, NewDynamicInterruptError("支付处理失败", false)
		}

		// 大额支付需要额外确认 - 触发需要审批的中断
		if data.Amount > 10000 {
			fmt.Printf("⚠️  大额支付，需要额外确认\n")
			return data, NewDynamicInterruptError("大额支付确认", true)
		}

		// 支付成功，更新状态和元数据
		data.Status = "支付成功"
		data.Metadata["payment_time"] = time.Now().Format(time.RFC3339)
		fmt.Printf("✅ 支付处理成功\n")
		return data, nil // 返回更新后的数据，无错误
	}
}

// createOrderCompletionNode 创建订单完成节点
// 返回一个业务步骤函数，用于完成订单的最终处理
// 这是工作流的最后一个步骤，标记订单为已完成状态
// 记录完成时间等最终信息，通常不会触发中断
func createOrderCompletionNode() BusinessStep {
	return func(ctx context.Context, data BusinessData) (BusinessData, error) {
		fmt.Printf("🎉 完成订单处理，订单ID: %s\n", data.OrderID)

		// 更新为最终处理步骤和完成状态
		data.ProcessStep = 4
		data.Status = "订单完成"
		// 记录订单完成时间到元数据
		data.Metadata["completion_time"] = time.Now().Format(time.RFC3339)

		return data, nil // 返回最终完成的订单数据
	}
}

// DynamicInterruptHandler 动态中断处理器
// 负责管理和执行业务工作流，支持智能的动态中断处理
// 可以根据业务条件自动决定中断策略，支持重试机制和人工审批流程
type DynamicInterruptHandler struct {
	maxRetries int // 最大重试次数，用于控制步骤失败时的重试行为
}

// NewDynamicInterruptHandler 创建新的动态中断处理器
// 初始化处理器的最大重试次数配置
// 参数 maxRetries: 每个业务步骤的最大重试次数
// 返回配置好的动态中断处理器实例
func NewDynamicInterruptHandler(maxRetries int) *DynamicInterruptHandler {
	return &DynamicInterruptHandler{maxRetries: maxRetries}
}

// processBusinessWorkflow 处理业务流程，支持动态中断和重试
// 按顺序执行业务步骤列表，当遇到动态中断时自动处理并重试
// 参数 ctx: 上下文对象，用于控制执行流程和超时
// 参数 data: 初始业务数据，在各步骤间传递和更新
// 参数 steps: 要执行的业务步骤列表
// 返回处理完成的业务数据和可能的错误
func (dih *DynamicInterruptHandler) processBusinessWorkflow(ctx context.Context, data BusinessData, steps []BusinessStep) (BusinessData, error) {
	currentData := data

	for stepIndex, step := range steps {
		stepName := fmt.Sprintf("步骤%d", stepIndex+1)

		for retry := 0; retry < dih.maxRetries; retry++ {
			result, err := step(ctx, currentData)

			if err != nil {
				if interruptErr, ok := err.(*DynamicInterruptError); ok {
					fmt.Printf("⏸️  %s 在第 %d 次尝试中被动态中断: %s\n", stepName, retry+1, interruptErr.InterruptReason)

					// 处理中断
					handled := dih.handleInterrupt(interruptErr)
					if handled {
						fmt.Printf("🔧 中断已处理，重试 %s\n", stepName)
						currentData = result // 更新状态
						continue             // 重试当前步骤
					} else {
						return result, fmt.Errorf("无法处理中断: %v", interruptErr)
					}
				} else {
					return result, fmt.Errorf("步骤 %s 执行失败: %v", stepName, err)
				}
			} else {
				// 步骤成功，更新当前数据并继续下一步骤
				currentData = result
				break
			}
		}
	}

	return currentData, nil
}

// handleInterrupt 处理具体的中断情况
// 根据中断错误的类型和原因，执行相应的处理策略
// 支持人工审批、自动重试、资源补充等多种处理方式
// 参数 interruptErr: 动态中断错误对象，包含中断原因和处理要求
// 返回 bool: true表示中断已成功处理可以重试，false表示无法处理
func (dih *DynamicInterruptHandler) handleInterrupt(interruptErr *DynamicInterruptError) bool {
	fmt.Printf("🔧 处理中断: %s\n", interruptErr.InterruptReason)

	// 根据中断类型进行不同的处理
	switch {
	case interruptErr.RequiresApproval: // 需要人工审批的中断类型
		fmt.Println("👤 需要人工审核...")  // 提示开始人工审核流程
		time.Sleep(1 * time.Second) // 模拟人工审核处理时间
		fmt.Println("✅ 人工审核通过")     // 审核完成提示
		return true                 // 审核通过，可以继续执行

	case interruptErr.InterruptReason == "中等风险订单检测": // 中等风险订单需要额外验证
		fmt.Println("🔍 执行额外风险验证...")       // 提示开始额外验证流程
		time.Sleep(800 * time.Millisecond) // 模拟风险验证处理时间
		fmt.Println("✅ 额外验证通过")            // 验证完成提示
		return true                        // 验证通过，可以继续执行

	case contains(interruptErr.InterruptReason, "库存不足"): // 库存不足类型的中断
		fmt.Println("📦 正在补货...")            // 提示开始自动补货流程
		time.Sleep(1200 * time.Millisecond) // 模拟补货处理时间
		fmt.Println("✅ 库存已补充")              // 补货完成提示
		return true                         // 补货完成，可以继续执行

	case interruptErr.InterruptReason == "支付处理失败": // 支付失败需要重试
		fmt.Println("💳 重新处理支付...")         // 提示开始支付重试流程
		time.Sleep(500 * time.Millisecond) // 模拟支付重试处理时间
		fmt.Println("✅ 支付处理就绪")            // 支付重试完成提示
		return true                        // 支付重试成功，可以继续执行

	default: // 未知的中断类型
		fmt.Printf("❌ 未知的中断类型: %s\n", interruptErr.InterruptReason) // 输出未知中断类型信息
		return false                                                // 无法处理未知类型，不能继续执行
	}
}

// contains 辅助函数：检查字符串是否包含指定子字符串
// 用于检查中断原因字符串中是否包含特定的关键词
// 支持前缀匹配和后缀匹配两种模式
// 参数 s: 要检查的主字符串
// 参数 substr: 要查找的子字符串
// 返回 bool: true表示包含子字符串，false表示不包含
func contains(s, substr string) bool {
	// 检查前缀匹配：主字符串长度足够且开头匹配子字符串
	// 或者检查后缀匹配：主字符串长度足够且结尾匹配子字符串
	return len(s) >= len(substr) && s[:len(substr)] == substr ||
		len(s) > len(substr) && (s[len(s)-len(substr):] == substr)
}

// dynamicInterruptDemo 动态中断演示函数
// 演示基本的动态中断处理机制，使用多个不同类型的测试订单
// 展示风险评估、库存检查、支付处理等步骤中的动态中断和自动恢复
// 通过不同金额和商品类型的订单，触发各种中断场景并观察处理结果
func dynamicInterruptDemo(ctx context.Context) {
	fmt.Println("=== 动态中断演示 ===")

	// 设置随机种子，确保每次运行都有不同的随机结果
	rand.Seed(time.Now().UnixNano())

	// 创建完整的业务处理步骤链，按顺序执行
	steps := []BusinessStep{
		createRiskAssessmentNode(),  // 1. 风险评估节点 - 可能因高风险触发中断
		createInventoryCheckNode(),  // 2. 库存检查节点 - 可能因库存不足触发中断
		createPaymentProcessNode(),  // 3. 支付处理节点 - 可能因支付失败或大额支付触发中断
		createOrderCompletionNode(), // 4. 订单完成节点 - 最终处理步骤
	}

	// 创建动态中断处理器，设置最大重试次数为5次
	handler := NewDynamicInterruptHandler(5)

	// 准备多个测试订单，覆盖不同的业务场景和中断触发条件
	orders := []BusinessData{
		{ // 测试订单1：中等金额订单，可能触发中等风险中断
			OrderID:  "ORD-001",
			Amount:   500.00,                       // 中等金额，风险评估可能通过
			Items:    []string{"手机", "充电器"},        // 常见商品，库存相对充足
			Status:   "待处理",                        // 初始状态
			Metadata: make(map[string]interface{}), // 初始化元数据存储
		},
		{ // 测试订单2：高金额订单，很可能触发高风险中断和大额支付中断
			OrderID:  "ORD-002",
			Amount:   15000.00,                      // 高金额，可能触发风险评估中断和大额支付中断
			Items:    []string{"笔记本电脑", "鼠标", "键盘"}, // 多个商品，增加库存检查复杂度
			Status:   "待处理",
			Metadata: make(map[string]interface{}),
		},
		{ // 测试订单3：低金额订单，预期很少触发中断
			OrderID:  "ORD-003",
			Amount:   200.00,               // 低金额，风险评估通过概率高
			Items:    []string{"书籍", "文具"}, // 简单商品，库存充足概率高
			Status:   "待处理",
			Metadata: make(map[string]interface{}),
		},
	}

	// 逐个处理每个测试订单，观察不同订单的中断处理行为
	for i, order := range orders {
		fmt.Printf("\n📝 开始处理订单 %d: %s (金额: %.2f)\n", i+1, order.OrderID, order.Amount)

		// 执行完整的业务流程，支持动态中断处理和自动重试
		result, err := handler.processBusinessWorkflow(ctx, order, steps)

		// 输出每个订单的最终处理结果
		if err != nil {
			fmt.Printf("❌ 订单 %s 处理失败: %v\n", order.OrderID, err)
		} else {
			fmt.Printf("✅ 订单 %s 处理成功！最终状态: %s\n", result.OrderID, result.Status)
		}
	}
}

// dynamicConditionalInterruptDemo 条件动态中断演示函数
// 演示基于多种业务条件的动态中断处理机制
// 通过预定义的条件检查规则，测试不同类型的中断触发和处理策略
// 包括紧急中断（需要立即处理）和非紧急中断（可以稍后处理）两种类型
func dynamicConditionalInterruptDemo(ctx context.Context) {
	fmt.Println("\n=== 条件中断演示 ===")

	// 创建条件检查步骤函数，实现多种业务条件的验证逻辑
	// 该函数会根据订单数据检查各种业务规则，并在违反规则时触发相应的中断
	conditionalStep := func(ctx context.Context, data BusinessData) (BusinessData, error) {
		fmt.Printf("🔍 执行条件检查，订单ID: %s\n", data.OrderID)

		// 更新处理步骤计数器
		data.ProcessStep += 1

		// 定义多种条件检查规则，每个规则包含检查逻辑、错误消息和紧急程度
		conditions := []struct {
			name    string                  // 条件检查的名称
			check   func(BusinessData) bool // 条件检查函数，返回true表示触发中断
			message string                  // 中断时显示的错误消息
			urgent  bool                    // 是否为紧急情况，需要立即处理
		}{
			{ // 条件1：订单金额验证 - 检查金额是否有效
				name:    "金额验证",
				check:   func(d BusinessData) bool { return d.Amount <= 0 }, // 金额小于等于0为无效
				message: "订单金额无效",
				urgent:  true, // 金额无效是紧急情况，需要立即处理
			},
			{ // 条件2：商品列表验证 - 检查是否有商品
				name:    "商品验证",
				check:   func(d BusinessData) bool { return len(d.Items) == 0 }, // 商品列表为空
				message: "订单商品列表为空",
				urgent:  true, // 无商品是紧急情况，需要立即处理
			},
			{ // 条件3：高价值订单检查 - 检查是否为超高价值订单
				name:    "高价值检查",
				check:   func(d BusinessData) bool { return d.Amount > 20000 }, // 金额超过20000
				message: "超高价值订单需要特殊处理",
				urgent:  false, // 高价值订单不是紧急情况，可以稍后处理
			},
			{ // 条件4：商品数量检查 - 检查商品种类是否过多
				name:    "商品数量检查",
				check:   func(d BusinessData) bool { return len(d.Items) > 10 }, // 商品种类超过10种
				message: "商品种类过多需要分批处理",
				urgent:  false, // 商品过多不是紧急情况，可以稍后处理
			},
		}

		// 遍历所有条件检查规则，按顺序执行验证
		for _, condition := range conditions {
			if condition.check(data) {
				fmt.Printf("⚠️  条件 '%s' 触发: %s\n", condition.name, condition.message)

				// 在业务数据的元数据中记录中断相关信息
				data.Metadata["interrupt_condition"] = condition.name // 记录触发的条件名称
				data.Metadata["urgent"] = condition.urgent            // 记录是否为紧急情况

				// 创建并返回动态中断错误
				return data, NewDynamicInterruptError(condition.message, condition.urgent)
			}
		}

		// 所有条件检查都通过，更新订单状态
		fmt.Println("✅ 所有条件检查通过")
		data.Status = "条件验证通过"
		return data, nil
	}

	// 创建动态中断处理器，设置最大重试次数为3次
	handler := NewDynamicInterruptHandler(3)

	// 创建多个测试用例，覆盖各种条件检查场景
	testCases := []BusinessData{
		{OrderID: "TEST-001", Amount: -100, Items: []string{"商品1"}},    // 测试用例1：金额无效（负数）
		{OrderID: "TEST-002", Amount: 100, Items: []string{}},          // 测试用例2：商品列表为空
		{OrderID: "TEST-003", Amount: 25000, Items: []string{"奢侈品"}},   // 测试用例3：超高价值订单
		{OrderID: "TEST-004", Amount: 1000, Items: make([]string, 15)}, // 测试用例4：商品种类过多（15种）
		{OrderID: "TEST-005", Amount: 100, Items: []string{"正常商品"}},    // 测试用例5：正常订单（应该通过所有检查）
	}

	// 逐个执行每个测试用例，观察不同条件下的中断处理行为
	for _, testCase := range testCases {
		fmt.Printf("\n📝 测试订单: %s (金额: %.2f, 商品数: %d)\n",
			testCase.OrderID, testCase.Amount, len(testCase.Items))

		// 为每个测试用例初始化元数据存储
		testCase.Metadata = make(map[string]interface{})

		// 执行条件检查步骤
		result, err := conditionalStep(ctx, testCase)

		// 处理执行结果，区分正常完成和各种类型的中断
		if err != nil {
			// 检查是否为动态中断错误
			if interruptErr, ok := err.(*DynamicInterruptError); ok {
				fmt.Printf("⏸️  按预期中断: %s\n", interruptErr.InterruptReason)

				// 根据中断的紧急程度显示不同的处理提示
				if interruptErr.RequiresApproval {
					fmt.Println("🚨 紧急情况，需要立即处理") // 紧急中断需要立即关注
				} else {
					fmt.Println("⏳ 非紧急情况，可以稍后处理") // 非紧急中断可以延后处理
				}

				// 尝试使用中断处理器处理当前中断
				handled := handler.handleInterrupt(interruptErr)
				if handled {
					fmt.Printf("✅ 中断已处理，订单可以继续\n")
				} else {
					fmt.Printf("❌ 中断无法处理\n")
				}
			} else {
				// 处理非动态中断的其他错误类型
				fmt.Printf("❌ 意外错误: %v\n", err)
			}
		} else {
			// 条件检查全部通过，订单处理成功
			fmt.Printf("✅ 订单处理成功: %s\n", result.Status)
		}
	}
}

// initDynamicConfig 初始化动态中断配置
// 设置配置文件的读取路径和默认参数值
// 支持从YAML配置文件中读取中断处理、业务规则等相关配置
// 如果配置文件不存在，将使用预设的默认值确保系统正常运行
func initDynamicConfig() {
	// 设置配置文件的基本信息
	viper.SetConfigName("dynamic_interrupt_config") // 配置文件名（不含扩展名）
	viper.SetConfigType("yaml")                     // 配置文件格式为YAML
	viper.AddConfigPath(".")                        // 在当前目录查找配置文件
	viper.AddConfigPath("./config")                 // 在config子目录查找配置文件

	// 设置中断处理相关的默认配置值
	viper.SetDefault("interrupt.max_retries", 3)        // 默认最大重试次数为3次
	viper.SetDefault("interrupt.retry_delay", "2s")     // 默认重试间隔为2秒
	viper.SetDefault("interrupt.enable_approval", true) // 默认启用人工审批功能

	// 设置业务规则相关的默认配置值
	viper.SetDefault("business.risk_threshold", 10000.0) // 默认风险评估阈值为10000元
	viper.SetDefault("business.inventory_threshold", 10) // 默认库存不足阈值为10件
	viper.SetDefault("business.payment_retry_limit", 3)  // 默认支付重试限制为3次

	// 尝试读取配置文件，如果失败则使用默认配置
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("配置文件读取失败，使用默认配置: %v", err) // 记录配置文件读取失败信息
	} else {
		log.Printf("配置文件加载成功: %s", viper.ConfigFileUsed()) // 记录成功加载的配置文件路径
	}
}

// RunDynamicInterruptDemo 运行动态中断机制演示
// 这是动态中断演示的主入口函数，展示各种动态中断处理场景
// 包括基本动态中断处理和条件动态中断处理的完整演示流程
func main() {
	initDynamicConfig()
	ctx := context.Background()

	// 运行动态中断演示
	dynamicInterruptDemo(ctx)
	dynamicConditionalInterruptDemo(ctx)

	fmt.Println("\n🎉 动态中断演示完成！")
}
