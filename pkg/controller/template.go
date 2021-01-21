package controller

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"text/template"

	"github.com/Masterminds/sprig"
	serrors "github.com/goharbor/harbor-operator/pkg/controller/errors"
	resources "github.com/goharbor/harbor-operator/pkg/resources"
	template2 "github.com/goharbor/harbor-operator/pkg/template"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
)

func (c *Controller) Funcs(ctx context.Context, owner resources.Resource) template.FuncMap {
	client := c.Client
	namespace := owner.GetNamespace()

	return template.FuncMap{
		"secretData":           template2.GetSecretDataFunc(ctx, client, namespace, false),
		"secretDataAllowEmpty": template2.GetSecretDataFunc(ctx, client, namespace, true),
	}
}

func (c *Controller) GetTemplatedConfig(ctx context.Context, templateConfig string, owner resources.Resource) ([]byte, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "GetTemplatedConfig")
	defer span.Finish()

	t, err := template.New("template").
		Funcs(sprig.TxtFuncMap()).
		Funcs(c.Funcs(ctx, owner)).
		Parse(templateConfig)
	if err != nil {
		return nil, errors.Wrap(err, "invalid template")
	}

	reader, writer := io.Pipe()
	defer reader.Close()

	var errTemplate error

	go func() {
		defer writer.Close()

		errTemplate = t.Execute(writer, owner)
	}()

	configContent, err := ioutil.ReadAll(reader)

	if errTemplate != nil {
		if errors.As(err, &template.ExecError{}) {
			return nil, serrors.UnrecoverrableError(errTemplate, "operatorCompatibility", fmt.Sprintf("cannot process config template: %v", errTemplate))
		}

		return nil, serrors.UnrecoverrableError(errTemplate, serrors.OperatorReason, fmt.Sprintf("cannot process config template: %v", errTemplate))
	}

	if err != nil {
		return nil, errors.Wrap(err, "cannot read processed config")
	}

	return configContent, nil
}
