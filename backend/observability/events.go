package observability

func MediaStorageFailed(operation string, err error) {
	Event("MEDIA_STORAGE_FAILED", map[string]any{"operation": operation, "error": err.Error()})
}
