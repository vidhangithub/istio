// Copyright Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package schema

import (
	"fmt"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/onsi/gomega"

	"istio.io/api/networking/v1alpha3"
	"istio.io/istio/pkg/config"
	"istio.io/istio/pkg/config/analysis/msg"
	"istio.io/istio/pkg/config/analysis/testing/fixtures"
	"istio.io/istio/pkg/config/resource"
	"istio.io/istio/pkg/config/schema/collections"
	"istio.io/istio/pkg/config/schema/gvk"
	resource2 "istio.io/istio/pkg/config/schema/resource"
	"istio.io/istio/pkg/config/validation"
)

func TestCorrectArgs(t *testing.T) {
	g := gomega.NewWithT(t)

	m1 := &v1alpha3.VirtualService{}

	testSchema := schemaWithValidateFn(func(cfg config.Config) (warnings validation.Warning, errs error) {
		g.Expect(cfg.Name).To(gomega.Equal("name"))
		g.Expect(cfg.Namespace).To(gomega.Equal("ns"))
		g.Expect(cfg.Spec).To(gomega.Equal(m1))
		return nil, nil
	})
	ctx := &fixtures.Context{
		Resources: []*resource.Instance{
			{
				Message: &v1alpha3.VirtualService{},
				Metadata: resource.Metadata{
					FullName: resource.NewFullName("ns", "name"),
				},
				Origin: fakeOrigin{},
			},
		},
	}
	a := ValidationAnalyzer{s: testSchema}
	a.Analyze(ctx)
}

func TestSchemaValidationWrapper(t *testing.T) {
	testCol := gvk.VirtualService

	m1 := &v1alpha3.VirtualService{}
	m2 := &v1alpha3.VirtualService{}
	m3 := &v1alpha3.VirtualService{}

	testSchema := schemaWithValidateFn(func(cfg config.Config) (warnings validation.Warning, errs error) {
		if cfg.Spec == m1 {
			return nil, nil
		}
		if cfg.Spec == m2 {
			return nil, fmt.Errorf("")
		}
		if cfg.Spec == m3 {
			return nil, multierror.Append(fmt.Errorf(""), fmt.Errorf(""))
		}
		return nil, nil
	})

	a := ValidationAnalyzer{s: testSchema}

	t.Run("CheckMetadataInputs", func(t *testing.T) {
		g := gomega.NewWithT(t)
		g.Expect(a.Metadata().Inputs).To(gomega.ConsistOf(testCol))
	})

	t.Run("NoErrors", func(t *testing.T) {
		g := gomega.NewWithT(t)
		ctx := &fixtures.Context{
			Resources: []*resource.Instance{
				{
					Message: m1,
				},
			},
		}
		a.Analyze(ctx)
		g.Expect(ctx.Reports).To(gomega.BeEmpty())
	})

	t.Run("SingleError", func(t *testing.T) {
		g := gomega.NewWithT(t)

		ctx := &fixtures.Context{
			Resources: []*resource.Instance{
				{
					Message: m2,
					Origin:  fakeOrigin{},
				},
			},
		}
		a.Analyze(ctx)
		g.Expect(ctx.Reports).To(gomega.HaveLen(1))
		g.Expect(ctx.Reports[0].Type).To(gomega.Equal(msg.SchemaValidationError))
	})

	t.Run("MultiError", func(t *testing.T) {
		g := gomega.NewWithT(t)
		ctx := &fixtures.Context{
			Resources: []*resource.Instance{
				{
					Message: m3,
					Origin:  fakeOrigin{},
				},
			},
		}
		a.Analyze(ctx)
		g.Expect(ctx.Reports).To(gomega.HaveLen(2))
		g.Expect(ctx.Reports[0].Type).To(gomega.Equal(msg.SchemaValidationError))
		g.Expect(ctx.Reports[1].Type).To(gomega.Equal(msg.SchemaValidationError))
	})
}

func schemaWithValidateFn(validateFn func(cfg config.Config) (validation.Warning, error)) resource2.Schema {
	original := collections.VirtualService
	return resource2.Builder{
		ClusterScoped: original.IsClusterScoped(),
		Kind:          original.Kind(),
		Plural:        original.Plural(),
		Group:         original.Group(),
		Version:       original.Version(),
		Proto:         original.Proto(),
		ProtoPackage:  original.ProtoPackage(),
		ValidateProto: validateFn,
	}.MustBuild()
}

type fakeOrigin struct{}

func (fakeOrigin) FriendlyName() string          { return "myFriendlyName" }
func (fakeOrigin) Comparator() string            { return "myFriendlyName" }
func (fakeOrigin) Namespace() resource.Namespace { return "myNamespace" }
func (fakeOrigin) Reference() resource.Reference { return fakeReference{} }
func (fakeOrigin) FieldMap() map[string]int      { return make(map[string]int) }

type fakeReference struct{}

func (fakeReference) String() string { return "" }
