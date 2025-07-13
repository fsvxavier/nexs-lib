package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-lib/v2/observability/tracer"
	"github.com/fsvxavier/nexs-lib/v2/observability/tracer/providers/newrelic"
)

// Message represents a message in the queue
type Message struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Payload   map[string]interface{} `json:"payload"`
	Metadata  map[string]string      `json:"metadata"`
	Timestamp time.Time              `json:"timestamp"`
	Retry     int                    `json:"retry"`
}

// GRPCService simulates a gRPC service
type GRPCService struct {
	tracer tracer.Tracer
	queue  *MessageQueue
}

// MessageQueue simulates a message queue system
type MessageQueue struct {
	tracer   tracer.Tracer
	messages chan Message
	workers  []*Worker
	mu       sync.RWMutex
	stats    QueueStats
}

type QueueStats struct {
	Published  int64
	Consumed   int64
	Failed     int64
	Retried    int64
	Processing int64
}

// Worker processes messages from the queue
type Worker struct {
	ID     int
	tracer tracer.Tracer
	queue  *MessageQueue
	stop   chan bool
}

// ProcessOrderRequest simulates a gRPC method
func (s *GRPCService) ProcessOrderRequest(ctx context.Context, orderID string, customerData map[string]interface{}) error {
	// Start server span for gRPC request
	ctx, span := s.tracer.StartSpan(ctx, "grpc.ProcessOrderRequest",
		tracer.WithSpanKind(tracer.SpanKindServer),
		tracer.WithSpanAttributes(map[string]interface{}{
			"grpc.service": "OrderService",
			"grpc.method":  "ProcessOrderRequest",
			"order.id":     orderID,
			"customer.id":  customerData["customer_id"],
		}),
	)
	defer span.End()

	span.AddEvent("request.started", map[string]interface{}{
		"timestamp": time.Now().Unix(),
		"order_id":  orderID,
	})

	// Validate request
	if err := s.validateOrderRequest(ctx, orderID, customerData); err != nil {
		span.SetStatus(tracer.StatusCodeError, fmt.Sprintf("Validation failed: %v", err))
		span.RecordError(err, map[string]interface{}{
			"error.type": "validation_error",
		})
		return err
	}

	// Transform data
	transformedData, err := s.transformOrderData(ctx, customerData)
	if err != nil {
		span.SetStatus(tracer.StatusCodeError, fmt.Sprintf("Transformation failed: %v", err))
		span.RecordError(err, map[string]interface{}{
			"error.type": "transformation_error",
		})
		return err
	}

	// Publish to message queue for async processing
	message := Message{
		ID:   fmt.Sprintf("order_%s_%d", orderID, time.Now().Unix()),
		Type: "order_processing",
		Payload: map[string]interface{}{
			"order_id":         orderID,
			"customer_data":    transformedData,
			"processing_stage": "initial",
		},
		Metadata: map[string]string{
			"source":     "grpc_api",
			"trace_id":   span.Context().TraceID,
			"span_id":    span.Context().SpanID,
			"priority":   "high",
			"created_by": "order_service",
		},
		Timestamp: time.Now(),
	}

	if err := s.queue.Publish(ctx, message); err != nil {
		span.SetStatus(tracer.StatusCodeError, fmt.Sprintf("Failed to publish message: %v", err))
		span.RecordError(err, map[string]interface{}{
			"error.type": "queue_publish_error",
		})
		return err
	}

	span.SetAttribute("message.published", true)
	span.SetAttribute("message.id", message.ID)
	span.AddEvent("message.published", map[string]interface{}{
		"message_id": message.ID,
		"queue":      "order_processing",
		"timestamp":  time.Now().Unix(),
	})

	span.SetStatus(tracer.StatusCodeOk, "Order request processed successfully")
	return nil
}

