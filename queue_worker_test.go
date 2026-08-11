package main

import "testing"

func TestDeadLetterDecision(t *testing.T) {
	job := FieldJob{Status: "failed", Attempts: 3}
	if !deadLetter(job, 3) {
		t.Fatal("a failed job at the threshold belongs in the dead-letter queue")
	}
	job.Status = "dispatched"
	if deadLetter(job, 3) {
		t.Fatal("a non-failed job must stay out of the dead-letter queue")
	}
}
