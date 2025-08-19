package main

import (
	"context"
	"fmt"
	"log"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/tool/utils"
)

// =============================================================================
//
//  文件: infertool_example.go
//  功能: 演示如何使用 InferTool 自动从结构体标签推断工具信息
//  说明: InferTool 是最简洁的工具创建方式，通过结构体标签自动生成工具描述
//
// =============================================================================

// --- 示例 1: 用户管理工具 ---

// CreateUserRequest 创建用户的请求参数，使用 jsonschema 标签定义约束
type CreateUserRequest struct {
	Name     string `json:"name" jsonschema:"required,description=用户姓名"`
	Email    string `json:"email" jsonschema:"required,format=email,description=用户邮箱地址"`
	Age      int    `json:"age" jsonschema:"minimum=0,maximum=120,description=用户年龄"`
	Gender   string `json:"gender" jsonschema:"enum=male,enum=female,enum=other,description=用户性别"`
	City     string `json:"city" jsonschema:"description=所在城市"`
	IsActive bool   `json:"is_active" jsonschema:"description=账户是否激活,default=true"`
}

// CreateUserResponse 创建用户的响应
type CreateUserResponse struct {
	Success   bool   `json:"success"`
	UserID    string `json:"user_id"`
	Message   string `json:"message"`
	CreatedAt string `json:"created_at"`
}

// createUser 创建用户的业务函数
func createUser(ctx context.Context, req *CreateUserRequest) (*CreateUserResponse, error) {
	log.Printf("[CreateUser] 创建用户: %s (%s)", req.Name, req.Email)

	// 验证邮箱格式
	if _, err := mail.ParseAddress(req.Email); err != nil {
		return &CreateUserResponse{
			Success: false,
			Message: "邮箱格式不正确",
		}, nil
	}

	// 模拟生成用户ID
	userID := fmt.Sprintf("user_%d", time.Now().Unix())

	return &CreateUserResponse{
		Success:   true,
		UserID:    userID,
		Message:   fmt.Sprintf("用户 %s 创建成功", req.Name),
		CreatedAt: time.Now().Format(time.RFC3339),
	}, nil
}

// --- 示例 2: 订单管理工具 ---

// CalculateOrderRequest 计算订单的请求参数
type CalculateOrderRequest struct {
	Items       []OrderItem `json:"items" jsonschema:"required,description=订单商品列表"`
	CouponCode  string      `json:"coupon_code" jsonschema:"description=优惠券代码"`
	ShippingFee float64     `json:"shipping_fee" jsonschema:"minimum=0,description=运费,default=10"`
	TaxRate     float64     `json:"tax_rate" jsonschema:"minimum=0,maximum=1,description=税率,default=0.1"`
}

// OrderItem 订单商品项
type OrderItem struct {
	ProductID string  `json:"product_id" jsonschema:"required,description=商品ID"`
	Name      string  `json:"name" jsonschema:"required,description=商品名称"`
	Price     float64 `json:"price" jsonschema:"required,minimum=0,description=单价"`
	Quantity  int     `json:"quantity" jsonschema:"required,minimum=1,description=数量"`
}

// CalculateOrderResponse 计算订单的响应
type CalculateOrderResponse struct {
	SubTotal      float64     `json:"sub_total"`
	Discount      float64     `json:"discount"`
	ShippingFee   float64     `json:"shipping_fee"`
	Tax           float64     `json:"tax"`
	Total         float64     `json:"total"`
	Items         []OrderItem `json:"items"`
	CouponApplied bool        `json:"coupon_applied"`
}

