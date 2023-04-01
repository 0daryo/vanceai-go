//go:build integration

package vanceai

import (
	"context"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestClient_Upload(t *testing.T) {
	cli, err := NewClient(os.Getenv("API_KEY"))
	if err != nil {
		t.Fatal(err)
	}
	image, err := os.Open("testdata/cat.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer image.Close()
	resp, err := cli.UploadImage(context.Background(), image, "cat.jpg")
	if err != nil {
		t.Fatal(err)
	}
	want := Response{
		Code:   200,
		CSCode: 200,
		Data: Data{
			Name:      "cat.jpg",
			Thumbnail: "",
			W:         1200,
			H:         1199,
			FileSize:  98052,
		},
	}
	if diff := cmp.Diff(want, resp,
		cmpopts.IgnoreFields(
			Response{}, "IP",
		),
		cmpopts.IgnoreFields(
			Data{}, "UID",
		),
	); diff != "" {
		t.Errorf("response mismatch (-want +got):\n%s", diff)
	}
}
