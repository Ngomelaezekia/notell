package observability

import "log"

func MediaJobClaimed(jobID uint) {
	log.Printf("event=MEDIA_JOB_CLAIMED job_id=%d", jobID)
}

func MediaProcessingStarted(jobID, uploadID uint) {
	log.Printf("event=MEDIA_PROCESSING_STARTED job_id=%d upload_id=%d", jobID, uploadID)
}

func MediaProcessingCompleted(jobID uint) {
	log.Printf("event=MEDIA_PROCESSING_COMPLETED job_id=%d", jobID)
}

func MediaProcessingFailed(jobID uint, err error) {
	log.Printf("event=MEDIA_PROCESSING_FAILED job_id=%d error=%v", jobID, err)
}

func MediaStorageFailed(operation string, err error) {
	log.Printf("event=MEDIA_STORAGE_FAILED operation=%s error=%v", operation, err)
}
