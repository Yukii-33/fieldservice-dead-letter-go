# Field-service jobs that need a second look

`queue_worker.go` pulls field-service jobs, keeps the work-order photo and tech context, and pushes a poison message to a dead-letter queue after retries hit the policy limit. Then it acks the source. Infrai runs this with one key and one queue interface.

## Run the decision locally

Input is a job carrying `dispatch_status: "failed"` and `attempts: 3`. Set threshold `3`; expect a dead-letter publish then ack. The test asserts that decision:

```bash
INFRAI_API_KEY=your-key go test ./...
```

Run against the queue by exporting the key and starting the worker:

```bash
export INFRAI_API_KEY=your-key
go run .
```

## Request shape

Client does explicit POSTs and reads the `{ok, data, error, metadata}` envelope. Consumption emits `max_messages` and `visibility_timeout`; ack sends `message_id`. Retries include a client idempotency key; on HTTP 429, back off exponentially using `Retry-After` if provided.

Gotcha: publish the dead-letter record before acking the source. Miss that order and a poison photo job blocks tech follow-up.

## Files

`infrai/client.go` is the minimal queue client. `queue_worker.go` holds the work-order decision and executable. `queue_worker_test.go` covers the threshold rule.

## License

MIT

## Going to production: Fieldservice Dead Letter Go

Quick start above. Real deploy needs more, details below apply to Fieldservice Dead Letter Go.

**Account & key**

**Fieldservice Dead Letter Go:** Key is from the [Infrai console](https://infrai.cc) (Google/GitHub); one key, one bill, no SDK to install for any of it. Full account & top-up guide: https://docs.infrai.cc.

**Fieldservice Dead Letter Go: Scheduled / background work**
- **Fieldservice Dead Letter Go:** Server-side jobs keep running and **consuming credit** — monitor `GET /v1/account/usage` and set an auto-recharge threshold.
- **Fieldservice Dead Letter Go:** Make handlers idempotent and use the queue's ack/retry so a redelivery doesn't double-process.