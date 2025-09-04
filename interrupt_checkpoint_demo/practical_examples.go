package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// 数据处理工作流相关结构
type DataProcessingJob struct {
	JobID         string                 `json:"job_id"`
	DataSource    string                 `json:"data_source"`
	ProcessedRows int                    `json:"processed_rows"`
	TotalRows     int                    `json:"total_rows"`
	CurrentStage  string                 `json:"current_stage"`
	Metadata      map[string]interface{} `json:"metadata"`
	Errors        []string               `json:"errors"`
}

// 审批工作流相关结构
type ApprovalRequest struct {
	RequestID     string                 `json:"request_id"`
	RequestType   string                 `json:"request_type"`
	Amount        float64                `json:"amount"`
	Requester     string                 `json:"requester"`
	Status        string                 `json:"status"`
	ApprovalLevel int                    `json:"approval_level"`
	Metadata      map[string]interface{} `json:"metadata"`
	Comments      []string               `json:"comments"`
}

// API调用工作流相关结构
type APICallJob struct {
	JobID        string                 `json:"job_id"`
	APIEndpoint  string                 `json:"api_endpoint"`
	RequestCount int                    `json:"request_count"`
	SuccessCount int                    `json:"success_count"`
	FailureCount int                    `json:"failure_count"`
	CurrentBatch int                    `json:"current_batch"`
	TotalBatches int                    `json:"total_batches"`
	Metadata     map[string]interface{} `json:"metadata"`
	LastError    string                 `json:"last_error"`
}

// 自定义中断错误类型
type InterruptError struct {
	Message     string
	RequiresFix bool
	Metadata    map[string]interface{}
}

func (e *InterruptError) Error() string {
	return e.Message
}

func NewInterruptError(message string, requiresFix bool) *InterruptError {
	return &InterruptError{
		Message:     message,
		RequiresFix: requiresFix,
		Metadata:    make(map[string]interface{}),
	}
}

// 1. 数据处理工作流演示
func dataIngestionStep(ctx context.Context, job DataProcessingJob) (DataProcessingJob, error) {
	fmt.Printf("📥 开始数据获取，作业ID: %s\n", job.JobID)

	job.CurrentStage = "数据获取"
	job.TotalRows = 1000000 // 模拟大数据集
	job.ProcessedRows = 0

	// 模拟数据获取过程
	for i := 0; i < 5; i++ {
		time.Sleep(200 * time.Millisecond)
		job.ProcessedRows += 200000
		fmt.Printf("📊 已获取数据: %d/%d 行 (%.1f%%)\n",
			job.ProcessedRows, job.TotalRows,
			float64(job.ProcessedRows)/float64(job.TotalRows)*100)
	}

	if job.Metadata == nil {
		job.Metadata = make(map[string]interface{})
	}
	job.Metadata["ingestion_completed"] = time.Now().Format(time.RFC3339)
	fmt.Printf("✅ 数据获取完成\n")

	return job, nil
}

func dataCleaningStep(ctx context.Context, job DataProcessingJob) (DataProcessingJob, error) {
	fmt.Printf("🧹 开始数据清洗，作业ID: %s\n", job.JobID)

	job.CurrentStage = "数据清洗"

	// 模拟发现数据质量问题
	dataQualityScore := rand.Float64()
	if job.Metadata == nil {
		job.Metadata = make(map[string]interface{})
	}
	job.Metadata["data_quality_score"] = dataQualityScore

	if dataQualityScore < 0.7 {
		errorMsg := fmt.Sprintf("数据质量分数过低: %.2f, 需要人工介入", dataQualityScore)
		job.Errors = append(job.Errors, errorMsg)
		fmt.Printf("⚠️  %s\n", errorMsg)

		// 返回中断错误，需要修复
		return job, NewInterruptError(errorMsg, true)
	}

	// 模拟清洗过程
	job.ProcessedRows = 0
	for i := 0; i < 10; i++ {
		time.Sleep(100 * time.Millisecond)
		job.ProcessedRows += job.TotalRows / 10
		fmt.Printf("📊 数据清洗进度: %d/%d 行\n", job.ProcessedRows, job.TotalRows)
	}

	job.Metadata["cleaning_completed"] = time.Now().Format(time.RFC3339)
	fmt.Printf("✅ 数据清洗完成\n")

	return job, nil
}

