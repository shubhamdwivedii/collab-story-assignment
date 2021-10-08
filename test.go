package main

import "fmt"

/*
Lets say Given free-time schedule in the form (a - b) i.e., from 'a' to 'b' of n people, print all time intervals where all n participants are available.
It's like a calendar application suggesting possible meeting timinings.

Example:

Person1: (4 - 16), (18 - 24)
Person2: (2 - 14), (17 - 24)
		4 -

Person3: (6 - 8), (12 - 20)
Person4: (10 - 22)



Time interval when all are available: (12 - 14), (18 - 20).
*/

func main() {
	p1d1 := duration{
		start: 4,
		end:   16,
	}

	p1d2 := duration{
		start: 18,
		end:   24,
	}

	p2d1 := duration{
		start: 2,
		end:   14,
	}

	p2d2 := duration{
		start: 17,
		end:   24,
	}

	p3d1 := duration{
		start: 6,
		end:   8,
	}

	p3d2 := duration{
		start: 12,
		end:   20,
	}

	p4d1 := duration{
		start: 10,
		end:   22,
	}

	commonInterval := findCommonFree([][]duration{
		{p1d1, p1d2},
		{p2d1, p2d2},
		{p3d1, p3d2},
		{p4d1},
	})

	fmt.Println("Result", commonInterval)

}

type duration struct {
	start int
	end   int
}

func findCommonFree(durations [][]duration) [][2]int { // 10
	// 24
	durationMap := make(map[int][]int)

	for i := 1; i < 25; i++ {
		durationMap[i] = []int{} // 4 [Person1, Person2]
	}

	for idx, person := range durations {
		for _, interval := range person {
			for i := interval.start; i <= interval.end; i++ {
				durationMap[i] = append(durationMap[i], idx)
			}
		}
	}
	// 24 interations
	starthr := 1
	var intervals [][2]int
	var current [2]int //
	for starthr < 25 {
		persons := durationMap[starthr]
		if len(persons) == len(durations) {
			if current[0] == 0 {
				current[0] = starthr
			} else { // start
				current[1] = starthr
			}
		} else {
			if current[0] != 0 && current[1] != 0 {
				intervals = append(intervals, current)
				current[0], current[1] = 0, 0
			}
		}
		starthr++
	}

	return intervals
}

////////////////////////

func checkChars(arr []int) bool {
	stk := -1

	for i, v := range arr {
		if stk == 1 {
			stk = -1 //pop
			continue
		} else { // empty stack
			if v == 0 {
				if i == len(arr)-1 { // last element is 0 and stk is empty
					return true
				}
				continue
			} else {
				stk = v // 1
			}
		}
	}
	return false
}

// func main() {

// 	opt := checkChars([]int{0, 0, 0, 0})
// 	fmt.Println(opt)
// }
