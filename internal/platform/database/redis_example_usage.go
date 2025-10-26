package database

// File này chứa các ví dụ sử dụng RedisClient
// KHÔNG build file này - chỉ dùng để tham khảo

/*
import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// ===============================================
// Example 1: Basic String Operations
// ===============================================

func ExampleStringOperations(rdb *RedisClient) error {
	ctx := context.Background()

	// Set key-value với expiration
	err := rdb.Set(ctx, "user:email:123", "user@example.com", 1*time.Hour)
	if err != nil {
		return err
	}

	// Get value
	email, err := rdb.Get(ctx, "user:email:123")
	if err == redis.Nil {
		// Key không tồn tại
		return nil
	} else if err != nil {
		return err
	}

	fmt.Println("Email:", email) // Output: user@example.com

	return nil
}

// ===============================================
// Example 2: Caching với JSON
// ===============================================

type User struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func ExampleCachingUser(rdb *RedisClient) error {
	ctx := context.Background()

	// Serialize user to JSON
	user := User{
		ID:    123,
		Name:  "John Doe",
		Email: "john@example.com",
	}

	userJSON, err := json.Marshal(user)
	if err != nil {
		return err
	}

	// Cache user data trong 30 phút
	key := fmt.Sprintf("user:cache:%d", user.ID)
	err = rdb.Set(ctx, key, userJSON, 30*time.Minute)
	if err != nil {
		return err
	}

	// Retrieve từ cache
	cachedJSON, err := rdb.Get(ctx, key)
	if err != nil {
		return err
	}

	var cachedUser User
	if err := json.Unmarshal([]byte(cachedJSON), &cachedUser); err != nil {
		return err
	}

	fmt.Printf("Cached User: %+v\n", cachedUser)

	return nil
}

// ===============================================
// Example 3: Hash Operations (User Profile)
// ===============================================

func ExampleHashOperations(rdb *RedisClient) error {
	ctx := context.Background()

	// Set multiple fields trong hash
	err := rdb.HSet(ctx, "user:profile:123",
		"name", "John Doe",
		"email", "john@example.com",
		"age", "30",
	)
	if err != nil {
		return err
	}

	// Get single field
	name, err := rdb.HGet(ctx, "user:profile:123", "name")
	if err != nil {
		return err
	}
	fmt.Println("Name:", name)

	// Get all fields
	profile, err := rdb.HGetAll(ctx, "user:profile:123")
	if err != nil {
		return err
	}
	fmt.Printf("Profile: %+v\n", profile)

	return nil
}

// ===============================================
// Example 4: Counter Operations
// ===============================================

func ExampleCounterOperations(rdb *RedisClient) error {
	ctx := context.Background()

	// Increment page views
	views, err := rdb.Incr(ctx, "page:views:homepage")
	if err != nil {
		return err
	}
	fmt.Println("Total views:", views)

	// Decrement stock
	stock, err := rdb.Decr(ctx, "product:stock:456")
	if err != nil {
		return err
	}
	fmt.Println("Remaining stock:", stock)

	return nil
}

// ===============================================
// Example 5: Key Expiration và TTL
// ===============================================

func ExampleExpirationAndTTL(rdb *RedisClient) error {
	ctx := context.Background()

	// Set key without expiration
	err := rdb.Set(ctx, "session:temp", "data", 0)
	if err != nil {
		return err
	}

	// Set expiration sau đó
	err = rdb.Expire(ctx, "session:temp", 5*time.Minute)
	if err != nil {
		return err
	}

	// Check TTL
	ttl, err := rdb.TTL(ctx, "session:temp")
	if err != nil {
		return err
	}
	fmt.Println("TTL:", ttl) // Output: ~5m0s

	return nil
}

// ===============================================
// Example 6: Check Key Existence
// ===============================================

func ExampleCheckExistence(rdb *RedisClient) error {
	ctx := context.Background()

	// Check if key exists
	exists, err := rdb.Exists(ctx, "user:123")
	if err != nil {
		return err
	}

	if exists {
		// Key tồn tại, lấy data
		data, _ := rdb.Get(ctx, "user:123")
		fmt.Println("User data:", data)
	} else {
		// Key không tồn tại, fetch từ DB
		fmt.Println("User not in cache, fetch from DB")
	}

	return nil
}

// ===============================================
// Example 7: Delete Keys
// ===============================================

func ExampleDeleteKeys(rdb *RedisClient) error {
	ctx := context.Background()

	// Delete single key
	err := rdb.Del(ctx, "temp:key")
	if err != nil {
		return err
	}

	// Delete multiple keys
	err = rdb.Del(ctx, "temp:key1", "temp:key2", "temp:key3")
	if err != nil {
		return err
	}

	return nil
}

// ===============================================
// Example 8: Session Management
// ===============================================

type Session struct {
	UserID    int64     `json:"user_id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

func ExampleSessionManagement(rdb *RedisClient) error {
	ctx := context.Background()

	sessionID := "sess_abc123xyz"

	// Create session
	session := Session{
		UserID:    123,
		Email:     "user@example.com",
		CreatedAt: time.Now(),
	}

	sessionJSON, _ := json.Marshal(session)
	err := rdb.Set(ctx, "session:"+sessionID, sessionJSON, 24*time.Hour)
	if err != nil {
		return err
	}

	// Retrieve session
	sessionData, err := rdb.Get(ctx, "session:"+sessionID)
	if err == redis.Nil {
		// Session expired hoặc không tồn tại
		return fmt.Errorf("session not found")
	} else if err != nil {
		return err
	}

	var retrievedSession Session
	json.Unmarshal([]byte(sessionData), &retrievedSession)
	fmt.Printf("Session: %+v\n", retrievedSession)

	// Invalidate session (logout)
	err = rdb.Del(ctx, "session:"+sessionID)
	if err != nil {
		return err
	}

	return nil
}

// ===============================================
// Example 9: Rate Limiting
// ===============================================

func ExampleRateLimiting(rdb *RedisClient) (bool, error) {
	ctx := context.Background()

	userID := "user:123"
	key := fmt.Sprintf("rate_limit:%s", userID)

	// Increment request count
	count, err := rdb.Incr(ctx, key)
	if err != nil {
		return false, err
	}

	// Set expiration on first request
	if count == 1 {
		err = rdb.Expire(ctx, key, 1*time.Minute)
		if err != nil {
			return false, err
		}
	}

	// Check limit (e.g., 100 requests per minute)
	if count > 100 {
		return false, fmt.Errorf("rate limit exceeded")
	}

	return true, nil
}

// ===============================================
// Example 10: Cache-Aside Pattern
// ===============================================

func ExampleCacheAsidePattern(rdb *RedisClient) (*User, error) {
	ctx := context.Background()
	userID := int64(123)
	cacheKey := fmt.Sprintf("user:%d", userID)

	// 1. Try to get from cache first
	cachedData, err := rdb.Get(ctx, cacheKey)
	if err == nil {
		// Cache hit
		var user User
		if err := json.Unmarshal([]byte(cachedData), &user); err == nil {
			fmt.Println("Cache HIT")
			return &user, nil
		}
	}

	// 2. Cache miss - fetch from database
	fmt.Println("Cache MISS - fetching from DB")
	user, err := fetchUserFromDB(userID)
	if err != nil {
		return nil, err
	}

	// 3. Store in cache for next time
	userJSON, _ := json.Marshal(user)
	rdb.Set(ctx, cacheKey, userJSON, 1*time.Hour)

	return user, nil
}

func fetchUserFromDB(userID int64) (*User, error) {
	// Simulate DB fetch
	return &User{
		ID:    userID,
		Name:  "John Doe",
		Email: "john@example.com",
	}, nil
}

// ===============================================
// Example 11: Using Direct Redis Client for Advanced Operations
// ===============================================

func ExampleAdvancedOperations(rdb *RedisClient) error {
	ctx := context.Background()

	// List operations
	err := rdb.Client.RPush(ctx, "queue:jobs", "job1", "job2", "job3").Err()
	if err != nil {
		return err
	}

	job, err := rdb.Client.LPop(ctx, "queue:jobs").Result()
	if err != nil {
		return err
	}
	fmt.Println("Processing job:", job)

	// Set operations
	err = rdb.Client.SAdd(ctx, "tags:product:123", "electronics", "laptop", "gaming").Err()
	if err != nil {
		return err
	}

	isMember, err := rdb.Client.SIsMember(ctx, "tags:product:123", "laptop").Result()
	if err != nil {
		return err
	}
	fmt.Println("Is laptop a tag:", isMember)

	// Sorted Set operations
	err = rdb.Client.ZAdd(ctx, "leaderboard", redis.Z{Score: 100, Member: "player1"}).Err()
	if err != nil {
		return err
	}

	return nil
}

// ===============================================
// Example 12: Pipeline for Batch Operations
// ===============================================

func ExamplePipeline(rdb *RedisClient) error {
	ctx := context.Background()

	// Pipeline giúp gửi nhiều commands cùng lúc
	pipe := rdb.Client.Pipeline()

	// Add multiple commands to pipeline
	pipe.Set(ctx, "key1", "value1", 0)
	pipe.Set(ctx, "key2", "value2", 0)
	pipe.Set(ctx, "key3", "value3", 0)
	pipe.Incr(ctx, "counter")

	// Execute all commands at once
	_, err := pipe.Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

// ===============================================
// Example 13: Transaction (WATCH/MULTI/EXEC)
// ===============================================

func ExampleTransaction(rdb *RedisClient) error {
	ctx := context.Background()

	// Transaction với optimistic locking
	err := rdb.Client.Watch(ctx, func(tx *redis.Tx) error {
		// Get current value
		val, err := tx.Get(ctx, "balance:user:123").Int()
		if err != nil && err != redis.Nil {
			return err
		}

		// Calculate new value
		newVal := val + 100

		// Execute transaction
		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.Set(ctx, "balance:user:123", newVal, 0)
			return nil
		})

		return err
	}, "balance:user:123")

	return err
}

// ===============================================
// Example 14: Pub/Sub (Publisher/Subscriber)
// ===============================================

func ExamplePubSub(rdb *RedisClient) error {
	ctx := context.Background()

	// Publish message
	err := rdb.Client.Publish(ctx, "notifications", "New user registered").Err()
	if err != nil {
		return err
	}

	// Subscribe (usually runs in goroutine)
	go func() {
		pubsub := rdb.Client.Subscribe(ctx, "notifications")
		defer pubsub.Close()

		ch := pubsub.Channel()
		for msg := range ch {
			fmt.Printf("Received: %s from channel: %s\n", msg.Payload, msg.Channel)
		}
	}()

	return nil
}
*/
