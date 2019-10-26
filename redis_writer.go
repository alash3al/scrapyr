package main

import "io"

// RedisWriter an io.Writer based on redis pub/sub
type RedisWriter struct {
	ch string
}

// NewRedisWriter initiate a new redis writer
func NewRedisWriter(ch string) io.Writer {
	return RedisWriter{ch: ch}
}

// Write write to the redis channel
func (rw RedisWriter) Write(b []byte) (n int, err error) {
	_, err = redisConn.Publish(rw.ch, string(b)).Result()
	n = len(b)

	return
}
