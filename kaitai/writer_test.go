package kaitai

// TODO: impl write test
// import (
// 	"bytes"
// 	"reflect"
// 	"testing"
// )

// func TestNewWriter(t *testing.T) {
// 	tests := []struct {
// 		name  string
// 		want  *Writer
// 		wantW string
// 	}{
// 		{"Test", NewWriter(&bytes.Buffer{}), ""},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			w := &bytes.Buffer{}
// 			if got := NewWriter(w); !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("NewWriter() = %v, want %v", got, tt.want)
// 			}
// 			if gotW := w.String(); gotW != tt.wantW {
// 				t.Errorf("NewWriter() = %v, want %v", gotW, tt.wantW)
// 			}
// 		})
// 	}
// }

// func TestWriter_WriteU1(t *testing.T) {
// 	type args struct {
// 		v uint8
// 	}
// 	tests := []struct {
// 		name    string
// 		k       *Writer
// 		args    args
// 		wantErr bool
// 	}{
// 		{"Test", NewWriter(&bytes.Buffer{}), args{1}, false},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if err := tt.k.WriteU1(tt.args.v); (err != nil) != tt.wantErr {
// 				t.Errorf("Writer.WriteU1() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }

// func TestWriter_WriteU2be(t *testing.T) {
// 	type args struct {
// 		v uint16
// 	}
// 	tests := []struct {
// 		name    string
// 		k       *Writer
// 		args    args
// 		wantErr bool
// 	}{
// 		{"Test", NewWriter(&bytes.Buffer{}), args{1}, false},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if err := tt.k.WriteU2be(tt.args.v); (err != nil) != tt.wantErr {
// 				t.Errorf("Writer.WriteU2be() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }

// func TestWriter_WriteU4be(t *testing.T) {
// 	type args struct {
// 		v uint32
// 	}
// 	tests := []struct {
// 		name    string
// 		k       *Writer
// 		args    args
// 		wantErr bool
// 	}{
// 		{"Test", NewWriter(&bytes.Buffer{}), args{1}, false},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if err := tt.k.WriteU4be(tt.args.v); (err != nil) != tt.wantErr {
// 				t.Errorf("Writer.WriteU4be() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }

// func TestWriter_WriteU8be(t *testing.T) {
// 	type args struct {
// 		v uint64
// 	}
// 	tests := []struct {
// 		name    string
// 		k       *Writer
// 		args    args
// 		wantErr bool
// 	}{
// 		{"Test", NewWriter(&bytes.Buffer{}), args{1}, false},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if err := tt.k.WriteU8be(tt.args.v); (err != nil) != tt.wantErr {
// 				t.Errorf("Writer.WriteU8be() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }

// func TestWriter_WriteU2le(t *testing.T) {
// 	type args struct {
// 		v uint16
// 	}
// 	tests := []struct {
// 		name    string
// 		k       *Writer
// 		args    args
// 		wantErr bool
// 	}{
// 		{"Test", NewWriter(&bytes.Buffer{}), args{1}, false},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if err := tt.k.WriteU2le(tt.args.v); (err != nil) != tt.wantErr {
// 				t.Errorf("Writer.WriteU2le() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }

// func TestWriter_WriteU4le(t *testing.T) {
// 	type args struct {
// 		v uint32
// 	}
// 	tests := []struct {
// 		name    string
// 		k       *Writer
// 		args    args
// 		wantErr bool
// 	}{
// 		{"Test", NewWriter(&bytes.Buffer{}), args{1}, false},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if err := tt.k.WriteU4le(tt.args.v); (err != nil) != tt.wantErr {
// 				t.Errorf("Writer.WriteU4le() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }

// func TestWriter_WriteU8le(t *testing.T) {
// 	type args struct {
// 		v uint64
// 	}
// 	tests := []struct {
// 		name    string
// 		k       *Writer
// 		args    args
// 		wantErr bool
// 	}{
// 		{"Test", NewWriter(&bytes.Buffer{}), args{1}, false},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if err := tt.k.WriteU8le(tt.args.v); (err != nil) != tt.wantErr {
// 				t.Errorf("Writer.WriteU8le() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }

// func TestWriter_WriteS1(t *testing.T) {
// 	type args struct {
// 		v int8
// 	}
// 	tests := []struct {
// 		name    string
// 		k       *Writer
// 		args    args
// 		wantErr bool
// 	}{
// 		{"Test", NewWriter(&bytes.Buffer{}), args{1}, false},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if err := tt.k.WriteS1(tt.args.v); (err != nil) != tt.wantErr {
// 				t.Errorf("Writer.WriteS1() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }

// func TestWriter_WriteS2be(t *testing.T) {
// 	type args struct {
// 		v int16
// 	}
// 	tests := []struct {
// 		name    string
// 		k       *Writer
// 		args    args
// 		wantErr bool
// 	}{
// 		{"Test", NewWriter(&bytes.Buffer{}), args{1}, false},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if err := tt.k.WriteS2be(tt.args.v); (err != nil) != tt.wantErr {
// 				t.Errorf("Writer.WriteS2be() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }

