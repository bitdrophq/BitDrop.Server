package handlers

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/richiethie/BitDrop.Server/internal/db"
	"github.com/richiethie/BitDrop.Server/internal/models"
	"github.com/richiethie/BitDrop.Server/internal/utils"
)

// UploadDropHandler handles video uploads and creates a Drop record
func UploadDropHandler(c *gin.Context) {
	// Parse multipart form
	err := c.Request.ParseMultipartForm(100 << 20) // 100MB max
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form: " + err.Error()})
		return
	}

	file, header, err := c.Request.FormFile("video")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Video file is required"})
		return
	}
	defer file.Close()

	caption := c.PostForm("caption")
	groupIDStr := c.PostForm("group_id")
	var groupID *uuid.UUID
	if groupIDStr != "" {
		gid, err := uuid.Parse(groupIDStr)
		if err == nil {
			groupID = &gid
		}
	}

	// Get user ID from context (set by AuthMiddleware)
	userIDVal, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userIDStr, ok := userIDVal.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID in context"})
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User ID is not a valid UUID"})
		return
	}

	// Generate a unique filename for the video
	ext := ""
	if header != nil {
		name := header.Filename
		for i := len(name) - 1; i >= 0; i-- {
			if name[i] == '.' {
				ext = name[i:]
				break
			}
		}
	}
	filename := uuid.New().String() + ext

	// Save uploaded video to a temp file
	tmpVideoFile, err := os.CreateTemp("", "upload-*.mp4")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create temp video file: " + err.Error()})
		return
	}
	defer os.Remove(tmpVideoFile.Name())

	_, err = io.Copy(tmpVideoFile, file)
	if err != nil {
		tmpVideoFile.Close()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save temp video file: " + err.Error()})
		return
	}
	tmpVideoFile.Close()

	// Log temp video file size
	if fi, err := os.Stat(tmpVideoFile.Name()); err == nil {
		log.Println("Temp video file size:", fi.Size())
	} else {
		log.Println("Could not stat temp video file:", err)
	}

	// Reopen the temp file for reading (for upload)
	videoReader, err := os.Open(tmpVideoFile.Name())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open temp video file: " + err.Error()})
		return
	}
	defer videoReader.Close()

	// Upload file to Supabase Storage 'drops' bucket
	videoURL, err := utils.UploadToSupabase(videoReader, filename, "drops")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload video: " + err.Error()})
		return
	}

	// Use a dedicated subdirectory for thumbnails
	thumbDir := filepath.Join(os.TempDir(), "bitdrop_thumbs")
	if err := os.MkdirAll(thumbDir, 0o755); err != nil {
		log.Println("Failed to create thumbnail directory:", thumbDir, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create thumbnail directory: " + err.Error()})
		return
	}
	thumbPath := filepath.Join(thumbDir, "thumb-"+uuid.New().String()+"-"+strconv.FormatInt(time.Now().UnixNano(), 10)+".jpg")
	log.Println("Thumbnail output path:", thumbPath)

	// If the file already exists, delete it and abort if removal fails
	if _, err := os.Stat(thumbPath); err == nil {
		log.Println("Thumbnail file already exists, removing:", thumbPath)
		if err := os.Remove(thumbPath); err != nil {
			log.Println("Failed to remove existing thumbnail file:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove existing thumbnail file: " + err.Error()})
			return
		}
	}
	defer os.Remove(thumbPath)

	// Run ffmpeg to create the thumbnail
	var ffmpegOut bytes.Buffer
	cmd := exec.Command("ffmpeg", "-y", "-i", tmpVideoFile.Name(), "-ss", "00:00:01.000", "-vframes", "1", thumbPath)
	cmd.Stderr = &ffmpegOut
	err = cmd.Run()
	log.Println("ffmpeg output:", ffmpegOut.String())
	if err != nil {
		log.Println("ffmpeg error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate thumbnail: " + err.Error()})
		return
	}

	// Log thumbnail file size
	if thumbFi, err := os.Stat(thumbPath); err == nil {
		log.Println("Thumbnail file size:", thumbFi.Size())
	} else {
		log.Println("Could not stat thumbnail file:", err)
	}

	// Only now, open the file for upload
	thumbReader, err := os.Open(thumbPath)
	if err != nil {
		log.Println("Failed to open generated thumbnail:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open generated thumbnail: " + err.Error()})
		return
	}
	defer thumbReader.Close()

	thumbFilename := "thumbnails/" + uuid.New().String() + ".jpg"
	thumbURL, err := utils.UploadToSupabase(thumbReader, thumbFilename, "drops")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload thumbnail: " + err.Error()})
		return
	}

	// Save Drop to DB
	drop := models.Drop{
		ID:        uuid.New(),
		UserID:    userID,
		GroupID:   groupID,
		VideoURL:  videoURL,
		Thumbnail: thumbURL,
		Caption:   caption,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Votes:     0,
	}

	_, err = db.DB.Exec(context.Background(),
		`INSERT INTO drops (id, user_id, group_id, video_url, thumbnail, caption, created_at, updated_at, votes)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		drop.ID, drop.UserID, drop.GroupID, drop.VideoURL, drop.Thumbnail, drop.Caption, drop.CreatedAt, drop.UpdatedAt, drop.Votes,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert drop: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, drop)
}

// GetUserDropsHandler returns all drops for the authenticated user
func GetUserDropsHandler(c *gin.Context) {
	userIDVal, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userIDStr, ok := userIDVal.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID in context"})
		return
	}
	rows, err := db.DB.Query(context.Background(),
		`SELECT id, user_id, group_id, video_url, thumbnail, caption, created_at, updated_at, votes, visibility
		 FROM drops WHERE user_id = $1 ORDER BY created_at DESC`, userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch drops: " + err.Error()})
		return
	}
	defer rows.Close()

	drops := []models.Drop{}
	for rows.Next() {
		var d models.Drop
		err := rows.Scan(&d.ID, &d.UserID, &d.GroupID, &d.VideoURL, &d.Thumbnail, &d.Caption, &d.CreatedAt, &d.UpdatedAt, &d.Votes, &d.Visibility)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan drop: " + err.Error()})
			return
		}
		drops = append(drops, d)
	}
	c.JSON(http.StatusOK, drops)
}

// Handler to get drop details with user info
func GetDropDetailsHandler(c *gin.Context) {
	dropID := c.Param("id")
	var drop models.Drop
	var username, avatarURL string
	err := db.DB.QueryRow(context.Background(),
		`SELECT d.id, d.user_id, d.group_id, d.video_url, d.thumbnail, d.caption, d.created_at, d.updated_at, d.votes, d.visibility,
		        u.username, u.avatar_url
		 FROM drops d
		 JOIN users u ON d.user_id = u.id
		 WHERE d.id = $1`, dropID).
		Scan(&drop.ID, &drop.UserID, &drop.GroupID, &drop.VideoURL, &drop.Thumbnail, &drop.Caption, &drop.CreatedAt, &drop.UpdatedAt, &drop.Votes, &drop.Visibility, &username, &avatarURL)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Drop not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"drop": drop,
		"user": gin.H{"id": drop.UserID, "username": username, "avatar_url": avatarURL},
	})
}

// Handler to delete a drop by id (only if user is owner)
func DeleteDropHandler(c *gin.Context) {
	dropID := c.Param("id")
	userIDVal, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userIDStr, ok := userIDVal.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID in context"})
		return
	}
	log.Println("DeleteDropHandler dropID:", dropID)
	// Check if drop exists and belongs to user, and get video/thumbnail URLs
	var ownerID, videoURL, thumbURL string
	err := db.DB.QueryRow(context.Background(), "SELECT user_id, video_url, thumbnail FROM drops WHERE id = $1", dropID).Scan(&ownerID, &videoURL, &thumbURL)
	if err != nil {
		log.Println("DB error:", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Drop not found"})
		return
	}
	if ownerID != userIDStr {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to delete this drop"})
		return
	}
	// Extract file paths from URLs (after /drops/)
	getPath := func(url string) string {
		idx := strings.Index(url, "/drops/")
		if idx == -1 {
			return ""
		}
		return url[idx+len("/drops/"):]
	}
	videoPath := getPath(videoURL)
	thumbPath := getPath(thumbURL)
	// Delete files from Supabase Storage
	if videoPath != "" {
		_ = utils.DeleteFromSupabase(videoPath, "drops")
	}
	if thumbPath != "" {
		_ = utils.DeleteFromSupabase(thumbPath, "drops")
	}
	_, err = db.DB.Exec(context.Background(), "DELETE FROM drops WHERE id = $1", dropID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete drop: " + err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
} 