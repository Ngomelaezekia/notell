package postaccess

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var subscriberHTTPClient = &http.Client{Timeout: 3 * time.Second}

func hasSubscriberAccess(viewerID, postID uint) bool {
	active, err := subscriberEntitled(context.Background(), viewerID, postID)
	return err == nil && active
}

func subscriberEntitled(ctx context.Context, viewerID, postID uint) (bool, error) {
	if viewerID == 0 || postID == 0 {
		return false, nil
	}
	base := strings.TrimRight(strings.TrimSpace(os.Getenv("PAYMENT_SERVICE_URL")), "/")
	key := strings.TrimSpace(os.Getenv("INTERNAL_SERVICE_KEY"))
	if base == "" || key == "" {
		return false, nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/v1/internal/entitlements/post/"+strconv.FormatUint(uint64(postID), 10), nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("X-Internal-Service-Key", key)
	req.Header.Set("X-User-ID", strconv.FormatUint(uint64(viewerID), 10))
	resp, err := subscriberHTTPClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return false, nil
	}
	if resp.StatusCode != http.StatusOK {
		return false, errors.New("entitlement service unavailable")
	}
	var result struct {
		Active bool `json:"active"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, err
	}
	return result.Active, nil
}
