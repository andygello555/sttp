package main

import (
	"fmt"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	for _, test := range []string{
		"(a | b*) a (b | e)*",
		"c*(a|b)((a|c)b)*",
		"a*a(ba|(b|e))(b|e)*",
		"(a|b)a(b|e)*",
	}{
		_, parsed := Parse(test)
		if parsed.String() != strings.Replace(test, " ", "", -1) {
			t.Errorf("Parsed regular expression: '%s', does not match '%s'", test, parsed)
		}
	}
}

func TestThompson(t *testing.T) {
	for _, test := range []struct{
		regex string
		adjacencyList AdjacencyList
	}{
		{
			regex: "(a | b*) a (b | e)*",
			adjacencyList: AdjacencyList{
				StateKeyString("0"): []Edge{{"a", State(0), State(1)}},
				StateKeyString("1"): []Edge{{"e", State(1), State(7)}},
				StateKeyString("2"): []Edge{{"b", State(2), State(3)}},
				StateKeyString("3"): []Edge{
					{"e", State(3), State(2)},
					{"e", State(3), State(5)},
				},
				StateKeyString("4"): []Edge{
					{"e", State(4), State(5)},
					{"e", State(4), State(2)},
				},
				StateKeyString("5"): []Edge{{"e", State(5), State(7)}},
				StateKeyString("6"): []Edge{
					{"e", State(6), State(0)},
					{"e", State(6), State(4)},
				},
				StateKeyString("7"): []Edge{{"e", State(7), State(8)}},
				StateKeyString("8"): []Edge{{"a", State(8), State(9)}},
				StateKeyString("9"): []Edge{{"e", State(9), State(16)}},
				StateKeyString("10"): []Edge{{"b", State(10), State(11)}},
				StateKeyString("11"): []Edge{{"e", State(11), State(15)}},
				StateKeyString("12"): []Edge{{"e", State(12), State(13)}},
				StateKeyString("13"): []Edge{{"e", State(13), State(15)}},
				StateKeyString("14"): []Edge{
					{"e", State(14), State(10)},
					{"e", State(14), State(12)},
				},
				StateKeyString("15"): []Edge{
					{"e", State(15), State(14)},
					{"e", State(15), State(17)},
				},
				StateKeyString("16"): []Edge{
					{"e", State(16), State(17)},
					{"e", State(16), State(14)},
				},
			},
		},
		{
			regex: "c*(a|b)((a|c)b)*",
			adjacencyList: AdjacencyList{
				StateKeyString("0"): {{"c", State(0), State(1)}},
				StateKeyString("1"): {
					{"e", State(1), State(0)},
					{"e", State(1), State(3)},
				},
				StateKeyString("10"): {{"a", State(10), State(11)}},
				StateKeyString("11"): {{"e", State(11), State(15)}},
				StateKeyString("12"): {{"c", State(12), State(13)}},
				StateKeyString("13"): {{"e", State(13), State(15)}},
				StateKeyString("14"): {
					{"e", State(14), State(10)},
					{"e", State(14), State(12)},
				},
				StateKeyString("15"): {{"e", State(15), State(16)}},
				StateKeyString("16"): {{"b", State(16), State(17)}},
				StateKeyString("17"): {
					{"e", State(17), State(14)},
					{"e", State(17), State(19)},
				},
				StateKeyString("18"): {
					{"e", State(18), State(19)},
					{"e", State(18), State(14)},
				},
				StateKeyString("2"): {
					{"e", State(2), State(3)},
					{"e", State(2), State(0)},
				},
				StateKeyString("3"): {{"e", State(3), State(8)}},
				StateKeyString("4"): {{"a", State(4), State(5)}},
				StateKeyString("5"): {{"e", State(5), State(9)}},
				StateKeyString("6"): {{"b", State(6), State(7)}},
				StateKeyString("7"): {{"e", State(7), State(9)}},
				StateKeyString("8"): {
					{"e", State(8), State(4)},
					{"e", State(8), State(6)},
				},
				StateKeyString("9"): {{"e", State(9), State(18)}},
			},},
		{
			regex: "a*a(ba|(b|e))(b|e)*",
			adjacencyList: AdjacencyList{
				StateKeyString("0"): {{"a", State(0), State(1)}},
				StateKeyString("1"): {
					{"e", State(1), State(0)},
					{"e", State(1), State(3)},
				},
				StateKeyString("10"): {{"b", State(10), State(11)}},
				StateKeyString("11"): {{"e", State(11), State(15)}},
				StateKeyString("12"): {{"e", State(12), State(13)}},
				StateKeyString("13"): {{"e", State(13), State(15)}},
				StateKeyString("14"): {
					{"e", State(14), State(10)},
					{"e", State(14), State(12)},
				},
				StateKeyString("15"): {{"e", State(15), State(17)}},
				StateKeyString("16"): {
					{"e", State(16), State(6)},
					{"e", State(16), State(14)},
				},
				StateKeyString("17"): {{"e", State(17), State(24)}},
				StateKeyString("18"): {{"b", State(18), State(19)}},
				StateKeyString("19"): {{"e", State(19), State(23)}},
				StateKeyString("2"): {
					{"e", State(2), State(3)},
					{"e", State(2), State(0)},
				},
				StateKeyString("20"): {{"e", State(20), State(21)}},
				StateKeyString("21"): {{"e", State(21), State(23)}},
				StateKeyString("22"): {
					{"e", State(22), State(18)},
					{"e", State(22), State(20)},
				},
				StateKeyString("23"): {
					{"e", State(23), State(22)},
					{"e", State(23), State(25)},
				},
				StateKeyString("24"): {
					{"e", State(24), State(25)},
					{"e", State(24), State(22)},
				},
				StateKeyString("3"): {{"e", State(3), State(4)}},
				StateKeyString("4"): {{"a", State(4), State(5)}},
				StateKeyString("5"): {{"e", State(5), State(16)}},
				StateKeyString("6"): {{"b", State(6), State(7)}},
				StateKeyString("7"): {{"e", State(7), State(8)}},
				StateKeyString("8"): {{"a", State(8), State(9)}},
				StateKeyString("9"): {{"e", State(9), State(17)}},
			},
		},
		{
			regex: "(a|b)a(b|e)*",
			adjacencyList: AdjacencyList{
				StateKeyString("0"): {{"a", State(0), State(1)}},
				StateKeyString("1"): {{"e", State(1), State(5)}},
				StateKeyString("10"): {{"e", State(10), State(11)}},
				StateKeyString("11"): {{"e", State(11), State(13)}},
				StateKeyString("12"): {
					{"e", State(12), State(8)},
					{"e", State(12), State(10)},
				},
				StateKeyString("13"): {
					{"e", State(13), State(12)},
					{"e", State(13), State(15)},
				},
				StateKeyString("14"): {
					{"e", State(14), State(15)},
					{"e", State(14), State(12)},
				},
				StateKeyString("2"): {{"b", State(2), State(3)}},
				StateKeyString("3"): {{"e", State(3), State(5)}},
				StateKeyString("4"): {
					{"e", State(4), State(0)},
					{"e", State(4), State(2)},
				},
				StateKeyString("5"): {{"e", State(5), State(6)}},
				StateKeyString("6"): {{"a", State(6), State(7)}},
				StateKeyString("7"): {{"e", State(7), State(14)}},
				StateKeyString("8"): {{"b", State(8), State(9)}},
				StateKeyString("9"): {{"e", State(9), State(13)}},
			},
		},
	}{
		graph, _, _, err := Thompson(test.regex)
		if err != nil || test.adjacencyList == nil || !graph.NFA.Equal(&test.adjacencyList) {
			t.Errorf("Thompson's construction of regular expression: %s, does not produce the expected AdjacencyList", test.regex)
		}
	}
}

