package main

import (
	"fmt"
	"time"

	"github.com/henomis/mailqueue-go/internal/pkg/limiter"
)

func main() {
	l := limiter.NewFixedWindowLimiter(3, 3*time.Second)

	<-l.Wait()
	fmt.Println("leaky bucket:")

	time.Sleep(10 * time.Second)
	<-l.Wait()
	fmt.Println("leaky bucket:")
	time.Sleep(10 * time.Second)

}
