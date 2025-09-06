// Package main 演示检查点存储机制的实现和使用
// 该文件展示了如何在工作流执行过程中保存和恢复检查点，
// 支持内存存储和文件存储两种方式，实现工作流的中断和恢复功能
package main

import (
	"context"       // 上下文管理
	"encoding/json" // JSON序列化和反序列化
	"fmt"           // 格式化输出
	"log"           // 日志记录
	"os"            // 操作系统接口
	"path/filepath" // 文件路径操作
	"sync"          // 同步原语
	"time"          // 时间处理

	"github.com/spf13/viper" // 配置文件管理
)

// CheckpointData 表示检查点数据结构
// 用于保存工作流执行过程中的状态信息，包括已处理的数据项、
// 当前执行步骤和相关元数据，支持工作流的中断和恢复
type CheckpointData struct {
	ProcessedItems []string               `json:"processed_items"` // 已处理的数据项列表
	CurrentStep    int                    `json:"current_step"`    // 当前执行步骤的索引
	Metadata       map[string]interface{} `json:"metadata"`        // 存储额外的元数据信息
}

// CheckpointStore 定义检查点存储接口
// 提供统一的检查点数据存储和检索功能，支持不同的存储后端实现
// 如内存存储、文件存储等，确保检查点数据的持久化和可恢复性
type CheckpointStore interface {
	// Get 根据键获取检查点数据
	// 参数: ctx - 上下文, key - 检查点键
	// 返回: value - 检查点数据字节数组, existed - 是否存在, err - 错误信息
	Get(ctx context.Context, key string) (value []byte, existed bool, err error)

	// Set 保存检查点数据
	// 参数: ctx - 上下文, key - 检查点键, value - 检查点数据字节数组
	// 返回: err - 错误信息
	Set(ctx context.Context, key string, value []byte) error
}

// MemoryCheckpointStore 内存检查点存储实现
// 将检查点数据存储在内存中，适用于临时存储和快速访问场景
// 注意：程序重启后数据会丢失，不适用于需要持久化的场景
type MemoryCheckpointStore struct {
	data map[string][]byte // 存储检查点数据的映射表，键为检查点ID，值为序列化后的数据
	mu   sync.RWMutex      // 读写锁，保证并发安全访问
}

// NewMemoryCheckpointStore 创建新的内存检查点存储实例
// 初始化内部数据结构，返回可用的内存存储对象
// 返回: *MemoryCheckpointStore - 内存检查点存储实例
func NewMemoryCheckpointStore() *MemoryCheckpointStore {
	return &MemoryCheckpointStore{
		data: make(map[string][]byte), // 初始化存储映射表
	}
}

// Get 从内存中获取检查点数据
// 使用读锁保证并发安全，根据键查找对应的检查点数据
// 参数: ctx - 上下文, key - 检查点键
// 返回: value - 检查点数据, existed - 是否存在, err - 错误信息
func (m *MemoryCheckpointStore) Get(ctx context.Context, key string) (value []byte, existed bool, err error) {
	m.mu.RLock()         // 获取读锁
	defer m.mu.RUnlock() // 确保函数结束时释放读锁

	value, existed = m.data[key] // 从映射表中查找数据
	fmt.Printf("📖 从内存读取检查点 '%s': %t\n", key, existed)
	return value, existed, nil
}

// Set 将检查点数据保存到内存中
// 使用写锁保证并发安全，将数据存储到内部映射表中
// 参数: ctx - 上下文, key - 检查点键, value - 检查点数据
// 返回: err - 错误信息
func (m *MemoryCheckpointStore) Set(ctx context.Context, key string, value []byte) error {
	m.mu.Lock()         // 获取写锁
	defer m.mu.Unlock() // 确保函数结束时释放写锁

	m.data[key] = value // 将数据存储到映射表中
	fmt.Printf("💾 保存检查点到内存 '%s': %d bytes\n", key, len(value))
	return nil
}

// FileCheckpointStore 文件检查点存储实现
// 将检查点数据持久化存储到文件系统中，支持程序重启后的数据恢复
// 适用于需要长期保存检查点数据的场景，提供可靠的持久化存储
type FileCheckpointStore struct {
	baseDir string // 检查点文件存储的基础目录路径
}

