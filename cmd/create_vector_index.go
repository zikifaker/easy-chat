package main

import (
	"easy-chat/dao"
	"github.com/redis/go-redis/v9"
	"golang.org/x/net/context"
	"log"
)

const embeddingDIM = 1536

func main() {
	ctx := context.Background()
	cmd := redis.NewCmd(
		ctx,
		"FT.CREATE",
		"qa_index",
		"ON", "HASH",
		"PREFIX", 1, "qa:",
		"SCHEMA",
		"query", "TEXT",
		"answer", "TEXT",
		"embedding", "VECTOR", "HNSW", 6, "TYPE", "FLOAT32", "DIM", embeddingDIM, "DISTANCE_METRIC", "COSINE",
	)

	client := dao.GetCacheClient()
	if err := client.Process(ctx, cmd); err != nil {
		log.Printf("failed to create vector index: %v", err)
	}
}