// calculateOrder 计算订单总价的业务函数
func calculateOrder(ctx context.Context, req *CalculateOrderRequest) (*CalculateOrderResponse, error) {
	log.Printf("[CalculateOrder] 计算订单，商品数量: %d", len(req.Items))

	var subTotal float64
	for _, item := range req.Items {
		subTotal += item.Price * float64(item.Quantity)
	}

	// 简单的优惠券逻辑
	var discount float64
	couponApplied := false
	if req.CouponCode == "SAVE10" {
		discount = subTotal * 0.1
		couponApplied = true
	}

	taxableAmount := subTotal - discount + req.ShippingFee
	tax := taxableAmount * req.TaxRate
	total := taxableAmount + tax

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

// AnalyzeDataRequest 数据分析请求参数
type AnalyzeDataRequest struct {
	Dataset      []float64 `json:"dataset" jsonschema:"required,description=要分析的数据集"`
	Operations   []string  `json:"operations" jsonschema:"required,description=分析操作列表,items={enum=[mean,median,mode,std,min,max]}"`
	Precision    int       `json:"precision" jsonschema:"minimum=0,maximum=10,description=小数点精度,default=2"`
	IncludeChart bool      `json:"include_chart" jsonschema:"description=是否包含图表信息,default=false"`
}

// AnalyzeDataResponse 数据分析响应
type AnalyzeDataResponse struct {
	Results     map[string]interface{} `json:"results"`
	DataSummary DataSummary            `json:"data_summary"`
	ChartInfo   *ChartInfo             `json:"chart_info,omitempty"`
}

// DataSummary 数据摘要
type DataSummary struct {
	Count    int     `json:"count"`
	Min      float64 `json:"min"`
	Max      float64 `json:"max"`
	Range    float64 `json:"range"`
	DataType string  `json:"data_type"`
}

// ChartInfo 图表信息
type ChartInfo struct {
	RecommendedType string    `json:"recommended_type"`
	XAxis           []string  `json:"x_axis"`
	YAxis           []float64 `json:"y_axis"`
}

// analyzeData 数据分析业务函数
func analyzeData(ctx context.Context, req *AnalyzeDataRequest) (*AnalyzeDataResponse, error) {
	log.Printf("[AnalyzeData] 分析数据集，数据点数量: %d", len(req.Dataset))

	if len(req.Dataset) == 0 {
		return nil, fmt.Errorf("数据集不能为空")
	}

	results := make(map[string]interface{})

	// 执行各种分析操作
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
		}
	}

	// 生成数据摘要
	min := req.Dataset[0]
	max := req.Dataset[0]
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

	// 如果需要图表信息
	if req.IncludeChart {
		xAxis := make([]string, len(req.Dataset))
		for i := range req.Dataset {
			xAxis[i] = fmt.Sprintf("Point %d", i+1)
		}

		response.ChartInfo = &ChartInfo{
			RecommendedType: "line",
			XAxis:           xAxis,
			YAxis:           req.Dataset,
		}
	}

	return response, nil
}

// --- 示例 4: 文本处理工具 ---

// ProcessTextRequest 文本处理请求
type ProcessTextRequest struct {
	Text          string   `json:"text" jsonschema:"required,description=要处理的文本"`
	Operations    []string `json:"operations" jsonschema:"required,description=处理操作,items={enum=[word_count,char_count,sentence_count,extract_emails,extract_urls,sentiment]}"`
	Language      string   `json:"language" jsonschema:"enum=zh,enum=en,description=文本语言,default=zh"`
	CaseSensitive bool     `json:"case_sensitive" jsonschema:"description=是否大小写敏感,default=false"`
}

// ProcessTextResponse 文本处理响应
type ProcessTextResponse struct {
	OriginalText string                 `json:"original_text"`
	Results      map[string]interface{} `json:"results"`
	Language     string                 `json:"language"`
	ProcessedAt  string                 `json:"processed_at"`
}

