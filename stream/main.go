package main

import (
	"context"
	"fmt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"io"
	"log"
	"strings"
	"time"
)

// 故事生成器 - 流式输出
func createStoryStreamer() *compose.Lambda {
	return compose.StreamableLambda(func(ctx context.Context, prompt string) (*schema.StreamReader[string], error) {
		// 创建一个流管道
		reader, writer := schema.Pipe[string](10)

		go func() {
			defer writer.Close()

			// 根据提示生成故事片段
			storyParts := generateStoryParts(prompt)

			for _, part := range storyParts {
				select {
				case <-ctx.Done():
					return
				default:
					// 发送故事片段
					if err := writer.Send(part, nil); err != nil {
						return
					}
					// 模拟生成延迟，让流式效果更明显
					time.Sleep(200 * time.Millisecond)
				}
			}
		}()

		return reader, nil
	})
}

// 生成故事片段的辅助函数
func generateStoryParts(prompt string) []string {
	baseStory := []string{
		"在一个遥远的王国里，",
		"住着一位勇敢的骑士。",
		"他有一匹神奇的马，",
		"能够飞越高山和河流。",
		"一天，王国遇到了危机，",
		"邪恶的龙威胁着村庄。",
		"骑士决定挺身而出，",
		"踏上了冒险的旅程...",
	}

	// 根据提示词个性化故事
	if strings.Contains(strings.ToLower(prompt), "科幻") {
		return []string{
			"在2024年的地球上，",
			"科技已经高度发达。",
			"人工智能与人类和谐共处，",
			"一起探索宇宙的奥秘。",
			"突然，从深空传来神秘信号，",
			"预示着外星文明的到来...",
		}
	}

	return baseStory
}

// 实时聊天应用示例
func buildStreamingChatApp() {
	var ctx context.Context
	chain := compose.NewChain[string, string]()

	streamer := createStoryStreamer()
	chain.AppendLambda(streamer)

	// 编译并运行
	runnable, err := chain.Compile(ctx)
	if err != nil {
		log.Printf("编译链失败: %v", err)
		return
	}
	// 使用invoke方法执行链
	result, err := runnable.Invoke(ctx, testUser)
	if err != nil {
		log.Printf("运行链失败: %v", err)
		return
	}

	fmt.Printf("  最终结果:\n%s\n", result)

}

func main() {
	buildStreamingChatApp()
}
