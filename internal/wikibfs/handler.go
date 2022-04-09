package wikibfs

import (
	"github.com/google/uuid"
	"github.com/lodthe/wiki-graph/internal/pathtask"
	"github.com/pkg/errors"
	zlog "github.com/rs/zerolog/log"
)

type Handler struct {
	repository pathtask.Repository
}

func NewHandler(repo pathtask.Repository) *Handler {
	return &Handler{
		repository: repo,
	}
}

func (h *Handler) HandleTask(taskID uuid.UUID) error {
	task, err := h.repository.Get(taskID)
	if err != nil {
		zlog.Error().Err(err).Str("id", taskID.String()).Msg("task cannot be fetched")
		return errors.Wrap(err, "fetch failed")
	}

	if task.Status != pathtask.StatusPending {
		zlog.Info().Str("id", task.ID.String()).Uint("status", uint(task.Status)).Msg("task has invalid status")
		return nil
	}

	zlog.Info().Str("id", taskID.String()).Msg("start processing task")

	err = h.repository.UpdateStatus(task.ID, pathtask.StatusPending, pathtask.StatusProcessing)
	if err != nil {
		zlog.Error().Err(err).Str("id", taskID.String()).Msg("failed to update task status to PROCESSING")
		return errors.Wrap(err, "failed to update status")
	}

	algo := newAlgorithm()
	path, err := algo.findShortestPath(task.ID, task.From, task.To)
	if err != nil {
		zlog.Error().Err(err).Str("id", taskID.String()).Msg("algorithm failed")
		return errors.Wrap(err, "algorithm failed")
	}

	err = h.repository.SetResult(task.ID, &pathtask.Result{ShortestPath: path})
	if err != nil {
		zlog.Error().Err(err).Fields(map[string]interface{}{
			"id":   task.ID.String(),
			"path": path,
		}).Msg("failed to set result")

		return errors.Wrap(err, "setting result failed")
	}

	err = h.repository.UpdateStatus(task.ID, pathtask.StatusProcessing, pathtask.StatusDone)
	if err != nil {
		zlog.Error().Err(err).Str("id", taskID.String()).Msg("failed to update task status to DONE")
		return errors.Wrap(err, "failed to update status")
	}

	return nil
}
