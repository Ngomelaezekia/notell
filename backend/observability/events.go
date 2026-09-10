package observability

func MediaStorageFailed(operation string, err error) {
	logEvent("MEDIA_STORAGE_FAILED", map[string]any{"operation": operation, "error": err.Error()})
}
