package wikigraphserver

import (
	"context"

	"github.com/google/uuid"
	"github.com/lodthe/wiki-graph/internal/pathtask"
	"github.com/lodthe/wiki-graph/internal/taskqueue"
	"github.com/lodthe/wiki-graph/pkg/wikigraphpb"
	"github.com/pkg/errors"
	zlog "github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	wikigraphpb.UnimplementedWikiGraphServer

	repo     pathtask.Repository
	producer *taskqueue.Producer
}

func New(repo pathtask.Repository, producer *taskqueue.Producer) *Server {
	return &Server{
		repo:     repo,
		producer: producer,
	}
}

func (s *Server) FindShortestPath(_ context.Context, in *wikigraphpb.FindShortestPathRequest) (*wikigraphpb.FindShortestPathResponse, error) {
	if in.GetFrom() == "" {
		return nil, status.Error(codes.InvalidArgument, "from is empty")
	}

	if in.GetTo() == "" {
		return nil, status.Error(codes.InvalidArgument, "to is empty")
	}

	task, err := s.repo.Create(in.GetFrom(), in.GetTo())
	if err != nil {
		zlog.Error().Err(err).Fields(map[string]interface{}{
			"from": in.GetFrom(),
			"to":   in.GetTo(),
		}).Msg("failed to create task")

		return nil, status.Error(codes.Internal, err.Error())
	}

	zlog.Info().Fields(map[string]interface{}{
		"id":   task.ID.String(),
		"from": task.From,
		"to":   task.To,
	}).Msg("created a new task")

	err = s.producer.Produce(taskqueue.Task{ID: task.ID})
	if err != nil {
		zlog.Error().Err(err).Str("id", task.ID.String()).Msg("failed to publish")
		return nil, status.Error(codes.Internal, errors.Wrap(err, "failed to enqueue the task").Error())
	}

	return &wikigraphpb.FindShortestPathResponse{
		TaskId: &wikigraphpb.TaskId{
			Id: task.ID.String(),
		},
	}, nil
}

func (s *Server) GetTask(_ context.Context, in *wikigraphpb.GetTaskRequest) (*wikigraphpb.GetTaskResponse, error) {
	if in.GetTaskId().GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "task_id is empty")
	}

	id, err := uuid.Parse(in.GetTaskId().GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, errors.Wrap(err, "task_id is invalid").Error())
	}

	task, err := s.repo.Get(id)
	if errors.Is(err, pathtask.ErrNotFound) {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	if errors.Is(err, pathtask.ErrNotFound) {
		zlog.Error().Err(err).Str("id", id.String()).Msg("failed to get task")
		return nil, status.Error(codes.Internal, "failed to find a task")
	}

	return &wikigraphpb.GetTaskResponse{
		Task: s.taskToProto(task),
	}, nil
}

func (s *Server) taskToProto(task *pathtask.Task) *wikigraphpb.Task {
	converted := &wikigraphpb.Task{
		Id: &wikigraphpb.TaskId{
			Id: task.ID.String(),
		},
		From: task.From,
		To:   task.To,
	}
	if task.Result != nil {
		converted.Path = task.Result.ShortestPath
	}

	switch task.Status {
	case pathtask.StatusPending:
		converted.Status = wikigraphpb.Task_PENDING

	case pathtask.StatusProcessing:
		converted.Status = wikigraphpb.Task_PROCESSING

	case pathtask.StatusDone:
		converted.Status = wikigraphpb.Task_DONE

	default:
		zlog.Error().Fields(map[string]interface{}{
			"id":     task.ID.String(),
			"status": task.Status,
		}).Msg("unknown task status")
	}

	return converted
}
