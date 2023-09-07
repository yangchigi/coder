package main

import (
	"context"
	"encoding/json"
	"fmt"
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

	listVulns(dclient, jclient)
	for range ticker.C {
		listVulns(dclient, jclient)
	}
}

func listVulns(dclient *client.Client, jclient *jfroghttpclient.JfrogHttpClient) {
	containers := listCoderContainers(dclient)
	if len(containers) == 0 {
		fmt.Println("NO CONTAINERS, SKIPPING")
		return
	}
	results := fetchSecurityResults(jclient, defaultRepo)
	for _, container := range containers {
		ref, err := name.NewTag(container.Config.Image)
		must(err)
		repo := ref.Context().RepositoryStr() + ":" + ref.TagStr()
		result, ok := results[repo]
		if !ok {
			fmt.Println("no results!")
			return
		}
		cclient := agentsdk.New(&url.URL{
			Scheme: "https",
			Host:   mustAccessURL(),
		})
		cclient.SetSessionToken(agentToken(container))
		postMetadata(cclient, result.SecIssues.Critical, result.SecIssues.High)
		fmt.Println("Image: ", container.Config.Image)
		fmt.Println("\tCritical: ", result.SecIssues.Critical)
		fmt.Println("\tHigh: ", result.SecIssues.High)
		fmt.Println("\n")
	}
}

func postMetadata(cclient *agentsdk.Client, critical int, high int) {
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
	must(err)
}

func listCoderContainers(client *client.Client) []types.ContainerJSON {
	containers, err := client.ContainerList(context.Background(), types.ContainerListOptions{})
	must(err)

	filtered := make([]types.ContainerJSON, 0, len(containers))
	for _, container := range containers {
		inspect, err := client.ContainerInspect(context.Background(), container.ID)
		must(err)
		if token := agentToken(inspect); token == "" {
			fmt.Println("Skipping non-coder container ", inspect.Name)
			continue
		}
		filtered = append(filtered, inspect)
	}
	return filtered
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
func fetchSecurityResults(client *jfroghttpclient.JfrogHttpClient, repo string) map[string]artifact {
	path := fmt.Sprintf("https://cdr.jfrog.io/xray/api/v1/artifacts?repo=%s", repo)
	resp, body, _, err := client.SendGet(path, true, &httputils.HttpClientDetails{
		User:        mustUser(),
		AccessToken: mustAccessToken(),
	})
	must(err)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		panicf("unexpected status code %d", resp.StatusCode)
	}

	var response artifactsResponse
	err = json.Unmarshal(body, &response)
	must(err)

	artifacts := make(map[string]artifact, len(response.Data))
	for _, artifact := range response.Data {
		key := fmtKey(artifact)
		artifacts[key] = artifact
	}

	return artifacts
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
	client, err := client.NewClientWithOpts(client.FromEnv)
	must(err)
	return client
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
