package bullet

import (
	"bytes"
	"io"
	"reflect"
	"testing"
)

func TestReadFrom(t *testing.T) {
	type args struct {
		rd io.Reader
	}
	tests := []struct {
		name    string
		args    args
		want    *Buttle
		wantErr bool
	}{
		{
			args: args{
				rd: bytes.NewBuffer(NewBullet(10010, 1, []byte("abcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigk")).Bytes()),
			},
			want: &Buttle{
				guid: 10010,
				cmd:  1,
				data: []byte("abcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigklmnumadfadfabcdefghigk"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReadFrom(tt.args.rd)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadFrom() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReadFrom() = %v, want %v", got, tt.want)
			}
		})
	}
}