// NewFileCheckpointStore 创建新的文件检查点存储实例
// 自动创建存储目录（如果不存在），初始化文件存储对象
// 参数: baseDir - 检查点文件存储的基础目录
// 返回: *FileCheckpointStore - 文件检查点存储实例
func NewFileCheckpointStore(baseDir string) *FileCheckpointStore {
	// 确保目录存在，创建存储目录，权限设置为755
	os.MkdirAll(baseDir, 0755)
	return &FileCheckpointStore{baseDir: baseDir}
}

// Get 从文件中读取检查点数据
// 根据键构造文件路径，读取对应的检查点文件内容
// 参数: ctx - 上下文, key - 检查点键
// 返回: value - 检查点数据, existed - 文件是否存在, err - 错误信息
func (f *FileCheckpointStore) Get(ctx context.Context, key string) (value []byte, existed bool, err error) {
	filePath := filepath.Join(f.baseDir, key+".json") // 构造检查点文件路径

	data, err := os.ReadFile(filePath) // 读取文件内容
	if os.IsNotExist(err) {
		fmt.Printf("📖 文件检查点 '%s' 不存在\n", key)
		return nil, false, nil // 文件不存在，返回false
	} else if err != nil {
		return nil, false, fmt.Errorf("读取检查点文件失败: %w", err) // 读取文件出错
	}

	fmt.Printf("📖 从文件读取检查点 '%s': %d bytes\n", key, len(data))
	return data, true, nil // 成功读取文件
}

// Set 将检查点数据保存到文件中
// 根据键构造文件路径，将数据写入对应的检查点文件
// 参数: ctx - 上下文, key - 检查点键, value - 检查点数据
// 返回: err - 错误信息
func (f *FileCheckpointStore) Set(ctx context.Context, key string, value []byte) error {
	filePath := filepath.Join(f.baseDir, key+".json") // 构造检查点文件路径

	err := os.WriteFile(filePath, value, 0644) // 写入文件，权限设置为644
	if err != nil {
		return fmt.Errorf("写入检查点文件失败: %w", err) // 写入文件出错
	}

	fmt.Printf("💾 保存检查点到文件 '%s': %d bytes\n", key, len(value))
	return nil // 成功保存文件
}

// CheckpointWorkflowExecutor 支持检查点的工作流执行器
// 集成检查点存储功能，支持工作流的中断和恢复机制
// 可以在指定节点后自动创建检查点，实现工作流的断点续传功能
type CheckpointWorkflowExecutor struct {
	store          CheckpointStore // 检查点存储后端，支持内存和文件存储
	interruptAfter []string        // 需要在执行后中断的节点名称列表
}

// NewCheckpointWorkflowExecutor 创建新的检查点工作流执行器
// 初始化执行器实例，配置检查点存储后端
// 参数: store - 检查点存储实现（内存或文件存储）
// 返回: *CheckpointWorkflowExecutor - 检查点工作流执行器实例
func NewCheckpointWorkflowExecutor(store CheckpointStore) *CheckpointWorkflowExecutor {
	return &CheckpointWorkflowExecutor{
		store:          store,      // 设置检查点存储后端
		interruptAfter: []string{}, // 初始化中断节点列表为空
	}
}

// WithInterruptAfterNodes 配置需要在执行后中断的节点
// 设置工作流中哪些节点执行完成后需要创建检查点并中断执行
// 参数: nodes - 需要中断的节点名称列表
// 返回: *CheckpointWorkflowExecutor - 返回自身以支持链式调用
func (cwe *CheckpointWorkflowExecutor) WithInterruptAfterNodes(nodes []string) *CheckpointWorkflowExecutor {
	cwe.interruptAfter = nodes // 设置中断节点列表
	return cwe                 // 返回自身支持链式调用
}

// contains 检查字符串切片中是否包含指定项
// 辅助方法，用于判断当前节点是否在中断节点列表中
// 参数: slice - 字符串切片, item - 要查找的字符串
// 返回: bool - 是否包含指定项
func (cwe *CheckpointWorkflowExecutor) contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true // 找到匹配项
		}
	}
	return false // 未找到匹配项
}

