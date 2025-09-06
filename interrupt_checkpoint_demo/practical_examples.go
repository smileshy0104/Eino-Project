// Package main 实际应用场景中断演示
// 本包演示了在真实业务场景中如何使用中断机制来处理复杂的工作流
// 包含数据处理、审批流程、API调用等常见的企业级应用场景
// 展示了中断机制在实际生产环境中的应用价值和最佳实践
package main

import (
	"context"   // 上下文包，用于控制goroutine的生命周期和传递请求范围的值
	"fmt"       // 格式化I/O包，用于格式化输出和字符串处理
	"log"       // 日志包，用于记录程序运行时的日志信息
	"math/rand" // 随机数包，用于模拟随机事件和概率性场景
	"strings"   // 字符串处理包，用于字符串操作和匹配
	"time"      // 时间包，用于时间处理和延迟操作

	"github.com/spf13/viper" // 配置管理包，用于读取和管理应用配置
)

// DataProcessingJob 数据处理工作流相关结构
// 用于表示大数据处理任务的状态和进度信息
// 适用于ETL流程、数据清洗、数据分析等场景
type DataProcessingJob struct {
	JobID         string                 `json:"job_id"`         // 作业唯一标识符
	DataSource    string                 `json:"data_source"`    // 数据源名称或路径
	ProcessedRows int                    `json:"processed_rows"` // 已处理的数据行数
	TotalRows     int                    `json:"total_rows"`     // 总数据行数
	CurrentStage  string                 `json:"current_stage"`  // 当前处理阶段（获取、清洗、分析等）
	Metadata      map[string]interface{} `json:"metadata"`       // 作业元数据，存储额外信息
	Errors        []string               `json:"errors"`         // 处理过程中遇到的错误列表
}

// ApprovalRequest 审批工作流相关结构
// 用于表示需要审批的请求及其状态信息
// 适用于费用报销、采购申请、投资决策等审批场景
type ApprovalRequest struct {
	RequestID     string                 `json:"request_id"`     // 申请唯一标识符
	RequestType   string                 `json:"request_type"`   // 申请类型（设备采购、差旅报销等）
	Amount        float64                `json:"amount"`         // 申请金额
	Requester     string                 `json:"requester"`      // 申请人姓名
	Status        string                 `json:"status"`         // 审批状态（待审批、已通过、已拒绝等）
	ApprovalLevel int                    `json:"approval_level"` // 所需审批级别（1-普通主管，2-高级管理层）
	Metadata      map[string]interface{} `json:"metadata"`       // 申请元数据，存储额外信息
	Comments      []string               `json:"comments"`       // 审批过程中的评论和备注
}

// APICallJob API调用工作流相关结构
// 用于表示批量API调用任务的执行状态和统计信息
// 适用于数据同步、批量通知、第三方服务集成等场景
type APICallJob struct {
	JobID        string                 `json:"job_id"`        // 作业唯一标识符
	APIEndpoint  string                 `json:"api_endpoint"`  // API端点URL
	RequestCount int                    `json:"request_count"` // 总请求数量
	SuccessCount int                    `json:"success_count"` // 成功请求数量
	FailureCount int                    `json:"failure_count"` // 失败请求数量
	CurrentBatch int                    `json:"current_batch"` // 当前处理的批次号
	TotalBatches int                    `json:"total_batches"` // 总批次数量
	Metadata     map[string]interface{} `json:"metadata"`      // 作业元数据，存储额外信息
	LastError    string                 `json:"last_error"`    // 最后一次错误信息
}

// InterruptError 自定义中断错误类型
// 用于表示工作流执行过程中的中断情况
// 支持区分是否需要人工修复，以便采取不同的处理策略
type InterruptError struct {
	Message     string                 // 中断错误消息描述
	RequiresFix bool                   // 是否需要人工修复（true-需要修复，false-仅需等待）
	Metadata    map[string]interface{} // 中断相关的元数据信息
}

// Error 实现error接口
// 返回中断错误的消息描述，用于错误处理和日志记录
func (e *InterruptError) Error() string {
	return e.Message
}