func TestGraph_Subset(t *testing.T) {
	for _, test := range []struct{
		regex string
		adjacencyList AdjacencyList
	}{
		{
			regex: "(a | b*) a (b | e)*",
			adjacencyList: AdjacencyList{
				StateKeyString("[0 2 4 5 6 7 8]"): {
					{"b", StateKeyString("[0 2 4 5 6 7 8]"), StateKeyString("[2 3 5 7 8]")},
					{"a", StateKeyString("[0 2 4 5 6 7 8]"), StateKeyString("[1 7 8 9 10 12 13 14 15 16 17]")},
				},
				StateKeyString("[1 7 8 9 10 12 13 14 15 16 17]"): {
					{"b", StateKeyString("[1 7 8 9 10 12 13 14 15 16 17]"), StateKeyString("[10 11 12 13 14 15 17]")},
					{"a", StateKeyString("[1 7 8 9 10 12 13 14 15 16 17]"), StateKeyString("[9 10 12 13 14 15 16 17]")},
				},
				StateKeyString("[10 11 12 13 14 15 17]"): {
					{"b", StateKeyString("[10 11 12 13 14 15 17]"), StateKeyString("[10 11 12 13 14 15 17]")},
				},
				StateKeyString("[2 3 5 7 8]"): {
					{"b", StateKeyString("[2 3 5 7 8]"), StateKeyString("[2 3 5 7 8]")},
					{"a", StateKeyString("[2 3 5 7 8]"), StateKeyString("[9 10 12 13 14 15 16 17]")},
				},
				StateKeyString("[9 10 12 13 14 15 16 17]"): {
					{"b", StateKeyString("[9 10 12 13 14 15 16 17]"), StateKeyString("[10 11 12 13 14 15 17]")},
				},
			},
		},
		{
			regex: "c*(a|b)((a|c)b)*",
			adjacencyList: AdjacencyList{
				StateKeyString("[0 1 3 4 6 8]"): {
					{"c", StateKeyString("[0 1 3 4 6 8]"), StateKeyString("[0 1 3 4 6 8]")},
					{"a", StateKeyString("[0 1 3 4 6 8]"), StateKeyString("[5 9 10 12 14 18 19]")},
					{"b", StateKeyString("[0 1 3 4 6 8]"), StateKeyString("[7 9 10 12 14 18 19]")},
				},
				StateKeyString("[0 2 3 4 6 8]"): {
					{"a", StateKeyString("[0 2 3 4 6 8]"), StateKeyString("[5 9 10 12 14 18 19]")},
					{"b", StateKeyString("[0 2 3 4 6 8]"), StateKeyString("[7 9 10 12 14 18 19]")},
					{"c", StateKeyString("[0 2 3 4 6 8]"), StateKeyString("[0 1 3 4 6 8]")},
				},
				StateKeyString("[10 12 14 17 19]"): {
					{"a", StateKeyString("[10 12 14 17 19]"), StateKeyString("[11 15 16]")},
					{"c", StateKeyString("[10 12 14 17 19]"), StateKeyString("[13 15 16]")},
				},
				StateKeyString("[11 15 16]"): {
					{"b", StateKeyString("[11 15 16]"), StateKeyString("[10 12 14 17 19]")},
				},
				StateKeyString("[13 15 16]"): {
					{"b", StateKeyString("[13 15 16]"), StateKeyString("[10 12 14 17 19]")},
				},
				StateKeyString("[5 9 10 12 14 18 19]"): {
					{"a", StateKeyString("[5 9 10 12 14 18 19]"), StateKeyString("[11 15 16]")},
					{"c", StateKeyString("[5 9 10 12 14 18 19]"), StateKeyString("[13 15 16]")},
				},
				StateKeyString("[7 9 10 12 14 18 19]"): {
					{"a", StateKeyString("[7 9 10 12 14 18 19]"), StateKeyString("[11 15 16]")},
					{"c", StateKeyString("[7 9 10 12 14 18 19]"), StateKeyString("[13 15 16]")},
				},
			},
		},
		{
			regex: "a*a(ba|(b|e))(b|e)*",
			adjacencyList: AdjacencyList{
				StateKeyString("[0 1 3 4 5 6 10 12 13 14 15 16 17 18 20 21 22 23 24 25]"): {
					{"a", StateKeyString("[0 1 3 4 5 6 10 12 13 14 15 16 17 18 20 21 22 23 24 25]"), StateKeyString("[0 1 3 4 5 6 10 12 13 14 15 16 17 18 20 21 22 23 24 25]")},
					{"b", StateKeyString("[0 1 3 4 5 6 10 12 13 14 15 16 17 18 20 21 22 23 24 25]"), StateKeyString("[7 8 11 15 17 18 19 20 21 22 23 24 25]")},
				},
				StateKeyString("[0 2 3 4]"): {
					{"a", StateKeyString("[0 2 3 4]"), StateKeyString("[0 1 3 4 5 6 10 12 13 14 15 16 17 18 20 21 22 23 24 25]")},
				},
				StateKeyString("[18 19 20 21 22 23 25]"): {
					{"b", StateKeyString("[18 19 20 21 22 23 25]"), StateKeyString("[18 19 20 21 22 23 25]")},
				},
				StateKeyString("[7 8 11 15 17 18 19 20 21 22 23 24 25]"): {
					{"b", StateKeyString("[7 8 11 15 17 18 19 20 21 22 23 24 25]"), StateKeyString("[18 19 20 21 22 23 25]")},
					{"a", StateKeyString("[7 8 11 15 17 18 19 20 21 22 23 24 25]"), StateKeyString("[9 17 18 20 21 22 23 24 25]")},
				},
				StateKeyString("[9 17 18 20 21 22 23 24 25]"): {
					{"b", StateKeyString("[9 17 18 20 21 22 23 24 25]"), StateKeyString("[18 19 20 21 22 23 25]")},
				},
			},
		},
		{
			regex: "(a|b)a(b|e)*",
			adjacencyList: AdjacencyList{
				StateKeyString("[0 2 4]"): {
					{"a", StateKeyString("[0 2 4]"), StateKeyString("[1 5 6]")},
					{"b", StateKeyString("[0 2 4]"), StateKeyString("[3 5 6]")},
				},
				StateKeyString("[1 5 6]"): {
					{"a", StateKeyString("[1 5 6]"), StateKeyString("[7 8 10 11 12 13 14 15]")},
				},
				StateKeyString("[3 5 6]"): {
					{"a", StateKeyString("[3 5 6]"), StateKeyString("[7 8 10 11 12 13 14 15]")},
				},
				StateKeyString("[7 8 10 11 12 13 14 15]"): {
					{"b", StateKeyString("[7 8 10 11 12 13 14 15]"), StateKeyString("[8 9 10 11 12 13 15]")},
				},
				StateKeyString("[8 9 10 11 12 13 15]"): {
					{"b", StateKeyString("[8 9 10 11 12 13 15]"), StateKeyString("[8 9 10 11 12 13 15]")},
				},
			},
		},
	}{
		graph, _, _, err := Thompson(test.regex)
		graph.Subset()
		if err != nil || !graph.DFA.Equal(&test.adjacencyList) {
			t.Errorf("Subset construction of regular expression: %s, does not produce the expected AdjacencyList", test.regex)
		}
	}
}

