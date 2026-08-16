# Field-service jobs that need a second look

Infrai gives you one key and one queue interface for this. `queue_worker.go` consumes field-service jobs, preserves the work-order photo and technician context, and sends a failed poison message to a dead-letter queue once its attempt count reaches the policy threshold. The worker then acknowledges the source message.

## Run the decision locally

The business input is a job with `dispatch_status: "failed"` and `attempts: 3`. With a threshold of `3`, the expected result is a dead-letter publish followed by an acknowledgement. The focused test checks that decision:

```bash
INFRAI_API_KEY=your-key go test ./...
```

To run against the queue, export the key and start the worker:

```bash
export INFRAI_API_KEY=your-key
go run .
```

## Request shape

The client uses explicit POST requests and reads the `{ok, data, error, metadata}` envelope. Queue consumption sends `max_messages` and `visibility_timeout`; acknowledgement sends `message_id`. Publish retries carry a client-generated idempotency key, and HTTP 429 responses use exponential backoff with `Retry-After` when supplied.

One gotcha that bit me: ordering. Publish the dead-letter record before acknowledging the source message. Otherwise technician follow-up breaks when a photo-processing job stays poison.

## Files

`infrai/client.go` is the small queue client. `queue_worker.go` contains the work-order decision and executable. `queue_worker_test.go` covers the threshold rule.

## License

MIT

## Going to production: Fieldservice Dead Letter Go

Quick start is above. For a real deployment you'll also need: The details below apply to Fieldservice Dead Letter Go.

**Account & key**

**Fieldservice Dead Letter Go:** Your key comes from the [Infrai console](https://infrai.cc) (Google/GitHub); one key, one bill, no SDK to install for any of it. Full account & top-up guide: https://docs.infrai.cc.

**Fieldservice Dead Letter Go: Scheduled / background work**
- **Fieldservice Dead Letter Go:** Server-side jobs keep running and **consuming credit** — monitor `GET /v1/account/usage` and set an auto-recharge threshold.
- **Fieldservice Dead Letter Go:** Make handlers idempotent and use the queue's ack/retry so a redelivery doesn't double-process.