// NewInterruptError 创建新的中断错误实例
// 用于构造包含消息和修复标志的中断错误对象
// 参数 message: 错误消息描述
// 参数 requiresFix: 是否需要人工修复
// 返回新创建的InterruptError指针
func NewInterruptError(message string, requiresFix bool) *InterruptError {
	return &InterruptError{
		Message:     message,                      // 设置错误消息
		RequiresFix: requiresFix,                  // 设置是否需要修复标志
		Metadata:    make(map[string]interface{}), // 初始化元数据映射
	}
}

// dataIngestionStep 数据获取步骤 - 数据处理工作流第一阶段
// 模拟从数据源获取大量数据的过程，展示长时间运行任务的进度跟踪
// 这是数据处理管道的入口步骤，负责从各种数据源（数据库、文件、API等）获取原始数据
// 参数 ctx: 上下文对象，用于控制执行流程和传递请求信息
// 参数 job: 数据处理作业对象，包含作业状态和配置信息
// 返回更新后的作业对象和可能的错误
func dataIngestionStep(ctx context.Context, job DataProcessingJob) (DataProcessingJob, error) {
	fmt.Printf("📥 开始数据获取，作业ID: %s\n", job.JobID)

	job.CurrentStage = "数据获取" // 设置当前处理阶段
	job.TotalRows = 1000000   // 模拟大数据集，100万行数据
	job.ProcessedRows = 0     // 初始化已处理行数

	// 模拟数据获取过程，分批次获取数据
	for i := 0; i < 5; i++ {
		time.Sleep(200 * time.Millisecond) // 模拟数据获取耗时
		job.ProcessedRows += 200000        // 每批次获取20万行
		// 显示获取进度
		fmt.Printf("📊 已获取数据: %d/%d 行 (%.1f%%)\n",
			job.ProcessedRows, job.TotalRows,
			float64(job.ProcessedRows)/float64(job.TotalRows)*100)
	}

	// 初始化元数据映射（如果尚未初始化）
	if job.Metadata == nil {
		job.Metadata = make(map[string]interface{})
	}
	// 记录数据获取完成时间
	job.Metadata["ingestion_completed"] = time.Now().Format(time.RFC3339)
	fmt.Printf("✅ 数据获取完成\n")

	return job, nil // 返回更新后的作业对象，无错误
}

// dataCleaningStep 数据清洗步骤 - 数据处理工作流第二阶段
// 对获取的原始数据进行质量检查和清洗处理，可能因数据质量问题触发中断
// 这是数据处理管道的核心步骤，负责数据去重、格式标准化、异常值处理等
// 当数据质量分数低于阈值时会触发中断，需要人工介入处理
// 参数 ctx: 上下文对象，用于控制执行流程和传递请求信息
// 参数 job: 数据处理作业对象，包含作业状态和配置信息
// 返回更新后的作业对象和可能的中断错误
func dataCleaningStep(ctx context.Context, job DataProcessingJob) (DataProcessingJob, error) {
	fmt.Printf("🧹 开始数据清洗，作业ID: %s\n", job.JobID)

	job.CurrentStage = "数据清洗" // 设置当前处理阶段

	// 模拟数据质量评估过程
	dataQualityScore := rand.Float64() // 生成0-1之间的随机质量分数
	if job.Metadata == nil {
		job.Metadata = make(map[string]interface{})
	}
	job.Metadata["data_quality_score"] = dataQualityScore // 记录数据质量分数

	// 检查数据质量是否达标（阈值为0.7）
	if dataQualityScore < 0.7 {
		errorMsg := fmt.Sprintf("数据质量分数过低: %.2f, 需要人工介入", dataQualityScore)
		job.Errors = append(job.Errors, errorMsg) // 记录错误信息
		fmt.Printf("⚠️  %s\n", errorMsg)

		// 返回中断错误，标记需要人工修复
		return job, NewInterruptError(errorMsg, true)
	}

	// 数据质量合格，执行清洗过程
	job.ProcessedRows = 0 // 重置处理行数计数器
	for i := 0; i < 10; i++ {
		time.Sleep(100 * time.Millisecond)      // 模拟清洗处理耗时
		job.ProcessedRows += job.TotalRows / 10 // 每次处理10%的数据
		fmt.Printf("📊 数据清洗进度: %d/%d 行\n", job.ProcessedRows, job.TotalRows)
	}

	// 记录清洗完成时间
	job.Metadata["cleaning_completed"] = time.Now().Format(time.RFC3339)
	fmt.Printf("✅ 数据清洗完成\n")

	return job, nil // 返回清洗完成的作业对象，无错误
}

