package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/0daryo/vanceai-go"
)

func main() {
	cli, err := vanceai.NewClient(os.Getenv("API_KEY"), "")
	if err != nil {
		panic(err)
	}
	image, err := os.Open("../../testdata/cat.jpg")
	if err != nil {
		panic(err)
	}
	defer image.Close()

	// upload image
	uresp, err := cli.UploadImage(context.Background(), image, "check.jpg")
	if err != nil {
		panic(err)
	}

	fmt.Printf("upload finished: %+v\n", uresp)

	// process image
	presp, err := cli.ProcessImage(context.Background(), uresp.Data.UID, &vanceai.MultipleJob{
		Job: "workflow",
		Config: []vanceai.SingleConfig{
			{
				Name: "sketch",
				Config: vanceai.Config{
					Module: "sketch",
					ModuleParams: vanceai.ModuleParams{
						ModelName:  "SketchStable",
						SingleFace: true,
						Composite:  true,
						Sigma:      0,
						Alpha:      0,
					},
					OutParams: vanceai.OutParams{},
				},
			},
			{
				Name: "matting",
				Config: vanceai.Config{
					Module: "matting",
					ModuleParams: vanceai.ModuleParams{
						ModelName:   "MattingStable",
						Rescale:     532,
						WebAutoMode: false,
					},
					OutParams: vanceai.OutParams{
						Compress: vanceai.Compress{
							Quality: 100,
						},
					},
				},
			},
		},
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("process finished: %+v\n", presp)

	// check process status
	gresp, err := cli.GetProgress(context.Background(), presp.Data.TransID)
	if err != nil {
		panic(err)
	}
	fmt.Printf("get process: %+v\n", gresp)

	// download image
	dresp, err := cli.Download(context.Background(), presp.Data.TransID)
	if err != nil {
		panic(err)
	}
	if dresp == nil {
		panic("response is nil")
	}
	d := uresp.Data.UID
	if err := os.Mkdir(d, os.ModePerm); err != nil {
		panic(err)
	}
	f, err := os.Create(d + "/output.jpg")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	_, err = io.Copy(f, dresp)
	if err != nil {
		panic(err)
	}
	fmt.Printf("download finished: directory is %s\n", d)

	cresp, err := cli.GetPoint(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Printf("get point: %+v\n", cresp)
}
