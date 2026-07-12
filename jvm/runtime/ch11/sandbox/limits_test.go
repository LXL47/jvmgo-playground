package sandbox

import "testing"

func TestInstructionBudget(t *testing.T) {
	Configure(Limits{Enabled: true, MaxInstructions: 2})
	CountInstruction()
	CountInstruction()
	assertPanic(t, CountInstruction)
}

func TestHeapAndArrayBudgets(t *testing.T) {
	Configure(Limits{Enabled: true, MaxHeapBytes: 8, MaxArrayLength: 4})
	ReserveHeap(8)
	assertPanic(t, func() { ReserveHeap(1) })
	assertPanic(t, func() { CheckArrayLength(5) })
}

func TestOutputBudget(t *testing.T) {
	Configure(Limits{Enabled: true, MaxOutputBytes: 3})
	CountOutput(3)
	assertPanic(t, func() { CountOutput(1) })
}

func assertPanic(t *testing.T, action func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatal("预期触发预算异常")
		}
	}()
	action()
}
