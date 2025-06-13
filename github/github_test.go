package github

// This test doesn't work in CI, because we're running in GHA. Not sure if it's worth fixing.
// func TestActionsEnvVars(t *testing.T) {

// 	tests := []struct {
// 		name string
// 		env  map[string]string
// 		want actionsEnvVars
// 		err  error
// 	}{
// 		{
// 			name: "in gha",
// 			env: map[string]string{
// 				"GITHUB_ACTIONS": "true",
// 				"GITHUB_ACTION":  "test",
// 			},
// 			want: actionsEnvVars{
// 				Action: "test",
// 			},
// 		},
// 		{
// 			name: "not in gha",
// 			env: map[string]string{
// 				"GITHUB_ACTION": "test",
// 			},
// 			want: actionsEnvVars{},
// 			err:  ErrNotInActions,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			for k, v := range tt.env {
// 				t.Setenv(k, v)
// 			}

// 			vars, err := ActionsEnvVars()
// 			if tt.err != nil {
// 				require.Error(t, err)
// 				assert.ErrorIs(t, err, tt.err)
// 			} else {
// 				require.NoError(t, err)
// 				require.NotNil(t, vars)
// 			}
// 		})
// 	}
// }
