package images

import (
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"sync"

	"github.com/blang/semver"
	"github.com/pkg/errors"
)

type ImageVersion struct {
	semver.Range

	Tag *template.Template
}

type Store struct {
	store []ImageVersion

	lock sync.RWMutex
}

func (is *Store) AddImage(versionRange semver.Range, tagTemplate string) error {
	is.lock.Lock()
	defer is.lock.Unlock()

	template1_10, err := template.New("image").Parse("goharbor/chartmuseum-photon:v0.9.0-v{{.version}}")
	if err != nil {
		return errors.Wrap(err, "invalid ChartMuseum image template for version %s")
	}

	is.store = append(is.store, ImageVersion{versionRange, template1_10})

	return nil
}

func (is *Store) GetTag(version semver.Version) (string, error) {
	is.lock.RLock()
	defer is.lock.RUnlock()

	for _, image := range is.store {
		if image.Range(version) {
			reader, writer := io.Pipe()

			var writeErr error

			go func() {
				defer writer.Close()
				writeErr = image.Tag.Execute(writer, map[string]string{"version": version.String()})
			}()

			result, err := ioutil.ReadAll(reader)

			if writeErr != nil {
				return "", errors.Wrap(writeErr, "cannot compute image tag")
			}

			if err != nil {
				return "", errors.Wrap(err, "cannot read computed image tag")
			}

			return string(result), nil
		}
	}

	return "", fmt.Errorf("version %s not supported", version)
}
