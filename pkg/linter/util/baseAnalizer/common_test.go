package baseAnalizer

import (
	"github.com/MishaSytch/linter/pkg/config"
	"reflect"
	"testing"
)

func Test_containSensitiveData(t *testing.T) {
	type args struct {
		msg string
		cfg *config.Config
	}
	cfg := &config.Config{
		SensitiveRules: config.SensitiveRules{
			Patterns: []config.SensitivePattern{
				{
					Name:  "password",
					Regex: "([1-5]{4,}[a-zA-Z]{4,}|[a-zA-Z]{4,}[1-5]{4,})",
				},
				{
					Name:  "email",
					Regex: "[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}",
				},
			},
			SensitiveWords: []string{
				"password",
				"email",
				"token",
			},
		},
	}
	tests := []struct {
		name               string
		args               args
		wantFixedMsg       string
		wantAllFoundIssues []string
	}{
		{
			"all legal",
			args{
				msg: "this is ok veryLongword",
				cfg: cfg,
			},
			"this is ok veryLongword",
			nil,
		},
		{
			"password",
			args{
				msg: "password: 1232rf",
				cfg: cfg,
			},
			"********: 1232rf",
			[]string{"password"},
		},
		{
			"password and pattern Password",
			args{
				msg: "password: 1234Pass",
				cfg: cfg,
			},
			"********: ********",
			[]string{"password", "password"},
		},
		{
			"token and pattern Password",
			args{
				msg: "token: 1232Rffg",
				cfg: cfg,
			},
			"********: ********",
			[]string{"token", "password"},
		},
		{
			"token and longer Password",
			args{
				msg: "token: 1232Rffg",
				cfg: cfg,
			},
			"********: ********",
			[]string{"token", "password"},
		},
		{
			"all words",
			args{
				msg: "token: 1321, password: 1232rfasd, email: misha@misha.ru",
				cfg: cfg,
			},
			"********: 1321, ********: ********, ********: ********",
			[]string{"password", "email", "token", "password", "email"},
		},
		{
			"hard",
			args{
				msg: "token: 1321, password: 1232rfPass, email: mishamisha.ru",
				cfg: cfg,
			},
			"********: 1321, ********: ********, ********: mishamisha.ru",
			[]string{"password", "email", "token", "password"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFixedMsg, gotAllFoundIssues := containSensitiveData(tt.args.msg, tt.args.cfg)
			if gotFixedMsg != tt.wantFixedMsg {
				t.Errorf("containSensitiveData() gotFixedMsg = %v, want %v", gotFixedMsg, tt.wantFixedMsg)
			}
			if !reflect.DeepEqual(gotAllFoundIssues, tt.wantAllFoundIssues) {
				t.Errorf("containSensitiveData() gotAllFoundIssues = %v, want %v", gotAllFoundIssues, tt.wantAllFoundIssues)
			}
		})
	}
}

func Test_isNonEnglish(t *testing.T) {
	type args struct {
		r rune
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"ok",
			args{
				'a',
			},
			false,
		},
		{
			"no",
			args{
				'ñ',
			},
			true,
		},
		{
			"ok",
			args{
				'1',
			},
			false,
		},
		{
			"ok",
			args{
				'🫡',
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isNonEnglish(tt.args.r); got != tt.want {
				t.Errorf("isNonEnglish() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isSpecialChars(t *testing.T) {
	type args struct {
		r rune
	}
	tests := []struct {
		name string
		args args
		want bool
	}{

		{
			"ok",
			args{
				'a',
			},
			false,
		},
		{
			"no",
			args{
				'ñ',
			},
			false,
		},
		{
			"ok",
			args{
				'1',
			},
			false,
		},
		{
			"ok",
			args{
				'🫡',
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isSpecialChars(tt.args.r); got != tt.want {
				t.Errorf("isSpecialChars() = %v, want %v", got, tt.want)
			}
		})
	}
}
