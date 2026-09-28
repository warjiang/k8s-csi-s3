package mounter

import (
	"slices"
	"testing"

	"github.com/yandex-cloud/k8s-csi-s3/pkg/s3"
)

func TestS3fsMountArgsUseVirtualHostedStyle(t *testing.T) {
	mounter, err := newS3fsMounter(&s3.FSMeta{
		BucketName:   "example-bucket",
		Prefix:       "prefix",
		MountOptions: []string{"-o", "use_path_request_style", "-o", "nonempty"},
	}, &s3.Config{
		Endpoint: "https://oss-example.aliyuncs.com",
		Region:   "oss-example",
	})
	if err != nil {
		t.Fatal(err)
	}

	args := mounter.(*s3fsMounter).mountArgs("/target")
	if !slices.Contains(args, "-f") {
		t.Fatalf("s3fs must run in the foreground: %v", args)
	}
	if slices.Contains(args, "use_path_request_style") {
		t.Fatalf("s3fs must not force path-style requests: %v", args)
	}
	if slices.Contains(args, "-ouse_path_request_style") || slices.Contains(args, "-o=use_path_request_style") {
		t.Fatalf("s3fs must not force path-style requests: %v", args)
	}
	if slices.Contains(args, "-o") && slices.Index(args, "-o") == len(args)-1 {
		t.Fatalf("s3fs options must remain valid: %v", args)
	}
	if !hasS3fsMountOption(args, "compat_dir") {
		t.Fatalf("s3fs must enable compat_dir by default: %v", args)
	}
}

func TestS3fsMountArgsDoesNotDuplicateCompatDir(t *testing.T) {
	mounter, err := newS3fsMounter(&s3.FSMeta{
		BucketName:   "example-bucket",
		MountOptions: []string{"-o", "compat_dir"},
	}, &s3.Config{})
	if err != nil {
		t.Fatal(err)
	}

	args := mounter.(*s3fsMounter).mountArgs("/target")
	count := 0
	for i, arg := range args {
		if arg == "compat_dir" && i > 0 && args[i-1] == "-o" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("s3fs must pass compat_dir once: %v", args)
	}
}
