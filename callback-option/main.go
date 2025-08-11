// package callbackoption 展示了如何实现一个具有回调和选项处理能力的自定义聊天模型。
package callbackoption

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// MyChatModel 表示一个自定义的聊天模型客户端。
// 它持有与聊天服务交互所需的配置和状态。
type MyChatModel struct {
	client     *http.Client  // 用于发起请求的 HTTP 客户端。
	apiKey     string        // 用于身份验证的 API 密钥。
	baseURL    string        // 聊天服务的基础 URL。
	model      string        // 用于生成的默认模型名称。
	timeout    time.Duration // 默认的请求超时时间。
	retryCount int           // 失败时的默认重试次数。
}

// MyChatModelConfig 持有创建 MyChatModel 实例的初始配置。
type MyChatModelConfig struct {
	APIKey string // APIKey 是身份验证所必需的。
}

// MyChatModelOptions 定义了生成请求可用的选项。
// 它包括通用选项和特定于实现的选项。
type MyChatModelOptions struct {
	*model.Options               // 通用选项，如模型、温度等。
	Timeout        time.Duration // 特定于请求的超时时间。
	RetryCount     int           // 特定于请求的重试次数。
}

// NewMyChatModel 创建一个 MyChatModel 的新实例。
// 它需要一个包含 API 密钥的配置对象。
func NewMyChatModel(config *MyChatModelConfig, opts ...model.Option) (*MyChatModel, error) {
	if config.APIKey == "" {
		return nil, errors.New("api key is required")
	}

	// 使用默认值进行初始化
	m := &MyChatModel{
		client:     &http.Client{},
		apiKey:     config.APIKey,
		model:      "default-model",
		timeout:    60 * time.Second,
		retryCount: 3,
	}

	// 应用函数式选项来覆盖默认值
	options := &MyChatModelOptions{
		Options: &model.Options{
			Model: &m.model,
		},
		Timeout:    m.timeout,
		RetryCount: m.retryCount,
	}
	// 这是一个重用选项处理逻辑的小技巧。
	// 我们将选项应用于一个临时结构体，然后将值复制回来。
	implOpts := model.GetImplSpecificOptions(options, opts...)
	m.model = *implOpts.Options.Model
	m.timeout = implOpts.Timeout
	m.retryCount = implOpts.RetryCount

	return m, nil
}

// Generate 执行非流式聊天生成。
func (m *MyChatModel) Generate(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	// 1. 处理选项
	// 从模型的默认设置开始，并使用每个请求的选项进行覆盖。
	options := &MyChatModelOptions{
		Options: &model.Options{
			Model: &m.model,
		},
		RetryCount: m.retryCount,
		Timeout:    m.timeout,
	}
	options.Options = model.GetCommonOptions(options.Options, opts...)
	options = model.GetImplSpecificOptions(options, opts...)

	// 2. 触发 OnStart 回调
	// 这会通知监听器生成任务即将开始。
	ctx = callbacks.OnStart(ctx, &model.CallbackInput{
		Messages: messages,
		Config: &model.Config{
			Model: *options.Options.Model,
		},
	})

	// 3. 执行核心生成逻辑
	response, err := m.doGenerate(ctx, messages, options)

	// 4. 处理错误和完成回调
	if err != nil {
		// 如果发生错误，触发 OnError 回调。
		ctx = callbacks.OnError(ctx, err)
		return nil, err
	}

	// 成功完成后触发 OnEnd 回调。
	ctx = callbacks.OnEnd(ctx, &model.CallbackOutput{
		Message: response,
	})

	return response, nil
}

