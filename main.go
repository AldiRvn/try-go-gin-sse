// ? reference: https://medium.com/@akbarkusumanegaralth/stream-docker-log-with-server-sent-events-3f4fe19ca273
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
	"github.com/logrusorgru/aurora"
)

var AppPort = ":" + os.Getenv("APP_SSE_PORT")

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func stream(c *gin.Context) {
	target := c.Param("target")
	if _, found := listReceiverActive[target]; !found { //? Register Active
		listReceiverActive[target] = listReceiver[target]
	}

	//? Set Timeout Stream
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	//? Handle Closing Stream
	done := make(chan bool)
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

	//? Listening Channel
	isStreaming := c.Stream(func(w io.Writer) bool {
		for {
			select {
			case <-done:
				c.SSEvent("end", "end")
				return false
			case msg := <-listReceiverActive[target]:
				log.Println(aurora.BgGreen("-> New msg"), msg)
				c.SSEvent("message", msg)
				return true
			}
		}
	})
	if !isStreaming {
		<-listReceiverActive[target]       //? Throw Away Active Channel
		delete(listReceiverActive, target) //? Delete From Active
		log.Println(aurora.BgRed("CLOSED"))
	}
}

var listReceiver = map[string]chan string{
	"Marcus Silva":  make(chan string),
	"Matilda Logan": make(chan string),
	"Lenora Pope":   make(chan string),
	"Theodore King": make(chan string),
	"Ralph Bailey":  make(chan string),
}
var listReceiverActive = map[string]chan string{}

func getListReceiverKey() (res []string) {
	for key := range listReceiver {
		res = append(res, key)
	}
	return
}

func main() {
	log.Println("NumGoroutine", runtime.NumGoroutine())
	fmt.Printf("AppPort: %v\n", AppPort)

	router := gin.Default()

	router.StaticFile("/", "./src/public/index.html")
	router.GET("/target", func(ctx *gin.Context) { ctx.JSON(http.StatusOK, map[string]any{"data": getListReceiverKey()}) })
	router.GET("/stream/:target", stream)

	// ? Writing Channel
	go func() {
		log.Println("NumGoroutine", runtime.NumGoroutine())

		for {
			time.Sleep(500 * time.Millisecond)
			for key := range listReceiverActive {
				log.Println(aurora.BgYellow("<- Send msg to"), key)
				listReceiverActive[key] <- fmt.Sprint(key)

				time.Sleep(1 * time.Second)
				log.Println("NumGoroutine", runtime.NumGoroutine())
			}
		}
	}()

	log.Println("http://localhost" + AppPort)
	router.Run(AppPort)
}
