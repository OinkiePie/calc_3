package workers_test

import (
	"context"
	"errors"
	"github.com/OinkiePie/calc_3/agent/internal/workers"
	"github.com/OinkiePie/calc_3/config"
	"github.com/OinkiePie/calc_3/pkg/logger"
	"github.com/OinkiePie/calc_3/pkg/operators"
	pb "github.com/OinkiePie/calc_3/pkg/proto"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"io"
	"log"
	"sync"
	"testing"
	"time"
)

func init() {
	// Отключаем выводы и инициализируем конфиг
	log.SetOutput(io.Discard)
	_ = config.InitConfig()
	logger.InitLogger(logger.Options{Level: 6})
}

type MockOrchestratorClient struct {
	GetTaskFunc      func(context.Context, *pb.Empty) (*pb.TaskResponse, error)
	SubmitResultFunc func(context.Context, *pb.TaskCompleted) (*pb.Empty, error)
}

func (m *MockOrchestratorClient) GetTask(ctx context.Context, in *pb.Empty, _ ...grpc.CallOption) (*pb.TaskResponse, error) {
	return m.GetTaskFunc(ctx, in)
}

func (m *MockOrchestratorClient) SubmitResult(ctx context.Context, in *pb.TaskCompleted, _ ...grpc.CallOption) (*pb.Empty, error) {
	return m.SubmitResultFunc(ctx, in)
}

func float64Ptr(i float64) *float64 {
	return &i
}

