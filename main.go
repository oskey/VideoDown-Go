package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// 版本信息变量（编译时注入）
var (
	Version = "V1.0.0" // 默认版本号，可通过 -ldflags 覆盖
)

// 应用信息结构体
type AppInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Title   string `json:"title"`
}

// 请求结构体
type RunRequest struct {
	Platform string `json:"platform"`
	URL      string `json:"url"`
	TaskID   string `json:"taskID"` // 添加任务ID字段
	Config   Config `json:"config"` // 添加配置字段
}

// 停止请求结构体
type StopRequest struct {
	TaskID string `json:"taskID"` // 要停止的任务ID
}

// 视频文件信息结构体
type VideoInfo struct {
	Name      string    `json:"name"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"created_at"`
}

// 客户端连接信息
type ClientInfo struct {
	Conn   *websocket.Conn
	TaskID string // 当前任务ID
}

// WebSocket消息结构体
type WSMessage struct {
	TaskID  string `json:"taskID"`
	Message string `json:"message"`
	Type    string `json:"type"` // "log", "progress", "complete", "error"
}

// 配置结构体
type Config struct {
	EnableAdvanced       bool   `json:"enableAdvanced"`
	DownloadType         string `json:"downloadType"`
	SeparateDownload     string `json:"separateDownload"`
	VideoResolution      string `json:"videoResolution"`
	AudioFormat          string `json:"audioFormat"`
	DownloadSubtitle     bool   `json:"downloadSubtitle"`
	DownloadAutoSubtitle bool   `json:"downloadAutoSubtitle"`
	SubtitleLanguage     string `json:"subtitleLanguage"`
	EmbedSubtitle        bool   `json:"embedSubtitle"`
	SubtitleOnly         bool   `json:"subtitleOnly"`
	PlaylistStart        int    `json:"playlistStart"`
	PlaylistEnd          int    `json:"playlistEnd"`
	PlaylistMode         string `json:"playlistMode"`
	EnableThreads        bool   `json:"enableThreads"`
	ThreadCount          int    `json:"threadCount"`
	EnableRateLimit      bool   `json:"enableRateLimit"`
	RateLimit            string `json:"rateLimit"`
	ContinueOnError      bool   `json:"continueOnError"`
	EnableReferer        bool   `json:"enableReferer"`
}

// 版本信息结构体
type VersionInfo struct {
	CurrentVersion string `json:"currentVersion"`
	LatestVersion  string `json:"latestVersion"`
	HasUpdate      bool   `json:"hasUpdate"`
	DownloadURL    string `json:"downloadURL"`
}

// 更新请求结构体
type UpdateRequest struct {
	TaskID string `json:"taskID"`
}

// 中文编码转换函数
func convertGBKToUTF8(gbkStr string) string {
	decoder := simplifiedchinese.GBK.NewDecoder()
	utf8Str, _, err := transform.String(decoder, gbkStr)
	if err != nil {
		return gbkStr // 如果转换失败，返回原字符串
	}
	return utf8Str
}

// 从yt-dlp输出中提取文件名
func extractFilename(output string) string {
	// 匹配 "Destination: filename" 格式
	if strings.Contains(output, "Destination:") {
		parts := strings.Split(output, "Destination:")
		if len(parts) > 1 {
			filename := strings.TrimSpace(parts[1])
			return filename
		}
	}

	// 匹配 "[download] filename" 格式
	if strings.Contains(output, "[download]") && strings.Contains(output, "has already been downloaded") {
		re := regexp.MustCompile(`\[download\]\s+(.+?)\s+has already been downloaded`)
		matches := re.FindStringSubmatch(output)
		if len(matches) > 1 {
			return strings.TrimSpace(matches[1])
		}
	}

	// 匹配 "[download] Downloading video X of Y" 后面的文件名
	if strings.Contains(output, "[download]") && strings.Contains(output, "Downloading") {
		re := regexp.MustCompile(`\[download\]\s+Downloading\s+.*?:\s+(.+)`)
		matches := re.FindStringSubmatch(output)
		if len(matches) > 1 {
			return strings.TrimSpace(matches[1])
		}
	}

	// 匹配 "Merging formats into" 格式
	if strings.Contains(output, "Merging formats into") {
		re := regexp.MustCompile(`Merging formats into "(.+?)"`)
		matches := re.FindStringSubmatch(output)
		if len(matches) > 1 {
			return strings.TrimSpace(matches[1])
		}
	}

	return ""
}

// 全局变量
var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true // 允许所有来源的WebSocket连接
		},
	}
	clients       = make(map[*websocket.Conn]*ClientInfo) // 存储所有WebSocket连接及其信息
	clientsMu     sync.Mutex                              // 保护clients的互斥锁
	activeTasks   = make(map[string]*exec.Cmd)            // 存储活跃的下载任务
	tasksMu       sync.Mutex                              // 保护activeTasks的互斥锁
	taskFiles     = make(map[string]string)               // 存储任务对应的文件名
	filesMu       sync.Mutex                              // 保护taskFiles的互斥锁
	updateTasks   = make(map[string]context.CancelFunc)   // 存储活跃的更新任务
	updateTasksMu sync.Mutex                              // 保护updateTasks的互斥锁
)

