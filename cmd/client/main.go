package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/lodthe/wiki-graph/pkg/wikigraphpb"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conf := ReadConfig()

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zlog.Logger = zlog.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	ctx, cancel := context.WithCancel(context.Background())
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	conn, err := createConnection(conf.GRPCServer)
	if err != nil {
		zlog.Fatal().Err(err).Str("address", conf.GRPCServer.Address).Msg("failed to connect to gRPC server")
	}
	defer conn.Close()

	go processRequests(ctx, wikigraphpb.NewWikiGraphClient(conn))

	<-stop
	cancel()
}

func createConnection(cfg GRPCServer) (*grpc.ClientConn, error) {
	conn, err := grpc.Dial(
		cfg.Address,
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(
			grpc_retry.WithMax(cfg.MaxRetries),
			grpc_retry.WithPerRetryTimeout(cfg.RetryTimeout),
		)),
	)

	return conn, err
}

func processRequests(ctx context.Context, cli wikigraphpb.WikiGraphClient) {
	reader := bufio.NewReader(os.Stdin)

	for {
		select {
		case <-ctx.Done():
			return

		default:
		}

		fmt.Println("Enter the title of the page you want to start from:")
		from, _ := reader.ReadString('\n')
		from = from[:len(from)-1]

		fmt.Println("Enter the title of the page you want to end at:")
		to, _ := reader.ReadString('\n')
		to = to[:len(to)-1]

		createTaskResponse, err := cli.FindShortestPath(ctx, &wikigraphpb.FindShortestPathRequest{
			From: from,
			To:   to,
		})
		if err != nil {
			fmt.Printf("[!] Failed to create a task: %v\n\n", err)
			continue
		}

		fmt.Printf("\nWaiting for task %s to complete...\n\n", createTaskResponse.GetTaskId().GetId())

		var task *wikigraphpb.Task
		for {
			time.Sleep(time.Second)

			resp, err := cli.GetTask(ctx, &wikigraphpb.GetTaskRequest{
				TaskId: createTaskResponse.GetTaskId(),
			})
			if err != nil {
				fmt.Printf("[!] Failed to get task status: %v\n\n", err)
				continue
			}

			if resp.GetTask().GetStatus() == wikigraphpb.Task_DONE {
				task = resp.GetTask()
				break
			}

			fmt.Printf("Current status: %s\n", resp.GetTask().GetStatus())
		}

		fmt.Printf("Task %s completed\n\n", task.GetId().GetId())

		fmt.Printf("The shortest path:\n")
		for _, url := range task.GetPath() {
			fmt.Printf("%s\n", url)
		}

		if len(task.GetPath()) == 0 {
			fmt.Printf("Unfortunately, the path was not found. Probably, the path is too long or you have typos in the provided page titles.\nTry Apple and Fruits as an example.")
		}

		fmt.Println()
		fmt.Println()
	}
}