func TestWorker(t *testing.T) {
	// Вспомогательная функция для создания worker
	newTestWorker := func(t *testing.T, mockClient *MockOrchestratorClient) (*workers.Worker, context.CancelFunc, chan error) {
		t.Helper()
		var wg sync.WaitGroup
		errChan := make(chan error, 1)
		wg.Add(1)
		worker := workers.NewWorker(1, mockClient, &wg, errChan)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)

		go func() {
			defer wg.Done()
			worker.Start(ctx)
		}()

		return worker, cancel, errChan
	}

	tests := []struct {
		name      string
		setupMock func() *MockOrchestratorClient
		wantErr   bool
	}{
		{
			name: "successful task processing",
			setupMock: func() *MockOrchestratorClient {
				return &MockOrchestratorClient{
					GetTaskFunc: func(ctx context.Context, empty *pb.Empty) (*pb.TaskResponse, error) {
						return &pb.TaskResponse{
							Id:         1,
							Args:       []*pb.WrappedDouble{{Value: float64Ptr(2)}, {Value: float64Ptr(3)}},
							Operation:  operators.OpAdd,
							Expression: 1,
						}, nil
					},
					SubmitResultFunc: func(ctx context.Context, completed *pb.TaskCompleted) (*pb.Empty, error) {
						assert.Equal(t, int64(1), completed.Id)
						assert.Equal(t, int64(1), completed.Expression)
						return &pb.Empty{}, nil
					},
				}
			},
			wantErr: false,
		},
		{
			name: "get task error",
			setupMock: func() *MockOrchestratorClient {
				return &MockOrchestratorClient{
					GetTaskFunc: func(ctx context.Context, empty *pb.Empty) (*pb.TaskResponse, error) {
						return nil, errors.New("error")
					},
					SubmitResultFunc: func(ctx context.Context, completed *pb.TaskCompleted) (*pb.Empty, error) {
						return &pb.Empty{}, nil
					},
				}
			},
			wantErr: false,
		},
		{
			name: "get task unavailable",
			setupMock: func() *MockOrchestratorClient {
				return &MockOrchestratorClient{
					GetTaskFunc: func(ctx context.Context, empty *pb.Empty) (*pb.TaskResponse, error) {
						return nil, nil
					},
					SubmitResultFunc: func(ctx context.Context, completed *pb.TaskCompleted) (*pb.Empty, error) {
						return nil, errors.New("submit failed")
					},
				}
			},
			wantErr: false,
		},
		{
			name: "calculation add",
			setupMock: func() *MockOrchestratorClient {
				return &MockOrchestratorClient{
					GetTaskFunc: func(ctx context.Context, empty *pb.Empty) (*pb.TaskResponse, error) {
						return &pb.TaskResponse{
							Id:         1,
							Args:       []*pb.WrappedDouble{{Value: float64Ptr(2)}, {Value: float64Ptr(3)}},
							Operation:  operators.OpAdd,
							Expression: 1,
						}, nil
					},
					SubmitResultFunc: func(ctx context.Context, completed *pb.TaskCompleted) (*pb.Empty, error) {
						assert.Equal(t, 5.0, completed.Result)
						return &pb.Empty{}, nil
					},
				}
			},
			wantErr: false,
		},
		{
			name: "calculation subtract",
			setupMock: func() *MockOrchestratorClient {
				return &MockOrchestratorClient{
					GetTaskFunc: func(ctx context.Context, empty *pb.Empty) (*pb.TaskResponse, error) {
						return &pb.TaskResponse{
							Id:         1,
							Args:       []*pb.WrappedDouble{{Value: float64Ptr(2)}, {Value: float64Ptr(3)}},
							Operation:  operators.OpSubtract,
							Expression: 1,
						}, nil
					},
					SubmitResultFunc: func(ctx context.Context, completed *pb.TaskCompleted) (*pb.Empty, error) {
						assert.Equal(t, -1.0, completed.Result)
						return &pb.Empty{}, nil
					},
				}
			},
			wantErr: false,
		},
		{
			name: "calculation multiply",
			setupMock: func() *MockOrchestratorClient {
				return &MockOrchestratorClient{
					GetTaskFunc: func(ctx context.Context, empty *pb.Empty) (*pb.TaskResponse, error) {
						return &pb.TaskResponse{
							Id:         1,
							Args:       []*pb.WrappedDouble{{Value: float64Ptr(4)}, {Value: float64Ptr(2)}},
							Operation:  operators.OpDivide,
							Expression: 1,
						}, nil
					},
					SubmitResultFunc: func(ctx context.Context, completed *pb.TaskCompleted) (*pb.Empty, error) {
						assert.Equal(t, 2.0, completed.Result)
						return &pb.Empty{}, nil
					},
				}
			},
			wantErr: false,
		},
		{
			name: "calculation multiply",
			setupMock: func() *MockOrchestratorClient {
				return &MockOrchestratorClient{
					GetTaskFunc: func(ctx context.Context, empty *pb.Empty) (*pb.TaskResponse, error) {
						return &pb.TaskResponse{
							Id:         1,
							Args:       []*pb.WrappedDouble{{Value: float64Ptr(2)}, {Value: float64Ptr(3)}},
							Operation:  operators.OpPower,
							Expression: 1,
						}, nil
					},
					SubmitResultFunc: func(ctx context.Context, completed *pb.TaskCompleted) (*pb.Empty, error) {
						assert.Equal(t, 8.0, completed.Result)
						return &pb.Empty{}, nil
					},
				}
			},
			wantErr: false,
		},
		{
			name: "calculation unary minus",
			setupMock: func() *MockOrchestratorClient {
				return &MockOrchestratorClient{
					GetTaskFunc: func(ctx context.Context, empty *pb.Empty) (*pb.TaskResponse, error) {
						return &pb.TaskResponse{
							Id:         1,
							Args:       []*pb.WrappedDouble{{Value: float64Ptr(2)}, nil},
							Operation:  operators.OpUnaryMinus,
							Expression: 1,
						}, nil
					},
					SubmitResultFunc: func(ctx context.Context, completed *pb.TaskCompleted) (*pb.Empty, error) {
						assert.Equal(t, -2.0, completed.Result)
						return &pb.Empty{}, nil
					},
				}
			},
			wantErr: false,
		},
		{
			name: "calculation compiler error",
			setupMock: func() *MockOrchestratorClient {
				return &MockOrchestratorClient{
					GetTaskFunc: func(ctx context.Context, empty *pb.Empty) (*pb.TaskResponse, error) {
						return &pb.TaskResponse{
							Id:         1,
							Args:       []*pb.WrappedDouble{},
							Operation:  operators.OpMultiply,
							Expression: 1,
						}, nil
					},
					SubmitResultFunc: func(ctx context.Context, completed *pb.TaskCompleted) (*pb.Empty, error) {
						return &pb.Empty{}, nil
					},
				}
			},
			wantErr: true,
		},
		{
			name: "calculation error positive infinity",
			setupMock: func() *MockOrchestratorClient {
				return &MockOrchestratorClient{
					GetTaskFunc: func(ctx context.Context, empty *pb.Empty) (*pb.TaskResponse, error) {
						return &pb.TaskResponse{
							Id:         1,
							Args:       []*pb.WrappedDouble{{Value: float64Ptr(1000)}, {Value: float64Ptr(1000)}},
							Operation:  operators.OpPower,
							Expression: 1,
						}, nil
					},
					SubmitResultFunc: func(ctx context.Context, completed *pb.TaskCompleted) (*pb.Empty, error) {
						assert.Equal(t, "Результат - +Inf", completed.Error)
						return &pb.Empty{}, nil
					},
				}
			},
			wantErr: false,
		},
		{
			name: "calculation error positive infinity",
			setupMock: func() *MockOrchestratorClient {
				return &MockOrchestratorClient{
					GetTaskFunc: func(ctx context.Context, empty *pb.Empty) (*pb.TaskResponse, error) {
						return &pb.TaskResponse{
							Id:         1,
							Args:       []*pb.WrappedDouble{{Value: float64Ptr(-1000)}, {Value: float64Ptr(999)}},
							Operation:  operators.OpPower,
							Expression: 1,
						}, nil
					},
					SubmitResultFunc: func(ctx context.Context, completed *pb.TaskCompleted) (*pb.Empty, error) {
						assert.Equal(t, "Результат - -Inf", completed.Error)
						return &pb.Empty{}, nil
					},
				}
			},
			wantErr: false,
		},
		{
			name: "calculation error division by zero",
			setupMock: func() *MockOrchestratorClient {
				return &MockOrchestratorClient{
					GetTaskFunc: func(ctx context.Context, empty *pb.Empty) (*pb.TaskResponse, error) {
						return &pb.TaskResponse{
							Id:         1,
							Args:       []*pb.WrappedDouble{{Value: float64Ptr(1)}, {Value: float64Ptr(0)}},
							Operation:  operators.OpDivide,
							Expression: 1,
						}, nil
					},
					SubmitResultFunc: func(ctx context.Context, completed *pb.TaskCompleted) (*pb.Empty, error) {
						assert.Equal(t, "деление на ноль", completed.Error)
						return &pb.Empty{}, nil
					},
				}
			},
			wantErr: false,
		},
		{
			name: "calculation error nil first",
			setupMock: func() *MockOrchestratorClient {
				return &MockOrchestratorClient{
					GetTaskFunc: func(ctx context.Context, empty *pb.Empty) (*pb.TaskResponse, error) {
						return &pb.TaskResponse{
							Id:         1,
							Args:       []*pb.WrappedDouble{nil, {Value: float64Ptr(0)}},
							Operation:  operators.OpDivide,
							Expression: 1,
						}, nil
					},
					SubmitResultFunc: func(ctx context.Context, completed *pb.TaskCompleted) (*pb.Empty, error) {
						assert.Equal(t, "первый оператор не может быть nil", completed.Error)
						return &pb.Empty{}, nil
					},
				}
			},
			wantErr: false,
		},
		{
			name: "task submit error",
			setupMock: func() *MockOrchestratorClient {
				return &MockOrchestratorClient{
					GetTaskFunc: func(ctx context.Context, empty *pb.Empty) (*pb.TaskResponse, error) {
						return &pb.TaskResponse{
							Id:         1,
							Args:       []*pb.WrappedDouble{{Value: float64Ptr(0)}, {Value: float64Ptr(0)}},
							Operation:  operators.OpAdd,
							Expression: 1,
						}, nil
					},
					SubmitResultFunc: func(ctx context.Context, completed *pb.TaskCompleted) (*pb.Empty, error) {
						return nil, errors.New("error")
					},
				}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := tt.setupMock()
			_, cancel, errChan := newTestWorker(t, mockClient)
			defer cancel()

			select {
			case err := <-errChan:
				if !tt.wantErr {
					t.Fatalf("Unexpected error: %v", err)
				}
			case <-time.After(10 * time.Millisecond):
				if tt.wantErr {
					t.Error("Expected error but got none")
				}
			}
		})
	}
}
