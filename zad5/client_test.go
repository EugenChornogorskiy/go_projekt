package httpclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestGetItems_Success(t *testing.T) { 
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { 
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		 
		if r.URL.Path != "/items" {
			t.Errorf("expected /items, got %s", r.URL.Path)
		}
		 
		if ua := r.Header.Get("User-Agent"); ua != "go-http-client-pv" {
			t.Errorf("expected User-Agent 'go-http-client-pv', got '%s'", ua)
		}
		 
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := `[{"id":1,"name":"First item","description":"Example description"}]`
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()
	
	client := NewClient(server.URL)
	items, err := client.GetItems(context.Background())
	
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	
	if items[0].ID != 1 {
		t.Errorf("expected ID 1, got %d", items[0].ID)
	}
	
	if items[0].Name != "First item" {
		t.Errorf("expected Name 'First item', got '%s'", items[0].Name)
	}
}

func TestGetItems_EmptyList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("[]"))
	}))
	defer server.Close()
	
	client := NewClient(server.URL)
	items, err := client.GetItems(context.Background())
	
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	if len(items) != 0 {
		t.Fatalf("expected 0 items, got %d", len(items))
	}
}

func TestGetItems_ErrorStatusCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()
	
	client := NewClient(server.URL)
	items, err := client.GetItems(context.Background())
	
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	
	if items != nil {
		t.Errorf("expected nil items, got %v", items)
	}
	
	expectedErrMsg := "unexpected status code: 500"
	if !strings.Contains(err.Error(), expectedErrMsg) {
		t.Errorf("expected error containing '%s', got '%s'", expectedErrMsg, err.Error())
	}
}

func TestGetItems_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{invalid json}`))
	}))
	defer server.Close()
	
	client := NewClient(server.URL)
	items, err := client.GetItems(context.Background())
	
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	
	if items != nil {
		t.Errorf("expected nil items, got %v", items)
	}
}

func TestGetItems_ContextCancel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	
	client := NewClient(server.URL)
	
	ctx, cancel := context.WithCancel(context.Background())
	cancel()  
	
	items, err := client.GetItems(ctx)
	
	if err == nil {
		t.Fatal("expected error from cancelled context, got nil")
	}
	
	if items != nil {
		t.Errorf("expected nil items, got %v", items)
	}
}

func TestCreateItem_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { 
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		 
		if r.URL.Path != "/items" {
			t.Errorf("expected /items, got %s", r.URL.Path)
		}
		 
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected Content-Type 'application/json', got '%s'", ct)
		}
		 
		if ua := r.Header.Get("User-Agent"); ua != "go-http-client-pv" {
			t.Errorf("expected User-Agent 'go-http-client-pv', got '%s'", ua)
		}
		 
		var req CreateItemRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		
		if req.Name != "New item" {
			t.Errorf("expected name 'New item', got '%s'", req.Name)
		}
		 
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		response := `{"id":3,"name":"New item","description":"Description of the new item"}`
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()
	
	client := NewClient(server.URL)
	
	input := CreateItemRequest{
		Name:        "New item",
		Description: "Description of the new item",
	}
	
	item, err := client.CreateItem(context.Background(), input)
	
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	if item == nil {
		t.Fatal("expected item, got nil")
	}
	
	if item.ID != 3 {
		t.Errorf("expected ID 3, got %d", item.ID)
	}
	
	if item.Name != "New item" {
		t.Errorf("expected Name 'New item', got '%s'", item.Name)
	}
}

func TestCreateItem_ErrorStatusCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()
	
	client := NewClient(server.URL)
	
	input := CreateItemRequest{
		Name:        "Test",
		Description: "Test",
	}
	
	item, err := client.CreateItem(context.Background(), input)
	
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	
	if item != nil {
		t.Errorf("expected nil item, got %v", item)
	}
	
	expectedErrMsg := "unexpected status code: 400"
	if !strings.Contains(err.Error(), expectedErrMsg) {
		t.Errorf("expected error containing '%s', got '%s'", expectedErrMsg, err.Error())
	}
}

func TestCreateItem_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{invalid json}`))
	}))
	defer server.Close()
	
	client := NewClient(server.URL)
	
	input := CreateItemRequest{
		Name:        "Test",
		Description: "Test",
	}
	
	item, err := client.CreateItem(context.Background(), inpt)
	
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	
	if item != nil {
		t.Errorf("expected nil item, got %v", item)
	}
}

func TestCreateItem_ContextTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()
	
	client := NewClient(server.URL)
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	
	input := CreateItemRequest{
		Name:        "Test",
		Description: "Test",
	}
	
	item, err := client.CreateItem(ctx, input)
	
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	
	if item != nil {
		t.Errorf("expected nil item, got %v", item)
	}
}

func TestUserAgentTransport_AddsUserAgent(t *testing.T) { 
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ua := r.Header.Get("User-Agent")
		if ua != "go-http-client-pv" {
			t.Errorf("expected User-Agent 'go-http-client-pv', got '%s'", ua)
		}
		w.WriteHeader(http.StatusOK)
		response := `[{"id":1,"name":"First item","description":"Example description"}]`
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()
	
	client := NewClient(server.URL)
	 
	_, err := client.GetItems(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_ReusesHTTPClient(t *testing.T) {
	client1 := NewClient("http://example.com")
	client2 := NewClient("http://example.com")
	 
	if client1.httpClient == nil {
		t.Error("httpClient should not be nil")
	}
	 
	transport, ok := client1.httpClient.Transport.(*UserAgentTransport)
	if !ok {
		t.Error("expected UserAgentTransport")
	}
	
	if transport.UserAgent != "go-http-client-pv" {
		t.Errorf("expected UserAgent 'go-http-client-pv', got '%s'", transport.UserAgent)
	}
	 
	if client1 == client2 {
		t.Error("clients should be different instances")
	}
}