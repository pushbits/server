package credentials

import "testing"

type isPasswordPwnedTest struct {
	arg  string
	exp1 bool
	exp2 error
}

var isPasswordPwnedTests = []isPasswordPwnedTest{
	{"", true, nil},
	{"password", true, nil},
	{"2y6bWMETuHpNP08HCZq00QAAzE6nmwEb", false, nil},
}

func TestIsPasswordPwned(t *testing.T) {
	for _, test := range isPasswordPwnedTests {
		if out1, out2 := IsPasswordPwned(test.arg); out1 != test.exp1 || out2 != test.exp2 {
			t.Errorf("Output (%t,%q) not equal to expected (%t,%q)", out1, out2, test.exp1, test.exp2)
		}
	}
}
