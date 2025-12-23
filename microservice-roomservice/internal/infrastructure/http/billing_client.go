package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/shopspring/decimal"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"microservice-roomservice/pkg/logger"
)

// BillingClient is an HTTP client for the billing service
type BillingClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewBillingClient creates a new billing client
func NewBillingClient(baseURL string) *BillingClient {
	return &BillingClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
			// Use default transport without otelhttp wrapping
			// We'll handle tracing manually
		},
	}
}

// CreateFolderRequest represents the request to create a billing folder
type CreateFolderRequest struct {
	Name string `json:"name"`
}

// FolderResponse represents a billing folder from the billing service
type FolderResponse struct {
	ID        string  `json:"id"`
	Name      *string `json:"name"`
	CreatedAt string  `json:"created_at"`
	ClosedAt  *string `json:"closed_at"`
}

// CreateTicketRequestJSON represents the request to create a billing ticket
type CreateTicketRequestJSON struct {
	FolderID    string `json:"folder_id"`
	Total       string `json:"total"`
	ItemID      string `json:"item_id"`
	Description string `json:"description,omitempty"`
}

// TicketResponse represents a billing ticket from the billing service
type TicketResponse struct {
	ID          string         `json:"id"`
	Folder      FolderResponse `json:"folder"`
	Total       string         `json:"total"`
	ItemID      string         `json:"item_id"`
	Description *string        `json:"description"`
	State       string         `json:"state"`
	CreatedAt   string         `json:"created_at"`
	ClosedAt    *string        `json:"closed_at"`
}

// CreateTicket creates a new billing ticket
func (c *BillingClient) CreateTicket(ctx context.Context, folderID string, total decimal.Decimal, itemID, description string) (*TicketResponse, error) {
	// Check if context has a parent span
	parentSpan := trace.SpanFromContext(ctx)
	parentSpanCtx := parentSpan.SpanContext()
	logger.CtxDebug(ctx).Msgf("BillingClient.CreateTicket - Parent span context valid: %v, traceID: %s, spanID: %s",
		parentSpanCtx.IsValid(),
		parentSpanCtx.TraceID().String(),
		parentSpanCtx.SpanID().String())

	tracer := otel.Tracer("billing-client")
	logger.CtxDebug(ctx).Msgf("BillingClient.CreateTicket - Tracer type: %T", tracer)

	ctx, span := tracer.Start(ctx, "POST billing ticket")
	defer span.End()

	spanCtx := span.SpanContext()
	logger.CtxDebug(ctx).Msgf("BillingClient.CreateTicket - Created span: traceID=%s spanID=%s isRecording=%v isValid=%v isSampled=%v",
		spanCtx.TraceID().String(),
		spanCtx.SpanID().String(),
		span.IsRecording(),
		spanCtx.IsValid(),
		spanCtx.IsSampled())

	span.SetAttributes(
		attribute.String("folder.id", folderID),
		attribute.String("item.id", itemID),
	)

	url := fmt.Sprintf("%s/tickets", c.baseURL)

	reqBody := CreateTicketRequestJSON{
		FolderID:    folderID,
		Total:       total.String(),
		ItemID:      itemID,
		Description: description,
	}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	// Manually inject trace context into HTTP headers
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create ticket: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("billing service error creating ticket: %s (status: %d)", string(body), resp.StatusCode)
	}

	var ticket TicketResponse
	if err := json.NewDecoder(resp.Body).Decode(&ticket); err != nil {
		return nil, fmt.Errorf("failed to decode ticket response: %w", err)
	}

	return &ticket, nil
}

// DeleteTicket deletes a billing ticket by ID
func (c *BillingClient) DeleteTicket(ctx context.Context, ticketID string) error {
	// Check if context has a parent span
	parentSpan := trace.SpanFromContext(ctx)
	parentSpanCtx := parentSpan.SpanContext()
	logger.CtxDebug(ctx).Msgf("BillingClient.DeleteTicket - Parent span context valid: %v, traceID: %s, spanID: %s",
		parentSpanCtx.IsValid(),
		parentSpanCtx.TraceID().String(),
		parentSpanCtx.SpanID().String())

	tracer := otel.Tracer("billing-client")
	logger.CtxDebug(ctx).Msgf("BillingClient.DeleteTicket - Tracer type: %T", tracer)

	ctx, span := tracer.Start(ctx, "DELETE billing ticket")
	defer span.End()

	spanCtx := span.SpanContext()
	logger.CtxDebug(ctx).Msgf("BillingClient.DeleteTicket - Created span: traceID=%s spanID=%s isRecording=%v isValid=%v isSampled=%v",
		spanCtx.TraceID().String(),
		spanCtx.SpanID().String(),
		span.IsRecording(),
		spanCtx.IsValid(),
		spanCtx.IsSampled())

	span.SetAttributes(attribute.String("ticket.id", ticketID))

	url := fmt.Sprintf("%s/tickets/%s", c.baseURL, ticketID)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	// Manually inject trace context into HTTP headers
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete ticket: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("billing service error deleting ticket: %s (status: %d)", string(body), resp.StatusCode)
	}

	return nil
}
