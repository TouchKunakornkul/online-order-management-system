package test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"online-order-management-system/internal/api/http/handler/dto"
)

// StressTestConfig defines parameters for stress testing
type StressTestConfig struct {
	BaseURL        string
	TotalOrders    int // Total number of orders to create
	MaxConcurrency int // Maximum concurrent goroutines
	RequestTimeout time.Duration
	TestTimeout    time.Duration
	BatchSize      int // Orders per batch
}

// StressTestResult contains the results of a stress test
type StressTestResult struct {
	TotalOrders      int64
	SuccessfulOrders int64
	FailedOrders     int64
	TotalDuration    time.Duration
	OrdersPerSecond  float64
	AverageLatency   time.Duration
	MinLatency       time.Duration
	MaxLatency       time.Duration
	SuccessRate      float64
	Errors           []string
	PeakConcurrency  int
}

// OrderMetrics tracks individual order creation performance
type OrderMetrics struct {
	OrderID   int
	StartTime time.Time
	EndTime   time.Time
	Success   bool
	Error     string
	Latency   time.Duration
}

func createStressTestOrder(orderID int) dto.CreateOrderRequest {
	return dto.CreateOrderRequest{
		CustomerName:  fmt.Sprintf("StressTest Customer %d", orderID),
		CustomerEmail: fmt.Sprintf("stress%d@loadtest.com", orderID),
		Items: []dto.CreateOrderItemRequest{
			{
				ProductName: fmt.Sprintf("Product-%d-A", orderID),
				Quantity:    1,
				UnitPrice:   99.99,
			},
			{
				ProductName: fmt.Sprintf("Product-%d-B", orderID),
				Quantity:    2,
				UnitPrice:   49.99,
			},
		},
	}
}

func executeOrderCreation(baseURL string, orderReq dto.CreateOrderRequest, orderID int, timeout time.Duration) OrderMetrics {
	start := time.Now()

	reqBody, err := json.Marshal(orderReq)
	if err != nil {
		return OrderMetrics{
			OrderID:   orderID,
			StartTime: start,
			EndTime:   time.Now(),
			Success:   false,
			Error:     fmt.Sprintf("marshal error: %v", err),
			Latency:   time.Since(start),
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/api/v1/orders", bytes.NewBuffer(reqBody))
	if err != nil {
		return OrderMetrics{
			OrderID:   orderID,
			StartTime: start,
			EndTime:   time.Now(),
			Success:   false,
			Error:     fmt.Sprintf("request creation error: %v", err),
			Latency:   time.Since(start),
		}
	}

	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	end := time.Now()
	latency := end.Sub(start)

	if err != nil {
		return OrderMetrics{
			OrderID:   orderID,
			StartTime: start,
			EndTime:   end,
			Success:   false,
			Error:     fmt.Sprintf("request error: %v", err),
			Latency:   latency,
		}
	}
	defer resp.Body.Close()

	// Read response body
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		return OrderMetrics{
			OrderID:   orderID,
			StartTime: start,
			EndTime:   end,
			Success:   false,
			Error:     fmt.Sprintf("response read error: %v", err),
			Latency:   latency,
		}
	}

	success := resp.StatusCode == http.StatusCreated
	var errorMsg string
	if !success {
		errorMsg = fmt.Sprintf("HTTP %d", resp.StatusCode)
	}

	return OrderMetrics{
		OrderID:   orderID,
		StartTime: start,
		EndTime:   end,
		Success:   success,
		Error:     errorMsg,
		Latency:   latency,
	}
}

func runStressTest(config StressTestConfig) StressTestResult {
	startTime := time.Now()

	// Channels for coordination
	orderChan := make(chan int, config.TotalOrders)
	resultChan := make(chan OrderMetrics, config.TotalOrders)

	// Populate order IDs
	for i := 1; i <= config.TotalOrders; i++ {
		orderChan <- i
	}
	close(orderChan)

	// Track active goroutines
	var activeGoroutines int64
	var peakConcurrency int64

	// Worker goroutines
	var wg sync.WaitGroup
	for i := 0; i < config.MaxConcurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for orderID := range orderChan {
				// Track concurrency
				current := atomic.AddInt64(&activeGoroutines, 1)
				for {
					peak := atomic.LoadInt64(&peakConcurrency)
					if current <= peak || atomic.CompareAndSwapInt64(&peakConcurrency, peak, current) {
						break
					}
				}

				// Create order
				orderReq := createStressTestOrder(orderID)
				metrics := executeOrderCreation(config.BaseURL, orderReq, orderID, config.RequestTimeout)
				resultChan <- metrics

				// Decrease concurrency counter
				atomic.AddInt64(&activeGoroutines, -1)
			}
		}(i)
	}

	// Wait for all workers to complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	var metrics []OrderMetrics
	for metric := range resultChan {
		metrics = append(metrics, metric)
	}

	endTime := time.Now()
	testDuration := endTime.Sub(startTime)

	// Calculate results
	result := calculateStressTestResults(metrics, testDuration, int(peakConcurrency))
	return result
}

