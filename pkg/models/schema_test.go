package models

import "testing"

func TestConstraint_IsNotNullCheck(t *testing.T) {
	cases := []struct {
		name string
		c    Constraint
		want bool
	}{
		{
			name: "postgres synthetic not-null check",
			c:    Constraint{Name: "16385_16398_3_not_null", Type: Check, CheckExpression: "total IS NOT NULL"},
			want: true,
		},
		{
			name: "oracle synthetic not-null check (quoted, uppercase)",
			c:    Constraint{Name: "SYS_C008649", Type: Check, CheckExpression: `"TOTAL" IS NOT NULL`},
			want: true,
		},
		{
			name: "explicit not-null constraint type",
			c:    Constraint{Name: "nn_total", Type: NotNull},
			want: true,
		},
		{
			name: "real check constraint is kept",
			c:    Constraint{Name: "orders_total_check", Type: Check, CheckExpression: "total >= 0"},
			want: false,
		},
		{
			name: "check referencing not null but with more logic is kept",
			c:    Constraint{Name: "chk", Type: Check, CheckExpression: "total IS NOT NULL AND total > 0"},
			want: false,
		},
		{
			name: "primary key is not a not-null check",
			c:    Constraint{Name: "pk", Type: PrimaryKey, Columns: []string{"id"}},
			want: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.c.IsNotNullCheck(); got != tc.want {
				t.Errorf("IsNotNullCheck() = %v, want %v", got, tc.want)
			}
		})
	}
}
