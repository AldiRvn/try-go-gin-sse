// ? reference: https://medium.com/@akbarkusumanegaralth/stream-docker-log-with-server-sent-events-3f4fe19ca273
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
)

var AppPort = ":" + os.Getenv("APP_SSE_PORT")

func stream(c *gin.Context) {
	//? Set Timeout Stream
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	chanStream := make(chan string)
	done := make(chan bool)

	//? Handle Closing Stream
	go func() {
		for {
			select {
			case <-c.Request.Context().Done():
				done <- true
				return
			case <-ctx.Done():
				switch ctx.Err() {
				case context.DeadlineExceeded:
					log.Println("timeout")
				}
				done <- true
				return
			}
		}
	}()

	//? Writing Channel
	var mu sync.Mutex
	go func() {
		for {
			mu.Lock()

			layout := []string{time.Layout, time.ANSIC, time.UnixDate, time.RubyDate, time.RFC822, time.RFC822Z, time.RFC850, time.RFC1123, time.RFC1123Z, time.RFC3339, time.RFC3339Nano, time.Kitchen, time.Stamp, time.StampMilli, time.StampMicro, time.StampNano, time.DateTime, time.DateOnly, time.TimeOnly}
			chanStream <- time.Now().Format(layout[rand.Intn(len(layout))])

			mu.Unlock()
			time.Sleep(1 * time.Second)
		}
	}()

	//? Listening Channel
	isStreaming := c.Stream(func(w io.Writer) bool {
		for {
			select {
			case <-done:
				c.SSEvent("end", "end")
				return false
			case msg := <-chanStream:
				c.SSEvent("message", msg)
				return true
			}
		}
	})
	if !isStreaming {
		log.Println("closed")
	}
}

func main() {
	fmt.Printf("AppPort: %v\n", AppPort)

	router := gin.Default()

	router.StaticFile("/", "./src/public/index.html")
	router.GET("/stream", stream)

	log.Println("http://localhost" + AppPort)
	router.Run(AppPort)
}
