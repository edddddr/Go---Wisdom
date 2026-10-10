package main

import (
	"math" 
	"fmt"
)

// func console(value, x, y int) {
// 	fmt.Println("The sum of : ", x, "and ", y, " is: ", value)
// }

// func add(x , y int,) (int , int , int){
// 	return x + y, x, y
// }

// func split(sum int) (x, y int) {
// 	x = sum * 4 / 9
// 	y = sum - x
// 	return
// }

func isXGreat(x, y int) bool {
	if x > y { 
		return true
	}else{
		return false
	}
}
type Vertex struct {
	X, Y float64
}

func (v Vertex) Abs() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

func main() {
	// i, j := 1, 2
	// f := float64(i)
	// c, python, java := false, true, false
	// sum, x, y := add(2, 4)
	// fmt.Println(sum, x, y)
	// console(sum, x, y)
	// fmt.Println(split(10))
	// fmt.Println(i, c, python, java, j, f)

	// sum := 1

	// for i := 0; i <= 5; i++ {
	//  sum += i
	// }

	// for sum < 5 {
	// 	sum += sum
	// }

	// fmt.Println(sum)
	// defer fmt.Println(isXGreat(8, 5))
	
	// i, j := 42, 2701
	// p := &i
	
	// fmt.Println(*p, j)
	// v := Vertex{1, 3}
	
	// var a[10] int
	// a[0] = 2

	

	// fmt.Println(a[0])

	v := Vertex{3, 4}
	fmt.Println(v.Abs())


}


