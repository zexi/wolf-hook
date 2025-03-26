package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"yunion.io/x/log"
)

type writeHwdbController struct{}

func NewWriteHwdbController() http.Handler {
	return new(writeHwdbController)
}

type WriteHwdbParams struct {
	Path    string `json:"path"`    // 文件路径
	Content string `json:"content"` // 文件内容
}

func (w *writeHwdbController) ServeHTTP(resp http.ResponseWriter, r *http.Request) {
	log.Infof("Write hwdb request received: %s", r.URL.Path)

	// 解析请求参数
	params := new(WriteHwdbParams)
	if err := json.NewDecoder(r.Body).Decode(params); err != nil {
		log.Errorf("解析请求参数失败: %v", err)
		http.Error(resp, err.Error(), http.StatusBadRequest)
		return
	}

	// 确保目标目录存在
	dir := filepath.Dir(params.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Errorf("创建目录失败 %s: %v", dir, err)
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		return
	}

	// 写入文件
	if err := os.WriteFile(params.Path, []byte(params.Content), 0644); err != nil {
		log.Errorf("写入文件失败 %s: %v", params.Path, err)
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Infof("成功写入文件 %s", params.Path)
	resp.WriteHeader(http.StatusOK)
	resp.Write([]byte("OK"))
}
