package vdfutil

import (
	"strings"
	"testing"
)

func TestConvertToVdf_Basic(t *testing.T) {
	data := map[string]interface{}{
		"string":     "value",
		"int":        123,
		"bool_true":  true,
		"bool_false": false,
		"float":      456.0,
	}

	result, err := ConvertToVdf(data)
	if err != nil {
		t.Fatalf("ConvertToVdf failed: %v", err)
	}

	// 检查基本类型的输出格式
	if !strings.Contains(result, `"string"	"value"`) {
		t.Error("string value not found")
	}
	if !strings.Contains(result, `"int"	"123"`) {
		t.Error("int value not found")
	}
	if !strings.Contains(result, `"bool_true"	"1"`) {
		t.Error("bool true value not found")
	}
	if !strings.Contains(result, `"bool_false"	"0"`) {
		t.Error("bool false value not found")
	}
	if !strings.Contains(result, `"float"	"456"`) {
		t.Error("float value not found")
	}
}

func TestConvertToVdf_Nested(t *testing.T) {
	data := map[string]interface{}{
		"root": map[string]interface{}{
			"child": map[string]interface{}{
				"leaf": "val",
			},
		},
	}

	result, err := ConvertToVdf(data)
	if err != nil {
		t.Fatalf("ConvertToVdf failed: %v", err)
	}

	if !strings.Contains(result, `"root"`) || !strings.Contains(result, `"child"`) || !strings.Contains(result, `"leaf"	"val"`) {
		t.Errorf("nested structure not formatted as expected: %s", result)
	}
}

func TestConvertToVdf_SpecialChars(t *testing.T) {
	data := map[string]interface{}{
		`k"y`: `v"l`,
	}

	result, err := ConvertToVdf(data)
	if err != nil {
		t.Fatalf("ConvertToVdf failed: %v", err)
	}

	// 检查特殊字符是否正确转义
	if !strings.Contains(result, `"k\\"y"	"v\\"l"`) {
		t.Errorf("special chars not handled: %s", result)
	}
}
