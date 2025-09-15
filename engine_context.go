package runehammer

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/hyperjumptech/grule-rule-engine/ast"
)

// ============================================================================
// 数据上下文管理 - 处理规则执行时的数据注入和结果提取
// ============================================================================

// injectInputData 注入输入数据 - 将各种类型的输入数据注入到执行上下文
//
// 变量注入规则:
//   1. Map类型：将整个map作为"Params"变量注入
//   2. 结构体类型：作为单个对象注入，使用类型名（小写）作为变量名
//   3. 匿名结构体和其他类型：统一以"Params"名称注入
//
// 参数:
//   dataCtx - Grule数据上下文
//   input   - 输入数据，支持任意类型
//
// 返回值:
//   error - 注入过程中的错误
func (e *engineImpl[T]) injectInputData(dataCtx ast.IDataContext, input any) error {
	// 首先初始化result变量作为一个空的map
	result := make(map[string]any)
	if err := dataCtx.Add("result", result); err != nil {
		return fmt.Errorf("注入result变量失败: %w", err)
	}

	v := reflect.ValueOf(input)
	t := reflect.TypeOf(input)

	// 处理指针类型，获取实际的值和类型
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}

	switch v.Kind() {
	case reflect.Map:
		return e.injectMapData(dataCtx, v)
	case reflect.Struct:
		return e.injectStructData(dataCtx, input, t)
	default:
		return e.injectDefaultData(dataCtx, input)
	}
}

// injectMapData 注入Map类型数据 - 将整个map作为Params变量注入
func (e *engineImpl[T]) injectMapData(dataCtx ast.IDataContext, v reflect.Value) error {
	// 将整个map作为Params注入
	if err := dataCtx.Add("Params", v.Interface()); err != nil {
		return fmt.Errorf("注入Params变量失败: %w", err)
	}
	return nil
}

// injectStructData 注入结构体数据 - 将整个结构体作为单个对象注入
func (e *engineImpl[T]) injectStructData(dataCtx ast.IDataContext, input any, t reflect.Type) error {
	// 使用结构体类型名作为变量名，转为小写
	inputName := strings.ToLower(t.Name())
	if inputName == "" {
		inputName = "Params" // 匿名结构体使用统一的Params名称
	}
	
	if err := dataCtx.Add(inputName, input); err != nil {
		return fmt.Errorf("注入结构体 %s 失败: %w", inputName, err)
	}
	
	return nil
}

// injectDefaultData 注入其他类型数据 - 直接以Params名称注入
func (e *engineImpl[T]) injectDefaultData(dataCtx ast.IDataContext, input any) error {
	if err := dataCtx.Add("Params", input); err != nil {
		return fmt.Errorf("注入Params变量失败: %w", err)
	}
	return nil
}

// extractResult 提取执行结果 - 从执行上下文中提取result变量并转换为目标类型
//
// 支持的结果类型转换:
//   1. interface{}类型：直接返回
//   2. map类型：从grule上下文提取实际的map值
//   3. 指针类型：提取实际值后转换
//   4. 其他类型：通过JSON序列化/反序列化进行类型转换
//
// 参数:
//   dataCtx - Grule数据上下文
//
// 返回值:
//   T     - 转换后的结果，类型由泛型参数决定
//   error - 转换过程中的错误
func (e *engineImpl[T]) extractResult(dataCtx ast.IDataContext) (T, error) {
	var zero T

	// 获取result变量
	resultValue := dataCtx.Get("result")
	if resultValue == nil {
		// 如果规则没有设置result变量，返回零值
		return zero, nil
	}

	// 获取实际的值
	actualValue, err := resultValue.GetValue()
	if err != nil {
		return zero, fmt.Errorf("获取result值失败: %w", err)
	}

	// 获取实际的interface{}值
	actualData := actualValue.Interface()

	// 根据泛型类型进行相应的转换
	var result T
	resultType := reflect.TypeOf(result)

	switch resultType.Kind() {
	case reflect.Interface:
		return e.extractInterfaceResult(actualData)
	case reflect.Map:
		return e.extractMapResult(actualData)
	case reflect.Ptr:
		return e.extractPointerResult(actualData)
	default:
		return e.extractGenericResult(actualData)
	}
}

// extractInterfaceResult 提取interface{}类型结果
func (e *engineImpl[T]) extractInterfaceResult(resultValue interface{}) (T, error) {
	var zero T
	
	// interface{}类型直接返回
	if reflect.TypeOf(zero) == reflect.TypeOf((*any)(nil)).Elem() {
		return resultValue.(T), nil
	}
	
	return zero, fmt.Errorf("不支持的interface类型: %v", reflect.TypeOf(zero))
}

// extractMapResult 提取map类型结果
func (e *engineImpl[T]) extractMapResult(resultValue interface{}) (T, error) {
	var zero T
	
	if resultMap, ok := resultValue.(map[string]any); ok {
		return any(resultMap).(T), nil
	}
	
	return zero, fmt.Errorf("结果不是有效的map类型")
}

// extractPointerResult 提取指针类型结果
func (e *engineImpl[T]) extractPointerResult(resultValue interface{}) (T, error) {
	return any(resultValue).(T), nil
}

// extractGenericResult 提取其他类型结果 - 通过JSON序列化/反序列化转换
func (e *engineImpl[T]) extractGenericResult(resultValue interface{}) (T, error) {
	var zero T
	var result T
	
	// 通过JSON进行类型转换
	data, err := json.Marshal(resultValue)
	if err != nil {
		return zero, fmt.Errorf("序列化结果失败: %w", err)
	}
	
	if err := json.Unmarshal(data, &result); err != nil {
		return zero, fmt.Errorf("反序列化结果失败: %w", err)
	}
	
	return result, nil
}