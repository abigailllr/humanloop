package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/abigailtech/humanloop/backend/db"
	"github.com/abigailtech/humanloop/backend/handlers"
	"github.com/abigailtech/humanloop/backend/middleware"
	"github.com/abigailtech/humanloop/backend/pipeline"
	"github.com/abigailtech/humanloop/backend/queue"
	"github.com/abigailtech/humanloop/backend/storage"
)

func main() {
	ctx := context.Background()

	if os.Getenv("JWT_SECRET") == "" {
		log.Println("warning: JWT_SECRET not set, using insecure default")
	}

	extractedDir := os.Getenv("EXTRACTED_DIR")
	if extractedDir == "" {
		extractedDir = "./data/extracted"
	}

	var hmdfUpload func(string, io.Reader) (string, error)

	if os.Getenv("AWS_BUCKET") != "" {
		s3store, err := storage.NewS3Store()
		if err != nil {
			log.Fatal("s3 init failed:", err)
		}
		handlers.Store = s3store
		hmdfUpload = s3store.UploadHMDF
	}

	handlers.Pipeline = pipeline.New(4, extractedDir, hmdfUpload)

	if err := db.Connect(ctx); err != nil {
		log.Println("db unavailable:", err)
	} else {
		defer db.Close()
		if err := db.Migrate(ctx); err != nil {
			log.Println("db migrate:", err)
		} else {
			if err := db.SeedChallenges(ctx); err != nil {
				log.Println("db seed:", err)
			}
		}
		middleware.BuyerKeyFunc = func(ctx context.Context, hash string) (string, bool) {
			k, err := db.GetBuyerKeyByHash(ctx, hash)
			if err != nil {
				return "", false
			}
			return k.DatasetID, true
		}
	}

	q := queue.New()
	if err := q.Ping(ctx); err != nil {
		log.Println("redis unavailable:", err)
	} else {
		handlers.Queue = q
		go func() {
			for {
				data, err := q.Pop(ctx)
				if err != nil {
					continue
				}
				var job pipeline.Job
				if err := json.Unmarshal(data, &job); err != nil {
					continue
				}
				handlers.Pipeline.Enqueue(job)
			}
		}()
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", handlers.Health)
	mux.Handle("GET /metrics", middleware.APIKey(http.HandlerFunc(handlers.Metrics)))

	authRL := middleware.RateLimitAuth
	submitRL := middleware.RateLimitSubmit

	mux.Handle("POST /v1/auth/google", authRL(http.HandlerFunc(handlers.AuthGoogle)))
	mux.Handle("POST /v1/auth/apple", authRL(http.HandlerFunc(handlers.AuthApple)))

	mux.HandleFunc("GET /v1/challenges", handlers.GetChallenges)
	mux.Handle("POST /v1/challenges", middleware.APIKey(http.HandlerFunc(handlers.CreateChallenge)))
	mux.Handle("PUT /v1/challenges/{id}", middleware.APIKey(http.HandlerFunc(handlers.UpdateChallenge)))
	mux.Handle("DELETE /v1/challenges/{id}", middleware.APIKey(http.HandlerFunc(handlers.DeleteChallenge)))

	mux.Handle("POST /v1/submit/{challengeId}", submitRL(middleware.Auth(http.HandlerFunc(handlers.Submit))))
	mux.Handle("GET /v1/profile", middleware.Auth(http.HandlerFunc(handlers.GetProfile)))
	mux.Handle("GET /v1/submissions/{id}", middleware.Auth(http.HandlerFunc(handlers.GetSubmission)))
	mux.Handle("GET /v1/submissions", middleware.Auth(http.HandlerFunc(handlers.GetUserSubmissions)))
	mux.HandleFunc("GET /v1/leaderboard", handlers.GetLeaderboard)

	mux.Handle("GET /v1/data/export", middleware.APIKey(http.HandlerFunc(handlers.ExportData)))
	mux.Handle("GET /v1/data/stats", middleware.APIKey(http.HandlerFunc(handlers.ExportStats)))

	buyerOrAdmin := middleware.BuyerKeyOrAPIKey
	mux.Handle("GET /v1/datasets", buyerOrAdmin(http.HandlerFunc(handlers.GetDatasets)))
	mux.Handle("POST /v1/datasets", middleware.APIKey(http.HandlerFunc(handlers.CreateDataset)))
	mux.Handle("DELETE /v1/datasets/{id}", middleware.APIKey(http.HandlerFunc(handlers.DeleteDataset)))
	mux.Handle("GET /v1/datasets/{id}/export", buyerOrAdmin(http.HandlerFunc(handlers.ExportDataset)))
	mux.Handle("GET /v1/datasets/{id}/submissions", buyerOrAdmin(http.HandlerFunc(handlers.GetDatasetSubmissionList)))
	mux.Handle("GET /v1/datasets/{id}/stats", buyerOrAdmin(http.HandlerFunc(handlers.GetDatasetStats)))

	mux.HandleFunc("GET /v1/robots", handlers.GetRobots)

	mux.Handle("GET /v1/admin/submissions", middleware.APIKey(http.HandlerFunc(handlers.AdminGetSubmissions)))
	mux.Handle("POST /v1/admin/submissions/{id}/approve", middleware.APIKey(http.HandlerFunc(handlers.AdminApproveSubmission)))
	mux.Handle("POST /v1/admin/submissions/{id}/reject", middleware.APIKey(http.HandlerFunc(handlers.AdminRejectSubmission)))
	mux.Handle("PUT /v1/admin/submissions/{id}/tags", middleware.APIKey(http.HandlerFunc(handlers.AdminTagSubmission)))

	mux.Handle("GET /v1/admin/buyer-keys", middleware.APIKey(http.HandlerFunc(handlers.ListBuyerKeys)))
	mux.Handle("POST /v1/admin/buyer-keys", middleware.APIKey(http.HandlerFunc(handlers.CreateBuyerKey)))
	mux.Handle("DELETE /v1/admin/buyer-keys/{id}", middleware.APIKey(http.HandlerFunc(handlers.DeleteBuyerKey)))

	stack := middleware.Security(middleware.RealIP(middleware.Logger(mux)))
	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", stack))
}
