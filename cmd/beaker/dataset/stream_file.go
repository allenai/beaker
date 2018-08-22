package dataset

import (
	"context"
	"io"
	"os"

	beaker "github.com/allenai/beaker-api/client"
	"github.com/pkg/errors"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/allenai/beaker/config"
)

type streamFileOptions struct {
	dataset string
	file    string
}

func newStreamCmd(
	parent *kingpin.CmdClause,
	parentOpts *datasetOptions,
	config *config.Config,
) {
	o := &streamFileOptions{}
	cmd := parent.Command("stream-file", "Stream a single file from an existing dataset to stdout")
	cmd.Action(func(c *kingpin.ParseContext) error {
		beaker, err := beaker.NewClient(parentOpts.addr, config.UserToken)
		if err != nil {
			return err
		}
		return o.run(beaker)
	})

	cmd.Arg("dataset", "Dataset name or ID").Required().StringVar(&o.dataset)
	cmd.Arg("file", "File in dataset to fetch. Optional for single-file datasets.").StringVar(&o.file)
}

func (o *streamFileOptions) run(beaker *beaker.Client) error {
	ctx := context.TODO()
	dataset, err := beaker.Dataset(ctx, o.dataset)
	if err != nil {
		return err
	}

	manifest, err := dataset.Manifest(ctx)
	if err != nil {
		return err
	}

	var filename = o.file
	if filename == "" {
		if !manifest.SingleFile {
			return errors.Errorf("filename required for multi-file dataset %s", manifest.ID)
		}
		if len(manifest.Files) == 0 {
			return errors.Errorf("dataset %s has no files", manifest.ID)
		}
		filename = manifest.Files[0].File
	}

	r, err := dataset.FileRef(filename).Download(ctx)
	if err != nil {
		return err
	}
	defer r.Close()

	_, err = io.Copy(os.Stdout, r)
	return errors.WithStack(err)
}
