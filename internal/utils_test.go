package internal

import (
	"os"
	"reflect"
	"strconv"
	"testing"
	"time"
)

func Test_format(t *testing.T) {
	type args struct {
		text   string
		params map[string]string
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "empty test",
			args: args{"", map[string]string{}},
			want: "",
		},

		{
			name: "name test",
			args: args{"hello this is me, ${name}", map[string]string{"name": "Shravan"}},
			want: "hello this is me, Shravan",
		},

		{
			name: "long sentence test",
			args: args{"${go} offers cool concurrency features like ${c1} and ${c2}. and it's ${adj}!", map[string]string{"go": "Golang", "c1": "goroutines", "c2": "channels", "adj": "amazing"}},
			want: "Golang offers cool concurrency features like goroutines and channels. and it's amazing!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := format(tt.args.text, tt.args.params); got != tt.want {
				t.Errorf("format() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_writeToFile(t *testing.T) {
	type args struct {
		text     string
		filename string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "write to file",
			args: args{
				text:     "hello this is me, name",
				filename: "test.txt",
			},
			wantErr: false,
		},
		{
			name: "write to file error",
			args: args{
				text:     "this test must fail",
				filename: "&*$*hvsgrv87@#$/|\\",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := writeToFile(tt.args.text, tt.args.filename); (err != nil) != tt.wantErr {
				t.Errorf("writeToFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	t.Cleanup(func() {
		if err := os.Remove("./test.txt"); err != nil {
			t.Errorf("Error removing file: %v", err)
		}
	})

}

func TestDurationFromNumber(t *testing.T) {
	type args[T numberLike] struct {
		number T
		unit   time.Duration
	}
	type testCase[S numberLike] struct {
		name        string
		args        args[S]
		want        time.Duration
		shouldPanic bool
	}
	floatTests := []testCase[float64]{
		{"fail1", args[float64]{number: 45, unit: time.Hour + 4}, 0, true},
		{"fail2", args[float64]{number: 45, unit: time.Second * 4}, 0, true},
		{"pass1", args[float64]{number: 25, unit: time.Second}, time.Second * 25, false},
		{"pass2", args[float64]{number: 789, unit: time.Microsecond}, 789 * time.Microsecond, false},
	}
	for _, tt := range floatTests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if err := recover(); err != nil {
					if !tt.shouldPanic {
						t.Errorf("DurationFromNumber() panicked when it shouldn't: args=%v", tt.args)
					}
				}
			}()
			if got := DurationFromNumber(tt.args.number, tt.args.unit); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DurationFromNumber() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMapFunc(t *testing.T) {
	type args[T, S any] struct {
		function func(T) S
		slice    []T
	}
	tests := []struct {
		name string
		args args[int, string]
		want []string
	}{
		{name: "pass1", args: args[int, string]{func(i int) string { return strconv.Itoa(i) }, []int{}}, want: []string{}},
		{name: "pass2", args: args[int, string]{func(i int) string { return strconv.Itoa(i) }, []int{1, 2, 12, 15}}, want: []string{"1", "2", "12", "15"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MapFunc[[]int, []string](tt.args.function, tt.args.slice); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MapFunc() = %v, want %v", got, tt.want)
			}
		})
	}
}
