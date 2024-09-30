package extensiontests

import (
	"fmt"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"github.com/google/cel-go/common/types"
)

func (specs ExtensionTestSpecs) Walk(walkFn func(*ExtensionTestSpec)) ExtensionTestSpecs {
	for i := range specs {
		walkFn(specs[i])
	}

	return specs
}

func (specs ExtensionTestSpecs) MustFilter(celExprs []string) ExtensionTestSpecs {
	specs, err := specs.Filter(celExprs)
	if err != nil {
		panic(fmt.Sprintf("filter did not succeed: %s", err.Error()))
	}

	return specs
}

func (specs ExtensionTestSpecs) Filter(celExprs []string) (ExtensionTestSpecs, error) {
	var filteredSpecs ExtensionTestSpecs

	env, err := cel.NewEnv(
		cel.Declarations(
			decls.NewVar("source", decls.String),
			decls.NewVar("name", decls.String),
			decls.NewVar("other_names", decls.NewListType(decls.String)),
			decls.NewVar("labels", decls.NewListType(decls.String)),
			decls.NewVar("tags", decls.NewMapType(decls.String, decls.String)),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create CEL environment: %w", err)
	}

	// OR all expressions together
	for _, spec := range specs {
		include := false
		for _, celExpr := range celExprs {
			// Parse CEL expression
			ast, iss := env.Parse(celExpr)
			if iss.Err() != nil {
				return nil, fmt.Errorf("error parsing CEL expression '%s': %v", celExpr, iss.Err())
			}

			// Check the AST
			checked, iss := env.Check(ast)
			if iss.Err() != nil {
				return nil, fmt.Errorf("error checking CEL expression '%s': %v", celExpr, iss.Err())
			}

			// Create a CEL program from the checked AST
			prg, err := env.Program(checked)
			if err != nil {
				return nil, fmt.Errorf("error creating CEL program: %v", err)
			}

			out, _, err := prg.Eval(map[string]interface{}{
				"name":        spec.Name,
				"other_names": spec.OtherNames,
				"labels":      spec.Labels.UnsortedList(),
				"tags":        spec.Tags,
			})
			if err != nil {
				return nil, fmt.Errorf("error evaluating CEL expression: %v", err)
			}

			// If any CEL expression evaluates to true, include the TestSpec
			if out == types.True {
				include = true
				break
			}
		}
		if include {
			filteredSpecs = append(filteredSpecs, spec)
		}
	}

	return filteredSpecs, nil
}

func (specs ExtensionTestSpecs) AddLabel(labels ...string) ExtensionTestSpecs {
	for i := range specs {
		specs[i].Labels.Insert(labels...)
	}

	return specs
}

func (specs ExtensionTestSpecs) RemoveLabel(labels ...string) ExtensionTestSpecs {
	for i := range specs {
		specs[i].Labels.Delete(labels...)
	}

	return specs
}

func (specs ExtensionTestSpecs) SetTag(key, value string) ExtensionTestSpecs {
	for i := range specs {
		specs[i].Tags[key] = value
	}

	return specs
}

func (specs ExtensionTestSpecs) UnsetTag(key string) ExtensionTestSpecs {
	for i := range specs {
		delete(specs[i].Tags, key)
	}

	return specs
}
