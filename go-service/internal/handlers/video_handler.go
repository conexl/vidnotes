package handlers

import (
	"github.com/code-zt/vidnotes/internal/services"
	"github.com/code-zt/vidnotes/pkg/utils"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type VideoHandlers struct {
	videoService services.VideoService
}

func NewVideoHandlers(videoService services.VideoService) *VideoHandlers {
	return &VideoHandlers{
		videoService: videoService,
	}
}

type UploadVideoRequest struct {
	Filename string `json:"filename"`
}

func (h *VideoHandlers) UploadVideo(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Invalid user ID")
	}

	form, err := c.MultipartForm()
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Invalid form data")
	}

	files := form.File["file"]
	if len(files) == 0 {
		return utils.Error(c, fiber.StatusBadRequest, "No file provided")
	}

	fileHeader := files[0]
	file, err := fileHeader.Open()
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "Failed to read file")
	}
	defer file.Close()

	// Читаем файл в память
	fileBytes := make([]byte, fileHeader.Size)
	if _, err := file.Read(fileBytes); err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "Failed to read file")
	}

	video, err := h.videoService.UploadVideo(c.Context(), userObjectID, fileBytes, fileHeader.Filename)
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.Success(c, fiber.StatusAccepted, fiber.Map{
		"video":   video,
		"message": "Video uploaded and processing started",
	})
}

func (h *VideoHandlers) GetVideoStatus(c *fiber.Ctx) error {
	videoID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Invalid video ID")
	}

	video, err := h.videoService.GetVideoStatus(c.Context(), videoID)
	if err != nil {
		return utils.Error(c, fiber.StatusNotFound, "Video not found")
	}

	return utils.Success(c, fiber.StatusOK, video)
}

func (h *VideoHandlers) GetUserVideos(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Invalid user ID")
	}

	videos, err := h.videoService.GetUserVideos(c.Context(), userObjectID)
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "Failed to get videos")
	}

	return utils.Success(c, fiber.StatusOK, videos)
}

func (h *VideoHandlers) GetVideoResult(c *fiber.Ctx) error {
	videoID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Invalid video ID")
	}

	result, err := h.videoService.GetVideoResult(c.Context(), videoID)
	if err != nil {
		return utils.Error(c, fiber.StatusNotFound, "Video result not found")
	}

	return utils.Success(c, fiber.StatusOK, result)
}

func (h *VideoHandlers) DeleteVideo(c *fiber.Ctx) error {
	videoID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Invalid video ID")
	}

	if err := h.videoService.DeleteVideo(c.Context(), videoID); err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "Failed to delete video")
	}

	return utils.Success(c, fiber.StatusOK, fiber.Map{
		"message": "Video deleted successfully",
	})
}