// func TestWriter_WriteS4be(t *testing.T) {
// 	type args struct {
// 		v int32
// 	}
// 	tests := []struct {
// 		name    string
// 		k       *Writer
// 		args    args
// 		wantErr bool
// 	}{
// 		{"Test", NewWriter(&bytes.Buffer{}), args{1}, false},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if err := tt.k.WriteS4be(tt.args.v); (err != nil) != tt.wantErr {
// 				t.Errorf("Writer.WriteS4be() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }

// func TestWriter_WriteS8be(t *testing.T) {
// 	type args struct {
// 		v int64
// 	}
// 	tests := []struct {
// 		name    string
// 		k       *Writer
// 		args    args
// 		wantErr bool
// 	}{
// 		{"Test", NewWriter(&bytes.Buffer{}), args{1}, false},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if err := tt.k.WriteS8be(tt.args.v); (err != nil) != tt.wantErr {
// 				t.Errorf("Writer.WriteS8be() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }

// func TestWriter_WriteS2le(t *testing.T) {
// 	type args struct {
// 		v int16
// 	}
// 	tests := []struct {
// 		name    string
// 		k       *Writer
// 		args    args
// 		wantErr bool
// 	}{
// 		{"Test", NewWriter(&bytes.Buffer{}), args{1}, false},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if err := tt.k.WriteS2le(tt.args.v); (err != nil) != tt.wantErr {
// 				t.Errorf("Writer.WriteS2le() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }

// func TestWriter_WriteS4le(t *testing.T) {
// 	type args struct {
// 		v int32
// 	}
// 	tests := []struct {
// 		name    string
// 		k       *Writer
// 		args    args
// 		wantErr bool
// 	}{
// 		{"Test", NewWriter(&bytes.Buffer{}), args{1}, false},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if err := tt.k.WriteS4le(tt.args.v); (err != nil) != tt.wantErr {
// 				t.Errorf("Writer.WriteS4le() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }

// func TestWriter_WriteS8le(t *testing.T) {
// 	type args struct {
// 		v int64
// 	}
// 	tests := []struct {
// 		name    string
// 		k       *Writer
// 		args    args
// 		wantErr bool
// 	}{
// 		{"Test", NewWriter(&bytes.Buffer{}), args{1}, false},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if err := tt.k.WriteS8le(tt.args.v); (err != nil) != tt.wantErr {
// 				t.Errorf("Writer.WriteS8le() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }

// func TestWriter_WriteF4be(t *testing.T) {
// 	type args struct {
// 		v float32
// 	}
// 	tests := []struct {
// 		name    string
// 		k       *Writer
// 		args    args
// 		wantErr bool
// 	}{
// 		{"Test", NewWriter(&bytes.Buffer{}), args{1}, false},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if err := tt.k.WriteF4be(tt.args.v); (err != nil) != tt.wantErr {
// 				t.Errorf("Writer.WriteF4be() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }

// func TestWriter_WriteF8be(t *testing.T) {
// 	type args struct {
// 		v float64
// 	}
// 	tests := []struct {
// 		name    string
// 		k       *Writer
// 		args    args
// 		wantErr bool
// 	}{
// 		{"Test", NewWriter(&bytes.Buffer{}), args{1}, false},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if err := tt.k.WriteF8be(tt.args.v); (err != nil) != tt.wantErr {
// 				t.Errorf("Writer.WriteF8be() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }

// func TestWriter_WriteF4le(t *testing.T) {
// 	type args struct {
// 		v float32
// 	}
// 	tests := []struct {
// 		name    string
// 		k       *Writer
// 		args    args
// 		wantErr bool
// 	}{
// 		{"Test", NewWriter(&bytes.Buffer{}), args{1}, false},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if err := tt.k.WriteF4le(tt.args.v); (err != nil) != tt.wantErr {
// 				t.Errorf("Writer.WriteF4le() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }

// func TestWriter_WriteF8le(t *testing.T) {
// 	type args struct {
// 		v float64
// 	}
// 	tests := []struct {
// 		name    string
// 		k       *Writer
// 		args    args
// 		wantErr bool
// 	}{
// 		{"Test", NewWriter(&bytes.Buffer{}), args{1}, false},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if err := tt.k.WriteF8le(tt.args.v); (err != nil) != tt.wantErr {
// 				t.Errorf("Writer.WriteF8le() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }

// func TestWriter_WriteBytes(t *testing.T) {
// 	type args struct {
// 		b []byte
// 	}
// 	tests := []struct {
// 		name    string
// 		k       *Writer
// 		args    args
// 		wantErr bool
// 	}{
// 		{"Test", NewWriter(&bytes.Buffer{}), args{[]byte("test")}, false},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if err := tt.k.WriteBytes(tt.args.b); (err != nil) != tt.wantErr {
// 				t.Errorf("Writer.WriteBytes() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }
