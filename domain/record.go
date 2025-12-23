package domain

type Record struct {
	values  []interface{}
	keys    []string
	indices map[string]int
}

func NewRecord(values []interface{}, keys []string) *Record {
	r := &Record{
		values:  values,
		keys:    keys,
		indices: make(map[string]int, len(keys)),
	}

	for i, k := range keys {
		r.indices[k] = i
	}

	return r
}

func (r *Record) Keys() []string {
	return r.keys
}

func (r *Record) Values() []interface{} {
	return r.values
}

func (r *Record) Get(key string) (interface{}, bool) {
	if r.indices == nil {
		r.indices = make(map[string]int, len(r.keys))
		for i, k := range r.keys {
			r.indices[k] = i
		}
	}

	if idx, ok := r.indices[key]; ok {
		return r.values[idx], true
	}

	return nil, false
}

func (r *Record) GetByIndex(index int) interface{} {
	if index < len(r.values) {
		return r.values[index]
	}
	return nil
}
