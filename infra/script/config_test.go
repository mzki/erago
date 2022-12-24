package script

import "testing"

func Test_validateScriptPath(t *testing.T) {
	const baseDir = "testing"
	const baseDirEndsSep = baseDir + "/"
	type args struct {
		p       string
		baseDir string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"No Error", args{baseDirEndsSep + "a_file", baseDir}, false},
		{"No Error Dot Dot", args{baseDirEndsSep + "a_dir/../b_file", baseDir}, false},
		{"Error Upper BaseDir", args{baseDirEndsSep + "../c_file", baseDir}, true},
		{"Error Upper BaseDir from child", args{baseDirEndsSep + "d_file/../../../e_file", baseDir}, true},
		{"Error Upper BaseDir with same prefix name", args{baseDirEndsSep + "../" + baseDir + "2", baseDir}, true},
		{"Error Upper BaseDir with same prefix name, ends os.sep", args{baseDirEndsSep + "../" + baseDir + "2", baseDirEndsSep}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateScriptPath(tt.args.p, tt.args.baseDir); (err != nil) != tt.wantErr {
				t.Errorf("validateScriptPath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
