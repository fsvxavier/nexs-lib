package nethttp

import (
	"context"
	"net/http"
	"testing"

	"github.com/dock-tech/isis-golang-lib/domainerrors"
)

func TestWrapErrors_BadRequest_WithDescription(t *testing.T) {
	ctx := context.Background()
	body := []byte(`{"error": {"description": "invalid input"}}`)
	req := Requester{BaseURL: "http://target"}
	resp := &Response{
		Body:       body,
		StatusCode: http.StatusBadRequest,
		Request:    req,
	}

	err := WrapResponseErrors(ctx, resp, "field")
	badReqErr, ok := err.(*domainerrors.BadRequestError)
	if !ok {
		t.Fatalf("expected BadRequestError, got %T", err)
	}
	if badReqErr.Details["field"][0] != "invalid input" {
		t.Errorf("unexpected details: %v", badReqErr.Details)
	}
}

func TestWrapErrors_BadRequest_WithoutDescription(t *testing.T) {
	ctx := context.Background()
	body := []byte(`{"error": {}}`)

	req := Requester{BaseURL: "http://target"}

	resp := &Response{
		Body:       body,
		StatusCode: http.StatusBadRequest,
		Request:    req,
	}

	err := WrapResponseErrors(ctx, resp, "field")
	serverErr, ok := err.(*domainerrors.ServerError)
	if !ok {
		t.Fatalf("expected ServerError, got %T", err)
	}
	if serverErr.InternalError == nil {
		t.Error("expected internal error to be set")
	}
}

func TestWrapErrors_UnprocessableEntity_WithDescription(t *testing.T) {
	ctx := context.Background()
	body := []byte(`{"error": {"description": "unprocessable"}}`)
	req := Requester{BaseURL: "http://target"}
	resp := &Response{
		Body:       body,
		StatusCode: http.StatusUnprocessableEntity,
		Request:    req,
	}

	err := WrapResponseErrors(ctx, resp, "field")
	usecaseErr, ok := err.(*domainerrors.UsecaseError)
	if !ok {
		t.Fatalf("expected UsecaseError, got %T", err)
	}
	if usecaseErr.Description != "unprocessable" {
		t.Errorf("unexpected description: %v", usecaseErr.Description)
	}
}

func TestWrapErrors_UnprocessableEntity_WithoutDescription(t *testing.T) {
	ctx := context.Background()
	body := []byte(`{"error": {}}`)
	req := Requester{BaseURL: "http://target"}
	resp := &Response{
		Body:       body,
		StatusCode: http.StatusUnprocessableEntity,
		Request:    req,
	}

	err := WrapResponseErrors(ctx, resp, "field")
	serverErr, ok := err.(*domainerrors.ServerError)
	if !ok {
		t.Fatalf("expected ServerError, got %T", err)
	}
	if serverErr.InternalError == nil {
		t.Error("expected internal error to be set")
	}
}

func TestWrapErrors_NotFound_WithDescription(t *testing.T) {
	ctx := context.Background()
	body := []byte(`{"error": {"description": "not found"}}`)
	req := Requester{BaseURL: "http://target"}
	resp := &Response{
		Body:       body,
		StatusCode: http.StatusNotFound,
		Request:    req,
	}

	err := WrapResponseErrors(ctx, resp, "field")
	_, ok := err.(*domainerrors.NotFoundError)
	if !ok {
		t.Fatalf("expected NotFoundError, got %T", err)
	}
}

func TestWrapErrors_NotFound_WithoutDescription(t *testing.T) {
	ctx := context.Background()
	body := []byte(`{"error": {}}`)
	req := Requester{BaseURL: "http://target"}
	resp := &Response{
		Body:       body,
		StatusCode: http.StatusNotFound,
		Request:    req,
	}

	err := WrapResponseErrors(ctx, resp, "field")
	serverErr, ok := err.(*domainerrors.ServerError)
	if !ok {
		t.Fatalf("expected ServerError, got %T", err)
	}
	if serverErr.InternalError == nil {
		t.Error("expected internal error to be set")
	}
}

func TestWrapErrors_InvalidJSON(t *testing.T) {
	ctx := context.Background()
	body := []byte(`invalid json`)
	req := Requester{BaseURL: "http://target"}
	resp := &Response{
		Body:       body,
		StatusCode: http.StatusBadRequest,
		Request:    req,
	}

	err := WrapResponseErrors(ctx, resp, "field")
	serverErr, ok := err.(*domainerrors.ServerError)
	if !ok {
		t.Fatalf("expected ServerError, got %T", err)
	}
	if serverErr.InternalError == nil {
		t.Error("expected internal error to be set")
	}
}

func TestWrapErrors_DefaultCase(t *testing.T) {
	ctx := context.Background()
	body := []byte(`{"error": {"description": "server error"}}`)
	req := Requester{BaseURL: "http://target"}
	resp := &Response{
		Body:       body,
		StatusCode: http.StatusInternalServerError,
		Request:    req,
	}

	err := WrapResponseErrors(ctx, resp, "field")
	serverErr, ok := err.(*domainerrors.ServerError)
	if !ok {
		t.Fatalf("expected ServerError, got %T", err)
	}
	if serverErr.InternalError == nil {
		t.Error("expected internal error to be set")
	}
}
