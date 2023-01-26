package teal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func testVm(res *ProcessResult) *Vm {
	vm, err := NewVm(res)
	if err != nil {
		panic(err)
	}

	return vm
}

func TestCostly(t *testing.T) {
	res := Process(`#pragma version 7
	txn Sender
	base64_decode StdEncoding`)

	vm := testVm(res)
	vm.Run()
}

func BenchmarkVm(b *testing.B) {
	res := Process(`
	#pragma version 8
	txn ApplicationID
	bnz initialized

	byte "manager"
	txn Sender
	app_global_put
	b end

	initialized:
	byte "welcome back"
	log

	end:
	int 1
	return
	`)

	for i := 0; i < b.N; i++ {
		vm := testVm(res)
		vm.Run()
	}
}

func BenchmarkInfiniteLoop(b *testing.B) {
	res := Process(`main:
	b main`)

	for i := 0; i < b.N; i++ {
		vm := testVm(res)
		vm.Run()
	}
}

func TestVmError(t *testing.T) {
	res := Process(`int 1
	-
	int 1`)

	vm := testVm(res)
	vm.Run()

	assert.NotNil(t, vm.Error)
	assert.Equal(t, 1, vm.Branch.Line)
}

func testVmRunSource(source string) *Vm {
	res := Process(source)

	vm := testVm(res)
	vm.Run()

	return vm
}

func testVmBranch(t *testing.T, source string, count int) *Vm {
	vm := testVmRunSource(source)
	assert.Len(t, vm.Branches, count)
	return vm
}

func TestEmptyCodeBranches(t *testing.T) {
	testVmBranch(t, "", 1)
}

func TestBzConst0Branches(t *testing.T) {
	vm := testVmBranch(t, `#pragma version 2
	int 0
	bz test
	int 2
	test:`, 1)

	assert.False(t, vm.Visited[3])
}

func TestBzConst1Branches(t *testing.T) {
	vm := testVmBranch(t, `#pragma version 2
	int 1
	bz test
	int 2
	test:`, 1)

	assert.True(t, vm.Visited[3])
}

func TestBzRangeBranches(t *testing.T) {
	vm := testVmBranch(t, `#pragma version 2
	txn ApplicationID
	bz test
	int 2
	test:`, 2)

	assert.True(t, vm.Visited[3])
}

func TestBnzConst0Branches(t *testing.T) {
	vm := testVmBranch(t, `#pragma version 2
	int 0
	bnz test
	int 2
	test:`, 1)

	assert.True(t, vm.Visited[3])
}

func TestBnzConst1Branches(t *testing.T) {
	vm := testVmBranch(t, `#pragma version 2
	int 1
	bnz test
	int 2
	test:`, 1)

	assert.False(t, vm.Visited[3])
}

func TestBnzRangeBranches(t *testing.T) {
	vm := testVmBranch(t, `#pragma version 2
	txn ApplicationID
	bnz test
	int 2
	test:`, 2)

	assert.True(t, vm.Visited[3])
}

func TestSwitchConst0Of3Branches(t *testing.T) {
	vm := testVmBranch(t, `#pragma version 8
	int 0
	switch a b c
	a:
	b end
	b:
	b end
	c:
	b end
	end:
	int 1`, 1)

	assert.True(t, vm.Visited[4])
	assert.False(t, vm.Visited[6])
	assert.False(t, vm.Visited[8])
	assert.True(t, vm.Visited[10])
}

func TestSwitchConst3Of3Branches(t *testing.T) {
	vm := testVmBranch(t, `#pragma version 8
	int 3
	switch a b c
	a:
	b end
	b:
	b end
	c:
	b end
	end:
	int 1`, 1)

	assert.True(t, vm.Visited[4])
	assert.False(t, vm.Visited[6])
	assert.False(t, vm.Visited[8])
	assert.True(t, vm.Visited[10])
}

func TestSwitchRangeBranches(t *testing.T) {
	vm := testVmRunSource(`#pragma version 8
	txn OnCompletion
	switch a b c
	a:
	b end
	b:
	b end
	c:
	b end
	end:
	int 1`)

	assert.True(t, vm.Visited[4])
	assert.True(t, vm.Visited[6])
	assert.True(t, vm.Visited[8])
	assert.True(t, vm.Visited[10])
}

func TestAssertConst0(t *testing.T) {
	vm := testVmRunSource(`int 0
	assert`)

	assert.NotNil(t, vm.Error)
}

func TestAssertConst1(t *testing.T) {
	vm := testVmRunSource(`int 1
	assert`)

	assert.Nil(t, vm.Error)
}

func TestAssertRange(t *testing.T) {
	vm := testVmRunSource(`#pragma version 2
	txn ApplicationID
	assert`)

	assert.Nil(t, vm.Error)
}

func TestMatchConstInt(t *testing.T) {
	vm := testVmRunSource(`#pragma version 8
	int 1
	int 2
	int 3
	txn ApplicationID
	int 3
	match a b c d
	a:
	-
	b:
	-
	c:
	return
	d:
	return
	`)

	assert.False(t, vm.Visited[8])
	assert.False(t, vm.Visited[10])
	assert.True(t, vm.Visited[12])
	assert.True(t, vm.Visited[14])

	assert.Nil(t, vm.Error)
}

func testBreakOnError(t *testing.T, enabled bool, source string) *Vm {
	vm := testVm(Process(source))
	vm.BreakOnError = enabled
	vm.Run()
	return vm
}

const testBreakOnErrorSource = `#pragma version 8
txn ApplicationID
bnz test
-
test:
int 1`

func TestBreakOnErrorEnabled(t *testing.T) {
	vm := testBreakOnError(t, true, testBreakOnErrorSource)
	assert.False(t, vm.Visited[5])
}

func TestBreakOnErrorDisabled(t *testing.T) {
	vm := testBreakOnError(t, false, testBreakOnErrorSource)
	assert.True(t, vm.Visited[5])
}

func TestSwitchBranchOnRun(t *testing.T) {
	vm := testVm(Process(`#pragma version 8
	txn ApplicationID
	bnz test
	-
	test:
	int 1
	pop
	-`))

	vm.Run()
	assert.False(t, vm.Visited[5])
	vm.Run()
	assert.True(t, vm.Visited[5])
}
