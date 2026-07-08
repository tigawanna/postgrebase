package daos

import (
	"testing"

	"github.com/zhenruyan/postgrebase/models/schema"
	"github.com/zhenruyan/postgrebase/tools/types"
)

func TestIsMySQLRequiredJSONField(t *testing.T) {
	tests := []struct {
		name  string
		field *schema.SchemaField
		want  bool
	}{
		{
			name:  "json field remains nullable",
			field: &schema.SchemaField{Type: schema.FieldTypeJson},
			want:  false,
		},
		{
			name:  "multiple select uses required json",
			field: &schema.SchemaField{Type: schema.FieldTypeSelect, Options: &schema.SelectOptions{MaxSelect: 2}},
			want:  true,
		},
		{
			name:  "single relation is varchar",
			field: &schema.SchemaField{Type: schema.FieldTypeRelation, Options: &schema.RelationOptions{MaxSelect: types.Pointer(1)}},
			want:  false,
		},
		{
			name:  "multiple relation uses required json",
			field: &schema.SchemaField{Type: schema.FieldTypeRelation, Options: &schema.RelationOptions{}},
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isMySQLRequiredJSONField(tt.field); got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}
