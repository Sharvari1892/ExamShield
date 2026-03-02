package main

import (
	"context"
	"log"
	"net/http"
	"time"
	"errors"
	"encoding/json"
	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/gin-contrib/cors"
	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/Sharvari1892/examshield/internal/realtime"
	"github.com/Sharvari1892/examshield/internal/config"
	"github.com/Sharvari1892/examshield/internal/middleware"
	"github.com/Sharvari1892/examshield/internal/repository"
	"github.com/Sharvari1892/examshield/internal/service"
	"github.com/Sharvari1892/examshield/internal/worker"
	"github.com/Sharvari1892/examshield/internal/logger"
	"github.com/Sharvari1892/examshield/internal/metrics"
)

func main() {
	logger.Init()
	defer logger.Sync()
	metrics.Init()

	cfg := config.Load()
	ctx := context.Background()
	router := gin.Default()
	router.Use(cors.New(cors.Config{
        AllowOrigins:     []string{"http://localhost:3000", "http://localhost:5173"},
        AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: true,
        MaxAge:           12 * time.Hour,
    }))

	// ----------------------------
	// Database
	// ----------------------------
	db, err := pgxpool.New(ctx, cfg.DBUrl)
	if err != nil {
		log.Fatal(err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr,
	})

	if err := db.Ping(ctx); err != nil {
		log.Fatal(err)
	}

	worker.StartIntegrityWorker(ctx, db, rdb)

	// ----------------------------
	// Services & Repositories
	// ----------------------------
	authService := service.NewAuthService(cfg.JWTSecret)
	userRepo := repository.NewUserRepository(db)
	examRepo := repository.NewExamRepository(db)
	auditRepo := repository.NewAuditRepository(db)

	router.Use(middleware.RequestID())
	router.Use(middleware.LoggingMiddleware())
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	hub := realtime.NewHub()
	realtime.StartSubscriber(ctx, rdb, hub)

	// ============================
	// HEALTH
	// ============================
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// ============================
	// AUTH
	// ============================

	router.POST("/register", func(c *gin.Context) {
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		hash, _ := authService.HashPassword(req.Password)
		user, err := userRepo.CreateUser(c, req.Email, hash, "student")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user exists"})
			return
		}
		c.JSON(http.StatusOK, user)
	})

	router.POST("/login", func(c *gin.Context) {
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		user, err := userRepo.GetByEmail(c, req.Email)
		if err != nil || authService.CheckPassword(user.PasswordHash, req.Password) != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}

		access, _ := authService.GenerateAccessToken(user.ID, user.Role)
		refresh, _ := authService.GenerateRefreshToken(user.ID)

		c.JSON(http.StatusOK, gin.H{
			"access_token":  access,
			"refresh_token": refresh,
		})
	})

	// ============================
	// ADMIN ROUTES
	// ============================

	admin := router.Group("/admin")
	admin.Use(
		middleware.AuthMiddleware(authService),
		middleware.AdminOnly(),
	)

	admin.POST("/exam", func(c *gin.Context) {
		var req struct {
			Title    string `json:"title"`
			Duration int    `json:"duration"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		exam, err := examRepo.CreateExam(c, req.Title, req.Duration)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
			return
		}
		c.JSON(http.StatusOK, exam)
	})

	admin.POST("/question", func(c *gin.Context) {
		var req struct {
			ExamID     string `json:"exam_id"`
			Difficulty int    `json:"difficulty"`
			Content    string `json:"content"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		q, err := examRepo.CreateQuestion(c, req.ExamID, req.Difficulty, req.Content)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
			return
		}
		c.JSON(http.StatusOK, q)
	})

	// ============================
	// START EXAM
	// ============================

	router.POST("/exam/start",
		middleware.AuthMiddleware(authService),
		func(c *gin.Context) {

			var req struct {
				ExamID string `json:"exam_id"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
				return
			}

			userID := c.GetString("user_id")
			lockKey := "active_session:" + userID

			ok, _ := rdb.SetNX(ctx, lockKey, req.ExamID, time.Hour).Result()
			if !ok {
				c.JSON(http.StatusConflict, gin.H{"error": "active session exists"})
				return
			}

			exam, _, err := examRepo.GetExamWithQuestions(c, req.ExamID)
			if err != nil {
				rdb.Del(ctx, lockKey)
				c.JSON(http.StatusNotFound, gin.H{"error": "exam not found"})
				return
			}

			start := time.Now().UTC()
			end := start.Add(time.Duration(exam.DurationSeconds) * time.Second)
			sessionID := uuid.New().String()

			_, err = db.Exec(ctx,
				`INSERT INTO exam_sessions
				(id,user_id,exam_id,start_time,end_time,status)
				VALUES($1,$2,$3,$4,$5,$6)`,
				sessionID, userID, req.ExamID, start, end, "active")

			if err != nil {
				rdb.Del(ctx, lockKey)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
				return
			}

			reqID := c.GetString("request_id")

			logger.Log.Info("exam session started",
				zap.String("request_id", reqID),
				zap.String("session_id", sessionID),
				zap.String("user_id", userID),
			)

			// Emit SESSION_STARTED
			auditRepo.CreateEvent(c, sessionID, "SESSION_STARTED",
				map[string]interface{}{
					"user_id": userID,
					"exam_id": req.ExamID,
				})

			c.JSON(http.StatusOK, gin.H{
				"session_id": sessionID,
				"start_time": start,
				"end_time":   end,
			})
		})

	// ============================
	// ANSWER UPDATE
	// ============================

	router.POST("/exam/:id/answer",
		middleware.AuthMiddleware(authService),
		func(c *gin.Context) {

			var req struct {
				SessionID  string `json:"session_id"`
				QuestionID string `json:"question_id"`
				Answer     string `json:"answer"`
				Version    int    `json:"version"`
			}

			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
				return
			}

			auditRepo.CreateEvent(c, req.SessionID, "ANSWER_UPDATED",
				map[string]interface{}{
					"question_id": req.QuestionID,
					"answer":      req.Answer,
					"version":     req.Version,
				})

			c.JSON(http.StatusOK, gin.H{"status": "recorded"})
		})

	//=============================
	//SUBMIT ANSWER
	//=============================
	router.POST("/exam/:id/resume",
	middleware.AuthMiddleware(authService),
	func(c *gin.Context) {

		var req struct {
			SessionID string `json:"session_id"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		// Increment grace_used
		_, err := db.Exec(ctx,
			`UPDATE exam_sessions
			 SET grace_used = grace_used + 1
			 WHERE id = $1`,
			req.SessionID,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update grace"})
			return
		}

		// 🔥 STEP 8 METRIC HERE
		metrics.GraceUsage.Inc()

		reqID := c.GetString("request_id")

		logger.Log.Info("session resumed",
			zap.String("request_id", reqID),
			zap.String("session_id", req.SessionID),
		)

		// Emit audit event
		auditRepo.CreateEvent(c, req.SessionID, "RESUMED",
			map[string]interface{}{
				"resumed_at": time.Now().UTC(),
			},
		)

		c.JSON(http.StatusOK, gin.H{"status": "resumed"})
	})	

	// ============================
	// SUBMIT EXAM
	// ============================

	router.POST("/exam/:id/submit",
		middleware.AuthMiddleware(authService),
		func(c *gin.Context) {

			var req struct {
				SessionID string `json:"session_id"`
			}

			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
				return
			}

			auditRepo.CreateEvent(c, req.SessionID, "SUBMITTED",
				map[string]interface{}{
					"submitted_at": time.Now().UTC(),
				})

			reqID := c.GetString("request_id")

			logger.Log.Info("exam submitted",
				zap.String("request_id", reqID),
				zap.String("session_id", req.SessionID),
			)

			c.JSON(http.StatusOK, gin.H{"status": "submitted"})
		})

	// ============================
	// SYNC ENDPOINT
	// ============================
		router.POST("/sync",
	middleware.AuthMiddleware(authService),
	func(c *gin.Context) {

		start := time.Now()
		defer func() {
			metrics.SyncLatency.Observe(time.Since(start).Seconds())
		}()

		var req struct {
			SessionID string `json:"session_id"`
			Answers   []struct {
				QuestionID string `json:"question_id"`
				Answer     string `json:"answer"`
				Version    int    `json:"version"`
			} `json:"answers"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			metrics.SyncFailures.Inc()
			c.JSON(400, gin.H{"error": "invalid request"})
			return
		}

		reqID := c.GetString("request_id")

		logger.Log.Info("sync request received",
			zap.String("request_id", reqID),
			zap.String("session_id", req.SessionID),
			zap.Int("answer_count", len(req.Answers)),
		)

		// ----------------------------------
		// Validate session active
		// ----------------------------------
		var status string
		err := db.QueryRow(ctx,
			`SELECT status FROM exam_sessions WHERE id=$1`,
			req.SessionID,
		).Scan(&status)

		if err != nil || status != "active" {
			metrics.SyncFailures.Inc()
			c.JSON(400, gin.H{"error": "invalid or inactive session"})
			return
		}

		accepted := []string{}
		rejected := []string{}
		serverVersions := map[string]int{}

		// ----------------------------------
		// Process each answer
		// ----------------------------------
		for _, ans := range req.Answers {

			var serverVersion int

			err := db.QueryRow(ctx,
				`SELECT version FROM answers
				 WHERE session_id=$1 AND question_id=$2`,
				req.SessionID,
				ans.QuestionID,
			).Scan(&serverVersion)

			// ==================================
			// CASE 1: No existing answer → INSERT
			// ==================================
			if errors.Is(err, pgx.ErrNoRows) {

				answerJSON, marshalErr := json.Marshal(ans.Answer)
				if marshalErr != nil {
					rejected = append(rejected, ans.QuestionID)
					continue
				}

				_, insertErr := db.Exec(ctx,
					`INSERT INTO answers
					 (id, session_id, question_id, answer_data, version, updated_at)
					 VALUES ($1,$2,$3,$4,$5,$6)`,
					uuid.New().String(),
					req.SessionID,
					ans.QuestionID,
					answerJSON,
					ans.Version,
					time.Now().UTC(),
				)

				if insertErr != nil {
				logger.Log.Error("failed to insert answer",
				zap.String("session_id", req.SessionID),
				zap.String("question_id", ans.QuestionID),
				zap.Int("version", ans.Version),
				zap.Error(insertErr),
				)

				rejected = append(rejected, ans.QuestionID)
				metrics.SyncFailures.Inc()
				continue
			}

				accepted = append(accepted, ans.QuestionID)
				serverVersions[ans.QuestionID] = ans.Version
				continue
			}

			// ==================================
			// CASE 2: Real DB error
			// ==================================
			if err != nil {
				logger.Log.Error("failed to fetch answer version",
				zap.String("session_id", req.SessionID),
				zap.String("question_id", ans.QuestionID),
				zap.Error(err),
			)

	rejected = append(rejected, ans.QuestionID)
	metrics.SyncFailures.Inc()
	continue
}

			// ==================================
			// CASE 3: Existing row → Version check
			// ==================================
			if ans.Version > serverVersion {

				answerJSON, marshalErr := json.Marshal(ans.Answer)
				if marshalErr != nil {
					rejected = append(rejected, ans.QuestionID)
					continue
				}

				newVersion := ans.Version

				_, updateErr := db.Exec(ctx,
					`UPDATE answers
					 SET answer_data=$1,
					     version=$2,
					     updated_at=$3
					 WHERE session_id=$4 AND question_id=$5`,
					answerJSON,
					newVersion,
					time.Now().UTC(),
					req.SessionID,
					ans.QuestionID,
				)

				if updateErr != nil {

					reqID := c.GetString("request_id")

					logger.Log.Error("failed to update answer",
						zap.String("request_id", reqID),
						zap.String("session_id", req.SessionID),
						zap.String("question_id", ans.QuestionID),
						zap.Int("client_version", ans.Version),
						zap.Error(updateErr),
					)

					rejected = append(rejected, ans.QuestionID)
					metrics.SyncFailures.Inc()
					continue
				}

				accepted = append(accepted, ans.QuestionID)
				serverVersions[ans.QuestionID] = newVersion

			} else {
				// Version conflict
				rejected = append(rejected, ans.QuestionID)
				serverVersions[ans.QuestionID] = serverVersion
			}
		}

		logger.Log.Info("sync completed",
			zap.String("request_id", reqID),
			zap.String("session_id", req.SessionID),
			zap.Int("accepted", len(accepted)),
			zap.Int("rejected", len(rejected)),
		)

		c.JSON(200, gin.H{
			"accepted":        accepted,
			"rejected":        rejected,
			"server_versions": serverVersions,
		})
	})

		// ============================
// VERIFY TAMPERING (ADMIN)
// ============================

admin.GET("/verify/:session_id", func(c *gin.Context) {

	sessionID := c.Param("session_id")

	events, err := auditRepo.GetEventsBySession(c, sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch events"})
		return
	}

	err = service.VerifyChain(events)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "CHAIN BROKEN",
			"error":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "CHAIN VALID"})
})

//========================
//Hub
//========================
upgrader := websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

router.GET("/ws",
	middleware.AuthMiddleware(authService),
	func(c *gin.Context) {

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}

		role := c.GetString("role")
		userID := c.GetString("user_id")

		client := &realtime.Client{
			Conn:   conn,
			Role:   role,
			UserID: userID,
		}

		hub.Register(client)

		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				hub.Unregister(client)
				conn.Close()
				break
			}
		}
	})

	logger.Log.Info("server started",
	zap.String("port", cfg.Port),
	)
	router.Run(":" + cfg.Port)
}