func (s *GRPCService) validateOrderRequest(ctx context.Context, orderID string, data map[string]interface{}) error {
	_, span := s.tracer.StartSpan(ctx, "order.validate",
		tracer.WithSpanKind(tracer.SpanKindInternal),
		tracer.WithSpanAttributes(map[string]interface{}{
			"order.id":   orderID,
			"validation": "business_rules",
		}),
	)
	defer span.End()

	// Simulate validation logic
	time.Sleep(50 * time.Millisecond)

	if orderID == "" {
		err := fmt.Errorf("order ID cannot be empty")
		span.SetStatus(tracer.StatusCodeError, err.Error())
		return err
	}

	if data["customer_id"] == nil {
		err := fmt.Errorf("customer ID is required")
		span.SetStatus(tracer.StatusCodeError, err.Error())
		return err
	}

	span.SetStatus(tracer.StatusCodeOk, "Validation successful")
	return nil
}

func (s *GRPCService) transformOrderData(ctx context.Context, data map[string]interface{}) (map[string]interface{}, error) {
	_, span := s.tracer.StartSpan(ctx, "order.transform",
		tracer.WithSpanKind(tracer.SpanKindInternal),
		tracer.WithSpanAttributes(map[string]interface{}{
			"transformation": "normalize_data",
		}),
	)
	defer span.End()

	// Simulate data transformation
	time.Sleep(30 * time.Millisecond)

	transformed := make(map[string]interface{})
	for k, v := range data {
		transformed[k] = v
	}

	// Add processing metadata
	transformed["transformed_at"] = time.Now().Unix()
	transformed["processor_version"] = "v2.1.0"

	span.SetAttribute("fields.transformed", len(transformed))
	span.SetStatus(tracer.StatusCodeOk, "Data transformation completed")

	return transformed, nil
}

// NewMessageQueue creates a new message queue with workers
func NewMessageQueue(tr tracer.Tracer, workerCount int) *MessageQueue {
	mq := &MessageQueue{
		tracer:   tr,
		messages: make(chan Message, 1000), // Buffer for 1000 messages
		workers:  make([]*Worker, 0, workerCount),
	}

	// Start workers
	for i := 0; i < workerCount; i++ {
		worker := &Worker{
			ID:     i + 1,
			tracer: tr,
			queue:  mq,
			stop:   make(chan bool),
		}
		mq.workers = append(mq.workers, worker)
		go worker.Start()
	}

	log.Printf("Message queue started with %d workers", workerCount)
	return mq
}

// Publish sends a message to the queue
func (mq *MessageQueue) Publish(ctx context.Context, message Message) error {
	_, span := mq.tracer.StartSpan(ctx, "queue.publish",
		tracer.WithSpanKind(tracer.SpanKindProducer),
		tracer.WithSpanAttributes(map[string]interface{}{
			"message.id":   message.ID,
			"message.type": message.Type,
			"queue.name":   "order_processing",
		}),
	)
	defer span.End()

	select {
	case mq.messages <- message:
		mq.mu.Lock()
		mq.stats.Published++
		mq.mu.Unlock()

		span.SetAttribute("queue.size", len(mq.messages))
		span.SetStatus(tracer.StatusCodeOk, "Message published successfully")
		return nil
	case <-time.After(5 * time.Second):
		err := fmt.Errorf("queue is full, message publishing timeout")
		span.SetStatus(tracer.StatusCodeError, err.Error())
		span.RecordError(err, map[string]interface{}{
			"error.type": "queue_full",
		})
		return err
	}
}

// Start begins processing messages
func (w *Worker) Start() {
	log.Printf("Worker %d started", w.ID)

	for {
		select {
		case message := <-w.queue.messages:
			w.processMessage(message)
		case <-w.stop:
			log.Printf("Worker %d stopped", w.ID)
			return
		}
	}
}