// generateCheckpointID 生成唯一的检查点标识符
// 使用当前时间的纳秒时间戳生成唯一ID，确保每个检查点都有不同的标识
// 返回: string - 格式为"checkpoint_纳秒时间戳"的检查点ID
func (cwe *CheckpointWorkflowExecutor) generateCheckpointID() string {
	return fmt.Sprintf("checkpoint_%d", time.Now().UnixNano()) // 基于纳秒时间戳生成唯一ID
}

// saveCheckpoint 保存检查点数据到存储后端
// 将检查点数据序列化为JSON格式，然后保存到配置的存储后端中
// 参数: ctx - 上下文, checkpointID - 检查点ID, data - 要保存的检查点数据
// 返回: error - 保存过程中的错误信息
func (cwe *CheckpointWorkflowExecutor) saveCheckpoint(ctx context.Context, checkpointID string, data CheckpointData) error {
	jsonData, err := json.Marshal(data) // 将检查点数据序列化为JSON
	if err != nil {
		return fmt.Errorf("序列化检查点数据失败: %w", err)
	}

	return cwe.store.Set(ctx, checkpointID, jsonData) // 保存到存储后端
}

// loadCheckpoint 从存储后端加载检查点数据
// 根据检查点ID从存储后端读取数据，并反序列化为CheckpointData结构
// 参数: ctx - 上下文, checkpointID - 检查点ID
// 返回: CheckpointData - 检查点数据, error - 加载过程中的错误信息
func (cwe *CheckpointWorkflowExecutor) loadCheckpoint(ctx context.Context, checkpointID string) (CheckpointData, error) {
	var data CheckpointData

	jsonData, existed, err := cwe.store.Get(ctx, checkpointID) // 从存储后端读取数据
	if err != nil {
		return data, fmt.Errorf("读取检查点失败: %w", err)
	}

	if !existed {
		return data, fmt.Errorf("检查点 %s 不存在", checkpointID) // 检查点不存在
	}

	err = json.Unmarshal(jsonData, &data) // 反序列化JSON数据
	if err != nil {
		return data, fmt.Errorf("反序列化检查点数据失败: %w", err)
	}

	return data, nil // 成功加载检查点数据
}

// DataProcessingStep 数据处理步骤函数类型
// 定义工作流中每个处理步骤的函数签名，接收上下文和检查点数据，
// 返回处理后的检查点数据和可能的错误信息
type DataProcessingStep func(ctx context.Context, data CheckpointData) (CheckpointData, error)

// createDataProcessingNode 创建数据处理节点
// 工厂函数，根据节点名称和步骤编号创建具体的数据处理步骤
// 每个步骤会模拟数据处理过程，更新检查点数据并记录执行信息
// 参数: name - 节点名称, stepNum - 步骤编号
// 返回: DataProcessingStep - 数据处理步骤函数
func createDataProcessingNode(name string, stepNum int) DataProcessingStep {
	return func(ctx context.Context, data CheckpointData) (CheckpointData, error) {
		fmt.Printf("🔄 执行 %s (步骤 %d)\n", name, stepNum)

		// 模拟数据处理：生成处理结果并添加到已处理项目列表
		newItem := fmt.Sprintf("处理结果-%s", name)
		data.ProcessedItems = append(data.ProcessedItems, newItem)
		data.CurrentStep = stepNum // 更新当前步骤编号

		// 更新元数据：记录步骤完成信息
		if data.Metadata == nil {
			data.Metadata = make(map[string]interface{}) // 初始化元数据映射
		}
		data.Metadata[name] = fmt.Sprintf("完成于步骤%d", stepNum) // 记录步骤完成信息

		// 模拟处理时间，增加真实感
		time.Sleep(500 * time.Millisecond)

		fmt.Printf("📊 当前状态: 已处理 %d 项, 当前步骤: %d\n", len(data.ProcessedItems), data.CurrentStep)

		return data, nil // 返回更新后的检查点数据
	}
}