func main() {
	// 设置静态文件服务
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// 设置路由
	http.HandleFunc("/", handleHome)
	http.HandleFunc("/ws", handleWebSocket)
	http.HandleFunc("/run", handleRun)
	http.HandleFunc("/stop", handleStop)
	http.HandleFunc("/api/videos", handleVideoList)
	http.HandleFunc("/api/video/", handleVideoStream)
	http.HandleFunc("/api/thumbnail/", handleThumbnail)
	http.HandleFunc("/api/delete", handleDelete)
	http.HandleFunc("/api/rename", handleRename)
	http.HandleFunc("/api/batch-delete", handleBatchDelete)
	http.HandleFunc("/api/config/save", handleConfigSave)
	http.HandleFunc("/api/config/load", handleConfigLoad)
	http.HandleFunc("/api/version/check", handleVersionCheck)
	http.HandleFunc("/api/version/update", handleVersionUpdate)
	http.HandleFunc("/api/version/cancel", handleVersionCancel)
	http.HandleFunc("/api/app/info", handleAppInfo)

	// 启动服务器
	port := "8888"
	log.Printf("服务器启动在 http://127.0.0.1:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// 处理主页请求
func handleHome(w http.ResponseWriter, r *http.Request) {
	templ := template.Must(template.ParseFiles("templates/index.html"))
	templ.Execute(w, nil)
}

// 处理WebSocket连接
func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	// 注册新客户端
	clientsMu.Lock()
	clients[conn] = &ClientInfo{
		Conn:   conn,
		TaskID: "", // 初始时没有任务ID
	}
	clientsMu.Unlock()
	defer func() {
		clientsMu.Lock()
		delete(clients, conn)
		clientsMu.Unlock()
	}()

	// 处理客户端消息
	for {
		var msg WSMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("读取WebSocket消息错误: %v", err)
			break
		}

		// 处理任务ID注册
		if msg.Type == "register" && msg.TaskID != "" {
			clientsMu.Lock()
			if clientInfo, exists := clients[conn]; exists {
				clientInfo.TaskID = msg.TaskID
				log.Printf("客户端注册任务ID: %s", msg.TaskID)
			}
			clientsMu.Unlock()
		}
	}
}

// 处理视频列表API请求
func handleVideoList(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 获取排序参数
	sortBy := r.URL.Query().Get("sort")
	if sortBy == "" {
		sortBy = "time" // 默认按创建时间排序
	}

	// 获取当前工作目录
	cwd, err := os.Getwd()
	if err != nil {
		http.Error(w, "Failed to get working directory", http.StatusInternalServerError)
		return
	}

	// 读取目录中的所有文件
	entries, err := os.ReadDir(cwd)
	if err != nil {
		http.Error(w, "Failed to read directory", http.StatusInternalServerError)
		return
	}

	videos := make([]VideoInfo, 0) // 初始化为空数组而不是nil切片
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		// 只处理.mp4文件，排除临时文件
		if strings.HasSuffix(strings.ToLower(filename), ".mp4") && !strings.Contains(filename, ".mp4.") {
			info, err := entry.Info()
			if err != nil {
				continue
			}

			videos = append(videos, VideoInfo{
				Name:      filename,
				Size:      info.Size(),
				CreatedAt: info.ModTime(), // 使用修改时间作为创建时间
			})
		}
	}

	// 根据排序参数进行排序
	switch sortBy {
	case "time":
		// 按创建时间降序排序（最新的在前）
		for i := 0; i < len(videos)-1; i++ {
			for j := i + 1; j < len(videos); j++ {
				if videos[i].CreatedAt.Before(videos[j].CreatedAt) {
					videos[i], videos[j] = videos[j], videos[i]
				}
			}
		}
	case "size":
		// 按文件大小降序排序（大文件在前）
		for i := 0; i < len(videos)-1; i++ {
			for j := i + 1; j < len(videos); j++ {
				if videos[i].Size < videos[j].Size {
					videos[i], videos[j] = videos[j], videos[i]
				}
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(videos)
}

// 处理视频流API请求
func handleVideoStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 从URL路径中提取文件名
	filename := strings.TrimPrefix(r.URL.Path, "/api/video/")
	if filename == "" {
		http.Error(w, "Filename not provided", http.StatusBadRequest)
		return
	}

	// URL解码文件名
	decodedFilename, err := url.QueryUnescape(filename)
	if err != nil {
		http.Error(w, "Invalid filename", http.StatusBadRequest)
		return
	}

	// 获取基础文件名，防止路径遍历攻击
	decodedFilename = filepath.Base(decodedFilename)

	// 获取当前工作目录
	cwd, err := os.Getwd()
	if err != nil {
		http.Error(w, "Failed to get working directory", http.StatusInternalServerError)
		return
	}

	// 构建完整文件路径
	filePath := filepath.Join(cwd, decodedFilename)

	// 检查文件是否存在且为.mp4文件
	if !strings.HasSuffix(strings.ToLower(decodedFilename), ".mp4") {
		http.Error(w, "Invalid file type", http.StatusBadRequest)
		return
	}

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// 设置响应头
	w.Header().Set("Content-Type", "video/mp4")
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// 提供文件服务
	http.ServeFile(w, r, filePath)
}

// 向所有WebSocket客户端广播消息
// 向指定任务ID的客户端发送消息
func sendMessageToTask(taskID, message, msgType string) {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	wsMsg := WSMessage{
		TaskID:  taskID,
		Message: message,
		Type:    msgType,
	}

	sentCount := 0
	for conn, clientInfo := range clients {
		if clientInfo.TaskID == taskID {
			err := conn.WriteJSON(wsMsg)
			if err != nil {
				log.Printf("发送消息错误: %v", err)
				conn.Close()
				delete(clients, conn)
			} else {
				sentCount++
			}
		}
	}

	log.Printf("向任务 %s 发送消息: %s (发送给 %d 个客户端)", taskID, message, sentCount)
}

// 兼容性函数：广播消息给所有客户端（用于系统消息）
func broadcastMessage(message string) {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	wsMsg := WSMessage{
		TaskID:  "",
		Message: message,
		Type:    "system",
	}

	log.Printf("系统广播消息: %s (客户端数量: %d)", message, len(clients))

	for conn := range clients {
		err := conn.WriteJSON(wsMsg)
		if err != nil {
			log.Printf("发送消息错误: %v", err)
			conn.Close()
			delete(clients, conn)
		}
	}
}

// 处理运行yt-dlp的请求
func handleRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "只允许POST请求", http.StatusMethodNotAllowed)
		return
	}

	// 解析请求参数
	var req RunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "无效的JSON参数", http.StatusBadRequest)
		return
	}

	// 验证参数
	if req.Platform == "" || req.URL == "" || req.TaskID == "" {
		http.Error(w, "平台、URL和任务ID参数不能为空", http.StatusBadRequest)
		return
	}

	// 检查任务ID是否已存在
	tasksMu.Lock()
	if _, exists := activeTasks[req.TaskID]; exists {
		tasksMu.Unlock()
		http.Error(w, "任务ID已存在，请使用不同的任务ID", http.StatusConflict)
		return
	}
	tasksMu.Unlock()

	// 向任务相关的客户端发送开始运行的消息
	sendMessageToTask(req.TaskID, fmt.Sprintf("[%s] 开始运行yt-dlp...", time.Now().Format("2006-01-02 15:04:05")), "log")
	sendMessageToTask(req.TaskID, fmt.Sprintf("平台: %s", req.Platform), "log")
	sendMessageToTask(req.TaskID, fmt.Sprintf("URL: %s", req.URL), "log")

	// 在后台运行yt-dlp
	go func() {
		// 获取当前工作目录的绝对路径
		execPath := "./bin/yt-dlp.exe" // 使用相对路径，但确保在正确的工作目录中执行

		// 根据是否启用高级选项构建命令参数
		var args []string
		if req.Config.EnableAdvanced {
			args = buildAdvancedCommandArgs(req.Config, req.URL)
		} else {
			args = buildCommandArgs(req.Platform, req.URL)
		}

		// 显示完整的拼接命令
		fullCommand := execPath
		for _, arg := range args {
			// 如果参数包含空格或特殊字符，用反引号包围
			if strings.Contains(arg, " ") || strings.Contains(arg, "?") || strings.Contains(arg, "&") {
				fullCommand += " `" + arg + "`"
			} else {
				fullCommand += " " + arg
			}
		}
		sendMessageToTask(req.TaskID, fmt.Sprintf("执行命令: %s", fullCommand), "log")

		// 创建命令
		cmd := exec.Command(execPath, args...)
		// 设置工作目录为当前目录
		cmd.Dir = "."
		// 设置环境变量禁用缓冲
		cmd.Env = append(os.Environ(), "PYTHONUNBUFFERED=1")

		// 保存任务命令引用
		tasksMu.Lock()
		activeTasks[req.TaskID] = cmd
		tasksMu.Unlock()

		// 将stderr重定向到stdout，这样所有输出都从一个管道读取
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			sendMessageToTask(req.TaskID, fmt.Sprintf("错误：无法获取标准输出 - %v", err), "error")
			return
		}
		// 将stderr重定向到stdout
		cmd.Stderr = cmd.Stdout

		// 启动命令
		if err := cmd.Start(); err != nil {
			sendMessageToTask(req.TaskID, fmt.Sprintf("错误：无法启动工具 - %v", err), "error")
			return
		}

		// 使用WaitGroup确保goroutine完成
		var wg sync.WaitGroup
		wg.Add(1) // 一个goroutine处理所有输出

		// 读取合并后的输出（stdout + stderr）
		go func() {
			defer wg.Done()
			scanner := bufio.NewScanner(stdout)
			// 设置更大的缓冲区以处理长行
			scanner.Buffer(make([]byte, 64*1024), 64*1024)
			for scanner.Scan() {
				text := convertGBKToUTF8(scanner.Text())

				// 尝试从输出中提取文件名
				if filename := extractFilename(text); filename != "" {
					filesMu.Lock()
					taskFiles[req.TaskID] = filename
					filesMu.Unlock()
					sendMessageToTask(req.TaskID, fmt.Sprintf("检测到下载文件: %s", filename), "progress")
				}

				// 立即发送消息，不等待缓冲
				sendMessageToTask(req.TaskID, text, "log")
			}
			if err := scanner.Err(); err != nil {
				sendMessageToTask(req.TaskID, fmt.Sprintf("错误：读取输出失败 - %v", err), "error")
			}
		}()

		// 等待命令完成
		cmdErr := cmd.Wait()

		// 清理任务引用和文件名
		tasksMu.Lock()
		delete(activeTasks, req.TaskID)
		tasksMu.Unlock()

		filesMu.Lock()
		delete(taskFiles, req.TaskID)
		filesMu.Unlock()

		// 等待所有输出读取完成
		wg.Wait()

		// 发送完成消息
		if cmdErr != nil {
			sendMessageToTask(req.TaskID, fmt.Sprintf("命令执行完成，但有错误：%v", cmdErr), "error")
		} else {
			sendMessageToTask(req.TaskID, fmt.Sprintf("[%s] 命令执行完成", time.Now().Format("2006-01-02 15:04:05")), "complete")
		}
		sendMessageToTask(req.TaskID, "COMMAND_FINISHED", "complete") // 发送完成信号
	}()

	// 返回成功响应
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("命令已启动"))
}