// processText 文本处理业务函数
func processText(ctx context.Context, req *ProcessTextRequest) (*ProcessTextResponse, error) {
	log.Printf("[ProcessText] 处理文本，长度: %d", len(req.Text))

	results := make(map[string]interface{})

	text := req.Text
	if !req.CaseSensitive {
		text = strings.ToLower(text)
	}

	for _, op := range req.Operations {
		switch op {
		case "word_count":
			words := strings.Fields(text)
			results[op] = len(words)
		case "char_count":
			results[op] = len(text)
		case "sentence_count":
			// 简单的句子计数
			sentences := strings.Split(text, "。")
			if req.Language == "en" {
				sentences = strings.Split(text, ".")
			}
			results[op] = len(sentences) - 1 // 减去最后一个空元素
		case "extract_emails":
			emailRegex := regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
			emails := emailRegex.FindAllString(text, -1)
			results[op] = emails
		case "extract_urls":
			urlRegex := regexp.MustCompile(`https?://[^\s]+`)
			urls := urlRegex.FindAllString(text, -1)
			results[op] = urls
		case "sentiment":
			// 简单的情感分析模拟
			positiveWords := []string{"好", "棒", "优秀", "喜欢", "happy", "good", "excellent"}
			negativeWords := []string{"坏", "差", "糟糕", "讨厌", "bad", "terrible", "hate"}

			score := 0
			for _, word := range positiveWords {
				if strings.Contains(text, word) {
					score += 1
				}
			}
			for _, word := range negativeWords {
				if strings.Contains(text, word) {
					score -= 1
				}
			}

			sentiment := "neutral"
			if score > 0 {
				sentiment = "positive"
			} else if score < 0 {
				sentiment = "negative"
			}

			results[op] = map[string]interface{}{
				"sentiment": sentiment,
				"score":     score,
			}
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

// roundTo 将浮点数四舍五入到指定精度
func roundTo(value float64, precision int) float64 {
	factor := 1.0
	for i := 0; i < precision; i++ {
		factor *= 10
	}
	return float64(int(value*factor+0.5)) / factor
}

// --- 演示函数 ---

func demonstrateInferTool() {
	ctx := context.Background()

	fmt.Println("=== InferTool 自动推断演示 ===\n")

	// 1. 使用 InferTool 创建用户管理工具
	fmt.Println("--- 创建用户管理工具 ---")
	userTool, err := utils.InferTool("create_user", "创建新用户账户", createUser)
	if err != nil {
		log.Printf("创建用户工具失败: %v", err)
	} else {
		// 获取工具信息
		toolInfo, _ := userTool.Info(ctx)
		fmt.Printf("工具名称: %s\n", toolInfo.Name)
		fmt.Printf("工具描述: %s\n", toolInfo.Desc)

		// 执行工具
		userResult, err := userTool.InvokableRun(ctx, `{
			"name": "张三",
			"email": "zhangsan@example.com", 
			"age": 28,
			"gender": "male",
			"city": "北京",
			"is_active": true
		}`)
		if err != nil {
			log.Printf("用户工具执行失败: %v", err)
		} else {
			fmt.Printf("执行结果: %s\n\n", userResult)
		}
	}

	// 2. 使用 InferTool 创建订单计算工具
	fmt.Println("--- 创建订单计算工具 ---")
	orderTool, err := utils.InferTool("calculate_order", "计算订单总价", calculateOrder)
	if err != nil {
		log.Printf("创建订单工具失败: %v", err)
	} else {
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
	}

	// 3. 使用 InferTool 创建数据分析工具
	fmt.Println("--- 创建数据分析工具 ---")
	analysisTool, err := utils.InferTool("analyze_data", "分析数值数据集", analyzeData)
	if err != nil {
		log.Printf("创建分析工具失败: %v", err)
	} else {
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
	}

	// 4. 使用 InferTool 创建文本处理工具
	fmt.Println("--- 创建文本处理工具 ---")
	textTool, err := utils.InferTool("process_text", "处理和分析文本", processText)
	if err != nil {
		log.Printf("创建文本工具失败: %v", err)
	} else {
		textResult, err := textTool.InvokableRun(ctx, `{
			"text": "这是一个好的例子。联系我: user@example.com 或访问 https://example.com",
			"operations": ["word_count", "char_count", "extract_emails", "extract_urls", "sentiment"],
			"language": "zh",
			"case_sensitive": false
		}`)
		if err != nil {
			log.Printf("文本工具执行失败: %v", err)
		} else {
			fmt.Printf("处理结果: %s\n\n", textResult)
		}
	}
}

// func main() {
// 	demonstrateInferTool()
// }
