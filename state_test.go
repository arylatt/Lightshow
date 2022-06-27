package lightshow

// func TestStateStepsTo(t *testing.T) {
// 	tests := []struct {
// 		inputs struct {
// 			start  State
// 			steps  int
// 			target State
// 		}
// 		output []State
// 	}{{
// 		struct {
// 			start  State
// 			steps  int
// 			target State
// 		}{
// 			State{"#FFFFFF", 1}, 1, StateOff,
// 		},
// 		[]State{{"#FFFFFF", 1}, {"#000000", 0}},
// 	}, {
// 		struct {
// 			start  State
// 			steps  int
// 			target State
// 		}{
// 			State{"#FFFFFF", 1}, 2, StateOff,
// 		},
// 		[]State{{"#FFFFFF", 1}, {"8080", 0.5}, {"#000000", 0}},
// 	}, {
// 		struct {
// 			start  State
// 			steps  int
// 			target State
// 		}{
// 			State{"#FFFFFF", 1}, 10, State{"#00FF00", 0.8},
// 		},
// 		[]State{{"#FFFFFF", 1}, {"e6ff", 0.98}, {"cdff", 0.96}, {"b4ff", 0.9400000000000001}, {"9bff", 0.92}, {"82ff", 0.90}, {"69ff", 0.88}, {"50ff", 0.86}, {"37ff", 0.8400000000000001}, {"1eff", 0.8200000000000001}, {"#00FF00", 0.8}},
// 	},
// 	}

// 	for _, test := range tests {
// 		assert.Equal(t, test.output, test.inputs.start.StepsTo(test.inputs.steps, test.inputs.target))
// 	}
// }
