// Package main 演示 Eino 框架中真正的 Lambda 组件用法
// 使用真实的 StreamableLambda、CollectableLambda、TransformableLambda、AnyLambda
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// UserInput 用户输入结构
type UserInput struct {
	Name  string `json:"name"`
	Age   int    `json:"age"`
	City  string `json:"city"`
	Email string `json:"email"`
}

// ProcessedData 处理后的数据结构
type ProcessedData struct {
	ID          string                 `json:"id"`
	UserInfo    string                 `json:"user_info"`
	ProcessTime string                 `json:"process_time"`
	Category    string                 `json:"category"`
	Metadata    map[string]interface{} `json:"metadata"`
}

func main() {
	fmt.Println("=== Eino 真正的 Lambda 组件演示 ===")

	ctx := context.Background()

	// 获取命令行参数决定运行哪个示例
	if len(os.Args) > 1 {
		example := os.Args[1]
		switch example {
		case "basic":
			fmt.Println("运行基础InvokableLambda演示...")
			basicInvokableDemo(ctx)
		case "stream":
			fmt.Println("运行StreamableLambda演示...")
			streamableLambdaDemo(ctx)
		case "collect":
			fmt.Println("运行CollectableLambda演示...")
			collectableLambdaDemo(ctx)
		case "transform":
			fmt.Println("运行TransformableLambda演示...")
			transformableLambdaDemo(ctx)
		case "chain":
			fmt.Println("运行Lambda链式演示...")
			lambdaChainDemo(ctx)
		default:
			fmt.Printf("未知示例: %s\n", example)
			showUsage()
		}
		return
	}

	// 默认运行所有演示
	runAllDemos(ctx)
}

func runAllDemos(ctx context.Context) {
	//fmt.Println("📝 演示1: InvokableLambda - 基础用法")
	//basicInvokableDemo(ctx)

	//fmt.Println("\n🌊 演示2: StreamableLambda - 生成数据流")
	//streamableLambdaDemo(ctx)

	//fmt.Println("\n📊 演示3: CollectableLambda - 收集数据流")
	//collectableLambdaDemo(ctx)

	fmt.Println("\n⚡ 演示4: TransformableLambda - 流转换")
	transformableLambdaDemo(ctx)

	//fmt.Println("\n🔗 演示5: Lambda链式组合")
	//lambdaChainDemo(ctx)

	fmt.Println("\n✅ 所有真正的 Lambda 演示完成！")
}

// 1. 基础 InvokableLambda 演示
func basicInvokableDemo(ctx context.Context) {
	// 创建一个简单的数据处理 Lambda
	processLambda := compose.InvokableLambda(func(ctx context.Context, input UserInput) (*ProcessedData, error) {
		category := "adult"
		if input.Age < 18 {
			category = "minor"
		} else if input.Age >= 60 {
			category = "senior"
		}

		processed := &ProcessedData{
			ID:          fmt.Sprintf("user_%d", time.Now().Unix()),
			UserInfo:    fmt.Sprintf("%s (%d岁) 来自 %s", input.Name, input.Age, input.City),
			ProcessTime: time.Now().Format("2006-01-02 15:04:05"),
			Category:    category,
			Metadata: map[string]interface{}{
				"email": input.Email,
				"year":  time.Now().Year(),
			},
		}

		fmt.Printf("  处理用户数据: %s\n", processed.UserInfo)
		return processed, nil
	})

	// 在链中使用
	chain := compose.NewChain[UserInput, *ProcessedData]()
	chain.AppendLambda(processLambda)

	runnable, err := chain.Compile(ctx)
	if err != nil {
		log.Printf("编译失败: %v", err)
		return
	}

	// 测试
	testUser := UserInput{Name: "张三", Age: 25, City: "北京", Email: "zhangsan@example.com"}
	fmt.Printf("  输入: %+v\n", testUser)

	result, err := runnable.Invoke(ctx, testUser)
	if err != nil {
		log.Printf("执行失败: %v", err)
		return
	}

	fmt.Printf("  结果: %+v\n", result)
}

