package ranking

import (
	"reflect"
	"testing"
)

func TestQuery_tokenize(t *testing.T) {
	type fields struct {
		Id    string
		Text  string
		Terms []string
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{
			name: "Simple words",
			fields: fields{
				Id:    "1",
				Text:  "hello world",
				Terms: nil,
			},
			want: []string{"hello", "world"},
		},
		{
			name: "Text with extra spaces",
			fields: fields{
				Id:    "2",
				Text:  "  spaced   out   text  ",
				Terms: nil,
			},
			want: []string{"spaced", "out", "text"},
		},
		{
			name: "Empty text",
			fields: fields{
				Id:    "3",
				Text:  "",
				Terms: nil,
			},
			want: []string{},
		},
		{
			name: "Text with special characters",
			fields: fields{
				Id:    "4",
				Text:  "hello, world! how's it going?",
				Terms: nil,
			},
			want: []string{"hello,", "world!", "how's", "it", "going?"},
		},
		{
			name: "Single word",
			fields: fields{
				Id:    "5",
				Text:  "single",
				Terms: nil,
			},
			want: []string{"single"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &Query{
				Id:    tt.fields.Id,
				Text:  tt.fields.Text,
				Terms: tt.fields.Terms,
			}
			q.tokenize()
			if !reflect.DeepEqual(q.Terms, tt.want) {
				t.Errorf("tokenize() = %v, want %v", q.Terms, tt.want)
			}
		})
	}
}
