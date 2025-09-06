package Eino_Dev_Generate

import (
	"context"

	"github.com/cloudwego/eino/compose"
)

func BuildDemoGraph(ctx context.Context) (r compose.Runnable[any, any], err error) {
	const (
		ChatModel1    = "ChatModel1"
		ChatTemplate1 = "ChatTemplate1"
		ChatTemplate2 = "ChatTemplate2"
		Retriever1    = "Retriever1"
	)
	g := compose.NewGraph[any, any]()
	chatModel1KeyOfChatModel, err := newChatModel(ctx)
	if err != nil {
		return nil, err
	}
	_ = g.AddChatModelNode(ChatModel1, chatModel1KeyOfChatModel)
	chatTemplate1KeyOfChatTemplate, err := newChatTemplate(ctx)
	if err != nil {
		return nil, err
	}
	_ = g.AddChatTemplateNode(ChatTemplate1, chatTemplate1KeyOfChatTemplate)
	chatTemplate2KeyOfChatTemplate, err := newChatTemplate1(ctx)
	if err != nil {
		return nil, err
	}
	_ = g.AddChatTemplateNode(ChatTemplate2, chatTemplate2KeyOfChatTemplate)
	retriever1KeyOfRetriever, err := newRetriever(ctx)
	if err != nil {
		return nil, err
	}
	_ = g.AddRetrieverNode(Retriever1, retriever1KeyOfRetriever)
	_ = g.AddEdge(compose.START, Retriever1)
	_ = g.AddEdge(ChatModel1, compose.END)
	_ = g.AddEdge(ChatTemplate1, ChatModel1)
	_ = g.AddEdge(ChatTemplate2, ChatModel1)
	_ = g.AddEdge(Retriever1, ChatTemplate1)
	_ = g.AddEdge(Retriever1, ChatTemplate2)
	r, err = g.Compile(ctx, compose.WithGraphName("DemoGraph"), compose.WithNodeTriggerMode(compose.AnyPredecessor))
	if err != nil {
		return nil, err
	}
	return r, err
}
