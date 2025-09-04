package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/spf13/viper"
)

// 定义数据结构
type CheckpointData struct {
	ProcessedItems []string               `json:"processed_items"`
	CurrentStep    int                    `json:"current_step"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// 检查点存储接口
type CheckpointStore interface {
	Get(ctx context.Context, key string) (value []byte, existed bool, err error)
	Set(ctx context.Context, key string, value []byte) error
}

// 内存检查点存储实现
type MemoryCheckpointStore struct {
	data map[string][]byte
	mu   sync.RWMutex
}

func NewMemoryCheckpointStore() *MemoryCheckpointStore {
	return &MemoryCheckpointStore{
		data: make(map[string][]byte),
	}
}

func (m *MemoryCheckpointStore) Get(ctx context.Context, key string) (value []byte, existed bool, err error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	value, existed = m.data[key]
	fmt.Printf("📖 从内存读取检查点 '%s': %t\n", key, existed)
	return value, existed, nil
}

func (m *MemoryCheckpointStore) Set(ctx context.Context, key string, value []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.data[key] = value
	fmt.Printf("💾 保存检查点到内存 '%s': %d bytes\n", key, len(value))
	return nil
}

// 文件检查点存储实现
type FileCheckpointStore struct {
	baseDir string
}

func NewFileCheckpointStore(baseDir string) *FileCheckpointStore {
	// 确保目录存在
	os.MkdirAll(baseDir, 0755)
	return &FileCheckpointStore{baseDir: baseDir}
}

func (f *FileCheckpointStore) Get(ctx context.Context, key string) (value []byte, existed bool, err error) {
	filePath := filepath.Join(f.baseDir, key+".json")

	data, err := os.ReadFile(filePath)
	if os.IsNotExist(err) {
		fmt.Printf("📖 文件检查点 '%s' 不存在\n", key)
		return nil, false, nil
	} else if err != nil {
		return nil, false, fmt.Errorf("读取检查点文件失败: %w", err)
	}

	fmt.Printf("📖 从文件读取检查点 '%s': %d bytes\n", key, len(data))
	return data, true, nil
}

func (f *FileCheckpointStore) Set(ctx context.Context, key string, value []byte) error {
	filePath := filepath.Join(f.baseDir, key+".json")

	err := os.WriteFile(filePath, value, 0644)
	if err != nil {
		return fmt.Errorf("写入检查点文件失败: %w", err)
	}

	fmt.Printf("💾 保存检查点到文件 '%s': %d bytes\n", key, len(value))
	return nil
}

// 工作流执行器，支持检查点
type CheckpointWorkflowExecutor struct {
	store          CheckpointStore
	interruptAfter []string
}

func NewCheckpointWorkflowExecutor(store CheckpointStore) *CheckpointWorkflowExecutor {
	return &CheckpointWorkflowExecutor{
		store:          store,
		interruptAfter: []string{},
	}
}

func (cwe *CheckpointWorkflowExecutor) WithInterruptAfterNodes(nodes []string) *CheckpointWorkflowExecutor {
	cwe.interruptAfter = nodes
	return cwe
}

func (cwe *CheckpointWorkflowExecutor) contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// 生成检查点ID
func (cwe *CheckpointWorkflowExecutor) generateCheckpointID() string {
	return fmt.Sprintf("checkpoint_%d", time.Now().UnixNano())
}

// 保存检查点
func (cwe *CheckpointWorkflowExecutor) saveCheckpoint(ctx context.Context, checkpointID string, data CheckpointData) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("序列化检查点数据失败: %w", err)
	}

	return cwe.store.Set(ctx, checkpointID, jsonData)
}

// 加载检查点
func (cwe *CheckpointWorkflowExecutor) loadCheckpoint(ctx context.Context, checkpointID string) (CheckpointData, error) {
	var data CheckpointData

	jsonData, existed, err := cwe.store.Get(ctx, checkpointID)
	if err != nil {
		return data, fmt.Errorf("读取检查点失败: %w", err)
	}

	if !existed {
		return data, fmt.Errorf("检查点 %s 不存在", checkpointID)
	}

	err = json.Unmarshal(jsonData, &data)
	if err != nil {
		return data, fmt.Errorf("反序列化检查点数据失败: %w", err)
	}

	return data, nil
}

// 数据处理节点
type DataProcessingStep func(ctx context.Context, data CheckpointData) (CheckpointData, error)

func createDataProcessingNode(name string, stepNum int) DataProcessingStep {
	return func(ctx context.Context, data CheckpointData) (CheckpointData, error) {
		fmt.Printf("🔄 执行 %s (步骤 %d)\n", name, stepNum)

		// 模拟数据处理
		newItem := fmt.Sprintf("处理结果-%s", name)
		data.ProcessedItems = append(data.ProcessedItems, newItem)
		data.CurrentStep = stepNum

		// 更新元数据
		if data.Metadata == nil {
			data.Metadata = make(map[string]interface{})
		}
		data.Metadata[name] = fmt.Sprintf("完成于步骤%d", stepNum)

		// 模拟处理时间
		time.Sleep(500 * time.Millisecond)

		fmt.Printf("📊 当前状态: 已处理 %d 项, 当前步骤: %d\n", len(data.ProcessedItems), data.CurrentStep)

		return data, nil
	}
}

// 执行工作流步骤
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

// 从检查点恢复执行
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

// 检查点中断错误
type CheckpointInterruptError struct {
	Message      string
	CheckpointID string
	StepName     string
}

func (e *CheckpointInterruptError) Error() string {
	return fmt.Sprintf("检查点中断: %s (ID: %s)", e.Message, e.CheckpointID)
}

// 内存存储演示
func memoryStoreDemo(ctx context.Context) {
	fmt.Println("=== 内存检查点存储演示 ===")

	// 创建内存存储
	store := NewMemoryCheckpointStore()
	executor := NewCheckpointWorkflowExecutor(store)

	// 创建步骤
	steps := map[string]DataProcessingStep{
		"collect": createDataProcessingNode("数据收集", 1),
		"clean":   createDataProcessingNode("数据清洗", 2),
		"analyze": createDataProcessingNode("数据分析", 3),
		"export":  createDataProcessingNode("结果导出", 4),
	}

	stepOrder := []string{"collect", "clean", "analyze", "export"}

	// 配置在clean后和analyze后设置中断
	executor.WithInterruptAfterNodes([]string{"clean", "analyze"})

	// 初始数据
	initialData := CheckpointData{
		ProcessedItems: []string{},
		CurrentStep:    0,
		Metadata:       make(map[string]interface{}),
	}

	fmt.Println("📝 开始执行数据处理流程...")

	var checkpointID string

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

// 文件存储演示
func fileStoreDemo(ctx context.Context) {
	fmt.Println("\n=== 文件检查点存储演示 ===")

	// 创建文件存储
	store := NewFileCheckpointStore("./checkpoints")
	executor := NewCheckpointWorkflowExecutor(store)

	// 创建一个复杂处理步骤
	complexStep := createDataProcessingNode("复杂处理", 1)

	steps := map[string]DataProcessingStep{
		"complex": complexStep,
	}

	stepOrder := []string{"complex"}

	// 配置中断
	executor.WithInterruptAfterNodes([]string{"complex"})

	// 初始数据
	initialData := CheckpointData{
		ProcessedItems: []string{"初始数据"},
		CurrentStep:    0,
		Metadata:       map[string]interface{}{"start_time": "2024-01-01"},
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

// 检查点数据修改演示
func checkpointModificationDemo(ctx context.Context) {
	fmt.Println("\n=== 检查点数据修改演示 ===")

	store := NewMemoryCheckpointStore()
	executor := NewCheckpointWorkflowExecutor(store)

	// 创建验证步骤
	validateStep := createDataProcessingNode("验证", 1)

	steps := map[string]DataProcessingStep{
		"validate": validateStep,
	}

	stepOrder := []string{"validate"}

	// 配置中断
	executor.WithInterruptAfterNodes([]string{"validate"})

	// 初始数据
	initialData := CheckpointData{
		ProcessedItems: []string{"原始数据1", "原始数据2"},
		CurrentStep:    0,
		Metadata:       make(map[string]interface{}),
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

func initCheckpointConfig() {
	viper.SetConfigFile("../config.yaml")
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("读取配置文件失败: %v (使用默认配置)", err)
	}
}

func main() {
	initCheckpointConfig()
	ctx := context.Background()

	// 运行各种检查点存储演示
	memoryStoreDemo(ctx)
	fileStoreDemo(ctx)
	checkpointModificationDemo(ctx)

	fmt.Println("\n🎉 检查点存储演示完成！")
}
