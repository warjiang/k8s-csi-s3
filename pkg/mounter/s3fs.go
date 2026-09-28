package mounter

import (
	"fmt"
	"os"
	"strings"

	"github.com/yandex-cloud/k8s-csi-s3/pkg/s3"
)

// Implements Mounter
type s3fsMounter struct {
	meta          *s3.FSMeta
	url           string
	region        string
	pwFileContent string
}

const (
	s3fsCmd = "s3fs"
)

func newS3fsMounter(meta *s3.FSMeta, cfg *s3.Config) (Mounter, error) {
	return &s3fsMounter{
		meta:          meta,
		url:           cfg.Endpoint,
		region:        cfg.Region,
		pwFileContent: cfg.AccessKeyID + ":" + cfg.SecretAccessKey,
	}, nil
}

func (s3fs *s3fsMounter) Mount(target, volumeID string) error {
	if err := writes3fsPass(s3fs.pwFileContent); err != nil {
		return err
	}
	return fuseMountForeground(target, s3fsCmd, s3fs.mountArgs(target), nil)
}

func (s3fs *s3fsMounter) mountArgs(target string) []string {
	args := []string{
		fmt.Sprintf("%s:/%s", s3fs.meta.BucketName, s3fs.meta.Prefix),
		target,
		"-f",
		"-o", fmt.Sprintf("url=%s", s3fs.url),
		"-o", "allow_other",
		"-o", "mp_umask=000",
	}
	if !hasS3fsMountOption(s3fs.meta.MountOptions, "compat_dir") {
		args = append(args, "-o", "compat_dir")
	}
	if s3fs.region != "" {
		args = append(args, "-o", fmt.Sprintf("endpoint=%s", s3fs.region))
	}
	for i := 0; i < len(s3fs.meta.MountOptions); i++ {
		option := strings.TrimSpace(s3fs.meta.MountOptions[i])
		if option == "-o" && i+1 < len(s3fs.meta.MountOptions) &&
			strings.TrimSpace(s3fs.meta.MountOptions[i+1]) == "use_path_request_style" {
			i++
			continue
		}
		if option == "use_path_request_style" ||
			option == "-ouse_path_request_style" ||
			option == "-o=use_path_request_style" {
			continue
		}
		args = append(args, s3fs.meta.MountOptions[i])
	}
	return args
}

func hasS3fsMountOption(options []string, name string) bool {
	for i, option := range options {
		option = strings.TrimSpace(option)
		if option == "-o" && i+1 < len(options) && strings.TrimSpace(options[i+1]) == name {
			return true
		}
		if option == name || option == "-o"+name || option == "-o="+name {
			return true
		}
	}
	return false
}

func writes3fsPass(pwFileContent string) error {
	pwFileName := fmt.Sprintf("%s/.passwd-s3fs", os.Getenv("HOME"))
	pwFile, err := os.OpenFile(pwFileName, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	_, err = pwFile.WriteString(pwFileContent)
	if err != nil {
		return err
	}
	pwFile.Close()
	return nil
}
