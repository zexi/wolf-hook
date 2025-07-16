package vdfutil

import (
	"fmt"
	"strings"
)

// ConvertToVdf 将 map[string]interface{} 转换为 VDF 格式字符串
func ConvertToVdf(data map[string]interface{}) (string, error) {
	var result strings.Builder
	err := WriteVdfNode(&result, data, 0)
	if err != nil {
		return "", err
	}
	return result.String(), nil
}

func escapeVdfString(s string) string {
	return strings.ReplaceAll(s, "\"", `\\"`)
}

// WriteVdfNode 递归写入 VDF 节点
func WriteVdfNode(builder *strings.Builder, data map[string]interface{}, indent int) error {
	indentStr := strings.Repeat("\t", indent)

	for key, value := range data {
		builder.WriteString(indentStr)
		builder.WriteString("\"")
		builder.WriteString(escapeVdfString(key))
		builder.WriteString("\"")

		switch v := value.(type) {
		case map[string]interface{}:
			builder.WriteString("\n")
			builder.WriteString(indentStr)
			builder.WriteString("{\n")
			err := WriteVdfNode(builder, v, indent+1)
			if err != nil {
				return err
			}
			builder.WriteString(indentStr)
			builder.WriteString("}\n")
		case string:
			builder.WriteString("\t\"")
			builder.WriteString(escapeVdfString(v))
			builder.WriteString("\"\n")
		case int:
			builder.WriteString("\t\"")
			builder.WriteString(fmt.Sprintf("%d", v))
			builder.WriteString("\"\n")
		case float64:
			builder.WriteString("\t\"")
			builder.WriteString(fmt.Sprintf("%.0f", v))
			builder.WriteString("\"\n")
		case bool:
			builder.WriteString("\t\"")
			if v {
				builder.WriteString("1")
			} else {
				builder.WriteString("0")
			}
			builder.WriteString("\"\n")
		default:
			builder.WriteString("\t\"")
			builder.WriteString(escapeVdfString(fmt.Sprintf("%v", v)))
			builder.WriteString("\"\n")
		}
	}

	return nil
}
