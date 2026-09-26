package observability

import (
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    "log"
    "strings"
    "time"
)

func HashSensitive(value string) string {
    sum := sha256.Sum256([]byte(value))
    return hex.EncodeToString(sum[:])
}

func HashError(err error) string {
    if err == nil { return "" }
    return HashSensitive(err.Error())
}

func SafeError(err error) string {
    if err == nil { return "" }
    return "error_hash=" + HashError(err)
}

func Event(name string, fields map[string]any) {
    safe := make(map[string]any, len(fields))
    for key, value := range fields {
        switch strings.ToLower(strings.TrimSpace(key)) {
        case "error", "err", "token", "secret", "authorization", "cookie", "url", "key", "path", "filename", "email":
            safe[key+"_hash"] = hashField(value)
        default:
            safe[key] = value
        }
    }
    log.Printf("event=%s fields=%v", name, safe)
}

func hashField(value any) string {
    switch v := value.(type) {
    case string: return HashSensitive(v)
    case error: return HashError(v)
    default: return HashSensitive(fmt.Sprint(v))
    }
}

func Duration(start time.Time) float64 {
    return time.Since(start).Seconds()
}
