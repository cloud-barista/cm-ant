package configuration

import (
	"testing"
)

func TestInitConfig(t *testing.T) {
	type test struct {
		desc string
		arg  string
	}

	cases := []test{
		{
			"아무런 값을 넘기지 않을 경우 파일을 잘 읽어온다.",
			"",
		},
		{
			"상대 경로 지정시 파일을 잘 읽어 온다.",
			"../..",
		},
		{
			"절대경로 지정시 파일을 잘 읽어 온다.",
			RootPath(),
		},
	}

	for _, v := range cases {
		t.Run(v.desc, func(t *testing.T) {
			err := InitConfig(v.arg)
			if err != nil {
				t.Errorf(err.Error())
			}
		})
	}
}
