// services/video_service.go
package services

import (
	"context"
	"fmt"
	"time"

	pb "github.com/code-zt/vidnotes/api/proto"
	"github.com/code-zt/vidnotes/internal/models"
	"github.com/code-zt/vidnotes/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc"
)

type VideoService interface {
	UploadVideo(ctx context.Context, userID primitive.ObjectID, file []byte, filename string) (*models.Video, error)
	GetVideoStatus(ctx context.Context, videoID primitive.ObjectID) (*models.Video, error)
	GetUserVideos(ctx context.Context, userID primitive.ObjectID) ([]*models.Video, error)
	GetVideoResult(ctx context.Context, videoID primitive.ObjectID) (string, error)
	DeleteVideo(ctx context.Context, videoID primitive.ObjectID) error
}

type videoService struct {
	videoRepo   repository.VideoRepository
	userService UserService
	grpcClient  pb.VideoProcessorClient
}

func NewVideoService(
	videoRepo repository.VideoRepository,
	userService UserService,
	grpcConn *grpc.ClientConn,
) VideoService {
	return &videoService{
		videoRepo:   videoRepo,
		userService: userService,
		grpcClient:  pb.NewVideoProcessorClient(grpcConn),
	}
}

func (s *videoService) UploadVideo(ctx context.Context, userID primitive.ObjectID, file []byte, filename string) (*models.Video, error) {
	// Проверяем лимиты пользователя
	if err := s.userService.CanPerformAnalysis(ctx, userID); err != nil {
		return nil, err
	}

	// Проверяем что файл не пустой
	if len(file) == 0 {
		return nil, fmt.Errorf("uploaded file is empty")
	}

	// Создаем запись видео в базе
	video := &models.Video{
		UserID:    userID,
		Title:     filename,
		URL:       fmt.Sprintf("/videos/%s", primitive.NewObjectID().Hex()),
		Status:    "uploaded",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	videoID, err := s.videoRepo.Create(ctx, video)
	if err != nil {
		return nil, err
	}

	// Запускаем обработку в фоне
	go s.processVideo(context.Background(), videoID, file, filename)

	// Увеличиваем счетчик анализов
	if err := s.userService.RecordAnalysis(ctx, userID); err != nil {
		fmt.Printf("Failed to record analysis: %v\n", err)
	}

	return video, nil
}

func (s *videoService) processVideo(ctx context.Context, videoID primitive.ObjectID, file []byte, filename string) error {
	// Обновляем статус на "processing"
	if err := s.videoRepo.UpdateStatus(ctx, videoID, "processing"); err != nil {
		return fmt.Errorf("failed to update video status: %w", err)
	}

	// Создаем контекст с таймаутом для gRPC вызова
	grpcCtx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()

	// Создаем gRPC stream
	stream, err := s.grpcClient.ProcessVideo(grpcCtx)
	if err != nil {
		s.videoRepo.UpdateStatus(ctx, videoID, "failed")
		return fmt.Errorf("failed to create gRPC stream: %w", err)
	}

	// Отправляем первый чанк с метаданными
	metadataChunk := &pb.VideoChunk{
		Filename: filename,
		VideoId:  videoID.Hex(),
		Data:     []byte{}, // Пустые данные для первого чанка
	}

	if err := stream.Send(metadataChunk); err != nil {
		s.videoRepo.UpdateStatus(ctx, videoID, "failed")
		return fmt.Errorf("failed to send metadata: %w", err)
	}

	fmt.Printf("Sending video data: %d bytes in chunks\n", len(file))

	// Отправляем данные файла чанками
	chunkSize := 64 * 1024 // 64KB chunks для лучшей производительности
	sentBytes := 0

	for i := 0; i < len(file); i += chunkSize {
		end := i + chunkSize
		if end > len(file) {
			end = len(file)
		}

		chunk := &pb.VideoChunk{
			Data: file[i:end],
		}

		if err := stream.Send(chunk); err != nil {
			s.videoRepo.UpdateStatus(ctx, videoID, "failed")
			return fmt.Errorf("failed to send chunk [%d-%d]: %w", i, end, err)
		}

		sentBytes += len(chunk.Data)

		// Логируем прогресс каждые 1MB
		if sentBytes%(1024*1024) == 0 {
			fmt.Printf("Sent %d/%d bytes (%.1f%%)\n", sentBytes, len(file), float64(sentBytes)/float64(len(file))*100)
		}
	}

	fmt.Printf("All data sent: %d bytes\n", sentBytes)

	// Закрываем поток и получаем ответ
	resp, err := stream.CloseAndRecv()
	if err != nil {
		s.videoRepo.UpdateStatus(ctx, videoID, "failed")
		return fmt.Errorf("failed to receive response: %w", err)
	}

	// Проверяем статус ответа
	if resp.Status == "failed" || resp.Error != "" {
		s.videoRepo.UpdateStatus(ctx, videoID, "failed")
		return fmt.Errorf("processing failed: %s", resp.Error)
	}

	// Сохраняем summary в видео
	if err := s.videoRepo.UpdateSummary(ctx, videoID, resp.Summary); err != nil {
		s.videoRepo.UpdateStatus(ctx, videoID, "failed")
		return fmt.Errorf("failed to save video summary: %w", err)
	}

	// Обновляем статус видео на "completed"
	if err := s.videoRepo.UpdateStatus(ctx, videoID, "completed"); err != nil {
		return fmt.Errorf("failed to update video status: %w", err)
	}

	fmt.Printf("Video %s processed successfully. Summary length: %d\n", videoID.Hex(), len(resp.Summary))
	return nil
}

func (s *videoService) GetVideoStatus(ctx context.Context, videoID primitive.ObjectID) (*models.Video, error) {
	video, err := s.videoRepo.GetByID(ctx, videoID)
	if err != nil {
		return nil, err
	}
	return video, nil
}

func (s *videoService) GetUserVideos(ctx context.Context, userID primitive.ObjectID) ([]*models.Video, error) {
	videos, err := s.videoRepo.GetByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	return videos, nil
}

func (s *videoService) GetVideoResult(ctx context.Context, videoID primitive.ObjectID) (string, error) {
	video, err := s.videoRepo.GetByID(ctx, videoID)
	if err != nil {
		return "", err
	}

	if video.Status != "completed" {
		return "", fmt.Errorf("video processing not completed")
	}

	return video.Summary, nil
}

func (s *videoService) DeleteVideo(ctx context.Context, videoID primitive.ObjectID) error {
	return s.videoRepo.Delete(ctx, videoID)
}