// dataAnalysisStep 数据分析步骤 - 数据处理工作流第三阶段
// 对清洗后的数据执行深度分析，包括统计分析、关联分析、趋势分析、异常检测和模型训练
// 这是数据处理管道的核心分析环节，可能在模型训练阶段因资源不足触发中断
// 当GPU资源不足时会触发中断，需要等待资源释放后重试
// 参数 ctx: 上下文对象，用于控制执行流程和传递请求信息
// 参数 job: 数据处理作业对象，包含清洗后的数据和状态信息
// 返回包含分析结果的作业对象和可能的中断错误
func dataAnalysisStep(ctx context.Context, job DataProcessingJob) (DataProcessingJob, error) {
	fmt.Printf("📈 开始数据分析，作业ID: %s\n", job.JobID)

	job.CurrentStage = "数据分析" // 设置当前处理阶段

	// 定义分析流程的各个步骤，涵盖完整的数据分析生命周期
	analysisSteps := []string{
		"统计分析", // 基础统计指标计算
		"关联分析", // 数据间关联关系分析
		"趋势分析", // 时间序列趋势识别
		"异常检测", // 异常值和离群点检测
		"模型训练", // 机器学习模型训练
	}

	// 逐步执行分析任务
	for i, step := range analysisSteps {
		fmt.Printf("🔄 执行 %s...\n", step)
		time.Sleep(500 * time.Millisecond) // 模拟分析处理耗时

		// 在模型训练阶段模拟GPU资源不足问题（40%概率）
		if step == "模型训练" && rand.Float64() < 0.4 {
			errorMsg := "GPU资源不足，需要等待资源释放"
			job.Errors = append(job.Errors, errorMsg) // 记录资源不足错误
			fmt.Printf("⚠️  %s\n", errorMsg)

			// 返回中断错误，标记需要修复（等待资源）
			return job, NewInterruptError(errorMsg, true)
		}

		// 初始化元数据存储（如果尚未初始化）
		if job.Metadata == nil {
			job.Metadata = make(map[string]interface{})
		}
		// 记录每个步骤的完成状态
		job.Metadata[fmt.Sprintf("step_%d_completed", i)] = step
	}

	// 记录分析完成时间
	job.Metadata["analysis_completed"] = time.Now().Format(time.RFC3339)
	fmt.Printf("✅ 数据分析完成\n")

	return job, nil // 返回包含分析结果的作业对象，无错误
}

// autoApprovalStep 自动审批步骤 - 审批工作流第一阶段
// 基于预设规则对审批请求进行自动处理，适用于低风险、小金额的常规审批
// 当请求超出自动审批范围时会触发中断，转入人工审批流程
// 自动审批规则：≤1000自动批准，≤10000需主管审批，>10000需高级管理层审批
// 参数 ctx: 上下文对象，用于控制执行流程和传递请求信息
// 参数 request: 审批请求对象，包含请求详情和状态信息
// 返回更新后的审批请求对象和可能的中断错误
func autoApprovalStep(ctx context.Context, request ApprovalRequest) (ApprovalRequest, error) {
	fmt.Printf("🤖 自动审批检查，申请ID: %s\n", request.RequestID)

	request.Status = "自动审批中" // 设置当前审批状态

	// 执行分层自动审批规则检查
	if request.Amount <= 1000 {
		// 小额申请直接自动批准
		request.Status = "自动批准"
		request.Comments = append(request.Comments, "小额申请，自动批准")
		fmt.Printf("✅ 小额申请自动批准: %.2f\n", request.Amount)
		return request, nil // 无需中断，直接完成
	}

	if request.Amount <= 10000 {
		// 中额申请需要主管级别审批
		request.Status = "需要主管审批"
		request.ApprovalLevel = 1 // 设置为一级审批
		request.Comments = append(request.Comments, "中额申请，需要主管审批")
		fmt.Printf("⏸️  中额申请需要主管审批: %.2f\n", request.Amount)

		// 返回中断错误，等待主管审批（无需修复）
		return request, NewInterruptError(fmt.Sprintf("需要主管审批: %.2f", request.Amount), false)
	}

	// 大额申请需要高级管理层多级审批
	request.Status = "需要高级管理层审批"
	request.ApprovalLevel = 2 // 设置为二级审批
	request.Comments = append(request.Comments, "大额申请，需要高级管理层审批")
	fmt.Printf("⏸️  大额申请需要高级管理层审批: %.2f\n", request.Amount)

	// 返回中断错误，等待高级管理层审批（无需修复）
	return request, NewInterruptError(fmt.Sprintf("需要高级管理层审批: %.2f", request.Amount), false)
}

