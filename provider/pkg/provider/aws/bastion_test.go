package aws

import (
	"testing"
)

func TestUserDataArgs_JoinedTags(t *testing.T) {
	type fields struct {
		ParameterName string
		Route         string
		Region        string
		TailscaleTags []string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "no tags",
			fields: fields{
				TailscaleTags: []string{},
			},
			want: "",
		},
		{
			name: "single tag",
			fields: fields{
				TailscaleTags: []string{"tag1"},
			},
			want: "tag:tag1",
		},
		{
			name: "multiple tags",
			fields: fields{
				TailscaleTags: []string{"tag1", "tag2", "tag3"},
			},
			want: "tag:tag1,tag:tag2,tag:tag3",
		},
		{
			name: "tags with spaces",
			fields: fields{
				TailscaleTags: []string{"tag 1", "tag 2"},
			},
			want: "tag:tag 1,tag:tag 2",
		},
		{
			name: "tags with special characters",
			fields: fields{
				TailscaleTags: []string{"tag-1", "tag@2", "tag#3"},
			},
			want: "tag:tag-1,tag:tag@2,tag:tag#3",
		},
		{
			name: "nil slice",
			fields: fields{
				TailscaleTags: nil,
			},
			want: "",
		},
		{
			name: "slice with empty strings",
			fields: fields{
				TailscaleTags: []string{"", ""},
			},
			want: "tag:,tag:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uda := &UserDataArgs{
				ParameterName: tt.fields.ParameterName,
				Route:         tt.fields.Route,
				Region:        tt.fields.Region,
				TailscaleTags: tt.fields.TailscaleTags,
			}
			if got := uda.JoinedTags(); got != tt.want {
				t.Errorf("UserDataArgs.JoinedTags() = %v, want %v", got, tt.want)
			}
		})
	}
}
