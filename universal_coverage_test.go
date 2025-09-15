package runehammer

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

// TestUniversalEngineCoverage 专门用于测试新增通用引擎代码的覆盖率
func TestUniversalEngineCoverage(t *testing.T) {
	Convey("通用引擎覆盖率测试", t, func() {
		
		Convey("测试NewBaseEngine和NewTypedEngine", func() {
			// 测试NewBaseEngine
			baseEngine, err := NewBaseEngine(
				WithDSN("sqlite:file:coverage_test.db?mode=memory&cache=shared&_fk=1"),
				WithAutoMigrate(),
				WithLogger(NewNoopLogger()),
			)
			So(err, ShouldBeNil)
			So(baseEngine, ShouldNotBeNil)
			defer baseEngine.Close()
			
			// 测试NewTypedEngine
			typedEngine := NewTypedEngine[map[string]interface{}](baseEngine)
			So(typedEngine, ShouldNotBeNil)
			So(typedEngine.base, ShouldEqual, baseEngine)
		})
		
		Convey("测试baseEngineWrapper的ExecRaw方法", func() {
			// 创建mock引擎来测试wrapper
			mockEngine := &mockEngine[map[string]interface{}]{
				execResult: map[string]interface{}{"test": "value"},
				execError:  nil,
			}
			
			wrapper := &baseEngineWrapper{engine: mockEngine}
			
			result, err := wrapper.ExecRaw(context.Background(), "test", map[string]interface{}{"input": "test"})
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			So(result["test"], ShouldEqual, "value")
		})
		
		Convey("测试baseEngineWrapper的Close方法", func() {
			mockEngine := &mockEngine[map[string]interface{}]{
				closeError: nil,
			}
			
			wrapper := &baseEngineWrapper{engine: mockEngine}
			err := wrapper.Close()
			So(err, ShouldBeNil)
		})
		
		Convey("测试TypedEngine的Exec方法", func() {
			// 测试成功的情况
			mockBase := &mockBaseEngine{
				execRawResult: map[string]interface{}{"name": "test", "value": 42},
				execRawError:  nil,
			}
			
			typedEngine := &TypedEngine[TestStruct]{base: mockBase}
			
			result, err := typedEngine.Exec(context.Background(), "test", nil)
			So(err, ShouldBeNil)
			So(result.Name, ShouldEqual, "test")
			So(result.Value, ShouldEqual, 42)
			
			// 测试失败的情况
			mockBase.execRawError = ErrInvalidConfig
			_, err = typedEngine.Exec(context.Background(), "test", nil)
			So(err, ShouldNotBeNil)
		})
		
		Convey("测试TypedEngine的Close方法", func() {
			mockBase := &mockBaseEngine{closeError: nil}
			typedEngine := &TypedEngine[TestStruct]{base: mockBase}
			
			err := typedEngine.Close()
			So(err, ShouldBeNil)
		})
		
		Convey("测试convertToType的所有分支", func() {
			// 测试直接匹配的情况 - map[string]interface{}
			rawResult := map[string]interface{}{"key": "value"}
			result1, err := convertToType[map[string]interface{}](rawResult)
			So(err, ShouldBeNil)
			So(result1["key"], ShouldEqual, "value")
			
			// 测试map[string]any的情况
			result2, err := convertToType[map[string]any](rawResult)
			So(err, ShouldBeNil)
			So(result2["key"], ShouldEqual, "value")
			
			// 测试结构体转换的情况
			structData := map[string]interface{}{"name": "test", "value": 123}
			result3, err := convertToType[TestStruct](structData)
			So(err, ShouldBeNil)
			So(result3.Name, ShouldEqual, "test")
			So(result3.Value, ShouldEqual, 123)
			
			// 测试interface{}的情况
			result4, err := convertToType[interface{}](rawResult)
			So(err, ShouldBeNil)
			So(result4, ShouldNotBeNil)
		})
		
		Convey("测试convertMapToStruct的所有分支", func() {
			// 测试正常结构体转换
			rawResult := map[string]interface{}{"name": "test", "value": 456}
			result1, err := convertMapToStruct[TestStruct](rawResult)
			So(err, ShouldBeNil)
			So(result1.Name, ShouldEqual, "test")
			So(result1.Value, ShouldEqual, 456)
			
			// 测试nil类型的错误情况（interface{}的零值）
			_, err = convertMapToStruct[interface{}](rawResult)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "无法确定目标类型")
			
			// 测试错误情况：无法JSON序列化的数据
			invalidData := map[string]interface{}{
				"name":    "test",
				"invalid": make(chan int), // channel无法JSON序列化
			}
			_, err = convertMapToStruct[TestStruct](invalidData)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "JSON序列化失败")
			
			// 测试JSON反序列化失败的情况
			type BadStruct struct {
				Value int `json:"value,string"` // 要求string类型但给的是int
			}
			badData := map[string]interface{}{
				"value": map[string]interface{}{"nested": "object"}, // 复杂对象无法转换为int
			}
			_, err = convertMapToStruct[BadStruct](badData)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "JSON反序列化失败")
		})
		
		Convey("测试NewBaseEngine的错误情况", func() {
			// 测试数据库配置错误
			_, err := NewBaseEngine(
				WithDSN("invalid://invalid"),
				WithAutoMigrate(),
				WithLogger(NewNoopLogger()),
			)
			So(err, ShouldNotBeNil)
		})
	})
}

// 测试用结构体
type TestStruct struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

// Mock引擎实现，用于测试
type mockEngine[T any] struct {
	execResult T
	execError  error
	closeError error
}

func (m *mockEngine[T]) Exec(ctx context.Context, bizCode string, input any) (T, error) {
	return m.execResult, m.execError
}

func (m *mockEngine[T]) Close() error {
	return m.closeError
}

// Mock BaseEngine实现
type mockBaseEngine struct {
	execRawResult map[string]interface{}
	execRawError  error
	closeError    error
}

func (m *mockBaseEngine) ExecRaw(ctx context.Context, bizCode string, input any) (map[string]interface{}, error) {
	return m.execRawResult, m.execRawError
}

func (m *mockBaseEngine) Close() error {
	return m.closeError
}