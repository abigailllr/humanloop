package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
		go func() {
			for range time.Tick(6 * time.Hour) {
				db.PurgeExpiredRefreshTokens(context.Background())
			}
		}()
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
	mux.Handle("POST /v1/auth/refresh", authRL(http.HandlerFunc(handlers.AuthRefresh)))
	mux.Handle("POST /v1/auth/revoke", middleware.Auth(http.HandlerFunc(handlers.AuthRevoke)))

	mux.HandleFunc("GET /v1/challenges", handlers.GetChallenges)
	mux.Handle("POST /v1/challenges", middleware.APIKey(http.HandlerFunc(handlers.CreateChallenge)))
	mux.Handle("PUT /v1/challenges/{id}", middleware.APIKey(http.HandlerFunc(handlers.UpdateChallenge)))
	mux.Handle("DELETE /v1/challenges/{id}", middleware.APIKey(http.HandlerFunc(handlers.DeleteChallenge)))
	mux.HandleFunc("GET /v1/challenges/{id}/stats", handlers.GetChallengeStats)
	mux.HandleFunc("GET /v1/challenges/{id}/leaderboard", handlers.GetChallengeLeaderboard)

	mux.Handle("POST /v1/submit/{challengeId}", submitRL(middleware.Auth(http.HandlerFunc(handlers.Submit))))
	mux.Handle("GET /v1/profile", middleware.Auth(http.HandlerFunc(handlers.GetProfile)))
	mux.Handle("GET /v1/profile/stats", middleware.Auth(http.HandlerFunc(handlers.GetUserStats)))
	mux.Handle("GET /v1/credits/history", middleware.Auth(http.HandlerFunc(handlers.GetCreditHistory)))
	mux.Handle("GET /v1/submissions/{id}", middleware.Auth(http.HandlerFunc(handlers.GetSubmission)))
	mux.Handle("GET /v1/submissions/{id}/stream", middleware.Auth(http.HandlerFunc(handlers.StreamSubmission)))
	mux.Handle("GET /v1/submissions", middleware.Auth(http.HandlerFunc(handlers.GetUserSubmissions)))
	mux.HandleFunc("GET /v1/leaderboard", handlers.GetLeaderboard)
	mux.HandleFunc("GET /v1/robots", handlers.GetRobots)

	mux.Handle("GET /v1/data/export", middleware.APIKey(http.HandlerFunc(handlers.ExportData)))
	mux.Handle("GET /v1/data/stats", middleware.APIKey(http.HandlerFunc(handlers.ExportStats)))
	mux.HandleFunc("GET /v1/data/heatmap", handlers.GetHeatmap)

	buyerOrAdmin := middleware.BuyerKeyOrAPIKey
	mux.Handle("GET /v1/datasets", buyerOrAdmin(http.HandlerFunc(handlers.GetDatasets)))
	mux.Handle("POST /v1/datasets", middleware.APIKey(http.HandlerFunc(handlers.CreateDataset)))
	mux.Handle("PATCH /v1/datasets/{id}", middleware.APIKey(http.HandlerFunc(handlers.PatchDataset)))
	mux.Handle("DELETE /v1/datasets/{id}", middleware.APIKey(http.HandlerFunc(handlers.DeleteDataset)))
	mux.Handle("GET /v1/datasets/{id}/export", buyerOrAdmin(http.HandlerFunc(handlers.ExportDataset)))
	mux.Handle("GET /v1/datasets/{id}/submissions", buyerOrAdmin(http.HandlerFunc(handlers.GetDatasetSubmissionList)))
	mux.Handle("GET /v1/datasets/{id}/stats", buyerOrAdmin(http.HandlerFunc(handlers.GetDatasetStats)))
	mux.Handle("GET /v1/datasets/{datasetId}/submissions/{submissionId}/download", buyerOrAdmin(http.HandlerFunc(handlers.DownloadHMDF)))

	mux.Handle("GET /v1/admin/submissions", middleware.APIKey(http.HandlerFunc(handlers.AdminGetSubmissions)))
	mux.Handle("POST /v1/admin/submissions/{id}/approve", middleware.APIKey(http.HandlerFunc(handlers.AdminApproveSubmission)))
	mux.Handle("POST /v1/admin/submissions/{id}/reject", middleware.APIKey(http.HandlerFunc(handlers.AdminRejectSubmission)))
	mux.Handle("PUT /v1/admin/submissions/{id}/tags", middleware.APIKey(http.HandlerFunc(handlers.AdminTagSubmission)))
	mux.Handle("POST /v1/admin/submissions/{id}/retry", middleware.APIKey(http.HandlerFunc(handlers.AdminRetrySubmission)))

	mux.Handle("GET /v1/admin/dlq", middleware.APIKey(http.HandlerFunc(handlers.AdminGetDLQ)))
	mux.Handle("POST /v1/admin/dlq/{id}/retry", middleware.APIKey(http.HandlerFunc(handlers.AdminRetryDLQ)))

	mux.Handle("GET /v1/admin/buyer-keys", middleware.APIKey(http.HandlerFunc(handlers.ListBuyerKeys)))
	mux.Handle("POST /v1/admin/buyer-keys", middleware.APIKey(http.HandlerFunc(handlers.CreateBuyerKey)))
	mux.Handle("DELETE /v1/admin/buyer-keys/{id}", middleware.APIKey(http.HandlerFunc(handlers.DeleteBuyerKey)))

	mux.Handle("GET /v1/admin/webhooks", middleware.APIKey(http.HandlerFunc(handlers.ListWebhooks)))
	mux.Handle("POST /v1/admin/webhooks", middleware.APIKey(http.HandlerFunc(handlers.CreateWebhook)))
	mux.Handle("DELETE /v1/admin/webhooks/{id}", middleware.APIKey(http.HandlerFunc(handlers.DeleteWebhook)))

	mux.Handle("GET /v1/admin/analytics", middleware.APIKey(http.HandlerFunc(handlers.AdminGetAnalytics)))
	mux.Handle("DELETE /v1/account", middleware.Auth(http.HandlerFunc(handlers.DeleteAccount)))

	mux.Handle("GET /v1/referral", middleware.Auth(http.HandlerFunc(handlers.GetReferral)))
	mux.Handle("POST /v1/referral/redeem", middleware.Auth(http.HandlerFunc(handlers.RedeemReferral)))

	stack := middleware.Security(middleware.RealIP(middleware.Logger(mux)))
	server := &http.Server{
		Addr:         ":8080",
		Handler:      stack,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-quit
		log.Println("shutting down...")
		shutCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		server.Shutdown(shutCtx)
	}()

	log.Println("listening on :8080")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
	log.Println("server stopped")
}
