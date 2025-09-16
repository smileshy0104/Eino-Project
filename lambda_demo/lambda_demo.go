// Package main 演示 Eino 框架中 Lambda 组件的各种用法
// Lambda 是 Eino 中的核心组件，用于在工作流中嵌入自定义函数逻辑
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/cloudwego/eino/compose"
)

//// UserInput 用户输入结构
//type UserInput struct {
//	Name string `json:"name"`
//	Age  int    `json:"age"`
//	City string `json:"city"`
//}
//
//// ProcessedData 处理后的数据结构
//type ProcessedData struct {
//	UserInfo    string `json:"user_info"`
//	ProcessTime string `json:"process_time"`
//	Category    string `json:"category"`
//}

func main() {
	fmt.Println("=== Eino Lambda 组件演示 ===")

	ctx := context.Background()

	// 演示1: InvokableLambda 在 Chain 中的使用
	fmt.Println("📝 演示1: InvokableLambda - 数据处理链")
	runInvokableLambdaDemo(ctx)

	// 演示2: 复杂 JSON 处理链
	fmt.Println("\n🔗 演示2: 复杂数据处理链")
	runChainWithLambdaDemo(ctx)

	// 演示3: 文本处理链
	fmt.Println("\n📄 演示3: 文本处理链")
	runTextProcessingDemo(ctx)

	// 演示4: 数据验证和转换链
	fmt.Println("\n✅ 演示4: 数据验证链")
	runValidationChainDemo(ctx)

	fmt.Println("\n✅ 所有 Lambda 演示完成！")
}

// 演示1: InvokableLambda 在 Chain 中的使用
func runInvokableLambdaDemo1(ctx context.Context) {
	// 创建一个用户数据处理链
	chain := compose.NewChain[UserInput, string]()

	// Step 1: 处理用户数据
	processUserData := compose.InvokableLambda(func(ctx context.Context, input UserInput) (*ProcessedData, error) {
		category := "adult"
		if input.Age < 18 {
			category = "minor"
		} else if input.Age >= 60 {
			category = "senior"
		}

		// 处理用户信息
		processed := &ProcessedData{
			UserInfo:    fmt.Sprintf("%s (%d岁) 来自 %s", input.Name, input.Age, input.City),
			ProcessTime: time.Now().Format("2006-01-02 15:04:05"),
			Category:    category,
		}

		fmt.Printf("  步骤1 - 处理用户数据: %s\n", processed.UserInfo)
		return processed, nil
	})

	// Step 2: 格式化输出
	formatOutput := compose.InvokableLambda(func(ctx context.Context, data *ProcessedData) (string, error) {
		output := fmt.Sprintf("=== 用户信息报告 ===\n姓名: %s\n分类: %s\n处理时间: %s",
			data.UserInfo, data.Category, data.ProcessTime)
		fmt.Printf("  步骤2 - 格式化完成\n")
		return output, nil
	})

	// 构建链
	chain.AppendLambda(processUserData)
	chain.AppendLambda(formatOutput)

	// 编译并运行
	runnable, err := chain.Compile(ctx)
	if err != nil {
		log.Printf("编译链失败: %v", err)
		return
	}

	// 测试数据
	testUser := UserInput{Name: "张三", Age: 25, City: "北京"}
	fmt.Printf("  输入: %+v\n", testUser)

	// 使用invoke方法执行链
	result, err := runnable.Invoke(ctx, testUser)
	if err != nil {
		log.Printf("运行链失败: %v", err)
		return
	}

	fmt.Printf("  最终结果:\n%s\n", result)
}

// 演示3: 文本处理链
func runTextProcessingDemo(ctx context.Context) {
	chain := compose.NewChain[string, string]()

	// Step 1: 文本清理
	cleanText := compose.InvokableLambda(func(ctx context.Context, input string) (string, error) {
		cleaned := strings.TrimSpace(input)
		cleaned = strings.ReplaceAll(cleaned, "  ", " ") // 去除多余空格
		fmt.Printf("  步骤1 - 文本清理: '%s'\n", cleaned)
		return cleaned, nil
	})

	// Step 2: 大小写转换
	transformCase := compose.InvokableLambda(func(ctx context.Context, input string) (string, error) {
		result := strings.ToUpper(input)
		fmt.Printf("  步骤2 - 大写转换: '%s'\n", result)
		return result, nil
	})

	// Step 3: 添加格式
	formatText := compose.InvokableLambda(func(ctx context.Context, input string) (string, error) {
		result := fmt.Sprintf("*** %s ***", input)
		fmt.Printf("  步骤3 - 格式化完成\n")
		return result, nil
	})

	chain.AppendLambda(cleanText)
	chain.AppendLambda(transformCase)
	chain.AppendLambda(formatText)

	runnable, err := chain.Compile(ctx)
	if err != nil {
		log.Printf("编译链失败: %v", err)
		return
	}

	input := "  hello   world  "
	fmt.Printf("  输入: '%s'\n", input)

	result, err := runnable.Invoke(ctx, input)
	if err != nil {
		log.Printf("运行链失败: %v", err)
		return
	}

	fmt.Printf("  最终结果: %s\n", result)
}