// 处理预览图生成API请求
func handleThumbnail(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 从URL路径中提取文件名
	filename := strings.TrimPrefix(r.URL.Path, "/api/thumbnail/")
	if filename == "" {
		http.Error(w, "Filename not provided", http.StatusBadRequest)
		return
	}

	// URL解码文件名
	decodedFilename, err := url.QueryUnescape(filename)
	if err != nil {
		http.Error(w, "Invalid filename", http.StatusBadRequest)
		return
	}

	// 获取基础文件名，防止路径遍历攻击
	decodedFilename = filepath.Base(decodedFilename)

	// 获取当前工作目录
	cwd, err := os.Getwd()
	if err != nil {
		http.Error(w, "Failed to get working directory", http.StatusInternalServerError)
		return
	}

	// 构建完整文件路径
	filePath := filepath.Join(cwd, decodedFilename)

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// 生成预览图文件名
	thumbnailName := strings.TrimSuffix(decodedFilename, filepath.Ext(decodedFilename)) + "_thumbnail.jpg"
	thumbnailPath := filepath.Join(cwd, "thumbnails", thumbnailName)

	// 创建thumbnails目录
	thumbnailDir := filepath.Join(cwd, "thumbnails")
	if err := os.MkdirAll(thumbnailDir, 0755); err != nil {
		http.Error(w, "Failed to create thumbnail directory", http.StatusInternalServerError)
		return
	}

	// 检查预览图是否已存在
	if _, err := os.Stat(thumbnailPath); err == nil {
		// 预览图已存在，直接返回
		http.ServeFile(w, r, thumbnailPath)
		return
	}

	// 使用FFmpeg生成预览图，保持宽高比
	ffmpegPath := filepath.Join("bin", "ffmpeg.exe")
	cmd := exec.Command(ffmpegPath, "-i", filePath, "-ss", "00:00:05", "-vframes", "1", "-vf", "scale='min(320,iw)':-1", "-y", thumbnailPath)

	if err := cmd.Run(); err != nil {
		log.Printf("FFmpeg error: %v", err)
		http.Error(w, "Failed to generate thumbnail", http.StatusInternalServerError)
		return
	}

	// 返回生成的预览图
	http.ServeFile(w, r, thumbnailPath)
}

