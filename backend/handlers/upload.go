package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"notell/models"
	"notell/services"
	mediaservice "notell/services/media"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// upload recovery hardening: failed storage cleanup is tracked instead of silently ignored.

