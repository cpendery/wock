package pipe

import (
	"context"
	"fmt"
	"os"
	"time"
)

const (
	timeout = 1 * time.Second
)

func waitForFile(filePath string, ctx context.Context) error {
	sleepChan := make(chan struct{})
	for {
		if _, err := os.Stat(filePath); err == nil {
			return nil
		}
		go func() {
			time.Sleep(time.Millisecond * 50)
			sleepChan <- struct{}{}
		}()
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout: file '%s' doesn't exist", filePath)
		case <-sleepChan:
		}
	}
}
