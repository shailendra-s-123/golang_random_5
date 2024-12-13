// main.go
import (
    "context"
    "log"
    "net/http"
    "yourmodule/errorreporter"
)

func main() {
    if err := errorreporter.Init("YOUR_SENTRY_DSN_HERE"); err != nil {
        log.Fatalf("Failed to initialize Sentry: %v", err)
    }

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        ctx := context.WithValue(r.Context(), "user_id", "123")
        ctx = context.WithValue(r.Context(), "request_id", r.Header.Get("X-Request-ID"))

        err := riskyOperation(ctx)
        if err != nil {
            tags := map[string]string{
                "error_type": "logic",
                "endpoint":   r.URL.Path,
            }
            errorreporter.CaptureError(ctx, err, tags)
            http.Error(w, "Internal Server Error", http.StatusInternalServerError)
            return
        }
        w.Write([]byte("Success!"))
    })

    http.ListenAndServe(":8080", nil)
}

func riskyOperation(ctx context.Context) error {
    // Simulate a logic error
    return &LogicError{msg: "business logic failed"}
}