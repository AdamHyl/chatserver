package acmachine

import (
	"testing"
)

var (
	dirty = []string{"assholes", "asswhole", "a_s_s", "b!tch", "b00bs", "b17ch", "b1tch"}
)

func TestReplace(t *testing.T) {
	ac := New(dirty)
	result := ac.MatchAndReplace("func ass asswholeb17b17ch bl b1tch", '*')
	t.Logf("%+v \n", result)
}

func BenchmarkReplace(b *testing.B) {
	ac := New(dirty)
	for i := 0; i < b.N; i++ {
		_ = ac.MatchAndReplace("func ass asswholeb17b17ch bl b1tch %v", '*')
	}
}
