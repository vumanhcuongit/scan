package gitscan

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/go-github/v47/github"
	"github.com/vumanhcuongit/scan/pkg/models"
	"go.uber.org/zap"
)

const (
	ruleID                    = "G101"
	typeSast                  = "sast"
	exampleFindingDescription = "Potential hardcoded credentials"
	exampleFindingSeverity    = "HIGH"
	secretPrivateKey          = "private_key"
	secretPublicKey           = "public_key"
)

type GitScan struct {
	sourceCodesDir string // directory contains repository's code
	githubClient   *github.Client
	httpClient     *http.Client
}

func NewGitScan(sourcesCodeDir string) *GitScan {
	httpClient := &http.Client{Timeout: 2 * time.Minute}
	githubClient := github.NewClient(httpClient)
	return &GitScan{
		sourceCodesDir: sourcesCodeDir,
		githubClient:   githubClient,
		httpClient:     httpClient,
	}
}

func (g *GitScan) Scan(
	ctx context.Context,
	ownerName string,
	repoName string,
) ([]models.Finding, error) {
	log := zap.S()
	log.Infof("starting to scan repository, owner name %s, repo name %s", ownerName, repoName)

	url, _, err := g.githubClient.Repositories.GetArchiveLink(
		ctx, ownerName, repoName, github.Tarball,
		&github.RepositoryContentGetOptions{}, true,
	)
	if err != nil {
		log.Warnf("failed to get archive link, err: %+v", err)
		return nil, err
	}

	err = g.downloadAndUntar(ctx, url.String(), g.sourceCodesDir)
	if err != nil {
		log.Warnf("failed to download tarball, err: %+v", err)
		return nil, err
	}

	repoFolderName := ""
	err = filepath.Walk(g.sourceCodesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if strings.HasPrefix(info.Name(), fmt.Sprintf("%s-%s", ownerName, repoName)) {
				log.Infof("found repo directory: %s", info.Name())
				repoFolderName = info.Name()
				return io.EOF
			}
			return nil
		}

		return nil
	})
	if err != nil && err != io.EOF {
		log.Warnf("failed to read file, err: %+v", err)
		return nil, err
	}

	if repoFolderName == "" {
		log.Warnf("empty repo directory")
		return nil, errors.New("empty repo directory")
	}

	repoDir := path.Join(g.sourceCodesDir, repoFolderName)
	defer os.RemoveAll(repoDir)
	findings := []models.Finding{}
	err = filepath.Walk(repoDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			log.Fatal(err)
		}
		defer func() {
			if err = f.Close(); err != nil {
				log.Fatal(err)
			}
		}()
		s := bufio.NewScanner(f)

		lineNumber := 0
		for s.Scan() {
			lineNumber++
			containPrivateKey := strings.HasPrefix(s.Text(), secretPrivateKey)
			containPublicKey := strings.HasPrefix(s.Text(), secretPublicKey)
			if !containPrivateKey && !containPublicKey {
				continue
			}

			extractedPath := strings.Join(strings.Split(path, "/")[2:], "/")
			findings = append(findings, models.Finding{
				Type:   typeSast,
				RuleID: ruleID,
				Location: models.Location{
					Path: extractedPath,
					Position: models.Position{
						Begin: models.Begin{
							Line: lineNumber,
						},
					},
				},
				Metadata: models.Metadata{
					Description: exampleFindingDescription,
					Severity:    exampleFindingSeverity,
				},
			})
		}
		err = s.Err()
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		log.Warnf("failed to scan source code, err: %+v", err)
		return nil, err
	}

	return findings, nil
}

func (g *GitScan) downloadAndUntar(ctx context.Context, downloadURL string, destPath string) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return err
	}
	resp, err := g.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("received non 200 response code")
	}

	_, err = g.untarWithIOReader(ctx, resp.Body, destPath)
	if err != nil {
		return err
	}

	return nil
}

func (g *GitScan) untarWithIOReader(
	ctx context.Context,
	tarFile io.ReadCloser,
	destination string,
) (string, error) {
	directory := ""

	gz, err := gzip.NewReader(tarFile)
	if err != nil {
		return directory, err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)

	absPath, err := filepath.Abs(destination)
	if err != nil {
		return directory, err
	}

	// untar each segment
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return directory, err
		}
		// determine proper file path info
		finfo := hdr.FileInfo()
		fileName := hdr.Name
		absFileName := filepath.Join(absPath, fileName)
		if finfo.Mode().IsDir() {
			if err = os.MkdirAll(absFileName, os.ModePerm); err != nil {
				return directory, err
			}

			continue
		}

		if directory == "" && filepath.Base(fileName) == "package.json" {
			directory = filepath.Dir(absFileName)
		}

		// Creating the files in the target directory
		if err = os.MkdirAll(filepath.Dir(absFileName), os.ModePerm); err != nil {
			return directory, err
		}

		// create new file with original file mode
		file, err := os.OpenFile(
			absFileName,
			os.O_RDWR|os.O_CREATE|os.O_TRUNC,
			finfo.Mode().Perm(),
		)
		if err != nil {
			return directory, err
		}
		defer file.Close()

		_, cpErr := io.Copy(file, tr)
		if cpErr != nil && cpErr != io.EOF {
			return directory, cpErr
		}
	}
	return directory, nil
}
