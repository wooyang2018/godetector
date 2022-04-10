package service

import (
	"fmt"
	"github.com/dengsgo/math-engine/engine"
	"testing"
)

func TestExample(t *testing.T) {
	s := "1 + 2 * 6 / 4 + (456 - 8 * 9.2) - (2 + 4 ^ 5)"
	// call top level function
	r, err := engine.ParseAndExec(s)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%s = %v", s, r)
}

func TestAnalyzer(t *testing.T) {
	input := NewInput()
	input.SetNumtext(map[string]float64{
		"其他": 1,
		"广告": 1,
		"政治": 3,
		"暴恐": 1,
		"民生": 1,
		"网址": 1,
		"色情": 2,
	})
	input.SetNsfw(map[string]float64{
		"porn":     0.6816046833992004,
		"sexy":     0.2803221046924591,
		"neutral":  0.03365671634674072,
		"hentai":   0.0038238749839365482,
		"drawings": 0.0005926391459070146,
	})
	input.SetProtest(map[string]float64{
		"sign":     0.17170631885528564,
		"protest":  0.17033971846103668,
		"children": 0.07608138769865036,
		"violence": 0.07127571851015091,
		"group_20": 0.06621399521827698,
	})
	input.SetCntext(0.9769131)
	input.SetEntext(map[string]float64{
		"toxicity":        0.990424,
		"severe_toxicity": 0.07567663,
		"obscene":         0.9390912,
		"threat":          0.0045800083,
		"insult":          0.84097433,
		"identity_attack": 0.00858633,
	})
	raw := map[string]string{
		"nsfw":    "or(equal(argmax(nsfw(1),nsfw(2),nsfw(3),nsfw(4),nsfw(5)),2),equal(argmax(nsfw(1),nsfw(2),nsfw(3),nsfw(4),nsfw(5)),4))",
		"protest": "or(pro(1)-0.12,pro(2)-0.12,pro(3)-0.12,pro(4)-0.12,pro(5)-0.12)",
		"cntext":  "and(cntext(1)-0.7)",
		"entext":  "or(entext(1)-0.7,entext(2)-0.7,entext(3)-0.7,entext(4)-0.7,entext(5)-0.7,entext(6)-0.7)",
		"numtext": "and(num(1)*0.1+num(2)*0.1+num(3)*0.3+num(4)*0.1+num(5)*0.1+num(6)*0.1+num(7)*0.2-2)",
	}
	analyzer := NewMathAnalyzer(raw)
	analyzer.ASTInput(input)
	r := analyzer.ExprASTResult()
	fmt.Printf("%v\n", r)
}