// Stream 执行流式聊天生成。
func (m *MyChatModel) Stream(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	// 1. 处理选项
	// 与 Generate 中相同，从默认值开始，并使用每个请求的选项进行覆盖。
	options := &MyChatModelOptions{
		Options: &model.Options{
			Model: &m.model,
		},
		RetryCount: m.retryCount,
		Timeout:    m.timeout,
	}
	options.Options = model.GetCommonOptions(options.Options, opts...)
	options = model.GetImplSpecificOptions(options, opts...)

	// 2. 触发 OnStart 回调
	ctx = callbacks.OnStart(ctx, &model.CallbackInput{
		Messages: messages,
		Config: &model.Config{
			Model: *options.Options.Model,
		},
	})

	// 3. 创建一个流管道
	// 管道提供了一个写入器和一个读取器。核心逻辑写入写入器，
	// 调用者从读取器读取。这是线程安全的。
	sr, sw := schema.Pipe[*model.CallbackOutput](1)

	// 4. 启动异步生成
	go func() {
		// 确保在 goroutine 完成时关闭写入器。
		defer sw.Close()

		// 核心流式逻辑将数据块写入流写入器 (sw)。
		m.doStream(ctx, messages, options, sw)
	}()

	// 5. 使用流触发 OnEnd 回调
	// OnEndWithStreamOutput 为回调处理流。它在内部
	// 复制流，以便回调系统和调用者
	// 可以独立地消费它。`nsr` 是供调用者使用的新流。
	_, nsr := callbacks.OnEndWithStreamOutput(ctx, sr)

	// 将 CallbackOutput 的流转换为 Message 的流。
	return schema.StreamReaderWithConvert(nsr, func(t *model.CallbackOutput) (*schema.Message, error) {
		return t.Message, nil
	}), nil
}

// WithTools 应该将工具绑定到聊天模型以实现工具调用功能。
// 这是一个占位符实现。
func (m *MyChatModel) WithTools(tools []*schema.ToolInfo) (model.ToolCallingChatModel, error) {
	// TODO: 实现将工具绑定到模型的逻辑。
	// 这可能涉及创建一个新的模型实例或修改现有的实例。
	return nil, errors.New("WithTools not implemented")
}

// doGenerate 包含发出非流式 API 调用的实际逻辑。
// 这是一个占位符实现。
func (m *MyChatModel) doGenerate(ctx context.Context, messages []*schema.Message, opts *MyChatModelOptions) (*schema.Message, error) {
	// TODO: 在此处实现生成逻辑。
	// 1. 准备 HTTP 请求（例如，从消息和选项创建 JSON 主体）。
	// 2. 设置标头（例如，使用 API 密钥进行授权）。
	// 3. 使用 m.client 发出 HTTP 请求。
	// 4. 处理 HTTP 响应，检查错误。
	// 5. 将响应主体解析为 *schema.Message。
	// 6. 返回消息。
	return &schema.Message{Content: "This is a generated message."}, nil
}

// doStream 包含发出流式 API 调用的实际逻辑。
// 它应该将生成的块发送到提供的流写入器。
// 这是一个占位符实现。
func (m *MyChatModel) doStream(ctx context.Context, messages []*schema.Message, opts *MyChatModelOptions, sw *schema.StreamWriter[*model.CallbackOutput]) {
	// TODO: 在此处实现流式逻辑。
	// 1. 准备并发出流式 HTTP 请求。
	// 2. 在循环中逐块读取响应主体。
	// 3. 对于每个块：
	//    a. 将其解析为 *schema.Message。
	//    b. 创建一个 *model.CallbackOutput。
	//    c. 将输出写入流写入器：sw.Send(&output, nil)。
	//    d. 处理写入时可能出现的错误。
	// 4. 如果在流传输过程中发生错误，请使用 sw.Send(nil, err) 写入错误。
	//
	// 写入块的示例：
	for i := 0; i < 5; i++ {
		time.Sleep(100 * time.Millisecond) // 模拟网络延迟
		chunk := &model.CallbackOutput{
			Message: &schema.Message{Content: " chunk"},
		}
		// Send 接受一个值和一个错误。如果发送了非 nil 的错误，
		// 流将被关闭。如果流已经关闭，它将返回 false。
		if !sw.Send(chunk, nil) {
			// 流已被读取器关闭，所以我们应该停止。
			return
		}
	}
}
