package main

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/go-github/v32/github"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v2"
)

type config struct {
	Server struct {
		Host    string `yaml:"host"`
		Port    string `yaml:"port"`
		Timeout struct {
			Server time.Duration `yaml:"server"`
			Write  time.Duration `yaml:"write"`
			Read   time.Duration `yaml:"read"`
			Idle   time.Duration `yaml:"idle"`
		} `yaml:"timeout"`
	} `yaml:"server"`
}

func newConfig(configPath string) (*config, error) {
	config := &config{}
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	d := yaml.NewDecoder(file)
	if err := d.Decode(&config); err != nil {
		return nil, err
	}

	return config, nil
}

type dashpageData struct {
	Commits []*github.RepositoryCommit
}

func getGithubClient(ctx context.Context) *github.Client {
	ghPat := os.Getenv("GH_PAT")

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: ghPat},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	return client
}

func getChangelog(ctx context.Context, client *github.Client) (*github.CommitsComparison, error) {
	refFrom, _, err := client.Git.GetRef(ctx, "lobsterdore", "lobstercms", "tags/v2.1.0")
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	refTo, _, err := client.Git.GetRef(ctx, "lobsterdore", "lobstercms", "tags/v2.7.0")
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	comparison, _, err := client.Repositories.CompareCommits(ctx, "lobsterdore", "lobstercms", *refFrom.Object.SHA, *refTo.Object.SHA)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	return comparison, nil
}

func newRouter() *http.ServeMux {
	router := http.NewServeMux()

	router.HandleFunc("/", func(respWriter http.ResponseWriter, request *http.Request) {
		log.Printf("Requested - '/' ")
		ctx := request.Context()
		tmpl := template.Must(template.ParseFiles("layout.html"))

		client := getGithubClient(ctx)
		comparison, err := getChangelog(ctx, client)
		if err != nil {
			log.Fatal(err)
		}

		data := dashpageData{
			Commits: comparison.Commits,
		}

		tmpl.Execute(respWriter, data)
	})

	return router
}

func main() {
	cfg, err := newConfig("./config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	var runChan = make(chan os.Signal, 1)
	ctx, cancel := context.WithTimeout(
		context.Background(),
		cfg.Server.Timeout.Server,
	)
	defer cancel()s

	server := &http.Server{
		Addr:         cfg.Server.Host + ":" + cfg.Server.Port,
		Handler:      newRouter(),
		ReadTimeout:  cfg.Server.Timeout.Read * time.Second,
		WriteTimeout: cfg.Server.Timeout.Write * time.Second,
		IdleTimeout:  cfg.Server.Timeout.Idle * time.Second,
	}

	signal.Notify(runChan, os.Interrupt, syscall.SIGTSTP)

	log.Printf("Server is starting on %s\n", server.Addr)

	go func() {
		if err := server.ListenAndServe(); err != nil {
			if err == http.ErrServerClosed {
			} else {
				log.Fatalf("Server failed to start due to err: %v", err)
			}
		}
	}()

	interrupt := <-runChan
	log.Printf("Server is shutting down due to %+v\n", interrupt)
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server was unable to gracefully shutdown due to err: %+v", err)
	}
}
