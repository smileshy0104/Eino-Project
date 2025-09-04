package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/spf13/viper"
)

// 业务数据结构
type BusinessData struct {
	OrderID     string                 `json:"order_id"`
	Amount      float64                `json:"amount"`
	Status      string                 `json:"status"`
	Items       []string               `json:"items"`
	Metadata    map[string]interface{} `json:"metadata"`
	ProcessStep int                    `json:"process_step"`
}

// 自定义动态中断错误
type DynamicInterruptError struct {
	Message          string
	RequiresApproval bool
	InterruptReason  string
	Metadata         map[string]interface{}
	Timestamp        string
}

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

// 业务步骤处理函数类型
type BusinessStep func(ctx context.Context, data BusinessData) (BusinessData, error)

// 风险评估节点 - 可能触发动态中断
func createRiskAssessmentNode() BusinessStep {
	return func(ctx context.Context, data BusinessData) (BusinessData, error) {
		fmt.Printf("🔍 执行风险评估，订单ID: %s, 金额: %.2f\n", data.OrderID, data.Amount)

		data.ProcessStep = 1
		data.Status = "风险评估中"

		// 模拟风险评估逻辑
		riskScore := rand.Float64() * 100

		if data.Metadata == nil {
			data.Metadata = make(map[string]interface{})
		}
		data.Metadata["risk_score"] = riskScore

		fmt.Printf("📊 风险分数: %.2f\n", riskScore)

		// 高风险订单需要人工审核
		if riskScore > 80 {
			fmt.Printf("⚠️  高风险订单，需要人工审核\n")
			return data, NewDynamicInterruptError("高风险订单检测", true)
		}

		// 中等风险订单需要额外验证
		if riskScore > 60 {
			fmt.Printf("⚠️  中等风险订单，需要额外验证\n")
			return data, NewDynamicInterruptError("中等风险订单检测", false)
		}

		data.Status = "风险评估通过"
		fmt.Printf("✅ 低风险订单，自动通过\n")
		return data, nil
	}
}

// 库存检查节点 - 可能触发动态中断
func createInventoryCheckNode() BusinessStep {
	return func(ctx context.Context, data BusinessData) (BusinessData, error) {
		fmt.Printf("📦 检查库存，订单项目: %v\n", data.Items)

		data.ProcessStep = 2
		data.Status = "库存检查中"

		// 模拟库存检查
		for i, item := range data.Items {
			stock := rand.Intn(100)
			data.Metadata[fmt.Sprintf("stock_%d", i)] = stock

			fmt.Printf("📊 商品 %s 库存: %d\n", item, stock)

			// 库存不足需要中断
			if stock < 10 {
				fmt.Printf("⚠️  商品 %s 库存不足，需要补货\n", item)
				return data, NewDynamicInterruptError(fmt.Sprintf("商品 %s 库存不足", item), false)
			}
		}

		data.Status = "库存充足"
		fmt.Printf("✅ 所有商品库存充足\n")
		return data, nil
	}
}

// 支付处理节点 - 可能触发动态中断
func createPaymentProcessNode() BusinessStep {
	return func(ctx context.Context, data BusinessData) (BusinessData, error) {
		fmt.Printf("💳 处理支付，订单金额: %.2f\n", data.Amount)

		data.ProcessStep = 3
		data.Status = "支付处理中"

		// 模拟支付处理
		paymentSuccess := rand.Float64() > 0.3 // 70% 成功率

		if !paymentSuccess {
			fmt.Printf("❌ 支付失败，需要重试\n")
			return data, NewDynamicInterruptError("支付处理失败", false)
		}

		// 大额支付需要额外确认
		if data.Amount > 10000 {
			fmt.Printf("⚠️  大额支付，需要额外确认\n")
			return data, NewDynamicInterruptError("大额支付确认", true)
		}

		data.Status = "支付成功"
		data.Metadata["payment_time"] = time.Now().Format(time.RFC3339)
		fmt.Printf("✅ 支付处理成功\n")
		return data, nil
	}
}

