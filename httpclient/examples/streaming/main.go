package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-lib/httpclient"
	"github.com/fsvxavier/nexs-lib/httpclient/interfaces"
)

// FileDownloadHandler handles streaming download to a file
type FileDownloadHandler struct {
	file       *os.File
	totalBytes int64
	startTime  time.Time
	mu         sync.Mutex
}

func NewFileDownloadHandler(filename string) (*FileDownloadHandler, error) {
	file, err := os.Create(filename)
	if err != nil {
		return nil, err
	}

	return &FileDownloadHandler{
		file:      file,
		startTime: time.Now(),
	}, nil
}

func (h *FileDownloadHandler) OnData(data []byte) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	n, err := h.file.Write(data)
	if err != nil {
		return err
	}

	h.totalBytes += int64(n)

	// Progress update every 100KB
	if h.totalBytes%(100*1024) == 0 {
		elapsed := time.Since(h.startTime)
		speed := float64(h.totalBytes) / elapsed.Seconds() / 1024 // KB/s
		fmt.Printf("📥 Downloaded: %d bytes (%.2f KB/s)\n", h.totalBytes, speed)
	}

	return nil
}

func (h *FileDownloadHandler) OnError(err error) {
	fmt.Printf("❌ Download error: %v\n", err)
	h.file.Close()
}

func (h *FileDownloadHandler) OnComplete() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.file.Close()
	elapsed := time.Since(h.startTime)
	avgSpeed := float64(h.totalBytes) / elapsed.Seconds() / 1024 // KB/s
	fmt.Printf("✅ Download completed: %d bytes in %v (avg %.2f KB/s)\n",
		h.totalBytes, elapsed, avgSpeed)
}

func (h *FileDownloadHandler) GetTotalBytes() int64 {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.totalBytes
}

// JSONStreamHandler handles streaming JSON responses
type JSONStreamHandler struct {
	chunks     [][]byte
	totalBytes int64
	objects    int
	mu         sync.Mutex
}

func NewJSONStreamHandler() *JSONStreamHandler {
	return &JSONStreamHandler{
		chunks: make([][]byte, 0),
	}
}

func (h *JSONStreamHandler) OnData(data []byte) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Copy data to avoid memory issues
	chunk := make([]byte, len(data))
	copy(chunk, data)
	h.chunks = append(h.chunks, chunk)
	h.totalBytes += int64(len(data))

	// Count JSON objects (simple count of '{')
	for _, b := range data {
		if b == '{' {
			h.objects++
		}
	}

	fmt.Printf("📄 JSON chunk received: %d bytes (%d total objects so far)\n",
		len(data), h.objects)
	return nil
}

func (h *JSONStreamHandler) OnError(err error) {
	fmt.Printf("❌ JSON stream error: %v\n", err)
}

func (h *JSONStreamHandler) OnComplete() {
	h.mu.Lock()
	defer h.mu.Unlock()

	fmt.Printf("✅ JSON stream completed: %d bytes, %d objects processed\n",
		h.totalBytes, h.objects)
}

func (h *JSONStreamHandler) GetData() []byte {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Combine all chunks
	totalSize := int(h.totalBytes)
	result := make([]byte, 0, totalSize)
	for _, chunk := range h.chunks {
		result = append(result, chunk...)
	}
	return result
}

// ProgressHandler shows download progress with a progress bar
type ProgressHandler struct {
	totalBytes   int64
	expectedSize int64
	startTime    time.Time
	lastUpdate   time.Time
	mu           sync.Mutex
}

func NewProgressHandler(expectedSize int64) *ProgressHandler {
	return &ProgressHandler{
		expectedSize: expectedSize,
		startTime:    time.Now(),
		lastUpdate:   time.Now(),
	}
}

func (h *ProgressHandler) OnData(data []byte) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.totalBytes += int64(len(data))
	now := time.Now()

	// Update progress every 500ms
	if now.Sub(h.lastUpdate) > 500*time.Millisecond {
		h.printProgress()
		h.lastUpdate = now
	}

	return nil
}

func (h *ProgressHandler) OnError(err error) {
	fmt.Printf("\n❌ Progress error: %v\n", err)
}

func (h *ProgressHandler) OnComplete() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.printProgress()
	elapsed := time.Since(h.startTime)
	avgSpeed := float64(h.totalBytes) / elapsed.Seconds() / 1024 // KB/s
	fmt.Printf("\n✅ Transfer completed in %v (avg %.2f KB/s)\n", elapsed, avgSpeed)
}

