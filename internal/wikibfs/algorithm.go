package wikibfs

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lodthe/wiki-graph/pkg/wikiclient"
	zlog "github.com/rs/zerolog/log"
)

type BFSConfig struct {
	// Maximum allowed distance from the root node.
	DistanceThreshold uint

	// Number of workers to parse pages.
	WorkerCount int
}

type parseResult struct {
	title           string
	mentionedTitles []string

	err error
}

type algorithm struct {
	wikiClient *wikiclient.Client
	cfg        BFSConfig

	visited map[string]struct{}
	prev    map[string]string
}

func newAlgorithm(wikiClient *wikiclient.Client, cfg BFSConfig) *algorithm {
	return &algorithm{
		wikiClient: wikiClient,
		cfg:        cfg,
		visited:    make(map[string]struct{}),
		prev:       make(map[string]string),
	}
}

func (a *algorithm) findShortestPath(taskID uuid.UUID, from, to string) ([]string, error) {
	pagesToParse := make(chan string, 1024)
	parseResults := make(chan parseResult, 1024)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer close(pagesToParse)

	for range make([]struct{}, a.cfg.WorkerCount) {
		go a.parseWorker(ctx, pagesToParse, parseResults)
	}

	to = a.normalize(to)
	from = a.normalize(from)

	a.visited[from] = struct{}{}
	queue := []string{from}

	var reached bool
	var distance uint
	for {
		_, reached = a.visited[to]
		if reached {
			break
		}

		distance++
		if distance > a.cfg.DistanceThreshold {
			break
		}

		startedAt := time.Now()
		zlog.Info().Int("queue_length", len(queue)).Msg("started a new BFS iteration")

		go func() {
			for _, title := range queue {
				select {
				case <-ctx.Done():
					return
				default:
				}

				pagesToParse <- title
			}
		}()

		newQueue := make([]string, 0, len(queue))
		for range queue {
			result := <-parseResults
			if result.err != nil {
				zlog.Error().Err(result.err).Fields(map[string]interface{}{
					"task_id":           taskID.String(),
					"from":              from,
					"to":                to,
					"current_distance":  distance,
					"faulty_page_title": result.title,
				}).Msg("page cannot be parsed")

				continue
			}

			for _, title := range result.mentionedTitles {
				title = a.normalize(title)
				_, visited := a.visited[title]
				if visited {
					continue
				}

				a.visited[title] = struct{}{}
				a.prev[title] = result.title
				newQueue = append(newQueue, title)
			}
		}

		queue = newQueue

		zlog.Info().
			Dur("elapsed", time.Since(startedAt)).
			Uint("distance", distance).
			Msgf("found %d new pages", len(queue))
	}

	if !reached {
		zlog.Info().Str("from", from).Str("to", to).Msg("page is not reachable")
		return nil, nil
	}

	zlog.Info().Fields(map[string]interface{}{
		"task_id":  taskID.String(),
		"from":     from,
		"to":       to,
		"distance": distance,
	}).Msg("BFS finished successfully")

	path := make([]string, 0, distance+1)
	for to != "" {
		path = append(path, to)
		to = a.prev[to]
	}

	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}

	return path, nil
}

func (a *algorithm) normalize(s string) string {
	return strings.ToLower(s)
}

func (a *algorithm) parseWorker(ctx context.Context, pageTitles <-chan string, results chan<- parseResult) {
	for title := range pageTitles {
		mentioned, err := a.wikiClient.GetMentionedPages(title)

		select {
		case <-ctx.Done():
			return
		case results <- parseResult{
			title:           title,
			mentionedTitles: mentioned,
			err:             err,
		}:
		}
	}
}
