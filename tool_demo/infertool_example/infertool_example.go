// package main 表明这是一个可执行程序
package main

import (
	"context"
	"fmt"
	"log"
	"net/mail"
	"regexp"
	"strings"
	"time"

	// 导入 InferTool 所在的包
	"github.com/cloudwego/eino/components/tool/utils"
)

// =============================================================================
//
//  文件: infertool_example.go
//  功能: 演示如何使用 InferTool 自动从 Go 结构体标签推断工具的 schema 信息。
//  说明: InferTool 是创建 Eino 工具最简洁、最高效的方式。它通过反射读取结构体的
//        `jsonschema` 标签，自动生成工具的参数定义，极大地减少了手动编写
//        `Info()` 方法的模板代码。
//
// =============================================================================

// --- 示例 1: 用户管理工具 ---

// CreateUserRequest 定义了创建用户工具的输入参数。
// `jsonschema` 标签是 InferTool 的核心，它被用来自动生成工具的参数 schema。
// - `required`: 标记该字段为必填项。
// - `description`: 字段的描述，会显示在工具的文档中。
// - `format=email`: 指定字段的格式，这里是邮箱格式。
// - `minimum=0, maximum=120`: 指定数值范围。
// - `enum=male,enum=female,enum=other`: 定义枚举值，限制输入选项。
// - `default=true`: 为字段提供默认值。
type CreateUserRequest struct {
	Name     string `json:"name" jsonschema:"required,description=用户姓名"`
	Email    string `json:"email" jsonschema:"required,format=email,description=用户邮箱地址"`
	Age      int    `json:"age" jsonschema:"minimum=0,maximum=120,description=用户年龄"`
	Gender   string `json:"gender" jsonschema:"enum=male,enum=female,enum=other,description=用户性别"`
	City     string `json:"city" jsonschema:"description=所在城市"`
	IsActive bool   `json:"is_active" jsonschema:"description=账户是否激活,default=true"`
}

// CreateUserResponse 定义了创建用户工具的输出结构。
// InferTool 会自动处理将此结构体序列化为 JSON 字符串作为工具的返回结果。
type CreateUserResponse struct {
	Success   bool   `json:"success"`
	UserID    string `json:"user_id"`
	Message   string `json:"message"`
	CreatedAt string `json:"created_at"`
}

// createUser 是实际执行创建用户逻辑的业务函数。
// 函数签名必须是 `func(context.Context, *Request) (*Response, error)` 的形式，
// 其中 Request 和 Response 分别是输入和输出的结构体指针。
func createUser(ctx context.Context, req *CreateUserRequest) (*CreateUserResponse, error) {
	log.Printf("[CreateUser] 正在创建用户: %s (%s)", req.Name, req.Email)

	// 简单的业务逻辑验证：验证邮箱格式
	if _, err := mail.ParseAddress(req.Email); err != nil {
		return &CreateUserResponse{
			Success: false,
			Message: "邮箱格式不正确",
		}, nil // 对于业务逻辑错误，通常返回 nil error，在响应体中指明错误信息
	}

	// 模拟业务处理，例如数据库插入操作
	userID := fmt.Sprintf("user_%d", time.Now().Unix())

	// 返回成功的响应
	return &CreateUserResponse{
		Success:   true,
		UserID:    userID,
		Message:   fmt.Sprintf("用户 %s 创建成功", req.Name),
		CreatedAt: time.Now().Format(time.RFC3339),
	}, nil
}

// --- 示例 2: 订单管理工具 ---

// CalculateOrderRequest 定义了计算订单价格工具的输入参数。
// 它包含了一个嵌套的结构体切片 `[]OrderItem`。
type CalculateOrderRequest struct {
	Items       []OrderItem `json:"items" jsonschema:"required,description=订单商品列表"`
	CouponCode  string      `json:"coupon_code" jsonschema:"description=优惠券代码"`
	ShippingFee float64     `json:"shipping_fee" jsonschema:"minimum=0,description=运费,default=10"`
	TaxRate     float64     `json:"tax_rate" jsonschema:"minimum=0,maximum=1,description=税率,default=0.1"`
}

// OrderItem 定义了订单中的单个商品项。
// InferTool 能够递归地解析嵌套结构体，并为它们生成正确的 schema。
type OrderItem struct {
	ProductID string  `json:"product_id" jsonschema:"required,description=商品ID"`
	Name      string  `json:"name" jsonschema:"required,description=商品名称"`
	Price     float64 `json:"price" jsonschema:"required,minimum=0,description=单价"`
	Quantity  int     `json:"quantity" jsonschema:"required,minimum=1,description=数量"`
}