func (w *Worker) processMessage(message Message) {
	// Create consumer span with parent context from message metadata
	ctx := context.Background()
	if traceID := message.Metadata["trace_id"]; traceID != "" {
		// In a real implementation, you'd reconstruct the span context
		// For this example, we'll create a new span with trace link
		// ctx = tracer.ContextWithSpan(ctx, nil) // Placeholder for context reconstruction
	}

	ctx, span := w.tracer.StartSpan(ctx, "queue.process_message",
		tracer.WithSpanKind(tracer.SpanKindConsumer),
		tracer.WithSpanAttributes(map[string]interface{}{
			"message.id":    message.ID,
			"message.type":  message.Type,
			"worker.id":     w.ID,
			"queue.name":    "order_processing",
			"message.retry": message.Retry,
		}),
	)
	defer span.End()

	w.queue.mu.Lock()
	w.queue.stats.Processing++
	w.queue.mu.Unlock()

	defer func() {
		w.queue.mu.Lock()
		w.queue.stats.Processing--
		w.queue.stats.Consumed++
		w.queue.mu.Unlock()
	}()

	span.AddEvent("processing.started", map[string]interface{}{
		"timestamp":  time.Now().Unix(),
		"worker_id":  w.ID,
		"message_id": message.ID,
	})

	// Process the message based on type
	switch message.Type {
	case "order_processing":
		if err := w.processOrder(ctx, message); err != nil {
			span.SetStatus(tracer.StatusCodeError, fmt.Sprintf("Order processing failed: %v", err))
			span.RecordError(err, map[string]interface{}{
				"error.type": "order_processing_error",
			})
			w.handleProcessingError(ctx, message, err)
			return
		}
	default:
		err := fmt.Errorf("unknown message type: %s", message.Type)
		span.SetStatus(tracer.StatusCodeError, err.Error())
		span.RecordError(err, map[string]interface{}{
			"error.type": "unknown_message_type",
		})
		return
	}

	span.AddEvent("processing.completed", map[string]interface{}{
		"timestamp":  time.Now().Unix(),
		"worker_id":  w.ID,
		"message_id": message.ID,
	})

	span.SetStatus(tracer.StatusCodeOk, "Message processed successfully")
}

func (w *Worker) processOrder(ctx context.Context, message Message) error {
	_, span := w.tracer.StartSpan(ctx, "worker.process_order",
		tracer.WithSpanKind(tracer.SpanKindInternal),
		tracer.WithSpanAttributes(map[string]interface{}{
			"order.id": message.Payload["order_id"],
			"stage":    message.Payload["processing_stage"],
		}),
	)
	defer span.End()

	orderID := message.Payload["order_id"].(string)

	// Simulate various processing stages
	stages := []string{"validation", "inventory_check", "payment_processing", "fulfillment"}

	for i, stage := range stages {
		if err := w.processOrderStage(ctx, orderID, stage); err != nil {
			span.SetStatus(tracer.StatusCodeError, fmt.Sprintf("Stage %s failed: %v", stage, err))
			return err
		}

		span.SetAttribute(fmt.Sprintf("stage.%s.completed", stage), true)

		// Simulate processing time
		time.Sleep(time.Duration(100+rand.Intn(200)) * time.Millisecond)

		// Publish intermediate results for complex workflows
		if i < len(stages)-1 {
			_ = Message{
				ID:   fmt.Sprintf("%s_stage_%d", message.ID, i+1),
				Type: "order_processing",
				Payload: map[string]interface{}{
					"order_id":         orderID,
					"processing_stage": stages[i+1],
					"previous_stage":   stage,
				},
				Metadata:  message.Metadata,
				Timestamp: time.Now(),
			}

			// In a real system, you might publish to different queues
			span.AddEvent("stage.transition", map[string]interface{}{
				"from_stage": stage,
				"to_stage":   stages[i+1],
				"order_id":   orderID,
			})
		}
	}

	span.SetStatus(tracer.StatusCodeOk, "Order processing completed")
	return nil
}

