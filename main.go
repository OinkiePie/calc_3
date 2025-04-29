package main

import "fmt"

func main() {
	minuscount := 0
	expression := "----3- +5"
	expressionres := ""
	for _, r := range expression {
		s := string(r)
		if s == "-" {
			minuscount++
			continue
		} else {
			if minuscount%2 != 0 {
				expressionres += "-"
				minuscount = 0
			}
			expressionres += s
		}
	}
	fmt.Println(expressionres)
}