// CalculateOrderResponse 定义了计算订单价格工具的输出。
type CalculateOrderResponse struct {
	SubTotal      float64     `json:"sub_total"`      // 商品总价
	Discount      float64     `json:"discount"`       // 折扣金额
	ShippingFee   float64     `json:"shipping_fee"`   // 运费
	Tax           float64     `json:"tax"`            // 税费
	Total         float64     `json:"total"`          // 最终总价
	Items         []OrderItem `json:"items"`          // 订单项详情
	CouponApplied bool        `json:"coupon_applied"` // 是否成功应用优惠券
}

// calculateOrder 是计算订单总价的业务函数。
func calculateOrder(ctx context.Context, req *CalculateOrderRequest) (*CalculateOrderResponse, error) {
	log.Printf("[CalculateOrder] 正在计算订单，商品数量: %d", len(req.Items))

	// 计算商品小计
	var subTotal float64
	for _, item := range req.Items {
		subTotal += item.Price * float64(item.Quantity)
	}

	// 应用简单的优惠券逻辑
	var discount float64
	couponApplied := false
	if req.CouponCode == "SAVE10" { // 假设 "SAVE10" 是一个有效的优惠券
		discount = subTotal * 0.1
		couponApplied = true
	}

	// 计算税费和总价
	taxableAmount := subTotal - discount + req.ShippingFee
	tax := taxableAmount * req.TaxRate
	total := taxableAmount + tax

	// 返回包含所有计算细节的响应
	return &CalculateOrderResponse{
		SubTotal:      subTotal,
		Discount:      discount,
		ShippingFee:   req.ShippingFee,
		Tax:           tax,
		Total:         total,
		Items:         req.Items,
		CouponApplied: couponApplied,
	}, nil
}

// --- 示例 3: 数据分析工具 ---

// AnalyzeDataRequest 定义了数据分析工具的输入。
// `items={enum=[...]}` 用于定义字符串切片中每个元素的有效值。
type AnalyzeDataRequest struct {
	Dataset      []float64 `json:"dataset" jsonschema:"required,description=要分析的数据集"`
	Operations   []string  `json:"operations" jsonschema:"required,description=分析操作列表,items={enum=[mean,median,mode,std,min,max]}"`
	Precision    int       `json:"precision" jsonschema:"minimum=0,maximum=10,description=小数点精度,default=2"`
	IncludeChart bool      `json:"include_chart" jsonschema:"description=是否包含图表信息,default=false"`
}

// AnalyzeDataResponse 定义了数据分析工具的输出，包含多个嵌套结构。
type AnalyzeDataResponse struct {
	Results     map[string]interface{} `json:"results"`              // 存储各种计算结果
	DataSummary DataSummary            `json:"data_summary"`         // 数据的摘要信息
	ChartInfo   *ChartInfo             `json:"chart_info,omitempty"` // 图表信息，`omitempty` 表示如果为空则在JSON中省略
}

// DataSummary 提供了数据集的摘要统计。
type DataSummary struct {
	Count    int     `json:"count"`
	Min      float64 `json:"min"`
	Max      float64 `json:"max"`
	Range    float64 `json:"range"`
	DataType string  `json:"data_type"`
}

// ChartInfo 包含了用于生成图表的建议信息。
type ChartInfo struct {
	RecommendedType string    `json:"recommended_type"` // 建议的图表类型
	XAxis           []string  `json:"x_axis"`           // X轴数据
	YAxis           []float64 `json:"y_axis"`           // Y轴数据
}

