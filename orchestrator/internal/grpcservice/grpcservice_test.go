package grpcservice_test

import (
	"context"
	"errors"
	"github.com/OinkiePie/calc_3/config"
	"github.com/OinkiePie/calc_3/orchestrator/internal/grpcservice"
	mm "github.com/OinkiePie/calc_3/orchestrator/internal/managers"
	"github.com/OinkiePie/calc_3/orchestrator/internal/providers"
	"github.com/OinkiePie/calc_3/pkg/logger"
	"github.com/OinkiePie/calc_3/pkg/models"
	pb "github.com/OinkiePie/calc_3/pkg/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"log"
	"net"
	"net/http"
	"testing"
)

func init() {
	// Отключаем выводы и инициализируем конфиг
	log.SetOutput(io.Discard)
	_ = config.InitConfig()
	logger.InitLogger(logger.Options{Level: 6})
}

func float64Ptr(i float64) *float64 {
	return &i
}

func TestGetTask_Success(t *testing.T) {
	mockEM := new(mm.MockExpressionManager)
	mockPr := &providers.Providers{ExprManager: mockEM}
	server := grpcservice.NewOrchestratorGRPCServer(mockPr)

	expectedTask := &models.Task{
		ID:         1,
		Args:       []*float64{float64Ptr(2.0), float64Ptr(3.0)},
		Operation:  "+",
		Expression: 1,
	}

	mockEM.On("ReadTask", mock.Anything).Return(expectedTask, nil, http.StatusOK)

	resp, err := server.GetTask(context.Background(), &pb.Empty{})

	assert.NoError(t, err)
	assert.Equal(t, expectedTask.ID, resp.Id)
	assert.Equal(t, expectedTask.Operation, resp.Operation)
	assert.Equal(t, expectedTask.Expression, resp.Expression)
	assert.Len(t, resp.Args, 2)
	assert.Equal(t, expectedTask.Args[0], resp.Args[0].Value)
	assert.Equal(t, expectedTask.Args[1], resp.Args[1].Value)
	mockEM.AssertExpectations(t)
}

func TestGetTask_NoTaskAvailable(t *testing.T) {
	mockEM := new(mm.MockExpressionManager)
	mockPr := &providers.Providers{ExprManager: mockEM}
	server := grpcservice.NewOrchestratorGRPCServer(mockPr)

	expectedErr := errors.New("error")

	mockEM.On("ReadTask", mock.Anything).Return((*models.Task)(nil), expectedErr, http.StatusNotFound)

	resp, err := server.GetTask(context.Background(), &pb.Empty{})

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Equal(t, expectedErr.Error(), err.Error())
	mockEM.AssertExpectations(t)
}

func TestSubmitResult_Success(t *testing.T) {
	mockEM := new(mm.MockExpressionManager)
	mockPr := &providers.Providers{ExprManager: mockEM}
	server := grpcservice.NewOrchestratorGRPCServer(mockPr)

	completedTask := &pb.TaskCompleted{
		Id:         1,
		Result:     5.0,
		Expression: 1,
	}

	mockEM.On("CompleteTask", mock.Anything, mock.Anything).Return(nil, http.StatusOK)

	resp, err := server.SubmitResult(context.Background(), completedTask)

	assert.NoError(t, err)
	assert.Equal(t, &pb.Empty{}, resp)
	mockEM.AssertExpectations(t)
}

func TestSubmitResult_Error(t *testing.T) {
	mockEM := new(mm.MockExpressionManager)
	mockPr := &providers.Providers{ExprManager: mockEM}
	server := grpcservice.NewOrchestratorGRPCServer(mockPr)

	completedTask := &pb.TaskCompleted{
		Id:         1,
		Result:     5.0,
		Expression: 1,
	}
	expectedErr := errors.New("error")

	mockEM.On("CompleteTask", mock.Anything, mock.Anything).Return(expectedErr, http.StatusInternalServerError)

	resp, err := server.SubmitResult(context.Background(), completedTask)

	assert.Error(t, err)
	assert.Equal(t, expectedErr.Error(), err.Error())
	assert.NotNil(t, resp)
	mockEM.AssertExpectations(t)
}

func TestGetTask_ManagerError(t *testing.T) {
	mockEM := new(mm.MockExpressionManager)
	mockPr := &providers.Providers{ExprManager: mockEM}
	server := grpcservice.NewOrchestratorGRPCServer(mockPr)

	expectedErr := errors.New("error")
	mockEM.On("ReadTask", mock.Anything).Return((*models.Task)(nil), expectedErr, http.StatusInternalServerError)

	resp, err := server.GetTask(context.Background(), &pb.Empty{})

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Equal(t, expectedErr.Error(), err.Error())
	mockEM.AssertExpectations(t)
}

func TestGRPCServerIntegration(t *testing.T) {
	mockEM := new(mm.MockExpressionManager)
	mockPr := &providers.Providers{ExprManager: mockEM}

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	grpcServer := grpc.NewServer()
	pb.RegisterOrchestratorServiceServer(grpcServer, grpcservice.NewOrchestratorGRPCServer(mockPr))
	go func() {
		_ = grpcServer.Serve(lis)
	}()
	defer grpcServer.Stop()

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	client := pb.NewOrchestratorServiceClient(conn)

	t.Run("GetTask", func(t *testing.T) {
		expectedTask := &models.Task{
			ID:         1,
			Args:       []*float64{float64Ptr(2.0), float64Ptr(3.0)},
			Operation:  "+",
			Expression: 1,
		}

		mockEM.On("ReadTask", mock.Anything).Return(expectedTask, nil, http.StatusOK)

		resp, err := client.GetTask(context.Background(), &pb.Empty{})
		require.NoError(t, err)
		assert.Equal(t, int64(1), resp.Id)
	})

	t.Run("SubmitResult", func(t *testing.T) {
		completedTask := &pb.TaskCompleted{
			Id:         1,
			Result:     5.0,
			Expression: 1,
		}

		mockEM.On("CompleteTask", mock.Anything, mock.Anything).Return(nil, http.StatusOK)

		_, err := client.SubmitResult(context.Background(), completedTask)
		assert.NoError(t, err)
	})
}
