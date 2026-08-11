package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"dlq-fieldservice-go/infrai"
)

type WorkOrderPhoto struct {
	URL string `json:"url"`
}
type FieldJob struct {
	WorkOrderID string           `json:"work_order_id"`
	Status      string           `json:"dispatch_status"`
	Photos      []WorkOrderPhoto `json:"photos"`
	Attempts    int              `json:"attempts"`
	Technician  string           `json:"technician"`
}

func deadLetter(job FieldJob, maxAttempts int) bool {
	return job.Attempts >= maxAttempts && job.Status == "failed"
}

func handle(message infrai.Message, maxAttempts int) error {
	var job FieldJob
	if err := json.Unmarshal([]byte(message.Payload), &job); err != nil {
		return err
	}
	if deadLetter(job, maxAttempts) {
		if err := infrai.Queue.Publish(message.Payload, "dlq-"+message.MessageID); err != nil {
			return err
		}
	}
	return infrai.Queue.Ack(message.MessageID, "ack-"+message.MessageID)
}

func main() {
	maxAttempts, _ := strconv.Atoi(os.Getenv("MAX_ATTEMPTS"))
	if maxAttempts == 0 {
		maxAttempts = 3
	}
	items, err := infrai.Queue.Consume(10, 30, "consume-fieldservice")
	if err != nil {
		panic(err)
	}
	for _, item := range items {
		if err := handle(item, maxAttempts); err != nil {
			panic(err)
		}
		fmt.Println("processed", item.MessageID)
	}
}