func (w *Worker) processOrderStage(ctx context.Context, orderID, stage string) error {
	_, span := w.tracer.StartSpan(ctx, fmt.Sprintf("order.stage.%s", stage),
		tracer.WithSpanKind(tracer.SpanKindInternal),
		tracer.WithSpanAttributes(map[string]interface{}{
			"order.id":   orderID,
			"stage.name": stage,
		}),
	)
	defer span.End()

	// Simulate different types of processing
	switch stage {
	case "validation":
		// Simulate business rule validation
		time.Sleep(50 * time.Millisecond)
		span.SetAttribute("validation.rules_checked", 15)

	case "inventory_check":
		// Simulate inventory system call
		if err := w.simulateExternalCall(ctx, "inventory-service", "GET", "/api/check"); err != nil {
			return err
		}
		span.SetAttribute("inventory.items_checked", 3)

	case "payment_processing":
		// Simulate payment gateway call
		if err := w.simulateExternalCall(ctx, "payment-gateway", "POST", "/api/charge"); err != nil {
			return err
		}
		span.SetAttribute("payment.amount", 99.99)
		span.SetAttribute("payment.currency", "USD")

	case "fulfillment":
		// Simulate warehouse system call
		if err := w.simulateExternalCall(ctx, "warehouse-system", "POST", "/api/ship"); err != nil {
			return err
		}
		span.SetAttribute("fulfillment.carrier", "UPS")
		span.SetAttribute("fulfillment.tracking_number", "1Z999AA1234567890")
	}

	span.SetStatus(tracer.StatusCodeOk, fmt.Sprintf("Stage %s completed", stage))
	return nil
}

func (w *Worker) simulateExternalCall(ctx context.Context, service, method, endpoint string) error {
	_, span := w.tracer.StartSpan(ctx, fmt.Sprintf("external.%s", service),
		tracer.WithSpanKind(tracer.SpanKindClient),
		tracer.WithSpanAttributes(map[string]interface{}{
			"external.service": service,
			"http.method":      method,
			"http.endpoint":    endpoint,
		}),
	)
	defer span.End()

	// Simulate network latency and occasional failures
	latency := time.Duration(100+rand.Intn(300)) * time.Millisecond
	time.Sleep(latency)

	// 5% chance of failure
	if rand.Float32() < 0.05 {
		err := fmt.Errorf("external service %s temporarily unavailable", service)
		span.SetStatus(tracer.StatusCodeError, err.Error())
		span.RecordError(err, map[string]interface{}{
			"error.type": "external_service_error",
		})
		return err
	}

	span.SetAttribute("external.latency_ms", latency.Milliseconds())
	span.SetStatus(tracer.StatusCodeOk, "External call successful")
	return nil
}

func (w *Worker) handleProcessingError(ctx context.Context, message Message, err error) {
	_, span := w.tracer.StartSpan(ctx, "queue.handle_error",
		tracer.WithSpanKind(tracer.SpanKindInternal),
		tracer.WithSpanAttributes(map[string]interface{}{
			"message.id":    message.ID,
			"error.message": err.Error(),
			"retry.count":   message.Retry,
		}),
	)
	defer span.End()

	w.queue.mu.Lock()
	w.queue.stats.Failed++
	w.queue.mu.Unlock()

	maxRetries := 3
	if message.Retry < maxRetries {
		// Retry with exponential backoff
		retryDelay := time.Duration(1<<uint(message.Retry)) * time.Second

		span.AddEvent("retry.scheduled", map[string]interface{}{
			"retry_count":   message.Retry + 1,
			"delay_seconds": retryDelay.Seconds(),
		})

		time.Sleep(retryDelay)

		message.Retry++
		w.queue.mu.Lock()
		w.queue.stats.Retried++
		w.queue.mu.Unlock()

		select {
		case w.queue.messages <- message:
			span.SetStatus(tracer.StatusCodeOk, "Message requeued for retry")
		default:
			span.SetStatus(tracer.StatusCodeError, "Failed to requeue message")
		}
	} else {
		// Send to dead letter queue (simulated)
		span.AddEvent("dead_letter.sent", map[string]interface{}{
			"reason":  "max_retries_exceeded",
			"retries": message.Retry,
		})
		span.SetStatus(tracer.StatusCodeError, "Message sent to dead letter queue")
	}
}

