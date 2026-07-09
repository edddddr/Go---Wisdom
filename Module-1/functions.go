package main


import "fmt"


// func add(x, y int) int {
// 	return x + y

// }


// func swap(x, y string) (string, string){
// 	return y, x
// }



func split(sum int) (x,y int){
	x = sum * 4/5
	y = sum - x
	return
}



func main(){


	// fmt.Println(add(10, 10))

	// fmt.Println(swap("hello", "worlds"))
	fmt.Println(split(17))

}