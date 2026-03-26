package libs

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContextFactoryIsolation(t *testing.T) {
	factory := &contextProviderFactory{
		client:        nil,
		imagedata:     nil,
		gctxStore:     nil,
		restMapper:    nil,
		cliEvaluation: true,
	}

	ctx1 := factory.NewContext()
	ctx2 := factory.NewContext()

	ctx1.SetGenerateContext("policy1", "trigger1", "ns1", "v1", "", "Pod", "uid1", false)
	ctx2.SetGenerateContext("policy2", "trigger2", "ns2", "v1", "", "Deployment", "uid2", true)

	cp1 := ctx1.(*contextProvider)
	cp2 := ctx2.(*contextProvider)
	assert.Equal(t, "policy1", cp1.genCtx.policyName)
	assert.Equal(t, "policy2", cp2.genCtx.policyName)
	assert.Equal(t, "trigger1", cp1.genCtx.triggerName)
	assert.Equal(t, "trigger2", cp2.genCtx.triggerName)
	assert.Equal(t, false, cp1.genCtx.restoreCache)
	assert.Equal(t, true, cp2.genCtx.restoreCache)

	assert.NotSame(t, cp1.generatedResources, cp2.generatedResources)
}

func TestContextFactoryConcurrentIsolation(t *testing.T) {
	factory := &contextProviderFactory{
		client:        nil,
		imagedata:     nil,
		gctxStore:     nil,
		restMapper:    nil,
		cliEvaluation: true,
	}

	const numGoroutines = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	results := make([]Context, numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			ctx := factory.NewContext()
			restoreCache := idx%2 == 0
			ctx.SetGenerateContext(
				"policy"+string(rune('A'+idx%26)),
				"trigger"+string(rune('0'+idx%10)),
				"ns"+string(rune('0'+idx%10)),
				"v1",
				"",
				"Pod",
				"uid"+string(rune('0'+idx%10)),
				restoreCache,
			)
			results[idx] = ctx
		}(i)
	}
	wg.Wait()

	for i := 0; i < numGoroutines; i++ {
		cp := results[i].(*contextProvider)
		expectedPolicyName := "policy" + string(rune('A'+i%26))
		assert.Equal(t, expectedPolicyName, cp.genCtx.policyName, "context %d has wrong policy name", i)

		expectedRestoreCache := i%2 == 0
		assert.Equal(t, expectedRestoreCache, cp.genCtx.restoreCache, "context %d has wrong restoreCache value", i)
	}
}

func TestContextFactorySharedResources(t *testing.T) {
	factory := &contextProviderFactory{
		client:        nil,
		imagedata:     nil,
		gctxStore:     nil,
		restMapper:    nil,
		cliEvaluation: true,
	}

	ctx1 := factory.NewContext()
	ctx2 := factory.NewContext()

	cp1 := ctx1.(*contextProvider)
	cp2 := ctx2.(*contextProvider)

	assert.Same(t, factory.client, cp1.client)
	assert.Same(t, factory.client, cp2.client)
	assert.Same(t, factory.imagedata, cp1.imagedata)
	assert.Same(t, factory.imagedata, cp2.imagedata)
	assert.Same(t, factory.gctxStore, cp1.gctxStore)
	assert.Same(t, factory.gctxStore, cp2.gctxStore)
	assert.Same(t, factory.restMapper, cp1.restMapper)
	assert.Same(t, factory.restMapper, cp2.restMapper)

	assert.Equal(t, factory.cliEvaluation, cp1.cliEvaluation)
	assert.Equal(t, factory.cliEvaluation, cp2.cliEvaluation)
}

func TestContextFactoryGeneratedResourcesIsolation(t *testing.T) {
	factory := &contextProviderFactory{
		client:        nil,
		imagedata:     nil,
		gctxStore:     nil,
		restMapper:    nil,
		cliEvaluation: true,
	}

	ctx1 := factory.NewContext()
	ctx2 := factory.NewContext()

	assert.Empty(t, ctx1.GetGeneratedResources())
	assert.Empty(t, ctx2.GetGeneratedResources())

	cp1 := ctx1.(*contextProvider)
	cp1.generatedResources = append(cp1.generatedResources, nil)

	assert.Len(t, ctx1.GetGeneratedResources(), 1)
	assert.Empty(t, ctx2.GetGeneratedResources())
}

func TestGetLibsCtxFactory(t *testing.T) {
	originalFactory := LibraryContextFactory

	testFactory := &contextProviderFactory{
		cliEvaluation: true,
	}
	LibraryContextFactory = testFactory

	result := GetLibsCtxFactory()
	assert.Same(t, testFactory, result)

	LibraryContextFactory = originalFactory
}
