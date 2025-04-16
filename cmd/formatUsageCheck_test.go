package cmd

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

func TestConvertFormatToInterpolation(t *testing.T) {
	testCases := []struct {
		name     string
		expr     string
		expected string
	}{
		{
			name:     "no-op: string literal",
			expr:     `"hoi"`,
			expected: `"hoi"`,
		},
		{
			name:     "no args: dissolve format()",
			expr:     `format("hoi")`,
			expected: `"hoi"`,
		},
		{
			name:     "string literal: inline into single string",
			expr:     `format("%s-%s", "hoi", "dag")`,
			expected: `"hoi-dag"`,
		},
		{
			name:     "expr: wrap in template interpretation",
			expr:     `format("%s-%s", var.hoi, local.dag)`,
			expected: `"${var.hoi}-${local.dag}"`,
		},
		{
			name:     "array: single item",
			expr:     `["hoi"]`,
			expected: `["hoi"]`,
		},
		{
			name:     "array: multiple items item",
			expr:     `["hoi", "dag"]`,
			expected: `["hoi","dag"]`,
		},
		{
			name:     "array: with format call",
			expr:     `[format("hoi")]`,
			expected: `["hoi"]`,
		},
		{
			name:     "array: with many format calls",
			expr:     `[format("hoi"), format("%s-%s", var.hoi, local.dag)]`,
			expected: `["hoi","${var.hoi}-${local.dag}"]`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tokens, diags := hclsyntax.LexConfig([]byte(tc.expr), "dummy.tf", hcl.Pos{Line: 1, Column: 1})
			if len(diags) > 0 {
				t.Errorf("TOKENS: expression '%s' is not valid HCL: diagnostics: %v", tc.expr, diags)
				return
			}

			result := convertFormatToInterpolation(tokens)
			resultString := string(result.Bytes())
			if resultString != tc.expected {
				t.Errorf("convertFormatToInterpolation(%s) = %s; want %s", tc.expr, resultString, tc.expected)
			}
		})
	}
}

func TestGetAttributeForWrite(t *testing.T) {
	testCases := []struct {
		name                  string
		hcl                   string
		address               hclAddress
		expectedAttributeName string
	}{
		{
			name: "single block",
			hcl: `block {
				hoi = "dag"
			}`,
			address:               hclAddress{[]hclBlockId{{"block", nil}}, "hoi"},
			expectedAttributeName: "hoi"},
	}
	for _, tc := range testCases {

		hclFile, _ := hclwrite.ParseConfig([]byte(tc.hcl), "dummy.tf", hcl.Pos{Line: 1, Column: 1})

		body, resultAttrName := getAttributeForWrite(hclFile, tc.address)
		if resultAttrName != tc.expectedAttributeName || body == nil || body.GetAttribute(resultAttrName) == nil {
			t.Errorf("getAttributeForWrite(%s) = %v, %s; want %s", tc.hcl, body, resultAttrName, tc.expectedAttributeName)
		}
	}
}
