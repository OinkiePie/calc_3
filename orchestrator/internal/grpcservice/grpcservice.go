package grpcservice

import (
	"context"
	"github.com/OinkiePie/calc_3/orchestrator/internal/managers"
	"github.com/OinkiePie/calc_3/orchestrator/internal/providers"
	"github.com/OinkiePie/calc_3/pkg/models"
	pb "github.com/OinkiePie/calc_3/pkg/proto"
)

type OrchestratorGRPCServer struct {
	pb.UnimplementedOrchestratorServiceServer
	exprManager managers.ExpressionManagerInterface
}

func NewOrchestratorGRPCServer(provider *providers.Providers) *OrchestratorGRPCServer {
	return &OrchestratorGRPCServer{exprManager: provider.ExprManager}
}

func (s *OrchestratorGRPCServer) GetTask(
	ctx context.Context,
	in *pb.Empty,
) (*pb.TaskResponse, error) {
	task, err, _ := s.exprManager.ReadTask(ctx)
	if task == nil {
		return nil, err
	}

	pbArgs := make([]*pb.WrappedDouble, 2)
	for i, ptr := range task.Args {
		pbArgs[i] = &pb.WrappedDouble{
			Value: ptr,
		}
	}

	response := &pb.TaskResponse{
		Id:         task.ID,
		Args:       pbArgs,
		Operation:  task.Operation,
		Expression: task.Expression,
	}

	return response, nil
}

func (s *OrchestratorGRPCServer) SubmitResult(
	ctx context.Context,
	in *pb.TaskCompleted,
) (*pb.Empty, error) {
	completed := &models.TaskCompleted{
		Result:     in.Result,
		ID:         in.Id,
		Expression: in.Expression,
		Error:      in.Error,
	}

	err, _ := s.exprManager.CompleteTask(ctx, completed)
	return &pb.Empty{}, err
}
