package s3

import (
	"testing"

	"github.com/minio/minio-go/v7"
)

func TestNewClientFromSecretBucketLookup(t *testing.T) {
	tests := []struct {
		value string
		want  minio.BucketLookupType
	}{
		{"", minio.BucketLookupAuto},
		{"auto", minio.BucketLookupAuto},
		{"path", minio.BucketLookupPath},
		{"dns", minio.BucketLookupDNS},
	}
	for _, test := range tests {
		client, err := NewClientFromSecret(map[string]string{
			"endpoint":     "https://example.com",
			"bucketLookup": test.value,
		})
		if err != nil {
			t.Fatalf("bucketLookup=%q: %v", test.value, err)
		}
		if client.Config.BucketLookup != test.want {
			t.Fatalf("bucketLookup=%q: got %v, want %v", test.value, client.Config.BucketLookup, test.want)
		}
	}
}

func TestNewClientFromSecretRejectsInvalidBucketLookup(t *testing.T) {
	_, err := NewClientFromSecret(map[string]string{
		"endpoint":     "https://example.com",
		"bucketLookup": "invalid",
	})
	if err == nil {
		t.Fatal("expected invalid bucketLookup to fail")
	}
}