// GetStats returns current queue statistics
func (mq *MessageQueue) GetStats() QueueStats {
	mq.mu.RLock()
	defer mq.mu.RUnlock()
	return mq.stats
}

// Shutdown stops all workers and closes the queue
func (mq *MessageQueue) Shutdown() {
	for _, worker := range mq.workers {
		close(worker.stop)
	}
	close(mq.messages)
}

func main() {
	// Configure New Relic provider
	config := &newrelic.Config{
		AppName:           "grpc-messagequeue-example",
		LicenseKey:        "your-license-key", // In real app, use environment variable
		Enabled:           true,
		Environment:       "development",
		DistributedTracer: true,
		MaxSamplesStored:  1000,
		FlushInterval:     10 * time.Second,
		AttributesEnabled: true,
		DatastoreTracer:   true,
		CodeLevelMetrics:  true,
	}

	// Create provider
	provider, err := newrelic.NewProvider(config)
	if err != nil {
		log.Fatalf("Failed to create New Relic provider: %v", err)
	}

	// Create tracer
	tr, err := provider.CreateTracer("grpc-mq",
		tracer.WithServiceName("grpc-messagequeue-example"),
		tracer.WithEnvironment("development"),
		tracer.WithSamplingRate(1.0),
	)
	if err != nil {
		log.Fatalf("Failed to create tracer: %v", err)
	}

	// Create message queue with 3 workers
	queue := NewMessageQueue(tr, 3)
	defer queue.Shutdown()

	// Create gRPC service
	grpcService := &GRPCService{
		tracer: tr,
		queue:  queue,
	}

	// Simulate multiple gRPC requests
	fmt.Println("Starting gRPC & Message Queue example...")
	fmt.Println("Simulating order processing requests...")

	orders := []struct {
		ID           string
		CustomerData map[string]interface{}
	}{
		{
			ID: "order_001",
			CustomerData: map[string]interface{}{
				"customer_id": "cust_123",
				"email":       "customer@example.com",
				"tier":        "premium",
			},
		},
		{
			ID: "order_002",
			CustomerData: map[string]interface{}{
				"customer_id": "cust_456",
				"email":       "customer2@example.com",
				"tier":        "standard",
			},
		},
		{
			ID: "order_003",
			CustomerData: map[string]interface{}{
				"customer_id": "cust_789",
				"email":       "customer3@example.com",
				"tier":        "premium",
			},
		},
	}

	// Process orders concurrently
	var wg sync.WaitGroup
	for i, order := range orders {
		wg.Add(1)
		go func(idx int, o struct {
			ID           string
			CustomerData map[string]interface{}
		}) {
			defer wg.Done()

			// Simulate staggered requests
			time.Sleep(time.Duration(idx) * 2 * time.Second)

			ctx := context.Background()
			if err := grpcService.ProcessOrderRequest(ctx, o.ID, o.CustomerData); err != nil {
				log.Printf("Failed to process order %s: %v", o.ID, err)
			} else {
				log.Printf("Order %s submitted for processing", o.ID)
			}
		}(i, order)
	}

	wg.Wait()

	// Let workers process messages
	fmt.Println("Waiting for message processing...")
	time.Sleep(10 * time.Second)

	// Print final statistics
	stats := queue.GetStats()
	fmt.Printf("\nFinal Queue Statistics:\n")
	fmt.Printf("Published: %d\n", stats.Published)
	fmt.Printf("Consumed: %d\n", stats.Consumed)
	fmt.Printf("Failed: %d\n", stats.Failed)
	fmt.Printf("Retried: %d\n", stats.Retried)
	fmt.Printf("Currently Processing: %d\n", stats.Processing)

	fmt.Println("\nExample completed. Check your New Relic dashboard for traces!")

	// Cleanup
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := provider.Shutdown(ctx); err != nil {
		log.Printf("Error shutting down provider: %v", err)
	}
}