// executeSteps 执行工作流步骤
// 按照指定顺序执行工作流中的各个步骤，支持在指定节点后创建检查点并中断执行
// 参数: ctx - 上下文, data - 初始检查点数据, steps - 步骤映射表, stepOrder - 步骤执行顺序
// 返回: CheckpointData - 执行结果数据, error - 执行过程中的错误或中断信息
func (cwe *CheckpointWorkflowExecutor) executeSteps(ctx context.Context, data CheckpointData, steps map[string]DataProcessingStep, stepOrder []string) (CheckpointData, error) {
	currentData := data

	for _, stepName := range stepOrder {
		step, exists := steps[stepName]
		if !exists {
			return currentData, fmt.Errorf("步骤 %s 不存在", stepName)
		}

		// 执行步骤
		result, err := step(ctx, currentData)
		if err != nil {
			return result, err
		}

		currentData = result

		// 检查是否需要在此步骤后中断
		if cwe.contains(cwe.interruptAfter, stepName) {
			// 生成检查点ID并保存
			checkpointID := cwe.generateCheckpointID()
			err = cwe.saveCheckpoint(ctx, checkpointID, currentData)
			if err != nil {
				return currentData, fmt.Errorf("保存检查点失败: %w", err)
			}

			// 返回中断错误，包含检查点ID
			return currentData, &CheckpointInterruptError{
				Message:      fmt.Sprintf("在步骤 %s 后中断", stepName),
				CheckpointID: checkpointID,
				StepName:     stepName,
			}
		}
	}

	return currentData, nil
}

// resumeFromCheckpoint 从检查点恢复执行
// 从指定的检查点恢复工作流执行，跳过已完成的步骤，从指定步骤的下一步开始继续执行
// 参数: ctx - 上下文, checkpointID - 检查点ID, steps - 步骤映射表, stepOrder - 步骤执行顺序, resumeFromStep - 恢复起始步骤名称
// 返回: CheckpointData - 执行结果数据, error - 恢复执行过程中的错误信息
func (cwe *CheckpointWorkflowExecutor) resumeFromCheckpoint(ctx context.Context, checkpointID string, steps map[string]DataProcessingStep, stepOrder []string, resumeFromStep string) (CheckpointData, error) {
	// 加载检查点数据
	data, err := cwe.loadCheckpoint(ctx, checkpointID)
	if err != nil {
		return data, err
	}

	fmt.Printf("🔄 从检查点恢复执行，检查点ID: %s\n", checkpointID)

	// 找到恢复点
	resumeIndex := -1
	for i, stepName := range stepOrder {
		if stepName == resumeFromStep {
			resumeIndex = i + 1 // 从下一个步骤开始
			break
		}
	}

	if resumeIndex == -1 {
		return data, fmt.Errorf("找不到恢复步骤 %s", resumeFromStep)
	}

	// 从恢复点继续执行
	if resumeIndex < len(stepOrder) {
		remainingSteps := stepOrder[resumeIndex:]
		return cwe.executeSteps(ctx, data, steps, remainingSteps)
	}

	// 已经是最后一步，直接返回
	return data, nil
}

// CheckpointInterruptError 检查点中断错误类型
// 当工作流在指定节点后需要创建检查点并中断执行时抛出的特殊错误
// 包含中断信息、检查点ID和步骤名称，用于后续的恢复执行
type CheckpointInterruptError struct {
	Message      string // 中断消息描述
	CheckpointID string // 创建的检查点ID
	StepName     string // 中断时的步骤名称
}

// Error 实现error接口
// 返回格式化的错误信息，包含中断消息和检查点ID
// 返回: string - 格式化的错误信息
func (e *CheckpointInterruptError) Error() string {
	return fmt.Sprintf("检查点中断: %s (ID: %s)", e.Message, e.CheckpointID)
}

