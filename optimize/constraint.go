package optimize

// ConstraintFunc is a function that returns true if the parameters are valid.
type ConstraintFunc func(params map[string]interface{}) bool

// ParamLessThan returns a constraint that requires p1 < p2.
func ParamLessThan(p1, p2 string) ConstraintFunc {
	return func(params map[string]interface{}) bool {
		v1, ok1 := getNumericValue(params[p1])
		v2, ok2 := getNumericValue(params[p2])
		if !ok1 || !ok2 {
			return true // If types don't match, allow it
		}
		return v1 < v2
	}
}

// ParamLessThanOrEqual returns a constraint that requires p1 <= p2.
func ParamLessThanOrEqual(p1, p2 string) ConstraintFunc {
	return func(params map[string]interface{}) bool {
		v1, ok1 := getNumericValue(params[p1])
		v2, ok2 := getNumericValue(params[p2])
		if !ok1 || !ok2 {
			return true
		}
		return v1 <= v2
	}
}

// ParamGreaterThan returns a constraint that requires p1 > p2.
func ParamGreaterThan(p1, p2 string) ConstraintFunc {
	return func(params map[string]interface{}) bool {
		v1, ok1 := getNumericValue(params[p1])
		v2, ok2 := getNumericValue(params[p2])
		if !ok1 || !ok2 {
			return true
		}
		return v1 > v2
	}
}

// ParamGreaterThanOrEqual returns a constraint that requires p1 >= p2.
func ParamGreaterThanOrEqual(p1, p2 string) ConstraintFunc {
	return func(params map[string]interface{}) bool {
		v1, ok1 := getNumericValue(params[p1])
		v2, ok2 := getNumericValue(params[p2])
		if !ok1 || !ok2 {
			return true
		}
		return v1 >= v2
	}
}

// ParamNotEqual returns a constraint that requires p1 != p2.
func ParamNotEqual(p1, p2 string) ConstraintFunc {
	return func(params map[string]interface{}) bool {
		v1, ok1 := getNumericValue(params[p1])
		v2, ok2 := getNumericValue(params[p2])
		if !ok1 || !ok2 {
			return true
		}
		return v1 != v2
	}
}

// ParamMinValue returns a constraint that requires param >= minVal.
func ParamMinValue(param string, minVal float64) ConstraintFunc {
	return func(params map[string]interface{}) bool {
		v, ok := getNumericValue(params[param])
		if !ok {
			return true
		}
		return v >= minVal
	}
}

// ParamMaxValue returns a constraint that requires param <= maxVal.
func ParamMaxValue(param string, maxVal float64) ConstraintFunc {
	return func(params map[string]interface{}) bool {
		v, ok := getNumericValue(params[param])
		if !ok {
			return true
		}
		return v <= maxVal
	}
}

// ParamRange returns a constraint that requires minVal <= param <= maxVal.
func ParamRange(param string, minVal, maxVal float64) ConstraintFunc {
	return func(params map[string]interface{}) bool {
		v, ok := getNumericValue(params[param])
		if !ok {
			return true
		}
		return v >= minVal && v <= maxVal
	}
}

// And combines multiple constraints with AND logic.
func And(constraints ...ConstraintFunc) ConstraintFunc {
	return func(params map[string]interface{}) bool {
		for _, c := range constraints {
			if !c(params) {
				return false
			}
		}
		return true
	}
}

// Or combines multiple constraints with OR logic.
func Or(constraints ...ConstraintFunc) ConstraintFunc {
	return func(params map[string]interface{}) bool {
		if len(constraints) == 0 {
			return true
		}
		for _, c := range constraints {
			if c(params) {
				return true
			}
		}
		return false
	}
}

// Not negates a constraint.
func Not(constraint ConstraintFunc) ConstraintFunc {
	return func(params map[string]interface{}) bool {
		return !constraint(params)
	}
}

// getNumericValue extracts a numeric value from an interface{}.
func getNumericValue(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case int:
		return float64(val), true
	case int8:
		return float64(val), true
	case int16:
		return float64(val), true
	case int32:
		return float64(val), true
	case int64:
		return float64(val), true
	case uint:
		return float64(val), true
	case uint8:
		return float64(val), true
	case uint16:
		return float64(val), true
	case uint32:
		return float64(val), true
	case uint64:
		return float64(val), true
	case float32:
		return float64(val), true
	case float64:
		return val, true
	default:
		return 0, false
	}
}