// 2. StreamableLambda 演示 - 将输入转换为流输出
func streamableLambdaDemo(ctx context.Context) {
	// StreamableLambda: 接收单个输入，生成流式输出
	wordStreamLambda := compose.StreamableLambda(func(ctx context.Context, input string) (*schema.StreamReader[string], error) {
		words := strings.Fields(input)

		fmt.Printf("  将文本 '%s' 转换为单词流\n", input)

		// 使用 StreamReaderFromArray 创建流
		streamReader := schema.StreamReaderFromArray(words)

		return streamReader, nil
	})

	// 使用 Graph 来执行完整流程：StreamableLambda + CollectableLambda
	// 创建一个收集器来处理流输出
	collectResult := compose.CollectableLambda(func(ctx context.Context, input *schema.StreamReader[string]) (string, error) {
		var words []string
		for {
			word, err := input.Recv()
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				return "", err
			}
			words = append(words, word)
			fmt.Printf("    收到单词: %s\n", word)
		}
		return fmt.Sprintf("收集的单词: [%s]", strings.Join(words, ", ")), nil
	})

	graph := compose.NewGraph[string, string]()
	graph.AddLambdaNode("stream_lambda", wordStreamLambda)
	graph.AddLambdaNode("collect_lambda", collectResult)
	graph.AddEdge(compose.START, "stream_lambda")
	graph.AddEdge("stream_lambda", "collect_lambda")
	graph.AddEdge("collect_lambda", compose.END)

	runnable, err := graph.Compile(ctx)
	if err != nil {
		log.Printf("编译失败: %v", err)
		return
	}

	input := "Hello world from Eino Lambda components"
	fmt.Printf("  输入: '%s'\n", input)

	finalResult, err := runnable.Invoke(ctx, input)
	if err != nil {
		log.Printf("执行失败: %v", err)
		return
	}

	fmt.Printf("  最终结果: %s\n", finalResult)
}

// 3. CollectableLambda 演示 - 收集流式输入为单个输出
func collectableLambdaDemo(ctx context.Context) {
	// 演示正确的 CollectableLambda 用法
	// CollectableLambda 通常与 StreamableLambda 组合使用

	// 步骤1: 创建 StreamableLambda 生成流
	streamLambda := compose.StreamableLambda(func(ctx context.Context, input string) (*schema.StreamReader[string], error) {
		words := strings.Fields(input)
		fmt.Printf("  分割文本 '%s' 为 %d 个单词\n", input, len(words))
		return schema.StreamReaderFromArray(words), nil
	})

	// 步骤2: 创建 CollectableLambda 收集流
	collectLambda := compose.CollectableLambda(func(ctx context.Context, input *schema.StreamReader[string]) (string, error) {
		var words []string
		var totalChars int

		fmt.Println("  开始收集流式输入...")

		for {
			word, err := input.Recv()
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				return "", fmt.Errorf("读取流失败: %w", err)
			}

			words = append(words, word)
			totalChars += len(word)
			fmt.Printf("    收集: '%s'\n", word)
		}

		result := fmt.Sprintf("收集了 %d 个单词，总字符数: %d，内容: [%s]",
			len(words), totalChars, strings.Join(words, ", "))

		fmt.Println("  收集完成")
		return result, nil
	})

	// 使用 Graph 组合 StreamableLambda -> CollectableLambda
	graph := compose.NewGraph[string, string]()
	graph.AddLambdaNode("stream", streamLambda)
	graph.AddLambdaNode("collect", collectLambda)
	graph.AddEdge(compose.START, "stream")
	graph.AddEdge("stream", "collect")
	graph.AddEdge("collect", compose.END)

	runnable, err := graph.Compile(ctx)
	if err != nil {
		log.Printf("编译失败: %v", err)
		return
	}

	input := "Eino Lambda Components CollectableLambda Demo"
	fmt.Printf("  输入: '%s'\n", input)

	result, err := runnable.Invoke(ctx, input)
	if err != nil {
		log.Printf("执行失败: %v", err)
		return
	}

	fmt.Printf("  最终结果: %s\n", result)
}

