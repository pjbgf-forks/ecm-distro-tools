// package images exports k3s container images into tarballs.
// This is needed for Rancher Manager's airgap support and
// replaces docker/podman save.
// Note that both docker and podman do not support the execution
// of their commands during a container image building process.
package images

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/containers/common/libimage"
	"github.com/containers/common/pkg/config"
	"github.com/containers/image/v5/types"
	"github.com/containers/storage"
	"github.com/containers/storage/pkg/reexec"
)

var (
	maxDownloadSize int64 = 1 << 20 // 1MB
	requestTimeout        = 30 * time.Second

	// Same as upstream default policy, which translates to no
	// image signatures being verified.
	noSignaturePolicy = `{"default":[{"type":"insecureAcceptAnything"}]}`

	// requiredImages defines the list of images that must be saved
	// into the output tar.
	requiredImages = map[string]struct{}{
		"docker.io/rancher/mirrored-pause":           struct{}{},
		"docker.io/rancher/mirrored-coredns-coredns": struct{}{},
	}

	k3sVersion = regexp.MustCompile(`^v\d+\.\d+\.\d+\+k3s\d+$`)

	fetcher = fetch
)

func fetch(url string) (io.ReadCloser, error) {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	fmt.Println("fetching", url)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create request: %w", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request unsuccessful: %w", err)
	}

	return res.Body, nil
}

func k3sImages(version string) ([]string, error) {
	if !k3sVersion.MatchString(version) {
		return nil, fmt.Errorf("invalid k3s version: %s", version)
	}

	url := fmt.Sprintf("https://github.com/rancher/k3s/releases/download/%s/k3s-images.txt", version)
	body, err := fetcher(url)
	if err != nil {
		return nil, err
	}
	defer io.Copy(io.Discard, body)

	images := []string{}
	scanner := bufio.NewScanner(io.LimitReader(body, maxDownloadSize))
	for scanner.Scan() {
		fqn := scanner.Text()
		if _, ok := requiredImages[strings.Split(fqn, ":")[0]]; ok {
			images = append(images, fqn)
		}
	}

	err = scanner.Err()
	if err != nil {
		return nil, fmt.Errorf("error processing k3s-images.txt file: %w", err)
	}

	return images, nil
}

func setupStorage() (string, error) {
	tmp, err := os.MkdirTemp("", "")
	if err != nil {
		return "", err
	}

	err = os.WriteFile(filepath.Join(tmp, "policy.json"),
		[]byte(noSignaturePolicy), 0o600)
	if err != nil {
		return "", err
	}

	return tmp, nil
}

func Save(version, output string) error {
	if reexec.Init() {
		return nil
	}

	path, err := setupStorage()
	if err != nil {
		return err
	}

	store, err := storage.GetStore(storage.StoreOptions{
		GraphRoot: path,
	})
	if err != nil {
		return err
	}
	defer os.RemoveAll(path)

	runtime, err := libimage.RuntimeFromStore(store, &libimage.RuntimeOptions{
		SystemContext: &types.SystemContext{
			SignaturePolicyPath: filepath.Join(path, "policy.json"),
		},
	})
	if err != nil {
		return err
	}

	imgs, err := k3sImages(version)
	if err != nil {
		return err
	}

	for _, img := range imgs {
		fmt.Printf("pulling image %s\n", img)
		_, err = runtime.Pull(context.TODO(), img, config.PullPolicyMissing, nil)
		if err != nil {
			return err
		}
	}

	return runtime.Save(context.TODO(), imgs, "docker-archive", output, nil)
}
