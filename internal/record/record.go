package record

type Record struct {
	Samples       []float32
	TriggerIndex  int
	TriggerOffset float64
}