// 处理文件删除API请求
func handleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Filename string `json:"filename"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Filename == "" {
		http.Error(w, "Filename not provided", http.StatusBadRequest)
		return
	}

	// 获取基础文件名，防止路径遍历攻击
	filename := filepath.Base(req.Filename)

	// 获取当前工作目录
	cwd, err := os.Getwd()
	if err != nil {
		http.Error(w, "Failed to get working directory", http.StatusInternalServerError)
		return
	}

	// 构建完整文件路径
	filePath := filepath.Join(cwd, filename)

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// 删除文件
	if err := os.Remove(filePath); err != nil {
		http.Error(w, "Failed to delete file", http.StatusInternalServerError)
		return
	}

	// 同时删除对应的预览图
	thumbnailName := strings.TrimSuffix(filename, filepath.Ext(filename)) + "_thumbnail.jpg"
	thumbnailPath := filepath.Join(cwd, "thumbnails", thumbnailName)
	if _, err := os.Stat(thumbnailPath); err == nil {
		os.Remove(thumbnailPath)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// 处理文件重命名API请求
func handleRename(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		OldFilename string `json:"oldFilename"`
		NewFilename string `json:"newFilename"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.OldFilename == "" || req.NewFilename == "" {
		http.Error(w, "Filenames not provided", http.StatusBadRequest)
		return
	}

	// 获取基础文件名，防止路径遍历攻击
	oldFilename := filepath.Base(req.OldFilename)
	newFilename := filepath.Base(req.NewFilename)

	// 确保新文件名有.mp4扩展名
	if !strings.HasSuffix(strings.ToLower(newFilename), ".mp4") {
		newFilename += ".mp4"
	}

	// 获取当前工作目录
	cwd, err := os.Getwd()
	if err != nil {
		http.Error(w, "Failed to get working directory", http.StatusInternalServerError)
		return
	}

	// 构建完整文件路径
	oldPath := filepath.Join(cwd, oldFilename)
	newPath := filepath.Join(cwd, newFilename)

	// 检查原文件是否存在
	if _, err := os.Stat(oldPath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// 检查新文件名是否已存在
	if _, err := os.Stat(newPath); err == nil {
		http.Error(w, "File with new name already exists", http.StatusConflict)
		return
	}

	// 重命名文件
	if err := os.Rename(oldPath, newPath); err != nil {
		http.Error(w, "Failed to rename file", http.StatusInternalServerError)
		return
	}

	// 同时重命名对应的预览图
	oldThumbnailName := strings.TrimSuffix(oldFilename, filepath.Ext(oldFilename)) + "_thumbnail.jpg"
	newThumbnailName := strings.TrimSuffix(newFilename, filepath.Ext(newFilename)) + "_thumbnail.jpg"
	oldThumbnailPath := filepath.Join(cwd, "thumbnails", oldThumbnailName)
	newThumbnailPath := filepath.Join(cwd, "thumbnails", newThumbnailName)
	if _, err := os.Stat(oldThumbnailPath); err == nil {
		os.Rename(oldThumbnailPath, newThumbnailPath)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "newName": newFilename})
}

