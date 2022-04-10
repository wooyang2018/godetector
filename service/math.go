package service

import (
	"errors"
	"fmt"
	"github.com/dengsgo/math-engine/engine"
	"sort"
	"sync"
)

/*
[drawings1 hentai2 neutral3 porn4 sexy5]
[children1 group_20-2 protest3 sign4 violence5]
[identity_attack1 insult2 obscene3 severe_toxicity4 threat5 toxicity6]
[其他1 广告2 政治3 暴恐4 民生5 网址6 色情7]
*/

type Input struct {
	nsfw    []float64 //低俗图片检测结果集
	protest []float64 //抗议暴力图片检测结果集
	cntext  []float64 //违规中文检测结果集
	entext  []float64 //违规英文检测结果集
	numtext []float64 //文本过滤结果集
}

func NewInput() *Input {
	i := Input{}
	i.nsfw = make([]float64, 5)
	i.protest = make([]float64, 5)
	i.cntext = make([]float64, 1)
	i.entext = make([]float64, 6)
	i.numtext = make([]float64, 7)
	return &i
}

func SortValueByKey(in map[string]float64) []float64 {
	var keys []string
	for k, _ := range in {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var res []float64
	for _, key := range keys {
		res = append(res, in[key])
	}
	return res
}

func (i *Input) SetNsfw(nsfw map[string]float64) {
	i.nsfw = SortValueByKey(nsfw)
}

func (i *Input) SetProtest(protest map[string]float64) {
	i.protest = SortValueByKey(protest)
}

func (i *Input) SetCntext(cntext float64) {
	i.cntext = []float64{cntext}
}

func (i *Input) SetEntext(entext map[string]float64) {
	i.entext = SortValueByKey(entext)
}

func (i *Input) SetNumtext(numtext map[string]float64) {
	i.numtext = SortValueByKey(numtext)
}

type MathAnalyzer struct {
	input *Input
	raw   map[string]string
	lock  sync.Mutex
	ast   map[string]engine.ExprAST
}

//NewAnalyzer 新建分析器
func NewMathAnalyzer(raw map[string]string) *MathAnalyzer {
	a := MathAnalyzer{}
	a.raw = raw
	a.input = NewInput()
	a.RegFunction()
	a.BuildAST()
	return &a
}

//RegFunction 注册自定义函数，保证在AST构建前调用该函数
func (a *MathAnalyzer) RegFunction() {
	engine.RegFunction("and", -1, func(expr ...engine.ExprAST) float64 {
		if len(expr) == 0 {
			panic(errors.New("calling function `and` must have at least one parameter."))
		}
		for i := 0; i < len(expr); i++ {
			v := engine.ExprASTResult(expr[i])
			if v <= 0 {
				return 0
			}
		}
		return 1
	})
	engine.RegFunction("or", -1, func(expr ...engine.ExprAST) float64 {
		if len(expr) == 0 {
			panic(errors.New("calling function `or` must have at least one parameter."))
		}
		for i := 0; i < len(expr); i++ {
			v := engine.ExprASTResult(expr[i])
			if v > 0 {
				return 1
			}
		}
		return 0
	})
	engine.RegFunction("argmax", -1, func(expr ...engine.ExprAST) float64 {
		if len(expr) == 0 {
			panic(errors.New("calling function `argmax` must have at least one parameter."))
		}
		index := 0
		maxNum := engine.ExprASTResult(expr[0])
		for i := 1; i < len(expr); i++ {
			v := engine.ExprASTResult(expr[i])
			if v > maxNum {
				maxNum = v
				index = i
			}
		}
		return float64(index + 1)
	})
	engine.RegFunction("argmin", -1, func(expr ...engine.ExprAST) float64 {
		if len(expr) == 0 {
			panic(errors.New("calling function `argmin` must have at least one parameter."))
		}
		index := 0
		minNum := engine.ExprASTResult(expr[0])
		for i := 1; i < len(expr); i++ {
			v := engine.ExprASTResult(expr[i])
			if v < minNum {
				minNum = v
				index = i
			}
		}
		return float64(index + 1)
	})
	engine.RegFunction("equal", 2, func(expr ...engine.ExprAST) float64 {
		if len(expr) != 2 {
			panic(errors.New("calling function `equal` must have two parameter."))
		}
		a := engine.ExprASTResult(expr[0])
		b := engine.ExprASTResult(expr[1])
		if a == b {
			return 1
		}
		return 0
	})
	engine.RegFunction("sign", 2, func(expr ...engine.ExprAST) float64 {
		if len(expr) != 2 {
			panic(errors.New("calling function `sign` must have two parameter."))
		}
		v := engine.ExprASTResult(expr[0])
		flag := engine.ExprASTResult(expr[1])
		if v >= flag {
			return 1
		}
		return 0
	})
	engine.RegFunction("nsfw", 1, func(expr ...engine.ExprAST) float64 {
		if len(expr) != 1 {
			panic(errors.New("calling function `nsfw` must have one parameter."))
		}
		v := engine.ExprASTResult(expr[0])
		if v < 1 || v > 5 {
			panic(errors.New("parameters must be between 1 to 5 for calling function `nsfw`."))
		}
		return a.input.nsfw[uint32(v)-1]
	})
	engine.RegFunction("pro", 1, func(expr ...engine.ExprAST) float64 {
		if len(expr) != 1 {
			panic(errors.New("calling function `protest` must have one parameter."))
		}
		v := engine.ExprASTResult(expr[0])
		if v < 1 || v > 5 {
			panic(errors.New("parameters must be between 1 to 5 for calling function `protest`."))
		}
		return a.input.protest[uint32(v)-1]
	})
	engine.RegFunction("cntext", 1, func(expr ...engine.ExprAST) float64 {
		if len(expr) != 1 {
			panic(errors.New("calling function `cntext` must have one parameter."))
		}
		v := engine.ExprASTResult(expr[0])
		if v != 1 {
			panic(errors.New("parameters must be 1 for calling function `cntext`."))
		}
		return a.input.cntext[uint32(v)-1]
	})
	engine.RegFunction("entext", 1, func(expr ...engine.ExprAST) float64 {
		if len(expr) != 1 {
			panic(errors.New("calling function `entext` must have one parameter."))
		}
		v := engine.ExprASTResult(expr[0])
		if v < 1 || v > 6 {
			panic(errors.New("parameters must be between 1 to 6 for calling function `entext`."))
		}
		return a.input.entext[uint32(v)-1]
	})
	engine.RegFunction("num", 1, func(expr ...engine.ExprAST) float64 {
		if len(expr) != 1 {
			panic(errors.New("calling function `numtext` must have one parameter."))
		}
		v := engine.ExprASTResult(expr[0])
		if v < 1 || v > 7 {
			panic(errors.New("parameters must be between 1 to 7 for calling function `numtext`."))
		}
		return a.input.numtext[uint32(v)-1]
	})
}

func buildAST(exp string) engine.ExprAST {
	// input text -> []token
	toks, err := engine.Parse(exp)
	if err != nil {
		fmt.Println("ERROR: " + err.Error())
		return nil
	}
	// []token -> AST Tree
	ast := engine.NewAST(toks, exp)
	if ast.Err != nil {
		fmt.Println("ERROR: " + ast.Err.Error())
		return nil
	}
	// AST builder
	ar := ast.ParseExpression()
	if ast.Err != nil {
		fmt.Println("ERROR: " + ast.Err.Error())
		return nil
	}
	return ar
}

func (a *MathAnalyzer) BuildAST() {
	a.ast = make(map[string]engine.ExprAST)
	for k, v := range a.raw {
		a.ast[k] = buildAST(v)
	}
}

func (a *MathAnalyzer) ASTInput(i *Input) {
	a.lock.Lock()
	a.input = i
	a.lock.Unlock()
}

func (a *MathAnalyzer) ExprASTResult() map[string]float64 {
	res := make(map[string]float64)
	// AST traversal -> result
	for k, v := range a.ast {
		r := engine.ExprASTResult(v)
		res[k] = r
	}
	return res
}

func (a *MathAnalyzer) ASTParseNsfw(nsfw map[string]float64) float64 {
	a.lock.Lock()
	defer a.lock.Unlock()
	a.input.SetNsfw(nsfw)
	res := engine.ExprASTResult(a.ast["nsfw"])
	return res
}

func (a *MathAnalyzer) ASTParseProtest(protest map[string]float64) float64 {
	a.lock.Lock()
	defer a.lock.Unlock()
	a.input.SetProtest(protest)
	res := engine.ExprASTResult(a.ast["protest"])
	return res
}

func (a *MathAnalyzer) ASTParseCntext(cntext float64) float64 {
	a.lock.Lock()
	defer a.lock.Unlock()
	a.input.SetCntext(cntext)
	res := engine.ExprASTResult(a.ast["cntext"])
	return res
}

func (a *MathAnalyzer) ASTParseEntext(entext map[string]float64) float64 {
	a.lock.Lock()
	defer a.lock.Unlock()
	a.input.SetEntext(entext)
	res := engine.ExprASTResult(a.ast["entext"])
	return res
}

func (a *MathAnalyzer) ASTParseNumtext(numtext map[string][]string) float64 {
	nummap := make(map[string]float64)
	for k, v := range numtext {
		nummap[k] = float64(len(v))
	}
	a.lock.Lock()
	defer a.lock.Unlock()
	a.input.SetNumtext(nummap)
	res := engine.ExprASTResult(a.ast["numtext"])
	return res
}