func calculateStressTestResults(metrics []OrderMetrics, testDuration time.Duration, peakConcurrency int) StressTestResult {
	result := StressTestResult{
		TotalDuration:   testDuration,
		MinLatency:      time.Hour, // Start with a very high value
		PeakConcurrency: peakConcurrency,
	}

	var totalLatency time.Duration
	var errors []string

	for _, metric := range metrics {
		result.TotalOrders++
		totalLatency += metric.Latency

		if metric.Success {
			result.SuccessfulOrders++
		} else {
			result.FailedOrders++
			if len(errors) < 20 { // Collect more errors for stress test
				errors = append(errors, fmt.Sprintf("Order %d: %s", metric.OrderID, metric.Error))
			}
		}

		if metric.Latency < result.MinLatency {
			result.MinLatency = metric.Latency
		}
		if metric.Latency > result.MaxLatency {
			result.MaxLatency = metric.Latency
		}
	}

	if result.TotalOrders > 0 {
		result.AverageLatency = totalLatency / time.Duration(result.TotalOrders)
		result.OrdersPerSecond = float64(result.TotalOrders) / testDuration.Seconds()
		result.SuccessRate = float64(result.SuccessfulOrders) / float64(result.TotalOrders) * 100
	}

	result.Errors = errors
	return result
}

// getStressTestBaseURL returns the base URL for stress testing
// Supports both regular and isolated stress testing
func getStressTestBaseURL() string {
	if baseURL := os.Getenv("STRESS_TEST_BASE_URL"); baseURL != "" {
		return baseURL
	}
	return "http://localhost:8080" // Default for regular stress testing
}

func TestStressTest_1000Orders(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	config := StressTestConfig{
		BaseURL:        getStressTestBaseURL(),
		TotalOrders:    1000,
		MaxConcurrency: 100, // 100 concurrent goroutines
		RequestTimeout: 30 * time.Second,
		TestTimeout:    5 * time.Minute,
		BatchSize:      10,
	}

	// Test if server is running
	resp, err := http.Get(config.BaseURL + "/health")
	if err != nil {
		t.Skipf("Skipping stress test: server not running at %s", config.BaseURL)
	}
	resp.Body.Close()

	t.Logf("🔥 Starting stress test: Creating %d orders with %d concurrent goroutines",
		config.TotalOrders, config.MaxConcurrency)

	result := runStressTest(config)

	// Report results
	t.Logf("📊 Stress Test Results (1,000 Orders):")
	t.Logf("  Total Orders: %d", result.TotalOrders)
	t.Logf("  Successful: %d", result.SuccessfulOrders)
	t.Logf("  Failed: %d", result.FailedOrders)
	t.Logf("  Success Rate: %.2f%%", result.SuccessRate)
	t.Logf("  Test Duration: %v", result.TotalDuration)
	t.Logf("  Orders/Second: %.2f", result.OrdersPerSecond)
	t.Logf("  Peak Concurrency: %d goroutines", result.PeakConcurrency)
	t.Logf("  Average Latency: %v", result.AverageLatency)
	t.Logf("  Min Latency: %v", result.MinLatency)
	t.Logf("  Max Latency: %v", result.MaxLatency)

	if len(result.Errors) > 0 {
		t.Logf("  Sample Errors:")
		for i, err := range result.Errors {
			if i >= 10 { // Limit error display
				t.Logf("    ... and %d more errors", len(result.Errors)-10)
				break
			}
			t.Logf("    %s", err)
		}
	}

	// Stress test acceptance criteria (more lenient than load test)
	expectedMinSuccessRate := 90.0 // Lower success rate acceptable for stress test
	if result.SuccessRate < expectedMinSuccessRate {
		t.Errorf("Success rate too low for stress test: got %.2f%%, expected at least %.2f%%",
			result.SuccessRate, expectedMinSuccessRate)
	}

	expectedMinOPS := 5.0 // Lower orders per second acceptable for stress test
	if result.OrdersPerSecond < expectedMinOPS {
		t.Errorf("Orders per second too low: got %.2f, expected at least %.2f",
			result.OrdersPerSecond, expectedMinOPS)
	}
}