// 订单完成节点
func createOrderCompletionNode() BusinessStep {
	return func(ctx context.Context, data BusinessData) (BusinessData, error) {
		fmt.Printf("🎉 完成订单处理，订单ID: %s\n", data.OrderID)

		data.ProcessStep = 4
		data.Status = "订单完成"
		data.Metadata["completion_time"] = time.Now().Format(time.RFC3339)

		return data, nil
	}
}

// 动态中断处理器
type DynamicInterruptHandler struct {
	maxRetries int
}

func NewDynamicInterruptHandler(maxRetries int) *DynamicInterruptHandler {
	return &DynamicInterruptHandler{maxRetries: maxRetries}
}

// 处理业务流程，支持动态中断和重试
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

// 处理具体的中断情况
func (dih *DynamicInterruptHandler) handleInterrupt(interruptErr *DynamicInterruptError) bool {
	fmt.Printf("🔧 处理中断: %s\n", interruptErr.InterruptReason)

	// 根据中断类型进行不同的处理
	switch {
	case interruptErr.RequiresApproval:
		fmt.Println("👤 需要人工审核...")
		time.Sleep(1 * time.Second) // 模拟人工审核时间
		fmt.Println("✅ 人工审核通过")
		return true

	case interruptErr.InterruptReason == "中等风险订单检测":
		fmt.Println("🔍 执行额外风险验证...")
		time.Sleep(800 * time.Millisecond)
		fmt.Println("✅ 额外验证通过")
		return true

	case contains(interruptErr.InterruptReason, "库存不足"):
		fmt.Println("📦 正在补货...")
		time.Sleep(1200 * time.Millisecond)
		fmt.Println("✅ 库存已补充")
		return true

	case interruptErr.InterruptReason == "支付处理失败":
		fmt.Println("💳 重新处理支付...")
		time.Sleep(500 * time.Millisecond)
		fmt.Println("✅ 支付处理就绪")
		return true

	default:
		fmt.Printf("❌ 未知的中断类型: %s\n", interruptErr.InterruptReason)
		return false
	}
}

// 辅助函数
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr ||
		len(s) > len(substr) && (s[len(s)-len(substr):] == substr)
}

// 动态中断演示
func dynamicInterruptDemo(ctx context.Context) {
	fmt.Println("=== 动态中断演示 ===")

	// 设置随机种子
	rand.Seed(time.Now().UnixNano())

	// 创建业务处理步骤
	steps := []BusinessStep{
		createRiskAssessmentNode(),
		createInventoryCheckNode(),
		createPaymentProcessNode(),
		createOrderCompletionNode(),
	}

	// 创建中断处理器
	handler := NewDynamicInterruptHandler(5)

	// 准备测试订单
	orders := []BusinessData{
		{
			OrderID:  "ORD-001",
			Amount:   500.00,
			Items:    []string{"手机", "充电器"},
			Status:   "待处理",
			Metadata: make(map[string]interface{}),
		},
		{
			OrderID:  "ORD-002",
			Amount:   15000.00,
			Items:    []string{"笔记本电脑", "鼠标", "键盘"},
			Status:   "待处理",
			Metadata: make(map[string]interface{}),
		},
		{
			OrderID:  "ORD-003",
			Amount:   200.00,
			Items:    []string{"书籍", "文具"},
			Status:   "待处理",
			Metadata: make(map[string]interface{}),
		},
	}

	// 处理每个订单
	for i, order := range orders {
		fmt.Printf("\n📝 开始处理订单 %d: %s\n", i+1, order.OrderID)

		result, err := handler.processBusinessWorkflow(ctx, order, steps)

		if err != nil {
			fmt.Printf("❌ 订单 %s 处理失败: %v\n", order.OrderID, err)
		} else {
			fmt.Printf("✅ 订单 %s 处理成功！状态: %s\n", result.OrderID, result.Status)
		}
	}
}