func dataAnalysisStep(ctx context.Context, job DataProcessingJob) (DataProcessingJob, error) {
	fmt.Printf("📈 开始数据分析，作业ID: %s\n", job.JobID)

	job.CurrentStage = "数据分析"

	// 模拟长时间运行的分析任务
	analysisSteps := []string{
		"统计分析",
		"关联分析",
		"趋势分析",
		"异常检测",
		"模型训练",
	}

	for i, step := range analysisSteps {
		fmt.Printf("🔄 执行 %s...\n", step)
		time.Sleep(500 * time.Millisecond)

		// 模拟在某个步骤可能出现资源不足
		if step == "模型训练" && rand.Float64() < 0.4 {
			errorMsg := "GPU资源不足，需要等待资源释放"
			job.Errors = append(job.Errors, errorMsg)
			fmt.Printf("⚠️  %s\n", errorMsg)

			return job, NewInterruptError(errorMsg, true)
		}

		if job.Metadata == nil {
			job.Metadata = make(map[string]interface{})
		}
		job.Metadata[fmt.Sprintf("step_%d_completed", i)] = step
	}

	job.Metadata["analysis_completed"] = time.Now().Format(time.RFC3339)
	fmt.Printf("✅ 数据分析完成\n")

	return job, nil
}

// 2. 审批工作流演示
func autoApprovalStep(ctx context.Context, request ApprovalRequest) (ApprovalRequest, error) {
	fmt.Printf("🤖 自动审批检查，申请ID: %s\n", request.RequestID)

	request.Status = "自动审批中"

	// 自动审批规则
	if request.Amount <= 1000 {
		request.Status = "自动批准"
		request.Comments = append(request.Comments, "小额申请，自动批准")
		fmt.Printf("✅ 小额申请自动批准: %.2f\n", request.Amount)
		return request, nil
	}

	if request.Amount <= 10000 {
		request.Status = "需要主管审批"
		request.ApprovalLevel = 1
		request.Comments = append(request.Comments, "中额申请，需要主管审批")
		fmt.Printf("⏸️  中额申请需要主管审批: %.2f\n", request.Amount)

		return request, NewInterruptError(fmt.Sprintf("需要主管审批: %.2f", request.Amount), false)
	}

	// 大额申请需要多级审批
	request.Status = "需要高级管理层审批"
	request.ApprovalLevel = 2
	request.Comments = append(request.Comments, "大额申请，需要高级管理层审批")
	fmt.Printf("⏸️  大额申请需要高级管理层审批: %.2f\n", request.Amount)

	return request, NewInterruptError(fmt.Sprintf("需要高级管理层审批: %.2f", request.Amount), false)
}

func manualApprovalStep(ctx context.Context, request ApprovalRequest) (ApprovalRequest, error) {
	fmt.Printf("👤 人工审批处理，申请ID: %s\n", request.RequestID)

	// 模拟人工审批决策
	approvalDecision := rand.Float64()

	if approvalDecision > 0.8 {
		request.Status = "审批拒绝"
		request.Comments = append(request.Comments, "风险过高，审批拒绝")
		fmt.Printf("❌ 审批被拒绝\n")
		return request, nil
	}

	if approvalDecision > 0.6 {
		request.Status = "需要补充材料"
		request.Comments = append(request.Comments, "需要提供更多支持材料")
		fmt.Printf("⏸️  需要补充材料\n")

		return request, NewInterruptError("需要补充材料", true)
	}

	request.Status = "审批通过"
	request.Comments = append(request.Comments, fmt.Sprintf("经%d级审批通过", request.ApprovalLevel))
	fmt.Printf("✅ 审批通过\n")

	return request, nil
}

// 3. API调用工作流演示
func apiBatchCallStep(ctx context.Context, job APICallJob) (APICallJob, error) {
	fmt.Printf("🌐 执行API批量调用，作业ID: %s\n", job.JobID)

	// 模拟API批量调用
	batchSize := 10
	totalRequests := 100

	job.TotalBatches = totalRequests / batchSize
	job.RequestCount = 0
	job.SuccessCount = 0
	job.FailureCount = 0

	for batch := 1; batch <= job.TotalBatches; batch++ {
		job.CurrentBatch = batch
		fmt.Printf("📦 处理批次 %d/%d\n", batch, job.TotalBatches)

		// 模拟批次处理
		for i := 0; i < batchSize; i++ {
			job.RequestCount++

			// 模拟API调用成功率
			if rand.Float64() > 0.1 { // 90% 成功率
				job.SuccessCount++
			} else {
				job.FailureCount++
				job.LastError = fmt.Sprintf("批次 %d 中的请求 %d 失败", batch, i+1)
			}
		}

		// 检查是否需要中断（错误率过高）
		if job.RequestCount > 20 && float64(job.FailureCount)/float64(job.RequestCount) > 0.2 {
			errorMsg := fmt.Sprintf("API错误率过高: %.1f%%",
				float64(job.FailureCount)/float64(job.RequestCount)*100)
			fmt.Printf("⚠️  %s\n", errorMsg)

			return job, NewInterruptError(errorMsg, true)
		}

		time.Sleep(200 * time.Millisecond) // 模拟批次间延迟
	}

	if job.Metadata == nil {
		job.Metadata = make(map[string]interface{})
	}
	job.Metadata["completion_time"] = time.Now().Format(time.RFC3339)
	fmt.Printf("✅ API批量调用完成，成功: %d, 失败: %d\n", job.SuccessCount, job.FailureCount)

	return job, nil
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
