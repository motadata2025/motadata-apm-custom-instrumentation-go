package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	motadata "github.com/motadata2025/motadata-apm-custom-instrumentation-go"
)

// ---- HTTP Handler layer ----

type UserHandler struct {
	repo *UserRepo
}

type CreateUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Age      int    `json:"age"`
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	ctx, span, _ := motadata.StartSpan(r.Context(), "CreateUser")
	defer span.End()

	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		span.RecordError(err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	_ = span.SetString("http.method", r.Method)
	_ = span.SetString("http.url", r.URL.Path)
	_ = span.SetString("operation", "create_user")
	_ = span.SetString("user.username", req.Username)
	_ = span.SetString("user.email", req.Email)
	_ = span.SetInt("user.age", int64(req.Age))

	user, err := h.repo.CreateUser(ctx, req)
	if err != nil {
		span.RecordError(err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(user)
}

// ---- Repository / Data Access layer ----

type UserRepo struct{}

type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

func (r *UserRepo) CreateUser(ctx context.Context, req CreateUserRequest) (*User, error) {
	_, span, _ := motadata.StartSpan(ctx, "db:CreateUser")
	defer span.End()

	_ = span.SetString("db.operation", "INSERT")
	_ = span.SetString("db.table", "users")
	_ = span.SetString("db.param.username", req.Username)
	_ = span.SetString("db.param.email", req.Email)
	_ = span.SetBool("db.transaction", true)

	// Simulated DB result — replace with real DB call passing ctx
	return &User{ID: 1, Username: req.Username, Email: req.Email}, nil
}

// ---- Service / Business Logic layer ----

func ProcessOrder(ctx context.Context, orderID string, itemCount int64, tags []string) error {
	_, span, _ := motadata.StartSpan(ctx, "ProcessOrder")
	defer span.End()

	_ = span.SetString("order.id", orderID)
	_ = span.SetString("operation", "process_order")
	_ = span.SetBool("order.express", true)
	_ = span.SetInt("order.items", itemCount)
	_ = span.SetFloat("order.weight.kg", 2.5)
	_ = span.SetStringSlice("order.tags", tags)

	return nil
}

func main() {
	repo := &UserRepo{}
	handler := &UserHandler{repo: repo}

	// ProcessOrder demonstrates service-layer instrumentation.
	// In a real app this would be called with a real context from an active request.
	if err := ProcessOrder(context.Background(), "ORD-001", 3, []string{"urgent", "vip"}); err != nil {
		log.Printf("ProcessOrder: %v", err)
	}

	http.HandleFunc("/users", handler.CreateUser)

	log.Println("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