func (h *ProgressHandler) printProgress() {
	percent := float64(h.totalBytes) / float64(h.expectedSize) * 100
	if h.expectedSize == 0 {
		percent = 0
	}

	// Create progress bar
	barWidth := 50
	filled := int(percent / 100 * float64(barWidth))
	bar := ""
	for i := 0; i < barWidth; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}

	elapsed := time.Since(h.startTime)
	speed := float64(h.totalBytes) / elapsed.Seconds() / 1024 // KB/s

	fmt.Printf("\r🔄 [%s] %.1f%% (%d/%d bytes) %.2f KB/s",
		bar, percent, h.totalBytes, h.expectedSize, speed)
}

func main() {
	fmt.Printf("🌊 Streaming Example\n")
	fmt.Printf("====================\n\n")

	// Create HTTP client
	client, err := httpclient.New(interfaces.ProviderNetHTTP, "https://httpbin.org")
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Example 1: Stream JSON data
	fmt.Println("1️⃣ Streaming JSON data from /json endpoint...")
	jsonHandler := NewJSONStreamHandler()

	err = client.Stream(ctx, "GET", "/json", jsonHandler)
	if err != nil {
		log.Printf("JSON streaming failed: %v", err)
	} else {
		// Show first 200 characters of the received data
		data := jsonHandler.GetData()
		preview := string(data)
		if len(preview) > 200 {
			preview = preview[:200] + "..."
		}
		fmt.Printf("📄 Received JSON data preview: %s\n", preview)
	}
	fmt.Println()

	// Example 2: Stream large response with progress
	fmt.Println("2️⃣ Streaming large response with progress bar...")
	// Simulate expected size (httpbin.org/bytes/10240 returns 10KB)
	progressHandler := NewProgressHandler(10240)

	err = client.Stream(ctx, "GET", "/bytes/10240", progressHandler)
	if err != nil {
		log.Printf("Progress streaming failed: %v", err)
	}
	fmt.Println()

	// Example 3: Download file with streaming
	fmt.Println("3️⃣ Downloading file with streaming...")
	filename := "downloaded_data.bin"
	fileHandler, err := NewFileDownloadHandler(filename)
	if err != nil {
		log.Printf("Failed to create file handler: %v", err)
	} else {
		err = client.Stream(ctx, "GET", "/bytes/50000", fileHandler) // Download 50KB
		if err != nil {
			log.Printf("File download failed: %v", err)
		} else {
			// Check file size
			if stat, err := os.Stat(filename); err == nil {
				fmt.Printf("📁 File saved: %s (%d bytes)\n", filename, stat.Size())
			}
		}
	}
	fmt.Println()

	// Example 4: Multiple concurrent streams
	fmt.Println("4️⃣ Multiple concurrent streaming downloads...")
	var wg sync.WaitGroup

	for i := 1; i <= 3; i++ {
		wg.Add(1)
		go func(streamID int) {
			defer wg.Done()

			handler := NewJSONStreamHandler()
			endpoint := fmt.Sprintf("/delay/%d", streamID) // Different delays

			fmt.Printf("🚀 Starting stream %d...\n", streamID)
			err := client.Stream(ctx, "GET", endpoint, handler)
			if err != nil {
				log.Printf("Stream %d failed: %v", streamID, err)
			} else {
				fmt.Printf("✅ Stream %d completed (%d bytes)\n", streamID, handler.totalBytes)
			}
		}(i)
	}

	wg.Wait()
	fmt.Println("All concurrent streams completed!")
	fmt.Println()

	// Example 5: Streaming with timeout
	fmt.Println("5️⃣ Streaming with timeout...")
	timeoutCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	slowHandler := NewJSONStreamHandler()
	err = client.Stream(timeoutCtx, "GET", "/delay/5", slowHandler) // 5s delay, 2s timeout
	if err != nil {
		fmt.Printf("⏰ Streaming timed out as expected: %v\n", err)
	}

	// Cleanup
	fmt.Println("\n🧹 Cleaning up...")
	if err := os.Remove(filename); err != nil {
		log.Printf("Failed to remove file: %v", err)
	} else {
		fmt.Printf("🗑️  Removed file: %s\n", filename)
	}

	fmt.Println("\n🎉 Streaming example completed!")
	fmt.Println("\n💡 Key Features Demonstrated:")
	fmt.Println("  • JSON data streaming")
	fmt.Println("  • Progress bar for large downloads")
	fmt.Println("  • File download with streaming")
	fmt.Println("  • Concurrent streaming operations")
	fmt.Println("  • Streaming with timeout handling")
	fmt.Println("  • Memory-efficient large data processing")
}