func TestTransitionTable_DeadStateMinimisation(t *testing.T) {
	for i, test := range []struct{
		inGraph Graph
		mergedStates MergedStates
	}{
		{
			inGraph: Graph{
				DFA: AdjacencyList{
					StateKeyString("T0"): {
						{"a", StateKeyString("T0"), StateKeyString("T1")},
						{"b", StateKeyString("T0"), StateKeyString("T3")},
					},
					StateKeyString("T1"): {
						{"a", StateKeyString("T1"), StateKeyString("T2")},
						{"b", StateKeyString("T1"), StateKeyString("T3")},
					},
					StateKeyString("T2"): {
						{"a", StateKeyString("T2"), StateKeyString("T2")},
						{"b", StateKeyString("T2"), StateKeyString("T3")},
						{"c", StateKeyString("T2"), StateKeyString("T6")},
					},
					StateKeyString("T3"): {
						{"b", StateKeyString("T3"), StateKeyString("T4")},
					},
					StateKeyString("T4"): {
						{"c", StateKeyString("T4"), StateKeyString("T5")},
					},
					StateKeyString("T5"): {},
					StateKeyString("T6"): {
						{"b", StateKeyString("T6"), StateKeyString("T7")},
					},
					StateKeyString("T7"): {},
				},
				AcceptingStates: StateSetExistence{
					StateKeyString("T0"): true,
					StateKeyString("T1"): true,
					StateKeyString("T2"): true,
					StateKeyString("T3"): true,
					StateKeyString("T5"): true,
				},
			},
		},
		//{
		//	inGraph: Graph{
		//		DFA: AdjacencyList{
		//			StateKeyString("T0"): {
		//				{"a", StateKeyString("T0"), StateKeyString("T2")},
		//				{"b", StateKeyString("T0"), StateKeyString("T1")},
		//			},
		//			StateKeyString("T1"): {
		//				{"a", StateKeyString("T1"), StateKeyString("T3")},
		//			},
		//			StateKeyString("T2"): {
		//				{"a", StateKeyString("T2"), StateKeyString("T4")},
		//			},
		//			StateKeyString("T3"): {
		//				{"c", StateKeyString("T3"), StateKeyString("T5")},
		//			},
		//			StateKeyString("T4"): {
		//				{"c", StateKeyString("T4"), StateKeyString("T6")},
		//			},
		//			StateKeyString("T5"): {},
		//			StateKeyString("T6"): {},
		//		},
		//		AcceptingStates: StateSetExistence{
		//			StateKeyString("T3"): true,
		//			StateKeyString("T4"): true,
		//			StateKeyString("T5"): true,
		//		},
		//	},
		//},
		//{
		//	inGraph: Graph{
		//		DFA: AdjacencyList{
		//			StateKeyString("T0"): {
		//				{"a", StateKeyString("T0"), StateKeyString("T1")},
		//				{"b", StateKeyString("T0"), StateKeyString("T4")},
		//			},
		//			StateKeyString("T1"): {
		//				{"a", StateKeyString("T1"), StateKeyString("T2")},
		//			},
		//			StateKeyString("T2"): {
		//				{"c", StateKeyString("T2"), StateKeyString("T3")},
		//			},
		//			StateKeyString("T3"): {},
		//			StateKeyString("T4"): {
		//				{"a", StateKeyString("T4"), StateKeyString("T5")},
		//			},
		//			StateKeyString("T5"): {
		//				{"c", StateKeyString("T5"), StateKeyString("T6")},
		//			},
		//			StateKeyString("T6"): {},
		//		},
		//		AcceptingStates: StateSetExistence{
		//			StateKeyString("T2"): true,
		//			StateKeyString("T5"): true,
		//			StateKeyString("T6"): true,
		//		},
		//	},
		//},
	}{
		tt := InitTT(&test.inGraph)
		tt.DeadStateMinimisation()
		tt.Visualisation("deadstate_test")
		fmt.Println(i)
		fmt.Println(tt.String())
		fmt.Println(tt.MergedStates.String())
	}
}
