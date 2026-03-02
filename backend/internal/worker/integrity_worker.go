package worker

import (
	"context"
	"encoding/json"
	"time"

	"go.uber.org/zap"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/Sharvari1892/examshield/internal/logger"
	"github.com/Sharvari1892/examshield/internal/metrics"
	"github.com/Sharvari1892/examshield/internal/domain"
)

func StartIntegrityWorker(ctx context.Context, db *pgxpool.Pool, rdb *redis.Client) {

	ticker := time.NewTicker(15 * time.Second)

	go func() {
		for {
			select {
			case <-ticker.C:
				runIntegrityCheck(ctx, db, rdb)
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}

func runIntegrityCheck(ctx context.Context, db *pgxpool.Pool, rdb *redis.Client) {

	var activeCount int
	db.QueryRow(ctx,
		`SELECT COUNT(*) FROM exam_sessions WHERE status='active'`,
	).Scan(&activeCount)

metrics.ActiveSessions.Set(float64(activeCount))

	rows, err := db.Query(ctx,
		`SELECT id, start_time, end_time
		 FROM exam_sessions
		 WHERE status='active'`,
	)
	if err != nil {
		logger.Log.Error("worker failed to query active sessions",
			zap.Error(err),
		)
		return
	}
	defer rows.Close()

	for rows.Next() {

		var sessionID string
		var start, end time.Time

		if err := rows.Scan(&sessionID, &start, &end); err != nil {
			continue
		}

		score := 100
		flags := []string{}

		// ----------------------------------
		// Exam time exceeded
		// ----------------------------------
		if time.Now().UTC().After(end) {
			flags = append(flags, "EXAM_TIME_EXCEEDED")
			score -= 20
		}

		// ----------------------------------
		// Fast answering rate
		// ----------------------------------
		var answerCount int
		db.QueryRow(ctx,
			`SELECT COUNT(*) FROM answers WHERE session_id=$1`,
			sessionID,
		).Scan(&answerCount)

		duration := time.Since(start).Seconds()

		if duration > 0 {
			rate := float64(answerCount) / duration
			if rate > 1.5 {
				flags = append(flags, "ABNORMALLY_FAST_ANSWERING")
				score -= 30
			}
		}

		// ----------------------------------
		// Too Many Resumes
		// ----------------------------------
		var resumeCount int
		db.QueryRow(ctx,
			`SELECT COUNT(*) FROM audit_events
			 WHERE session_id=$1 AND event_type='RESUMED'`,
			sessionID,
		).Scan(&resumeCount)

		if resumeCount > 3 {
			flags = append(flags, "TOO_MANY_RESUMES")
			score -= 15
		}

		// ----------------------------------
		// Grace Usage
		// ----------------------------------
		var graceUsed int
		db.QueryRow(ctx,
			`SELECT grace_used FROM exam_sessions WHERE id=$1`,
			sessionID,
		).Scan(&graceUsed)

		if graceUsed > 2 {
			flags = append(flags, "EXCESSIVE_GRACE_USAGE")
			score -= 15
		}

		// ----------------------------------
		// IP Switching
		// ----------------------------------
		var distinctIPs int
		db.QueryRow(ctx,
			`SELECT COUNT(DISTINCT payload->>'ip')
			 FROM audit_events
			 WHERE session_id=$1 AND payload ? 'ip'`,
			sessionID,
		).Scan(&distinctIPs)

		if distinctIPs > 2 {
			flags = append(flags, "IP_SWITCHING_DETECTED")
			score -= 25
		}

		// ----------------------------------
		// Device Fingerprint Change
		// ----------------------------------
		var distinctDevices int
		db.QueryRow(ctx,
			`SELECT COUNT(DISTINCT payload->>'fingerprint')
			 FROM audit_events
			 WHERE session_id=$1 AND payload ? 'fingerprint'`,
			sessionID,
		).Scan(&distinctDevices)

		if distinctDevices > 1 {
			flags = append(flags, "DEVICE_FINGERPRINT_CHANGED")
			score -= 30
		}

		// ----------------------------------
		// Automation Pattern Detection
		// ----------------------------------
		rows2, err := db.Query(ctx,
			`SELECT updated_at
			 FROM answers
			 WHERE session_id=$1
			 ORDER BY updated_at`,
			sessionID,
		)

		if err == nil {
			var prev time.Time
			consistentCount := 0

			for rows2.Next() {
				var ts time.Time
				rows2.Scan(&ts)

				if !prev.IsZero() {
					diff := ts.Sub(prev).Seconds()
					if diff >= 1.9 && diff <= 2.1 {
						consistentCount++
					}
				}
				prev = ts
			}
			rows2.Close()

			if consistentCount >= 20 {
				flags = append(flags, "AUTOMATION_PATTERN_DETECTED")
				score -= 50
			}
		}

		if score < 0 {
			score = 0
		}

		// ----------------------------------
		// Update DB
		// ----------------------------------
		flagsJSON, _ := json.Marshal(flags)

		db.Exec(ctx,
			`UPDATE exam_sessions
			 SET integrity_score=$1,
			     flags=$2
			 WHERE id=$3`,
			score,
			flagsJSON,
			sessionID,
		)

		// ----------------------------------
		// Publish Real-Time Alert
		// ----------------------------------
		if score < 60 {

			alert := domain.IntegrityAlert{
				SessionID: sessionID,
				Score:     score,
				Flags:     flags,
				Type:      "RISK_ALERT",
			}

			data, _ := json.Marshal(alert)

			rdb.Publish(ctx, "integrity_alerts", data)

			logger.Log.Info("alert published",
				zap.String("session_id", sessionID),
			)
		}
	}
}