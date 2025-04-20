package cmd

import (
	"errors"
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

type formatUsages map[string][]formatInvocation

type hclBlockId struct {
	typeName string
	labels   []string
}

type hclAddress struct {
	blocks   []hclBlockId
	attrName string
}

type formatInvocation struct {
	tokens     []hclsyntax.Token
	hclAddress hclAddress

	// TODO: remove expr, when fully switched over to hclAddress
	expr hcl.Expression
}

// CHECK

func checkForFormatUsage(files []string) (formatUsages, error) {
	result := make(map[string][]formatInvocation)
	for _, f := range files {
		violations, err := checkForFormatUsageInFile(f)
		if err != nil {
			return nil, err
		}

		if len(violations) > 0 {
			result[f] = violations
		}
	}
	return result, nil
}

func checkForFormatUsageInFile(file string) ([]formatInvocation, error) {
	tokens, err := readHclTokens(file)
	if err != nil {
		return nil, err
	}

	var violationTokens []hclsyntax.Token
	for _, token := range tokens {
		if isFormatToken(token) {
			violationTokens = append(violationTokens, token)
		}
	}

	result, err := convertTokensToExpressions(file, tokens, violationTokens)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func convertTokensToExpressions(filename string, fileTokens []hclsyntax.Token, violationTokens []hclsyntax.Token) ([]formatInvocation, error) {
	hclFile, diags := hclparse.NewParser().ParseHCLFile(filename)
	if diags.HasErrors() {
		return nil, errors.New("failed to parse TF file: " + diags.Error())
	}

	var result []formatInvocation
	for _, t := range violationTokens {
		expr := hclFile.OutermostExprAtPos(t.Range.Start)

		result = append(result, formatInvocation{
			getAllTokensForExpression(fileTokens, expr),
			getHclAddress(hclFile, expr),
			expr,
		})
	}
	return result, nil
}

func getAllTokensForExpression(fileTokens []hclsyntax.Token, expr hcl.Expression) []hclsyntax.Token {
	start := 0
	for i, t := range fileTokens {
		if t.Range.Start == expr.Range().Start {
			start = i
		}

		if t.Range.End == expr.Range().End {
			return fileTokens[start : i+1]
		}
	}

	panic("cannot reach")
}

func getHclAddress(hclFile *hcl.File, expr hcl.Expression) hclAddress {
	var blIds []hclBlockId

	blocks := hclFile.BlocksAtPos(expr.Range().Start)
	for _, bl := range blocks {
		blIds = append(blIds, hclBlockId{bl.Type, bl.Labels})
	}
	attr := hclFile.AttributeAtPos(expr.Range().Start)

	return hclAddress{blIds, attr.Name}
}

func isFormatToken(token hclsyntax.Token) bool {
	return token.Type == hclsyntax.TokenIdent && isTokenText(token, "format")
}

func (invoke formatInvocation) string() string {
	var str string
	for _, s := range invoke.tokens {
		str = str + string(s.Bytes)
	}
	return str
}

func (invoke formatInvocation) location() string {
	r := invoke.expr.Range()
	return fmt.Sprintf("%v:L%d:%d-%d", r.Filename, r.Start.Line, r.Start.Column, r.End.Column)
}

// FIX

func convertFormatUsageToInterpolation(report formatUsages) error {
	for filename, formatInvocations := range report {
		if len(formatInvocations) == 0 {
			continue
		}
		err := convertFormatUsageToInterpolationForFile(filename, formatInvocations)
		if err != nil {
			return err
		}
	}
	return nil
}

func convertFormatUsageToInterpolationForFile(filename string, formatInvocations []formatInvocation) error {
	return patchFile(filename, func(hclFile *hclwrite.File) (*hclwrite.File, error) {
		for _, fi := range formatInvocations {
			body, attrName := getAttributeForWrite(hclFile, fi.hclAddress)
			body.SetAttributeRaw(attrName, convertFormatToInterpolation(fi.tokens))
		}
		return hclFile, nil
	})
}

func getAttributeForWrite(hclFile *hclwrite.File, address hclAddress) (*hclwrite.Body, string) {
	body := hclFile.Body()

	for _, blAddr := range address.blocks {
		for _, bl := range hclFile.Body().Blocks() {
			if isAddr(blAddr, bl) {
				body = bl.Body()
				break
			}
		}
	}

	return body, address.attrName
}

func isAddr(blAddr hclBlockId, bl *hclwrite.Block) bool {
	return blAddr.typeName == bl.Type() && equalsLabels(blAddr.labels, bl.Labels())
}

func equalsLabels(l1 []string, l2 []string) bool {
	if len(l1) != len(l2) {
		return false
	}
	for i, l := range l1 {
		if l != l2[i] {
			return false
		}
	}
	return true
}

func convertFormatToInterpolation(tokens []hclsyntax.Token) hclwrite.Tokens {
	var resultTokens []*hclwrite.Token
	for i := 0; i < len(tokens); i++ {
		t := tokens[i]
		if t.Type == hclsyntax.TokenIdent && string(t.Bytes) == "format" {
			newTokens, consumedTokens := parseFormatAndReturnInterpolationTokens(tokens[i:])
			resultTokens = append(resultTokens, newTokens...)
			i += consumedTokens - 1 // account for the next loop increase
		} else {
			resultTokens = append(resultTokens, toHclwriteToken(t))
		}
	}
	return resultTokens
}

func parseFormatAndReturnInterpolationTokens(tokens []hclsyntax.Token) ([]*hclwrite.Token, int) {
	var resultTokens []*hclwrite.Token

	i := 0

	// eat format token
	i++

	// eat open bracket token
	i++

	fmtString, tokensConsumed := getFormatString(tokens[i:])
	i += tokensConsumed

	fmtArgs, tokensConsumed := getFormatArgs(tokens[i:])
	i += tokensConsumed

	// push opening quote token
	resultTokens = append(resultTokens, quoteToken()) // TODO: fix the type that we use here (is now closing quote, should be opening)
	// parse the fmt string
	resultTokens = append(resultTokens, parseFmtString(fmtString, fmtArgs)...)
	// push closing quote token
	resultTokens = append(resultTokens, quoteToken())

	return resultTokens, i
}

func getFormatString(tokens []hclsyntax.Token) (string, int) {
	var bytes []byte
	var tokensConsumed int

	// consume opening quote token
	tokens = tokens[1:]
	tokensConsumed++

	for _, t := range tokens {
		tokensConsumed++
		if t.Type == hclsyntax.TokenCQuote {
			break
		} else {
			bytes = append(bytes, t.Bytes...)
		}
	}

	return string(bytes), tokensConsumed
}

func getFormatArgs(tokens []hclsyntax.Token) ([][]hclsyntax.Token, int) {
	var resultTokens [][]hclsyntax.Token
	var tokensConsumed int

	start := 0
	for i, t := range tokens {
		tokensConsumed++

		if start != 0 && // skip the first comma
			(t.Type == hclsyntax.TokenComma || t.Type == hclsyntax.TokenCParen) {
			resultTokens = append(resultTokens, tokens[start:i])
		}

		if t.Type == hclsyntax.TokenComma {
			start = i + 1 // set start to next token after the current comma
		}

		if t.Type == hclsyntax.TokenCParen {
			break
		}
	}

	return resultTokens, tokensConsumed
}

func parseFmtString(fmtString string, fmtArgs [][]hclsyntax.Token) []*hclwrite.Token {
	var resultTokens []*hclwrite.Token
	var stringFragment []byte

	for i := 0; i < len(fmtString); i++ {
		s := fmtString[i:min(i+2, len(fmtString))]
		if s == "%s" {
			arg := fmtArgs[0]
			if arg[0].Type == hclsyntax.TokenOQuote {
				// expression is string literal, so we can inline it
				stringFragment = append(stringFragment, arg[1].Bytes...)
			} else {
				// otherwise we need to start a template interpretation
				resultTokens = append(resultTokens, stringLiteralToken(stringFragment))
				stringFragment = nil

				resultTokens = append(resultTokens, openingTemplate())
				resultTokens = append(resultTokens, toHclwriteTokens(arg)...)
				resultTokens = append(resultTokens, closingTemplate())
			}

			fmtArgs = fmtArgs[1:] // consume the arg
			i += 1                // consume the %s
		} else {
			stringFragment = append(stringFragment, fmtString[i])
		}
	}

	// append last string fragment
	resultTokens = append(resultTokens, stringLiteralToken(stringFragment))

	return resultTokens
}

func toHclwriteTokens(tokens []hclsyntax.Token) []*hclwrite.Token {
	var resultTokens []*hclwrite.Token
	for _, t := range tokens {
		resultTokens = append(resultTokens, toHclwriteToken(t))
	}
	return resultTokens
}

func toHclwriteToken(token hclsyntax.Token) *hclwrite.Token {
	return &hclwrite.Token{
		Type:  token.Type,
		Bytes: token.Bytes,
	}
}

func quoteToken() *hclwrite.Token {
	return &hclwrite.Token{
		Type:  hclsyntax.TokenOQuote,
		Bytes: []byte(`"`)}
}

func stringLiteralToken(b []byte) *hclwrite.Token {
	return &hclwrite.Token{
		Type:  hclsyntax.TokenQuotedLit,
		Bytes: b}
}

func openingTemplate() *hclwrite.Token {
	return &hclwrite.Token{
		Type:  hclsyntax.TokenTemplateInterp,
		Bytes: []byte(`${`)}
}

func closingTemplate() *hclwrite.Token {
	return &hclwrite.Token{
		Type:  hclsyntax.TokenTemplateSeqEnd,
		Bytes: []byte(`}`)}
}