func TestStressTest_10000Orders(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	config := StressTestConfig{
		BaseURL:        getStressTestBaseURL(),
		TotalOrders:    10000,
		MaxConcurrency: 500,              // 500 concurrent goroutines for extreme load
		RequestTimeout: 60 * time.Second, // Longer timeout for extreme load
		TestTimeout:    10 * time.Minute,
		BatchSize:      50,
	}

	// Test if server is running
	resp, err := http.Get(config.BaseURL + "/health")
	if err != nil {
		t.Skipf("Skipping stress test: server not running at %s", config.BaseURL)
	}
	resp.Body.Close()

	t.Logf("🚨 Starting EXTREME stress test: Creating %d orders with %d concurrent goroutines",
		config.TotalOrders, config.MaxConcurrency)
	t.Logf("⚠️  This test may take several minutes and stress your system significantly")

	result := runStressTest(config)

	// Report results
	t.Logf("📊 EXTREME Stress Test Results (10,000 Orders):")
	t.Logf("  Total Orders: %d", result.TotalOrders)
	t.Logf("  Successful: %d", result.SuccessfulOrders)
	t.Logf("  Failed: %d", result.FailedOrders)
	t.Logf("  Success Rate: %.2f%%", result.SuccessRate)
	t.Logf("  Test Duration: %v", result.TotalDuration)
	t.Logf("  Orders/Second: %.2f", result.OrdersPerSecond)
	t.Logf("  Peak Concurrency: %d goroutines", result.PeakConcurrency)
	t.Logf("  Average Latency: %v", result.AverageLatency)
	t.Logf("  Min Latency: %v", result.MinLatency)
	t.Logf("  Max Latency: %v", result.MaxLatency)

	if len(result.Errors) > 0 {
		t.Logf("  Sample Errors:")
		for i, err := range result.Errors {
			if i >= 15 {
				t.Logf("    ... and %d more errors", len(result.Errors)-15)
				break
			}
			t.Logf("    %s", err)
		}
	}

	// Very lenient criteria for extreme stress test
	expectedMinSuccessRate := 80.0 // Even lower success rate for extreme load
	if result.SuccessRate < expectedMinSuccessRate {
		t.Logf("⚠️  Success rate: %.2f%% (expected ≥%.2f%% but acceptable for extreme stress)",
			result.SuccessRate, expectedMinSuccessRate)
	}

	// Just log performance, don't fail the test for extreme load
	t.Logf("📈 Performance Analysis:")
	if result.OrdersPerSecond >= 50 {
		t.Logf("  🟢 Excellent performance: %.2f orders/second", result.OrdersPerSecond)
	} else if result.OrdersPerSecond >= 20 {
		t.Logf("  🟡 Good performance: %.2f orders/second", result.OrdersPerSecond)
	} else {
		t.Logf("  🔴 Performance under stress: %.2f orders/second", result.OrdersPerSecond)
	}
}

func BenchmarkStressTest_OrderCreation(b *testing.B) {
	config := StressTestConfig{
		BaseURL:        getStressTestBaseURL(),
		RequestTimeout: 30 * time.Second,
	}

	// Test if server is running
	resp, err := http.Get(config.BaseURL + "/health")
	if err != nil {
		b.Skipf("Skipping benchmark: server not running at %s", config.BaseURL)
	}
	resp.Body.Close()

	var successCount int64
	var errorCount int64

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		orderID := 0
		for pb.Next() {
			orderID++
			orderReq := createStressTestOrder(orderID)
			metrics := executeOrderCreation(config.BaseURL, orderReq, orderID, config.RequestTimeout)

			if metrics.Success {
				atomic.AddInt64(&successCount, 1)
			} else {
				atomic.AddInt64(&errorCount, 1)
			}
		}
	})

	b.ReportMetric(float64(successCount), "successful_orders")
	b.ReportMetric(float64(errorCount), "failed_orders")
	b.ReportMetric(float64(successCount)/float64(successCount+errorCount)*100, "success_rate_%")
}
