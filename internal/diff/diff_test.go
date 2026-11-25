package diff

import "testing"

func TestUnified(t *testing.T) {
	cases := []struct {
		name     string
		old, new string
		context  int
		want     string
	}{
		{
			name: "identical",
			old:  "a\nb\n",
			new:  "a\nb\n",
			want: "",
		},
		{
			name:    "changed line with context",
			old:     "a\nb\nc\nd\ne\n",
			new:     "a\nb\nC\nd\ne\n",
			context: 1,
			want:    "@@\n  b\n- c\n+ C\n  d\n",
		},
		{
			name:    "distant changes get separate hunks",
			old:     "1\n2\n3\n4\n5\n6\n7\n8\n",
			new:     "x\n2\n3\n4\n5\n6\n7\ny\n",
			context: 1,
			want:    "@@\n- 1\n+ x\n  2\n@@\n  7\n- 8\n+ y\n",
		},
		{
			name: "new file",
			old:  "",
			new:  "a\n",
			want: "@@\n+ a\n",
		},
		{
			name:    "appended lines",
			old:     "a\n",
			new:     "a\nb\nc\n",
			context: 3,
			want:    "@@\n  a\n+ b\n+ c\n",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := Unified(testCase.old, testCase.new, testCase.context); got != testCase.want {
				t.Errorf("Unified() =\n%s\nwant\n%s", got, testCase.want)
			}
		})
	}
}