// 演示4: 数据验证链
func runValidationChainDemo(ctx context.Context) {
	chain := compose.NewChain[map[string]interface{}, string]()

	// Step 1: 数据验证
	validateData := compose.InvokableLambda(func(ctx context.Context, data map[string]interface{}) (map[string]interface{}, error) {
		// 检查必需字段
		requiredFields := []string{"name", "age"}
		for _, field := range requiredFields {
			if _, exists := data[field]; !exists {
				return nil, fmt.Errorf("缺少必需字段: %s", field)
			}
		}

		// 验证年龄
		if age, ok := data["age"].(float64); ok {
			if age < 0 || age > 150 {
				return nil, fmt.Errorf("年龄无效: %.0f", age)
			}
		}

		fmt.Printf("  步骤1 - 数据验证通过\n")
		return data, nil
	})

	// Step 2: 数据标准化
	normalizeData := compose.InvokableLambda(func(ctx context.Context, data map[string]interface{}) (map[string]string, error) {
		result := make(map[string]string)
		result["name"] = strings.TrimSpace(data["name"].(string))
		result["age"] = fmt.Sprintf("%.0f", data["age"].(float64))
		result["processed_at"] = time.Now().Format("2006-01-02 15:04:05")

		fmt.Printf("  步骤2 - 数据标准化完成\n")
		return result, nil
	})

	// Step 3: 生成报告
	generateReport := compose.InvokableLambda(func(ctx context.Context, data map[string]string) (string, error) {
		report := fmt.Sprintf("用户报告\n姓名: %s\n年龄: %s\n处理时间: %s",
			data["name"], data["age"], data["processed_at"])
		fmt.Printf("  步骤3 - 报告生成完成\n")
		return report, nil
	})

	chain.AppendLambda(validateData)
	chain.AppendLambda(normalizeData)
	chain.AppendLambda(generateReport)

	runnable, err := chain.Compile(ctx)
	if err != nil {
		log.Printf("编译链失败: %v", err)
		return
	}

	// 测试有效数据
	validData := map[string]interface{}{
		"name": "  王五  ",
		"age":  float64(30),
		"city": "上海",
	}

	fmt.Printf("  输入: %+v\n", validData)

	result, err := runnable.Invoke(ctx, validData)
	if err != nil {
		log.Printf("处理失败: %v", err)

		// 测试无效数据
		fmt.Printf("\n  测试无效数据:\n")
		invalidData := map[string]interface{}{
			"name": "测试",
			// 缺少 age 字段
		}

		fmt.Printf("  输入: %+v\n", invalidData)
		_, err2 := runnable.Invoke(ctx, invalidData)
		if err2 != nil {
			fmt.Printf("  验证错误: %v\n", err2)
		}
		return
	}

	fmt.Printf("  最终结果:\n%s\n", result)
}

// 演示2: 复杂数据处理链
func runChainWithLambdaDemo(ctx context.Context) {
	// 创建一个处理链，演示 Lambda 在工作流中的应用

	// Step 1: 解析 JSON 输入
	parseJSON := compose.InvokableLambda(func(ctx context.Context, jsonStr string) (map[string]interface{}, error) {
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
			return nil, fmt.Errorf("JSON解析失败: %w", err)
		}
		fmt.Printf("  步骤1 - JSON解析: %v\n", data)
		return data, nil
	})

	// Step 2: 数据验证和转换
	validateAndTransform := compose.InvokableLambda(func(ctx context.Context, data map[string]interface{}) (map[string]string, error) {
		result := make(map[string]string)

		// 验证必需字段
		requiredFields := []string{"name", "email"}
		for _, field := range requiredFields {
			if value, exists := data[field]; exists {
				result[field] = fmt.Sprintf("%v", value)
			} else {
				return nil, fmt.Errorf("缺少必需字段: %s", field)
			}
		}

		// 添加处理时间戳
		result["processed_at"] = time.Now().Format(time.RFC3339)

		fmt.Printf("  步骤2 - 验证转换: %v\n", result)
		return result, nil
	})

	// Step 3: 格式化输出
	formatOutput := compose.InvokableLambda(func(ctx context.Context, data map[string]string) (string, error) {
		output := fmt.Sprintf("用户信息处理完成:\n  姓名: %s\n  邮箱: %s\n  处理时间: %s",
			data["name"], data["email"], data["processed_at"])
		fmt.Printf("  步骤3 - 格式化输出完成\n")
		return output, nil
	})

	// Step 4: 创建链
	chain := compose.NewChain[string, string]()
	// 将各个步骤添加到链中
	chain.AppendLambda(parseJSON)
	chain.AppendLambda(validateAndTransform)
	chain.AppendLambda(formatOutput)

	// 编译并运行链
	runnable, err := chain.Compile(ctx)
	if err != nil {
		log.Printf("编译链失败: %v", err)
		return
	}

	// 测试数据
	testJSON := `{"name": "李四", "email": "lisi@example.com", "age": 30}`
	fmt.Printf("  输入JSON: %s\n", testJSON)

	finalResult, err := runnable.Invoke(ctx, testJSON)
	if err != nil {
		log.Printf("运行链失败: %v", err)
		return
	}

	fmt.Printf("  最终结果:\n%s\n", finalResult)
}
