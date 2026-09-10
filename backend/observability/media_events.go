package observability

func MediaJobClaimed(jobID uint) {
	Event("MEDIA_JOB_CLAIMED", map[string]any{"job_id": jobID})
}

func MediaProcessingStarted(jobID, uploadID uint) {
	Event("MEDIA_PROCESSING_STARTED", map[string]any{"job_id": jobID, "upload_id": uploadID})
}

func MediaProcessingCompleted(jobID uint) {
	Event("MEDIA_PROCESSING_COMPLETED", map[string]any{"job_id": jobID})
}

func MediaProcessingFailed(jobID uint, err error) {
	Event("MEDIA_PROCESSING_FAILED", map[string]any{"job_id": jobID, "error": err.Error()})
}