// memoryStoreDemo 内存存储演示函数
// 演示如何使用内存检查点存储进行工作流的中断和恢复
// 包含多个处理步骤，展示在不同节点的中断和恢复机制
// 参数: ctx - 上下文对象
func memoryStoreDemo(ctx context.Context) {
	fmt.Println("=== 内存检查点存储演示 ===")

	// 创建内存存储实例
	store := NewMemoryCheckpointStore()
	executor := NewCheckpointWorkflowExecutor(store)

	// 创建多个处理步骤，模拟完整的数据处理流水线
	steps := map[string]DataProcessingStep{
		"collect": createDataProcessingNode("数据收集", 1), // 第一步：数据收集
		"clean":   createDataProcessingNode("数据清洗", 2), // 第二步：数据清洗
		"analyze": createDataProcessingNode("数据分析", 3), // 第三步：数据分析
		"export":  createDataProcessingNode("结果导出", 4), // 第四步：结果导出
	}

	// 定义步骤执行顺序
	stepOrder := []string{"collect", "clean", "analyze", "export"}

	// 配置中断点：在数据清洗和数据分析步骤后创建检查点
	executor.WithInterruptAfterNodes([]string{"clean", "analyze"})

	// 准备初始数据
	initialData := CheckpointData{
		ProcessedItems: []string{},                   // 初始处理项目列表为空
		CurrentStep:    0,                            // 当前步骤为0
		Metadata:       make(map[string]interface{}), // 初始化元数据映射
	}

	fmt.Println("📝 开始执行数据处理流程...")

	var checkpointID string // 用于存储检查点ID

	// 第一次执行到第一个中断点
	result, err := executor.executeSteps(ctx, initialData, steps, stepOrder)
	if err != nil {
		if checkpointErr, ok := err.(*CheckpointInterruptError); ok {
			checkpointID = checkpointErr.CheckpointID
			fmt.Printf("⏸️  第一次中断，检查点ID: %s\n", checkpointID)

			// 从检查点恢复执行到第二个中断点
			fmt.Println("🔄 从第一个检查点恢复执行...")
			result, err = executor.resumeFromCheckpoint(ctx, checkpointID, steps, stepOrder, checkpointErr.StepName)

			if err != nil {
				if checkpointErr2, ok := err.(*CheckpointInterruptError); ok {
					checkpointID = checkpointErr2.CheckpointID
					fmt.Printf("⏸️  第二次中断，检查点ID: %s\n", checkpointID)

					// 从第二个检查点恢复并完成
					fmt.Println("🔄 从第二个检查点恢复执行...")
					result, err = executor.resumeFromCheckpoint(ctx, checkpointID, steps, stepOrder, checkpointErr2.StepName)
					if err != nil {
						log.Printf("最终执行失败: %v", err)
						return
					}
				}
			}
		} else {
			log.Printf("执行失败: %v", err)
			return
		}
	}

	fmt.Printf("✅ 内存存储演示完成，最终结果: %+v\n", result)
}

// fileStoreDemo 文件存储演示函数
// 演示如何使用文件检查点存储进行工作流的持久化存储和恢复
// 展示程序重启后从文件恢复检查点的能力
// 参数: ctx - 上下文对象
func fileStoreDemo(ctx context.Context) {
	fmt.Println("\n=== 文件检查点存储演示 ===")

	// 创建文件存储实例，指定检查点文件存储目录
	store := NewFileCheckpointStore("./checkpoints")
	executor := NewCheckpointWorkflowExecutor(store)

	// 创建一个复杂处理步骤，模拟耗时的数据处理任务
	complexStep := createDataProcessingNode("复杂处理", 1)

	// 构建步骤映射表
	steps := map[string]DataProcessingStep{
		"complex": complexStep, // 复杂处理步骤
	}

	// 定义步骤执行顺序
	stepOrder := []string{"complex"}

	// 配置中断点：在复杂处理步骤后创建检查点
	executor.WithInterruptAfterNodes([]string{"complex"})

	// 准备初始数据，包含预设的处理项目和元数据
	initialData := CheckpointData{
		ProcessedItems: []string{"初始数据"},                                   // 预设的初始数据项
		CurrentStep:    0,                                                  // 当前步骤为0
		Metadata:       map[string]interface{}{"start_time": "2024-01-01"}, // 包含开始时间的元数据
	}

	fmt.Println("📝 开始文件存储演示...")

	// 第一次执行
	result, err := executor.executeSteps(ctx, initialData, steps, stepOrder)
	if err != nil {
		if checkpointErr, ok := err.(*CheckpointInterruptError); ok {
			checkpointID := checkpointErr.CheckpointID
			fmt.Printf("⏸️  执行中断，检查点已保存到文件，ID: %s\n", checkpointID)

			// 演示从文件恢复
			fmt.Println("🔄 模拟程序重启，从文件恢复检查点...")

			// 创建新的executor（模拟程序重启）
			newStore := NewFileCheckpointStore("./checkpoints")
			newExecutor := NewCheckpointWorkflowExecutor(newStore)

			// 从检查点恢复（这里不设置中断，直接完成）
			result, err = newExecutor.resumeFromCheckpoint(ctx, checkpointID, steps, stepOrder, checkpointErr.StepName)
			if err != nil {
				log.Printf("从检查点恢复失败: %v", err)
				return
			}
		} else {
			log.Printf("执行失败: %v", err)
			return
		}
	}

	fmt.Printf("✅ 文件存储演示完成，最终结果: %+v\n", result)
}

