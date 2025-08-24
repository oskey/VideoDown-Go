package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"html/template"
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

// 请求结构体
type RunRequest struct {
	Platform string `json:"platform"`
	URL      string `json:"url"`
	TaskID   string `json:"taskID"` // 添加任务ID字段
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
	clients     = make(map[*websocket.Conn]*ClientInfo) // 存储所有WebSocket连接及其信息
	clientsMu   sync.Mutex                              // 保护clients的互斥锁
	activeTasks = make(map[string]*exec.Cmd)            // 存储活跃的下载任务
	tasksMu     sync.Mutex                              // 保护activeTasks的互斥锁
	taskFiles   = make(map[string]string)               // 存储任务对应的文件名
	filesMu     sync.Mutex                              // 保护taskFiles的互斥锁
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

		// 根据平台类型构建命令参数
		var args []string
		switch req.Platform {
		case "youtube":
			// YouTube
			args = []string{
				"-f", "bv*+ba/b",
				"-S", "res,codec",
				"--merge-output-format", "mp4",
				"--cookies-from-browser", "firefox",
				"--newline", // 强制每行输出后换行
				req.URL,
			}
		case "tiktok":
			// TikTok
			args = []string{
				"-f", "bv*+ba/b",
				"-S", "res:desc,br:desc",
				"--merge-output-format", "mp4",
				"--cookies-from-browser", "firefox",
				"--newline", // 强制每行输出后换行
				req.URL,
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
				req.URL,
			}
		case "generic1":
			// 其他通用1（最高画质）
			args = []string{
				"-f", "bv*+ba/b",
				"-S", "res,codec",
				"--merge-output-format", "mp4",
				"--cookies-from-browser", "firefox",
				"--newline", // 强制每行输出后换行
				req.URL,
			}
		case "generic2":
			// 其他通用2
			args = []string{
				"--merge-output-format", "mp4",
				"--cookies-from-browser", "firefox",
				"--newline", // 强制每行输出后换行
				req.URL,
			}
		default:
			// 默认情况
			args = []string{
				"--newline", // 强制每行输出后换行
				req.URL,
			}
		}

		// 显示完整的拼接命令
		fullCommand := execPath
		for _, arg := range args {
			fullCommand += " " + arg
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
		"success":     true,
		"deleted":     deletedFiles,
		"failed":      failedFiles,
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
