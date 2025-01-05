/*
# Software Name : SIP Load Balancer
# SPDX-FileCopyrightText: Copyright (c) 2025 - Orange Business - OINIS/Services/NSF

# Authors:
# - Moatassem Talaat <moatassem.talaat@orange.com>

---
*/

package redis

// package main

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Redis variables
var (
	// var ctx,x = context.WithTimeout(context.Background(), 5*time.Second)//ammar
	ctx            = context.Background() // Check out other contexts
	redisKeyExpiry = 15 * time.Minute     // redisKeyExpiry = 5 * time.Second
	client         *redis.Client          // Redis client
)

func SetupCheckRedis(raddr string, pswrd string, db int, expiryInMin int) (string, error) {
	client = redis.NewClient(&redis.Options{
		Addr:     raddr,
		Password: pswrd,
		DB:       db,
	})
	redisKeyExpiry = time.Duration(expiryInMin) * time.Minute
	_, err := client.Ping(ctx).Result()
	return client.Options().Addr, err
}

// TODO check if we can just reset the expiry timer on the key/value pair .. instead of setting it every 5mins
// ResetKeyExpiry resets the expiration time for a given key
func ResetKeyExpiry(fixed string, variable string) error {
	key := fmt.Sprintf("%s_%s", fixed, variable)
	exists, err := client.Exists(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("error checking key existence %s: %w", key, err)
	}
	if exists == 0 {
		return fmt.Errorf("key %s does not exist", key)
	}
	err = client.Expire(ctx, key, redisKeyExpiry).Err()
	if err != nil {
		return fmt.Errorf("error resetting expiry for key %s: %w", key, err)
	}
	return nil
}

// SetKeyWithExpiry sets a value in Redis with a composite key and a global expiry time
func setKeyWithAutoExpiry2(fixed string, variable string, value string) error {
	key := fmt.Sprintf("%s_%s", fixed, variable)
	err := client.Set(ctx, key, value, redisKeyExpiry).Err()
	if err != nil {
		return fmt.Errorf("error setting key %s: %w", key, err)
	}
	return nil
}

func SetKeyWithAutoExpiry(fixed string, variable string) error {
	var x uint8
	key := fmt.Sprintf("%s_%s", fixed, variable)
	err := client.Set(ctx, key, x, redisKeyExpiry).Err()
	if err != nil {
		return fmt.Errorf("error setting key %s: %w", key, err)
	}
	return nil
}

// GetValue returns the value for a given key
func getValue(key string) (string, error) {
	val, err := client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("key %s does not exist", key)
	} else if err != nil {
		return "", fmt.Errorf("error getting key %s: %w", key, err)
	}
	return val, nil
}

// CountKeysByPattern counts the number of keys that match a specific pattern
func countKeysByPattern2(pattern string) (int, error) {
	var count int
	var cursor uint64
	for {
		keys, cur, err := client.Scan(ctx, cursor, pattern, 0).Result() // check cursor usage
		if err != nil {
			return 0, fmt.Errorf("error scanning keys with pattern %s: %w", pattern, err)
		}
		count += len(keys)
		cursor = cur
		if cursor == 0 {
			break
		}
	}
	return count, nil
}

// CountKeysByPattern2 counts the number of keys that match a specific pattern
func CountKeysByPattern(fixed string) (uint, error) {
	var count uint
	pattern := fmt.Sprintf("%s_*", fixed)
	iter := client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		count++
	}
	if err := iter.Err(); err != nil {
		return count, err
	}
	return count, nil
}

func DeleteKey(fixed string, variable string) error {
	key := fmt.Sprintf("%s_%s", fixed, variable)
	err := client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("error deleting key %s: %w", key, err)
	}
	return nil
}

// func main() {
// 	fmt.Println(CheckRedisServer())
// 	// Example: Setting a key with composite name and value
// 	var wg sync.WaitGroup
// 	wg.Add(2)

// 	fxd := "scl_id_14423"
// 	var keys []string
// 	go func() {
// 		for i := 0; i < 100000; i++ {
// 			key := fmt.Sprintf("call_id_%d", i)
// 			keys = append(keys, key)
// 			err := SetKeyWithAutoExpiry(fxd, key)
// 			if err != nil {
// 				fmt.Println("Error:", err)
// 				return
// 			}
// 			<-time.After(20 * time.Millisecond)
// 		}
// 		fmt.Println("Keys set successfully.")
// 		wg.Done()
// 	}()

// 	// t := time.NewTimer(4 * time.Second)
// 	// <-t.C
// 	// for _, k := range keys {
// 	// 	ResetKeyExpiry("fromTag", k)
// 	// }

// 	// t.Reset(4200 * time.Millisecond)
// 	// <-t.C
// 	// for _, k := range keys {
// 	// 	ResetKeyExpiry("fromTag", k)
// 	// }

// 	// Example: Getting the value of a key
// 	// value, err := GetValue("CLname_fromTag")
// 	// if err != nil {
// 	// 	fmt.Println("Error:", err)
// 	// }
// 	// fmt.Printf("Value for key 'CLname_fromTag': %v\n", value)

// 	// Example: Counting keys matching a pattern
// 	go func() {
// 		for range 1000 {
// 			count, err := CountKeysByPattern(fxd)
// 			if err != nil {
// 				fmt.Println("Error:", err)
// 				return
// 			}
// 			fmt.Printf("Number of keys matching pattern: %d\n", count)
// 			<-time.After(300 * time.Millisecond)
// 		}
// 		wg.Done()
// 	}()

// 	wg.Wait()
// }
