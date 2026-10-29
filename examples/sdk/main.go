// Command sdk demonstrates the unified yatori SDK.
//
// Configure the environment before running:
//
//	YATORI_PLATFORM=XUEXITONG
//	YATORI_ACCOUNT=ACCOUNT
//	YATORI_PASSWORD=PASSWORD
//	YATORI_URL=            # required by some platforms
//	YATORI_SCHOOL=         # required by WEIBAN
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/yatori-dev/yatori-go-core"
)

func main() {
	platform, err := yatori.ParsePlatform(os.Getenv("YATORI_PLATFORM"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	client, err := yatori.New(yatori.Config{
		Platform: platform,
		Account: yatori.Account{
			Username: os.Getenv("YATORI_ACCOUNT"),
			Password: os.Getenv("YATORI_PASSWORD"),
			URL:      os.Getenv("YATORI_URL"),
			School:   os.Getenv("YATORI_SCHOOL"),
		},
		Options: yatori.Options{
			ProxyURL: os.Getenv("YATORI_PROXY"),
		},
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	if err := client.Login(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	courses, err := client.Courses(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for _, course := range courses {
		fmt.Printf("[%s] %s (%.1f%%)\n", course.ID, course.Name, course.Progress)
		nodes, err := client.Tree(ctx, course)
		if err != nil {
			fmt.Fprintf(os.Stderr, "tree %s: %v\n", course.Name, err)
			continue
		}
		for _, node := range yatori.Flatten(nodes) {
			if node.Completed || node.Kind == yatori.NodeChapter {
				continue
			}
			result, err := client.Study(ctx, yatori.StudyRequest{
				Course:   course,
				Node:     node,
				Action:   yatori.StudyActionProgress,
				Mode:     yatori.StudyModeNormal,
				Position: 60 * time.Second,
			})
			if err != nil {
				fmt.Fprintf(os.Stderr, "study %s: %v\n", node.Name, err)
				continue
			}
			fmt.Printf("study %s -> %.1f%%\n", node.Name, result.Progress)
		}
	}
}