// checkpointModificationDemo 检查点数据修改演示函数
// 演示如何读取、修改和保存检查点数据
// 展示检查点数据的灵活性和可编辑性，支持手动干预工作流执行
// 参数: ctx - 上下文对象
func checkpointModificationDemo(ctx context.Context) {
	fmt.Println("\n=== 检查点数据修改演示 ===")

	// 创建内存存储实例用于演示
	store := NewMemoryCheckpointStore()
	executor := NewCheckpointWorkflowExecutor(store)

	// 创建验证步骤，用于演示检查点修改功能
	validateStep := createDataProcessingNode("验证", 1)

	// 构建步骤映射表
	steps := map[string]DataProcessingStep{
		"validate": validateStep, // 数据验证步骤
	}

	// 定义步骤执行顺序
	stepOrder := []string{"validate"}

	// 配置中断点：在验证步骤后创建检查点
	executor.WithInterruptAfterNodes([]string{"validate"})

	// 准备初始数据，包含多个原始数据项
	initialData := CheckpointData{
		ProcessedItems: []string{"原始数据1", "原始数据2"},   // 预设的原始数据项
		CurrentStep:    0,                            // 当前步骤为0
		Metadata:       make(map[string]interface{}), // 初始化元数据映射
	}

	fmt.Println("📝 开始检查点修改演示...")

	// 执行到中断点
	result, err := executor.executeSteps(ctx, initialData, steps, stepOrder)
	if err != nil {
		if checkpointErr, ok := err.(*CheckpointInterruptError); ok {
			checkpointID := checkpointErr.CheckpointID
			fmt.Printf("⏸️  执行中断，检查点ID: %s\n", checkpointID)

			// 读取并展示检查点数据
			fmt.Println("🔧 读取检查点数据...")
			savedData, err := executor.loadCheckpoint(ctx, checkpointID)
			if err != nil {
				log.Printf("读取检查点数据失败: %v", err)
				return
			}

			fmt.Printf("📊 原始检查点数据: %+v\n", savedData)

			// 修改检查点数据
			fmt.Println("✏️  修改检查点数据...")
			savedData.ProcessedItems = append(savedData.ProcessedItems, "手工添加的数据")
			savedData.Metadata["modified"] = true
			savedData.Metadata["modification_time"] = time.Now().Format(time.RFC3339)

			// 保存修改后的检查点
			newCheckpointID := executor.generateCheckpointID()
			err = executor.saveCheckpoint(ctx, newCheckpointID, savedData)
			if err != nil {
				log.Printf("保存修改后的检查点失败: %v", err)
				return
			}

			fmt.Printf("💾 已保存修改后的检查点，新ID: %s\n", newCheckpointID)

			// 从修改后的检查点继续执行
			result, err = executor.resumeFromCheckpoint(ctx, newCheckpointID, steps, stepOrder, checkpointErr.StepName)
			if err != nil {
				log.Printf("从修改后的检查点恢复失败: %v", err)
				return
			}
		} else {
			log.Printf("执行失败: %v", err)
			return
		}
	}

	fmt.Printf("✅ 检查点修改演示完成，最终结果: %+v\n", result)
}

// initCheckpointConfig 初始化检查点配置
// 读取配置文件设置，如果配置文件不存在或读取失败则使用默认配置
// 配置文件用于设置检查点存储相关的参数和选项
func initCheckpointConfig() {
	viper.SetConfigFile("../config.yaml") // 设置配置文件路径
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("读取配置文件失败: %v (使用默认配置)", err) // 配置文件读取失败时的提示
	}
}

// RunCheckpointStoreDemo 运行检查点存储演示
// 主演示函数，展示内存存储、文件存储和检查点修改等功能
// 包含完整的检查点存储机制演示流程
func main() {
	initCheckpointConfig() // 初始化配置
	ctx := context.Background()

	// 运行各种检查点存储演示
	memoryStoreDemo(ctx)            // 内存存储演示
	fileStoreDemo(ctx)              // 文件存储演示
	checkpointModificationDemo(ctx) // 检查点修改演示

	fmt.Println("\n🎉 检查点存储演示完成！")
}
