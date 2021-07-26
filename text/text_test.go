package text

import "testing"

func TestNormalize(t *testing.T) {
	cases := []struct {
		value, expected string
	}{
		{"àéîõüç", "aeiouc"},
		{"aeiou", "aeiou"},
		{"", ""},
		{"/mnt/data/CPI do Genocídio", "/mnt/data/CPI do Genocidio"},
		{"c:\\Teste", "c:\\Teste"},
	}
	for _, c := range cases {
		res, err := Normalize(c.value)
		if err != nil {
			t.Errorf("Expeceted no error for %s, got: %s", c.value, err)
		}
		if res != c.expected {
			t.Errorf("Expeceted %s for %s, got: %s", c.expected, c.value, res)
		}
	}
}
