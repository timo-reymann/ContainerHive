package rendering

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/timo-reymann/ContainerHive/internal/file_resolver"
	"github.com/timo-reymann/ContainerHive/pkg/model"
	"golang.org/x/sync/errgroup"
)

func processImagesForName(ctx context.Context, rootPath string, images []*model.Image) error {
	eg, _ := errgroup.WithContext(ctx)
	for _, image := range images {
		image := image

		for tag, _ := range image.Tags {
			tag := tag

			// Tag and variant are safe to run in parallel
			eg.Go(func() error {
				tagPath := filepath.Join(rootPath, tag)
				if err := setupImageDir(tagPath, image); err != nil {
					return err
				}

				for _, variantDef := range image.Variants {
					rootPath := rootPath
					variantDef := variantDef

					eg.Go(func() error {
						variantPath := filepath.Join(rootPath, tag+variantDef.TagSuffix)
						if err := setupVariantDir(variantPath, variantDef, image); err != nil {
							return err
						}
						return nil
					})

				}
				return nil
			})
		}
	}

	return eg.Wait()
}

func createTestsFolder(rootPath string) (string, error) {
	testsRoot := filepath.Join(rootPath, "tests")
	if err := mkdir(testsRoot); err != nil {
		return "", errors.Join(errors.New("failed to create tests directory"), err)
	}
	return testsRoot, nil
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

	if image.TestConfigFilePath != "" {
		testsRoot, err := createTestsFolder(tagPath)
		if err != nil {
			return err
		}

		if err := file_resolver.CopyAndRenderFile(image.TestConfigFilePath, filepath.Join(testsRoot, "image.yml")); err != nil {
			return err
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

	if image.TestConfigFilePath != "" || variantDef.TestConfigFilePath != "" {
		testsRoot, err := createTestsFolder(variantPath)
		if err != nil {
			return err
		}

		if image.TestConfigFilePath != "" {
			if err := file_resolver.CopyAndRenderFile(image.TestConfigFilePath, filepath.Join(testsRoot, "image.yml")); err != nil {
				return errors.Join(errors.New("failed to copy test config file"), err)
			}
		}

		if variantDef.TestConfigFilePath != "" {
			if err := file_resolver.CopyAndRenderFile(variantDef.TestConfigFilePath, filepath.Join(testsRoot, "variant.yml")); err != nil {
				return errors.Join(errors.New("failed to copy test config file"), err)
			}
		}
	}

	return nil
}

func RenderProject(ctx context.Context, project *model.ContainerHiveProject, targetPath string) error {
	_ = os.RemoveAll(targetPath)

	err := mkdir(targetPath)
	if err != nil {
		return errors.Join(errors.New("failed to create target directory"), err)
	}
	eg, _ := errgroup.WithContext(ctx)

	for name, images := range project.ImagesByName {
		images := images
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