// manualApprovalStep 人工审批步骤 - 审批工作流第二阶段
// 由人工审批员对自动审批无法处理的请求进行人工决策
// 基于风险评估和业务规则进行审批决策，可能需要补充材料或直接拒绝
// 审批决策概率：20%拒绝，20%需补充材料，60%通过
// 参数 ctx: 上下文对象，用于控制执行流程和传递请求信息
// 参数 request: 审批请求对象，包含请求详情和当前审批状态
// 返回包含审批决策的请求对象和可能的中断错误
func manualApprovalStep(ctx context.Context, request ApprovalRequest) (ApprovalRequest, error) {
	fmt.Printf("👤 人工审批处理，申请ID: %s\n", request.RequestID)

	// 模拟人工审批员的决策过程（基于随机概率模拟不同审批结果）
	approvalDecision := rand.Float64()

	// 20%概率：审批拒绝（高风险或不符合政策）
	if approvalDecision > 0.8 {
		request.Status = "审批拒绝"
		request.Comments = append(request.Comments, "风险过高，审批拒绝")
		fmt.Printf("❌ 审批被拒绝\n")
		return request, nil // 审批流程结束，无需中断
	}

	// 20%概率：需要补充材料（信息不足或需要更多证明）
	if approvalDecision > 0.6 {
		request.Status = "需要补充材料"
		request.Comments = append(request.Comments, "需要提供更多支持材料")
		fmt.Printf("⏸️  需要补充材料\n")

		// 返回中断错误，需要申请人补充材料后重新提交
		return request, NewInterruptError("需要补充材料", true)
	}

	// 60%概率：审批通过
	request.Status = "审批通过"
	request.Comments = append(request.Comments, fmt.Sprintf("经%d级审批通过", request.ApprovalLevel))
	fmt.Printf("✅ 审批通过\n")

	return request, nil // 审批通过，流程正常结束
}

// apiBatchCallStep API批量调用步骤 - API调用工作流核心阶段
// 执行批量API调用任务，处理网络请求、响应解析和错误重试
// 在API调用过程中可能因网络问题、服务不可用等原因触发中断
// 当API错误率超过20%时会触发中断，需要人工介入处理
// 参数 ctx: 上下文对象，用于控制执行流程和传递请求信息
// 参数 job: API调用作业对象，包含批量调用配置和状态信息
// 返回包含调用结果的作业对象和可能的中断错误
func apiBatchCallStep(ctx context.Context, job APICallJob) (APICallJob, error) {
	fmt.Printf("🌐 执行API批量调用，作业ID: %s\n", job.JobID)

	// 模拟API批量调用配置
	batchSize := 10      // 每批次处理的请求数量
	totalRequests := 100 // 总请求数量

	// 初始化作业统计信息
	job.TotalBatches = totalRequests / batchSize // 计算总批次数
	job.RequestCount = 0                         // 重置请求计数器
	job.SuccessCount = 0                         // 重置成功计数器
	job.FailureCount = 0                         // 重置失败计数器

	// 分批次执行API调用
	for batch := 1; batch <= job.TotalBatches; batch++ {
		job.CurrentBatch = batch
		fmt.Printf("📦 处理批次 %d/%d\n", batch, job.TotalBatches)

		// 模拟单批次内的API调用处理
		for i := 0; i < batchSize; i++ {
			job.RequestCount++ // 增加总请求计数

			// 模拟API调用成功率（90%成功率）
			if rand.Float64() > 0.1 {
				job.SuccessCount++ // API调用成功
			} else {
				job.FailureCount++ // API调用失败
				// 记录最后一次错误信息
				job.LastError = fmt.Sprintf("批次 %d 中的请求 %d 失败", batch, i+1)
			}
		}

		// 检查API错误率是否超过阈值（20%）
		if job.RequestCount > 20 && float64(job.FailureCount)/float64(job.RequestCount) > 0.2 {
			errorMsg := fmt.Sprintf("API错误率过高: %.1f%%",
				float64(job.FailureCount)/float64(job.RequestCount)*100)
			fmt.Printf("⚠️  %s\n", errorMsg)

			// 返回中断错误，需要人工介入处理API问题
			return job, NewInterruptError(errorMsg, true)
		}

		time.Sleep(200 * time.Millisecond) // 模拟批次间延迟，避免API限流
	}

	// 初始化元数据存储（如果尚未初始化）
	if job.Metadata == nil {
		job.Metadata = make(map[string]interface{})
	}
	// 记录API调用完成时间
	job.Metadata["completion_time"] = time.Now().Format(time.RFC3339)
	fmt.Printf("✅ API批量调用完成，成功: %d, 失败: %d\n", job.SuccessCount, job.FailureCount)

	return job, nil // 返回包含调用结果的作业对象，无错误
}

