package filter

import (
	"testing"
)

func TestFilterExample(t *testing.T) {
	filter := NewFilter()
	err := filter.LoadWordDict("./sensitive-dicts/其他.txt")
	if err != nil {
		t.Errorf("LoadWordDict Failed: %v", err)
	}
	filter.AddWord("长者")
	t.Log(filter.Filter("我为长者续一秒"))       // 我为续一秒
	t.Log(filter.Replace("我为长者续一秒", '*')) // 我为**续一秒
	t.Log(filter.FindIn("我为长者续一秒"))       // true, 长者
	t.Log(filter.Validate("我为长者续一秒"))     // False, 长者
	t.Log(filter.FindAll("我为长者续一秒"))      // [长者]
	//测试过滤功能的抗噪效果，默认不开启抗噪效果
	t.Log(filter.FindIn("我为长@者续一秒")) // false
	filter.UpdateNoisePattern(`[\|\s&%$@*]+`)
	t.Log(filter.FindIn("我为长@者续一秒"))   // true, 长者
	t.Log(filter.Validate("我为长@者续一秒")) // False, 长者
}

func TestLoadDict(t *testing.T) {
	filter := NewFilter()
	err := filter.LoadWordDict("./sensitive-dicts/其他.txt")
	if err != nil {
		t.Errorf("fail to load dict %v", err)
	}
}

func TestLoadNetWordDict(t *testing.T) {
	filter := NewFilter()
	err := filter.LoadNetWordDict("https://raw.githubusercontent.com/importcjj/sensitive/master/dict/dict.txt")
	if err != nil {
		t.Errorf("fail to load dict %v", err)
	}
	if len(filter.trie.Root.Children) == 0 {
		t.Errorf("load dict empty")
	}
}
