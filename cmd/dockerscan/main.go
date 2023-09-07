package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/jfrog/jfrog-client-go/config"
	"github.com/jfrog/jfrog-client-go/http/jfroghttpclient"
	"github.com/jfrog/jfrog-client-go/utils/io/httputils"
	"github.com/jfrog/jfrog-client-go/xray"
	"github.com/jfrog/jfrog-client-go/xray/auth"
	"golang.org/x/xerrors"

	"github.com/coder/coder/v2/codersdk/agentsdk"
)

const (
	defaultRepo         = "docker-local"
	defaultScanInterval = time.Second * 5
	defaultMetadataKey  = "99_image_vuln"
)

func main() {
	var (
		jclient = jfrogClient()
		dclient = dockerClient()
	)

	ticker := time.NewTicker(defaultScanInterval)
	defer ticker.Stop()

	err := listVulns(dclient, jclient)
	if err != nil {
		fmt.Printf("err: %v\n", err)
	}

	for range ticker.C {
		err := listVulns(dclient, jclient)
		if err != nil {
			fmt.Printf("err: %v\n", err)
		}
	}
}

func listVulns(dclient *client.Client, jclient *jfroghttpclient.JfrogHttpClient) error {
	containers, err := listCoderContainers(dclient)
	if err != nil {
		return xerrors.Errorf("list coder containers: %w", err)
	}

	if len(containers) == 0 {
		fmt.Println("NO CONTAINERS, SKIPPING")
		return nil
	}

	results, err := fetchSecurityResults(jclient, defaultRepo)
	if err != nil {
		return xerrors.Errorf("fetch results: %w", err)
	}

	for _, container := range containers {
		ref, err := name.NewTag(container.Config.Image)
		if err != nil {
			return xerrors.Errorf("new tag: %w", err)
		}
		repo := ref.Context().RepositoryStr() + ":" + ref.TagStr()
		result, ok := results[repo]
		if !ok {
			fmt.Println("no results!")
			return nil
		}
		accessurl := mustAccessURL()
		host, _, err := net.SplitHostPort(accessurl)
		if err != nil {
			return xerrors.Errorf("split host post %s: %w", accessurl, err)
		}

		scheme := "https"
		if strings.Contains(host, "localhost") {
			scheme = "http"
		}

		cclient := agentsdk.New(&url.URL{
			Scheme: scheme,
			Host:   accessurl,
		})
		cclient.SetSessionToken(agentToken(container))
		err = postMetadata(cclient, result.SecIssues.Critical, result.SecIssues.High)
		if err != nil {
			return err
		}
		fmt.Println("Image: ", container.Config.Image)
		fmt.Println("\tCritical: ", result.SecIssues.Critical)
		fmt.Println("\tHigh: ", result.SecIssues.High)
		fmt.Println("\n")
	}
	return nil
}

func postMetadata(cclient *agentsdk.Client, critical int, high int) error {
	var errStr string
	var value string
	if critical > 0 || high > 0 {
		errStr = fmt.Sprintf("Crit(%d) High(%d)", critical, high)
		value = errStr
	} else {
		value = "None"
	}
	fmt.Println("error: ", errStr)
	fmt.Println("value: ", value)
	err := cclient.PostMetadata(context.Background(), defaultMetadataKey, agentsdk.PostMetadataRequest{
		CollectedAt: time.Now(),
		Age:         0,
		Value:       value,
		Error:       errStr,
	})
	if err != nil {
		return xerrors.Errorf("post metadata: %w", err)
	}
	return nil
}

func listCoderContainers(client *client.Client) ([]types.ContainerJSON, error) {
	containers, err := client.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		return nil, xerrors.Errorf("container list: %w", err)
	}

	filtered := make([]types.ContainerJSON, 0, len(containers))
	for _, container := range containers {
		inspect, err := client.ContainerInspect(context.Background(), container.ID)
		if err != nil {
			return nil, xerrors.Errorf("container inspect: %w", err)
		}
		if token := agentToken(inspect); token == "" {
			fmt.Println("Skipping non-coder container ", inspect.Name)
			continue
		}
		filtered = append(filtered, inspect)
	}
	return filtered, nil
}

func agentToken(c types.ContainerJSON) string {
	for _, env := range c.Config.Env {
		if strings.HasPrefix(env, "CODER_AGENT_TOKEN=") {
			idx := strings.Index(env, "=")
			return env[idx+1:]
		}
	}
	return ""
}

// fetchSecurityResults fetches results for images in a repo
func fetchSecurityResults(client *jfroghttpclient.JfrogHttpClient, repo string) (map[string]artifact, error) {
	path := fmt.Sprintf("https://cdr.jfrog.io/xray/api/v1/artifacts?repo=%s", repo)
	resp, body, _, err := client.SendGet(path, true, &httputils.HttpClientDetails{
		User:        mustUser(),
		AccessToken: mustAccessToken(),
	})
	if err != nil {
		return nil, xerrors.Errorf("send get: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, xerrors.Errorf("unexpected status code %d", resp.StatusCode)
	}

	var response artifactsResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, xerrors.Errorf("unmarshal: %w", err)
	}

	artifacts := make(map[string]artifact, len(response.Data))
	for _, artifact := range response.Data {
		key := fmtKey(artifact)
		artifacts[key] = artifact
	}

	return artifacts, nil
}

func fmtKey(a artifact) string {
	key := a.RepoFullPath
	// Strip manifest.json
	key = filepath.Dir(key)
	lastSlash := strings.LastIndex(key, "/")
	// e.g. Replace /latest -> :latest
	return key[:lastSlash] + ":" + key[lastSlash+1:]
}

type artifactsResponse struct {
	Data   []artifact `json:"data"`
	Offset int        `json:"offset"`
}

type artifact struct {
	Name         string    `json:"name"`
	RepoPath     string    `json:"repo_path"`
	PackageID    string    `json:"package_id"`
	Version      string    `json:"version"`
	SecIssues    SecIssues `json:"sec_issues"`
	Size         string    `json:"size"`
	Violations   int       `json:"violations"`
	Created      time.Time `json:"created"`
	DeployedBy   string    `json:"deployed_by"`
	RepoFullPath string    `json:"repo_full_path"`
}

type SecIssues struct {
	Critical int `json:"critical"`
	High     int `json:"high"`
	Medium   int `json:"medium"`
	Low      int `json:"low"`
	Total    int `json:"total"`
}

func jfrogClient() *jfroghttpclient.JfrogHttpClient {
	details := auth.NewXrayDetails()
	details.SetAccessToken(mustAccessToken())
	details.SetUser(mustUser())
	details.SetUrl("https://cdr.jfrog.io")
	conf, err := config.NewConfigBuilder().SetServiceDetails(details).Build()
	must(err)
	mgr, err := xray.New(conf)
	must(err)
	return mgr.Client()
}

func dockerClient() *client.Client {
	dclient, err := client.NewClientWithOpts(client.FromEnv)
	must(err)
	return dclient
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func mustUser() string {
	user := os.Getenv("JFROG_USER")
	if user == "" {
		panic("must set JFROG_USER")
	}

	return user

}

func mustAccessToken() string {
	token := os.Getenv("JFROG_ACCESS_TOKEN")
	if token == "" {
		panic("must set JFROG_ACCESS_TOKEN")
	}

	return token
}

func mustAccessURL() string {
	host := os.Getenv("TEST_HOST")
	if host == "" {
		panic("must set TEST_HOST")
	}

	return host
}

func panicf(msg string, args ...interface{}) {
	panic(fmt.Sprintf(msg, args...))
}