// 处理批量删除API请求
func handleBatchDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Filenames []string `json:"filenames"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if len(req.Filenames) == 0 {
		http.Error(w, "No filenames provided", http.StatusBadRequest)
		return
	}

	// 获取当前工作目录
	cwd, err := os.Getwd()
	if err != nil {
		http.Error(w, "Failed to get working directory", http.StatusInternalServerError)
		return
	}

	var deletedFiles []string
	var failedFiles []string

	for _, filename := range req.Filenames {
		// 获取基础文件名，防止路径遍历攻击
		cleanFilename := filepath.Base(filename)
		filePath := filepath.Join(cwd, cleanFilename)

		// 检查文件是否存在并删除
		if _, err := os.Stat(filePath); err == nil {
			if err := os.Remove(filePath); err == nil {
				deletedFiles = append(deletedFiles, cleanFilename)

				// 同时删除对应的预览图
				thumbnailName := strings.TrimSuffix(cleanFilename, filepath.Ext(cleanFilename)) + "_thumbnail.jpg"
				thumbnailPath := filepath.Join(cwd, "thumbnails", thumbnailName)
				if _, err := os.Stat(thumbnailPath); err == nil {
					os.Remove(thumbnailPath)
				}
			} else {
				failedFiles = append(failedFiles, cleanFilename)
			}
		} else {
			failedFiles = append(failedFiles, cleanFilename)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":      true,
		"deleted":      deletedFiles,
		"failed":       failedFiles,
		"deletedCount": len(deletedFiles),
		"failedCount":  len(failedFiles),
	})
}

// 处理停止请求
func handleStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "只支持POST请求", http.StatusMethodNotAllowed)
		return
	}

	// 解析请求参数
	var req StopRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "无效的JSON参数", http.StatusBadRequest)
		return
	}

	// 验证任务ID
	if req.TaskID == "" {
		http.Error(w, "任务ID不能为空", http.StatusBadRequest)
		return
	}

	// 查找并停止指定任务
	tasksMu.Lock()
	cmd, exists := activeTasks[req.TaskID]
	if !exists {
		tasksMu.Unlock()
		http.Error(w, "指定的任务不存在或已完成", http.StatusBadRequest)
		return
	}
	tasksMu.Unlock()

	// 获取任务对应的文件名（用于删除未完成的文件）
	filesMu.Lock()
	filename := taskFiles[req.TaskID]
	filesMu.Unlock()

	// 终止进程
	if err := cmd.Process.Kill(); err != nil {
		sendMessageToTask(req.TaskID, fmt.Sprintf("停止命令时出错：%v", err), "error")
		http.Error(w, "停止命令失败", http.StatusInternalServerError)
		return
	}

	sendMessageToTask(req.TaskID, fmt.Sprintf("[%s] 用户手动停止了下载", time.Now().Format("2006-01-02 15:04:05")), "log")

	// 延迟2秒后再检测和删除文件，确保进程完全停止
	sendMessageToTask(req.TaskID, "等待2秒后开始检测临时文件...", "log")
	time.Sleep(2 * time.Second)

	// 删除未完成的文件
	if filename != "" {
		// 获取当前工作目录的绝对路径
		cwd, err := os.Getwd()
		if err != nil {
			sendMessageToTask(req.TaskID, fmt.Sprintf("[调试] 获取工作目录失败: %v", err), "log")
			cwd = "."
		}

		// 如果filename是相对路径，转换为绝对路径
		var absoluteFilename string
		if filepath.IsAbs(filename) {
			absoluteFilename = filename
		} else {
			absoluteFilename = filepath.Join(cwd, filename)
		}

		dir := filepath.Dir(absoluteFilename)
		baseName := filepath.Base(absoluteFilename)

		sendMessageToTask(req.TaskID, fmt.Sprintf("[调试] 当前工作目录: %s", cwd), "log")
		sendMessageToTask(req.TaskID, fmt.Sprintf("[调试] 检测目录: %s", dir), "log")
		sendMessageToTask(req.TaskID, fmt.Sprintf("[调试] 基础文件名: %s", baseName), "log")
		sendMessageToTask(req.TaskID, fmt.Sprintf("[调试] 完整文件路径: %s", absoluteFilename), "log")

		// 新的删除策略：匹配所有以.mp4结尾但后面还有额外后缀的文件
		// 例如：filename.mp4.part, filename.mp4.part1, filename.mp4.temp 等
		deletedCount := 0

		// 由于filepath.Glob在Windows上处理中文字符和特殊字符有问题，
		// 改用ReadDir直接读取目录内容进行匹配
		sendMessageToTask(req.TaskID, fmt.Sprintf("[调试] 开始读取目录内容: %s", dir), "log")

		if entries, err := os.ReadDir(dir); err == nil {

			for _, entry := range entries {
				if entry.IsDir() {
					continue // 跳过目录
				}

				fileName := entry.Name()

				// 检查文件名是否以基础文件名开头
				if strings.HasPrefix(fileName, baseName) {
					// 只删除以.mp4结尾但后面还有额外后缀的文件
					// 例如：xxx.mp4.part, xxx.mp4.temp 等，但不删除 xxx.mp4
					if strings.Contains(fileName, ".mp4.") {
						sendMessageToTask(req.TaskID, fmt.Sprintf("[调试] 文件名匹配基础名称: %s", fileName), "log")
						sendMessageToTask(req.TaskID, fmt.Sprintf("[调试] 文件符合删除条件（.mp4后有额外后缀）: %s", fileName), "log")

						fullPath := filepath.Join(dir, fileName)
						if _, err := os.Stat(fullPath); err == nil {
							sendMessageToTask(req.TaskID, fmt.Sprintf("[调试] 文件存在，尝试删除: %s", fullPath), "log")
							if removeErr := os.Remove(fullPath); removeErr == nil {
								sendMessageToTask(req.TaskID, fmt.Sprintf("已删除文件: %s", fileName), "log")
								deletedCount++
							} else {
								sendMessageToTask(req.TaskID, fmt.Sprintf("删除文件失败 %s: %v", fileName, removeErr), "error")
							}
						} else {
							sendMessageToTask(req.TaskID, fmt.Sprintf("[调试] 文件不存在或无法访问: %s, 错误: %v", fullPath, err), "log")
						}
					}
				}
			}
		} else {
			sendMessageToTask(req.TaskID, fmt.Sprintf("[调试] 读取目录失败: %v", err), "error")
		}

		if deletedCount == 0 {
			sendMessageToTask(req.TaskID, "未找到需要删除的临时文件", "log")
		} else {
			sendMessageToTask(req.TaskID, fmt.Sprintf("共删除了 %d 个文件", deletedCount), "log")
		}
	} else {
		sendMessageToTask(req.TaskID, "[调试] 当前文件名为空，无法进行文件删除", "log")
	}

	// 清理任务相关数据
	tasksMu.Lock()
	delete(activeTasks, req.TaskID)
	tasksMu.Unlock()

	filesMu.Lock()
	delete(taskFiles, req.TaskID)
	filesMu.Unlock()

	sendMessageToTask(req.TaskID, "COMMAND_FINISHED", "complete") // 发送完成信号

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("任务已停止"))
}

// 处理配置保存请求
func handleConfigSave(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var config Config
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// 将配置保存到文件
	configData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		http.Error(w, "Failed to marshal config", http.StatusInternalServerError)
		return
	}

	if err := os.WriteFile("config.json", configData, 0644); err != nil {
		http.Error(w, "Failed to save config", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "配置保存成功"})
}

// 处理配置加载请求
func handleConfigLoad(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 从文件读取配置
	configData, err := os.ReadFile("config.json")
	if err != nil {
		// 如果文件不存在，返回默认配置
		if os.IsNotExist(err) {
			defaultConfig := Config{
				EnableAdvanced:       false,
				DownloadType:         "best",
				SeparateDownload:     "",
				VideoResolution:      "1080p",
				AudioFormat:          "mp3",
				DownloadSubtitle:     false,
				DownloadAutoSubtitle: false,
				SubtitleLanguage:     "zh-CN,en",
				EmbedSubtitle:        false,
				SubtitleOnly:         false,
				PlaylistStart:        1,
				PlaylistEnd:          0,
				EnableThreads:        false,
				ThreadCount:          4,
				EnableRateLimit:      false,
				RateLimit:            "1M",
				ContinueOnError:      false,
				EnableReferer:        false,
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(defaultConfig)
			return
		}
		http.Error(w, "Failed to read config", http.StatusInternalServerError)
		return
	}

	var config Config
	if err := json.Unmarshal(configData, &config); err != nil {
		http.Error(w, "Failed to parse config", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(config)
}

// 从URL中提取主域名作为referer
func extractReferer(urlStr string) string {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "" // 如果解析失败，返回空字符串
	}

	// 返回协议+主机名，例如: https://v.qq.com
	return parsedURL.Scheme + "://" + parsedURL.Host
}

// 根据平台类型构建命令参数
func buildCommandArgs(platform, url string) []string {
	var args []string

	switch platform {
	case "youtube":
		// YouTube
		args = []string{
			"-f", "bv*+ba/b",
			"-S", "res,codec",
			"--merge-output-format", "mp4",
			"--cookies-from-browser", "firefox",
			"--newline", // 强制每行输出后换行
			url,
		}
	case "tiktok":
		// TikTok
		args = []string{
			"-f", "bv*+ba/b",
			"-S", "res:desc,br:desc",
			"--merge-output-format", "mp4",
			"--cookies-from-browser", "firefox",
			"--newline", // 强制每行输出后换行
			url,
		}
	case "bilibili":
		// Bilibili
		args = []string{
			"-f", "bv*+ba",
			"-S", "res:desc,br:desc",
			"--merge-output-format", "mp4",
			"--cookies-from-browser", "firefox",
			"--sub-langs", "all",
			"--newline", // 强制每行输出后换行
			url,
		}
	case "generic1":
		// 其他通用1（最高画质）
		referer := extractReferer(url)
		args = []string{
			"-f", "bv*+ba/b",
			"-S", "res,codec",
			"--merge-output-format", "mp4",
			"--cookies-from-browser", "firefox",
		}
		if referer != "" {
			args = append(args, "--referer", referer)
		}
		args = append(args, "--newline") // 强制每行输出后换行
		args = append(args, url) // 添加URL到最后
	case "generic2":
		// 其他通用2
		referer := extractReferer(url)
		args = []string{
			"--merge-output-format", "mp4",
			"--cookies-from-browser", "firefox",
		}
		if referer != "" {
			args = append(args, "--referer", referer)
		}
		args = append(args, "--newline") // 强制每行输出后换行
		args = append(args, url) // 添加URL到最后
	default:
		// 默认情况
		args = []string{
			"--newline", // 强制每行输出后换行
			url,
		}
	}

	return args
}

// 根据高级配置构建命令参数
func buildAdvancedCommandArgs(config Config, url string) []string {
	var args []string

	// 固定参数
	args = append(args, "--cookies-from-browser", "firefox")
	args = append(args, "--newline") // 强制每行输出后换行

	// 下载设置参数（-f 参数需要放在最前面）
	// 优先处理单独下载选项，如果没有选择单独下载，则使用常规下载类型
	if config.SeparateDownload == "video" {
		// 单独下载视频
		height := extractHeightFromResolution(config.VideoResolution)
		if height != "" {
			args = append([]string{"-f", fmt.Sprintf("bestvideo[height<=%s]", height), "--merge-output-format", "mp4"}, args...)
		} else {
			args = append([]string{"-f", "bestvideo", "--merge-output-format", "mp4"}, args...)
		}
	} else if config.SeparateDownload == "audio" {
		// 单独下载音频
		if config.AudioFormat != "default" && config.AudioFormat != "" {
			args = append([]string{"-f", "bestaudio", "-x", "--audio-format", config.AudioFormat}, args...)
		} else {
			args = append([]string{"-f", "bestaudio"}, args...)
		}
	} else {
		// 常规下载类型
		switch config.DownloadType {
		case "bestQuality":
			args = append([]string{"-f", "bestvideo", "--merge-output-format", "mp4"}, args...)
		case "bestAudio":
			args = append([]string{"-f", "bestaudio"}, args...)
		case "bestMerge":
			args = append([]string{"-f", "bestvideo+bestaudio", "--merge-output-format", "mp4"}, args...)
		}
	}

	// 字幕相关参数
	if config.SubtitleLanguage != "" {
		switch config.SubtitleLanguage {
		case "all":
			args = append(args, "--sub-langs", "all")
		case "zh-CN,en":
			args = append(args, "--sub-langs", "\"zh-CN,en\"")
		case "zh-CN":
			args = append(args, "--sub-langs", "\"zh-CN\"")
		case "en":
			args = append(args, "--sub-langs", "\"en\"")
		}
	}

	if config.DownloadSubtitle {
		args = append(args, "--write-subs")
	}

	if config.DownloadAutoSubtitle {
		args = append(args, "--write-auto-subs")
	}

	if config.EmbedSubtitle {
		args = append(args, "--embed-subs")
	}

	if config.SubtitleOnly {
		args = append(args, "--skip-download")
	}

	// 播放列表相关参数
	switch config.PlaylistMode {
	case "single":
		args = append(args, "--no-playlist")
	case "force":
		args = append(args, "--yes-playlist")
	case "range":
		if config.PlaylistStart > 0 {
			args = append(args, "--playlist-start", fmt.Sprintf("%d", config.PlaylistStart))
		}
		if config.PlaylistEnd > 0 {
			args = append(args, "--playlist-end", fmt.Sprintf("%d", config.PlaylistEnd))
		}
	// 默认情况（"default"或"all"）：不添加任何参数，下载整个播放列表
	}

	// 常见控制参数
	if config.EnableThreads {
		args = append(args, "-N", fmt.Sprintf("%d", config.ThreadCount))
	}

	if config.EnableRateLimit && config.RateLimit != "" {
		args = append(args, "--limit-rate", config.RateLimit)
	}

	if config.ContinueOnError {
		args = append(args, "--ignore-errors")
	}

	if config.EnableReferer {
		referer := extractReferer(url)
		if referer != "" {
			args = append(args, "--referer", referer)
		}
	}

	// 添加URL到最后
	args = append(args, url)

	return args
}

// 从分辨率字符串中提取高度值
func extractHeightFromResolution(resolution string) string {
	switch resolution {
	case "4320p":
		return "4320"
	case "2160p":
		return "2160"
	case "1440p":
		return "1440"
	case "1080p":
		return "1080"
	case "720p":
		return "720"
	default:
		return ""
	}
}

// 获取yt-dlp当前版本
func getCurrentYtDlpVersion() (string, error) {
	execPath := "./bin/yt-dlp.exe"
	cmd := exec.Command(execPath, "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("获取yt-dlp版本失败: %v", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// 获取yt-dlp最新版本信息
func getLatestYtDlpVersion() (string, string, error) {
	resp, err := http.Get("https://api.github.com/repos/yt-dlp/yt-dlp/releases/latest")
	if err != nil {
		return "", "", fmt.Errorf("获取最新版本信息失败: %v", err)
	}
	defer resp.Body.Close()

	var release struct {
		TagName string `json:"tag_name"`
		Assets  []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", "", fmt.Errorf("解析版本信息失败: %v", err)
	}

	// 查找Windows可执行文件
	var downloadURL string
	for _, asset := range release.Assets {
		if asset.Name == "yt-dlp.exe" {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		return "", "", fmt.Errorf("未找到Windows版本下载链接")
	}

	return release.TagName, downloadURL, nil
}

// 处理版本检查请求
func handleVersionCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "只允许GET请求", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// 获取最新版本
	latestVersion, downloadURL, err := getLatestYtDlpVersion()
	if err != nil {
		log.Printf("获取最新版本失败: %v", err)
		http.Error(w, "获取最新版本失败", http.StatusInternalServerError)
		return
	}

	// 获取当前版本
	currentVersion, err := getCurrentYtDlpVersion()
	var versionInfo VersionInfo
	
	if err != nil {
		// 如果获取当前版本失败（通常是文件不存在），提示下载最新版本
		log.Printf("获取当前版本失败: %v", err)
		versionInfo = VersionInfo{
			CurrentVersion: "未安装",
			LatestVersion:  latestVersion,
			HasUpdate:      true,
			DownloadURL:    downloadURL,
		}
	} else {
		// 比较版本
		hasUpdate := currentVersion != latestVersion
		versionInfo = VersionInfo{
			CurrentVersion: currentVersion,
			LatestVersion:  latestVersion,
			HasUpdate:      hasUpdate,
			DownloadURL:    downloadURL,
		}
	}

	json.NewEncoder(w).Encode(versionInfo)
}

// 处理版本更新请求
func handleVersionUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "只允许POST请求", http.StatusMethodNotAllowed)
		return
	}

	// 解析请求参数
	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "无效的JSON参数", http.StatusBadRequest)
		return
	}

	if req.TaskID == "" {
		http.Error(w, "任务ID不能为空", http.StatusBadRequest)
		return
	}

	// 创建可取消的context
	ctx, cancel := context.WithCancel(context.Background())
	
	// 将取消函数存储到全局map中
	updateTasksMu.Lock()
	updateTasks[req.TaskID] = cancel
	updateTasksMu.Unlock()
	
	// 在后台执行更新
	go func() {
		defer func() {
			// 清理任务
			updateTasksMu.Lock()
			delete(updateTasks, req.TaskID)
			updateTasksMu.Unlock()
		}()
		
		// 发送开始更新消息
		sendUpdateProgress(req.TaskID, 0, "开始更新yt-dlp...", "正在准备更新")
		
		// 检查是否已取消
		select {
		case <-ctx.Done():
			sendUpdateProgress(req.TaskID, 0, "更新已取消", "用户取消了更新操作")
			return
		default:
		}

		// 获取最新版本信息
		latestVersion, downloadURL, err := getLatestYtDlpVersion()
		if err != nil {
			sendUpdateProgress(req.TaskID, 0, "更新失败", fmt.Sprintf("获取最新版本信息失败: %v", err))
			return
		}

		sendUpdateProgress(req.TaskID, 5, fmt.Sprintf("发现新版本: %s", latestVersion), "开始下载新版本...")

		// 下载新版本
		resp, err := http.Get(downloadURL)
		if err != nil {
			sendUpdateProgress(req.TaskID, 0, "更新失败", fmt.Sprintf("下载失败: %v", err))
			return
		}
		defer resp.Body.Close()

		// 创建临时文件
		tempFile := "./bin/yt-dlp.exe.new"
		file, err := os.Create(tempFile)
		if err != nil {
			sendUpdateProgress(req.TaskID, 0, "更新失败", fmt.Sprintf("创建临时文件失败: %v", err))
			return
		}

		// 下载文件并显示进度
		totalSize := resp.ContentLength
		var downloaded int64
		buffer := make([]byte, 32*1024) // 32KB buffer

		for {
			// 检查是否已取消
			select {
			case <-ctx.Done():
				file.Close()
				os.Remove(tempFile)
				sendUpdateProgress(req.TaskID, 0, "更新已取消", "用户取消了更新操作")
				return
			default:
			}
			
			n, err := resp.Body.Read(buffer)
			if n > 0 {
				file.Write(buffer[:n])
				downloaded += int64(n)
				if totalSize > 0 {
					progress := int(float64(downloaded) / float64(totalSize) * 100)
					sendUpdateProgress(req.TaskID, progress, fmt.Sprintf("正在下载 yt-dlp... %.1f%%", float64(progress)), fmt.Sprintf("已下载 %.1f MB / %.1f MB", float64(downloaded)/(1024*1024), float64(totalSize)/(1024*1024)))
				}
			}
			if err != nil {
				break
			}
		}
		file.Close()

		if downloaded != totalSize && totalSize > 0 {
			sendUpdateProgress(req.TaskID, 0, "下载失败", "下载不完整，请重试")
			os.Remove(tempFile)
			return
		}

		sendUpdateProgress(req.TaskID, 95, "下载完成，准备替换文件...", "正在安装新版本")
		
		// 检查是否已取消
		select {
		case <-ctx.Done():
			os.Remove(tempFile)
			sendUpdateProgress(req.TaskID, 0, "更新已取消", "用户取消了更新操作")
			return
		default:
		}

		// 停止所有使用yt-dlp的任务
		tasksMu.Lock()
		for taskID, cmd := range activeTasks {
			if cmd != nil {
				sendMessageToTask(taskID, "检测到yt-dlp更新，正在停止当前任务...", "log")
				cmd.Process.Kill()
			}
		}
		// 清空活跃任务
		for taskID := range activeTasks {
			delete(activeTasks, taskID)
		}
		tasksMu.Unlock()

		// 等待进程完全停止
		time.Sleep(2 * time.Second)

		// 备份当前版本
		backupFile := "./bin/yt-dlp.exe.backup"
		originalFile := "./bin/yt-dlp.exe"

		if _, err := os.Stat(originalFile); err == nil {
			if err := os.Rename(originalFile, backupFile); err != nil {
				sendUpdateProgress(req.TaskID, 0, "更新失败", fmt.Sprintf("备份原文件失败: %v", err))
				os.Remove(tempFile)
				return
			}
		}

		// 替换文件
		if err := os.Rename(tempFile, originalFile); err != nil {
			sendUpdateProgress(req.TaskID, 0, "更新失败", fmt.Sprintf("替换文件失败: %v", err))
			// 恢复备份
			if _, backupErr := os.Stat(backupFile); backupErr == nil {
				os.Rename(backupFile, originalFile)
			}
			return
		}

		// 删除备份文件
		os.Remove(backupFile)

		// 验证新版本
		newVersion, err := getCurrentYtDlpVersion()
		if err != nil {
			sendUpdateProgress(req.TaskID, 0, "更新失败", fmt.Sprintf("验证新版本失败: %v", err))
			return
		}

		sendUpdateProgress(req.TaskID, 100, "更新完成！", fmt.Sprintf("当前版本: %s", newVersion))
		sendUpdateProgress(req.TaskID, 100, "UPDATE_COMPLETE", "更新完成")
	}()

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("更新已开始"))
}

// 处理取消更新请求
func handleVersionCancel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "只允许POST请求", http.StatusMethodNotAllowed)
		return
	}

	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "无效的JSON参数", http.StatusBadRequest)
		return
	}

	if req.TaskID == "" {
		http.Error(w, "任务ID不能为空", http.StatusBadRequest)
		return
	}

	// 查找并取消对应的更新任务
	updateTasksMu.Lock()
	cancel, exists := updateTasks[req.TaskID]
	updateTasksMu.Unlock()

	if exists {
		// 调用取消函数
		cancel()
		log.Printf("取消更新任务: %s", req.TaskID)
	} else {
		log.Printf("未找到要取消的更新任务: %s", req.TaskID)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "cancelled"})
}

// 发送更新进度消息
func sendUpdateProgress(taskID string, progress int, message, status string) {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	progressMsg := map[string]interface{}{
		"type":     "update_progress",
		"taskID":   taskID,
		"progress": progress,
		"message":  message,
		"status":   status,
	}

	sentCount := 0
	for conn, clientInfo := range clients {
		if clientInfo.TaskID == taskID {
			err := conn.WriteJSON(progressMsg)
			if err != nil {
				log.Printf("发送更新进度消息错误: %v", err)
				conn.Close()
				delete(clients, conn)
			} else {
				sentCount++
			}
		}
	}

	log.Printf("向任务 %s 发送更新进度: %s (发送给 %d 个客户端)", taskID, message, sentCount)
}

// 停止所有任务
func stopAllTasks() {
	// 这里可以添加停止所有正在运行的yt-dlp进程的逻辑
	// 目前简单等待一下确保进程释放
	time.Sleep(2 * time.Second)
}

// 下载yt-dlp文件
func downloadYtDlp(downloadURL, taskID string) (string, error) {
	resp, err := http.Get(downloadURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("下载失败，状态码: %d", resp.StatusCode)
	}

	// 创建临时文件
	tempFile, err := os.CreateTemp("", "yt-dlp-*.exe")
	if err != nil {
		return "", err
	}
	defer tempFile.Close()

	// 获取文件大小用于进度计算
	contentLength := resp.ContentLength
	var downloaded int64

	// 创建进度读取器
	buffer := make([]byte, 32*1024) // 32KB buffer
	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			_, writeErr := tempFile.Write(buffer[:n])
			if writeErr != nil {
				return "", writeErr
			}
			downloaded += int64(n)

			// 计算并发送进度
			if contentLength > 0 {
				progress := int(30 + (downloaded*50)/contentLength) // 30-80%的进度用于下载
				sendUpdateProgress(taskID, progress, "下载新版本...", fmt.Sprintf("已下载 %.1f MB / %.1f MB", float64(downloaded)/(1024*1024), float64(contentLength)/(1024*1024)))
			}
		}
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return "", err
		}
	}

	return tempFile.Name(), nil
}

// 复制文件
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// 读取版本文件
func readVersionFromFile() string {
	content, err := os.ReadFile("version.txt")
	if err != nil {
		log.Printf("读取版本文件失败: %v", err)
		return Version // 返回默认版本
	}
	return strings.TrimSpace(string(content))
}

// 处理应用信息请求
func handleAppInfo(w http.ResponseWriter, r *http.Request) {
	log.Printf("收到应用信息请求: %s %s", r.Method, r.URL.Path)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 从文件读取版本号
	version := readVersionFromFile()
	
	// 构建应用标题
	title := fmt.Sprintf("X-KT 视频下载器 %s", version)

	appInfo := AppInfo{
		Name:    "X-KT 视频下载器",
		Version: version,
		Title:   title,
	}

	json.NewEncoder(w).Encode(appInfo)
}
