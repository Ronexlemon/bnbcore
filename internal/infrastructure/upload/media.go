package upload


import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/redis/go-redis/v9"
)

type MediaService struct {
	cld   *cloudinary.Cloudinary
	rdb   *redis.Client
	cacheTTL time.Duration
}

func NewMediaService(cldURL string, rdb *redis.Client) (*MediaService, error) {
	cld, err := cloudinary.NewFromURL(cldURL)
	if err != nil {
		return nil, fmt.Errorf("failed to init Cloudinary: %w", err)
	}

	return &MediaService{
		cld:      cld,
		rdb:      rdb,
		cacheTTL: 24 * time.Hour, // Cache asset URLs for 24 hours
	}, nil
}

func (s *MediaService) UploadAndCacheStream(ctx context.Context, fileStream io.Reader, assetKey string, folder string) (string, error) {
	cachedURL, err := s.rdb.Get(ctx, assetKey).Result()
	if err == nil {
		return cachedURL, nil 
	} else if err != redis.Nil {
		log.Printf("Redis error (fallback to uploading): %v", err)
	}

	uploadCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
isUnique := true
	//  Configure streaming options with automated WebP/Next-Gen optimization
	uploadParams := uploader.UploadParams{
		Folder:          folder,
		UniqueFilename:  &isUnique,
		// f_auto: dynamically converts to webp/avif based on consumer browser capability
		// q_auto: adjusts compression without visible quality degradation
		Transformation:  "f_auto,q_auto", 
	}

	result, err := s.cld.Upload.Upload(uploadCtx, fileStream, uploadParams)
	if err != nil {
		return "", fmt.Errorf("streaming upload to Cloudinary failed: %w", err)
	}
	if result.Error.Message != "" {
    return "", fmt.Errorf("cloudinary API error: %s ", result.Error.Message)
}
	fmt.Println("THE RESUKT",result)
	log.Println("THE RESUKT",result)

	go func() {
		cacheCtx, cacheCancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cacheCancel()
		
		if err := s.rdb.Set(cacheCtx, assetKey, result.SecureURL, s.cacheTTL).Err(); err != nil {
			log.Printf("Failed to write asset URL to Redis cache: %v", err)
		}
	}()

	return result.SecureURL, nil
}