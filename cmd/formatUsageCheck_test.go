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
			name:                  "single attribute",
			hcl:                   `hoi = "dag"`,
			address:               hclAddress{[]hclBlockId{}, "hoi"},
			expectedAttributeName: "hoi",
		},
		{
			name: "block without labels",
			hcl: `block1 {
				bloeb = "blaat"
			}

			block2 {
				hoi = "dag"
			}`,
			address:               hclAddress{[]hclBlockId{{"block2", nil}}, "hoi"},
			expectedAttributeName: "hoi",
		},
		{
			name: "block with labels",
			hcl: `block "id1" "id2" {
				bloeb = "blaat"
			}

			block "id3" "id4" {
				hoi = "dag"
			}`,
			address:               hclAddress{[]hclBlockId{{"block", []string{"id3", "id4"}}}, "hoi"},
			expectedAttributeName: "hoi",
		},
	}
	for _, tc := range testCases {
		hclFile, _ := hclwrite.ParseConfig([]byte(tc.hcl), "dummy.tf", hcl.Pos{Line: 1, Column: 1})

		body, resultAttrName := getAttributeForWrite(hclFile, tc.address)
		if resultAttrName != tc.expectedAttributeName || body == nil || body.GetAttribute(resultAttrName) == nil {
			t.Errorf("getAttributeForWrite(%s) = %v, %s; want %s", tc.hcl, body, resultAttrName, tc.expectedAttributeName)
		}
	}
}

func TestEqualsLabels(t *testing.T) {
	testCases := []struct {
		name           string
		l1             []string
		l2             []string
		expectedEquals bool
	}{
		{
			name: "empty labels",
			l1: []string{},
			l2: []string{},
			expectedEquals: true,
		},
		{
			name: "single labels, same",
			l1: []string{"hoi"},
			l2: []string{"hoi"},
			expectedEquals: true,
		},
		{
			name: "multiple labels, same",
			l1: []string{"hoi", "dag", "bloeb"},
			l2: []string{"hoi", "dag", "bloeb"},
			expectedEquals: true,
		},
		{
			name: "single labels, different",
			l1: []string{"hoi"},
			l2: []string{"dag"},
			expectedEquals: false,
		},
		{
			name: "labels vs no labels",
			l1: []string{"hoi"},
			l2: []string{},
			expectedEquals: false,
		},
		{
			name: "labels vs no labels",
			l1: []string{},
			l2: []string{"hoi"},
			expectedEquals: false,
		},
		{
			name: "multiple labels, first different",
			l1: []string{"hoi", "dag", "bloeb"},
			l2: []string{"hello", "dag", "bloeb"},
			expectedEquals: false,
		},
	}

	for _, tc := range testCases {
		result := equalsLabels(tc.l1, tc.l2)
		if result != tc.expectedEquals {
			t.Errorf("equalsLabels(%s, %s) == %t; want %t", tc.l1, tc.l2, result, tc.expectedEquals)
		}
	}
}
