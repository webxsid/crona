package api

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Subscribe connects to the kernel SSE stream and sends each parsed KernelEvent
// to the returned channel. The goroutine reconnects on disconnect.
// Close the done channel to stop.
func Subscribe(baseURL, token string, done <-chan struct{}) <-chan KernelEvent {
	ch := make(chan KernelEvent, 32)

	go func() {
		defer close(ch)
		for {
			select {
			case <-done:
				return
			default:
			}

			err := readStream(baseURL, token, ch, done)
			if err != nil {
				// brief backoff before reconnect
				select {
				case <-done:
					return
				case <-time.After(2 * time.Second):
				}
			}
		}
	}()

	return ch
}

func readStream(baseURL, token string, ch chan<- KernelEvent, done <-chan struct{}) error {
	req, err := http.NewRequest(http.MethodGet, baseURL+"/events", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "text/event-stream")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("SSE status %d", resp.StatusCode)
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		select {
		case <-done:
			return nil
		default:
		}

		line := scanner.Text()
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "" {
			continue
		}

		var event KernelEvent
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			continue
		}

		select {
		case ch <- event:
		case <-done:
			return nil
		}
	}

	return scanner.Err()
}
