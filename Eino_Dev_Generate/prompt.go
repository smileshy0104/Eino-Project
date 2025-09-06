package Eino_Dev_Generate

import (
	"context"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

type ChatTemplateConfig struct {
	FormatType schema.FormatType
	Templates  []schema.MessagesTemplate
}

// newChatTemplate component initialization function of node 'ChatTemplate1' in graph 'DemoGraph'
func newChatTemplate(ctx context.Context) (ctp prompt.ChatTemplate, err error) {
	// TODO Modify component configuration here.
	config := &ChatTemplateConfig{}
	ctp = prompt.FromMessages(config.FormatType, config.Templates...)
	return ctp, nil
}

type ChatTemplate1Config struct {
	FormatType schema.FormatType
	Templates  []schema.MessagesTemplate
}

// newChatTemplate1 component initialization function of node 'ChatTemplate2' in graph 'DemoGraph'
func newChatTemplate1(ctx context.Context) (ctp prompt.ChatTemplate, err error) {
	// TODO Modify component configuration here.
	config := &ChatTemplate1Config{}
	ctp = prompt.FromMessages(config.FormatType, config.Templates...)
	return ctp, nil
}
