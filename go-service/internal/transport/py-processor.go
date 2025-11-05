package transport

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	pb "github.com/code-zt/vidnotes/api/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	// Увеличиваем лимит до 500MB
	r.Body = http.MaxBytesReader(w, r.Body, 500<<20)

	if err := r.ParseMultipartForm(500 << 20); err != nil {
		http.Error(w, "Form error: "+err.Error(), http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "File error: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Подключаемся к gRPC серверу
	conn, err := grpc.Dial("localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(500*1024*1024),
			grpc.MaxCallSendMsgSize(500*1024*1024),
		),
	)
	if err != nil {
		http.Error(w, "gRPC connection failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	client := pb.NewVideoProcessorClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	stream, err := client.ProcessVideo(ctx)
	if err != nil {
		http.Error(w, "gRPC stream failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Отправляем имя файла
	if err := stream.Send(&pb.VideoChunk{Filename: header.Filename}); err != nil {
		http.Error(w, "Failed to send filename: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Отправляем данные файла большими чанками
	const chunkSize = 256 * 1024 // 256KB chunks для лучшей производительности
	buf := make([]byte, chunkSize)
	totalSent := 0

	for {
		n, err := file.Read(buf)
		if n > 0 {
			totalSent += n
			if err := stream.Send(&pb.VideoChunk{Data: buf[:n]}); err != nil {
				http.Error(w, "Chunk send failed: "+err.Error(), http.StatusInternalServerError)
				return
			}
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			http.Error(w, "File read error: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Получаем ответ
	resp, err := stream.CloseAndRecv()
	if err != nil {
		http.Error(w, "Response error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if resp.Error != "" {
		http.Error(w, "Processing error: "+resp.Error, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