// 条件中断演示
func conditionalInterruptDemo(ctx context.Context) {
	fmt.Println("\n=== 条件中断演示 ===")

	// 创建条件检查步骤
	conditionalStep := func(ctx context.Context, data BusinessData) (BusinessData, error) {
		fmt.Printf("🔍 执行条件检查，订单ID: %s\n", data.OrderID)

		data.ProcessStep += 1

		// 多种条件检查
		conditions := []struct {
			name    string
			check   func(BusinessData) bool
			message string
			urgent  bool
		}{
			{
				name:    "金额验证",
				check:   func(d BusinessData) bool { return d.Amount <= 0 },
				message: "订单金额无效",
				urgent:  true,
			},
			{
				name:    "商品验证",
				check:   func(d BusinessData) bool { return len(d.Items) == 0 },
				message: "订单商品列表为空",
				urgent:  true,
			},
			{
				name:    "高价值检查",
				check:   func(d BusinessData) bool { return d.Amount > 20000 },
				message: "超高价值订单需要特殊处理",
				urgent:  false,
			},
			{
				name:    "商品数量检查",
				check:   func(d BusinessData) bool { return len(d.Items) > 10 },
				message: "商品种类过多需要分批处理",
				urgent:  false,
			},
		}

		for _, condition := range conditions {
			if condition.check(data) {
				fmt.Printf("⚠️  条件 '%s' 触发: %s\n", condition.name, condition.message)

				// 设置中断元数据
				data.Metadata["interrupt_condition"] = condition.name
				data.Metadata["urgent"] = condition.urgent

				return data, NewDynamicInterruptError(condition.message, condition.urgent)
			}
		}

		fmt.Println("✅ 所有条件检查通过")
		data.Status = "条件验证通过"
		return data, nil
	}

	handler := NewDynamicInterruptHandler(3)

	// 测试不同条件的订单
	testCases := []BusinessData{
		{OrderID: "TEST-001", Amount: -100, Items: []string{"商品1"}},    // 金额无效
		{OrderID: "TEST-002", Amount: 100, Items: []string{}},          // 商品为空
		{OrderID: "TEST-003", Amount: 25000, Items: []string{"奢侈品"}},   // 高价值
		{OrderID: "TEST-004", Amount: 1000, Items: make([]string, 15)}, // 商品过多
		{OrderID: "TEST-005", Amount: 100, Items: []string{"正常商品"}},    // 正常订单
	}

	for _, testCase := range testCases {
		fmt.Printf("\n📝 测试订单: %s (金额: %.2f, 商品数: %d)\n",
			testCase.OrderID, testCase.Amount, len(testCase.Items))

		testCase.Metadata = make(map[string]interface{})

		result, err := conditionalStep(ctx, testCase)

		if err != nil {
			if interruptErr, ok := err.(*DynamicInterruptError); ok {
				fmt.Printf("⏸️  按预期中断: %s\n", interruptErr.InterruptReason)

				if interruptErr.RequiresApproval {
					fmt.Println("🚨 紧急情况，需要立即处理")
				} else {
					fmt.Println("⏳ 非紧急情况，可以稍后处理")
				}

				// 尝试处理中断
				handled := handler.handleInterrupt(interruptErr)
				if handled {
					fmt.Printf("✅ 中断已处理，订单可以继续\n")
				} else {
					fmt.Printf("❌ 中断无法处理\n")
				}
			} else {
				fmt.Printf("❌ 意外错误: %v\n", err)
			}
		} else {
			fmt.Printf("✅ 订单处理成功: %s\n", result.Status)
		}
	}
}

func initDynamicConfig() {
	viper.SetConfigFile("../config.yaml")
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("读取配置文件失败: %v (使用默认配置)", err)
	}
}

func main() {
	initDynamicConfig()
	ctx := context.Background()

	// 运行动态中断演示
	dynamicInterruptDemo(ctx)
	conditionalInterruptDemo(ctx)

	fmt.Println("\n🎉 动态中断演示完成！")
}
