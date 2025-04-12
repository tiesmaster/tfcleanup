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
		{"no-op: string literal", `"hoi"`, `"hoi"`},
		{"no args: dissolve format()", `format("hoi")`, `"hoi"`},
		// {"string literal: inline into single string", `format("%s-%s", "hoi", "dag")`, `hoi-dag`},
		// TODO: add
		//    with variables, and locals
		//    enclosed in an array
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// expr, diags := hclsyntax.ParseExpression([]byte(tc.expr), "dummy.tf", hcl.Pos{Line: 1, Column: 1})
			// if len(diags) > 0 {
			// 	t.Errorf("EXPR: expression '%s' is not valid HCL: diagnostics: %v", tc.expr, diags)
			// 	return
			// }

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
		blockPos              hcl.Pos
		expectedAttributeName string
	}{
		{
			"single block",
			`block {
				hoi = "dag"
			}`,
			hcl.Pos{Line: 1, Column: 1},
			"hoi"},
	}
	for _, tc := range testCases {

		hclFile, _ := hclwrite.ParseConfig([]byte(tc.hcl), "dummy.tf", hcl.Pos{Line: 1, Column: 1})

		expr := getExpression(tc.hcl, tc.blockPos, tc.expectedAttributeName)

		_, resultAttrName := getAttributeForWrite(hclFile, expr)
		if resultAttrName != tc.expectedAttributeName {
			t.Errorf("getAttributeForWrite(%s) = _, %s; want %s", tc.hcl, resultAttrName, tc.expectedAttributeName)
		}
	}
}

func getExpression(hclText string, blockPos hcl.Pos, attrName string) hcl.Expression {
	hclFile, _ := hclsyntax.ParseConfig([]byte(hclText), "dummy.tf", hcl.Pos{Line: 1, Column: 1})
	bl := hclFile.BlocksAtPos(blockPos)
	attributes, _ := bl[0].Body.JustAttributes()
	return attributes[attrName].Expr
}
