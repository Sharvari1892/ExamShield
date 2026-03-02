package realtime

import (
	"context"
	"encoding/json"
	"go.uber.org/zap"

	"github.com/redis/go-redis/v9"
	"github.com/Sharvari1892/examshield/internal/domain"
	"github.com/Sharvari1892/examshield/internal/logger"
)

func StartSubscriber(ctx context.Context, rdb *redis.Client, hub *Hub) {

	pubsub := rdb.Subscribe(ctx, "integrity_alerts")

	go func() {
		ch := pubsub.Channel()

		for msg := range ch {

			var alert domain.IntegrityAlert
			if err := json.Unmarshal([]byte(msg.Payload), &alert); err != nil {
				continue
			}

			hub.Broadcast(alert)
			logger.Log.Info("alert broadcasted",
				zap.String("session_id", alert.SessionID),
				zap.Int("score", alert.Score),
				zap.Strings("flags", alert.Flags),
			)
		}
	}()
}