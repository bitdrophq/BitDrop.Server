package utils

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

// UploadToSupabase uploads a file to Supabase Storage and returns the public URL or error
func UploadToSupabase(file multipart.File, filename, bucket string) (string, error) {
	supabaseURL := os.Getenv("SUPABASE_URL")
	serviceKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	if supabaseURL == "" || serviceKey == "" {
		return "", fmt.Errorf("Supabase URL or service key not set in env")
	}

	// Read file into buffer
	buf := new(bytes.Buffer)
	_, err := io.Copy(buf, file)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	uploadURL := fmt.Sprintf("%s/storage/v1/object/%s/%s", supabaseURL, bucket, filename)

	req, err := http.NewRequest("POST", uploadURL, buf)
	if err != nil {
		return "", fmt.Errorf("failed to create upload request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", serviceKey))
	req.Header.Set("Content-Type", "application/octet-stream")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("upload request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("upload failed: %s", string(body))
	}

	// Public URL format: {SUPABASE_URL}/storage/v1/object/public/{bucket}/{filename}
	publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", supabaseURL, bucket, filename)
	return publicURL, nil
}

// DeleteFromSupabase deletes a file from Supabase Storage
func DeleteFromSupabase(filename, bucket string) error {
	supabaseURL := os.Getenv("SUPABASE_URL")
	serviceKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	if supabaseURL == "" || serviceKey == "" {
		return fmt.Errorf("Supabase URL or service key not set in env")
	}

	deleteURL := fmt.Sprintf("%s/storage/v1/object/%s/%s", supabaseURL, bucket, filename)
	req, err := http.NewRequest("DELETE", deleteURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create delete request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", serviceKey))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("delete request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete failed: %s", string(body))
	}
	return nil
} 