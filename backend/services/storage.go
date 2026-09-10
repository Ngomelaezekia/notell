package services

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"notell/config"
	"notell/models"
	"notell/observability"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"gorm.io/gorm"
)

type MediaStorage interface {
	Put(ctx context.Context, key, localPath, contentType string) error
	Delete(ctx context.Context, key string) error
	Open(ctx context.Context, key string) (io.ReadCloser, string, int64, error)
	PublicURL(key string) string
}

// ... rest of storage implementation remains unchanged
