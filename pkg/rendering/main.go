package rendering

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/timo-reymann/ContainerHive/pkg/model"
	"golang.org/x/sync/errgroup"
)

func mkdir(targetPath string) error {
	return os.MkdirAll(targetPath, 0755)
}

func processImagesForName(ctx context.Context, rootPath string, images []*model.Image) error {
	eg, _ := errgroup.WithContext(ctx)
	for _, image := range images {
		for tag, _ := range image.Tags {
			// Tag and variant are safe to run in parallel
			eg.Go(func() error {
				tagPath := filepath.Join(rootPath, tag)
				if err := setupImageDir(tagPath, image); err != nil {
					return err
				}

				for _, variantDef := range image.Variants {
					variantPath := filepath.Join(rootPath, tag+variantDef.TagSuffix)
					if err := setupVariantDir(variantPath, variantDef, image); err != nil {
						return err
					}
				}
				return nil
			})
		}
	}

	return eg.Wait()
}

func setupImageDir(tagPath string, image *model.Image) error {
	if err := mkdir(tagPath); err != nil {
		return errors.Join(errors.New("failed to create tag directory"), err)
	}

	if image.RootFSDir != "" {
		if err := copyRootFs(image.RootFSDir, tagPath); err != nil {
			return errors.Join(errors.New("failed to copy rootfs"), err)
		}
	}
	return nil
}

func setupVariantDir(variantPath string, variantDef *model.ImageVariant, image *model.Image) error {
	if err := mkdir(variantPath); err != nil {
		return errors.Join(errors.New("failed to create variant directory"), err)
	}

	if image.RootFSDir != "" {
		if err := copyRootFs(image.RootFSDir, variantPath); err != nil {
			return errors.Join(errors.New("failed to copy rootfs for variant from base version"), err)
		}
	}

	if variantDef.RootFSDir != "" {
		if err := copyRootFs(variantDef.RootFSDir, variantPath); err != nil {
			return errors.Join(errors.New("failed to copy rootfs for variant"), err)
		}
	}
	return nil
}

func RenderProject(ctx context.Context, project *model.ContainerHiveProject, targetPath string) error {
	err := mkdir(targetPath)
	if err != nil {
		return errors.Join(errors.New("failed to create target directory"), err)
	}
	eg, _ := errgroup.WithContext(ctx)

	for name, images := range project.ImagesByName {
		nameRootPath := filepath.Join(targetPath, name)
		err := mkdir(nameRootPath)
		if err != nil {
			return errors.Join(errors.New("failed to create image directory for "+name), err)
		}

		eg.Go(func() error {
			return processImagesForName(ctx, nameRootPath, images)
		})
	}

	return eg.Wait()
}