// analyzeData 是执行数据分析的业务函数。
func analyzeData(ctx context.Context, req *AnalyzeDataRequest) (*AnalyzeDataResponse, error) {
	log.Printf("[AnalyzeData] 正在分析数据集，数据点数量: %d", len(req.Dataset))

	if len(req.Dataset) == 0 {
		return nil, fmt.Errorf("数据集不能为空") // 对于系统级错误，返回 non-nil error
	}

	results := make(map[string]interface{})

	// 执行请求的分析操作
	for _, op := range req.Operations {
		switch op {
		case "mean":
			sum := 0.0
			for _, v := range req.Dataset {
				sum += v
			}
			results[op] = roundTo(sum/float64(len(req.Dataset)), req.Precision)
		case "min":
			min := req.Dataset[0]
			for _, v := range req.Dataset[1:] {
				if v < min {
					min = v
				}
			}
			results[op] = min
		case "max":
			max := req.Dataset[0]
			for _, v := range req.Dataset[1:] {
				if v > max {
					max = v
				}
			}
			results[op] = max
			// 其他操作（median, mode, std）在此省略以保持示例简洁
		}
	}

	// 生成数据摘要
	min, max := req.Dataset[0], req.Dataset[0]
	for _, v := range req.Dataset {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}

	summary := DataSummary{
		Count:    len(req.Dataset),
		Min:      min,
		Max:      max,
		Range:    max - min,
		DataType: "numerical",
	}

	response := &AnalyzeDataResponse{
		Results:     results,
		DataSummary: summary,
	}

	// 如果请求包含图表信息，则生成它
	if req.IncludeChart {
		xAxis := make([]string, len(req.Dataset))
		for i := range req.Dataset {
			xAxis[i] = fmt.Sprintf("Point %d", i+1)
		}

		response.ChartInfo = &ChartInfo{
			RecommendedType: "line", // 推荐使用折线图
			XAxis:           xAxis,
			YAxis:           req.Dataset,
		}
	}

	return response, nil
}

// --- 示例 4: 文本处理工具 ---

// ProcessTextRequest 定义了文本处理工具的输入。
type ProcessTextRequest struct {
	Text          string   `json:"text" jsonschema:"required,description=要处理的文本"`
	Operations    []string `json:"operations" jsonschema:"required,description=处理操作,items={enum=[word_count,char_count,sentence_count,extract_emails,extract_urls,sentiment]}"`
	Language      string   `json:"language" jsonschema:"enum=zh,enum=en,description=文本语言,default=zh"`
	CaseSensitive bool     `json:"case_sensitive" jsonschema:"description=是否大小写敏感,default=false"`
}

// ProcessTextResponse 定义了文本处理工具的输出。
type ProcessTextResponse struct {
	OriginalText string                 `json:"original_text"`
	Results      map[string]interface{} `json:"results"`
	Language     string                 `json:"language"`
	ProcessedAt  string                 `json:"processed_at"`
}

// processText 是执行文本处理的业务函数。
func processText(ctx context.Context, req *ProcessTextRequest) (*ProcessTextResponse, error) {
	log.Printf("[ProcessText] 正在处理文本，长度: %d", len(req.Text))

	results := make(map[string]interface{})
	text := req.Text
	// 根据是否大小写敏感，预处理文本
	if !req.CaseSensitive {
		text = strings.ToLower(text)
	}

	for _, op := range req.Operations {
		switch op {
		case "word_count":
			results[op] = len(strings.Fields(text))
		case "char_count":
			results[op] = len([]rune(text)) // 使用 []rune 以正确处理多字节字符
		case "sentence_count":
			// 简化的句子计数逻辑
			var sentences []string
			if req.Language == "en" {
				sentences = strings.Split(text, ".")
			} else {
				sentences = strings.Split(text, "。")
			}
			count := 0
			for _, s := range sentences {
				if strings.TrimSpace(s) != "" {
					count++
				}
			}
			results[op] = count
		case "extract_emails":
			emailRegex := regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
			results[op] = emailRegex.FindAllString(req.Text, -1) // 在原始文本中查找
		case "extract_urls":
			urlRegex := regexp.MustCompile(`https?://[^\s]+`)
			results[op] = urlRegex.FindAllString(req.Text, -1) // 在原始文本中查找
		case "sentiment":
			// 极其简化的情感分析模拟
			positiveWords := []string{"好", "棒", "优秀", "喜欢", "happy", "good", "excellent"}
			negativeWords := []string{"坏", "差", "糟糕", "讨厌", "bad", "terrible", "hate"}
			score := 0
			for _, word := range positiveWords {
				if strings.Contains(text, word) {
					score++
				}
			}
			for _, word := range negativeWords {
				if strings.Contains(text, word) {
					score--
				}
			}
			sentiment := "neutral"
			if score > 0 {
				sentiment = "positive"
			} else if score < 0 {
				sentiment = "negative"
			}
			results[op] = map[string]interface{}{"sentiment": sentiment, "score": score}
		}
	}

	return &ProcessTextResponse{
		OriginalText: req.Text,
		Results:      results,
		Language:     req.Language,
		ProcessedAt:  time.Now().Format(time.RFC3339),
	}, nil
}

