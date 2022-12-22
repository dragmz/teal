package teal

import "testing"

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
		vm := NewVm(res)
		vm.Run()
	}
}

func BenchmarkInfiniteLoop(b *testing.B) {
	res := Process(`main:
	b main`)

	for i := 0; i < b.N; i++ {
		vm := NewVm(res)
		vm.Run()
	}
}
