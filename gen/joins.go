package gen

import (
	"fmt"
)

func getTablesFromJoin(joined StructMeta) []StructMeta {
	original := []StructMeta{}
	if joined.StructType == JoinedStruct {
		for _, s := range joined.JoinMembers {
			original = append(original, getTablesFromJoin(s)...)
		}
	} else {
		original = []StructMeta{joined}
	}
	return original
}

func JoinTables(original []StructMeta) []StructMeta {
	counter := 0
	for {
		var newelements int
		// fmt.Println("--------------original-----------------")
		// for _, j := range original {
		// 	fmt.Println(j.StructName)
		// 	for _, jm := range j.JoinMembers {
		// 		fmt.Print(jm.StructName, ", ")
		// 	}
		// 	fmt.Println()
		// 	fmt.Println("-------------------------------")
		// }

		joins := findJoins(original)
		// fmt.Println("--------------join-----------------")
		// for _, j := range joins {
		// 	fmt.Println(j.StructName)
		// 	for _, jm := range j.JoinMembers {
		// 		fmt.Print(jm.StructName, ", ")
		// 	}
		// 	fmt.Println()
		// 	fmt.Println("-------------------------------")
		// }
		original, newelements = mergeLists(original, joins)
		// fmt.Println(len(original), len(joins), newelements)
		// original, newelements = mergeLists(original, joins)
		// fmt.Println(joins)
		fmt.Println("=======================================")
		counter++
		if newelements == 0 || counter > 5 {
			break
		}
	}
	return original
}

func mergeLists(original, new []StructMeta) ([]StructMeta, int) {
	var newelements int
	for _, s := range new {
		found := find(s, original)
		if found == false {
			original = append(original, s)
			newelements++
		}
	}
	return original, newelements
}

func findByName(name string, structs []StructMeta) *StructMeta {
	for _, s := range structs {
		if s.StructName == name {
			return &s
		}
	}
	return nil
}

func isStructListEquals(first []StructMeta, second []StructMeta) bool {
	if len(first) != len(second) {
		return false
	}
	for _, s1 := range first {
		if !find(s1, second) {
			return false
		}
	}
	return true
}

func find(wanted StructMeta, structs []StructMeta) bool {
	if wanted.StructType == RegularStruct {
		for _, s := range structs {
			if wanted.StructName == s.StructName {
				return true
			}
		}
	} else {
		for _, s := range structs {
			if isStructListEquals(wanted.JoinMembers, s.JoinMembers) {
				return true
			}
		}
	}
	return false
}

func findJoins(original []StructMeta) []StructMeta {
	joinedList := []StructMeta{}
	for _, s := range original {
		for _, f := range s.Fields {
			if f.RelatedModelName != "" {
				related := findByName(f.RelatedModelName, original)
				if related != nil {
					newjoined := createJoinedStruct(s, *related, f.SqlFieldName)
					if !find(newjoined, original) && !find(newjoined, joinedList) {
						joinedList = append(joinedList, createJoinedStruct(s, *related, f.SqlFieldName))
					}
				}
			}
		}
	}
	return joinedList
}

func createJoinedStruct(left, right StructMeta, foreignKeyName string) StructMeta {
	joinedName := fmt.Sprintf("%sJoin%s", left.StructName, right.StructName)
	joinedTable := fmt.Sprintf(
		"%s join %s on %s.%s = %s.%s",
		left.TableName,
		right.TableName,
		left.TableName, foreignKeyName,
		right.TableName, "id")
	joinMembers, _ := mergeLists(getTablesFromJoin(left), getTablesFromJoin(right))
	fields := append(makeFieldsWithStructName(left, foreignKeyName), makeFieldsWithStructName(right, "")...)
	return StructMeta{
		StructType:  JoinedStruct,
		StructName:  joinedName,
		TableName:   joinedTable,
		Fields:      fields,
		JoinMembers: joinMembers,
		PackageName: left.PackageName,
		SelectOnly:  true,
	}
}

func makeFieldsWithStructName(s StructMeta, disableForeignKeyName string) []StructFieldMeta {
	fields := []StructFieldMeta{}
	for _, f := range s.Fields {
		newfield := f
		newfield.FieldName = fmt.Sprintf("%s%s", s.StructName, f.FieldName)
		// newfield := StructFieldMeta{
		// 	FieldName:    fmt.Sprintf("%s%s", s.StructName, f.FieldName),
		// 	OrigFieldName: f.OrigFieldName,
		// 	Type:         f.Type,
		// 	SqlFieldName: f.SqlFieldName,
		// 	TableName:    f.TableName,

		// }
		if disableForeignKeyName == f.SqlFieldName {
			newfield.RelatedModelName = ""
		}
		fields = append(fields, newfield)
	}
	return fields
}