// --- 辅助函数 ---

// roundTo 将浮点数四舍五入到指定的小数位数。
func roundTo(value float64, precision int) float64 {
	factor := 1.0
	for i := 0; i < precision; i++ {
		factor *= 10
	}
	return float64(int(value*factor+0.5)) / factor
}

// --- 演示函数 ---

// demonstrateInferTool 演示了如何使用 `utils.InferTool` 来创建和调用工具。
func demonstrateInferTool() {
	ctx := context.Background()

	fmt.Println("=== InferTool 自动推断演示 ===")

	// 1. 使用 InferTool 创建用户管理工具
	// 只需提供工具名称、描述和业务函数，InferTool 会自动完成剩下的工作。
	fmt.Println("--- 1. 创建用户管理工具 ---")
	// TODO 将本地业务函数 `createUser` 传递给 InferTool，InferTool 会自动生成工具的 schema 和 Info 方法。
	userTool, err := utils.InferTool("create_user", "创建一个新的用户账户", createUser)
	if err != nil {
		log.Fatalf("创建用户工具失败: %v", err)
	}
	// 打印自动生成的工具信息
	toolInfo, _ := userTool.Info(ctx)
	fmt.Printf("工具名称: %s\n", toolInfo.Name)
	fmt.Printf("工具描述: %s\n", toolInfo.Desc)

	// 调用工具，传入 JSON 字符串作为参数
	userResult, err := userTool.InvokableRun(ctx, `{
		"name": "张三",
		"email": "zhangsan@example.com",
		"age": 28,
		"gender": "male",
		"city": "北京"
	}`)
	if err != nil {
		log.Printf("用户工具执行失败: %v", err)
	} else {
		fmt.Printf("执行结果: %s\n\n", userResult)
	}

	// 2. 使用 InferTool 创建订单计算工具
	fmt.Println("--- 2. 创建订单计算工具 ---")
	orderTool, err := utils.InferTool("calculate_order", "根据商品列表、优惠券和运费计算订单总价", calculateOrder)
	if err != nil {
		log.Fatalf("创建订单工具失败: %v", err)
	}
	orderResult, err := orderTool.InvokableRun(ctx, `{
		"items": [
			{"product_id": "P001", "name": "苹果", "price": 5.99, "quantity": 3},
			{"product_id": "P002", "name": "香蕉", "price": 3.99, "quantity": 2}
		],
		"coupon_code": "SAVE10",
		"shipping_fee": 15.0,
		"tax_rate": 0.08
	}`)
	if err != nil {
		log.Printf("订单工具执行失败: %v", err)
	} else {
		fmt.Printf("计算结果: %s\n\n", orderResult)
	}

	// 3. 使用 InferTool 创建数据分析工具
	fmt.Println("--- 3. 创建数据分析工具 ---")
	analysisTool, err := utils.InferTool("analyze_data", "对给定的数值数据集进行统计分析", analyzeData)
	if err != nil {
		log.Fatalf("创建分析工具失败: %v", err)
	}
	analysisResult, err := analysisTool.InvokableRun(ctx, `{
		"dataset": [1.2, 3.4, 5.6, 7.8, 9.0, 2.1, 4.3, 6.5],
		"operations": ["mean", "min", "max"],
		"precision": 3,
		"include_chart": true
	}`)
	if err != nil {
		log.Printf("分析工具执行失败: %v", err)
	} else {
		fmt.Printf("分析结果: %s\n\n", analysisResult)
	}

	// 4. 使用 InferTool 创建文本处理工具
	fmt.Println("--- 4. 创建文本处理工具 ---")
	textTool, err := utils.InferTool("process_text", "对输入文本进行多种处理和分析", processText)
	if err != nil {
		log.Fatalf("创建文本工具失败: %v", err)
	}
	textResult, err := textTool.InvokableRun(ctx, `{
		"text": "这是一个好的例子。联系我: user@example.com 或访问 https://example.com",
		"operations": ["word_count", "char_count", "extract_emails", "extract_urls", "sentiment"],
		"language": "zh"
	}`)
	if err != nil {
		log.Printf("文本工具执行失败: %v", err)
	} else {
		fmt.Printf("处理结果: %s\n\n", textResult)
	}
}

// main 是程序的入口点。
func main() {
	// 调用演示函数来执行所有 InferTool 的示例
	demonstrateInferTool()
}
