package qfuncs

import "github.com/swelf19/tsqb/qtypes"

func RemoveEmpty(src []string) []string {
	dst := []string{}
	for _, s := range src {
		if s != "" {
			dst = append(dst, s)
		}
	}
	return dst
}

func Compose(composeMethod qtypes.ComposeMethod, nodes ...qtypes.Condition) qtypes.Condition {

	cn := qtypes.ComposedCondition{
		Conditions:    []qtypes.Condition{},
		ComposeMethod: composeMethod,
	}

	cn.Conditions = append(cn.Conditions, nodes...)

	return cn
}

func ComposeAnd(nodes ...qtypes.Condition) qtypes.Condition {
	return Compose(qtypes.WhereAnd, nodes...)
}

func ComposeOr(nodes ...qtypes.Condition) qtypes.Condition {
	return Compose(qtypes.WhereOr, nodes...)
}
