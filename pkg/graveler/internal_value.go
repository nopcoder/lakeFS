package graveler

// rangeToValue returns the Value representing rng in a MetaRange
func rangeToValue(rng Range) (Value, error) {
	data, err := MarshalRange(rng)
	if err != nil {
		return nil, err
	}
	rangeValue := &Value{
		Identity: []byte(rng.ID),
		Data:     data,
	}
	return MarshalValue(rangeValue)
}
