package process

import (
	"context"
	"reflect"
	"testing"
	"github.com/wptide/pkg/message"
)

func generateProcs(ctx context.Context, procs []Processor) <-chan Processor {
	out := make(chan Processor, len(procs))
	go func() {
		for _, proc := range procs {
			proc.SetContext(ctx)
			out <- proc
		}
	}()
	return out
}

func TestProcess_Run(t *testing.T) {
	type fields struct {
		context   context.Context
		Message   message.Message
		Result    *Result
		FilesPath string
	}
	tests := []struct {
		name    string
		fields  fields
		want    <-chan error
		wantErr bool
	}{
		{
			"No Override",
			fields{},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Process{
				context:   tt.fields.context,
				Message:   tt.fields.Message,
				Result:    tt.fields.Result,
				FilesPath: tt.fields.FilesPath,
			}
			got, err := p.Run()
			if (err != nil) != tt.wantErr {
				t.Errorf("Process.Run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Process.Run() = %v, want %v", got, tt.want)
			}
		})
	}
}