// 工作流执行器，支持中断和重试
type PracticalWorkflowExecutor struct {
	maxRetries int
}

func NewPracticalWorkflowExecutor(maxRetries int) *PracticalWorkflowExecutor {
	return &PracticalWorkflowExecutor{maxRetries: maxRetries}
}

func (we *PracticalWorkflowExecutor) executeWithRetry(ctx context.Context, stepName string, stepFunc func() error) error {
	for retry := 0; retry < we.maxRetries; retry++ {
		err := stepFunc()
		if err == nil {
			return nil
		}

		if interruptErr, ok := err.(*InterruptError); ok {
			fmt.Printf("⏸️  步骤 '%s' 中断 (第%d次重试): %s\n", stepName, retry+1, interruptErr.Message)

			if interruptErr.RequiresFix {
				fmt.Printf("🔧 正在处理问题: %s\n", interruptErr.Message)

				// 模拟问题修复
				if strings.Contains(interruptErr.Message, "数据质量") {
					fmt.Println("👤 等待人工审核...")
					time.Sleep(1 * time.Second)
					fmt.Println("✅ 数据质量问题已修复")
				} else if strings.Contains(interruptErr.Message, "资源不足") {
					fmt.Println("⏳ 等待GPU资源...")
					time.Sleep(2 * time.Second)
					fmt.Println("✅ GPU资源已可用")
				} else if strings.Contains(interruptErr.Message, "错误率过高") {
					fmt.Println("🔍 调查API问题...")
					time.Sleep(2 * time.Second)
					fmt.Println("✅ API问题已修复")
				}

				continue // 重试
			} else {
				// 不需要修复的中断（如人工审批）
				fmt.Printf("👤 等待外部处理: %s\n", interruptErr.Message)
				time.Sleep(1 * time.Second)
				continue
			}
		} else {
			fmt.Printf("❌ 步骤 '%s' 失败 (第%d次重试): %v\n", stepName, retry+1, err)
			if retry == we.maxRetries-1 {
				return err
			}
		}
	}

	return fmt.Errorf("步骤 '%s' 在 %d 次重试后仍然失败", stepName, we.maxRetries)
}

// 数据处理工作流演示
func dataProcessingWorkflowDemo(ctx context.Context) {
	fmt.Println("=== 数据处理工作流演示 ===")

	executor := NewPracticalWorkflowExecutor(3)

	// 初始化作业
	job := DataProcessingJob{
		JobID:      "DATA-JOB-001",
		DataSource: "customer_database",
		Metadata:   make(map[string]interface{}),
		Errors:     []string{},
	}

	fmt.Printf("📝 开始数据处理作业: %s\n", job.JobID)

	// 步骤1: 数据获取
	fmt.Println("\n--- 步骤1: 数据获取 ---")
	err := executor.executeWithRetry(ctx, "数据获取", func() error {
		result, err := dataIngestionStep(ctx, job)
		if err == nil {
			job = result
		}
		return err
	})
	if err != nil {
		fmt.Printf("❌ 数据获取步骤最终失败: %v\n", err)
		return
	}

	// 步骤2: 数据清洗
	fmt.Println("\n--- 步骤2: 数据清洗 ---")
	err = executor.executeWithRetry(ctx, "数据清洗", func() error {
		result, err := dataCleaningStep(ctx, job)
		if err == nil {
			job = result
		}
		return err
	})
	if err != nil {
		fmt.Printf("❌ 数据清洗步骤最终失败: %v\n", err)
		return
	}

	// 步骤3: 数据分析
	fmt.Println("\n--- 步骤3: 数据分析 ---")
	err = executor.executeWithRetry(ctx, "数据分析", func() error {
		result, err := dataAnalysisStep(ctx, job)
		if err == nil {
			job = result
		}
		return err
	})
	if err != nil {
		fmt.Printf("❌ 数据分析步骤最终失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 数据处理作业完成！最终状态: %s\n", job.CurrentStage)
}

// 审批工作流演示
func approvalWorkflowDemo(ctx context.Context) {
	fmt.Println("\n=== 审批工作流演示 ===")

	executor := NewPracticalWorkflowExecutor(3)

	// 测试不同金额的申请
	requests := []ApprovalRequest{
		{RequestID: "REQ-001", RequestType: "设备采购", Amount: 500, Requester: "张三"},
		{RequestID: "REQ-002", RequestType: "差旅报销", Amount: 5000, Requester: "李四"},
		{RequestID: "REQ-003", RequestType: "投资决策", Amount: 50000, Requester: "王五"},
	}

	for _, request := range requests {
		fmt.Printf("\n📝 处理申请: %s (金额: %.2f)\n", request.RequestID, request.Amount)

		request.Metadata = make(map[string]interface{})
		request.Comments = []string{}

		currentRequest := request

		// 自动审批步骤
		err := executor.executeWithRetry(ctx, "自动审批", func() error {
			result, err := autoApprovalStep(ctx, currentRequest)
			if err == nil {
				currentRequest = result
			}
			return err
		})

		// 如果自动审批失败（需要人工审批），继续人工审批
		if err != nil {
			fmt.Println("🔄 转入人工审批流程")

			err = executor.executeWithRetry(ctx, "人工审批", func() error {
				result, err := manualApprovalStep(ctx, currentRequest)
				if err == nil {
					currentRequest = result
				}
				return err
			})
		}

		if err != nil {
			fmt.Printf("❌ 申请 %s 处理失败: %v\n", request.RequestID, err)
		} else {
			fmt.Printf("✅ 申请处理完成！状态: %s\n", currentRequest.Status)
			if len(currentRequest.Comments) > 0 {
				fmt.Printf("💬 最终评论: %s\n", strings.Join(currentRequest.Comments, "; "))
			}
		}
	}
}

// API调用工作流演示
func apiCallWorkflowDemo(ctx context.Context) {
	fmt.Println("\n=== API调用工作流演示 ===")

	executor := NewPracticalWorkflowExecutor(3)

	// 初始化API调用作业
	job := APICallJob{
		JobID:       "API-JOB-001",
		APIEndpoint: "https://api.example.com/data",
		Metadata:    make(map[string]interface{}),
	}

	fmt.Printf("📝 开始API批量调用作业: %s\n", job.JobID)

	// 执行API批量调用
	err := executor.executeWithRetry(ctx, "API批量调用", func() error {
		result, err := apiBatchCallStep(ctx, job)
		if err == nil {
			job = result
		}
		return err
	})

	if err != nil {
		fmt.Printf("❌ API调用作业最终失败: %v\n", err)
	} else {
		fmt.Printf("✅ API调用作业完成！总计: %d 成功: %d 失败: %d\n",
			job.RequestCount, job.SuccessCount, job.FailureCount)
	}
}

func initPracticalConfig() {
	viper.SetConfigFile("../config.yaml")
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("读取配置文件失败: %v (使用默认配置)", err)
	}
}

// RunPracticalExamplesDemo 运行实际应用场景中断检查点演示
// 这是实际应用场景演示的主入口函数，展示三种典型的业务场景：
// 1. 数据处理工作流 - 展示数据获取、清洗、分析过程中的中断处理
// 2. 审批工作流 - 展示自动审批和人工审批的中断机制
// 3. API调用工作流 - 展示批量API调用中的错误处理和重试机制
// 每个演示都包含完整的中断触发、处理和恢复流程
func main() {
	initPracticalConfig()
	ctx := context.Background()

	// 设置随机种子
	rand.Seed(time.Now().UnixNano())

	// 运行实际应用场景演示
	dataProcessingWorkflowDemo(ctx)
	approvalWorkflowDemo(ctx)
	apiCallWorkflowDemo(ctx)

	fmt.Println("\n🎉 实际应用场景演示完成！")
}
