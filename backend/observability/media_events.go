package observability

func MediaJobClaimed(jobID uint) {
	Event("MEDIA_JOB_CLAIMED", map[string]any{"job_id": jobID})
}

func MediaProcessingStarted(jobID uint) {
	Event("MEDIA_PROCESSING_STARTED", map[string]any{"job_id": jobID})
}

func MediaProcessingCompleted(jobID uint, seconds float64) {
	Event("MEDIA_PROCESSING_COMPLETED", map[string]any{"job_id": jobID, "duration": seconds})
}

func MediaProcessingFailed(jobID uint, err error) {
	Event("MEDIA_PROCESSING_FAILED", map[string]any{"job_id": jobID, "error": err.Error()})
}
