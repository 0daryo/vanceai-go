//go:build integration

package vanceai

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// first, you need to set API_KEY environment variable
// https://vanceai.com/ja/my-account/api/
func TestClient_Upload_Process_Check_Download(t *testing.T) {
	cli, err := NewClient(os.Getenv("API_KEY"), "")
	if err != nil {
		t.Fatal(err)
	}
	image, err := os.Open("testdata/cat.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer image.Close()

	// upload image
	uresp, err := cli.UploadImage(context.Background(), image, "cat.jpg")
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
	if diff := cmp.Diff(want, uresp,
		cmpopts.IgnoreFields(
			Response{}, "IP",
		),
		cmpopts.IgnoreFields(
			Data{}, "UID",
		),
	); diff != "" {
		t.Errorf("response mismatch (-want +got):\n%s", diff)
	}

	// process image
	presp, err := cli.ProcessImage(context.Background(), uresp.Data.UID, &JobConfig{
		Job: "enlarge",
		Config: Config{
			Module: "enlarge",
			ModuleParams: ModuleParams{
				ModelName:     "EnlargeStable",
				SuppressNoise: 26,
				RemoveBlur:    26,
				Scale:         "2x",
			},
			OutParams: OutParams{},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	want = Response{
		Code:   200,
		CSCode: 200,
		Data: Data{
			Status: "finish",
		},
	}
	if diff := cmp.Diff(want, presp,
		cmpopts.IgnoreFields(
			Response{}, "IP",
		),
		cmpopts.IgnoreFields(
			Data{}, "TransID",
		),
	); diff != "" {
		t.Errorf("response mismatch (-want +got):\n%s", diff)
	}

	// check process status
	sresp, err := cli.GetProgress(context.Background(), presp.Data.TransID)
	if err != nil {
		t.Fatal(err)
	}
	want = Response{
		Code:   200,
		CSCode: 200,
		Data: Data{
			Status:   "finish",
			FileSize: 1288112,
		},
	}
	if diff := cmp.Diff(want, sresp,
		cmpopts.IgnoreFields(
			Response{}, "IP",
		),
	); diff != "" {
		t.Errorf("response mismatch (-want +got):\n%s", diff)
	}

	// download image
	dresp, err := cli.Download(context.Background(), presp.Data.TransID)
	if err != nil {
		t.Fatal(err)
	}
	if dresp == nil {
		t.Fatal("response is nil")
	}
	d := t.TempDir()
	f, err := os.Create(d + "/cat.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	written, err := io.Copy(f, dresp)
	if err != nil {
		t.Fatal(err)
	}
	if written <= uresp.Data.FileSize {
		// enlarged image size is larger than original image size
		t.Errorf("written size mismatch: want %d, got %d", sresp.Data.FileSize, written)
	}
}
