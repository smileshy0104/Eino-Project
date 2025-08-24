package pipeline

import (
	"context"

	"github.com/cloudwego/eino/compose"
)

func BuildmyEino(ctx context.Context) (r compose.Runnable[string, string], err error) {
	// --- 1. 创建 Graph ---
	const (
		Lambda1    = "Lambda1"
		ChatModel1 = "ChatModel1"
	)
	// --- 2. 添加节点和边 ---
	// 创建一个新的有向图
	g := compose.NewGraph[string, string]()
	_ = g.AddLambdaNode(Lambda1, compose.InvokableLambda(newLambda))
	chatModel1KeyOfChatModel, err := newChatModel(ctx)
	if err != nil {
		return nil, err
	}
	// 创建一个新的 ChatModel
	_ = g.AddChatModelNode(ChatModel1, chatModel1KeyOfChatModel)
	_ = g.AddEdge(compose.START, Lambda1)
	_ = g.AddEdge(ChatModel1, compose.END)
	_ = g.AddEdge(Lambda1, ChatModel1)
	// --- 3. 编译 Graph ---
	r, err = g.Compile(ctx, compose.WithGraphName("myEino"), compose.WithNodeTriggerMode(compose.AnyPredecessor))
	if err != nil {
		return nil, err
	}
	return r, err
}
