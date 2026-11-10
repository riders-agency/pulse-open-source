package kernel

import "testing"

func TestRuntimeWorkerCountFromEnv(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want int
	}{
		{name: "empty", raw: "", want: defaultRuntimeWorkerCount},
		{name: "invalid", raw: "many", want: defaultRuntimeWorkerCount},
		{name: "zero", raw: "0", want: defaultRuntimeWorkerCount},
		{name: "above maximum", raw: "65", want: defaultRuntimeWorkerCount},
		{name: "minimum", raw: "1", want: 1},
		{name: "configured", raw: "16", want: 16},
		{name: "maximum", raw: "64", want: 64},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := RuntimeWorkerCountFromEnv(test.raw); got != test.want {
				t.Fatalf("RuntimeWorkerCountFromEnv(%q) = %d, want %d", test.raw, got, test.want)
			}
		})
	}
}

func TestSignalRuntimeWorkersSignalsEveryAvailableWorkerOnce(t *testing.T) {
	signals := []chan struct{}{make(chan struct{}, 1), make(chan struct{}, 1), make(chan struct{}, 1)}
	signalRuntimeWorkers(signals)
	signalRuntimeWorkers(signals)
	for index, signal := range signals {
		select {
		case <-signal:
		default:
			t.Fatalf("worker %d was not signalled", index)
		}
		select {
		case <-signal:
			t.Fatalf("worker %d received duplicate buffered signal", index)
		default:
		}
	}
}

func TestWithWorkerCountRejectsOutOfRangeValues(t *testing.T) {
	kernel := &Kernel{workerCount: defaultRuntimeWorkerCount}
	WithWorkerCount(32)(kernel)
	if kernel.workerCount != 32 {
		t.Fatalf("worker count = %d, want 32", kernel.workerCount)
	}
	WithWorkerCount(0)(kernel)
	WithWorkerCount(65)(kernel)
	if kernel.workerCount != 32 {
		t.Fatalf("out-of-range option changed worker count to %d", kernel.workerCount)
	}
}
