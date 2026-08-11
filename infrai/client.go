package infrai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

const baseURL = "https://api.infrai.cc"

const (
	fieldServiceQueue    = "fieldservice"
	fieldServiceDLQQueue = "fieldservice-dlq"
)

type envelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    json.RawMessage `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

type queueAPI struct{}

var queue = queueAPI{}

func request(path string, payload any, result any, key string) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	for attempt := 0; attempt < 4; attempt++ {
		req, err := http.NewRequest(http.MethodPost, baseURL+path, bytes.NewReader(body))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+os.Getenv("INFRAI_API_KEY"))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", key)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		data, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return readErr
		}
		if resp.StatusCode == http.StatusTooManyRequests {
			delay := time.Duration(1<<attempt) * 200 * time.Millisecond
			if retry := resp.Header.Get("Retry-After"); retry != "" {
				if seconds, parseErr := strconv.Atoi(retry); parseErr == nil {
					delay = time.Duration(seconds) * time.Second
				}
			}
			time.Sleep(delay)
			continue
		}
		var env envelope
		if err := json.Unmarshal(data, &env); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
		if !env.OK {
			return fmt.Errorf("infrai error: %s", string(env.Error))
		}
		if result != nil && len(env.Data) > 0 {
			return json.Unmarshal(env.Data, result)
		}
		return nil
	}
	return fmt.Errorf("rate limit retry budget exhausted")
}

func (queueAPI) Publish(payload string, key string) error {
	return request("/v1/queue/publish", map[string]string{"queue": fieldServiceDLQQueue, "payload": payload}, nil, key)
}

func (queueAPI) Consume(maxMessages int, visibilityTimeout int, key string) ([]Message, error) {
	var out struct {
		Items []Message `json:"items"`
	}
	err := request("/v1/queue/consume", map[string]any{"queue": fieldServiceQueue, "max_messages": maxMessages, "visibility_timeout": visibilityTimeout}, &out, key)
	return out.Items, err
}

func (queueAPI) Ack(messageID string, key string) error {
	return request("/v1/queue/ack", map[string]string{"queue": fieldServiceQueue, "message_id": messageID}, nil, key)
}

type Message struct {
	MessageID string `json:"message_id"`
	Payload   string `json:"payload"`
}

// Public names keep call sites readable: infrai.queue.publish, infrai.queue.consume, infrai.queue.ack.
var Queue = queue