// 4. TransformableLambda 演示 - 流到流的转换
func transformableLambdaDemo(ctx context.Context) {
	// TransformableLambda: 接收流式输入，生成流式输出
	uppercaseTransformLambda := compose.TransformableLambda(func(ctx context.Context, input *schema.StreamReader[string]) (*schema.StreamReader[string], error) {
		fmt.Println("  开始流转换处理...")

		// 创建输出流
		sr, sw := schema.Pipe[string](10)

		// 在 goroutine 中处理转换
		go func() {
			defer sw.Close()

			for {
				word, err := input.Recv()
				if errors.Is(err, io.EOF) {
					break
				}
				if err != nil {
					fmt.Printf("读取输入流错误: %v\n", err)
					return
				}

				// 转换：大写并添加前缀
				transformed := fmt.Sprintf("TRANSFORMED_%s", strings.ToUpper(word))
				fmt.Printf("    转换: '%s' -> '%s'\n", word, transformed)

				// Send 需要两个参数：chunk 和 error
				closed := sw.Send(transformed, nil)
				if closed {
					fmt.Println("输出流已关闭")
					return
				}

				time.Sleep(100 * time.Millisecond) // 模拟处理时间
			}
		}()

		return sr, nil
	})

	// 创建测试流数据
	testWords := []string{"hello", "world", "eino", "lambda"}
	inputStream := schema.StreamReaderFromArray(testWords)

	// 使用 Graph 来执行 TransformableLambda
	graph := compose.NewGraph[*schema.StreamReader[string], *schema.StreamReader[string]]()
	graph.AddLambdaNode("transform_lambda", uppercaseTransformLambda)
	graph.AddEdge(compose.START, "transform_lambda")
	graph.AddEdge("transform_lambda", compose.END)

	runnable, err := graph.Compile(ctx)
	if err != nil {
		log.Printf("编译失败: %v", err)
		return
	}

	fmt.Printf("  输入流: %v\n", testWords)

	outputStream, err := runnable.Invoke(ctx, inputStream)
	if err != nil {
		log.Printf("执行失败: %v", err)
		return
	}

	// 读取转换后的流
	fmt.Println("  转换结果:")
	for {
		transformed, err := outputStream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			log.Printf("读取输出流失败: %v", err)
			break
		}
		fmt.Printf("    - %s\n", transformed)
	}
}

// 5. Lambda 链式组合演示
func lambdaChainDemo(ctx context.Context) {
	fmt.Println("  创建 StreamableLambda -> CollectableLambda 链")

	// Step 1: StreamableLambda - 文本分割为流
	textToStream := compose.StreamableLambda(func(ctx context.Context, input string) (*schema.StreamReader[string], error) {
		words := strings.Fields(input)
		fmt.Printf("    步骤1: 分割文本为 %d 个单词\n", len(words))
		return schema.StreamReaderFromArray(words), nil
	})

	// Step 2: CollectableLambda - 收集并处理
	streamToResult := compose.CollectableLambda(func(ctx context.Context, input *schema.StreamReader[string]) (string, error) {
		var processedWords []string

		for {
			word, err := input.Recv()
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				return "", err
			}

			// 处理每个单词
			processed := fmt.Sprintf("[%s:%d]", strings.ToUpper(word), len(word))
			processedWords = append(processedWords, processed)
		}

		result := fmt.Sprintf("链式处理结果: %s", strings.Join(processedWords, " -> "))
		fmt.Printf("    步骤2: 收集并格式化 %d 个单词\n", len(processedWords))
		return result, nil
	})

	// 创建链 - 完整的流水线：string -> StreamReader[string] -> string
	chain1 := compose.NewChain[string, *schema.StreamReader[string]]()
	chain1.AppendLambda(textToStream)

	runnable1, err := chain1.Compile(ctx)
	if err != nil {
		log.Printf("编译第一链失败: %v", err)
		return
	}

	chain2 := compose.NewChain[*schema.StreamReader[string], string]()
	chain2.AppendLambda(streamToResult)

	runnable2, err := chain2.Compile(ctx)
	if err != nil {
		log.Printf("编译第二链失败: %v", err)
		return
	}

	// 测试完整流程
	input := "Eino Lambda Chain Demonstration"
	fmt.Printf("  输入: '%s'\n", input)

	// 执行第一步
	stream, err := runnable1.Invoke(ctx, input)
	if err != nil {
		log.Printf("第一步执行失败: %v", err)
		return
	}

	// 执行第二步
	finalResult, err := runnable2.Invoke(ctx, stream)
	if err != nil {
		log.Printf("第二步执行失败: %v", err)
		return
	}

	fmt.Printf("  最终结果: %s\n", finalResult)
}

func showUsage() {
	fmt.Println("用法: go run main_real_lambdas.go [example]")
	fmt.Println("示例:")
	fmt.Println("  basic      - 基础InvokableLambda演示")
	fmt.Println("  stream     - StreamableLambda演示")
	fmt.Println("  collect    - CollectableLambda演示")
	fmt.Println("  transform  - TransformableLambda演示")
	fmt.Println("  chain      - Lambda链式组合演示")
	fmt.Println("\n不带参数运行所有演示")
}
